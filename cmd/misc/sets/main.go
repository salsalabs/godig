package main

import (
	"fmt"
)

//Members is a list of supporter keys for a group.
type Members map[int]bool

//Intersect returns the intersection of two Members.
func Intersect(m1 Members, m2 Members) Members {
	m3 := make(Members)
	for k := range m1 {
		_, ok := m2[k]
		if ok {
			m3[k] = true
		}
	}
	return m3
}

//Difference return the first set without the members of the second set.
func Difference(m1 Members, m2 Members) Members {
	m3 := make(Members)
	for k := range m1 {
		_, ok := m2[k]
		if !ok {
			m3[k] = true
		}

	}
	return m3
}
func main() {
	m1 := make(Members)
	m2 := make(Members)
	for i := 0; i < 10; i++ {
		m1[i] = true
		m2[i*2] = true
	}

	fmt.Println("m1", m1)
	fmt.Println("m2", m2)

	m3 := Intersect(m1, m2)
	fmt.Println("+", m3)

	m3 = Difference(m1, m2)
	fmt.Println("-", m3)
}
