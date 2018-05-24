package addressfixer

import "log"

//Fix updates a supporter record using SmartyStreets.
//See https://smartystreets.com/docs/sdk/go
func Fix(c1 chan []Supporter, c2 chan []Supporter, c3 chan Mod) {
	defer close(c2)
	defer close(c3)
	var offset int32
	offset = 0
	totalSkipped := 0
	totalSent := 0
	for a := range c1 {
		var t []Supporter
		skipped := 0
		sent := 0

		for _, r := range a {
			var mods []Mod
			err := Zippo(r, mods)
			if err != nil {
				log.Printf("Fix:     %v on %v\n", err, r.Zip)
			} else {
				if len(mods) != 0 {
					log.Printf("Fix:     Zippo returned %v mods for %v\n", len(mods), r.Zip)
					for _, m := range mods {
						c3 <- m
						log.Printf("send %+v to Audit\n", m)
					}
					t = append(t, r)
				} else {
					skipped = skipped + 1
				}
			}
		}
		if len(t) != 0 {
			c2 <- t
			sent = sent + len(t)
		}
		//log.Printf("Fix:     offset %7d, skipped %v, sent %v", offset, skipped, sent)
		totalSent = totalSent + sent
		totalSkipped = totalSkipped + skipped
		offset = offset + int32(len(a))
	}
	log.Printf("Fix:     done, %v records in, sent %v, skipped %v\n", offset, totalSent, totalSkipped)
}
