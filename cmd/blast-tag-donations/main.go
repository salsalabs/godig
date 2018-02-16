package main

import (
	"database/sql"
	"fmt"
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

//EBCriteria defines which email blasts that we're interested in.  NOTE that
// EndDate is actually the day after the end date.
type EBCriteria struct {
	StartDate string
	EndDate   string
}

//EBLink is the URL used to retrieve email blasts for a specific date range.
const EBLink = "https://%v/api/getObjects.sjs?json&object=email_blast&condition=Date_Created>=%v&condition=Date_Created<%v&include=email_blast_KEY,Subject"

//Tag defines the fields returned from the tag data for an email blast tag.
type TagFields struct {
	Tag_KEY string
}

//TagLink is the URL temp[late for retrieving an email blast tag.
const TagLink = "https://%v/api/getObjects.sjs?json&object=tag&include=tag_KEY&condition=prefix=email_blast&condition=tag=%v"

//TagData defines the donation page fields retrieved for an email blast tag.
type TagDataFields struct {
	Table_KEY string
}

//TagDataLink is the URL template for retrieving donation table keys for
//an emal blast tag.
const TagDataLink = "https://%v/api/getObjects.sjs?json&object=tag_data&include=table_KEY&condition=database_table_KEY=45&condition=tag_KEY=%v"

//EmailBlasts uses a URL template to retrieve blasts in a timeframe.
func EmailBlasts(w *godig.Wrapper, t string, criteria EBCriteria) ([]EBFields, error) {
	x := fmt.Sprintf(t, w.Host, criteria.StartDate, criteria.EndDate)
	var all []EBFields
	var a []EBFields

	offset := 0
	count := 500
	for count > 0 {
		err := godig.Many(w, x, offset, count, &a)
		if err != nil {
			return nil, err
		}
		for _, r := range a {
			r.KeyInt = godig.ParseInt(r.Email_Blast_KEY)
			all = append(all, r)
		}
		count = int(len(a))
		offset = offset + count
	}
	return all, nil
}

//Tags uses a URL template to retrieve blasts in a timeframe.
func Tags(w *godig.Wrapper, t string, ebk string) ([]TagFields, error) {
	x := fmt.Sprintf(t, w.Host, ebk)
	var all []TagFields
	var a []TagFields

	offset := 0
	count := 500
	for count > 0 {
		err := godig.Many(w, x, offset, count, &a)
		if err != nil {
			return nil, err
		}
		for _, r := range a {
			all = append(all, r)
		}
		count = int(len(a))
		offset = offset + count
	}
	return all, nil
}

//TagData uses a URL template to retrieve tag_data records for a donations with a tag.
func TagData(w *godig.Wrapper, t string, tk string) ([]TagDataFields, error) {
	x := fmt.Sprintf(t, w.Host, tk)
	var all []TagDataFields
	var a []TagDataFields

	offset := 0
	count := 500
	for count > 0 {
		err := godig.Many(w, x, offset, count, &a)
		if err != nil {
			return nil, err
		}
		for _, r := range a {
			all = append(all, r)
		}
		count = int(len(a))
		offset = offset + count
	}
	return all, nil
}

//Mainline retrieves data via the API and aggregates it into a local database.
func main() {
	cpath := kingpin.Flag("credentials", "YAML file containing credentials for Salsa Classic API").Required().PlaceHolder("FILENAME").String()
	sdate := kingpin.Flag("start", "Start date").PlaceHolder("YYYY-MM-DD").Required().String()
	edate := kingpin.Flag("end", "Day AFTER end date").PlaceHolder("YYYY-MM-DD").Required().String()
	dbname := kingpin.Flag("database", "Use this database for this run").Default("./db/godig.sqlite").String()
	kingpin.Parse()

	db, err := sql.Open("sqlite3", *dbname)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	w := godig.NewWrapper(db, nil, 0)
	err = godig.Authenticate(w, *cpath)
	if err != nil {
		log.Fatal(err)
	}

	criteria := EBCriteria{StartDate: *sdate, EndDate: *edate}
	allBlasts, err := EmailBlasts(w, EBLink, criteria)
	if err != nil {
		log.Fatal(err)
	}
	for _, r := range allBlasts {
		allTags, err := Tags(w, TagLink, r.Email_Blast_KEY)
		if err != nil {
			log.Fatal(err)
		}

		// There should be zero or one matching record...
		if len(allTags) == 1 {
			tagKey := allTags[0].Tag_KEY
			allTagData, err := TagData(w, TagDataLink, tagKey)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("%v: %3d donations", r.Email_Blast_KEY, len(allTagData))
		} else {
			log.Printf("%v: %3d tags", r.Email_Blast_KEY, len(allTags))
		}
	}
}
