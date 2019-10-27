//Query for a record, see JSON.
package main

import (
	"fmt"
	"log"

	"github.com/salsalabs/godig/pkg"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	cpath := kingpin.Flag("login", "YAML file containing login credentials for Salsa Classic API").PlaceHolder("FILENAME").Required().String()
	kingpin.Parse()
	api, err := godig.YAMLAuth(*cpath)
	if err != nil {
		log.Fatalf("Authentication error %v\n", err)
	}
	t := api.NewTable("groups")
	crit := ""
	count := 500
	offset := int32(0)
	for count > 0 {
		fmt.Printf("offset %d, count %d\n", offset, count)
		b, err := t.ManyMap(offset, count, crit)
		if err != nil {
			panic(err)
		}
		for i, r := range b {
			x := offset + int32(i) + 1
			fmt.Printf("%5d: %8v %-30v\n", x, r["groups_KEY"], r["Group_Name"])
			var ds godig.DeleteStatus
			t.Delete(r["groups_KEY"], &ds)
			fmt.Printf("%5d: %8v %-30v %+v\n", x, r["groups_KEY"], r["Group_Name"], ds)
		}
		count = len(b)
		offset += int32(count)
	}
}
