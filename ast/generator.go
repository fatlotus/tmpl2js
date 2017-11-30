package ast

import (
	"fmt"
)

func (l *Literal) expr() string {
	if l.FloatVal != nil {
		return fmt.Sprintf("%f", *l.FloatVal)
	} else if l.BoolVal != nil {
		if *l.BoolVal {
			return "true"
		}
		return "false"
	}
	return quote(*l.StringVal)
}
func (f Method) expr() string {
	lbl, _ := f.Subject.typ().FieldNamed(f.Name)
	res := fmt.Sprintf("%s.%s(", f.Subject.expr(), lbl)
	for i, arg := range f.Args {
		if i != 0 {
			res += ", "
		}
		res += arg.expr()
	}
	res += ")"
	return res
}

func (f Field) expr() string {
	lbl, _ := f.Subject.typ().FieldNamed(f.Name)
	return fmt.Sprintf("%s.%s", f.Subject.expr(), lbl)
}

func (l Local) expr() string {
	return l.Name
}

func (c Context) expr() string { return "ctx" }

func (g Global) expr() string { return "$" }

func (sl SetLocal) stmt() string {
	return fmt.Sprintf("var %s=%s;", sl.Name, sl.Value.expr())
}

func (t Text) stmt() string {
	return fmt.Sprintf("out+=%s;", quote(t.Text))
}

func (e Append) stmt() string { return "out+=" + e.Expression.expr() + ";" }

func (l Loop) stmt() string {
	sv := ""
	if l.IndexVar != "" {
		sv = fmt.Sprintf("var %s=i,%s=it[i];", l.IndexVar, l.ValueVar)
	} else if l.ValueVar != "" {
		sv = fmt.Sprintf("var %s=it[i];", l.ValueVar)
	}

	return fmt.Sprintf(""+
		"var any=false;"+
		"var it=%s;"+
		"for(var i=0;it&&i<it.length;i++){%s;any=true;}"+
		"if(!any){%s}",
		l.Subject.expr(), l.wrap("it[i]", sv, l.Body), l.wrap("ctx", "", l.Else))
}

func (c Conditional) stmt() string {
	call := "ctx"
	if c.SetContext {
		call = "v"
	}
	sv := ""
	if c.CondVar != "" {
		sv = fmt.Sprintf("var %s=v;", c.CondVar)
	}
	return fmt.Sprintf("var v=%s;if(v){%s}else{%s}",
		c.Conditional.expr(), c.wrap(call, sv, c.Body), c.wrap("ctx", "", c.Else))
}

func (i Include) stmt() string {
	if i.Context != nil {
		return fmt.Sprintf("out+=_tmpls[%s](%s);", quote(i.Name), i.Context.expr())
	}
	return fmt.Sprintf("out+=_tmpls[%s]();", quote(i.Name))
}

func (s Scope) wrap(callee string, before string, inner []Statement) string {
	return fmt.Sprintf("(function(ctx){%s%s})(%s)", before, catStmts(inner), callee)
}

func catStmts(s []Statement) string {
	res := ""
	for _, arg := range s {
		res += arg.stmt()
	}
	return res
}
