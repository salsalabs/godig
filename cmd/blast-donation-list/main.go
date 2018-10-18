//TagTagData reads all tags and all tag data.
package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/salsalabs/godig"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

//form is the date format.
const form = "Mon Jan 02 2006 15:04:05 GMT-0700 (MST)"

//Fields contains the contents to return.
type Fields struct {
	Tag             string
	EmailBlastKey   string `json:"email_blast_KEY"`
	Subject         string
	DateCreated     string `json:"Date_Created"`
	DonationKEY     string `json:"donation_KEY"`
	TransactionDate string `json:"Transaction_Date"`
	TransactionType string `json:"Transaction_Type"`
	Result          string
	Amount          string
}

//All reads all of the records and sends them to a Fields channel.
//parses the buffer for records then outputs them to cout.
func All(t *godig.Table, crit string, cout chan Fields) {
	offset := int32(0)
	count := 500
	for count > 0 {
		log.Printf("All: offset %6d\n", offset)
		if offset > 0 && offset%5000 == 0 {
			log.Printf("All: offset %6d\n", offset)
		}
		var a []Fields
		err := t.LeftJoin(offset, count, crit, &a)
		if err != nil {
			log.Fatalf("All: offset %6d %v\n", offset, err)
			close(cout)
			return
		}
		count = len(a)
		if count == 0 {
			log.Printf("All: offset %6d, done\n", offset)
			close(cout)
			return
		}
		for _, r := range a {
			cout <- r
		}
		offset = offset + int32(count)
	}
}

//Use reads Fields records from a channel and displays them.
func Use(cin chan Fields, w *csv.Writer) {
	w.Write(strings.Split("EmailBlastKey,Subject,DateCreated,DonationKey,TransactionDate,TransactionType,Result,Amount", ","))
	for r := range cin {
		var a []string
		a = append(a, r.EmailBlastKey)
		a = append(a, r.Subject)
		t, _ := time.Parse(form, r.DateCreated)
		s := t.Format("2006-01-02")
		a = append(a, s)
		a = append(a, r.DonationKEY)
		t, _ = time.Parse(form, r.TransactionDate)
		s = t.Format("2006-01-02")
		a = append(a, s)
		a = append(a, r.TransactionType)
		a = append(a, r.Result)
		a = append(a, r.Amount)
		err := w.Write(a)
		if err != nil {
			fmt.Printf("Use: error %v writing %v\n", err, a)
		}
	}
}

//Mainline.  Find supporters and display some info about each.
func main() {
	cpath := kingpin.Flag("credentials", "YAML file containing credentials for Salsa Classic API").PlaceHolder("FILENAME").Required().String()
	crit := kingpin.Flag("criteria", "Search for records matching this criteria").PlaceHolder("CRITERIA").String()
	kingpin.Parse()

	a, err := godig.YAMLAuth(*cpath)
	if err != nil {
		log.Fatalf("Authentication error: %+v\n", err)
	}

	// It turns out that conditions will not only join on table1(table1 primary key)table2,
	// they will also join on something like table1(table1.whatever=table2.whatever)table2.
	// This table name does the necessary joins to retrieve donations tagged by email blasts.
	//
	// SELECT eb.email_blast_KEY,
	//          eb.Subject,
	//          count(d.donation_KEY),
	//          min(d.amount),
	//          max(d.amount),
	//          avg(d.amount),
	//          sum(d.amount)
	// FROM email_blast eb,
	// JOIN tags t
	//     ON t.tag = eb.email_blast_KEY
	// JOIN tag_data td
	//     ON td.tag_KEY = td.tag_KEY
	// JOIN donation d
	//     ON td.table_KEY = donation.donation_KEY
	// WHERE td.database_table_KEY = 45;

	clauses := []string{"tag(tag_KEY)",
		"tag_data(tag.tag=email_blast_KEY)",
		"email_blast(tag_data.table_KEY=donation_KEY)",
		"donation"}
	cond := "tag_data.database_table_KEY=45&condition=tag.prefix=email_blast"

	tableName := strings.Join(clauses, "")
	t := a.NewTable(tableName)

	c := make(chan Fields, 100)
	var wg sync.WaitGroup

	f, err := os.Create("results.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	writer := csv.NewWriter(f)

	log.Println("Main: start")
	wg.Add(1)
	go func(w *sync.WaitGroup) {
		defer w.Done()
		Use(c, writer)
	}(&wg)
	log.Println("Main: Use started")
	wg.Add(1)
	go func(w *sync.WaitGroup) {
		defer w.Done()
		if len(*crit) != 0 {
			cond = cond + "&condition=" + *crit
		}
		All(&t, cond, c)
	}(&wg)
	log.Println("Main: All started")

	log.Println("Main: waiting...")
	wg.Wait()
}
