package ast

import (
	"fmt"
)

func (l Literal) typ() Type {
	if l.FloatVal != nil {
		return number{}
	} else if l.BoolVal != nil {
		return boolean{}
	} else if l.StringVal != nil {
		return str{}
	}

	panic("should not exist")
}
func (m Method) typ() Type {
	_, typ := m.Subject.typ().FieldNamed(m.Name)
	ret := typ.(function).Return
	if ret == nil {
		panic(fmt.Sprintf("function %s returns void", m.Name))
	}
	return ret
}
func (f Field) typ() Type {
	_, typ := f.Subject.typ().FieldNamed(f.Name)
	return typ
}

func (f Local) typ() Type    { return f.T }
func (f SetLocal) typ() Type { return f.Value.typ() }
func (f Context) typ() Type  { return f.T }
func (f Global) typ() Type   { return f.S }

func (f function) String() string {
	args := ""
	for i, arg := range f.Args {
		if i != 0 {
			args += ", "
		}
		args += arg.String()
	}

	ret := ""
	if f.Return != nil {
		ret = ": " + f.Return.String()
	}
	return fmt.Sprintf("function (%s)%s", args, ret)
}
func (f function) FieldNamed(s string) (string, Type) {
	panic(fmt.Sprintf("Function has no field %#v", s))
}
func (f function) Iterate() Type {
	panic("Functions are not iterable")
}

func (o object) String() string {
	props := ""
	first := true
	for s := range o.Fields {
		label, typ := o.FieldNamed(s)

		if first {
			first = false
		} else {
			props += ", "
		}
		props += label + ": " + typ.String()
	}

	return "{" + props + "}"
}
func (o object) FieldNamed(s string) (string, Type) {
	typ, ok := o.Fields[s]
	if !ok {
		panic(fmt.Sprintf("Object %s has no field %#v", o, s))
	}
	label, ok := o.Labels[s]
	if !ok {
		label = s
	}
	return label, typ
}
func (o object) Iterate() Type {
	panic("Objects are not iterable")
}

func (a array) String() string { return fmt.Sprintf("Array.<%s>", a.Contains) }
func (a array) FieldNamed(s string) (string, Type) {
	panic(fmt.Sprintf("Array %s has no field %#v", a, s))
}
func (a array) Iterate() Type {
	return a.Contains
}

func (b boolean) String() string { return "boolean" }
func (b boolean) FieldNamed(s string) (string, Type) {
	panic(fmt.Sprintf("Boolean has no field %#v", s))
}
func (b boolean) Iterate() Type {
	panic("Booleans are not iterable")
}

func (s str) String() string { return "string" }
func (s str) FieldNamed(n string) (string, Type) {
	panic(fmt.Sprintf("Strings have no field %#v", n))
}
func (s str) Iterate() Type {
	panic("Strings are not iterable")
}

func (n number) String() string { return "number" }
func (n number) FieldNamed(s string) (string, Type) {
	panic(fmt.Sprintf("Numbers have no field %#v", s))
}
func (n number) Iterate() Type {
	panic("Numbers are not iterable")
}

// Pretty-prints the current global scope.
func (s *Scope) String() string { return "$" }

func (s *Scope) child() *Scope {
	return &Scope{
		Context:   s.Context,
		Variables: map[string]Type{},
		Parent:    s,
	}
}

// FieldNamed returns the given field of the current scope.
func (s *Scope) FieldNamed(name string) (string, Type) {
	typ, ok := s.Variables[name]
	if !ok {
		if s.Parent != nil {
			return s.Parent.FieldNamed(name)
		}
		panic(fmt.Sprintf("no global variable %s (candidates %#v)", name,
			s.Variables))
	}
	return name, typ
}

// Iterate throws, since the user cannot iterate over $.
func (s *Scope) Iterate() Type {
	panic(fmt.Errorf("cannot Iterate over global object"))
}

func fieldOrMethod(subject Expression, name string, args []Expression) Expression {
	_, typ := subject.typ().FieldNamed(name)
	switch typ.(type) {
	case function:
		return &Method{Subject: subject, Name: name, Args: args}
	default:
		if len(args) > 0 {
			panic(fmt.Sprintf("%#v is not callable", typ))
		}
		return &Field{Subject: subject, Name: name}
	}
}
