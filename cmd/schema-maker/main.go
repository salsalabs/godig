//An application to accept a table and and create a Go file containing
//a table schema.  The table schema can be used to retrieve data from
//Salsa for the table.
package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/salsalabs/godig"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

//Item is a single description item.
type Item struct {
	Name         string
	Nullable     string
	Type         string
	DefaultValue string
	Label        string
}

func main() {
	var (
		login = kingpin.Flag("login", "YAML file with login credentials").Required().String()
		table = kingpin.Flag("table", "Describe this table").Required().String()
	)
	kingpin.Parse()
	api, err := (godig.YAMLAuth(*login))
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	t := api.NewTable(*table)
	var a []Item
	err = t.Describe(&a)
	if err != nil {
		panic(err)
	}
	b, err := json.MarshalIndent(a, "", "\t")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}
