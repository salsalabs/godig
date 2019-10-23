package godig

import (
	"fmt"
	"strings"
	"time"
)

// SalsaTimestamp provides a way to unmarshal Salsa's time into a time object.
//Salsa's time is "" when exported.
//The desired time should be ""/
//Many thanks to OneOfOne
//https://stackoverflow.com/questions/25087960/json-unmarshal-time-that-isnt-in-rfc-3339-format
type SalsaTimestamp struct {
	time.Time
}

//               "Wed Aug 01 2018 11:30:51 GMT-0400 (EDT)"
const ctLayout = "Mon Jan 2 2006 15:04:05 (MST)"
const fmtLayout = "2006-Jan-02 15:04:05"

//UnmarshalJSON parses a byte slice in Salsa format and stores a time object.
func (ct *SalsaTimestamp) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		ct.Time = time.Time{}
		return
	}
	//GMT-500 is pain to parse in Go.  Drop it.
	p := strings.Split(s, " ")
	if len(p) < 7 {
		ct.Time = time.Time{}
		return
	}
	p = append(p[0:5], p[6])
	s = strings.Join(p, " ")
	ct.Time, err = time.Parse(ctLayout, s)
	if err != nil {
		fmt.Printf("Date parse error %v\n", err)
	}
	return
}

//MarshalJSON converts a Time into a Salsa timestamp string.
func (ct *SalsaTimestamp) MarshalJSON() ([]byte, error) {
	if ct.Time.UnixNano() == nilTime {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("%s", ct.Time.Format(fmtLayout))), nil
}

var nilTime = (time.Time{}).UnixNano()

//IsSet returns true if the provided timestamp is not empty.
func (ct *SalsaTimestamp) IsSet() bool {
	return ct.UnixNano() != nilTime
}
