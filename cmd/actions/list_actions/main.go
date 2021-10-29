package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/salsalabs/godig/pkg"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

//Templates for formatting action page URLs.
const (
	National = "http://engage.jewishpublicaffairs.org/p/dia/action4/common/public/?action_KEY=%v"
	Chapter  = "http://engage.jewishpublicaffairs.org/c/%v/p/dia/action4/common/public/?action_KEY=%v"
	Filename = "%v %05v %v"
	Output   = "fetch_actions.bash"
	Process  = "node hn.js \"%v\" \"pdfs/actions/%v\"\n"
	Punct    = "[[:punct:]]"
)

//Fields are retrieved from the action record.
type Fields struct {
	Key     string `json:"action_KEY"`
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

//Use accepts events and formats them as a script.  Output goes to
//a file.
func Use(cin chan Fields) {
	re := regexp.MustCompile(Punct)
	f, err := os.Create(Output)
	if err != nil {
		panic(err)
	}
	for e := range cin {
		u := fmt.Sprintf(National, e.Key)
		if len(e.Chapter) > 0 && e.Chapter != "0" {
			u = fmt.Sprintf(Chapter, e.Chapter, e.Key)
		}
		d := godig.EngageDate(e.Date)
		t := strings.TrimSpace(e.Title)
		if len(t) == 0 {
			t = strings.TrimSpace(e.RefName)
		}
		t = re.ReplaceAllString(t, "")
		p := fmt.Sprintf(Filename, d, e.Key, t)
		p = strings.TrimSpace(p)
		p = p + ".pdf"
		s := fmt.Sprintf(Process, u, p)
		_, _ = f.WriteString(s)
	}
	f.Close()
}

//Mainline.  Find actions and display some info about each.
func main() {
	login := kingpin.Flag("login", "YAML file containing credentials for Salsa Classic API").PlaceHolder("FILENAME").Required().String()
	crit := kingpin.Flag("criteria", "Search for records matching this criteria").PlaceHolder("CRITERIA").String()
	kingpin.Parse()

	a, err := godig.YAMLAuth(*login)
	if err != nil {
		log.Fatalf("Authentication error: %+v\n", err)
	}

	t := a.NewTable("action")
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
