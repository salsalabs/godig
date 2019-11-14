//Return and display an organization.  The API always returns the org
//with key zero when a client calls this method.  We'll just leave that out...
package main

import (
	"fmt"
	"log"

	godig "github.com/salsalabs/godig/pkg"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	cpath := kingpin.Flag("login", "YAML file containing credentials for Salsa Classic API").PlaceHolder("FILENAME").Required().String()
	kingpin.Parse()
	api, err := godig.YAMLAuth(*cpath)
	if err != nil {
		log.Fatalf("Authentication error %v\n", err)
	}
	t := api.Org()
	var orgs []godig.Organization
	crit := "organization_KEY>0"
	err = t.Many(int32(0), 2, crit, &orgs)
	if err != nil {
		log.Fatalf("Read error %v\n", err)
	}
	if len(orgs) == 0 {
		log.Fatalf("No organizations found for your credentials\n")
	}
	fmt.Printf("OrganizationKEY:    %v\n", orgs[0].OrganizationKEY)
	fmt.Printf("Name:               %v\n", orgs[0].Name)
	fmt.Printf("DateCreated:        %v\n", godig.ShortDate(orgs[0].DateCreated))
	fmt.Printf("LastModified:       %v\n", godig.ShortDate(orgs[0].LastModified))
	fmt.Printf("PRIVATEDateCreated: %v\n", godig.ShortDate(orgs[0].PRIVATEDateCreated))
	fmt.Printf("Type:               %v\n", orgs[0].Type)
	fmt.Printf("Status:             %v\n", orgs[0].Status)
	fmt.Printf("BaseURL:            %v\n", orgs[0].BaseURL)
	fmt.Printf("SecureURL:          %v\n", orgs[0].SecureURL)
	fmt.Printf("ClosedDate:         %v\n", godig.ShortDate(orgs[0].ClosedDate))

}
