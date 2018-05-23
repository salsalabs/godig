package main

import (
	"log"
	"sync"

	"github.com/salsalabs/godig"
	"github.com/salsalabs/godig/cmd/addressfixer"
	"github.com/salsalabs/godig/cmd/addressfixer/active"
	"github.com/salsalabs/godig/cmd/addressfixer/passive"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

//All implements Reader.  Reads all supporters for a criteria
//then passes arrays of supporter JSON downstream.
func All(t *godig.Table, crit string, c chan []addressfixer.Supporter) {
	offset := 0
	count := 500

	for count > 0 {
		log.Printf("All: %v offset %6d\n", t.Name, offset)
		if offset > 0 && offset%5000 == 0 {
			log.Printf("All: %v offset %6d\n", t.Name, offset)
		}
		var a []addressfixer.Supporter
		err := t.Many(offset, count, crit, &a)
		if err != nil {
			log.Fatalf("All: %v offset %6d %v\n", t.Name, offset, err)
			return
		}
		count = len(a)
		if count == 0 {
			log.Printf("All: %v offset %6d, done\n", t.Name, offset)
			//close(c)
		} else {
			c <- a
			offset = offset + count
		}
	}
}

//Audit record changes to a supporter record.
func Audit(c chan addressfixer.Mod) {
	for a := range c {
		log.Printf("Audit: %+v\n", a)
	}
}

//Split accepts a buffer and splits it into supporter records.
//Supporter records then flow through the channel.
func Split(c1 chan []addressfixer.Supporter, c2 chan addressfixer.Supporter) {
	defer close(c2)
	for a := range c1 {
		log.Printf("Split: received %v supporters\n", len(a))
		for _, r := range a {
			c2 <- r
		}
	}
}

//Mainline.  Find supporters and fix their addresses.
func main() {
	cpath := kingpin.Flag("credentials", "YAML file containing credentials for Salsa Classic API").PlaceHolder("FILENAME").Required().String()
	crit := kingpin.Flag("criteria", "Search for records matching this criteria").PlaceHolder("CRITERIA").String()
	live := kingpin.Flag("live", "Update the database.  USE EXTREME CAUTION!!!").PlaceHolder("LIVE").Default("false").Bool()
	kingpin.Parse()

	a, err := godig.YAMLAuth(*cpath)
	if err != nil {
		log.Fatalf("Authentication error: %+v\n", err)
	}
	t := a.Supporter()

	c1 := make(chan []addressfixer.Supporter, 100)
	c2 := make(chan addressfixer.Supporter, 100)
	c3 := make(chan addressfixer.Supporter, 100)
	c4 := make(chan addressfixer.Mod, 100)
	var wg sync.WaitGroup

	log.Println("Main: start")

	wg.Add(1)
	go func(w *sync.WaitGroup) {
		defer w.Done()
		Audit(c4)
	}(&wg)
	log.Println("Main: Audit started")

	wg.Add(1)
	go func(w *sync.WaitGroup) {
		defer w.Done()
		if *live {
			active.Finish(c3)
		} else {
			passive.Finish(c3)
		}
	}(&wg)
	log.Println("Main: Finish started")

	wg.Add(1)
	go func(w *sync.WaitGroup) {
		defer w.Done()
		if *live {
			active.Fix(c2, c3, c4)
		} else {
			passive.Fix(c2, c3, c4)
		}
	}(&wg)
	log.Println("Main: Fix started")

	wg.Add(1)
	go func(w *sync.WaitGroup) {
		defer w.Done()
		Split(c1, c2)
	}(&wg)
	log.Println("Main: Fix started")

	wg.Add(1)
	go func(w *sync.WaitGroup) {
		defer w.Done()
		All(&t, *crit, c1)
	}(&wg)
	log.Println("Main: All started")

	log.Println("Main: waiting...")
	wg.Wait()
	log.Println("Main: done")
}
