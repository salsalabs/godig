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
	Abbr  string `json:"state_abbreviation"`
	Lat   string `json:"latitude"`
}

//Result is return by Zippotamus for all places that match a
//ZIP/postal code.
type ZResult struct {
	PostCode    string `json:"post code"`
	Country     string `json:"country"`
	CountryCode string `json:"contry abbreviation"`
	Places      []ZPlace
}

//Zippo does a lookup using the free service from http://www.zippopotam.us/.
//Note that ambiguous results from Zippopotamus are not applied.
func Zippo(s Supporter, r []Mod) error {
	short := regexp.MustCompile("\\d{}5}").MatchString(s.Zip)
	long := regexp.MustCompile("\\d{5}-\\d{4}").MatchString(s.Zip)

	if len(s.Zip) == 0 || !short || !long {
		log.Printf("Zippo: Key %v, Zip %v skipped\n", s.Key, s.Zip)
		return nil
	}

	u := fmt.Sprintf("http://api.zippopotam.us/us/%v", s.Zip)
	var body []byte
	var target ZResult
	resp, err := http.Get(u)
	if err == nil {

		defer resp.Body.Close()
		body, err = ioutil.ReadAll(resp.Body)
		if err == nil {
			err = json.Unmarshal(body, &target)
			if err == nil {
				log.Printf("Zippo: Result is %+v\n", target)
			}
		}
	}

	return err
}
