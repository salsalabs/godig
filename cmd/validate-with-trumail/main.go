package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

//CSVToMap accepts a Reader and returns an array of maps.
//Many thanks to https://gist.github.com/drernie/5684f9def5bee832ebc50cabb46c377a
func CSVToMap(reader io.Reader) ([]map[string]string, error) {
	r := csv.NewReader(reader)
	rows := []map[string]string{}
	var header []string
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return rows, err
		}
		if header == nil {
			header = record
		} else {
			dict := map[string]string{}
			for i := range header {
				dict[header[i]] = record[i]
			}
			rows = append(rows, dict)
		}
	}
	return rows, nil
}

//Save reads supporters from a queue and writes them to an output
//CSV file.
func Save(fn string, i chan map[string]string) {
	log.Println("Save start")
	f, err := os.Create(fn)
	if err != nil {
		log.Fatalf("%v on %v", err, fn)
	}
	w := csv.NewWriter(f)
	var h []string
	for {
		r, ok := <-i
		if !ok {
			break
		}
		if h == nil {
			for k, _ := range r {
				h = append(h, k)
			}
			w.Write(h)
		}
		var a []string
		for _, s := range h {
			a = append(a, r[s])
		}
		err := w.Write(a)
		if err != nil {
			log.Fatalf("%v on %v", err, fn)
		}
	}
	w.Flush()
	log.Println("Save done")
}

//Pump reads maps from a Reader and writes them to the channel.
//The channel is closed when the reader empties.
func Pump(r io.Reader, c chan map[string]string) {
	log.Println("Pump start")
	a, err := CSVToMap(r)
	if err != nil {
		log.Fatalf("%v converting CSV to a map", err)
	}
	for _, r := range a {
		c <- r
	}
	close(c)
	log.Println("Pump done")
}

//Mainline.  Find supporters and display some info about each.
func main() {
	ipath := kingpin.Flag("csv", "CSV file to read").PlaceHolder("INPUT").Required().String()
	opath := kingpin.Flag("out", "CSV file to write").PlaceHolder("OUTPUT").Required().String()
	kingpin.Parse()

	f, err := os.Open(*ipath)
	if err != nil {
		log.Fatalf("%v on %v", err, *ipath)
	}
	var wg sync.WaitGroup
	c := make(chan map[string]string, 100)

	go func(wg *sync.WaitGroup, r io.Reader, c chan map[string]string) {
		wg.Add(1)
		Pump(f, c)
		wg.Done()
	}(&wg, f, c)

	go func(wg *sync.WaitGroup, fn string, c chan map[string]string) {
		wg.Add(1)
		Save(fn, c)
		wg.Done()
	}(&wg, *opath, c)

	fmt.Println("Main waiting")
	wg.Wait()
	fmt.Println("Main done")
}
