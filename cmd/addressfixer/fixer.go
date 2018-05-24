package addressfixer

//Fix updates a supporter record using SmartyStreets.
//See https://smartystreets.com/docs/sdk/go
func Fix(c1 chan []Supporter, c2 chan []Supporter, c3 chan Mod) {
	defer close(c2)
	defer close(c3)
	for s := range c1 {
		var t []Supporter
		for _, r := range s {
			if len(r.Country) == 0 {
				m := Mod{
					Key:   r.Key,
					Field: "Country",
					Old:   r.Country,
					New:   "US"}
				r.Country = "US"
				c3 <- m
			}
			if len(r.Zip) == 0 {
				r.Zip = "78757"
				m := Mod{
					Key:   r.Key,
					Field: "Zip",
					Old:   "(empty)",
					New:   r.Zip}
				c3 <- m
			}
			t = append(t, r)
		}
		c2 <- t
	}
}
