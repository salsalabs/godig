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

	godig "github.com/salsalabs/godig/pkg"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

const outFile = "blast_donation_report.csv"

//Fields contains the contents to return.
type Fields struct {
	Tag           string
	EmailBlastKey string                `json:"email_blast_KEY"`
	DateRequested *godig.SalsaTimestamp `json:"Date_Requested"`
	Subject       string                `json:"Subject"`
	Amount        string                `json:"Amount"`
}

//Stats contains an email blast and some statistic donations.
type Stats struct {
	EmailBlastKey string
	DateRequested *godig.SalsaTimestamp
	Subject       string
	Count         int
	Min           float64
	Max           float64
	Sum           float64
	Avg           float64
}

//All reads all of the records and sends them to a Fields channel.
//parses the buffer for records then outputs them to cout.
func All(t *godig.Table, crit string, cout chan Fields) {
	offset := int32(0)
	count := 500
	for count > 0 {
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
	close(cout)
}

//Store reads stats from a channel and writes them to the CSV file.
func Store(cin chan Stats, outFile string) error {
	f, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	if err != nil {
		return err
	}
	headers := []string{
		"EmailBlastKey",
		"DateRequested",
		"Subject",
		"Count",
		"Min",
		"Max",
		"Avg",
		"Sum",
	}
	w.Write(headers)
	w.Flush()

	for {
		x, ok := <-cin
		if !ok {
			break
		}
		dateRequested := x.DateRequested.Time.Format(godig.DateFormat)

		row := []string{
			x.EmailBlastKey,
			dateRequested,
			x.Subject,
			fmt.Sprintf("%d", x.Count),
			fmt.Sprintf("%.2f", x.Min),
			fmt.Sprintf("%.2f", x.Max),
			fmt.Sprintf("%.2f", x.Avg),
			fmt.Sprintf("%.2f", x.Sum),
		}
		w.Write(row)
		w.Flush()
	}
	return nil
}

//Use reads Fields records from a channel and accumulates
//statistical info by email blast.
func Use(cin chan Fields, cout chan Stats) {
	prevKey := ""
	var s Stats

	for {
		r, ok := <-cin
		if !ok {
			break
		}

		if r.EmailBlastKey != prevKey {
			if s.EmailBlastKey != "" {
				cout <- s
			}
			s = Stats{
				EmailBlastKey: r.EmailBlastKey,
				DateRequested: r.DateRequested,
				Subject:       r.Subject}
			prevKey = r.EmailBlastKey
		}

		a, _ := strconv.ParseFloat(r.Amount, 64)
		s.Count = s.Count + 1
		if s.Min == 0.0 {
			s.Min = a
		} else {
			s.Min = math.Min(s.Min, a)
		}
		s.Max = math.Max(s.Max, a)
		s.Sum = s.Sum + a
		s.Avg = s.Sum / float64(s.Count)
	}

	if s.EmailBlastKey != "" {
		cout <- s
	}
	close(cout)

}

//Mainline.  Find email blasts and display donation stats.
func main() {
	var (
		cpath      = kingpin.Flag("login", "YAML file of login credentials for Salsa Classic API").PlaceHolder("FILENAME").Required().String()
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
	//			eb.Date_Requested,
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
	cond := "tag_data.database_table_KEY=45&condition=tag.prefix=email_blast&donation.RESULT IN (0,-1)"

	tableName := strings.Join(clauses, "")
	t := a.NewTable(tableName)

	cin := make(chan Fields, 100)
	cout := make(chan Stats, 100)
	var wg sync.WaitGroup

	log.Println("Main: start")
	wg.Add(1)
	go func(cin chan Fields, cout chan Stats, w *sync.WaitGroup) {
		defer w.Done()
		Use(cin, cout)
	}(cin, cout, &wg)
	log.Println("Main: Use started")

	wg.Add(1)
	go func(cout chan Stats, w *sync.WaitGroup) {
		defer w.Done()
		err := Store(cout, outFile)
		if err != nil {
			log.Fatalf("Store: %v\n", err)
		}
	}(cout, &wg)
	log.Println("Main: Store started")

	wg.Add(1)
	go func(cin chan Fields, w *sync.WaitGroup) {
		defer w.Done()
		if len(*crit) != 0 {
			cond = cond + "&condition=" + *crit
		}
		All(&t, cond, cin)
	}(cin, &wg)
	log.Println("Main: All started")

	log.Println("Main: waiting...")
	wg.Wait()

	log.Printf("Main: done, results in %s", outFile)
}
