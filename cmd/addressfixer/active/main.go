package active

import (
	"log"
	"strings"

	"github.com/salsalabs/godig"
	"github.com/salsalabs/godig/cmd/addressfixer"
)

//Finish accepts a supporter record at the end of the processing chain.
//This could be saving the record to disk.  It could also be a sink.
func Finish(t *godig.Table, c chan addressfixer.Supporter) {
	for s := range c {
		log.Printf("Active Finish: %+v\n", s)
		args := []string{
			"Street=" + strings.Replace(s.Street, " St.", "", -1),
			"Street_2=",
			"City=" + strings.Replace(s.City, "Ox", "", -1),
			"State=",
			"Zip=" + strings.Replace(s.Zip, "-12", "", -1)}
		a := strings.Join(args, "&")
		err := t.Save(s.Key, a)
		if err != nil {
			panic(err)
		}
		log.Printf("Active Finish: saved %v\n", s.Key)
	}
}
