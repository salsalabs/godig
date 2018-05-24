package addressfixer

import (
	"log"

	"github.com/salsalabs/godig"
)

//ReadAll implements Reader.  Reads all supporters for a criteria
//then passes arrays of supporter JSON downstream.
func ReadAll(t *godig.Table, crit string, c chan []Supporter) {
	offset := 0
	count := 500

	for count > 0 {
		if offset > 0 && offset%5000 == 0 {
			log.Printf("ReadAll: %v offset %6d\n", t.Name, offset)
		}
		var a []Supporter
		err := t.Many(offset, count, crit, &a)
		if err != nil {
			log.Fatalf("ReadAll: %v offset %6d %v\n", t.Name, offset, err)
			return
		}
		count = len(a)
		if count == 0 {
			log.Printf("ReadAll: offset %7d, done\n", offset)
			close(c)
		} else {
			c <- a
			log.Printf("ReadAll: offset %7d, sent %v\n", offset, len(a))
			offset = offset + count
		}
	}
}
