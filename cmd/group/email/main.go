package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"github.com/salsalabs/godig"
	"gopkg.in/alecthomas/kingpin.v2"
)

//Fetch is used to format the API call that retrieves data for a group.
const Fetch = `https://%s/api/getLeftJoin.sjs?json&object=groups(groups_KEY)supporter_groups(supporter_KEY)email&condition=groups.groups_KEY=%v&condition=email.Time_Sent>=2017-08-01&condition=email.Time_Sent<2017-12-31&include=groups.groups_KEY,email.email_KEY,email.Time_Sent`

// Record will contain the information retrieved from the database.
type Record struct {
	Groups_KEY string
	Email_KEY  string
	Time_Sent  string
}

//Mainline retrieves data via the API and aggregates it into a local database.
func main() {
	group := kingpin.Flag("groups_KEY", "Retrieve stats for this group.").Required().Int()
	cpath := kingpin.Flag("credentials", "YAML file containing credentials for Salsa Classic API").Required().PlaceHolder("FILENAME").String()
	dbname := kingpin.Flag("database", "Use this database for this run").Default("./db/godig.sqlite").String()
	kingpin.Parse()

	db, err := sql.Open("sqlite3", *dbname)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	stmt, err := db.Prepare("update group_email set count = count + ? where groups_KEY = ?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	w := godig.NewWrapper(db, stmt, *group)
	err = godig.Authenticate(w, *cpath)
	if err != nil {
		log.Fatal(err)
	}

	offset := 0
	count := 500
	var a []Record
	u := fmt.Sprintf(Fetch, w.Host, w.GroupsKey)
	for count > 0 {
		err := godig.Many(w, u, offset, count, &a)
		if err != nil {
			log.Fatal(err)
		}
		count := len(a)
		log.Printf("group %v count %3d\n", w.GroupsKey, count)
		_, err = w.Statement.Exec(count, w.GroupsKey)
		if err != nil {
			log.Fatal(err)
		}
		offset = offset + count
	}
}
