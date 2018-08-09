//An application to accept a table and and create a Go file containing
//a table schema.  The table schema can be used to retrieve data from
//Salsa for the table.
package main

import (
	"log"

	"github.com/salsalabs/godig"
	"github.com/tidwall/gjson"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	var (
		login = kingpin.Flag("login", "YAML file with login credentials").Required().String()
		table = kingpin.Flag("table", "Test with this table").Required().String()
		crit  = kingpin.Flag("criteria", "Use this criteria (without leading &condition").Default("").String()
		//match = kingpin.Flag("match", "GoJSON match string").Default("#Email#").String()
	)
	kingpin.Parse()
	api, err := (godig.YAMLAuth(*login))
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	t := api.NewTable(*table)
	b, err := t.ManyRaw(0, 10, *crit)
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	//result := gjson.GetBytes(b, *match)
	//log.Printf("results\n%+v\n", result)

	result := gjson.ParseBytes(b)
	result.ForEach(func(key, value gjson.Result) bool {
		//e := value.Get("Email")
		//log.Printf("Email: %v\n", e)
		m := value.Map()
		log.Printf("map\n%v, %v, %v, %v\n", m["supporter_KEY"], m["Email"], m["First_Name"], m["Last_Name"])
		return true // keep iterating
	})
}
