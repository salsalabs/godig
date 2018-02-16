//APICondition is a test program to find out why adding &condition=whatever to a
//Salsa API URL messes it up.
package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/salsalabs/godig"
	"gopkg.in/alecthomas/kingpin.v2"
)

//Fetch is the format for submitting the read command to Salsa. Data doesn't need to be
//ordered to work correctly. We'll use the SQLite database offline to sort.
//Note that adding "&orderBy=Other_Data_2" really, REALLY, slows this down.
const Fetch = `https://%s/api/getLeftJoin.sjs?json&object=supporter(supporter_KEY)email&condition=Other_Data_2 IS NOT EMPTY&include=Other_Data_2,email.Status,email.Last_Modified`

const Works = `https://%s/api/getLeftJoin.sjs?json&object=supporter(supporter_KEY)email&include=Other_Data_2,email.Status,email.Last_Modified&condition=Other_Data_2%20IS%20NOT%20EMPTY&condition=Other_Data_2%3E0&orderBy=Other_Data_2`

//FetchFields are retrieved from a join of the supporer and email tables in Salsa.
//Field names were derived from retrieving a record and seeing exactly what Salsa passed
//back from the API.  Just sayin...
type FetchFields struct {
	Location     string `json:"Other_Data_2"`
	LastModified string `json:"Last_Modified"`
	Status       string
	Year         int
}

//Update populates the year in a FetchFields object.
func (f *FetchFields) Update() {
	const form = "Mon Jan 02 2006 15:04:05 GMT-0700 (MST)"
	t, err := time.Parse(form, f.LastModified)
	if err != nil {
		log.Fatalf("Time Parse error on %v: %v\n", f.LastModified, err)
	}
	f.Year = t.Year()
}

//Mainline retrieves data via the API and aggregates it into a local database.
func main() {
	cpath := kingpin.Flag("credentials", "YAML file containing credentials for Salsa Classic API").PlaceHolder("FILENAME").Required().String()
	kingpin.Parse()

	db, err := sql.Open("mysql", "godig:speaker-1-operational@/godig")
	if err != nil {
		log.Fatalf("SQL Open error %v\n", err)
	}
	defer db.Close()
	w := godig.NewWrapper(db, nil, 0)
	err = godig.Authenticate(w, *cpath)
	if err != nil {
		log.Fatalf("Authentication error %v\n", err)
	}

	// Base URL to retrieve data.
	x := fmt.Sprintf(Works, w.Host)
	log.Println("URL:", x)

	// Read data and process it.
	offset := 0
	count := 500

	var a []FetchFields
	for count > 0 {
		err = godig.Many(w, x, offset, count, &a)
		if err != nil {
			log.Fatalf("godig.Many error %v at offset %v", err, offset)
		}
		for _, f := range a {
			if len(f.Location) != 0 && len(f.LastModified) != 0 {
				f.Update()
				log.Printf("Record: %+v\n")
			}
			count = int(len(a))
			offset = offset + count
			log.Printf("Offset: %6d\n", offset)
		}
		log.Printf("Done: %d\n", offset)
	}
}
