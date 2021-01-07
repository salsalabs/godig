// Cleanup tool.  Accept a date range for donations.  Delete the donationss
// and the supporters that made them.

package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	godig "github.com/salsalabs/godig/pkg"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

const (
	//DonationCount is the number of donation deleters to start.
	DonationCount = 5
	//SupporterCount is the number of supporter deleters to start.
	SupporterCount = 5
	//DonationFlag appears in log lines for donation activity
	DonationFlag = "d"
	//SupporterFlag appears in log lines for supporter activity
	SupporterFlag = "d"
)

//drive reads donation records that match "criteria" from Salsa Classic.
//Writes the donation key to "dc" and the supporter key to "sc".  Closes
//both after all matching donations are read.
func drive(t *godig.Table, criteria string, dc, sc chan string) {
	log.Printf("drive: start\n")
	count := 500
	offset := int32(0)
	total := int32(0)
	for count == 500 {
		b, err := t.ManyMap(offset, count, criteria)
		log.Printf("drive: %7d\n", offset)
		if err != nil {
			panic(err)
		}
		for _, r := range b {
			dc <- r["donation_KEY"]
			if len(r["supporter_KEY"]) > 0 {
				sc <- r["supporter_KEY"]
			}
		}
		count = len(b)
		offset += int32(count)
		total += int32(count)
	}
	close(dc)
	close(sc)
	log.Printf("drive: end %d records\n", total)
}

//whack accepts primary keys from a channel, then uses the table to delete
//the matching records.  Displays delete results for every record.  Sends
//a message on the "done" channel when the keys channel is empty.
func whack(i int, f string, t *godig.Table, c chan string, done chan bool) {
	log.Printf("whack-%s-%02d: start\n", f, i)
	for true {
		k, ok := <-c
		if !ok {
			break
		}
		var ds godig.DeleteStatus
		t.Delete(k, &ds)
		if ds.Result == "error" {
			for _, m := range ds.Messages {
				log.Printf("whack-%s-%02d: key %s, %s %s\n", f, i, k, ds.Result, m)
			}
		} else {
			log.Printf("whack-%s-%02d: key %s, %s\n", f, i, k, ds.Result)
		}
	}
	done <- true
	log.Printf("whack-%s-%02d: end\n", f, i)
}

//watch waits for a number of messages on the "done" queue.  Returns after
//that.  This function makes sure that all channel readers are done before
//the app terminates.
func watch(x int, done chan bool) {
	log.Println("watch: start")
	for x > 0 {
		log.Printf("watch: waiting for %d task(s)\n", x)
		_, _ = <-done
		x--
	}
	log.Println("watch: end")
}

//main accepts command line arguments then deletes donations that match the
//criteria.
func main() {
	cpath := kingpin.Flag("login", "YAML file containing login credentials for Salsa Classic API").PlaceHolder("FILENAME").Required().String()
	sdate := kingpin.Flag("start-date", "First last modified date as YYYY-MM-YY").Default("2021-01-01").String()
	edate := kingpin.Flag("end-date", "Day after last modified date as YYYY-MM-dd").Default("2021-02-01").String()
	verbose := kingpin.Flag("verbose", "Lots and *lots* of debug noise.  Not recommended...").Bool()
	kingpin.Parse()
	if *edate <= *sdate {
		log.Fatalf("End date must be after start date!\n")
	}
	api, err := godig.YAMLAuth(*cpath)
	api.Verbose = *verbose
	if err != nil {
		log.Fatalf("Authentication error %v\n", err)
	}
	donation := api.Donation()
	supporter := api.Supporter()
	criteria := fmt.Sprintf("Last_Modified>%s&condition=Last_Modified<%s", *sdate, *edate)

	dc := make(chan string, 100)
	sc := make(chan string, 100)
	done := make(chan bool, 10)
	var wg sync.WaitGroup

	// Start donation listeners
	for i := 0; i < DonationCount; i++ {
		go (func(i int, wg *sync.WaitGroup, t *godig.Table, c chan string, done chan bool) {
			wg.Add(1)
			whack(i, DonationFlag, t, c, done)
			wg.Done()
		})(i+1, &wg, &donation, dc, done)
	}

	// Start supporter listeners
	for i := 0; i < SupporterCount; i++ {
		go (func(i int, wg *sync.WaitGroup, t *godig.Table, c chan string, done chan bool) {
			wg.Add(1)
			whack(i, SupporterFlag, t, c, done)
			wg.Done()
		})(i+1, &wg, &supporter, sc, done)
	}

	// Start termination listener.
	go (func(wg *sync.WaitGroup, done chan bool) {
		wg.Add(1)
		watch(DonationCount+SupporterCount, done)
		wg.Done()
	})(&wg, done)

	// Start driver.
	go (func(wg *sync.WaitGroup, t *godig.Table, criteria string, dc, sc chan string) {
		wg.Add(1)
		drive(t, criteria, dc, sc)
		wg.Done()
	})(&wg, &donation, criteria, dc, sc)

	//Classic starts slowly.  Nap a bit before waiting for things to be done.
	time.Sleep(2 * time.Second)

	// Wait for things to complete.
	wg.Wait()
}
