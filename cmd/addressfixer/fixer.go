package addressfixer

import "log"

//Pen writes an audit record to the audit channe.
func Pen(c3 chan Mod, r Supporter, f string, p string, n string) {
	m := Mod{
		Key:   r.Key,
		Field: f,
		Old:   p,
		New:   n}
	c3 <- m
}

//Fix updates a supporter record using SmartyStreets.
//See https://smartystreets.com/docs/sdk/go
func Fix(c1 chan []Supporter, c2 chan []Supporter, c3 chan Mod) {
	defer close(c2)
	defer close(c3)
	var offset int32
	offset = 0
	totalSkipped := 0
	totalSent := 0
	for a := range c1 {
		var t []Supporter
		skipped := 0
		sent := 0

		for _, r := range a {
			m := false

			x := r.Country
			switch r.Country {
			case "United States":
				r.Country = "US"
			case "Canada":
				r.Country = "CA"
			}
			if x != r.Country {
				Pen(c3, r, "Country", x, r.Country)
				m = true
			}
			if m {
				t = append(t, r)
			} else {
				skipped = skipped + 1
			}
		}
		if len(t) != 0 {
			c2 <- t
			sent = sent + len(t)
		}
		//log.Printf("Fix:     offset %7d, skipped %v, sent %v", offset, skipped, sent)
		totalSent = totalSent + sent
		totalSkipped = totalSkipped + skipped
		offset = offset + int32(len(a))
	}
	log.Printf("Fix:     done, %v records in, sent %v, skipped %v\n", offset, totalSent, totalSkipped)
}

/*
[
  {
    "input_id": "0",
    "input_index": 0,
    "candidate_index": 0,
    "delivery_line_1": "7800 Shoal Creek Blvd",
    "last_line": "Austin TX 78757-1098",
    "delivery_point_barcode": "787571098990",
    "components": {
      "primary_number": "7800",
      "street_name": "Shoal Creek",
      "street_suffix": "Blvd",
      "city_name": "Austin",
      "state_abbreviation": "TX",
      "zipcode": "78757",
      "plus4_code": "1098",
      "delivery_point": "99",
      "delivery_point_check_digit": "0"
    },
    "metadata": {
      "record_type": "H",
      "zip_type": "Standard",
      "county_fips": "48453",
      "county_name": "Travis",
      "carrier_route": "C046",
      "congressional_district": "10",
      "building_default_indicator": "Y",
      "rdi": "Commercial",
      "elot_sequence": "0192",
      "elot_sort": "A",
      "latitude": 30.36008,
      "longitude": -97.74208,
      "precision": "Zip9",
      "time_zone": "Central",
      "utc_offset": -6,
      "dst": true
    },
    "analysis": {
      "dpv_match_code": "D",
      "dpv_footnotes": "AAN1",
      "dpv_cmra": "N",
      "dpv_vacant": "N",
      "active": "N",
      "footnotes": "H#L#"
    }
  }
]
*/

/*
SmartyStreets needs the country for international
422 The following required fields were missing: [country]

Raw Request:

GET /verify?auth-id=21102174564513388&agent=smartystreets%20(website%3Ademo%2Fapis%40latest)&address1=7800%20Shoal%20Creek%20Blvd&locality=Astin&administrative_area=TX&input_id=0 HTTP/1.1
Host: international-street.api.smartystreets.com
Accept: (star)/*
Accept-Encoding: gzip, deflate, br
Accept-Language: en-GB,en;q=0.5
Content-Type: application/json
Dnt: 1
Origin: https://smartystreets.com
Referer: https://smartystreets.com/products/apis/international-street-api
User-Agent: Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:60.0) Gecko/20100101 Firefox/60.0
X-Forwarded-For: 70.123.154.212
X-Forwarded-Proto: https
X-Security-Account-Guid: 4af850d8-0000-0000-0000-000000000000
X-Security-Account-Id: 1257787608
X-Security-Authentication: website-key:hostname
X-Security-Authentication-Key: 21102174564513388
X-Security-Remote-Address: 70.123.154.212
*/
