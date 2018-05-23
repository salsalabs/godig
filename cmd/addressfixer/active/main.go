package active

import (
	"log"

	"github.com/salsalabs/godig"
	"github.com/salsalabs/godig/cmd/addressfixer"
)

//All implements Reader.  Reads all supporters for a criteria
//then passes arrays of supporter JSON downstream.
func All(t *godig.Table, crit string, c chan []byte) {
	offset := 0
	count := 500
	for count > 0 {
		log.Printf("All: %v offset %6d\n", t.Name, offset)
		if offset > 0 && offset%5000 == 0 {
			log.Printf("All: %v offset %6d\n", t.Name, offset)
		}
		var a []Fields
		err := t.Many(offset, count, crit, &a)
		if err != nil {
			log.Fatalf("All: %v offset %6d %v\n", t.Name, offset, err)
			close(cout)
			return
		}
		count = len(a)
		if count == 0 {
			log.Printf("All: %v offset %6d, done\n", t.Name, offset)
			close(cout)
			return
		}
		for _, r := range a {
			cout <- r
		}
		offset = offset + count
	}
}

//Split accepts a buffer and splits it into supporter records.
//Supporter records then flow through the channel.
func Split(c1 chan []byte, c2 chan addressfixer.Supporter) {
}

//Audit record changes to a supporter record.
func Audit(c chan addressfixer.Mod) {
}

//Fix updates a supporter record using SmartyStreets.
func Fix(c1 chan addressfixer.Supporter, c2 chan addressfixer.Supporter, c3 chan addressfixer.Mod) {
}

//Save accepts a supporter record at the end of the processing chain.
//This could be saving the record to disk.  It could also be a sink.
func Save(c1 chan addressfixer.Supporter) {
}
