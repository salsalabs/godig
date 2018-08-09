package main

import (
	"fmt"
	"log"

	"github.com/salsalabs/godig"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	var (
		login = kingpin.Flag("login", "YAML file with login credentials").Required().String()
		table = kingpin.Flag("table", "Get database_table record for this table").Default("supporter").String()
	)
	kingpin.Parse()
	api, err := (godig.YAMLAuth(*login))
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	dt := api.NewTable("database_table")
	var a []godig.DatabaseTable
	crit := fmt.Sprintf("table_name=%v\n", *table)
	err = dt.Many(int32(0), 500, crit, &a)
	if err != nil {
		panic(err)
	}
	for _, d := range a {
		fmt.Printf("%+v\n", d)
	}
}
