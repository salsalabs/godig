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
	resp, body, err := a.respGet(x)
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

//Get retrieves the content for a URL.
func (a *API) Get(u string) ([]byte, error) {
	_, body, err := a.respGet(u)
	return body, err
}

//Many reads many records via the API.  The caller provides a URL because the
//read many URLs can be fairly complex.  Many reads the URL and returns up to
//"count" or 500 records, whichever is smaller.  End of data is when count is zero.
//The target is populated with an array of records.  An empty target means that
//no more data is available.
func (t *Table) Many(u string, offset int, count int, target interface{}) error {
	x := fmt.Sprintf("%v&limit=%d,%d", u, offset, count)
	body, err := t.Get(x)
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
	x := fmt.Sprintf("https://%s/api/getObject.sjs?json&object=%s&key=%s", t.Host, t.Name, key)
	body, err := t.Get(x)
	if err == nil {
		err = json.Unmarshal(body, target)
	}
	return err
}

//Resp reads the provided URL and returns the HTTP response, a body and an error.
//Quote a lot like GET except for returning the response.  Useful for getting
//to the internals after GET-ing a URL.  Not useful in day-to-day operations.
func (a *API) respGet(u string) (*http.Response, []byte, error) {
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

//Raw reads many records via the API.  The caller provides a full URL.  Many reads
//the URL and returns up to "count" or 500 records, whichever is smaller.  End of
//data is when count is zero. The channel receives the "raw" JSON text from the API call.
//Done receives a true at end of data.
func (t *Table) Raw(u string, offset int, count int, cout chan []byte, done chan bool) error {
	x := fmt.Sprintf("%v&limit=%d,%d", u, offset, count)
	body, err := t.Get(x)
	if err == nil {
		if len(body) <= 2 {
			done <- true
			return nil
		}
		cout <- body
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
	// TODO: figure out what to do with the response from /save.
	_, err = t.Client.Do(req)
	if err != nil {
		return err
	}
	//defer resp.Body.Close()
	//body, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	// 	return err
	//}

	// err = json.Unmarshal(body, &target)
	//return err
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
