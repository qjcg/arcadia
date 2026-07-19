package cueutil

import (
	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/format"
)

// NewContext creates a new CUE context.
func NewContext() *cue.Context {
	return cuecontext.New()
}

// CompileString compiles a CUE source string into a Value.
func CompileString(ctx *cue.Context, src string) cue.Value {
	return ctx.CompileString(src)
}

// CompileBytes compiles CUE source bytes into a Value.
func CompileBytes(ctx *cue.Context, b []byte) cue.Value {
	return ctx.CompileBytes(b)
}

// EncodeGo encodes a Go value into a CUE value.
func EncodeGo(ctx *cue.Context, x any) cue.Value {
	return ctx.Encode(x)
}

// FormatValue formats a CUE value as a string.
func FormatValue(v cue.Value) (string, error) {
	syn := v.Syntax(cue.Final(), cue.Concrete(true))
	b, err := format.Node(syn)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// FormatValueRaw formats a CUE value preserving all fields (including optional/pattern).
func FormatValueRaw(v cue.Value) (string, error) {
	syn := v.Syntax()
	b, err := format.Node(syn)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// LookupPath looks up a path in a CUE value and returns the result.
// Path syntax: "field.subfield" or "field[0].subfield"
func LookupPath(v cue.Value, path string) cue.Value {
	p := cue.ParsePath(path)
	if p.Err() != nil {
		return v.Context().CompileString("_|_").LookupPath(cue.ParsePath(""))
	}
	return v.LookupPath(p)
}

// Validate checks if a value is valid according to its constraints.
func Validate(v cue.Value) error {
	return v.Validate(cue.Concrete(true))
}

// WalkFields walks all fields of a struct value, calling fn for each.
func WalkFields(v cue.Value, fn func(name string, val cue.Value) bool) {
	iter, err := v.Fields(cue.All())
	if err != nil {
		return
	}
	for iter.Next() {
		if !fn(iter.Label(), iter.Value()) {
			break
		}
	}
}

// Unify unifies two CUE values.
func Unify(a, b cue.Value) cue.Value {
	return a.Unify(b)
}

// ToJSON marshals a CUE value to JSON.
func ToJSON(v cue.Value) ([]byte, error) {
	return v.MarshalJSON()
}

// IsConcrete reports whether the value is concrete.
func IsConcrete(v cue.Value) bool {
	return v.IsConcrete()
}

// Err returns the error associated with a value, if any.
func Err(v cue.Value) error {
	return v.Err()
}
