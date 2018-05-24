package addressfixer

import "log"

//Fix updates a supporter record using SmartyStreets.
//See https://smartystreets.com/docs/sdk/go
func Fix(c1 chan []Supporter, c2 chan []Supporter, c3 chan Mod) {
	defer close(c2)
	defer close(c3)
	for s := range c1 {
		for _, r := range s {
			log.Printf("Active Fix: %+v\n", r)
			r.Country = "US"
			if len(r.Zip) == 0 {
				r.Zip = "78757"
			}
			m := Mod{
				Key:   r.Key,
				Field: "None",
				Old:   "Old-None",
				New:   "New-None"}
			c3 <- m
		}
		c2 <- s
	}
}
