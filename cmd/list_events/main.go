package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"

	"github.com/salsalabs/godig"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

//Templates for formatting event page URLs.
const (
	National = "http://engage.jewishpublicaffairs.org/p/salsa/event/common/public/?event_KEY=%v"
	Chapter  = "http://engage.jewishpublicaffairs.org/c/%v/p/salsa/event/common/public/?event_KEY=%v"
	Filename = "%v %05v %v"
	Out      = "node hn.js \"%v\" \"pdfs/events/%v\"\n"
	Punct    = `"[,.;!?(){}\\[\\]<>%]"`
)

//Fields are retrieved from the event record.
type Fields struct {
	Key     string `json:"event_KEY"`
	Chapter string `json:"chapter_KEY"`
	Date    string `json:"Date_Created"`
	RefName string `json:"Reference_Name"`
	Title   string `json:"Title"`
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
	close(cout)
}

//Use accepts Fields records from a channel and displays them.
func Use(cin chan Fields) {
	re := regexp.MustCompile(Punct)
	for f := range cin {
		u := fmt.Sprintf(National, f.Key)
		if len(f.Chapter) > 0 {
			u = fmt.Sprintf(Chapter, f.Chapter, f.Key)
		}
		d := godig.EngageDate(f.Date)
		t := strings.TrimSpace(f.Title)
		t = re.ReplaceAllString(t, "")
		if len(t) == 0 {
			t = strings.TrimSpace(f.RefName)
		}
		t = re.ReplaceAllString(t, "")
		p := fmt.Sprintf(Filename, d, f.Key, t)
		fmt.Printf(Out, u, p)
	}
}

//Mainline.  Find events and display some info about each.
func main() {
	login := kingpin.Flag("login", "YAML file containing credentials for Salsa Classic API").PlaceHolder("FILENAME").Required().String()
	crit := kingpin.Flag("criteria", "Search for records matching this criteria").PlaceHolder("CRITERIA").String()
	kingpin.Parse()

	a, err := godig.YAMLAuth(*login)
	if err != nil {
		log.Fatalf("Authentication error: %+v\n", err)
	}

	t := a.NewTable("event")
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
