//ChangeFirstName demonstrates changing a single field in a supporter record.
//You provide the supporter key and the new First_Name field.
package main

import (
	"fmt"
	"log"

	"github.com/salsalabs/godig"
	"gopkg.in/alecthomas/kingpin.v2"
)

//Fields are retrieved from the supporter record.
type Fields struct {
	FirstName string `json:"First_Name"`
}

//Name is the table name to read.
const Name = "supporter"

//FirstName retrieves the first name from the supporter record.
func FirstName(t godig.Table, k string) (string, error) {
	var f Fields
	err := t.One(k, &f)
	if err != nil {
		return "", err
	}
	return f.FirstName, nil
}

func main() {
	cpath := kingpin.Flag("credentials", "YAML file containing credentials for Salsa Classic API").PlaceHolder("FILENAME").Required().String()
	supporterKey := kingpin.Flag("supporter_KEY", "Modify this supporter").Required().String()
	firstName := kingpin.Flag("first_name", "New value of First_Name").Required().String()
	kingpin.Parse()

	a, err := godig.YAMLAuth(*cpath)
	if err != nil {
		log.Fatalf("Authentication error %v using %v\n", err, *cpath)
	}

	t := a.Supporter()

	before, err := FirstName(t, *supporterKey)
	log.Printf("Initial first name is '%s'\n", before)

	s := fmt.Sprintf("First_Name=%s", *firstName)
	log.Println("Changing name to", s)
	err = t.Save(*supporterKey, s)
	if err != nil {
		log.Fatalf("Save error %+v on %s, key %s, value %s\n", err, t.Name, *supporterKey, s)
	}
	after, err := FirstName(t, *supporterKey)
	log.Printf("Modified first name is '%s'\n", after)
}
