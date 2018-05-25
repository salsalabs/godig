package addressfixer

import (
	"github.com/salsalabs/godig"
	"log"
	"sync"
)

//ReadAll implements Reader.  Reads all supporters for a criteria
//then passes arrays of supporter JSON downstream.
func ReadAll(t *godig.Table, crit string, c1 chan []Supporter, id int, m *sync.Mutex, c2 chan int32, done chan bool) {
	count := 500
	ok := true

	for ok {
		m.Lock()
		offset, ok := <-c2
		m.Unlock()
		if !ok {
			break
		}
		log.Printf("ReadAll: id %2d popped %7d", id, offset)
		var a []Supporter
		err := t.Many(int(offset), count, crit, &a)
		if err != nil {
			log.Fatalf("ReadAll: id %2d offset %6d %v\n", id, offset, err)
		}
		count = len(a)
		if count == 0 {
			log.Printf("ReadAll: id %2d offset %7d, done\n", id, offset)
			done <- true
		} else {
			c1 <- a
			offset = offset + int32(count)
			// c2 <- offset
			// log.Printf("ReadAll: id %2d sent %3d pushed offset %7d", id, len(a), offset)
		}
	}
}
