//Count the number of records for a table name.
package main

import (
	"log"

	"github.com/salsalabs/godig"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

//EmailBlast is the content that this app will read from Salsa Classic.
type EmailBlast struct {
	EmailBlastKey int32  `json:"email_blast_KEY"`
	DateRequested string `json:"Date_Requested"`
	Subject       string `json:"Subject"`
	HTMLContent   string `json:"HTML_Content"`
	TextContent   string `json:"Text_Content"`
}

func main() {
	login := kingpin.Flag("login", "YAML file containing credentials for Salsa Classic API").PlaceHolder("FILENAME").Required().String()
	kingpin.Parse()
	api, err := godig.YAMLAuth(*login)
	if err != nil {
		log.Fatalf("Authentication error %v\n", err)
	}
}
