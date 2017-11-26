package ast

import (
	"fmt"
	"reflect"
	"strings"
)

func newMethod(recv bool, t reflect.Type) Type {
	i := 0
	if recv {
		i = 1
	}

	f := function{Args: []Type{}, Return: nil}
	for ; i < t.NumIn(); i++ {
		f.Args = append(f.Args, NewType(t.In(i)))
	}

	switch t.NumOut() {
	case 0:
	case 1:
		f.Return = NewType(t.Out(0))
	default:
		panic("TODO: cannot handle non-singular return values")
	}

	return f
}

func NewType(t reflect.Type) Type {
	switch t.Kind() {
	case reflect.Ptr:
		return NewType(t.Elem())
	case reflect.Bool:
		return boolean{}
	case reflect.Int,
		reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Uintptr, reflect.Float32, reflect.Float64:
		return number{}
	case reflect.Array, reflect.Slice:
		return array{Contains: NewType(t.Elem())}
	case reflect.String:
		return str{}
	case reflect.Struct:
		o := &object{
			Fields: map[string]Type{},
			Labels: map[string]string{},
		}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			o.Fields[f.Name] = NewType(f.Type)

			// Extract `json:"value"` struct tag
			tag := f.Tag.Get("json")
			i := strings.Index(tag, ",")
			if i < 0 {
				i = len(tag)
			}
			if tag != "" {
				o.Labels[f.Name] = tag[:i]
			}
		}

		for i := 0; i < t.NumMethod(); i++ {
			m := t.Method(i)
			o.Fields[m.Name] = newMethod(true, m.Type)
		}
		return o
	case reflect.Func:
		return newMethod(false, t)
	default:
		panic(fmt.Sprintf("cannot process type %s", t))
	}
}

// Creates a global template context ready for the given root object.
// (Primarily, this means setting things like $ and lt).
func NewScope(ctx Type) *Scope {
	return &Scope{
		Context: ctx,
		Variables: map[string]Type{
			"$": ctx,
			"$lt": function{
				Args:   []Type{number{}, number{}},
				Return: boolean{},
			},
			"$le": function{
				Args:   []Type{number{}, number{}},
				Return: boolean{},
			},
			"$ne": function{
				Args:   []Type{number{}, number{}},
				Return: boolean{},
			},
			"$gt": function{
				Args:   []Type{number{}, number{}},
				Return: boolean{},
			},
			"$ge": function{
				Args:   []Type{number{}, number{}},
				Return: boolean{},
			},
			"$eq": function{
				Args:   []Type{number{}, number{}},
				Return: boolean{},
			},
			"$printf": function{
				Args:   []Type{str{}, str{}},
				Return: str{},
			},
			"$_html_template_htmlescaper": function{
				Args:   []Type{str{}},
				Return: str{},
			},
			"$_html_template_urlescaper": function{
				Args:   []Type{str{}},
				Return: str{},
			},
			"$_html_template_attrescaper": function{
				Args:   []Type{str{}},
				Return: str{},
			},
			"$_html_template_jsvalescaper": function{
				Args:   []Type{str{}},
				Return: str{},
			},
			"$_html_template_jsstrescaper": function{
				Args:   []Type{str{}},
				Return: str{},
			},
			"$json": function{
				Args:   []Type{str{}},
				Return: str{},
			},
		},
	}
}
