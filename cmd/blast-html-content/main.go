//HTMLContent retrieves the HTML content from an email blast and saves it to disk.
//Standard credentials protocol.
//You provide the email blast KEY in the invocation.
package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/salsalabs/godig"
	"gopkg.in/alecthomas/kingpin.v2"
)

//FetchFields are retrieved from the table.
type FetchFields struct {
	HtmlContent string `json:"HTML_Content"`
}

//Mainline retrieves data via the API and aggregates it into a local database.
func main() {
	cpath := kingpin.Flag("credentials", "YAML file containing credentials for Salsa Classic API").PlaceHolder("FILENAME").Required().String()
	blastKey := kingpin.Flag("email_blast_KEY", "Read this email blast").Required().String()
	kingpin.Parse()

	w := godig.NewWrapper(nil, nil, 0)
	err := godig.Authenticate(w, *cpath)
	if err != nil {
		log.Fatal(err)
	}

	var f FetchFields
	err = godig.One(w, "email_blast", godig.ParseInt(*blastKey), &f)
	if err != nil {
		log.Fatalf("Read error: %+v on email_blast_template key %v\n", err, blastKey)
	}
	c := f.HtmlContent

	b := []byte(c)
	fn := fmt.Sprintf("template_%s_html_content.html", *blastKey)
	err = ioutil.WriteFile(fn, b, 0777)
	if err != nil {
		log.Fatalf("Write error: %+v on %v\n", err, fn)
	}
	log.Println("Output is in ", fn)
}
