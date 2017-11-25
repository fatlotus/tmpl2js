package ast

import (
	"encoding/json"
	"fmt"
	"strconv"
	"text/template/parse"
)

func quote(in string) string {
	data, err := json.Marshal(in)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func processPipe(n *parse.PipeNode, sc *Scope) Expression {
	if n == nil {
		return nil
	}

	// handle pipelines via nested function calls
	callee := Expression(nil)

	for _, c := range n.Cmds {
		args := make([]Expression, len(c.Args)-1)
		for i, a := range c.Args[1:] {
			args[i] = processExpr(a, sc)
		}
		if callee != Expression(nil) {
			args = append(args, callee)
		}

		switch arg := c.Args[0].(type) {
		case *parse.IdentifierNode: // {{ "foo" | func }}
			callee = fieldOrMethod(
				&Global{S: sc},
				"$"+c.Args[0].(*parse.IdentifierNode).Ident,
				args,
			)
		case *parse.FieldNode: // {{ "foo" | obj.method 4 }}
			callee = processExpr(arg, sc)
			if len(args) > 0 {
				callee.(*Method).Args = args
			}
		default:
			callee = processExpr(arg, sc)
		}
	}

	if callee == nil {
		panic("callee is nil!")
	}

	// By this point, variables (n.Decl) have already been set
	// by the container.

	return callee
}

func processExpr(n parse.Node, sc *Scope) Expression {
	switch n := n.(type) {
	case *parse.NumberNode:
		value, err := strconv.ParseFloat(n.Text, 64)
		if err != nil {
			panic(err)
		}
		return &Literal{FloatVal: &value}
	case *parse.StringNode:
		return &Literal{StringVal: &n.Text}
	case *parse.BoolNode:
		return &Literal{BoolVal: &n.True}
	case *parse.PipeNode:
		return processPipe(n, sc)
	case *parse.DotNode:
		return &Context{T: sc.Context}
	case *parse.FieldNode:
		callee := Expression(&Context{T: sc.Context})
		for _, field := range n.Ident {
			callee = fieldOrMethod(callee, field, nil)
		}

		return callee
	case *parse.IdentifierNode:
		_, typ := sc.FieldNamed(n.Ident)
		return Local{Name: "$" + n.Ident, T: typ}
	case *parse.VariableNode:
		_, typ := sc.FieldNamed(n.Ident[0])
		callee := Expression(&Local{Name: n.Ident[0], T: typ})
		for _, field := range n.Ident[1:] {
			callee = fieldOrMethod(callee, field, nil)
		}

		return callee
	default:
		panic(fmt.Sprintf("unknown expr: %#v\n", n))
	}
}

func processStmts(ln *parse.ListNode, sc *Scope) []Statement {
	if ln == nil {
		return nil
	}
	res := make([]Statement, len(ln.Nodes))
	for i, node := range ln.Nodes {
		res[i] = processStmt(node, sc)
	}
	return res
}

func processExprCtx(n parse.Node, sc *Scope) Expression {
	expr := processExpr(n, sc)
	sc.Context = expr.typ()
	return expr
}

func vToName(v *parse.VariableNode) string {
	if len(v.Ident) != 1 {
		panic("TODO: multi-level assignment")
	}
	return v.Ident[0]
}

func extractVar(n *parse.PipeNode, t Type, sc *Scope) string {
	switch len(n.Decl) {
	case 0:
		return ""
	case 1:
		v := vToName(n.Decl[0])
		sc.Variables[v] = t
		return v
	default:
		panic("TODO: multiple assignment in if/with block")
	}
}

func extractVars(n *parse.PipeNode, t Type, sc *Scope) (index, value string) {
	switch len(n.Decl) {
	case 0:
		return "", ""
	case 1:
		b := vToName(n.Decl[0])
		sc.Variables[b] = t
		return "", b
	case 2:
		a, b := vToName(n.Decl[0]), vToName(n.Decl[1])
		sc.Variables[a] = number{}
		sc.Variables[b] = t
		return a, b
	default:
		panic("TODO: multiple assignment in if/with block")
	}
}

func processStmt(n parse.Node, sc *Scope) Statement {
	defer addCtxToPanics(n)

	switch n := n.(type) {
	case *parse.TextNode:
		return &Text{string(n.Text)}
	case *parse.IfNode:
		sub := sc.child()
		cond := processExpr(n.Pipe, sub)
		return &Conditional{
			Conditional: cond,
			SetContext:  false,
			CondVar:     extractVar(n.Pipe, cond.typ(), sub),
			Body:        processStmts(n.List, sub),
			Else:        processStmts(n.ElseList, sc),
			Scope:       sub,
		}
	case *parse.WithNode:
		sub := sc.child()
		cond := processExprCtx(n.Pipe, sub)
		return &Conditional{
			Conditional: cond,
			SetContext:  true,
			CondVar:     extractVar(n.Pipe, cond.typ(), sub),
			Body:        processStmts(n.List, sub),
			Else:        processStmts(n.ElseList, sc),
			Scope:       sub,
		}
	case *parse.RangeNode:
		sub := sc.child()
		subj := processExprCtx(n.Pipe, sub)
		sub.Context = sub.Context.Iterate()
		index, value := extractVars(n.Pipe, subj.typ().Iterate(), sub)
		return &Loop{
			Subject:  subj,
			Body:     processStmts(n.List, sub),
			Else:     processStmts(n.ElseList, sc),
			IndexVar: index,
			ValueVar: value,
			Scope:    sub,
		}
	case *parse.ActionNode:
		switch len(n.Pipe.Decl) {
		case 0:
			return &Append{processExpr(n.Pipe, sc)}
		case 1:
			inside := processExpr(n.Pipe, sc)
			extractVar(n.Pipe, inside.typ(), sc)
			return &SetLocal{n.Pipe.Decl[0].Ident[0], inside}
		default:
			panic("TODO: multiple assignment?")
		}
	case *parse.TemplateNode:
		return &Include{
			Name:    n.Name,
			Context: processExpr(n.Pipe, sc),
		}
	default:
		panic(fmt.Sprintf("unknown stmt: %#v\n", n))
	}
}

func Process(t *parse.Tree, sc *Scope) (r string, err error) {
	defer func() {
		exc := recover()
		if cerr, ok := exc.(contextError); ok {
			cerr.Tree = t
			err = cerr
		} else {
			if exc != nil {
				panic(exc)
			}
		}
	}()

	r = catStmts(processStmts(t.Root, sc))
	return
}
