//Delete a publish record with the provided database_table_KEY and table_KEY.
package main

import (
	"fmt"
	"log"

	"github.com/salsalabs/godig"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

//Fields contains the fields from a publish record
type Fields struct {
	PublishKey       string `json:"publish_KEY"`
	DatabaseTableKEY string `json:"database_table-KEY"`
	TableKey         string `json:"table_KEY"`
	TemplateKEY      string `json:"template_KEY"`
	DateCreated      string `json:"Date_Created"`
}

func main() {
	cpath := kingpin.Flag("credentials", "YAML file containing credentials for Salsa Classic API").PlaceHolder("FILENAME").Required().String()
	dtk := kingpin.Flag("database_table_KEY", "database table key").PlaceHolder("DTK").Required().String()
	tk := kingpin.Flag("table_KEY", "table key").PlaceHolder("TK").Required().String()
	kingpin.Parse()
	api, err := godig.YAMLAuth(*cpath)
	if err != nil {
		log.Fatalf("Authentication error %v\n", err)
	}
	// Get the publish_KEY for the record.  It's okay if the record does not exist.
	t := api.Publish()
	crit := fmt.Sprintf("database_table_KEY=%v&condition=table_KEY=%v", *dtk, *tk)
	fmt.Printf("delete-publish, criteria is %v\n", crit)
	var b []Fields
	err = t.Many(0, 500, crit, &b)
	if err != nil {
		log.Fatalf("Many error %v\n", err)
	}
	if len(b) == 0 {
		log.Printf("No record match criteria %v\n", crit)
		return
	}
	f := b[0]
	log.Printf("Match found, %+v\n", f)
	_, err = t.Delete(f.PublishKey)
	if err != nil {
		log.Fatalf("Delete error %v\n", err)
	}
	log.Printf("Delete successful\n")
}
