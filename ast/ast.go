package ast

// An Expression is a snippet of JavaScript code that can be evaluated.
//
// Expressions are always one of:
//  Literal      "foo bar"
//  Field        (a.b)
//  Method       (a.b(...))
//  Local        (a)
//  SetLocal     (a = (...))
//  Context      this
//  Global       window
type Expression interface {
	expr() string
	typ() Type
}

// A Statement is snippet of JavaScript code that modifies the surrounding
// environment.
//
// Statements are always one of:
//  Append         out += (...);
//  Text           out += "(...)";
//  Conditional    if (...) { ... } else { ... }
//  Loop           while (...) { ... }
//  Include        out += renderTemplate("name", ...);
type Statement interface {
	stmt() string
}

// Expressions

// A Literal is a JSON value that can be injected into the template.
type Literal struct {
	FloatVal  *float64
	BoolVal   *bool
	StringVal *string
}

// A Field accesses a named property of an object.
type Field struct {
	Subject Expression
	Name    string
}

// A Method accesses and invokes a named property of an object.
type Method struct {
	Subject Expression
	Name    string
	Args    []Expression
}

// A Local reads a given local variable from the environment.
type Local struct {
	Name string
	T    Type
}

// A Context returns the current object (this).
type Context struct {
	T Type
}

// A Global returns $ (also known as window).
type Global struct {
	S *Scope
}

// Statements

// A Text statement writes a string to the result.
type Text struct {
	Text string
}

// An Append statement evaluates an expression and writes it out.
type Append struct {
	Expression Expression
}

// A loop iterates over the given subject; if it never loops, then
// the else branch is run.
type Loop struct {
	Subject Expression
	Body    []Statement
	Else    []Statement

	// If not empty, set this variable to the value of the index.
	IndexVar string

	// If not empty, set this variable to current iterate.
	ValueVar string
	*Scope
}

// A Conditional conditionally runs the Body block, or, if false, the Else
// block.
type Conditional struct {
	Conditional Expression
	Body        []Statement
	Else        []Statement

	// If true, sets the current context to the result of the conditional.
	SetContext bool

	// If not empty, set this variable to the result of the conditional.
	CondVar string
	*Scope
}

// An Include evaluates the given template with the given context expression.
type Include struct {
	Name    string
	Context Expression
}

// A SetLocal modifies the current scope and sets the given variable.
type SetLocal struct {
	Name  string
	Value Expression
}

// Types

// A Function is a JavaScript callable with the given (optional) return Type.
type function struct {
	Args []Type

	// If null, function returns void.
	Return Type
}

// An Object is a struct with fixed properties.
type object struct {
	Fields map[string]Type

	// Since `json:"tag"` struct tags might rename the properties,
	// allow the use of different labels for certain fields.
	Labels map[string]string
}

// An Array is a container of many objects of the same type.
type array struct {
	Contains Type
}

// A boolean is either true or false.
type boolean struct{}

// A Number is a JavaScipt Number ~ float64.
type number struct{}

// A String is a JavaScript UTF-8 string.
type str struct{}

// A Type is a JavaScript type.
type Type interface {
	// Returns the label and type of the given field.
	// Panics if the given field does not exist.
	FieldNamed(name string) (string, Type)

	// Returns code to iterate over the given object.
	// Panics of the given object does not exist.
	Iterate() Type
	String() string
}

// A Scope is the type of the Context (the "this" object) plus
// all variables in the current scope.
type Scope struct {
	Context   Type
	Variables map[string]Type
	Parent    *Scope
}
