package godig

import (
	"net/http"

	"github.com/tidwall/gjson"
)

//API hold the data that we need to do Salsa API calls.  That includes
//the cookies from authentication.
type API struct {
	Client  *http.Client
	Cookies []*http.Cookie
	Host    string
	Verbose bool
}

//Table links an API to a Salsa API object.
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

//MapList is a slice of FieldMaps.
type MapList []gjson.Result

//NewAPI initializes and returns an API object.
func NewAPI() *API {
	c := API{}
	c.Client = &http.Client{}
	return &c
}

//NewTable creates a table using a tab/object name.
func (a *API) NewTable(n string) Table {
	t := Table{a, n}
	return t
}

//Blast is a shortcut for creating an email_blast Table.
func (a *API) Blast() Table {
	return a.NewTable("email_blast")
}

//Donation is a shortcut for creating a donation Table.
func (a *API) Donation() Table {
	return a.NewTable("donation")
}

//Groups is a shorcut for creating a groups Table.  Note that
//"groups" is the only table in the API that's plural.
func (a *API) Groups() Table {
	return a.NewTable("groups")
}

//GroupsSupporters is a shortcut to join groups to supporters
//via the supporter_groups table. Use LeftJoin to get
//data for this object.
func (a *API) GroupsSupporters() Table {
	return a.NewTable("groups(groups_KEY)supporter_groups(supporter_KEY)supporter")
}

//Supporter is a shortcut for creating a supporter Table.
func (a *API) Supporter() Table {
	return a.NewTable("supporter")
}
//EmailBlast is a shortcut for creating a supporter Table.
func (a *API) EmailBlast() Table {
	return a.NewTable("email_blast")
}

//SupporterDonation is a shortcut for creating a Table that
//holds supporter and donation records.  Use LeftJoin to get
//data for this object.
func (a *API) SupporterDonation() Table {
	return a.NewTable("supporter(supporter_KEY)donation")
}

//SupporterGroups is a shortcut for creating a supporter_group Table.
//SupporterGroups is a shortcut for creating a supporter_group Table.
func (a *API) SupporterGroups() Table {
	return a.NewTable("supporter_groups")
}

//Publish is a shortcut for creating a publish Table.
func (a *API) Publish() Table {
	return a.NewTable("publish")
}
