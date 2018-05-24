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

//isCA returns true if the provided postal code is a ZIP code.  Note
//that other countries also use five numeric digits for postal codes.
//We are ignoring the ambiguity for the time being.
func isCA(z string) bool {
	p := `^[A-Za-z]\d[A-Za-z][ -]?\d[A-Za-z]\d$`
	m := regexp.MustCompile(p).MatchString(z)
	return m
}

//isUS returns true if the provided postal code is a ZIP code.  Note
//that other countries also use five numeric digits for postal codes.
//We are ignoring the ambiguity for the time being.
func isUS(z string) bool {
	// Add check against phone number
	// Add check for domain name
	p := `^\d{5}(?:[-\s]\d{4})?$`
	m := regexp.MustCompile(p).MatchString(z)
	return m
}

//State checks to see if the supporter's state is correct.  If not, then
//the record is changed and a Mod is added to the list of modifications.
func State(s Supporter, t ZResult, r []Mod) []Mod {
	x := t.Places[0]
	if s.State != x.Abbr {
		m := Mod{
			Key:   s.Key,
			Field: "State",
			Old:   s.State,
			New:   x.Abbr}
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
			Key:   s.Key,
			Field: "Country",
			Old:   s.Country,
			New:   t.CountryCode}
		r = append(r, m)
		s.Country = t.CountryCode
	}
	return r
}

//Fetch retrieves information for a zip code.
func Fetch(s Supporter, c string) (ZResult, error) {
	// Zippopotamus only needs the first three digits for Canada.
	p := s.Zip
	if c == "CA" {
		p = p[0:3]
	}
	u := fmt.Sprintf("http://api.zippopotam.us/%v/%v", c, p)
	//log.Printf("Zippo:   Reading %v\n", u)
	var body []byte
	var zr ZResult
	resp, err := http.Get(u)
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
	//log.Printf("Zippo:   Result is %+v\n", zr)
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
	country := "US"
	if !isUS(s.Zip) {
		if isCA(s.Zip) {
			country = "CA"
		} else {
			log.Printf("Zippo:   Key: %-8s Zip: %-9s Comment: unknown country\n", s.Key, s.Zip)
			return r, nil
		}
	}
	zr, err := Fetch(s, country)
	if err != nil {
		return r, err
	}
	r = State(s, zr, r)
	r = Country(s, zr, r)
	return r, err
}
