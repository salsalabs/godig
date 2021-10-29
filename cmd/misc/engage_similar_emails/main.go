package main

import (
	"encoding/csv"
	"io"
	"log"
	"math"
	"os"
	"strings"
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

//Filter reads supporter information from a channel.  The email
//is compared with the previously read record.  If they are "similar"
//for some values of similarity, then the supporter records are
//written to the output channel.
func Filter(s chan map[string]string, t chan map[string]string) {
	log.Println("Filter start")
	var p map[string]string
	for {
		r, ok := <-s
		if !ok {
			break
		}
		if p == nil {
			p = r
		} else {
			if Similar(p, r) {
				t <- p
				t <- r
			}
			p = r
		}
	}
	close(t)
	log.Println("Filter done")
}

//Pump reads maps from a Reader and writes them to the channel.
//The channel is closed when the reader empties.
func Pump(r io.Reader, s chan map[string]string) {
	log.Println("Pump start")
	a, err := CSVToMap(r)
	if err != nil {
		log.Fatalf("%v converting CSV to a map", err)
	}
	for _, r := range a {
		s <- r
	}
	close(s)
	log.Println("Pump done")
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
			for k := range r {
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

//Identify accepts an email and returns a value for similarity
//testing.
func Identify(e string) string {
	a := strings.Split(e, "@")
	if len(a) > 1 {
		e = a[0]
		a = strings.Split(a[1], ".")
		if len(a) > 1 {
			e = e + a[0]
		}
	}
	return e
}

//MatchPercent accepts two strings and returns a percentage.
//The percentage is computed as the number of characters in the
//first string that match the second string dividied by the length
//of the second string.
func MatchPercent(a, b string) float64 {
	s1 := []rune(a)
	s2 := []rune(b)
	x := math.Min(float64(len(a)), float64(len(b)))
	count := 0
	for i := 0; i < int(x); i++ {
		if s1[i] == s2[i] {
			count = count + 1
		}
	}
	return float64(count) * 100.0 / float64(len(b))
}

//Similar compares two supporter records to determine if they
//are similar for certain values of similarity.  Returns true
//if they are similar, false otherwise.
func Similar(p, r map[string]string) bool {
	if p["InternalID"] != r["InternalID"] {
		n1 := Identify(p["Email"])
		n2 := Identify(r["Email"])
		pc := MatchPercent(n1, n2)
		return pc > 70.0
	}
	return false
}

//Mainline.  Find supporters and display some info about each.
//This app only works because the input file is sorted.  If it's
//not sorted on email then things will go sideways really quickly.
func main() {
	ipath := kingpin.Flag("in", "CSV file to read *sorted by email*").PlaceHolder("INPUT").Required().String()
	opath := kingpin.Flag("out", "CSV file to write").PlaceHolder("OUTPUT").Required().String()
	kingpin.Parse()

	f, err := os.Open(*ipath)
	if err != nil {
		log.Fatalf("%v on %v", err, *ipath)
	}
	var wg sync.WaitGroup
	s := make(chan map[string]string, 100)
	t := make(chan map[string]string, 100)

	go func(wg *sync.WaitGroup, s chan map[string]string, t chan map[string]string) {
		wg.Add(1)
		Filter(s, t)
		wg.Done()
	}(&wg, s, t)

	go func(wg *sync.WaitGroup, fn string, t chan map[string]string) {
		wg.Add(1)
		Save(fn, t)
		wg.Done()
	}(&wg, *opath, t)

	go func(wg *sync.WaitGroup, r io.Reader, s chan map[string]string) {
		wg.Add(1)
		Pump(f, s)
		wg.Done()
	}(&wg, f, s)

	log.Println("Main waiting")
	wg.Wait()
	log.Println("Main done")
}
