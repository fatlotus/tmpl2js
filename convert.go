package tmpl2js

import (
	"github.com/fatlotus/tmpl2js/ast"
	"reflect"
	"strings"

	html_template "html/template"
	text_template "text/template"
	"text/template/parse"
)

func minify(x string) string {
	return strings.Replace(strings.Replace(x, "\t", "", -1), "\n", "", -1)
}

var header = minify(`
(function(ctx) {
	var out = "";
	var $ = ctx || {};
	var MAP = {
		'&': '&amp;',
		'<': '&lt;',
		'>': '&gt;',
		'"': '&quot;',
		"'": '&#39;'
	};
	$.$le = function(a, b) { return a <= b };
	$.$lt = function(a, b) { return a < b };
	$.$gt = function(a, b) { return a > b };
	$.$ge = function(a, b) { return a >= b };
	$.$eq = function(a, b) { return a == b };
	$.$ne = function(a, b) { return a != b };
	$.$_html_template_htmlescaper = $_html_template_attrescaper = function(s) {
		return (""+s).replace(/[&<>'"]/g, function(c) {return MAP[c];});
	};
	$.$_html_template_urlescaper = function(s) {
		return encodeURIComponent("" + s);
	};
	$.$_html_template_jsvalescaper = $.$html_template_jsstrescaper = function(s) {
		throw new Error(
			"This application is trying to embed template values inside a " +
			"script tag {{ }}, which is unsupported."
		);
	};
	$.$json = function(s) {
		return JSON.stringify("" + s);
	};
`)

var footer = minify(`
	return out
})`)

// ConvertTree converts the given template parse tree into a JavaScript
// function.
//
// It accepts a single argument, which is the context used for the template.
func ConvertTree(tree *parse.Tree, exampleContext interface{}, funcMap map[string]interface{}) (string, error) {
	rootType := ast.NewType(reflect.TypeOf(exampleContext))
	scope := ast.NewScope(rootType)
	for key, value := range funcMap {
		scope.Variables["$"+key] = ast.NewType(reflect.TypeOf(value))
	}
	code, err := ast.Process(tree, scope)
	return header + code + footer, err
}

// ConvertText compiles a parsed *template.Template into a JavaScript function.
//
// It accepts a single argument ctx, which is the context used for the template.
func ConvertText(tmpl *text_template.Template, exampleContext interface{}, funcMap text_template.FuncMap) (string, error) {
	js := "(_tmpls={},"
	name := ""
	for _, tmpl := range tmpl.Templates() {
		newJs, err := ConvertTree(tmpl.Tree, exampleContext, funcMap)
		if err != nil {
			return "", err
		}
		js += "_tmpls[\"" + tmpl.Tree.Name + "\"]=" + newJs + ","
		name = tmpl.Tree.Name
	}
	return js + "_tmpls[\"" + name + "\"])", nil
}

// ConvertHTML compiles a parsed *template.Template into a JavaScript function.
//
// It accepts a single argument ctx, which is the context used for the template.
func ConvertHTML(tmpl *html_template.Template, exampleContext interface{}, funcMap html_template.FuncMap) (string, error) {
	js := "(_tmpls={},"
	name := ""
	for _, tmpl := range tmpl.Templates() {
		newJs, err := ConvertTree(tmpl.Tree, exampleContext, funcMap)
		if err != nil {
			return "", err
		}
		js += "_tmpls[\"" + tmpl.Tree.Name + "\"]=" + newJs + ","
		name = tmpl.Tree.Name
	}
	return js + "_tmpls[\"" + name + "\"])", nil
}
