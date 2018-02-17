package godig

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"gopkg.in/yaml.v2"
)

//Authenticate and save the cookies for later.
func (a *API) Authenticate(c CredData) error {
	p := "https://%s/api/authenticate.sjs?json&email=%s&password=%s"
	x := fmt.Sprintf(p, c.Host, c.Email, c.Password)
	resp, body, err := a.Get(x)
	if err != nil {
		return err
	}
	var as AuthStatus
	err = yaml.Unmarshal(body, &as)
	if err == nil {
		if as.Status == "error" {
			err = errors.New(as.Message)
		}
	}
	if err == nil {
		a.Host = c.Host
		a.Cookies = resp.Cookies()
	}
	return err
}

//Count returns the number of records in the table that match the criteria.
//You're responsible for providing valid criteria for the selected table.
//To just count records, pass an empty string in the criteria.
func (t *Table) Count(c string) (string, error) {
	p := "https://%s/api/getCount.sjs?json&object=%s&countColumn=%s_KEY"
	x := fmt.Sprintf(p, t.Host, t.Name, t.Name)
	if len(c) != 0 {
		x = x + "&condition=" + c
	}
	_, body, err := t.Get(x)
	fmt.Println("Count: body is", body)
	//The API does not return valid JSON for getCount.sjs.
	//The body is the count as a string.
	return string(body), err
}

//Credentials retrieves the login credentials from a yaml credentials file
// in the current directory.
func Credentials(cpath string) (CredData, error) {
	raw, err := ioutil.ReadFile(cpath)
	var c CredData
	if err == nil {
		err = yaml.Unmarshal(raw, &c)
	}
	return c, err
}

//Get reads the provided URL and returns the HTTP response, a body and an error.
//Get also adds the cookies that the API needs to prove authentication.
//Your application would probably be better off using One or Many.
func (a *API) Get(u string) (*http.Response, []byte, error) {
	var body []byte
	var resp *http.Response
	req, err := http.NewRequest("GET", u, nil)
	if err == nil {
		// Salsa's API needs these cookies to verify authentication.
		for _, c := range a.Cookies {
			req.AddCookie(c)
		}
		resp, err = a.Client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			body, err = ioutil.ReadAll(resp.Body)
		}
	}
	return resp, body, err
}

//Many reads many records via the API. Many reads the URL, applies the criteria if
// it's not empty.  The read returns up to "count" or 500 records, whichever is
// smaller.  End of data is when number of records is zero. The target is populated
// with an array of records.  An empty target means that no more data is available.
func (t *Table) Many(offset int, count int, crit string, target interface{}) error {
	p := "https://%s/api/getObjects.sjs?json&object=%s&limit=%d,%d"
	x := fmt.Sprintf(p, t.Host, t.Name, offset, count)
	if len(crit) != 0 {
		x = x + "&condition=" + crit
	}
	_, body, err := t.Get(x)
	if err == nil {
		err = json.Unmarshal(body, &target)
	}
	return err
}

//One retrieves a single record using the provided primary key.  The "target"
//is expected to be a struct that defines which fields should be extracted.
//The record will be retrieved into "target".  You can use a Go trick to find
//out if the record number was accurate.
//
//Note that there is not currently a way to retrieve all fields.
//
//TBD Determine how to retrieve all fields from a record.
func (t *Table) One(key string, target interface{}) error {
	p := "https://%s/api/getObject.sjs?json&object=%s&key=%s"
	x := fmt.Sprintf(p, t.Host, t.Name, key)
	_, body, err := t.Get(x)
	if err == nil {
		err = json.Unmarshal(body, target)
	}
	return err
}

//Save does a Salsa API /save.  The caller provides a buffer of fields to
//change.  That will go into the body of a POST request.  The buffer can
//be inordinately long.  Salsa may not process a truly long buffer.  YMWV.
func (t *Table) Save(key string, s string) error {
	u := "https://%s/save"
	x := fmt.Sprintf(u, t.Host)
	p := fmt.Sprintf("json=true&object=%s&key=%s&%s", t.Name, key, s)
	b := bytes.NewReader([]byte(p))
	req, err := http.NewRequest("POST", x, b)
	if err != nil {
		return err
	}
	// Salsa's API needs these cookies to verify authentication.
	for _, c := range t.Cookies {
		req.AddCookie(c)
	}
	// TODO: figure out what to do with the an error response from /save.
	_, err = t.Client.Do(req)
	return nil
}

//YAMLAuth accepts campaign manager credentials (email, password, host)
//from a YAML file and authenticates.
func YAMLAuth(f string) (*API, error) {
	a := NewAPI()
	c, err := Credentials(f)
	if err == nil {
		err = a.Authenticate(c)
	}
	return a, err
}
