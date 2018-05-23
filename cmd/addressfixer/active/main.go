package active

import (
	"log"

	"github.com/salsalabs/godig/cmd/addressfixer"
)

//Fix updates a supporter record using SmartyStreets.
func Fix(c1 chan addressfixer.Supporter, c2 chan addressfixer.Supporter, c3 chan addressfixer.Mod) {
	defer close(c2)
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
func Finish(c chan addressfixer.Supporter) {
	for s := range c {
		log.Printf("Active Finish: %+v\n", s)
	}
}
