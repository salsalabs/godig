//Demonstrate parsing MySQL Timestap.  Sample: Sun Nov 03 2013 08:59:25 GMT-0500 (EST)
package main

import (
	"fmt"
	"time"
)

func main() {
	date := "Sun Nov 03 2013 08:59:25 GMT-0500 (EST)"
	const form = "Mon Jan 02 2006 15:04:05 GMT-0700 (MST)"
	t, err := time.Parse(form, date)
	if err != nil {
		fmt.Printf("Time parse error %v on %v\n", err, date)
	}
	fmt.Printf("%+v\n", t)
	year := t.Year()
	fmt.Println(year)
}
