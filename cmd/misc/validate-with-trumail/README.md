# Validate email addresses using trumail.io

This application accepts a CSV file of emails from engage and validates emails via trumail.io.  Output is a CSV of the input record with trumail indicators appended.

The process requires a CSV of email addresses.  Each list typically contains these Engage fields
* InternalID
* ExternalID
* Email


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
go run validate-with-trumail/main.go --in emails.csv --out validated.csv
```

## Example

### Input file
```
Email,InternalID,ExternalID
able@wow.way.com,5e61a1b3-a31e-421e-8b7d-88af4e8da73f,91622931
bravo@woway.com,b8cc450c-b3f5-440e-b159-b9162f364383,89896535
charlie@dot.jong.com,669f5586-3797-45ff-a276-c38e2646ad1a,85626875
delta@dotjong.com,d3d0f3e7-317d-4006-8cab-ccd971ac7bb2,85626876
```

### Output file
```
The output file contains the same fields as the input file.  These TruMail values are appended to each line
* ValidFormat
* Deliverable
* HostExists
* Suggestion
Email,ValidFormat,Deliverable,HostExists,InternalID,ExternalID,Suggestion
able@wow.way.com,true,false,false,5e61a1b3-a31e-421e-8b7d-88af4e8da73f,91622931,
bravo@woway.com,false,false,false,b8cc450c-b3f5-440e-b159-b9162f364383,89896535,
charlie@dot.jong.com,false,false,false,669f5586-3797-45ff-a276-c38e2646ad1a,85626875,
delta@dotjong.com,false,false,false,d3d0f3e7-317d-4006-8cab-ccd971ac7bb2,85626876,
```
## Warnings

TruMail has limits on what it will do for free.  You'll start to see a lot of false indicators when the emails are clearly valid when you've reached the limit.  Grab a credit card for more access.

## Questions?  

Use the [Issues](https://github.com/salsalabs/godig/issues) page in the `godig` repository to ask questions or to report errors.  Don't waste your time contacting Salsalabs Support.  They bite if you get too close to their nest.
