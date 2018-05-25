package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

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
	fileLog := kingpin.Flag("file-log", "Write the log to a timestamped File").PlaceHolder("File").Default("false").Bool()
	fixerCount := kingpin.Flag("fixer-count", "Number of fixer threads").Default("1").Int()
	kingpin.Parse()

	a, err := godig.YAMLAuth(*cpath)
	if err != nil {
		log.Fatalf("Authentication error: %+v\n", err)
	}
	t := a.Supporter()

	//Redirect 'log' to a timestamped file.
	if *fileLog {
		now := time.Now()
		fn := fmt.Sprintf("addressfixer-%v.log", now.Format(time.RFC3339))
		log.Printf("Main:    Logging to %v\n", fn)
		f, err := os.Create(fn)
		if err != nil {
			log.Fatalf("Main:    %v on %v\n", err, fn)
		}
		defer f.Close()
		writer := bufio.NewWriter(f)
		log.SetOutput(writer)
	}
	log.Printf("Main:    Start on %v with criteria '%v'\n", a.Host, *crit)

	c1 := make(chan []addressfixer.Supporter, 1000)
	c2 := make(chan []addressfixer.Supporter, 1000)
	c3 := make(chan []addressfixer.Supporter, 1000)
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

	log.Printf("Main:    Starting %v fixer(s)\n", *fixerCount)
	fm := &sync.Mutex{}
	for i := 1; i <= *fixerCount; i++ {
		wg.Add(1)
		log.Printf("Main:    Fix %v started\n", i)
		go func(w *sync.WaitGroup) {
			defer w.Done()
			addressfixer.Fix(c2, c3, c4, fm, i)
		}(&wg)
	}

	wg.Add(1)
	go func(w *sync.WaitGroup) {
		defer w.Done()
		addressfixer.Chunk(c1, c2, *chunkSize)
	}(&wg)
	log.Println("Main:    Chunk started")

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
