package addressfixer

import "log"

//Split accepts a buffer and splits it into supporter records.
//Supporter records then flow through the channel.
func Split(c1 chan []Supporter, c2 chan Supporter) {
	defer close(c2)
	for a := range c1 {
		log.Printf("Split: received %v supporters\n", len(a))
		for _, r := range a {
			c2 <- r
		}
	}
}
