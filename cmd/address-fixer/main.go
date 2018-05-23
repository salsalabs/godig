package addressFixer

//Supporter defines the parts of the supporter record that this app uses.
type Supporter struct {
	Key     string `json:"supporter_KEY"`
	Email   string
	Street  string
	Street2 string
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

//Splitter accepts a buffer and splits it into supporter records.
//Supporter records then flow through the channel.
type Splitter interface {
	Split(b []buf, c chan Supporter)
}

//Auditor record changes to a supporter record.
type Auditor interface {
	Audit(c chan Mod)
}

//Fixer updates a supporter record using SmartyStreets.
type Fixer interface {
	Fix(c1 chan Supporter, c2 chan Supporter, c3 chan Mod)
}

//Saver accepts a supporter record at the end of the processing chain.
//This could be saving the record to disk.  It could also be a sink.
type Saver interface {
	Save(c1 chan Supporter)
}
