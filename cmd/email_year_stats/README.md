## email_year_stats
Reporting against the email table in Salsa Classic is not
a good idea.  The email table is in constant use, and reporting
tends to lock the table.  That, in turn, causes problems
in the highly-important email sending area.

This app solves the locking problem by reading the email table and extracting information.  Salsa allows an API-based
app to read up to 500 records at a time.  That means that an
app like this one can do be a low-impact way to retrieve data
from the email table.

### Installation

#### Go language
Thie program requires the [Go](https://golang.org/) programming language.  If you are starting from scratch, then you'll need to install Go, then create a directory structure like this in your home directory.

```
home
 |
 + -- go
      |
      + -- src
      + -- bin
      + -- pkg
```
For various technical and historical reasons, this directory
structure and its placement are non-negotiable.
#### Sqlite3
Install [SQLite version 3](https://www.SQLite.org/index.html).
#### Classic API in Go
```bash
go get github.com/salsalabs/godig
go install github.com/salsalabs/godig
```
### Credentials
The application needs Salsa Classic campaign manager credentials in order to the the API.  You'll pass these credentials to the app via a YAML file.  Here's a sample YAML file.

```yaml
host: salsa4.salsalabs.com
email: barney.blue@frog.bizi
password: 0iquhVthecteqwn0xdhmQnih
```
[Read the API doc](https://help.salsalabs.com/hc/en-us/articles/115000341773-Salsa-Application-Program-Interface-API-) if you have questions.
### Execution
```bash
cd ~/go/src/github.com/salsalabs/godig
go run cmd/email_year_stats/main.go --login LOGIN.YAML
```
Where LOGIN.YAML is described in the previous section.
### Database
The application will create a SQLite database in the current
directory.  The database will contain extractions from records
in the email table.  The database will not contain personally-identifiable information (PII).

There is only one table in the SQLite database.
```SQL
CREATE TABLE IF NOT EXISTS (
    year integer,
    supporter_KEY integer,
    status text)
)
```
### Questions?
Use the Issues link at the top of the [Github repository](https://github.com/salsalabs/godig).  Do not contact Salsalabs support.  It will just confuse them and you won't
get the answer that you need.
