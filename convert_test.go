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
	`Comparison: {{lt 1 2}}`,
	`Helper: {{helper 42}} also: {{ 42 | helper }}`,
	`Value of assignment: {{$x := ($y := 2)}}{{$x}} {{($y := .F).G}}`,
}

func TestConvertHTML(t *testing.T) {
	ctx := &Context{
		A: "fieldA",
		B: "",
		C: []struct{ D int }{{D: 4}},
		E: []string{"E", "E2", "E3"},
		F: struct{ G string }{G: "GggGG"},
	}
	helpers := html.FuncMap{"helper": func(x int) int { return 2 * x }}

	for _, test := range positive {
		t.Log(test)

		// Render via html/template.
		tmpl, err := html.New("").Funcs(helpers).Parse(test)
		if err != nil {
			t.Fatal(err)
		}

		buf := bytes.Buffer{}
		if err := tmpl.Execute(&buf, ctx); err != nil {
			t.Fatal(err)
		}

		// Package up the template as a Template, plus an example type.
		js, err := tmpl2js.ConvertHTML(tmpl, &Context{}, helpers)
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
		js += "x.I=function(a, b){return a + b},"
		js += "x.$helper=function(x){return 2 * x},x)"
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
	helpers := text.FuncMap{"helper": func(x int) int { return 2 * x }}

	for _, test := range positive {
		t.Log(test)

		// Render via text/template.
		tmpl, err := text.New("").Funcs(helpers).Parse(test)
		if err != nil {
			t.Fatal(err)
		}

		buf := bytes.Buffer{}
		if err := tmpl.Execute(&buf, ctx); err != nil {
			t.Fatal(err)
		}

		// Package up the template as a Template, plus an example type.
		js, err := tmpl2js.ConvertText(tmpl, &Context{}, helpers)
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
		js += "x.I=function(a, b){return a + b},"
		js += "x.$helper=function(x){return 2 * x},x)"
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
	`{{($y := "").NotExist}}`,
	`{{.I.NotExist}}`,
	`{{range .I}}{{end}}`,
	`{{.C.D}}`,
	`{{range true}}{{end}}`,
	`{{range 0}}{{end}}`,
	`{{range ""}}{{end}}`,
	`{{range .}}{{end}}`,
	`{{range $}}{{end}}`,
	`{{fake}}`,
}

func TestFailureText(t *testing.T) {
	helpers := text.FuncMap{"fake": func() int { return 42 }}
	for _, test := range negative {
		t.Log(test)
		tmpl, err := text.New("").Funcs(helpers).Parse(test)
		if err != nil {
			t.Fatal(err)
		}
		_, err = tmpl2js.ConvertText(tmpl, &Context{}, nil)
		if err == nil {
			t.Fatalf("expecting error from: %s", test)
		}
		t.Log(err.Error())
	}
}

func TestFailureHTML(t *testing.T) {
	helpers := html.FuncMap{"fake": func() int { return 42 }}
	for _, test := range negative {
		tmpl, err := html.New("").Funcs(helpers).Parse(test)
		if err != nil {
			t.Fatal(err)
		}
		_, err = tmpl2js.ConvertHTML(tmpl, &Context{}, nil)
		if err == nil {
			t.Fatalf("expecting error from: %s", test)
		}
		t.Log(err.Error())
	}
}
