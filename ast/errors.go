package ast

import (
	"fmt"
	"text/template/parse"
)

type contextError struct {
	Tree *parse.Tree
	Node parse.Node
	Msg  string
}

func (c contextError) Error() string {
	loc, ctx := c.Tree.ErrorContext(c.Node)
	return fmt.Sprintf("javascript/template: %s: executing at \"%s\": %s",
		loc, ctx, c.Msg)
}

func addCtxToPanics(n parse.Node) {
	x := recover()
	if x == nil {
		return
	}
	if msg, ok := x.(string); ok {
		panic(contextError{Node: n, Msg: msg})
	} else {
		panic(x)
	}
}
