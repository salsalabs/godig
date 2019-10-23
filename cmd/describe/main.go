//Describe shows the structure of a table.
package main

import (
	"log"

	"github.com/salsalabs/godig/pkg"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	cpath := kingpin.Flag("login", "YAML file containing login for Salsa Classic API").PlaceHolder("FILENAME").Required().String()
	table := kingpin.Flag("table", "show description for this table").PlaceHolder("TABLE").String()
	kingpin.Parse()

	a, err := godig.YAMLAuth(*cpath)
	if err != nil {
		log.Fatalf("Authentication error: %+v\n", err)
	}

	t := a.NewTable(*table)
	var target []godig.Fields
	err = t.Describe(&target)
	if err != nil {
		log.Fatalf("Describe %v on %v\n", err, *table)
	}
	log.Printf("describe %v\n", *table)
	for _, r := range target {
		log.Printf("%+v", r)
	}
}
