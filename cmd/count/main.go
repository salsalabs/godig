//Count the number of records for a table name.
package main

import (
	"log"

	"github.com/salsalabs/godig/pkg"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	cpath := kingpin.Flag("credentials", "YAML file containing credentials for Salsa Classic API").PlaceHolder("FILENAME").Required().String()
	table := kingpin.Flag("table", "table name ([supporter], donation, groups, etc.)").PlaceHolder("TABLE").Default("supporter").String()
	cond := kingpin.Flag("criteria", "(Optional) Salsa-formatted API condition").PlaceHolder("CONDITION").String()
	kingpin.Parse()
	api, err := godig.YAMLAuth(*cpath)
	if err != nil {
		log.Fatalf("Authentication error %v\n", err)
	}
	s := api.NewTable(*table)
	x, err := s.Count(*cond)
	if err != nil {
		log.Fatalf("Count error %v\n", err)
	}
	log.Printf("Count results %v\n", x)
}
