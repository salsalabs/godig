//Count the number of records for a table name.
package main

import (
	"encoding/csv"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/salsalabs/godig"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

//Criteria is the read conditions for fetching email blasts from Classic.
const Criteria = "Stage=Complete"

//CSVFilename is where the blast content ends up.
const CSVFilename = "./blast_content.csv"

//Headers are the CSV headers for EmailBlast records.
const Headers = "SupporterKey,Date,Subjject,HTML,Text"

//EmailBlast is the content that this app will read from Salsa Classic.
type EmailBlast struct {
	EmailBlastKey string `json:"email_blast_KEY"`
	DateRequested string `json:"Date_Requested"`
	Subject       string `json:"Subject"`
	HTMLContent   string `json:"HTML_Content"`
	TextContent   string `json:"Text_Content"`
}

//Push counts the number of records to read, then pushes
//read offsets into a channel.
func Push(t *godig.Table, c chan int32) error {
	log.Println("Push start")
	x, err := t.Count(Criteria)
	if err != nil {
		return err
	}
	y, err := strconv.ParseInt(x, 10, 32)
	if err != nil {
		return err
	}
	log.Printf("Push %d records\n", y)
	max := int32(y) + 500
	for i := int32(0); i < max; i += 500 {
		c <- i
	}
	close(c)
	log.Println("Push done")
	return nil
}

//Fetch accepts offsets from a channel, reads emal blast records,
//then pushes them onto a channel.  A true is pushed onto the done
//channel when the offset channel is closed.
func Fetch(t *godig.Table, c chan int32, e chan EmailBlast, d chan bool) error {
	log.Println("Fetch start")
	for {
		x, ok := <-c
		log.Printf("Fetch popped %v, %v\n", x, ok)
		if !ok {
			break
		}
		var a []EmailBlast
		err := t.Many(x, 500, Criteria, &a)
		if err != nil {
			return err
		}
		log.Printf("Fetch returned %d records\n", len(a))
		for _, r := range a {
			e <- r
		}
	}
	d <- true
	log.Println("Fetch done")
	return nil
}

//Store reads email blasts from the blast channel and writes them to a
//CSV file.  CSV filename is hard coded.
func Store(e chan EmailBlast) error {
	log.Println("Store start")
	f, err := os.Create(CSVFilename)
	if err != nil {
		return err
	}
	w := csv.NewWriter(f)
	h := strings.Split(Headers, ",")
	w.Write(h)

	for {
		r, ok := <-e
		if !ok {
			break
		}
		d := godig.EngageDate(r.DateRequested)
		a := []string{
			r.EmailBlastKey,
			d,
			r.Subject,
			r.HTMLContent,
			r.TextContent,
		}
		w.Write(a)
		log.Printf("%-8s %v %v\n", r.EmailBlastKey, d, r.Subject)
		w.Flush()
	}
	w.Flush()
	err = f.Close()
	return err
}

//WaitFor waits for 'n' messages to appear on the done channel.
//When that happens, WaitFor closes the blast channel.
func WaitFor(d chan bool, n int, e chan EmailBlast) {
	log.Printf("WaitFor start")
	for n > 0 {
		_, ok := <-d
		if !ok {
			break
		}
		n--
		log.Printf("WaitFor waiting for %d\n", n)
	}
	close(e)
	log.Println("WaitFor done")
}

//main is the main program.
func main() {
	login := kingpin.Flag("login", "YAML file containing Salsa Classic credentials").Required().String()
	kingpin.Parse()
	api, err := godig.YAMLAuth(*login)
	if err != nil {
		log.Fatalf("Authentication error %v\n", err)
	}
	t := api.NewTable("email_blast")
	c := make(chan int32, 500)
	d := make(chan bool, 10)
	e := make(chan EmailBlast, 500)
	var w sync.WaitGroup
	log.Println("Main start")

	go (func(e chan EmailBlast, w *sync.WaitGroup) {
		w.Add(1)
		err := Store(e)
		if err != nil {
			panic(err)
		}
		w.Done()
	})(e, &w)

	go (func(d chan bool, n int, e chan EmailBlast, w *sync.WaitGroup) {
		w.Add(1)
		WaitFor(d, n, e)
		w.Done()
	})(d, 5, e, &w)

	for i := 0; i < 5; i++ {
		go (func(t *godig.Table, c chan int32, e chan EmailBlast, d chan bool, w *sync.WaitGroup) {
			w.Add(1)
			err := Fetch(t, c, e, d)
			if err != nil {
				panic(err)
			}
			w.Done()
		})(&t, c, e, d, &w)
	}

	go (func(t *godig.Table, c chan int32, w *sync.WaitGroup) {
		w.Add(1)
		err := Push(t, c)
		if err != nil {
			panic(err)
		}
		w.Done()
	})(&t, c, &w)

	//Settle time then wait for things to end.
	time.Sleep(10000)
	w.Wait()
	log.Println("Main done")
}
