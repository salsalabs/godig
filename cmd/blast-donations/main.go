//BlastDonations accepts a date range then attributes donations to the blasts
//that were created in that timeframe.  Output goes to the "blast_donations"
//table in the app-level.  This app needs Salsa Classic campaign manager
//credentials to run.  The credentials should have "Suppoter Management",
//"Donations" and "Enterprise marker" permissions.
//
//Note that the "blast_donations" table needs to be empty before running this
//app.  Making that automatic is TBD.
package main

import (
	"database/sql"
	"fmt"
	"math"

	"log"

	_ "github.com/mattn/go-sqlite3"
	"github.com/salsalabs/godig"
	"gopkg.in/alecthomas/kingpin.v2"
)

//EBFields contains the fields that are captured when reading email blasts.
type EBFields struct {
	Email_Blast_KEY string
	KeyInt          int
	Subject         string
}

//EBLink is the URL used to retrieve email blasts for a specific date range.
const EBLink = "https://%v/api/getObjects.sjs?json&object=email_blast&condition=Date_Created>=%v&condition=Date_Created<%v&include=email_blast_KEY,Subject"

//DFields contains the fields that are captured when reading donations.
type DFields struct {
	Donation_KEY     string
	RESULT           string
	Transaction_Date string
	Amount           string
}

//DLink is the URL used to retrieve donations tagged by a specific email blast.
const DLink = "https://%v/api/getTaggedObjects.sjs?json&object=donation&tag=email_blast:%v"

//Mainline retrieves data via the API and aggregates it into a local database.
func main() {
	cpath := kingpin.Flag("credentials", "YAML file containing credentials for Salsa Classic API").Required().PlaceHolder("FILENAME").String()
	sdate := kingpin.Flag("start", "Start date").PlaceHolder("YYYY-MM-DD").Required().String()
	edate := kingpin.Flag("end", "Day AFTER end date").PlaceHolder("YYYY-MM-DD").Required().String()
	dbname := kingpin.Flag("database", "Use this database").Default("./db/godig.sqlite").String()
	kingpin.Parse()

	db, err := sql.Open("sqlite3", *dbname)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	stmt, err := db.Prepare("insert into blast_donations values(?, ?, ?, ?, ?, ?);")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	w := godig.NewWrapper(db, stmt, 0)
	err = godig.Authenticate(w, *cpath)
	if err != nil {
		log.Fatal(err)
	}

	// Get all of the email blasts in the specified date range.
	criteria := godig.Dates{StartDate: *sdate, EndDate: *edate}
	x := fmt.Sprintf(EBLink, w.Host, criteria.StartDate, criteria.EndDate)

	offset := 0
	count := 500

	var allBlasts []EBFields
	var a []EBFields
	for count > 0 {
		err = godig.Many(w, x, offset, count, &a)
		if err != nil {
			log.Fatal(err)
		}
		for _, r := range a {
			r.KeyInt = godig.ParseInt(r.Email_Blast_KEY)
			allBlasts = append(allBlasts, r)
		}
		count = int(len(a))
		offset = offset + count
	}

	for _, r := range allBlasts {
		x := fmt.Sprintf(DLink, w.Host, r.KeyInt)

		offset = 0
		count = 500
		var dRecs []DFields
		var ebRecs []DFields

		for count > 0 {
			err = godig.Many(w, x, offset, count, &dRecs)
			if err != nil {
				log.Fatal(err)
			}
			count = int(len(dRecs))
			offset = offset + count
			for _, r := range dRecs {
				ebRecs = append(ebRecs, r)
			}
		}
		// Compute count, minimum, maximum and sum for this email blast
		n := 0
		t := float64(0.0)
		min := float64(1e9)
		max := float64(0.0)
		for _, s := range ebRecs {
			x := godig.ParseInt(s.RESULT)
			if x == -1 || x == 0 {
				n = n + 1
				a := float64(godig.ParseFloat(s.Amount))
				t = t + a
				max = math.Max(max, a)
				min = math.Min(min, a)
			}
		}

		if n != 0 {
			_, err = w.Statement.Exec(r.KeyInt, r.Subject, n, min, max, t)
			if err != nil {
				log.Fatal(err)
			}
		}
		log.Printf("%v: %3d", r.KeyInt, len(ebRecs))
	}
}
