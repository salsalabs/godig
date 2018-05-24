package addressfixer

import "log"

//Split accepts a buffer and splits it into supporter records.
//Supporter records then flow through the channel.
func Split(c1 chan []Supporter, c2 chan []Supporter, chunkSize int) {
	defer close(c2)
	var count int32
	count = 0
	for a := range c1 {
		log.Printf("Split: received %v\n", len(a))
		for i := 0; i < len(a); i += chunkSize {
			j := i + chunkSize
			if j > len(a) {
				j = len(a)
			}
			var b []Supporter
			for k := i; k < j; k++ {
				b = append(b, a[k-i])
			}
			c2 <- b
			count = count + int32(len(b))
			log.Printf("Split: offset %7d, sent %v\n", count, len(b))
		}
	}
	log.Printf("Split: done, %v records\n", count)
}
