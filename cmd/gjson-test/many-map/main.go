//An application to accept a table and and create a Go file containing
//a table schema.  The table schema can be used to retrieve data from
//Salsa for the table.
package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/salsalabs/godig"
	"github.com/tidwall/gjson"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

func push(t *godig.Table, m chan int32) {

}
func main() {
	defCrit := "EMAIL%20IS%20NOT%20EMPTY&condition=First_Name%20IS%20NOT%20EMPTY&condition=Receive_Email>0"
	var (
		login = kingpin.Flag("login", "YAML file with login credentials").Required().String()
		table = kingpin.Flag("table", "Test with this table").Required().String()
		crit  = kingpin.Flag("criteria", "Use this criteria (without leading &condition").Default(defCrit).String()
	)
	kingpin.Parse()
	api, err := (godig.YAMLAuth(*login))
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	var headers []string
	var records [][]string
	t := api.NewTable(*table)
	c, err := t.Count(*crit)
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	m, _ := strconv.ParseInt(c, 10, 32)
	m = ((m + 499) / 500) * 500
	n := int32(m)
	fmt.Printf("Actual count %v, modified count %v\n", c, n)
	f, err := t.ManyMap(0, 500, *crit)
	for _, m := range f {
		var record []string
		var h []string
		log.Printf("\n")
		m.ForEach(func(key, value gjson.Result) bool {
			if len(headers) == 0 {
				h = append(h, key.String())
			}
			record = append(record, value.String())
			return true // keep iterating
		})
		records = append(records, record)
		if len(headers) == 0 {
			headers = h
		}
	}
	s, err := os.Create("test.csv")
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	w := csv.NewWriter(s)
	err = w.Write(headers)
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	err = w.WriteAll(records)
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	log.Printf("Wrote %d records to %v\n", len(records), "test.csv")
}
