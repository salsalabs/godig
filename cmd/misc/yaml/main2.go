package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
)

func main() {
	var data = `
default: mysql_dev
SetMaxOpenConns: 300
SetMaxIdleConns: 10
mysql_dev: 
  host:     192.168.200.248
  username: gcore
  password: gcore
  port:     3306
  database: test
  charset:  utf8
  protocol: tcp
  prefix:  null
  driver:  mysql
`

	type CredData struct {
		Default         string
		SetMaxOpenConns string
		SetMaxIdleConns string
		Mysql_dev       struct {
			Host     string
			Username string
			Password string
			Port     string
			Database string
			Charset  string
			Protocol string
			Prefix   string
			Driver   string
		}
	}

	c := CredData{}

	fmt.Println("Data: ", data)
	err := yaml.Unmarshal([]byte(data), &c)
	if err != nil {
		fmt.Printf("Unmarshall error %v on %v\n", err, data)
	}
	fmt.Printf("Unpacked: %+v\n", c)
}
