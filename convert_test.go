package tmpl2js_test

import (
	"bytes"
	"encoding/json"
	"github.com/fatlotus/tmpl2js"
	"github.com/robertkrimen/otto"
	"html/template"
	"testing"
)

type Person struct {
	Name   string   `json:"name"`
	Powers []string `json:"powers"`
}

type Context struct {
	A string `json:"a"`
	B string
	C []struct {
		D int
	}
	E []string
	F struct {
		G string
	}
}

func (c Context) H() struct{ G string } {
	return c.F
}

var tests = []string{
	`{{$var := .A}}{{$var}}`,
	`{{range $i, $x := .C}}{{$i}}: {{$x.D}} = {{.D}}{{end}}`,
	`{{range $i, $x := .E}}{{else}}nop{{end}}`,
	`{{if $x := .A}}{{.A}}{{$x}}{{else}}not{{end}}`,
	`{{if $x := .B}}{{.A}}{{$x}}{{else}}not{{end}}`,
	`{{with $x := .F}}{{.G}}{{$x.G}}{{else}}not{{end}}`,
	`{{with $x := .H}}{{.G}}{{$x.G}}{{end}}`,
	`{{.H.G}}`,
}

func TestFixtures(t *testing.T) {
	ctx := &Context{
		A: "fieldA",
		B: "",
		C: []struct{ D int }{{D: 4}},
		E: []string{"E", "E2", "E3"},
		F: struct{ G string }{G: "GggGG"},
	}

	for _, test := range tests {
		// Render via text/template.
		tmpl := template.Must(template.New("index.html").Parse(test))
		buf := bytes.Buffer{}
		if err := tmpl.Execute(&buf, ctx); err != nil {
			t.Fatal(err)
		}

		// Package up the template as a Template, plus an example type.
		js, err := tmpl2js.Convert(tmpl.Tree, &Context{})
		if err != nil {
			t.Fatal(err)
		}

		// Send the data along as well.
		data, err := json.Marshal(ctx)
		if err != nil {
			t.Fatal(err)
		}

		// Stub out the given method on the object.
		js += "(x=" + string(data) + ",x.H=function() {return this.F},x)"

		// Evaluate the resulting bundle, with the associated functions.
		_, val, err := otto.Run(js)
		if err != nil {
			t.Fatal(err)
		}

		if val.String() != string(buf.Bytes()) {
			t.Fatalf("%s != %s", val.String(), string(buf.Bytes()))
		}
	}
}
