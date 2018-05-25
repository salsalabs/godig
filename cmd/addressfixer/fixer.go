package addressfixer

import "log"
import "sync"

//Fix updates a supporter record using SmartyStreets.
//See https://smartystreets.com/docs/sdk/go

func Fix(c1 chan []Supporter, c2 chan []Supporter, c3 chan Mod, m *sync.Mutex, id int) {
	var offset int32
	offset = 0
	totalSkipped := 0
	totalSent := 0
	ok := true

	for ok {
		m.Lock()
		a, ok := <-c1
		m.Unlock()

		if !ok {
			break
		}
		if len(a) > 0 {
			f := a[0].Key
			log.Printf("Fix:     id %2d %3d records starting at %7s", id, len(a), f)
		}

		var t []Supporter
		skipped := 0
		sent := 0

		for _, r := range a {
			var mods []Mod
			// Get country code for long country name.
			// Do this before jumping into the postal code lookup.
			mods, err := RestCountries(&r, mods)
			if err != nil {
				log.Printf("Fix:    id %2d %v", id, err)
			} else {
				mods, err := Zippo(r, mods)
				if err != nil {
					log.Printf("Fix:     id %2d %v\n", id, err)
				} else {
					if len(mods) != 0 {
						for _, m := range mods {
							c3 <- m
						}
						t = append(t, r)
					} else {
						skipped = skipped + 1
					}
				}
			}
		}
		if len(t) != 0 {
			c2 <- t
			sent = sent + len(t)
		}
		log.Printf("Fix:     id %2d offset %7d, skipped %v, sent %v", id, offset, skipped, sent)
		totalSent = totalSent + sent
		totalSkipped = totalSkipped + skipped
		offset = offset + int32(len(a))
	}
	log.Printf("Fix:     id %2d done, %v records in, sent %v, skipped %v\n", id, offset, totalSent, totalSkipped)
}
