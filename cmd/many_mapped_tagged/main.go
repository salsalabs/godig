//Query for a record, see JSON.
package main

import (
	"fmt"
	"log"

	"github.com/salsalabs/godig"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	cpath := kingpin.Flag("login", "YAML file containing login credentials for Salsa Classic API").PlaceHolder("FILENAME").Required().String()
	table := kingpin.Flag("table", "table name ([supporter], donation, groups, etc.)").PlaceHolder("TABLE").Required().String()
	tag := kingpin.Flag("tag", "return records with this tag").PlaceHolder("TAG").Required().String()
	kingpin.Parse()
	api, err := godig.YAMLAuth(*cpath)
	if err != nil {
		log.Fatalf("Authentication error %v\n", err)
	}
	t := api.NewTable(*table)
	_ = []string{
		"EMAIL IS NOT EMPTY",
		"EMAIL LIKE %@%.%",
		"Receive_Email>0",
	}
	crit := "" //strings.Join(c, "&condition=")
	count := 500
	offset := int32(0)
	for count > 0 {
		b, err := t.ManyMapTagged(offset, count, crit, *tag)
		if err != nil {
			panic(err)
		}
		for i, r := range b {
			x := offset + int32(i) + 1
			fmt.Printf("%5d: %8v %-30v\n", x, r["supporter_KEY"], r["Email"])
		}
		count := len(b)
		offset += int32(count)
	}
}
