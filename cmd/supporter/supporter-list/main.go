package main

import (
	"log"
	"sync"

	godig "github.com/salsalabs/godig/pkg"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

//Fields are retrieved from the supporter record.
type Fields struct {
	SupporterKey string `json:"supporter_KEY"`
	FirstName    string `json:"First_Name,omitempty"`
	LastName     string `json:"Last_Name,omitempty"`
	Email        string `json:"Email,omitempty"`
}

//All reads from Salsa via the API.  If the criteria is not empty,
//then records that match that criteria are returned.  Each read
//parses the buffer for records then outputs them to cout.
func All(t *godig.Table, crit string, cout chan Fields) {
	offset := int32(0)
	count := 500
	for count > 0 {
		log.Printf("All: %v offset %6d\n", t.Name, offset)
		if offset > 0 && offset%5000 == 0 {
			log.Printf("All: %v offset %6d\n", t.Name, offset)
		}
		var a []Fields
		err := t.Many(offset, count, crit, &a)
		if err != nil {
			log.Fatalf("All: %v offset %6d %v\n", t.Name, offset, err)
			close(cout)
			return
		}
		count = len(a)
		if count == 0 {
			log.Printf("All: %v offset %6d, done\n", t.Name, offset)
			close(cout)
			return
		}
		for _, r := range a {
			cout <- r
		}
		offset = offset + int32(count)
	}
}

//Use accepts Fields records from a channel and displays them.
func Use(cin chan Fields) {
	for f := range cin {
		log.Printf("%s %s\n", f.SupporterKey, f.Email)
	}
}

//Mainline.  Find supporters and display some info about each.
func main() {
	cpath := kingpin.Flag("login", "YAML file containing credentials for Salsa Classic API").PlaceHolder("FILENAME").Required().String()
	crit := kingpin.Flag("criteria", "Search for records matching this criteria").PlaceHolder("CRITERIA").String()
	kingpin.Parse()

	a, err := godig.YAMLAuth(*cpath)
	if err != nil {
		log.Fatalf("Authentication error: %+v\n", err)
	}

	t := a.Supporter()
	count, err := t.Count("")
	log.Printf("Main: %v count is %v, err is %v\n", t.Name, count, err)
	if err != nil {
		panic(err)
	}
	c := make(chan Fields, 100)
	var wg sync.WaitGroup

	log.Println("Main: start")
	wg.Add(1)
	go func(w *sync.WaitGroup) {
		defer w.Done()
		Use(c)
	}(&wg)
	log.Println("Main: Use started")
	wg.Add(1)
	go func(w *sync.WaitGroup) {
		defer w.Done()
		All(&t, *crit, c)
	}(&wg)
	log.Println("Main: All started")

	log.Println("Main: waiting...")
	wg.Wait()
	log.Println("Main: done")
}
