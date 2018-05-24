package main

import (
	"log"
	"sync"

	"github.com/salsalabs/godig"
	"github.com/salsalabs/godig/cmd/addressfixer"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

//Mainline.  Find supporters and fix their addresses.
func main() {
	cpath := kingpin.Flag("credentials", "YAML file containing credentials for Salsa Classic API").PlaceHolder("FILENAME").Required().String()
	crit := kingpin.Flag("criteria", "Search for records matching this criteria").PlaceHolder("CRITERIA").String()
	chunkSize := kingpin.Flag("chunk-size", "Records per chunk").Default("50").Int()
	live := kingpin.Flag("live", "Update the database.  USE EXTREME CAUTION!!!").PlaceHolder("LIVE").Default("false").Bool()
	kingpin.Parse()

	a, err := godig.YAMLAuth(*cpath)
	if err != nil {
		log.Fatalf("Authentication error: %+v\n", err)
	}
	t := a.Supporter()

	c1 := make(chan []addressfixer.Supporter, 100)
	c2 := make(chan []addressfixer.Supporter, 100)
	c3 := make(chan []addressfixer.Supporter, 100)
	c4 := make(chan addressfixer.Mod, 1000)
	var wg sync.WaitGroup

	log.Println("Main:    start")

	wg.Add(1)
	go func(w *sync.WaitGroup) {
		defer w.Done()
		addressfixer.Audit(c4)
	}(&wg)
	log.Println("Main:    Audit started")

	wg.Add(1)
	go func(w *sync.WaitGroup) {
		defer w.Done()
		addressfixer.Finish(&t, c3, *live)
	}(&wg)
	log.Println("Main:    Finish started")

	wg.Add(1)
	go func(w *sync.WaitGroup) {
		defer w.Done()
		addressfixer.Fix(c2, c3, c4)
	}(&wg)
	log.Println("Main:    Fix started")

	wg.Add(1)
	go func(w *sync.WaitGroup) {
		defer w.Done()
		addressfixer.Split(c1, c2, *chunkSize)
	}(&wg)
	log.Println("Main:    Fix started")

	wg.Add(1)
	go func(w *sync.WaitGroup) {
		defer w.Done()
		addressfixer.ReadAll(&t, *crit, c1)
	}(&wg)
	log.Println("Main:    All started")

	log.Println("Main:    waiting...")
	wg.Wait()
	log.Println("Main:    done")

}
