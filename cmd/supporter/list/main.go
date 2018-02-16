package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/salsalabs/godig"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

//Fields are retrieved from the supporter record.
type Fields struct {
	SupporterKey string `json:"supporter_KEY"`
	FirstName    string `json:"First_Name"`
	LastName     string `json:"Last_Name"`
	Email        string `json:"Email"`
}

//Table is the table name to read.
const Table = "supporter"

//GetObjects is the template used to retrive records.
const GetObjects = "https://%s/api/getObjects.sjs?json&object=%s&include=First_Name,Last_Name,Email"

//Do accepts Fields records from a channel and displays them.
func Do(cin chan Fields) {
	for f := range cin {
		log.Printf("%s %s %s %s", f.SupporterKey, f.FirstName, f.LastName, f.Email)
	}
}

//Parse accepts a channel for a buffer and sends fields on another channel.
func Parse(cin chan []byte, cout chan Fields, done chan bool) {
	for body := range cin {
		var target []Fields
		err := json.Unmarshal(body, &target)
		if err != nil {
			log.Fatalf("Parse: JSON unmarshall error %v\n", err)
		}
		if len(target) == 0 {
			done <- true
			return
		}
		for _, f := range target {
			cout <- f
		}
	}
}

//Fetch reads from Salsa via the API.  Each read puts a byte buffer
//onto the provided channel.
func Fetch(t *godig.Table, c1 chan []byte) {
	u := fmt.Sprintf(GetObjects, t.Host, Table)

	// Read data and process it.
	offset := 0
	count := 500
	done := make(chan bool)
	for count > 0 {
		if offset > 0 && offset%5000 == 0 {
			log.Printf("Offset: %6d\n", offset)
		}
		t.Raw(u, offset, count, c1, done)
		offset = offset + 500
	}
	<- done
	log.Print("Done.")
}

//Mainline.  Find supporters and display some info about each.
func main() {
	cpath := kingpin.Flag("credentials", "YAML file containing credentials for Salsa Classic API").PlaceHolder("FILENAME").Required().String()
	kingpin.Parse()

	a, err := godig.YAMLAuth(*cpath)
	if err != nil {
		log.Fatalf("Authentication error: %+v\n", err)
	}

	c1 := make(chan []byte)
	c2 := make(chan Fields)
	done := make(chan bool)

	go Do(c2)
	go Parse(c1, c2, done)
	t := a.Supporter()
	go Fetch(&t, c1)
	<-done
}
