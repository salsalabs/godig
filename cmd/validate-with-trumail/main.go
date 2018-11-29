package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

//TruMail is returned by trumail.io when an email is submitted.
type TruMail struct {
	Address     string `json:"address"`
	Username    string `json:"username"`
	Domain      string `json:"domain"`
	Md5Hash     string `json:"md5Hash"`
	Suggestion  string `json:"suggestion"`
	ValidFormat string `json:"validFormat"`
	Deliverable string `json:"deliverable"`
	FullInbox   string `json:"fullInbox"`
	HostExists  string `json:"hostExists"`
	CatchAll    string `json:"catchAll"`
	Gravatar    string `json:"gravatar"`
	Role        string `json:"role"`
	Disposable  string `json:"disposable"`
	Free        string `json:"free"`
}

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

//Lookup sreads supporters from the source channel and looks up the
//Email address. The supporter record is augmented with the results
//of the lookup.  Lookup failures append empty fields to the supporter
//records.  Records are written to the target channel.  The target
//channel is closed when there is no mmore data on the source channel.
func Lookup(s chan map[string]string, t chan map[string]string) {
	trumail := "https://api.trumail.io/v2/lookups/json?email=%s"
	c := http.Client{
		Timeout: time.Second * 5,
	}
	for {
		r, ok := <-s
		if !ok {
			break
		}
		u := fmt.Sprintf(trumail, r["Email"])
		validFormat := ""
		deliverable := ""
		var err error
		req, err := http.NewRequest(http.MethodGet, u, nil)
		if err == nil {
			req.Header.Set("User-Agent", "salsalabs-lookup")
			res, err := c.Do(req)
			if err == nil {
				b, err := ioutil.ReadAll(res.Body)
				if err == nil {
					tr := TruMail{}
					err := json.Unmarshal(b, &tr)
					if err == nil {
						validFormat = tr.ValidFormat
						deliverable = tr.Deliverable
					}
				}
			}
		}
		if err != nil {
			log.Printf("%v on %v", err, u)
		}
		r["ValidFormat"] = validFormat
		r["Deliverable"] = deliverable
		t <- r
	}

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
	s := make(chan map[string]string, 100)
	t := make(chan map[string]string, 100)

	go func(wg *sync.WaitGroup, s chan map[string]string, t chan map[string]string) {
		wg.Add(1)
		Lookup(s, t)
		wg.Done()
	}(&wg, s, t)

	go func(wg *sync.WaitGroup, fn string, t chan map[string]string) {
		wg.Add(1)
		Save(fn, s)
		wg.Done()
	}(&wg, *opath, s)

	go func(wg *sync.WaitGroup, r io.Reader, s chan map[string]string) {
		wg.Add(1)
		Pump(f, s)
		wg.Done()
	}(&wg, f, s)

	log.Println("Main waiting")
	wg.Wait()
	log.Println("Main done")
}
