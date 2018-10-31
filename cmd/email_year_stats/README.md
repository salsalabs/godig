# email_year_stats
Reporting against the email table in Salsa Classic is not
a good idea.  The email table is in constant use, and reporting
tends to lock the table.  That, in turn, causes problems
in the email sending processes.

This app solves the locking problem by reading the email table and extracting information.  Salsa allows an API-based
app to read up to 500 records at a time.  That means that an
app like this one can do be a low-impact way to retrieve data
from the email table.

## Installation

### Go language
Thie program requires the [Go](https://golang.org/) programming language.  If you are starting from scratch, then you'll need to install Go, then create a directory structure like this in your home directory.

```
home
 |
 + - go
      + - src
      + - bin
      + - pkg
```
For various technical and historical reasons, this directory
structure and its placement are non-negotiable.
### Sqlite3
Install [SQLite version 3](https://www.SQLite.org/index.html).
### Classic API in Go
```bash
go get github.com/salsalabs/godig
go install github.com/salsalabs/godig
```
### Credentials
The application needs Salsa Classic campaign manager credentials in order to the the API.  You'll pass these credentials to the app via a YAML file.  Here's a sample YAML file.

```
host: salsa4.salsalabs.com
email: barney.blue@frog.bizi
password: 0iquhVthecteqwn0xdhmQnih
```
[Read the API doc](https://help.salsalabs.com/hc/en-us/articles/115000341773-Salsa-Application-Program-Interface-API-) if you have questions.

## Execution
```bash
cd ~/go/src/github.com/salsalabs/godig
go run cmd/email_year_stats/main.go --help

usage: main --login=LOGIN [<flags>]

Flags:
  --help                 Show context-sensitive help (also try --help-long and --help-man).
  --login=LOGIN          YAML file with login credentials
  --db="./data.sqlite3"  SQLite database to use
  --offset=0             Start reading at this offset
  --mysql                Use MySQL instead of SQLite
```
Where LOGIN is described in the previous section.

## Database
The application will use

* a SQLite database in the current directory, or
* a MySQL/MariaDB database

The MySQL/MariaDB database uses these parameters

| Param | Value |
| ----- | ----- |
|host | 127.0.0.1|
|database| generic |
|user | generic |
|password| hard coded|


Checking in a hard-coded database is suboptimal.  Secure-ish  (no PII and bad guys would have to find the database), but still suboptimal.

TODO: Create a mysql database parameter file.

### Schema

The database will contain extractions from records
in the email table.  The database will not contain personally-identifiable information (PII).

There is only one table in the database.
```SQL
CREATE TABLE IF NOT EXISTS (
    year integer,
    supporter_KEY integer,
    status text)
)
```
### Analysis
The initial requirement was for the number of supporters who 
at least opened an email, reported on a hearly basis.  We
can retrieve that from the database using these statements.
```ssql
.mode csv
SELECT year, status, count(*) 
FROM data
GROUP BY year, status
ORDER BY hear, status;
```
That should return data that looks like this.
```
2011,Opened,123456
2011,Sent and Opened,23456
2011,Sent and Clicked, 3456
2012, Opened, 234567
...
```
## Questions?
Use the Issues link at the top of the [Github repository](https://github.com/salsalabs/godig).  Do not contact Salsalabs support.  It will just confuse them and you won't
get the answer that you need.
