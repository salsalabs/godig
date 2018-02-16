//HTMLContentFix retrieves the HTML content from an email blast, changes an errant
//HTML entity to its HTML equivalent, then saves the change.
//Standard credentials protocol.
//You provide the email blast KEY in the invocation.
package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/salsalabs/godig"
	"gopkg.in/alecthomas/kingpin.v2"
)

//Fields are retrieved from the table.
type Fields struct {
	HtmlContent string `json:"HTML_Content"`
}

//Table is the table name being modified.
const Table = "email_blast"

//Mainline retrieves data via the API and aggregates it into a local database.
func main() {
	cpath := kingpin.Flag("credentials", "YAML file containing credentials for Salsa Classic API").PlaceHolder("FILENAME").Required().String()
	blastKey := kingpin.Flag("email_blast_KEY", "Modify this email blast").Required().String()
	kingpin.Parse()

	w := godig.NewWrapper(nil, nil, 0)
	err := godig.Authenticate(w, *cpath)
	if err != nil {
		log.Fatalf("Authentication error: %+v\n", err)
	}

	var f Fields
	k := godig.ParseInt(*blastKey)
	err = godig.One(w, Table, k, &f)
	if err != nil {
		log.Fatalf("Read error: %+v on %s key %v\n", err, Table, blastKey)
	}
	c := strings.Replace(f.HtmlContent, "&lt;", "<", -1)
	c = fmt.Sprintf("HTML_Content=%s", c)
	log.Println(c)
	err = godig.Save(w, Table, k, c)
	if err != nil {
		log.Fatalf("Save error: %+v on email_blast key %v\n", err, blastKey)
	}
}
