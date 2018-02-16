package main

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

func main() {
	var data = `
host: wfc2.wiredforchage.com
email: aleonard@basledi.com
password: extra-super-secret-password
`

	//Credentials contains the info that we need to get into the API.
	type CredData struct {
		Host     string
		Email    string
		Password string
	}

	c := CredData{}

	fmt.Println("Data: ", data)
	err := yaml.Unmarshal([]byte(data), &c)
	if err != nil {
		fmt.Printf("Unmarshall error %v on %v\n", err, data)
	}
	fmt.Printf("Unpacked: %+v\n", c)
}
