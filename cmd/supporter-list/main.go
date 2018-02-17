package main

import (
	"log"
	"sync"

	"github.com/salsalabs/godig"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

//Fields are retrieved from the supporter record.
type Fields struct {
	SupporterKey string `json:"supporter_KEY"`
	FirstName    string `json:"First_Name"`
	LastName     string `json:"Last_Name"`
	Email        string
}

//All reads from Salsa via the API.  Each read puts a byte buffer
//onto the provided channel.
func All(t *godig.Table, cout chan Fields) {
	offset := 0
	count := 500
	for count > 0 {
		log.Printf("All: %v offset %6d\n", t.Name, offset)
		if offset > 0 && offset%5000 == 0 {
			log.Printf("All: %v offset %6d\n", t.Name, offset)
		}
		var a []Fields
		err := t.Many(offset, count, &a)
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
		offset = offset + count
	}
}

//Use accepts Fields records from a channel and displays them.
func Use(cin chan Fields) {
	for f := range cin {
		x := f.SupporterKey
		x = x + "./"
		//log.Printf("%s %s %s %s", f.SupporterKey, f.FirstName, f.LastName, f.Email)
	}
}

//Mainline.  Find supporters and display some info about each.
func main() {
	cpath := kingpin.Flag("credentials", "YAML file containing credentials for Salsa Classic API").PlaceHolder("FILENAME").Required().String()
	kingpin.Parse()

	a, err := godig.YAMLAuth(*cpath)
	if err != nil {
		log.Fatalf("Authentication error: %+v\n", err)
	}

	t := a.Supporter()
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
		All(&t, c)
	}(&wg)
	log.Println("Main: All started")

	log.Println("Main: waiting...")
	wg.Wait()
	log.Println("Main: done")
}
