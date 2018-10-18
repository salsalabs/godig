//An application to accept a table and and create a Go file containing
//a table schema.  The table schema can be used to retrieve data from
//Salsa for the table.
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/salsalabs/godig"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

//Item is a single description item from describe2.sjs.
type Item struct {
	Name     string
	Nullable string
	Type     string
	Label    string
}

//Entry is the stuff that we need to format Go object elements.
type Entry struct {
	Cap  string
	Ext  string
	Type string
}

//Source is the source object used to format the Go file.
type Source struct {
	Now     string
	Package string
	Name    string
	CapName string
	Entries []Entry
}

const pattern = `
package {{.Package}}

//{{.CapName}} packages a {{.Name}} object from Salsa's database.
//Created {{.Now}} by schema-maker (github.com/salsalabs/godig/cmd/schema-maker/main.go)
type {{.CapName}} struct {
	{{range $i := .Entries -}}
	{{$i.Cap}} {{$i.Type}} {{$i.Ext}}
	{{end}}
}
`

func capitalize(s string) string {
	return strings.ToUpper(s[0:1]) + strings.ToLower(s[1:])
}

//name removes underbars and capitalizes names.  Leading and trailing
//underbars are ignored.
func goName(s string) string {
	s = strings.TrimSpace(s)
	m, _ := regexp.MatchString("^[\\d_]", s)
	if m {
		s = "F" + s
	}
	p := strings.Split(s, "_")
	var x []string
	for _, d := range p {
		if len(d) > 0 {
			x = append(x, capitalize(d))
		}
	}
	return strings.Join(x, "")
}
func main() {
	var (
		login = kingpin.Flag("login", "YAML file with login credentials").Required().String()
		table = kingpin.Flag("table", "Describe this table").Required().String()
		pkg   = kingpin.Flag("package", "Package for the created file").Default("main").String()
	)
	kingpin.Parse()
	api, err := (godig.YAMLAuth(*login))
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	t := api.NewTable(*table)
	var a []Item
	err = t.Describe(&a)
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	fn := fmt.Sprintf("%v.go", *table)
	f, err := os.Create(fn)
	if err != nil {
		log.Fatalf("%v, %v\n", err, fn)
	}

	cache := make(map[string]int)
	tableName := goName(*table)
	source := Source{
		Now:     time.Now().Format("2-Jan-2006 15:04:05"),
		Package: *pkg,
		Name:    *table,
		CapName: tableName,
	}
	for _, e := range a {
		cap := goName(e.Name)
		// Go likes "UID" and "ID".
		if cap == "Uid" {
			cap = "UID"
		}
		m, _ := regexp.MatchString("Id$", cap)
		if m {
			cap = strings.Replace(cap, "Id", "ID", -1)
		}

		//Remove duplicate API Names.
		_, ok := cache[cap]
		if !ok {
			cache[cap] = 1
			ext := fmt.Sprintf("`json:\"%v\"`", e.Name)
			t := e.Type
			switch e.Type {
			case "blob":
				t = "string"
			case "bool":
				t = "bool"
			case "currency":
				t = "float32"
			case "datetime":
				t = "string"
			case "enum":
				t = "string"
			case "float":
				t = "float32"
			case "int":
				t = "int32"
			case "mediumtext":
				t = "string"
			case "text":
				t = "string"
			case "time":
				t = "string"
			case "timestamp":
				if *pkg == "godig" {
					t = "SalsaTimestamp"
				} else {
					t = "*godig.SalsaTimestamp"
				}
			case "tinyint":
				t = "int32"
			case "varchar":
				t = "string"
			}
			entry := Entry{
				Cap:  cap,
				Ext:  ext,
				Type: t,
			}
			source.Entries = append(source.Entries, entry)
		}
	}
	tmpl, err := template.New("test").Parse(pattern)
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	if err := tmpl.Execute(w, source); err != nil {
		panic(err)
	}
	//"Don't forget to flush"...
	w.Flush()

	f.Write(buf.Bytes())
	fmt.Printf("Schema for %v is in %v\n", *table, fn)
}
