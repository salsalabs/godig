package active

import (
	"log"
	"strings"

	"github.com/salsalabs/godig"
	"github.com/salsalabs/godig/cmd/addressfixer"
)

//Fix updates a supporter record using SmartyStreets.
func Fix(c1 chan addressfixer.Supporter, c2 chan addressfixer.Supporter, c3 chan addressfixer.Mod) {
	defer close(c2)
	defer close(c3)
	for s := range c1 {
		log.Printf("Active Fix: %+v\n", s)
		m := addressfixer.Mod{
			Key:   s.Key,
			Field: "None",
			Old:   "Old-None",
			New:   "New-None",
		}
		c2 <- s
		c3 <- m
	}
}

//Finish accepts a supporter record at the end of the processing chain.
//This could be saving the record to disk.  It could also be a sink.
func Finish(t *godig.Table, c chan addressfixer.Supporter) {
	for s := range c {
		log.Printf("Active Finish: %+v\n", s)
		args := []string{
			"Street=" + s.Street,
			"Street_2=" + s.Street2,
			"City=" + s.City,
			"State=" + s.State,
			"Zip=" + s.Zip}
		a := strings.Join(args, "&")
		err := t.Save(s.Key, a)
		if err != nil {
			panic(err)
		}
		log.Printf("Active Finish: saved %v\n", s.Key)
	}
}
