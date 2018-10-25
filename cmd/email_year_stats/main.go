package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/salsalabs/godig"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

//env is the internal runtime environment.
type env struct {
	N      chan int32
	C      chan email
	D      chan bool
	T      *godig.Table
	DB     *sql.DB
	Insert *sql.Stmt
	Offset int32
}

//email is what Salsa records for each email.
type email struct {
	EmailKey      string `json:"email_KEY"`
	SupporterKey  string `json:"supporter_KEY"`
	EmailBlastKey string `json:"email_blast_KEY"`
	LastModified  string `json:"Last_Modified"`
	TimeRequested string `json:"Time_Requested"`
	TimeSent      string `json:"Time_Sent"`
	Status        string `json:"Status"`
	StatusCount   string `json:"Status_Count"`
	ThreadID      string `json:"thread_ID"`
}

//Fetch retrieves offsets from the offset channel, reads records
//from Salsa, then puts the records onto the save channel.
func (e *env) fetch() error {
	fmt.Println("fetch: start")
	for {
		offset, ok := <-e.N
		if !ok {
			e.D <- true
			break
		}
		fmt.Printf("fetch: popped %8d\n", offset)
		var a []email
		err := e.T.Many(offset, 500, "", &a)
		if err != nil {
			return err
		}
		for _, r := range a {
			e.C <- r
		}
	}
	fmt.Println("fetch: done")
	return nil
}

//push determines how many emails to process and pushes offsets
//onto the offset channel.
func (e *env) push() error {
	fmt.Println("push: start")

	s, err := e.T.Count("")
	x, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return err
	}
	max := int32(x)
	fmt.Printf("push: max is %v\n", max)
	for i := e.Offset; i <= max; i += 500 {
		e.N <- i
	}
	close(e.N)
	fmt.Println("push: done")
	return nil
}

//setup configures and return an env.
func setup(login string, dbPath string, offset int32, mysql *bool) (*env, error) {
	fmt.Println("setup: start")
	api, err := (godig.YAMLAuth(login))
	if err != nil {
		return nil, err
	}
	t := api.NewTable("email")
	var db *sql.DB
	if mysql != nil && *mysql {
		db, err = sql.Open("mysql", "generic:generic-at-loca@tcp(127.0.0.1:3306)/generic")
		fmt.Println("setup: opened MySQL database")
	} else {
		db, err = sql.Open("sqlite3", dbPath)
		fmt.Println("setup: opened SQLite database")
	}
	if err != nil {
		return nil, err
	}
	sqlTable := `
	CREATE TABLE IF NOT EXISTS data(
		year integer,
		supporter_KEY integer,
		status text
	    );
	`
	_, err = db.Exec(sqlTable)
	if err != nil {
		return nil, err
	}

	sqlInsert := `
	INSERT INTO data(year, supporter_KEY, status)
	VALUES(?, ?, ?);
	`
	s, err := db.Prepare(sqlInsert)
	if err != nil {
		panic(err)
	}
	c := make(chan email, 100)
	d := make(chan bool)
	n := make(chan int32, 1000)
	e := env{
		N:      n,
		C:      c,
		D:      d,
		T:      &t,
		DB:     db,
		Insert: s,
		Offset: offset}
	fmt.Println("setup: done")
	return &e, nil
}

//store stores parts of an email record to the database.
func (e *env) store() error {
	fmt.Println("store: start")
	count := int32(0)
	for {
		r, ok := <-e.C
		if !ok {
			break
		}
		// "Wed Aug 01 2018 11:30:51 GMT-0400 (EDT)"
		p := strings.Split(r.TimeSent, " ")
		y, err := strconv.ParseInt(p[3], 10, 32)
		if err != nil {
			m := fmt.Sprintf("%v on '%v'", err, p[3])
			err = errors.New(m)
			return err
		}
		sk, _ := strconv.ParseInt(r.SupporterKey, 10, 32)
		_, err = e.Insert.Exec(y, sk, r.Status)
		if err != nil {
			return err
		}
		count++
	}
	fmt.Printf("store: done, count is %v\n", count)
	return nil
}

//waitFor watches the done channel for 'n' messages to arrive.  When the n-th
//one arrives, then close save channel.
func (e *env) waitFor(n int) {
	fmt.Println("waitFor: start")
	for {
		_, ok := <-e.D
		if !ok {
			break
		}
		n--
		if n == 0 {
			break
		}
	}
	close(e.C)
	fmt.Println("waitFor: done")
}

func main() {
	var (
		login  = kingpin.Flag("login", "YAML file with login credentials").Required().String()
		dbPath = kingpin.Flag("db", "SQLite database to use").Default("./data.sqlite3").String()
		offset = kingpin.Flag("offset", "Start reading at this offset").Default("0").Int32()
		mysql  = kingpin.Flag("mysql", "Use MySQL instead of SQLite").Bool()
	)
	kingpin.Parse()
	if dbPath == nil || len(*dbPath) == 0 {
		fmt.Printf("--dbpath requires a filename")
		return
	}
	e, err := setup(*login, *dbPath, *offset, mysql)
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	var wg sync.WaitGroup

	// Read email records, write to the database.
	go (func(e *env, wg *sync.WaitGroup) {
		wg.Add(1)
		err := e.store()
		wg.Done()
		if err != nil {
			log.Fatalf("%v\n", err)
		}
	})(e, &wg)

	// Read offsets, push email records.
	for i := 0; i < 5; i++ {
		go (func(e *env, wg *sync.WaitGroup) {
			wg.Add(1)
			err := e.fetch()
			wg.Done()
			if err != nil {
				log.Fatalf("%v\n", err)
			}
		})(e, &wg)
	}

	// Read number of emails, push offsets.
	go (func(e *env, wg *sync.WaitGroup) {
		wg.Add(1)
		err := e.push()
		wg.Done()
		if err != nil {
			log.Fatalf("%v\n", err)
		}
	})(e, &wg)

	// Wait for fetchers to terminate.
	go (func(e *env, wg *sync.WaitGroup) {
		wg.Add(1)
		e.waitFor(5)
		wg.Done()
		if err != nil {
			log.Fatalf("%v\n", err)
		}
	})(e, &wg)

	// Settle for a bit to let Salsa I/O get started (it can
	// take a while), then wait for tasks to complete.
	time.Sleep(10000)
	wg.Wait()
}
