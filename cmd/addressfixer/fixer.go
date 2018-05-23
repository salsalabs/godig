package addressfixer

import "log"

//Fix updates a supporter record using SmartyStreets.
//See https://smartystreets.com/docs/sdk/go
func Fix(c1 chan Supporter, c2 chan Supporter, c3 chan Mod) {
	defer close(c2)
	defer close(c3)
	for s := range c1 {
		log.Printf("Active Fix: %+v\n", s)
		m := Mod{
			Key:   s.Key,
			Field: "None",
			Old:   "Old-None",
			New:   "New-None",
		}
		c2 <- s
		c3 <- m
	}
}
