package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/badoux/checkmail"
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

//Lookup reads supporter information from a channel. The email
//address' domain name is extracted and looked up.  The lookup results
//are appended to the record and it's put into the target channel.
func Lookup(s chan map[string]string, t chan map[string]string, done chan bool) {
	log.Println("Lookup start")
	for {
		r, ok := <-s
		if !ok {
			break
		}
		hasError := false
		m := ""
		err := checkmail.ValidateFormat(r["Email"])
		if err != nil {
			m = fmt.Sprintf("%v", err)
			hasError = true
		} else {
			m = ""
		}
		r["FormatErr"] = m

		err = checkmail.ValidateHost(r["Email"])
		if err != nil {
			m = fmt.Sprintf("%v", err)
			hasError = true
		} else {
			m = ""
		}
		r["HostErr"] = m
		err = checkmail.ValidateHost(r["Email"])
		if smtpErr, ok := err.(checkmail.SmtpError); ok && err != nil {
			m = fmt.Sprintf("%s", smtpErr.Code())
			hasError = true
			mp := strings.Split(m, ". ")
			if len(mp) > 1 {
				m = mp[0]
			}
		} else {
			m = ""
		}
		r["UserErr"] = m
		if hasError {
			t <- r
		}
	}
	done <- true
	log.Println("Lookup done")
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
		return a[1]
	}
	return e
}

//WaitFor waits for 'n' messages to appear on the 'done' channel.
//When all messages have arrived, WaitFor closed the target channel.
func WaitFor(n int, done chan bool, target chan map[string]string) {
	log.Printf("WaitFor start")
	for {
		_, ok := <-done
		if !ok {
			break
		}
		n--
		if n <= 0 {
			break
		}
		log.Printf("WaitFor waiting for %d\n", n)
	}
	close(target)
	log.Printf("WaitFor done")
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
	done := make(chan bool)
	count := 10

	for i := 0; i < count; i++ {
		go func(wg *sync.WaitGroup, s chan map[string]string, t chan map[string]string, d chan bool) {
			wg.Add(1)
			Lookup(s, t, d)
			wg.Done()
		}(&wg, s, t, done)
	}

	go func(wg *sync.WaitGroup, n int, d chan bool, t chan map[string]string) {
		wg.Add(1)
		WaitFor(n, d, t)
	}(&wg, count, done, t)

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
