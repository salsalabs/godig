//NYSNAData extracts opens and clicks by year for all of the supporter.Other_Data_2 values
//in the database.  Other_Data_2 is used by NYSNA as a location field.  The result will be
//a local database containing clicks and options by location.
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
const Fetch = `https://%s/api/getLeftJoin.sjs?json&object=supporter(supporter_KEY)email&include=Other_Data_2,email.Status,email.Last_Modified&condition=Other_Data_2%%20IS%%20NOT%%20EMPTY&condition=Other_Data_2>0`

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

	a := godig.NewAPI()
	c, err := godig.Credentials(*cpath)
	if err != nil {
		log.Fatalf("Error %v reading credential file %v\n", err, *cpath)
	}
	err = a.Authenticate(c)
	if err != nil {
		log.Fatalf("Authentication error %v\n", err)
	}

	db, err := sql.Open("mysql", "godig:speaker-1-operational@/godig")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	scrub, err := db.Prepare("delete from location_stats;")
	if err != nil {
		log.Fatalf("Scrub prepare error %v\n", err)
	}
	defer scrub.Close()

	check, err := db.Prepare("select location from location_stats where location = ? and year = ?")
	if err != nil {
		log.Fatalf("Check prepare error %v\n", err)
	}
	defer check.Close()

	insert, err := db.Prepare("insert into location_stats(location, year, sent, sent_and_opened, sent_and_clicked) values (?, ?, 0, 0, 0)")
	if err != nil {
		log.Fatalf("insert prepare error %v\n", err)
	}
	defer insert.Close()

	sentUpdate, err := db.Prepare("update location_stats set sent = sent + 1 where location = ? and year = ?;")
	if err != nil {
		log.Fatalf("Sent update prepare error %v\n", err)
	}
	defer sentUpdate.Close()

	sentOpenedUpdate, err := db.Prepare("update location_stats set sent_and_opened = sent_and_opened + 1 where location = ? and year = ?;")
	if err != nil {
		log.Fatalf("Sent and opened update prepare error %v\n", err)
	}
	defer sentUpdate.Close()

	sentClickedUpdate, err := db.Prepare("update location_stats set sent_and_clicked = sent_and_clicked + 1 where location = ? and year = ?;")
	if err != nil {
		log.Fatalf("Sent and clicked update prepare error %v\n", err)
	}
	defer sentUpdate.Close()

	// Clean out the database.
	_, err = scrub.Exec()
	if err != nil {
		log.Fatalf("Scrub: %v\n", err)
	}

	// Read data and process it.
	offset := 0
	count := 500

	// Base URL to retrieve data.
	x := fmt.Sprintf(Fetch, a.Host)
	log.Println(x)
	t := a.NewTable("supporter")

	var b []FetchFields
	for count > 0 {
		if offset%5000 == 0 {
			log.Printf("Offset: %6d\n", offset)
		}
		err = t.Many(x, offset, count, &b)
		if err != nil {
			log.Fatal(err)
		}
		for _, f := range b {
			if len(f.Location) != 0 && len(f.LastModified) != 0 {
				f.Update()

				//Record exists?
				var x string
				err := check.QueryRow(f.Location, f.Year).Scan(&x)
				if err != nil {
					if err == sql.ErrNoRows {
						// Nope, initialize it.
						_, err := insert.Exec(f.Location, f.Year)
						if err != nil {
							log.Printf("Record: %+v\n", f)
							log.Fatalf("Insert: %v\n", err)
						}
					} else {
						log.Printf("Record: %+v\n", f)
						log.Fatalf("Check:  %v\n", err)
					}
				}

				err = nil
				switch f.Status {
				case "Sent":
					_, err = sentUpdate.Exec(f.Location, f.Year)
				case "Sent and Opened":
					_, err = sentOpenedUpdate.Exec(f.Location, f.Year)
				case "Sent and Clicked":
					_, err = sentClickedUpdate.Exec(f.Location, f.Year)
				}
				if err != nil {
					log.Printf("Record: %+v\n", f)
					log.Fatalf("Update %s: %v\n", f.Status, err)
				}
			}
		}
		count = int(len(b))
		offset = offset + count
	}
	log.Printf("Done: %d\n", offset)
}
