package main

import (
	"encoding/csv"
	"log"
	"os"
	"sync"

	godig "github.com/salsalabs/godig/pkg"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

//Page is the contents that we're retrieving from each donation page.
type DonatePage struct {
	DonatePageKey   string `json:"donate_page_KEY"`
	OrganizationKey string `json:"organization_KEY"`
	ChapterKey      string `json:"chapter_KEY"`
	ReferenceName   string `json:"Reference_Name"`
	DateCreated     string `json:"Date_Created"`
}

//driver reads donation pages and pushes them onto a channel.
//The channel is closed at end of data.

func Generate(a *godig.API, c chan DonatePage, initOffset int32) error {
	t := a.NewTable("donate_page")
	offset := initOffset
	if initOffset != int32(0) {
		log.Printf("Generate: starting read at offset %d\n", initOffset)
	}
	count := 500
	for count > 0 {
		log.Printf("Generate: %v offset %6d\n", t.Name, offset)
		if offset > 0 && offset%5000 == 0 {
			log.Printf("Generate: %v offset %6d\n", t.Name, offset)
		}
		var a []DonatePage
		err := t.Many(offset, count, "", &a)
		if err != nil {
			log.Fatalf("Generate: %v offset %6d %v\n", t.Name, offset, err)
			return err
		}
		count = len(a)
		if count == 0 {
			log.Printf("Generate: %v offset %6d, done\n", t.Name, offset)
			return err
		}
		for _, r := range a {
			c <- r
		}
		offset = offset + int32(count)
	}
	close(c)
	return nil
}

//Consume accepts donate page records from a channel and writes
//them to `donate_pages.csv`.
func Consume(c chan DonatePage) error {
	csvFile := "donate_pages.csv"
	f, err := os.Create(csvFile)
	if err != nil {
		log.Fatalf("%v, %v\n", err, csvFile)
		return err
	}
	defer f.Close()
	writer := csv.NewWriter(f)
	first := true
	for {
		r, okay := <-c
		if !okay {
			break
		}
		if first {
			headers := []string{
				"DonatePageKey",
				"OrganizationKey",
				"ChapterKey",
				"ReferenceName",
				"DateCreated",
			}
			err := writer.Write(headers)
			if err != nil {
				log.Fatalf("%v, %v\n", err, csvFile)
				return err
			}
			first = false
		}
		row := []string{
			r.DonatePageKey,
			r.OrganizationKey,
			r.ChapterKey,
			r.ReferenceName,
			godig.ShortDate(r.DateCreated),
		}
		err := writer.Write(row)
		if err != nil {
			log.Fatalf("%v, %v\n", err, csvFile)
			return err
		}
	}
	writer.Flush()
	return nil
}

func main() {
	var (
		cpath      = kingpin.Flag("login", "YAML file of credentials for Salsa Classic API").PlaceHolder("FILENAME").Required().String()
		offset     = kingpin.Flag("offset", "Start reading at this offset").PlaceHolder("OFFSET").Default("0").Int32()
		apiVerbose = kingpin.Flag("verbose", "Show requests to, and responses from, the server. Can be very noisy.").Bool()
	)
	kingpin.Parse()
	a, err := godig.YAMLAuth(*cpath)
	if err != nil {
		log.Fatalf("Authentication error: %+v\n", err)
	}
	a.Verbose = *apiVerbose
	c := make(chan DonatePage, 100)
	var w sync.WaitGroup

	go func(c chan DonatePage, w *sync.WaitGroup) {
		defer w.Done()
		w.Add(1)
		err := Consume(c)
		if err != nil {
			panic(err)
		}
	}(c, &w)
	log.Println("main: Consume started")

	go func(a *godig.API, c chan DonatePage, w *sync.WaitGroup) {
		defer w.Done()
		w.Add(1)
		Generate(a, c, *offset)
	}(a, c, &w)

	log.Println("main: Generate started")
	log.Println("main: Waiting for terminations...")
	w.Wait()
	log.Println("main: done")
}
