package main

//Application to accept a CSV file of last-name, first-name,
//the display supporter_KEY, names and email.

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"

	godig "github.com/salsalabs/godig/pkg"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

// Supporter holds the data that we want from this supporter.
type Supporter struct {
	SupporterKEY string `json:"supporter_KEY"`
	FirstName    string `json:"First_Name,omitempty"`
	LastName     string `json:"Last_Name,omitempty"`
	Email        string `json:"Email,omitEmpty"`
}

//Mainline.  Find supporters and display some info about each.
func main() {
	cpath := kingpin.Flag("login", "YAML file containing credentials for Salsa Classic API").Required().String()
	csvFile := kingpin.Flag("csv-file", "Search for records in this file").Required().String()
	apiVerbose := kingpin.Flag("apiVerbose", "Show all interactions with the server.  Verrry noisy").Bool()
	kingpin.Parse()

	a, err := godig.YAMLAuth(*cpath)
	if err != nil {
		log.Fatalf("Authentication error: %+v\n", err)
	}
	a.Verbose = *apiVerbose
	t := a.Supporter()

	f, err := os.Open(*csvFile)
	if err != nil {
		log.Fatalf("%v on %v", err, *csvFile)
	}
	defer f.Close()
	r := csv.NewReader(f)
	all, err := r.ReadAll()

	f2, err := os.Create("names_and_emails.csv")
	w := csv.NewWriter(f2)
	h := []string{
		"SupporterKEY",
		"FirstName",
		"LastName",
		"Email",
	}
	err = w.Write(h)
	if err != nil {
		log.Fatalf("%s, %s", "names_and_emails.csv", err)
	}
	defer w.Flush()
	defer w.Close()
	var s []Supporter

	// Each row is a string slice of (Last_Name, First_Name)
	for _, row := range all {
		lastName := row[0]
		firstName := row[1]

		criteria := fmt.Sprintf("First_Name=%s&condition=last_Name=%s", firstName, lastName)
		err = t.Many(int32(0), 500, criteria, &s)
		if err != nil {
			log.Fatalf("'%s %s', %v\n", firstName, lastName, err)
		}
		if len(s) == 0 {
			fmt.Printf("'%s %s' not found\n", firstName, lastName)
		} else {
			for _, record := range s {
				fmt.Printf("'%s %s' %-8s %s\n", firstName, lastName, record.SupporterKEY, record.Email)
				a := []string{
					record.SupporterKEY,
					record.FirstName,
					record.LastName,
					record.Email,
				}
				err = w.Write(a)
				if err != nil {
					log.Fatalf("%s writing to CSV\n", err)
				}
			}
		}
	}
}
