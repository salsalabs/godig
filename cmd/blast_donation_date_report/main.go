// Application to attribute donations to email blasts, then
// create a report.  Attribution is via the "email_blast" tags
// that Salsa adds for any donation attributable to a blast.
package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	godig "github.com/salsalabs/godig/pkg"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

const outFile = "blast_donation.csv"

//Fields contains the contents to return.
type Fields struct {
	Tag             string
	EmailBlastKey   string `json:"email_blast_KEY"`
	Subject         string `json:"Subject"`
	DateCreated     string `json:"Date_Created"`
	DonationKEY     string `json:"donation_KEY"`
	TransactionDate string `json:"Transaction_Date"`
	TransactionType string `json:"Transaction_Type"`
	Result          string `json:"Result"`
	Amount          string `json:"Amount"`
}

//Stats contains an email blast and some statistic donations.
type Stats struct {
	EmailBlastKey string
	Subject       string
	DateCreated   string
	Count         int
	Min           float64
	Max           float64
	Sum           float64
	Avg           float64
}

//FieldMap is a mpa of email blast keys and some donation stats.
type FieldMap map[string]*Stats

//All reads all of the records and sends them to a Fields channel.
//parses the buffer for records then outputs them to cout.
func All(t *godig.Table, crit string, cout chan Fields) {
	offset := int32(0)
	count := 500
	for count > 0 {
		log.Printf("All: %v offset %6d\n", t.Name, offset)
		if offset > 0 && offset%5000 == 0 {
			log.Printf("All: %v offset %6d\n", t.Name, offset)
		}
		var a []Fields
		err := t.LeftJoin(offset, count, crit, &a)
		if err != nil {
			log.Fatalf("All: %v offset %6d %v\n", t.Name, offset, err)
			close(cout)
			return
		}
		count = len(a)
		if count == 0 {
			log.Printf("All: %v offset %6d, done\n", t.Name, offset)
			close(cout)
			return
		}
		for _, r := range a {
			cout <- r
		}
		offset = offset + int32(count)
	}
}

//Use reads Fields records from a channel and accumulates
//statistical info by email blast.
func Use(cin chan Fields, stats FieldMap) {
	for r := range cin {
		v, _ := strconv.ParseFloat(r.Amount, 64)
		_, ok := stats[r.EmailBlastKey]
		if !ok {
			s := Stats{EmailBlastKey: r.EmailBlastKey, Subject: r.Subject}
			const form = "Mon Jan 02 2006 15:04:05 GMT-0700 (MST)"
			t, _ := time.Parse(form, r.DateCreated)
			s.DateCreated = t.Format("2006-01-02")
			stats[r.EmailBlastKey] = &s
		}
		x, _ := stats[r.EmailBlastKey]
		x.Count = x.Count + 1
		if x.Min == 0.0 {
			x.Min = v
		} else {
			x.Min = math.Min(x.Min, v)
		}
		x.Max = math.Max(x.Max, v)
		x.Sum = x.Sum + v
		x.Avg = x.Sum / float64(x.Count)
	}
}

//Mainline.  Find supporters and display some info about each.
func main() {
	var (
		cpath      = kingpin.Flag("login", "YAML file of credentials for Salsa Classic API").PlaceHolder("FILENAME").Required().String()
		crit       = kingpin.Flag("criteria", "Search for records matching this criteria").PlaceHolder("CRITERIA").String()
		apiVerbose = kingpin.Flag("apiVerbose", "Show responses from Salsa.  Can be very noisy.").Bool()
	)
	kingpin.Parse()

	a, err := godig.YAMLAuth(*cpath)
	if err != nil {
		log.Fatalf("Authentication error: %+v\n", err)
	}
	a.Verbose = *apiVerbose

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

	stats := make(FieldMap)
	tableName := strings.Join(clauses, "")
	t := a.NewTable(tableName)

	c := make(chan Fields, 100)
	var wg sync.WaitGroup

	log.Println("Main: start")
	wg.Add(1)
	go func(w *sync.WaitGroup) {
		defer w.Done()
		Use(c, stats)
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

	f, err := os.Create(outFile)
	if err != nil {
		log.Fatalf("%v, %s\n", err, outFile)
	}
	defer f.Close()
	w := csv.NewWriter(f)
	if err != nil {
		log.Fatalf("%v, %s\n", err, outFile)
	}
	headers := []string{
		"EmailBlastKey",
		"Subject",
		"Date",
		"Count",
		"Min",
		"Max",
		"Avg",
		"Sum",
	}
	w.Write(headers)
	for _, x := range stats {
		row := []string{
			x.EmailBlastKey,
			x.Subject,
			x.DateCreated,
			fmt.Sprintf("%d", x.Count),
			fmt.Sprintf("%.2f", x.Min),
			fmt.Sprintf("%.2f", x.Max),
			fmt.Sprintf("%.2f", x.Avg),
			fmt.Sprintf("%.2f", x.Sum),
		}
		w.Write(row)
	}
	w.Flush()

	log.Printf("Main: done, results in %s", outFile)
}
