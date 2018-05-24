package addressfixer

import (
	"github.com/salsalabs/godig"
)

//Supporter defines the parts of the supporter record that this app uses.
type Supporter struct {
	Key     string `json:"supporter_KEY"`
	Email   string
	Street  string
	Street2 string `json:"Street_2"`
	City    string
	State   string
	Zip     string
	Country string
}

//Mod describes a modification done to a supporter record.
type Mod struct {
	Key   string
	Field string
	Old   string
	New   string
}

//Reader retrieves supporter records in batches and sends them
//down the process stream.
type Reader interface {
	All(t *godig.Table, c chan []Supporter)
}

//Chunkter accepts a buffer and splits it into supporter records.
//Supporter records then flow through the channel.
type Chunkter interface {
	Chunk(c1 chan []Supporter, c2 chan Supporter)
}

//Auditor record changes to a supporter record.
type Auditor interface {
	Audit(c chan Mod)
}

//Fixer updates a supporter record using SmartyStreets.
type Fixer interface {
	Fix(c1 chan Supporter, c2 chan Supporter, c3 chan Mod)
}

//Finisher accepts a supporter record at the end of the processing chain.
//This could be saving the record to disk.  It could also be a sink.
type Finisher interface {
	Finish(t *godig.Table, c1 chan Supporter)
}
