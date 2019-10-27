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
	"text/template"

	"github.com/salsalabs/godig/pkg"
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
	Name string
	Ext  string
	Type string
}

//Source is the source object used to format the Go file.
type Source struct {
	Name  string
	Items []Item
}

const pattern = `
<p>The {{.Name}} table contains ...</p>
<h2>Layout</h2>
<table>
<thead>
<tr valign="top">
<th>Field</th>
<th>Type</th>
<th>Notes</th>
</tr>
</thead>
<tbody>
{{range $i := .Items -}}
<tr>
<td>{{$i.Name}}</td>
<td>{{$i.Type}}</td>
<td>{{$i.Label}}</td>
</tr>
{{end}}
</tbody>
</table>
<h2>Format notes</h2>
<p>
<ul>
<li>Salsa timestamps are encoded as "YYYY-MM-DD HH:MM:SS", where all elements are numeric and "HH" is on a 24 hour clock.</li>
<li>Salsa timestamps are in the "America/NewYork" timezone (GMT-5), aka Eastern Standard Time (EST).</li>
</ul>
</p>
<h2>Schema</h2>
<p>This image shows how supporters are linked to an action via the supporter_action object.</p>
<p><img src="/hc/article_attachments/360021842313/Screenshot_at_Dec_28_11-38-18.png" alt="Screenshot_at_Dec_28_11-38-18.png" /></p>
`

func main() {
	var (
		login = kingpin.Flag("login", "YAML file with login credentials").Required().String()
		table = kingpin.Flag("table", "Output Zendesk doc HTML for this table").Required().String()
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
	fn := fmt.Sprintf("%v.html", *table)
	f, err := os.Create(fn)
	if err != nil {
		log.Fatalf("%v, %v\n", err, fn)
	}

	source := Source{
		Name:  *table,
		Items: a,
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
