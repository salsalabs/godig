//Describe shows the structure of a table.
package main

import (
	"log"

	"github.com/salsalabs/godig"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

//Fields are the fields returned for each database field.
type Fields struct {
	Name         string
	Nullable     string
	Type         string
	DefaultValue string `json:"defaultValue"`
	Label        string
}

func main() {
	cpath := kingpin.Flag("credentials", "YAML file containing credentials for Salsa Classic API").PlaceHolder("FILENAME").Required().String()
	table := kingpin.Flag("table", "show description for this table").PlaceHolder("TABLE").String()
	kingpin.Parse()

	a, err := godig.YAMLAuth(*cpath)
	if err != nil {
		log.Fatalf("Authentication error: %+v\n", err)
	}

	t := a.NewTable(*table)
	var target []Fields
	err = t.Describe(&target)
	if err != nil {
		log.Fatalf("Describe %v on %v\n", err, *table)
	}
	log.Printf("describe %v\n", *table)
	for _, r := range target {
		log.Printf("%+v", r)
	}
}
