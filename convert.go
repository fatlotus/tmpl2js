package tmpl2js

import (
	"github.com/fatlotus/tmpl2js/ast"
	"reflect"
	"strings"
	"text/template/parse"
)

func minify(x string) string {
	return strings.Replace(strings.Replace(x, "\t", "", -1), "\n", "", -1)
}

var header = minify(`
(function(ctx) {
	return (function() {
		out = "";
		window = $ = this;
		var MAP = {
			'&': '&amp;',
			'<': '&lt;',
			'>': '&gt;',
			'"': '&quot;',
			"'": '&#39;'
		};
		window.$_html_template_htmlescaper = $_html_template_attrescaper = function(s) {
			return (""+s).replace(/[&<>'"]/g, function(c) {return MAP[c];});
		};
		window.$_html_template_urlescaper = function(s) {
			return encodeURIComponent("" + s);
		};
		window.$_html_template_jsvalescaper = window.$html_template_jsstrescaper = function(s) {
			console.warning("This application may have XSS vulnerabilities. Check for {{ }} inside JavaScript.");
			return s;
		}; 
		window.$json = function(s) {
			return JSON.stringify("" + s);
		};
`)

var footer = minify(`
		return out
	}).call(ctx)
})`)

func Convert(tree *parse.Tree, example_context interface{}) (string, error) {
	root_type := ast.NewType(reflect.TypeOf(example_context))
	code, err := ast.Process(tree, ast.NewScope(root_type))
	return header + code + footer, err
}
