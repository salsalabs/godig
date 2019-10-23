//Save a value.
package main

import (
	"log"

	"github.com/salsalabs/godig/pkg"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	cpath := kingpin.Flag("credentials", "YAML file containing credentials for Salsa Classic API").PlaceHolder("FILENAME").Required().String()
	table := kingpin.Flag("table", "table name ([supporter], donation, groups, etc.)").PlaceHolder("TABLE").Default("supporter").String()
	key := kingpin.Flag("key", "primary key").PlaceHolder("KEY").Required().String()
	cond := kingpin.Flag("conditions", "(Optional) Salsa-formatted API condition").PlaceHolder("CONDITION").String()
	kingpin.Parse()
	api, err := godig.YAMLAuth(*cpath)
	if err != nil {
		log.Fatalf("Authentication error %v\n", err)
	}
	t := api.NewTable(*table)
	_, err = t.Save(*key, *cond)
	if err != nil {
		log.Fatalf("Save error %v\n", err)
	}
	log.Printf("Saved")
}
