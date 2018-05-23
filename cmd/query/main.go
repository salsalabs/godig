//Query for a record, see JSON.
package main

import (
	"log"
	"strings"

	"github.com/salsalabs/godig"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	cpath := kingpin.Flag("credentials", "YAML file containing credentials for Salsa Classic API").PlaceHolder("FILENAME").Required().String()
	table := kingpin.Flag("table", "table name ([supporter], donation, groups, etc.)").PlaceHolder("TABLE").Default("supporter").String()
	key := kingpin.Flag("key", "primary key").PlaceHolder("KEY").Required().String()
	kingpin.Parse()
	api, err := godig.YAMLAuth(*cpath)
	if err != nil {
		log.Fatalf("Authentication error %v\n", err)
	}
	t := api.NewTable(*table)
	b, err := t.OneRaw(*key)
	if err != nil {
		log.Fatalf("Query error %v\n", err)
	}
	s := strings.Replace(string(b), ",", ",\n", -1)
	log.Printf("%v\n", s)
}
