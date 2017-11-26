package tmpl2js_test

import (
	"fmt"
	"github.com/fatlotus/tmpl2js"
	"github.com/robertkrimen/otto"
	"html/template"
)

type Person struct {
	Name string
}

func Example() {
	// Parse the template
	tmpl, _ := template.New("").Parse(`Hello, {{.Name}}!`)

	// Compile it into a minified JavaScript function
	function, _ := tmpl2js.ConvertHTML(tmpl, &Person{}, nil)

	// Evaluate the JavaScript by invoking it
	_, result, _ := otto.Run(function + "({Name: 'World'})")
	fmt.Printf("Result: %s", result)

	// Output: Result: Hello, World!
}
