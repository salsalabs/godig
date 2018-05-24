package addressfixer

import "log"

//Audit record changes to a supporter record.
func Audit(c chan Mod) {
	var count int32
	count = 0
	for a := range c {
		log.Printf("Audit:   %+v\n", a)
		count = count + 1
	}
	log.Printf("Audit:   done, %d records\n", count)
}
