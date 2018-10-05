package godig

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"gopkg.in/yaml.v2"
)

//FixCrit Replace spaces and percent signs in the criteria so that Saosa
//consumes them correctly.
func FixCrit(c string) string {
	c = strings.Replace(c, "%", "%25", -1)
	c = strings.Replace(c, " ", "%20", -1)
	return c
}

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
		x = x + "&condition=" + FixCrit(c)
	}
	_, body, err := t.Get(x)
	//The API does not return valid JSON for getCount.sjs.
	//The body is the count as a string.
	return string(body), err
}

//Credentials retrieves the login credentials from a YAML login file.
func Credentials(p string) (CredData, error) {
	var c CredData
	raw, err := ioutil.ReadFile(p)
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
	if a.Verbose {
		fmt.Printf("Get: %v\n", u)
	}
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
			if a.Verbose {
				fmt.Printf("Get: %v\n", string(body))
			}
		}
	}
	return resp, body, err
}

//LeftJoinRaw does a left join using Salsa's API and returns a buffer of bytes.
func (t *Table) LeftJoinRaw(offset int32, count int, crit string) ([]byte, error) {
	p := "https://%s/api/getLeftJoin.sjs?json&object=%s&limit=%d,%d"
	x := fmt.Sprintf(p, t.Host, t.Name, offset, count)
	if len(crit) != 0 {
		x = x + "&condition=" + FixCrit(crit)
	}
	_, body, err := t.Get(x)
	return body, err
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
//The target is a slice of schemas.  The schemas contain the fields that
//you'd like to see.  Be sure to use the form "table.fieldName" in the
//JSON extensions to assure that the data is retrieved correctly.
func (t *Table) LeftJoin(offset int32, count int, crit string, target interface{}) error {
	body, err := t.LeftJoinRaw(offset, count, crit)
	if err == nil {
		err = json.Unmarshal(body, &target)
	}
	return err
}

//Many reads many records from a table. Reading starts at offset and retrieves
//count records.   Salsa will never return more than 500 records, however.
//The target is a slice of records that match the table schema. Many automatically
//unmarshals from JSON into the target.  An empty target indicates end of data.
func (t *Table) Many(offset int32, count int, crit string, target interface{}) error {
	body, err := t.ManyRaw(offset, count, crit)
	if err == nil {
		err = json.Unmarshal(body, &target)
	}
	return err
}

//ManyTagged reads many records from a table. Records share a common tag.
// Reading starts at offset and retrieves count records.   Salsa will never
// return more than 500 records, however.
//
//The target is a slice of records that match the table schema. Many automatically
//unmarshals from JSON into the target.  An empty target indicates end of data.
func (t *Table) ManyTagged(offset int32, count int, crit string, tag string, target interface{}) error {
	body, err := t.ManyRawTagged(offset, count, crit, tag)
	if err == nil {
		err = json.Unmarshal(body, &target)
	}
	return err
}

//ManyRawTagged reads many records from a table. Records share a common tag.
// Reading starts at offset and retrieves count records. Salsa will never
// return more than 500 records, however.
func (t *Table) ManyRawTagged(offset int32, count int, crit string, tag string) ([]byte, error) {
	p := "https://%s/api/getTaggedObjects.sjs?json&object=%s&tag=%s&limit=%d,%d"
	x := fmt.Sprintf(p, t.Host, t.Name, tag, offset, count)
	if len(crit) != 0 {
		x = x + "&condition=" + FixCrit(crit)
	}
	_, body, err := t.Get(x)
	return body, err
}

//ManyRaw reads many records from a table. Reading starts at offset and
//retrieves count records.   Salsa will never return more than 500 records,
//however.  The results are unmarshalled data in JSON format.
func (t *Table) ManyRaw(offset int32, count int, crit string) ([]byte, error) {
	p := "https://%s/api/getObjects.sjs?json&object=%s&limit=%d,%d"
	x := fmt.Sprintf(p, t.Host, t.Name, offset, count)
	if len(crit) != 0 {
		x = x + "&condition=" + FixCrit(crit)
	}
	_, body, err := t.Get(x)
	return body, err
}

//One retrieves a single record using the provided primary key.  The target
//is the address of a record schema, which defines which fields will
//be returned.  Note that there is not currently a way to retrieve all fields
//into a schema.  Use ManyMap to do that.
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
//changed in the form "fieldname=value".
//
//The buffer can be inordinately long.  Salsa may not process a truly
//long buffer.  YMWV.
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

	return body, err
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
