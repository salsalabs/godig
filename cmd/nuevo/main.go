//Nuevo demonstrates the Salsa API
package main

import (
	"log"

	"github.com/salsalabs/godig"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

//Fields defines the supporter fields to retrieve.
type Fields struct {
	FirstName string `json:"First_Name"`
	LastName  string `json:"Last_Name"`
	Email     string `json:"Email"`
}

func main() {
	cpath := kingpin.Flag("credentials", "YAML file containing credentials for Salsa Classic API").PlaceHolder("FILENAME").Required().String()
	kingpin.Parse()

	cred, err := godig.Credentials(*cpath)
	if err != nil {
		log.Fatalf("Credentials error %v on %v\n", err, *cpath)
	}
	api := godig.NewAPI()
	err = api.Authenticate(cred)
	if err != nil {
		log.Fatalf("Authentication error %v\n", err)
	}

	s := api.NewTable("supporter")
	// This is a string because that's what Salsa returns.
	key := "58534244"
	var f Fields
	err = s.One(key, &f)
	if err != nil {
		log.Fatalf("Get error %v on table %v key %v\n", err, s.Name, key)
	}
	log.Printf("Supporter record is %+v\n", f)
}
