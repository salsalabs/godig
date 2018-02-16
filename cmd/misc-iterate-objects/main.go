package main

import (
	"fmt"
)

type Blarg struct {
	Name string
	Count int32
}

func main() {
	x := struct {
		Name string
		Count int32 }{ "Bob", 123456 }
	fmt.Printf("%+v\n", x)
	for i, j := range x {
		fmt.Println("i=", i, ", j=", j)
	}
}

