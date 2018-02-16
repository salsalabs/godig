// Demonstrates reading CSV and TDF files.  Quoting items on output doesn't work the way that we want.
package main

import (
	"encoding/csv"
	"log"
	"os"
)

func main() {
	// os.Open is O_RDONLY.
	r, err := os.Open("export.txt")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer r.Close()
	reader := csv.NewReader(r)
	reader.Comma = '\011'
	list, err := reader.ReadAll()
	log.Println(list)

	for i, row := range list {
		log.Printf("%v: %v  contains %v items\n", i, row, len(row))
		for j, v := range row {
			row[j] = `"` + v + `"`
		}
		log.Printf("%v: %v  contains %v items\n", i, row, len(row))
	}
	w, err := os.OpenFile("copy-of-export.txt", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer w.Close()
	writer := csv.NewWriter(w)
	writer.Comma = '\011'
	err = writer.WriteAll(list)
	if err != nil {
		log.Fatal(err)
		return
	}
}
