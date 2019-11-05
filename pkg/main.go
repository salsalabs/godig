package godig

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/tidwall/gjson"
)

//ParseFormat is used to parse a Salsa database mesasge.  NOote that the only
//way to do that is to remove the hour offset before parsing.  See `date()`.
const ParseFormat = "Mon Jan 2 2006 15:04:05 (MST)"

//DateFormat is used to format a time so that Engage will recognize it.
const DateFormat = "2006-01-02"

//TimestampFormat is used to format a time so that Engage will recognize it.
const TimestampFormat = "2006-01-02T15:04:05"

//API hold the data that we need to do Salsa API calls.  That includes
//the cookies from authentication.
type API struct {
	Client  *http.Client
	Cookies []*http.Cookie
	Host    string
	Verbose bool
}

//Table links an API to a Salsa database table.
type Table struct {
	*API
	Name string
}

//AuthStatus contains the information returned by Authentication.
type AuthStatus struct {
	Status  string
	Message string
}

//CredData contains the info that we need to get into the API.
type CredData struct {
	Host     string
	Email    string
	Password string
}

//DeleteStatus contins the info returned by deleting a record.
type DeleteStatus struct {
	Object   string
	Key      string
	Result   string
	Messages []string
}

//Field is used to describe table fields when calling Describe.
type Field struct {
	DisplayToSupporters string `json:"display_to_supporters,omitempty"`
	Name                string `json:"name,omitempty"`
	DataColumn          string `json:"data_column,omitempty"`
	IsAZeroIndexEnum    bool   `json:"is_a_zero_index_enum,string,omitempty"`
	Label               string `json:"label,omitempty"`
	DataTable           string `json:"data_table,omitempty"`
	DisplayName         string `json:"displayName,omitempty"`
	Type                string `json:"type,omitempty"`
	IsCustom            bool   `json:"isCustom,omitempty,string"`
}

//FieldList is a slice of Fields returned by Describe.
type FieldList []Field

//MapList is a slice of FieldMaps.
type MapList []gjson.Result

//NewAPI initializes and returns an API object.
func NewAPI() *API {
	c := API{}
	c.Client = &http.Client{}
	return &c
}

//NewTable creates a table using a table/object name.
func (a *API) NewTable(n string) Table {
	t := Table{a, n}
	return t
}

//Donation is a shortcut for creating a donation Table object.
func (a *API) Donation() Table {
	return a.NewTable("donation")
}

//EmailBlast is a shortcut for creating an EmailBlast Table object.
func (a *API) EmailBlast() Table {
	return a.NewTable("email_blast")
}

//Groups is a shorcut for creating a groups Table.  Note that
//"groups" is the only table in the API that's plural.
func (a *API) Groups() Table {
	return a.NewTable("groups")
}

//GroupsSupporters is a shortcut to join groups to supporters
//via the supporter_groups table. Use LeftJoin to get data for
//this object.
func (a *API) GroupsSupporters() Table {
	return a.NewTable("groups(groups_KEY)supporter_groups(supporter_KEY)supporter")
}

//Supporter is a shortcut for creating a supporter Table object.
func (a *API) Supporter() Table {
	return a.NewTable("supporter")
}

//SupporterDonation is a shortcut for creating a Table that
//holds supporter and donation records.  Use LeftJoin to get
//data for this object.
func (a *API) SupporterDonation() Table {
	return a.NewTable("supporter(supporter_KEY)donation")
}

//SupporterGroups is a shortcut for creating a supporter_group Table object.
func (a *API) SupporterGroups() Table {
	return a.NewTable("supporter_groups")
}

//Publish is a shortcut for creating a publish Table object.
func (a *API) Publish() Table {
	return a.NewTable("publish")
}

//EngageDate parses converts a string containing a MySQL date to
//another string containing an Engage date.
func EngageDate(s string) string {
	// Date_Created, Transaction_Date, etc.  Convert dates from MySQL to Engage.
	p := strings.Split(s, " ")
	if len(p) >= 7 {
		//Pull out the timezone.
		p = append(p[0:5], p[6])
		x := strings.Join(p, " ")
		t, err := time.Parse(ParseFormat, x)
		if err != nil {
			log.Printf("Warning: parsing %v returned %v\n", s, err)
		} else {
			s = t.Format(DateFormat)
		}
	}
	return s
}

//EngageTimestamp parses converts a string containing a MySQL date to
//another string containing an Engage date and time.
func EngageTimestamp(s string) string {
	// Date_Created, Transaction_Date, etc.  Convert dates from MySQL to Engage.
	p := strings.Split(s, " ")
	if len(p) >= 7 {
		//Pull out the timezone.
		p = append(p[0:5], p[6])
		x := strings.Join(p, " ")
		t, err := time.Parse(ParseFormat, x)
		if err != nil {
			log.Printf("Warning: parsing %v returned %v\n", s, err)
		} else {
			s = t.Format(TimestampFormat)
		}
	}
	return s
}
