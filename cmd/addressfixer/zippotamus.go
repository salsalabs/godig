package addressfixer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
)

//Place is returned by Zippopotamus for a match with the zipcode.
type ZPlace struct {
	Name  string `json:"place_name"`
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
	p := `^\d{5}(?:[-\s]\d{4})?$`
	m := regexp.MustCompile(p).MatchString(z)
	return m
}

//isUS returns true if the provided postal code is a ZIP code.  Note
//that other countries also use five numeric digits for postal codes.
//We are ignoring the ambiguity for the time being.
func isUS(z string) bool {
	p := `^\d{5}(?:[-\s]\d{4})?$`
	m := regexp.MustCompile(p).MatchString(z)
	return m
}

//Zippo does a lookup using the free service from http://www.zippopotam.us/.
//Note that ambiguous results from Zippopotamus are not applied.
func Zippo(s Supporter, r []Mod) error {
	if len(s.Zip) == 0 {
		log.Printf("Zippo:   Key %v, Zip is empty\n", s.Key)
		return nil
	}
	country := "US"
	if !isUS(s.Zip) {
		if isCA(s.Zip) {
			log.Printf("Zippo:   Zip %v is CA\n", s.Zip)
			country = "CA"
		} else {
			log.Printf("Zippo:   Zip %v, is not US or CA\n", s.Zip)
		}
	}
	u := fmt.Sprintf("http://api.zippopotam.us/%v/%v", country, s.Zip)
	log.Printf("Zippo:   Reading %v\n", u)
	var body []byte
	var target ZResult
	resp, err := http.Get(u)
	if resp.StatusCode != 200 {
		err = fmt.Errorf("HTTP error %v on %v", err, u)
	}
	if err == nil {
		defer resp.Body.Close()
		body, err = ioutil.ReadAll(resp.Body)
		if err == nil {
			err = json.Unmarshal(body, &target)
			if err == nil {
				log.Printf("Zippo:   Result is %+v\n", target)
			}
		}
	}
	return err
}
