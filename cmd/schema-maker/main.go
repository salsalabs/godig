//An application to accept a table and and create a Go file containing
//a table schema.  The table schema can be used to retrieve data from
//Salsa for the table.
package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/salsalabs/godig"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

//Item is a single description item from describe2.sjs.
type Item struct {
	Name         string
	Nullable     string
	Type         string
	DefaultValue string
	Label        string
}

//Entry is the stuff that we need to format Go object elements.
type Entry struct {
	Cap string
	Ext string
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
	{{$i.Cap}} string {{$i.Ext}}
	{{end}}
}
`

func capitalize(s string) string {
	return strings.ToUpper(s[0:1]) + strings.ToLower(s[1:])
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
	//b, err := json.MarshalIndent(a, "", "\t")
	//if err != nil {
	//	log.Fatalf("%v\n", err)
	//}
	//fmt.Println(string(b))

	fn := fmt.Sprintf("%v.go", *table)
	f, err := os.Create(fn)
	if err != nil {
		log.Fatalf("%v, %v\n", err, fn)
	}

	source := Source{
		Now:     time.Now().Format("2-Jan-2006 15:04:05"),
		Package: *pkg,
		Name:    *table,
		CapName: capitalize(*table),
	}
	for _, e := range a {
		p := strings.Split(e.Name, "_")
		var x []string
		for _, d := range p {
			x = append(x, capitalize(d))
		}
		cap := strings.Join(x, "")
		ext := fmt.Sprintf("`json:\"%v\"`", e.Name)
		entry := Entry{
			Cap: cap,
			Ext: ext,
		}
		source.Entries = append(source.Entries, entry)
	}

	tmpl, err := template.New("test").Parse(pattern)
	if err != nil {
		panic(err)
	}
	if err := tmpl.Execute(f, source); err != nil {
		panic(err)
	}
	fmt.Printf("Schema for %v is in %v\n", *table, fn)
}
