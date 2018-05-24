package addressfixer

//Chunk accepts a buffer and splits it into supporter records.
//Supporter records then flow through the channel.
func Chunk(c1 chan []Supporter, c2 chan []Supporter, chunkSize int) {
	defer close(c2)
	var offset int32
	offset = 0
	for a := range c1 {
		for i := 0; i < len(a); i += chunkSize {
			j := i + chunkSize
			if j > len(a) {
				j = len(a)
			}
			var b []Supporter
			for k := i; k < j; k++ {
				b = append(b, a[k])
			}
			c2 <- b
			offset = offset + int32(len(b))
		}
		//log.Printf("Chunk:   offset %7d, sent %v\n", offset, len(a))
	}
	//log.Printf("Chunk:   done, %v records\n", offset)
}
