# text/template ↝ JavaScript

[![codecov](https://codecov.io/gh/fatlotus/tmpl2js/branch/master/graph/badge.svg)](https://codecov.io/gh/fatlotus/tmpl2js)

The `tmpl2js` package compiles Golang templates (from the `text/template`
package) into executable JavaScript. This way, developers can build responsive
applications (with dynamic, partial page updates), without writing much
JavaScript.

This code is intentionally structured as a library, to make it easier to embed
in your own work.

## Example

On the server side (`main.go`):

```go
type Person struct {
	Name string
}

func MyApp(w http.ResponseWriter, r *http.Request) {
	// Parse the template
	tmpl, _ := template.New("").Parse(`Hello, {{.Name}}!`)

	// Compile it into a minified JavaScript function
	function, _ := tmpl2js.ConvertHTML(tmpl, &Person{})
	w.Write([]byte("<script>var _tmpl = " + function + "</script>" +
	               "<script src="app.js"></script>"))
}
```

On the client side (`app.js`):

```js
var result = _tmpl({Name: "World"});
console.log(result);  // Prints "Hello, World!"
```

## License

All rights reserved, for the moment.