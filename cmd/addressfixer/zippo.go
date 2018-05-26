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
	//Ca is the regex that matches postal codes in Canada.
	CA string = `^[A-Za-z]\d[A-Za-z][ -]?\d[A-Za-z]\d$`
	//GB is the regex that matches postal codes in Great Britain.  Very long...
	GB string = `^([Gg][Ii][Rr] 0[Aa]{2})|((([A-Za-z][0-9]{1,2})|(([A-Za-z][A-Ha-hJ-Yj-y][0-9]{1,2})|(([AZa-z][0-9][A-Za-z])|([A-Za-z][A-Ha-hJ-Yj-y][0-9]?[A-Za-z]))))[0-9][A-Za-z]{2})$`
	//NL the the regex that mathes postal codes for the Netherlands.  Also very long...
	NL string = `(?:NL-)?(?:[1-9]\d{3} ?(?:[A-EGHJ-NPRTVWXZ][A-EGHJ-NPRSTVWXZ]|S[BCEGHJ-NPRTVWXZ]))$`
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
//postal code.
type ZResult struct {
	PostCode    string `json:"post code"`
	Country     string `json:"country"`
	CountryCode string `json:"country abbreviation"`
	Places      []ZPlace
}

//PMap maps a regex pattern to a list of matching country codes.
type PMap map[string][]string

//MatchPostal searches a list of regexes for a zipcode.  Returns
//a matched indicator and the country code.  Sources are StackTrace
//and https://rgxdb.com/
func MatchPostal(s Supporter) (bool, string) {
	m := PMap{
		CA:        []string{"CA"},
		GB:        []string{"GB"},
		NL:        []string{"NL"},
		`^\d{6}$`: []string{"BY", "CN", "NN", "EC", "KZ", "KG", "NG", "RO", "RU", "SG", "TJ", "TT", "TM", "UZ", "VN"},
		`^\d{5}$`: []string{"AX", "AX", "BA", "BR", "BT", "CC", "CP", "CP", "CR", "DE", "DO", "DZ", "EE", "EG", "ES",
			"FR", "GT", "HR", "ID", "IQ", "IT", "KH", "KR", "KW", "LA", "LB", "LK", "MA", "ME", "MM",
			"MN", "MU", "MV", "MX", "MY", "NI", "NP", "PE", "PK", "PO", "PO", "PO", "RS", "RS", "SD",
			"TH", "TR", "TZ", "UA", "UY", "XK", "ZM"},
		`^\d{5}-?\d{3}$`: []string{"BR"},
		`^\d{4}$`: []string{"AF", "AL", "AR", "AM", "AU", "AT", "BD", "BE", "BG", "CV",
			"CX", "CC", "CY", "DK", "SV", "ET", "GE", "DE", "GL", "GW",
			"HT", "HU", "LR", "LI", "LU", "MK", "MZ", "NZ", "NE", "NF",
			"NO", "PA", "PY", "PH", "PT", "SG", "ZA", "CH", "SJ", "TN"},
		`^\d{3}$`:                     []string{"FO", "GN", "IS", "LS", "NG", "OM", "PS", "PG"},
		`^00120$`:                     []string{"VA"},
		`^00[6-9](?:[-\s]\d{4})`:      []string{"PR"},
		`^008[0-5]\d`:                 []string{"VI"},
		`^4789\d$`:                    []string{"SM"},
		`^96799(?:[-\s]\d{4})$`:       []string{"AS"},
		`^9691\d{2}(?:[-\s]\d{4})?$`:  []string{"GU"},
		`^9695[0-2](?:[-\s]\d{4})?$`:  []string{"MP"},
		`^96960$`:                     []string{"PW"},
		`^969[6-7]\d(?:[-\s]\d{4})?$`: []string{"MH"},
		`^9694[1-4](?:[-\s]\d{4})?$`:  []string{"FM"},
		`^971\d{2}$`:                  []string{"GP"},
		`^97133$`:                     []string{"BL"},
		`^97150$`:                     []string{"MF"},
		`^972\d{2}`:                   []string{"MQ"},
		`^973\d{2}$`:                  []string{"GF"},
		`^974\d{2}$`:                  []string{"RE"},
		`^975\d{2}$`:                  []string{"PM"},
		`^976\d{2}$`:                  []string{"YT"},
		`^980\d{2}$`:                  []string{"MC"},
		`^986\d{2}$`:                  []string{"WF"},
		`^987\d{2}$`:                  []string{"PF"},
		`^988\d{2}$`:                  []string{"NC"},
		`^LC`:                         []string{"LC"},
		`^PCRN`:                       []string{"PN"},
		`^SIQQ`:                       []string{"GS"},
		`^TKCA`:                       []string{"TC"},
	}

	if len(s.Zip) == 0 {
		return false, ""
	}
	for p, c := range m {
		if regexp.MustCompile(p).MatchString(s.Zip) {
			for _, x := range c {
				e := strings.ToUpper(s.Email)
				if strings.HasSuffix(e, "."+x) {
					log.Printf("Zippo:   Key: %8s '%v''%v' changing '%v' to '%v'\n", s.Key, s.Email, s.Country, x)
					if x == "US" && len(s.Country) == 0 {
						return true, s.Country
					}
					return true, x
				}
			}
		}
	}
	// Default to the US.  Open for discussion.
	return false, "US"
}

//City checks to see if the supporter's state is correct.  If not, then
//the record is changed and a Mod is added to the list of modifications.
func City(s Supporter, t ZResult, r []Mod) []Mod {
	s.City = strings.TrimSpace(s.City)
	name := t.Places[0].Name
	// Zippopotamus shows neighboring towns in a neighborhood
	// in parens.  Not really a good city name.
	if strings.Contains(name, "(") {
		name = strings.Split(name, "(")[0]
		name = strings.TrimSpace(name)
	}
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
	// Make adjustment to the postal code submitted to Zippopotamus.
	p := s.Zip
	switch c {
	case "CA":
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
		err = fmt.Errorf("Key: %-8s HTTP null response object on %v", s.Key, u)
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

//FixShortZips adds a leading zero to a Zip code if the country is "US",
//the postal code has four digits, and the state is one of the US states
//that has a leading zero.
func FixShortZips(s Supporter) {
	r := regexp.MustCompile(`^\d{4}$`)
	if len(s.Country) == 0 && s.Country == "US" && r.MatchString(s.Zip) {
		zeroStates := strings.Split("CT,MA,MN,NH,NJ,PR,RI,VT,VI", ",")
		for _, x := range zeroStates {
			if s.State == x {
				s.Zip = "0" + s.Zip
			}
		}
	}

}

//Zippo does a lookup using the free service from http://www.zippopotam.us/.
//Note that ambiguous results from Zippopotamus are not applied.
func Zippo(s Supporter, r []Mod) ([]Mod, error) {
	s.Country = strings.TrimSpace(s.Country)
	s.Zip = strings.TrimSpace(s.Zip)
	if len(s.Zip) == 0 {
		return r, nil
	}
	FixShortZips(s)
	m, c := MatchPostal(s)
	if m {
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
