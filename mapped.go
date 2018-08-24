package godig
import (
	"github.com/tidwall/gjson"
)

//OneMap retrieves a single record using the provided primary key.  The
//returned record is a map of names and values.  Everything is a string.
func (t *Table) OneMap(key string) (map[string]string, error) {
	var b map[string]string
	body, err := t.OneRaw(key)
	if err != nil {
		return b, err
	}
	r := gjson.ParseBytes(body)
	b = unpackGJsonMap(r)
	return b, err
}

//ManyMap returns an array of records.  Each record is a map of field names
// and values. An empty array indicates end of data.
func (t *Table) ManyMap(offset int32, count int, crit string) ([]map[string]string, error) {
	var a []map[string]string
	body, err := t.ManyRaw(offset, count, crit)
	if err != nil {
		return a, err
	}
	a = unpackGJsonArray(body)
	return a, nil
}

//LeftJoinMap reads from Salsa and returns an array of maps. The results are
//unmarshalled using gjson. Each map containsa single record.
func (t *Table) LeftJoinMap(offset int32, count int, crit string) ([]map[string]string, error) {
	var a []map[string]string
	body, err := t.LeftJoinRaw(offset, count, crit)
	a = unpackGJsonArray(body)
	return a, err
}

func unpackGJsonMap(r gjson.Result) map[string]string {
	b := make(map[string]string, 0)
	r.ForEach(func(key, value gjson.Result) bool {
		b[key.String()] = value.String()
		return true
	})
	return b
}

func unpackGJsonArray(body []byte) []map[string]string {
	a := make([]map[string]string, 0)
	f := gjson.ParseBytes(body).Array()
	for _, r := range f {
		b := unpackGJsonMap(r)
		a = append(a, b)
	}
	return a
}
