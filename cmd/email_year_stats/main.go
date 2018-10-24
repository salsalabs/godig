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

	_ "github.com/mattn/go-sqlite3"
	"github.com/salsalabs/godig"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

//env is the internal runtime environment.
type env struct {
	C      chan email
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

//push reads emails and pumps them to a channel for email records.
func (e *env) push() error {
	fmt.Println("push: start")
	offset := e.Offset
	count := 500
	for count > 0 {
		var a []email
		err := e.T.Many(offset, count, "", &a)
		if err != nil {
			return err
		}
		count = len(a)
		for _, r := range a {
			e.C <- r
		}
		offset += int32(count)
		if offset%5000 == 0 {
			fmt.Printf("push: offset %8d\n", offset)
		}
	}
	close(e.C)
	fmt.Printf("push: done, offset is %v\n", offset)
	return nil
}

//setup configures and return an env.
func setup(login string, dbPath string, offset int32) (*env, error) {
	fmt.Println("setup: start")
	api, err := (godig.YAMLAuth(login))
	if err != nil {
		return nil, err
	}
	t := api.NewTable("email")
	db, err := sql.Open("sqlite3", dbPath)
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
	c := make(chan email)
	e := env{c, &t, db, s, offset}
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

func main() {
	var (
		login  = kingpin.Flag("login", "YAML file with login credentials").Required().String()
		dbPath = kingpin.Flag("db", "SQLite database to use").Default("./data.sqlite3").String()
		offset = kingpin.Flag("offset", "Start reading at this offset").Default("0").Int32()
	)
	kingpin.Parse()
	if dbPath == nil || len(*dbPath) == 0 {
		fmt.Printf("Oh come on. If you're going to screw with the parameters at least do it right!")
		return
	}
	e, err := setup(*login, *dbPath, *offset)
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	var wg sync.WaitGroup

	go (func(e *env, wg *sync.WaitGroup) {
		wg.Add(1)
		err := e.store()
		wg.Done()
		if err != nil {
			log.Fatalf("%v\n", err)
		}
	})(e, &wg)

	go (func(e *env, wg *sync.WaitGroup) {
		wg.Add(1)
		err := e.push()
		wg.Done()
		if err != nil {
			log.Fatalf("%v\n", err)
		}
	})(e, &wg)

	time.Sleep(10000)
	wg.Wait()
}
