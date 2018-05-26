package addressfixer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
)

const (
	Five = "D5"
)

//Place is returned by Zippopotamus for a match with the zipcode.
type ZPlace struct {
	Name  string `json:"place name"`
	Long  string `json:"longitude"`
	State string
	Abbr  string `json:"state abbreviation"`
	Lat   string `json:"latitude"`
}

//Result is return by Zippotamus for all places that match a
//ZIP/postal code.
type ZResult struct {
	PostCode    string `json:"post code"`
	Country     string `json:"country"`
	CountryCode string `json:"country abbreviation"`
	Places      []ZPlace
}

//MatchPostal searches a list of regexes for a zipcode.  Returns
//country and try if there's a match.  Sources are StackTrace
//and https://rgxdb.com/
func MatchPostal(z string) (bool, string) {
	m := map[string]string{
		// Lots of countries have just \d{5}.  Need to disambiguate if we
		// start to see those folks in our databases.
		Five: `^\d{5}$`,
		"BR": `^\d{5}-?\d{3}$`,
		"CA": `^[A-Za-z]\d[A-Za-z][ -]?\d[A-Za-z]\d$`,
		"GB": `^([Gg][Ii][Rr] 0[Aa]{2})|((([A-Za-z][0-9]{1,2})|(([A-Za-z][A-Ha-hJ-Yj-y][0-9]{1,2})|(([AZa-z][0-9][A-Za-z])|([A-Za-z][A-Ha-hJ-Yj-y][0-9]?[A-Za-z]))))[0-9][A-Za-z]{2})$`,
		"NL": `^(?:NL-)?(?:[1-9]\d{3} ?(?:[A-EGHJ-NPRTVWXZ][A-EGHJ-NPRSTVWXZ]|S[BCEGHJ-NPRTVWXZ]))$`,
		"US": `^\d{5}(?:[-\s]\d{4})?$`}
	for c, p := range m {
		if regexp.MustCompile(p).MatchString(z) {
			return true, c
		}
	}
	return false, ""
}

//fiveDigits disambiguates a supporter record that has five digits in
//the zip code. Sets the country code.
func fiveDigits(s Supporter) string {
	m := strings.Split("fr,es,de,it", ",")
	for _, x := range m {
		if strings.HasSuffix(s.Email, "."+x) {
			return strings.ToUpper(x)
		}
	}
	return "US"
}

//City checks to see if the supporter's state is correct.  If not, then
//the record is changed and a Mod is added to the list of modifications.
func City(s Supporter, t ZResult, r []Mod) []Mod {
	s.City = strings.TrimSpace(s.City)
	name := t.Places[0].Name
	if len(s.City) == 0 {
		m := Mod{
			Key:    s.Key,
			Field:  "City",
			Old:    s.City,
			New:    name,
			Reason: fmt.Sprintf("Z Lookup for %v", s.Zip)}
		r = append(r, m)
		s.City = name
	}
	return r
}

//State checks to see if the supporter's state is correct.  If not, then
//the record is changed and a Mod is added to the list of modifications.
func State(s Supporter, t ZResult, r []Mod) []Mod {
	x := t.Places[0]
	// Not a good result, don't use it.
	if strings.Contains(x.Abbr, "Whistler") {
		return r
	}
	if s.State != x.Abbr {
		m := Mod{
			Key:    s.Key,
			Field:  "State",
			Old:    s.State,
			New:    x.Abbr,
			Reason: fmt.Sprintf("Z Lookup for %v", s.Zip)}
		r = append(r, m)
		s.State = x.Abbr
	}
	return r
}

//Country checks to see if the supporter's state is correct.  If not, then
//the record is changed and a Mod is added to the list of modifications.
func Country(s Supporter, t ZResult, r []Mod) []Mod {
	s.Country = strings.TrimSpace(s.Country)
	if len(s.Country) != 0 && s.Country != t.CountryCode {
		m := Mod{
			Key:    s.Key,
			Field:  "Country",
			Old:    s.Country,
			New:    t.CountryCode,
			Reason: fmt.Sprintf("Z Lookup for %v", s.Zip)}
		r = append(r, m)
		s.Country = t.CountryCode
	}
	return r
}

//Fetch retrieves information for a zip code.
func Fetch(s Supporter, c string) (ZResult, error) {
	// Zippopotamus only needs the first three digits for Canada.
	p := s.Zip
	switch c {
	case "CA":
		//log.Printf("zippo:91 p is '%v' p has %d chars\n", p, len(p))
		//Zippopotamus only needs the first three digits (FSA).
		if len(p) > 2 {
			p = p[0:3]
		}
	case "GB":
		re := regexp.MustCompile(`^\w+\d+`)
		p = re.FindString(p)
	case "":
		c = "US"
	}
	if c == "US" {
		if len(s.Zip) == 4 {
			zeroStates := strings.Split("CT,MA,MN,NH,NJ,PR,RI,VT,VI", ",")
			for _, x := range zeroStates {
				if s.State == x {
					s.Zip = "0" + s.Zip
					p = s.Zip
				}
			}
		}
		if strings.Contains(s.Zip, "-") {
			p = strings.Split(s.Zip, "-")[0]
		}
	}
	u := fmt.Sprintf("http://api.zippopotam.us/%v/%v", c, p)
	if c != "US" {
		log.Printf("Zippo:   Reading %v\n", u)
	}
	var body []byte
	var zr ZResult
	resp, err := http.Get(u)
	if resp == nil {
		err = fmt.Errorf("Key: %-8s HTTP null resonse object on %v", s.Key, u)
		return zr, err
	}
	if resp.StatusCode != 200 {
		err = fmt.Errorf("Key: %-8s HTTP error %v on %v", s.Key, resp.Status, u)
	}
	if err != nil {
		return zr, err
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return zr, err
	}
	err = json.Unmarshal(body, &zr)
	if err != nil {
		return zr, err
	}
	if len(zr.Places) == 0 {
		err = fmt.Errorf("no results for %v", s.Zip)
		return zr, err
	}
	return zr, err
}

//Zippo does a lookup using the free service from http://www.zippopotam.us/.
//Note that ambiguous results from Zippopotamus are not applied.
func Zippo(s Supporter, r []Mod) ([]Mod, error) {
	if len(s.Zip) == 0 {
		return r, nil
	}
	m, c := MatchPostal(s.Zip)
	if m {
		switch c {
		case "":
			c = "US"
		case Five:
			c = fiveDigits(s)
		}
		if c != s.Country && (c != "US" && len(s.Country) == 0) {
			m := Mod{
				Key:    s.Key,
				Field:  "Country",
				Old:    s.Country,
				New:    c,
				Reason: fmt.Sprintf("Z pattern match for %v\n", s.Zip)}
			r = append(r, m)
			s.Country = c

		}
	}
	zr, err := Fetch(s, c)
	if err != nil {
		return r, err
	}
	r = City(s, zr, r)
	r = State(s, zr, r)
	r = Country(s, zr, r)
	return r, err
}
