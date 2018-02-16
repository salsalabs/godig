//GroupOverlay retrieves supporters for a list of groups KEYs then determines which
//supporters are in the two-way permutations of the groups KEYs.  Output is a
//CSV that shows the groups KEYs and the number of common (overlap) supporters.
//Output also provides lists of supporters for each of the groups KEYs and
//the permutations that have overlap.
//
//This application requires a YAML file of Salsa Classic credentials.
package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/salsalabs/godig"
	"gopkg.in/alecthomas/kingpin.v2"
)

//Directory for dumped data.
const Directory = "./overlap_details"

//Include is a *sorted* list of groups numbers.  Supporters in these groups
//are included in the analysis.
var Include = []int{112201, 112203, 124353, 124354, 155848, 165863, 166303}

//Exclude is a *sorted* list of groups numbers.  Supporters in these groups
//are excluded from the analysis.
var Exclude = []int{163402, 124215}

//Fields contains the fields from a groups record
type Fields struct {
	Supporter_KEY string
	Groups_KEY    string
	SupporterKey  int
	GroupsKey     int
}

//Link is the URL template that retrieves records for a group
const Link = "https://%v/api/getObjects.sjs?json&object=supporter_groups&condition=Last_Modified>=%v&condition=Last_Modified<%v&condition=groups_KEY=%v"

//Build accepts a list of group keys and builds a godig.Census.
func Build(w *godig.Wrapper, dates godig.Dates, keys []int) (godig.Census, error) {
	census := godig.NewCensus(keys)
	for _, k := range keys {
		Task(w, dates, k, census)
	}
	return census, nil
}

//Save writes godig.Members to a file.  The filename contains only the first groups_KEYs
//if both groups KEYs are the same.  The filename contains both groups KEYs if
//they differ.  Output goes into a file in "Directory".  The directory is created
//as needed.
func Save(result godig.Result, gk1, gk2 int) error {
	var a []string
	r := result[gk1][gk2]
	for i, _ := range r {
		s := strconv.Itoa(i)
		a = append(a, s)
	}
	sort.Strings(a)
	s := fmt.Sprintf("supporter_KEY\n%s\n", strings.Join(a, "\n"))
	d1 := []byte(s)
	fn := ""
	if gk1 == gk2 {
		fn = fmt.Sprintf("%d.csv", gk1)
	} else {
		fn = fmt.Sprintf("%d-%d.csv", gk1, gk2)
	}
	fn = path.Join(Directory, fn)

	err := ioutil.WriteFile(fn, d1, 0744)
	if err != nil {
		return err
	}
	log.Printf("Wrote %7d supporter_KEYs to %s", len(r), fn)
	return nil
}

//SaveResult stores the provided result on disk.  Three types of files are created.
//(1) a summary of groups keys and overlap counts, (2) a list of supporters for a group,
//and (3) a list of supporters for a (group, group) pair.
func SaveResult(result godig.Result) error {
	if _, err := os.Stat(Directory); os.IsNotExist(err) {
		err := os.Mkdir(Directory, 0744)
		if err != nil {
			return err
		}
	}

	filename := path.Join(Directory, "summary.csv")
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0744)
	if err != nil {
		return err
	}
	defer f.Close()

	text := fmt.Sprintf("%v,%v,%v\n", "groups_KEY1", "groups_KEY2", "Overlap")
	if _, err := f.WriteString(text); err != nil {
		return err
	}

	for gk1, m := range result {
		for gk2, r := range m {
			if len(r) > 0 {
				err := Save(result, gk1, gk2)
				if err != nil {
					log.Fatal(err)
				}
				text := fmt.Sprintf("%v,%v,%v\n", gk1, gk2, len(r))
				if _, err = f.WriteString(text); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

//SupporterGroups uses a URL template to retrieve supporter groups records by gruops key.
func SupporterGroups(w *godig.Wrapper, r godig.Dates, gk int) ([]Fields, error) {
	x := fmt.Sprintf(Link, w.Host, r.StartDate, r.EndDate, gk)
	var all []Fields
	var a []Fields

	offset := 0
	count := 500
	for count > 0 {
		err := godig.Many(w, x, offset, count, &a)
		if err != nil {
			return nil, err
		}
		for _, r := range a {
			r.SupporterKey = godig.ParseInt(r.Supporter_KEY)
			r.GroupsKey = godig.ParseInt(r.Groups_KEY)
			all = append(all, r)
		}
		count = int(len(a))
		offset = offset + count
	}
	return all, nil
}

//Task is used to retrieve supporter groups records and update a census.
//Errors are noted and ignored.
func Task(w *godig.Wrapper, dates godig.Dates, k int, census godig.Census) {
	a, err := SupporterGroups(w, dates, k)
	if err != nil {
		log.Printf("Supporter Groups error ", err)
		return
	}
	for _, sg := range a {
		members := census[int(k)]
		members[sg.SupporterKey] = true
	}
	log.Printf("Task: %6d %5d members\n", k, len(census[int(k)]))
}

//Mainline retrieves data via the API and aggregates it into a local database.
func main() {
	cpath := kingpin.Flag("credentials", "YAML file containing credentials for Salsa Classic API").Required().PlaceHolder("FILENAME").String()
	sdate := kingpin.Flag("start", "Start date").PlaceHolder("YYYY-MM-DD").Required().String()
	edate := kingpin.Flag("end", "Day AFTER end date").PlaceHolder("YYYY-MM-DD").Required().String()
	dbname := kingpin.Flag("database", "Use this database for this run").Default("./db/godig.sqlite").String()
	kingpin.Parse()

	db, err := sql.Open("sqlite3", *dbname)
	if err != nil {
		log.Println("Database open error")
		log.Fatal(err)
	}
	defer db.Close()

	w := godig.NewWrapper(db, nil, 0)
	err = godig.Authenticate(w, *cpath)
	if err != nil {
		log.Println("Authentication error")
		log.Fatal(err)
	}

	dates := godig.Dates{StartDate: *sdate, EndDate: *edate}
	included, err := Build(w, dates, Include)
	if err != nil {
		log.Println("Include error")
		log.Fatal(err)
	}
	log.Println("included:", len(included))

	excluded, err := Build(w, dates, Exclude)
	if err != nil {
		log.Println("Exclude error")
		log.Fatal(err)
	}
	log.Println("excluded:", len(excluded))

	result, err := godig.Analyze(included, excluded)
	if err != nil {
		log.Println("Analyze error")
		log.Fatal(err)
	}
	err = SaveResult(result)
	if err != nil {
		log.Println("SaveResult")
		log.Fatal(err)
	}
}
