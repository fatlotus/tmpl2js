package tmpl2js_test

import (
	"bytes"
	"encoding/json"
	"github.com/fatlotus/tmpl2js"
	"github.com/robertkrimen/otto"
	"testing"

	html "html/template"
	text "text/template"
)

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

func (c Context) I(a, b int) int {
	return a + b
}

var positive = []string{
	`{{$var := .A}}{{$var}}`,
	`{{range $i, $x := .C}}{{$i}}: {{$x.D}} = {{.D}}{{end}}`,
	`{{range $i, $x := .E}}{{else}}nop{{end}}`,
	`{{range $x := .C}}{{$x.D}}{{end}}`,
	`{{if $x := .A}}{{.A}}{{$x}}{{else}}not{{end}}`,
	`{{if $x := .B}}{{.A}}{{$x}}{{else}}not{{end}}`,
	`{{with $x := .F}}{{.G}}{{$x.G}}{{else}}not{{end}}`,
	`{{with $x := .H}}{{.G}}{{$x.G}}{{end}}`,
	`{{.H.G}} From global: {{$.H.G}}`,
	`{{define "sub"}}G is: {{.F.G}}{{end}}{{template "sub" .}}`,
	`{{define "sub"}}Sub{{end}}{{template "sub"}}`,
	`Int: {{1}} float: {{1.2}} string: {{"hello"}} Bools: {{true}} {{false}}`,
	`Variable: {{$x := .F}}{{$x.G}}`,
	`Args: {{.I 3 4}}`,
}

func TestConvertHTML(t *testing.T) {
	ctx := &Context{
		A: "fieldA",
		B: "",
		C: []struct{ D int }{{D: 4}},
		E: []string{"E", "E2", "E3"},
		F: struct{ G string }{G: "GggGG"},
	}

	for _, test := range positive {
		t.Log(test)

		// Render via html/template.
		tmpl := html.Must(html.New("").Parse(test))
		buf := bytes.Buffer{}
		if err := tmpl.Execute(&buf, ctx); err != nil {
			t.Fatal(err)
		}

		// Package up the template as a Template, plus an example type.
		js, err := tmpl2js.ConvertHTML(tmpl, &Context{})
		if err != nil {
			t.Fatal(err)
		}

		// Send the data along as well.
		data, err := json.Marshal(ctx)
		if err != nil {
			t.Fatal(err)
		}

		// Stub out the given method on the object.
		js += "(x=" + string(data) + ",x.H=function() {return this.F},"
		js += "x.I=function(a, b){return a + b},x)"
		t.Log(js)

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

func TestConvertText(t *testing.T) {
	ctx := &Context{
		A: "fieldA",
		B: "",
		C: []struct{ D int }{{D: 4}},
		E: []string{"E", "E2", "E3"},
		F: struct{ G string }{G: "GggGG"},
	}

	for _, test := range positive {
		t.Log(test)

		// Render via text/template.
		tmpl := text.Must(text.New("").Parse(test))
		buf := bytes.Buffer{}
		if err := tmpl.Execute(&buf, ctx); err != nil {
			t.Fatal(err)
		}

		// Package up the template as a Template, plus an example type.
		js, err := tmpl2js.ConvertText(tmpl, &Context{})
		if err != nil {
			t.Fatal(err)
		}

		// Send the data along as well.
		data, err := json.Marshal(ctx)
		if err != nil {
			t.Fatal(err)
		}

		// Stub out the given method on the object.
		js += "(x=" + string(data) + ",x.H=function() {return this.F},"
		js += "x.I=function(a, b){return a + b},x)"
		t.Log(js)

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

var negative = []string{
	`{{.NotExist}}`,
	`{{.A.NotExist}}`,
	`{{.a}}`,
	`{{$x := 0}}{{$x.NotExist}}`,
	`{{$x := true}}{{$x.NotExist}}`,
	`{{$x := ""}}{{$x.NotExist}}`,
	`{{$x := ($y := "")}}{{$x.NotExist}}`,
	`{{.I.NotExist}}`,
	`{{range .I}}{{end}}`,
	`{{.C.D}}`,
	`{{range true}}{{end}}`,
	`{{range 0}}{{end}}`,
	`{{range ""}}{{end}}`,
	`{{range .}}{{end}}`,
}

func TestFailure(t *testing.T) {
	for _, test := range negative {
		tmpl := html.Must(html.New("").Parse(test))
		_, err := tmpl2js.ConvertHTML(tmpl, &Context{})
		if err == nil {
			t.Fatalf("expecting error from: %s", test)
		}
		t.Log(err.Error())
	}
}
