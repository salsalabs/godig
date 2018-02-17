//TagTagData reads all tags and all tag data.
package main

import (
	"log"
	"sync"

	"github.com/salsalabs/godig"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

//Fields contains the contents to return.
type Fields struct {
	TagKey           string `json:"tag_KEY"`
	Prefix           string
	Tag              string
	DatabaseTableKey string `json:"database_table_KEY"`
	TableKey         string `json:"table_KEY"`
}

//All reads all of the records and sends them to a Fields channel.
//parses the buffer for records then outputs them to cout.
func All(t *godig.Table, crit string, cout chan Fields) {
	offset := 0
	count := 500
	for count > 0 {
		log.Printf("All: %v offset %6d\n", t.Name, offset)
		if offset > 0 && offset%5000 == 0 {
			log.Printf("All: %v offset %6d\n", t.Name, offset)
		}
		var a []Fields
		err := t.LeftJoin(offset, count, crit, &a)
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

//Use reads Fields records from a channel and displays them.
func Use(cin chan Fields) {
	for r := range cin {
		log.Printf("Use: %+v\n", r)
	}
}

//Mainline.  Find supporters and display some info about each.
func main() {
	cpath := kingpin.Flag("credentials", "YAML file containing credentials for Salsa Classic API").PlaceHolder("FILENAME").Required().String()
	crit := kingpin.Flag("criteria", "Search for records matching this criteria").PlaceHolder("CRITERIA").String()
	kingpin.Parse()

	a, err := godig.YAMLAuth(*cpath)
	if err != nil {
		log.Fatalf("Authentication error: %+v\n", err)
	}

	t := a.NewTable("tag(tag_KEY)tag_data")
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
