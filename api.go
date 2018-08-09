package godig

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/tidwall/gjson"
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

//Delete does a Salsa API /delete.  The caller provides a key. We whack that record.
func (t *Table) Delete(key string, target interface{}) error {
	u := "https://%s/delete?json=true&object=%s&key=%s"
	x := fmt.Sprintf(u, t.Host, t.Name, key)
	resp, body, err := t.Get(x)
	if err == nil {
		if resp.StatusCode != 200 {
			return errors.New(resp.Status)
		}
		err = json.Unmarshal(body, target)
	}
	return err
}

//Describe shows returns the table structure as an array of field descriptors.
func (t *Table) Describe(target interface{}) error {
	p := "https://%s/api/describe2.sjs?json&object=%s"
	x := fmt.Sprintf(p, t.Host, t.Name)
	_, body, err := t.Get(x)
	if err == nil {
		err = json.Unmarshal(body, target)
	}
	return err
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

//LeftJoin reads two or more tables from the database.  The tables are
//joined on the primary key of the left table.  For example, to find supporters
//and their donations would require a table join statement like
//'supporter(supporter_KEY)donation'.  The field supporter_KEY is the
//primary key in supporter and a foreign key in donation.
//
//LeftJoin reads staring at offset.  It will retrieve either count
//or 500 records, whichever is smaller.  If crit is defined, then that
//is added to the URL as a condition to limit the number of supporters.
//
//LeftJoin converts data to JSON into the target.  The target should be
//a slice of a record type.  The record type defines which fields to
//return.
func (t *Table) LeftJoin(offset int32, count int, crit string, target interface{}) error {
	p := "https://%s/api/getLeftJoin.sjs?json&object=%s&limit=%d,%d"
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

//Many reads many records via the API. Many reads the URL, applies the criteria if
// it's not empty.  The offiset is a 32-bin integer.  The count is a small integer because
// Salsa only allows 500 records when reading via the API.
//
//The read returns up to "count" or 500 records, whichever is
// smaller.  End of data is when number of records is zero. The target is populated
// with an array of records.  An empty target means that no more data is available.
func (t *Table) Many(offset int32, count int, crit string, target interface{}) error {
	body, err := t.ManyRaw(offset, count, crit)
	if err == nil {
		err = json.Unmarshal(body, &target)
		if err != nil {
			log.Printf("\nAPI.Many: Error %s\n offset: %v\ncount: %v\n,crit: '%v'\nbody: \n%v\n\n", err, offset, count, crit, string(body))
		}
	}
	return err
}

//ManyMap reads many records via the API and returns a FieldList. Many reads the URL,
/// applies the criteria, then fetches the data.  An empty field list is end of data.
//
//  Note that Salsa only allows 500 records when reading via the API.
//The read returns up to "count" or 500 records, whichever is
// smaller.  End of data is when number of records is zero. The target is populated
// with an array of records.  An empty target means that no more data is available.
func (t *Table) ManyMap(offset int32, count int, crit string) (f MapList, err error) {
	body, err := t.ManyRaw(offset, count, crit)
	if err != nil {
		return f, err
	}
	f = gjson.ParseBytes(body).Array()
	return f, err
}

//ManyRaw reads many records via the API and returns a buffer.
//offset is a 32-bin integer.  Count can be an int because we can't
//read more than 500 records from Salsa via the API.
func (t *Table) ManyRaw(offset int32, count int, crit string) ([]byte, error) {
	p := "https://%s/api/getObjects.sjs?json&object=%s&limit=%d,%d"
	x := fmt.Sprintf(p, t.Host, t.Name, offset, count)
	if len(crit) != 0 {
		x = x + "&condition=" + crit
	}
	//fmt.Printf("ManyRaw: %v\n", x)
	_, body, err := t.Get(x)
	return body, err
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
	body, err := t.OneRaw(key)
	if err == nil {
		err = json.Unmarshal(body, target)
	}
	return err
}

//OneRaw retrieves a single record using the provided primary key.
//Returns the buffer retrieved from the URL.
func (t *Table) OneRaw(key string) ([]byte, error) {
	p := "https://%s/api/getObject.sjs?json&object=%s&key=%s"
	x := fmt.Sprintf(p, t.Host, t.Name, key)

	resp, body, err := t.Get(x)
	if err == nil {
		if resp.StatusCode != 200 {
			return body, errors.New(resp.Status)
		}
	}
	return body, err
}

//Save does a Salsa API /save.  The caller provides a buffer of fields to
//change.  That will go into the body of a POST request.  The buffer can
//be inordinately long.  Salsa may not process a truly long buffer.  YMWV.
func (t *Table) Save(key string, s string) ([]byte, error) {
	p := fmt.Sprintf("&object=%s&key=%s&%s", t.Name, key, s)
	return t.SaveBulk(p)
}

//SaveBulk does a Salsa API /save.  The caller provides the contents
// of the body for a POST.  In general, the contents can be characterized
// as
// "&object=" followed by the table name,
// "&key=" followed by zero or the primary key
// and multiple instances of
// "&FieldName=Value"
// SaveBulk returns the body of the response and an error
func (t *Table) SaveBulk(s string) ([]byte, error) {
	u := "https://%s/save"
	x := fmt.Sprintf(u, t.Host)

	w := bytes.NewBufferString("?json")
	_, _ = w.WriteString(s)
	b := bytes.NewReader(w.Bytes())
	//log.Printf("SaveBulk: writing %s\n", w.String())
	req, err := http.NewRequest("POST", x, b)
	if err != nil {
		return nil, err
	}
	// Salsa's API needs these cookies to verify authentication.
	for _, c := range t.Cookies {
		req.AddCookie(c)
	}
	// TODO: figure out what to do with the an error response from /save.
	resp, err := t.Client.Do(req)
	var body []byte
	if err == nil {
		defer resp.Body.Close()
		body, err = ioutil.ReadAll(resp.Body)
	}

	return body, nil
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
