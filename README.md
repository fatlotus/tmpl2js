# text/template ↝ JavaScript

[![codecov](https://codecov.io/gh/fatlotus/tmpl2js/branch/master/graph/badge.svg)](https://codecov.io/gh/fatlotus/tmpl2js)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/fatlotus/tmpl2js)

The `tmpl2js` package compiles Golang templates (from the `text/template`
package) into executable JavaScript. This way, developers can build responsive
applications (with dynamic, partial page updates), without writing much
JavaScript.

This code is intentionally structured as a library, to make it easier to embed
in your own work.

## Usage

On the server side (`main.go`):

```go
type Person struct {
	Name string
}

func MyApp(w http.ResponseWriter, r *http.Request) {
	// Parse the template
	tmpl, _ := template.New("").Parse(`Hello, {{.Name}}!`)

	// Compile it into a minified JavaScript function
	function, _ := tmpl2js.ConvertHTML(tmpl, &Person{}, nil)
	w.Write([]byte("<script>var _tmpl = " + function + "</script>" +
	               "<script src='app.js'></script>"))
}
```

On the client side (`app.js`):

```js
var result = _tmpl({Name: "World"});
console.log(result);  // Prints "Hello, World!"
```

### Context methods

When rendering templates on the server side, it is possible to define methods
on the template. If the functions can be translated into JavaScript, this is
also possible:

```go
type Person struct {
	Name string
}

func (p Person) Greet(greeting string) string {
	return greeting + ", " + p.Name
}

func MyApp(w http.ResponseWriter, r *http.Request) {
	// Parse the template
	tmpl, _ := template.New("").Parse(`{{.Greet "Good morning"}}!`)

	// Compile it into a minified JavaScript function
	function, _ := tmpl2js.ConvertHTML(tmpl, &Person{}, nil)
	w.Write([]byte("<script>var _tmpl = " + function + "</script>" +
	               "<script src='app.js'></script>"))
}
```

On the client side (`app.js`):

```js
var result = _tmpl({Name: "World"});
result.Greet = function(greeting) {
	return greeting + ", " + this.Name + ", from JavaScript";
}
console.log(result);  // Prints "Good morning, World, from JavaScript!"
```

### Helper functions

Global helper functions (on the server side, specified with `tmpl.FuncMap`),
are translated into methods on the global context object. For example:

```go
func Greet(greeting, name string) string {
	return greeting + ", " + p.Name
}

func MyApp(w http.ResponseWriter, r *http.Request) {
	// Define a list of helper functions
	helpers := template.FuncMap{"greet": Greet}

	// Parse the template
	tmpl, _ := template.New("").Funcs(helpers).Parse(`{{greet "Howdy" .}}!`)

	// Compile it into a minified JavaScript function
	function, _ := tmpl2js.ConvertHTML(tmpl, "", helpers)
	w.Write([]byte("<script>var _tmpl = " + function + "</script>" +
	               "<script src='app.js'></script>"))
}
```

On the client side (`app.js`):

```js
var result = _tmpl("partner");
result.greet = function(greeting, name) {
	return greeting + ", " + name + ", from JavaScript";
}
console.log(result);  // Prints "Howdy, partner, from JavaScript!"
```
