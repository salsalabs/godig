package addressfixer

import "log"

//Audit record changes to a supporter record.
func Audit(c chan Mod) {
	var count int32
	count = 0
	for a := range c {
		if len(a.Old) == 0 {
			a.Old = "(empty)"
		}
		log.Printf("Audit:   Key: %-8s Field: %-10s  New: %-10s  Old: %-20s Reason: %s\n", a.Key, a.Field, a.New, a.Old, a.Reason)
		count = count + 1
	}
	log.Printf("Audit:   done, %d records\n", count)
}
