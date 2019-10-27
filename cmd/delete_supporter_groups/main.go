//Query for a record, see JSON.
package main

import (
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/salsalabs/godig/pkg"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

const (
	WhackCount = 10
	DriveCount = 5
)

func drive(id int, t *godig.Table, offsets chan int32, c chan string, done chan bool) {
	fmt.Printf("drive-%02d: start\n", id)
	for offset := range offsets {
		b, err := t.ManyMap(offset, 500, "")
		fmt.Printf("drive-%02d: %7d\n", id, offset)
		if err != nil {
			panic(err)
		}
		for _, r := range b {
			c <- r["supporter_groups_KEY"]
		}
	}
	close(c)
	fmt.Printf("drive-%02d: end", id)
	done <- true
}

func whack(id int, t *godig.Table, c chan string, done chan bool) {
	count := int32(0)
	fmt.Printf("whack-%02d: start\n", id)
	for k := range c {
		var ds godig.DeleteStatus
		t.Delete(k, &ds)
		count++
		if count%1000 == 0 {
			fmt.Printf("whack-%02d: %7d\n", id, count)
		}
	}
	fmt.Printf("whack-%02d: end", id)
	done <- true
}
func watch(x int, done chan bool) {
	fmt.Println("watch: start")
	for x > 0 {
		fmt.Printf("watch: watiing for %d\n", x)
		_, _ = <-done
		x--
	}
	fmt.Println("watch: end")
}

func main() {
	cpath := kingpin.Flag("login", "YAML file containing login credentials for Salsa Classic API").PlaceHolder("FILENAME").Required().String()
	kingpin.Parse()
	api, err := godig.YAMLAuth(*cpath)
	if err != nil {
		log.Fatalf("Authentication error %v\n", err)
	}
	t := api.NewTable("supporter_groups")
	var wg sync.WaitGroup
	s, _ := t.Count("")
	x, _ := strconv.ParseInt(s, 10, 64)
	limit := int32(x)
	fmt.Printf("main: processing %8v\n", limit)
	c := make(chan string, 1000)
	offsets := make(chan int32, 1000)
	done := make(chan bool, 20)

	for i := 0; i < 10; i++ {
		go (func(i int, wg *sync.WaitGroup, t *godig.Table, c chan string, done chan bool) {
			wg.Add(1)
			whack(i, t, c, done)
			wg.Done()
		})(i+1, &wg, &t, c, done)
	}
	for i := 0; i < 5; i++ {
		go (func(i int, wg *sync.WaitGroup, t *godig.Table, offsets chan int32, c chan string, done chan bool) {
			wg.Add(1)
			drive(i, t, offsets, c, done)
			wg.Done()
		})(i+1, &wg, &t, offsets, c, done)
	}
	var j int32
	for j = 0; j < limit; j += 500 {
		offsets <- j
		fmt.Printf("main: %7d pushed\n", j)
	}
	fmt.Println("main: waiting")
	watch(WhackCount+DriveCount, done)
	wg.Wait()
}
