package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// Record will contain the information retrieved from the database.
type Record struct {
	GroupsKey int64
	EmailKey  int64
}

//ClientWrapper holds a Client and its cookies.
type ClientWrapper struct {
	Client  *http.Client
	Cookies []*http.Cookie
}

//NewClientWrapper creates a new ClientWrapper, but without cookies.
func NewClientWrapper(c *http.Client) *ClientWrapper {
	return &ClientWrapper{c, nil}
}

//Auth is the format for submitting authentication request to Salsa.
const Auth = "https://hq-org2.salsalabs.com/api/authenticate.sjs?json&email=%s&password=%s"

//Fetch is the format for submitting the read command to Salsa.
const Fetch = `https://hq-org2.salsalabs.com/api/getLeftJoin.sjs?json&object=groups(groups_KEY)supporter_group(supporter_KEY)email&condition=groups.groups_KEY>=144806&condition=groups.groups_KEY<=144809&condition=email.Time_Sent>=2017-08-01&condition=email.Time_Sent<2017-12-31&include=groups.groups_KEY,email.email_KEY`

func authenticate(w *ClientWrapper, e string, p string) error {
	x := fmt.Sprintf(Auth, e, p)
	resp, err := w.Client.Get(x)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	w.Cookies = resp.Cookies()
	fmt.Println("authentication: ", string(body))
	return nil
}

//Read the data starting at record "i" for 500, if possible.
//Returns the number of records and an error.
func read(w *ClientWrapper, i int) (int, error) {
	limit := fmt.Sprintf("&limit=%d/500", i)
	x := Fetch + limit
	req, err := http.NewRequest("GET", x, nil)
	if err != nil {
		log.Fatalf("New request error %v on %v\n", err, x)
	}
	// Salsa's API needs these cookies to verify authentication.
	for _, c := range w.Cookies {
		req.AddCookie(c)
	}
	resp, err := w.Client.Do(req)
	if err != nil {
		log.Fatalf("Read error %v on %v\n", err, x)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("ReadAll error %v on %v\n", err, x)
	}
	path := fmt.Sprintf("%03d.json", i)
	err = ioutil.WriteFile(path, body, os.ModePerm)
	if err != nil {
		log.Fatalf("Write file error %v on %v\n", err, pat)
		return 0, err
	}
	var a []Record
	err = json.Unmarshal(body, &a)
	count := len(a)
	return count, nil
}

//main is the mainline for the application
func main() {
	client := &http.Client{}
	w := NewClientWrapper(client)
	err := authenticate(w, "aleonard@salsalabs.com", "extra-super-secret-password")
	if err != nil {
		log.Fatalf("Authentication error %v\n", err)
	}

	var i int
	var count = 500
	for count > 0 {
		count, err = read(w, i*500)
		i = i + count
	}
}
