package addressfixer

//Audit record changes to a supporter record.
func Audit(c chan Mod) {
	for a := range c {
		_ = a.Field
	}
}
