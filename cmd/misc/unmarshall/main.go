package main

import (
	"encoding/json"
	"fmt"
)

// Record will contain the information retrieved from the database.
type Record struct {
	Groups_KEY string
	Email_KEY  string
	Time_Sent  string
}

func unpack(b []byte, a interface{}) error {
	err := json.Unmarshal(b, &a)
	return err
}

func main() {
	body := []byte(`[{"groups_KEY":"144806","email_KEY":"3329273761","Time_Sent":"Thu Nov 16 2017 16:01:46 GMT-0500 (EST)"}]`)
	var a []Record
	err := unpack(body, &a)
	if err != nil {
		fmt.Printf("Unmarshall error %v on %v\n", err, string(body))
	}
	fmt.Println(a)
}
