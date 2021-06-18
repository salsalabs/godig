package main

// These work in the browser:
// https://org.salsalabs.com/api/getObjects.sjs
// ?json
// &object=supporter
// &include=Email,Receive_Email
// &condition=Email%20IS%20NOT%20EMPTY
// &condition=Email%20LIKE%20%25@%25.%25

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	godig "github.com/salsalabs/godig/pkg"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

const (
	activeCriteria                 = "Email IS NOT EMPTY&condition=Email LIKE %@%.%&condition=Receive_Email > 0"
	addedLast24HoursCriteria       = "Date_Created>=%s&condition=Date_Created<=%s"
	activeAddedLast24HoursCriteria = "Email IS NOT EMPTY&condition=Email LIKE %%@%%.%%&condition=Receive_Email > 0&condition=Date_Created>=%s&condition=Date_Created<=%s"
	unsubLast24HoursCriteria       = "Unsubscribe_Date>=%s&condition=Unsubscribe_Date<=%s"
)

// //Fields are retrieved from the supporter record.
// type Fields struct {
// 	SupporterKey string `json:"supporter_KEY"`
// 	DateCreated  string `json:"Date_Created,omitempty"`
// }

// //All reads from Salsa via the API.  If the criteria is not empty,
// //then records that match that criteria are returned.  Each read
// //parses the buffer for records then outputs them to cout.
// func All(t *godig.Table, crit string, cout chan Fields) {
// 	offset := int32(0)
// 	count := 500
// 	for count > 0 {
// 		log.Printf("All: %v offset %6d\n", t.Name, offset)
// 		if offset > 0 && offset%5000 == 0 {
// 			log.Printf("All: %v offset %6d\n", t.Name, offset)
// 		}
// 		var a []Fields
// 		err := t.Many(offset, count, crit, &a)
// 		if err != nil {
// 			log.Fatalf("All: %v offset %6d %v\n", t.Name, offset, err)
// 			close(cout)
// 			return
// 		}
// 		count = len(a)
// 		if count == 0 {
// 			log.Printf("All: %v offset %6d, done\n", t.Name, offset)
// 			close(cout)
// 			return
// 		}
// 		for _, r := range a {
// 			cout <- r
// 		}
// 		offset = offset + int32(count)
// 	}
// }

// //Use accepts Fields records from a channel and displays them.
// func Use(cin chan Fields) {
// 	for f := range cin {
// 		ts := godig.SalsaTimestamp{}
// 		ct := &ts
// 		err := ct.UnmarshalJSON([]byte(f.DateCreated))
// 		if err != nil {
// 			panic(err)
// 		}
// 		b, err := ct.MarshalDate()
// 		if err != nil {
// 			panic(err)
// 		}
// 		v := string(b)

// 		log.Printf("%s %s\n", f.SupporterKey, v)
// 	}
// }

//SalsaDate accepts a time and returns a Salsa-formatted date.
func SalsaDate(t time.Time) string {
	f := "2006-01-02 15:04:05"
	s := t.Format(f)
	return s
}

//SalsaToday returns the date for today as a Salsa date.
func SalsaToday() string {
	n := time.Now()
	return SalsaDate(n)
}

//SalsaYesterday returns the date for yesterday as a Salsa date.
func SalsaYesterday() string {
	n := time.Now()
	y := n.AddDate(0, 0, -1)
	return SalsaDate(y)
}

//BriefToday returns only the date for today.
func BriefToday() string {
	n := time.Now()
	return n.Format("2006-01-02")
}

//CountThese accepts a table and a criteria and returns the number of
//matching records.  Errors cause panics.
func CountThese(t godig.Table, s string) string {
	count, err := t.Count(s)
	if err != nil {
		panic(err)
	}
	return count
}

//Mainline.  Find supporters and display some info about each.
func main() {
	cpath := kingpin.Flag("login", "YAML file containing credentials for Salsa Classic API").PlaceHolder("FILENAME").Required().String()
	kingpin.Parse()

	a, err := godig.YAMLAuth(*cpath)
	if err != nil {
		log.Fatalf("Authentication error: %+v\n", err)
	}
	t := a.Supporter()
	u := a.Unsubscribe()
	yday := SalsaYesterday()
	today := SalsaToday()
	briefToday := BriefToday()

	addedLast24Hours := CountThese(t, fmt.Sprintf(addedLast24HoursCriteria, yday, today))
	activeAddedLast24Hours := CountThese(t, fmt.Sprintf(activeAddedLast24HoursCriteria, yday, today))
	unsubLast24Hours := CountThese(u, fmt.Sprintf(unsubLast24HoursCriteria, yday, today))
	activeSupporters := CountThese(t, activeCriteria)

	filename := "/Users/aleonard/Google Drive/My Drive/Clients/Consortium News/daily_readings.csv"
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	parts := []string{
		briefToday,
		addedLast24Hours,
		activeAddedLast24Hours,
		unsubLast24Hours,
		activeSupporters,
	}
	row := strings.Join(parts, ",")
	text := fmt.Sprintf("%s\n", row)
	_, err = f.WriteString(text)
	if err != nil {
		panic(err)
	}
}
