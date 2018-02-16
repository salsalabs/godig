package godig

import (
	"net/http"
)

//API hold the data that we need to do Salsa API calls.  That includes
//the cookies from authentication.
type API struct {
	Client  *http.Client
	Cookies []*http.Cookie
	Host    string
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

//Supporter is a shortcut for creating a supporter Table.
func (a *API) Supporter() Table {
	return a.NewTable("supporter")
}

//SupporterGroups is a shortcut for creating a supporter_group Table.
func (a *API) SupporterGroups() Table {
	return a.NewTable("supporter_groups")
}
