package addressfixer

import "log"

//Fix updates a supporter record using SmartyStreets.
//See https://smartystreets.com/docs/sdk/go
func Fix(c1 chan []Supporter, c2 chan []Supporter, c3 chan Mod) {
	defer close(c2)
	defer close(c3)
	var count int32
	count = 0
	for a := range c1 {
		log.Printf("Fix: offset %7d, got %v", count, len(a))
		var t []Supporter
		for _, r := range a {
			if len(r.Country) == 0 {
				m := Mod{
					Key:   r.Key,
					Field: "Country",
					Old:   r.Country,
					New:   "US"}
				r.Country = "US"
				c3 <- m
			}
			if len(r.Zip) == 0 {
				r.Zip = "78757"
				m := Mod{
					Key:   r.Key,
					Field: "Zip",
					Old:   "(empty)",
					New:   r.Zip}
				c3 <- m
			}
			t = append(t, r)
		}
		c2 <- t
		log.Printf("Fix: offset %7d, sent %v", count, len(t))
		count = count + int32(len(t))
	}
	log.Printf("Fix: done, %v records\n", count)
}
