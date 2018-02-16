//List gathers a list of groups asynchronously then displays it.
package main

import (
	"encoding/json"
	"fmt"
	"github.com/salsalabs/godig"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
	"log"
)

//Group contains data retrieved from the groups table.
type Group struct {
	Key  string `json:"groups_KEY"`
	Name string `json:"Group_Name"`
}

//allGroups returns an array of group object filled from the site.
func allGroups(a *godig.API, t []Group, done chan bool) {
	c1 := make(chan []byte, 1000)
	go func(c1 chan []byte) {
		for b := range c1 {
			var target []Group
			json.Unmarshal(b, &target)
			for _, r := range target {
				t = append(t, r)
			}
		}
	}(c1)

	log.Printf("starting All")
	g := a.NewTable("groups")
	All(g, c1, done)
}

//All reads from Salsa via the API.
func All(t godig.Table, c1 chan []byte, done chan bool) {
	templ := "https://%v/api/getObjects.sjs?json&object=%v"
	u := fmt.Sprintf(templ, t.Host, t.Name)

	// Read data and process it.
	offset := 0
	count := 500
	for count > 0 {
		if offset > 0 && offset%5000 == 0 {
			log.Printf("Offset: %6d\n", offset)
		}
		t.Raw(u, offset, count, c1, done)
		offset = offset + 500
	}
}

func main() {
	cpath := kingpin.Flag("credentials", "YAML file containing credentials for Salsa Classic API").PlaceHolder("FILENAME").Required().String()
	kingpin.Parse()

	a, err := godig.YAMLAuth(*cpath)
	if err != nil {
		log.Fatalf("Authentication error: %+v\n", err)
	}
	var t []Group
	done := make(chan bool)
	go allGroups(a, t, done)
	<-done
	for i, r := range t {
		log.Printf("%2d: %v %v", i, r.Key, r.Name)
	}
}
