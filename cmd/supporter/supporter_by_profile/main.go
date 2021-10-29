package main

//Application to accept a CSV file of recurring donation
//profile IDs and return supporter information.

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"

	godig "github.com/salsalabs/godig/pkg"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

const outFile = "profiles_and_supporters.csv"
const tableName = "recurring_donation(supporter_KEY)supporter"

// supporter holds the data that we want from this supporter.
type supporter struct {
	ProfileID    string `json:"PROFILEID,omitempty"`
	SupporterKEY string `json:"supporter_KEY"`
	FirstName    string `json:"First_Name,omitempty"`
	LastName     string `json:"Last_Name,omitempty"`
	Email        string `json:"Email,omitempty"`
}

//Mainline.  Find supporters and display some info about each.
func main() {
	cpath := kingpin.Flag("login", "YAML file containing credentials for Salsa Classic API").Required().String()
	csvFile := kingpin.Flag("csv-file", "Search for recurring profiles from this file").Required().String()
	apiVerbose := kingpin.Flag("apiVerbose", "Show all interactions with the server.  Verrry noisy").Bool()
	kingpin.Parse()

	a, err := godig.YAMLAuth(*cpath)
	if err != nil {
		log.Fatalf("Authentication error: %+v\n", err)
	}
	a.Verbose = *apiVerbose
	t := a.NewTable(tableName)

	f, err := os.Open(*csvFile)
	if err != nil {
		log.Fatalf("%v on %v", err, *csvFile)
	}
	defer f.Close()
	r := csv.NewReader(f)
	all, err := r.ReadAll()

	f2, err := os.Create(outFile)
	if err != nil {
		log.Fatalf("%v on %v", err, *csvFile)
	}
	defer f2.Close()

	w := csv.NewWriter(f2)
	h := []string{
		"PROFILEID",
		"SupporterKEY",
		"FirstName",
		"LastName",
		"Email",
	}
	err = w.Write(h)
	if err != nil {
		log.Fatalf("%s, %s", outFile, err)
	}
	defer w.Flush()
	var s []supporter

	// Each row is a string slice of (Last_Name, First_Name)
	for _, row := range all {
		profileID := row[0]

		criteria := fmt.Sprintf("recurring_donation.PROFILEID=%s", profileID)
		err = t.LeftJoin(int32(0), 500, criteria, &s)
		if err != nil {
			log.Fatalf("'%s', %v\n", profileID, err)
		}
		if len(s) == 0 {
			log.Printf("'%s', no supporter found\n", profileID)
		} else {
			for _, record := range s {
				a := []string{
					record.ProfileID,
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
