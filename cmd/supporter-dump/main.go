//An application to accept a table and and create a Go file containing
//a table schema.  The table schema can be used to retrieve data from
//Salsa for the table.
package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"sync"

	"github.com/salsalabs/godig"
	"github.com/tidwall/gjson"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

func push(t *godig.Table, crit string, offsets chan int32) {
	c, err := t.Count(crit)
	if err != nil {
		panic(err)
	}
	x, _ := strconv.ParseInt(c, 10, 32)
	x = ((x + 499) / 500) * 500
	n := int32(x)
	fmt.Printf("Push: actual count %v, modified count %v\n", c, n)

	for i := int32(0); i < n; i += 500 {
		offsets <- i
		fmt.Printf("Push: pushed %7d\n", i)
	}
}

func write(t *godig.Table, crit string, offsets chan int32, id int) {
	fn := fmt.Sprintf("./extract/supporter_%02d.csv", id)
	d := path.Dir(fn)
	err := os.MkdirAll(d, os.ModePerm)
	if err != nil {
		panic(err)
	}
	s, err := os.Create(fn)
	if err != nil {
		panic(err)
	}
	w := csv.NewWriter(s)

	for {
		off, ok := <-offsets
		log.Printf("Writer-%02d: offset %7d\n", id, off)
		if !ok {
			log.Printf("Writer-%02d: done\n", id)
			return
		}
		f, err := t.ManyMap(off, 500, crit)
		if err != nil {
			panic(err)
		}

		var h []string
		needHeaders := true
		var records [][]string

		for _, m := range f {
			var record []string
			m.ForEach(func(key, value gjson.Result) bool {
				if needHeaders {
					h = append(h, key.String())
				}
				record = append(record, value.String())
				return true // keep iterating
			})
			records = append(records, record)
		}
		if needHeaders {
			needHeaders = false
			err = w.Write(h)
			if err != nil {
				panic(err)
			}
		}
		err = w.WriteAll(records)
		if err != nil {
			panic(err)
		}
		w.Flush()
		log.Printf("Writer-%02d: offset %7d, wrote %d records to %v\n", id, off, len(records), fn)
	}
}
func main() {
	defCrit := "EMAIL%20IS%20NOT%20EMPTY&condition=First_Name%20IS%20NOT%20EMPTY&condition=Receive_Email>0"
	var (
		login = kingpin.Flag("login", "YAML file with login credentials").Required().String()
		crit  = kingpin.Flag("criteria", "Use this criteria (without leading &condition").Default(defCrit).String()
	)
	kingpin.Parse()
	api, err := (godig.YAMLAuth(*login))
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	offsets := make(chan int32)
	t := api.NewTable("supporter")
	var wg sync.WaitGroup

	for i := 0; i < 30; i++ {
		go (func(wg *sync.WaitGroup, t *godig.Table, crit string, offsets chan int32, id int) {
			wg.Add(1)
			write(t, crit, offsets, id)
			wg.Done()
		})(&wg, &t, *crit, offsets, i)
	}
	go (func(wg *sync.WaitGroup, t *godig.Table, crit string, offsets chan int32) {
		wg.Add(1)
		push(t, crit, offsets)
		wg.Done()

	})(&wg, &t, *crit, offsets)
	wg.Wait()
}
