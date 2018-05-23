package addressFixer

//Supporter defines the parts of the supporter record that this app uses.
type Supporter sruct {
	Key string `json:"supporter_KEY"`
	Email string
	Street string
	Street2 string
	City string
	State string
	Zip string
	Country string
}

//Mod describes a modification done to a supporter record.
type Mod struct {
	Key string
	Field string
	Old string
	New string
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
