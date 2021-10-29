// Application to show groups and supporters.  Typical output
// for exports.  You can control over which fields will appear
// and the select criteria.
package main

import (
	"encoding/csv"
	"log"
	"os"
	"strings"
	"sync"

	godig "github.com/salsalabs/godig/pkg"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

const (
	//outFile is the CSV groups output file
	outFile = "groups_and_supporters.csv"
	//tableName is the definition for getLeftJoin.sjs
	tableName = "supporter(supporter_KEY)supporter_groups(groups_KEY)groups"
	//criteria is the criteria for selecting and including fields.  This can be log...
	//Notice that "&condition=" is supplied by the API object
	criteria = "supporter_groups.Last_Modified>2021-09-28" +
		"&include=supporter.supporter_KEY,Email,groups.groups_KEY,Group_Name,supporter_groups.Last_Modified"
)

//Fields contains the contents to return.
type Fields struct {
	SupporterKey string `json:"supporter_KEY"`
	Email        string `json:"Email,omitempty"`
	GroupsKey    string `json:"groups_KEY"`
	GroupName    string `json:"Group_Name,omitempty"`
	JoinDate     string `json:"Last_Modified"`
}

//All reads all of the records and sends them to a Fields channel.
//parses the buffer for records then outputs them to cout.
func All(a *godig.API, cout chan Fields) {
	t := a.NewTable(tableName)
	offset := int32(0)
	count := 500
	for count > 0 {
		log.Printf("All: %v offset %6d\n", t.Name, offset)
		if offset > 0 && offset%5000 == 0 {
			log.Printf("All: %v offset %6d\n", t.Name, offset)
		}
		var a []Fields
		err := t.LeftJoin(offset, count, criteria, &a)
		if err != nil {
			log.Fatalf("All: %v offset %6d %v\n", t.Name, offset, err)
			break
		}
		count = len(a)
		if count == 0 {
			log.Printf("All: %v offset %6d, done\n", t.Name, offset)
			break
		}
		for _, r := range a {
			cout <- r
		}
		offset = offset + int32(count)
	}
	close(cout)
}

//Use reads Fields records from a channel and writes them
//to the CSV file.
func Use(cin chan Fields) {
	f, err := os.Create(outFile)
	if err != nil {
		log.Fatalf("%v, %s\n", err, outFile)
	}
	defer f.Close()
	w := csv.NewWriter(f)
	if err != nil {
		log.Fatalf("%v, %s\n", err, outFile)
	}
	headers := []string{
		"SupporterKey",
		"Email",
		"GroupsKey",
		"GroupName",
		"DateJoined",
	}
	w.Write(headers)
	w.Flush()
	for {
		r, ok := <-cin
		if !ok {
			break
		}
		a := []string{
			r.SupporterKey,
			r.Email,
			strings.TrimSpace(r.GroupName),
			r.GroupsKey,
			r.JoinDate,
		}
		w.Write(a)
		w.Flush()
	}
}

//Mainline.  Find supporters and display some info about each.
func main() {
	var (
		cpath      = kingpin.Flag("login", "YAML file of credentials for Salsa Classic API").PlaceHolder("FILENAME").Required().String()
		apiVerbose = kingpin.Flag("verbose", "Show requests to, and responses from, the server. Can be very noisy.").Bool()
	)
	kingpin.Parse()

	a, err := godig.YAMLAuth(*cpath)
	if err != nil {
		log.Fatalf("Authentication error: %+v\n", err)
	}
	a.Verbose = *apiVerbose

	c := make(chan Fields, 500)
	var wg sync.WaitGroup

	log.Println("Main: start")

	wg.Add(1)
	go func(c chan Fields, w *sync.WaitGroup) {
		defer w.Done()
		Use(c)
	}(c, &wg)
	log.Println("Main: Use started")

	wg.Add(1)
	go func(a *godig.API, c chan Fields, w *sync.WaitGroup) {
		defer w.Done()
		All(a, c)
	}(a, c, &wg)
	log.Println("Main: All started")

	log.Println("Main: waiting...")
	wg.Wait()

	log.Printf("Main: done, results in %s", outFile)
}
