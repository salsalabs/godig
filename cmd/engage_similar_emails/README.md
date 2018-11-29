# Find similar emails in Engage
This application accepts a CSV file of emails from engage and finds similar email addresses.  When similar addresses are found, the are written to a CSV file.

The process requires a CSV of email addresses sorted by email address.  Each list typically contains these Engage fields
* InternalID
* ExternalID
* Email

The process compares each email address to the one just before it.  If they are similar, then they are both written to the CSV file.  For this version, "similar" is defined as a 70% serial match in the contents of the two email addresses.

## Setup

The app is a Go program.  You'll need to start by installing Go and setting up the correct directory structure.  The directory structure is _very import_.

```
HOME_DIR
    + go
        + bin
        + pkg
        + src
```
## Installation

Installation is in two parts.  The first is to retrieve the "godig" package.  That provides access to Salsa Classic's API in Go.  The second is to retrieve all of the dependencies.  The "..." below is, quite literally, "...".

If you are on Windows, then use Windows notation.  Windows support for godig is sketchy at best.

```
go get github.com/salsalabs/godig
cd $(HOME)/go/src/github.com/salsalabs/godig
go get ...
```

## Execution

```
cd $(HOME)/go/src/github.com/salsalabs/godig
go run engage_similar_email/main.go --in sorted_emails.csv --out similar_emails.csv
```
## Warnings

The input CSV file MUST be sorted in ascending order by the Email field.

## Questions?  

Use the [Issues](https://github.com/salsalabs/godig/issues) page in the `godig` repository to ask questions or to report errors.  Don't waste your time contacting Salsalabs Support.  They bite if you get too close to their nest.
