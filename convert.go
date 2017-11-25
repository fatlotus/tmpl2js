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
	$.$lt = function(a, b) {
		return a < b;
	};
	$.$_html_template_htmlescaper = $_html_template_attrescaper = function(s) {
		return (""+s).replace(/[&<>'"]/g, function(c) {return MAP[c];});
	};
	$.$_html_template_urlescaper = function(s) {
		return encodeURIComponent("" + s);
	};
	$.$_html_template_jsvalescaper = $.$html_template_jsstrescaper = function(s) {
		console.warning("This application may have XSS vulnerabilities. Check for {{ }} inside JavaScript.");
		return s;
	}; 
	$.$json = function(s) {
		return JSON.stringify("" + s);
	};
`)

var footer = minify(`
	return out
})`)

// Compiles the given template parse tree into a JavaScript function.
//
// It accepts a single argument, which is the context used for the template.
func ConvertTree(tree *parse.Tree, example_context interface{}) (string, error) {
	root_type := ast.NewType(reflect.TypeOf(example_context))
	code, err := ast.Process(tree, ast.NewScope(root_type))
	return header + code + footer, err
}

// Compiles a parsed *template.Template into a JavaScript function.
//
// It accepts a single argument ctx, which is the context used for the template.
func ConvertText(tmpl *text_template.Template, example_context interface{}) (string, error) {
	js := "(_tmpls={},"
	for _, tmpl := range tmpl.Templates() {
		new_js, err := ConvertTree(tmpl.Tree, example_context)
		if err != nil {
			return "", err
		}
		js += "_tmpls[\"" + tmpl.Tree.Name + "\"]=" + new_js + ","
	}
	return js + "_tmpls[\"\"])", nil
}

// Compiles a parsed *template.Template into a JavaScript function.
//
// It accepts a single argument ctx, which is the context used for the template.
func ConvertHTML(tmpl *html_template.Template, example_context interface{}) (string, error) {
	js := "(_tmpls={},"
	for _, tmpl := range tmpl.Templates() {
		new_js, err := ConvertTree(tmpl.Tree, example_context)
		if err != nil {
			return "", err
		}
		js += "_tmpls[\"" + tmpl.Tree.Name + "\"]=" + new_js + ","
	}
	return js + "_tmpls[\"\"])", nil
}
