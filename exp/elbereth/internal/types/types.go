package types

import (
	"fmt"
	"strings"
)

// Type represents a type in the type system
type Type interface {
	String() string
	isType()
}

// ============================================================================
// Builtin Types
// ============================================================================

var (
	Int    = &NamedType{Name: "int"}
	Float  = &NamedType{Name: "float"}
	String = &NamedType{Name: "string"}
	Bool   = &NamedType{Name: "bool"}
	Byte   = &NamedType{Name: "byte"}
	Rune   = &NamedType{Name: "rune"}
	Nil    = &NamedType{Name: "nil"}
)

// NamedType represents a named type like int, string, or a custom type
type NamedType struct {
	Name string
}

func (t *NamedType) String() string { return t.Name }
func (t *NamedType) isType()        {}

// ============================================================================
// Composite Types
// ============================================================================

// SliceType represents [T]
type SliceType struct {
	EltType Type
}

func (t *SliceType) String() string {
	return fmt.Sprintf("[]%s", t.EltType)
}
func (t *SliceType) isType() {}

// ArrayType represents [N]T
type ArrayType struct {
	Size    int64
	EltType Type
}

func (t *ArrayType) String() string {
	return fmt.Sprintf("[%d]%s", t.Size, t.EltType)
}
func (t *ArrayType) isType() {}

// MapType represents {K V}
type MapType struct {
	KeyType Type
	ValType Type
}

func (t *MapType) String() string {
	return fmt.Sprintf("map[%s]%s", t.KeyType, t.ValType)
}
func (t *MapType) isType() {}

// ChanType represents (chan T)
type ChanType struct {
	EltType Type
	Buffer  int64 // 0 for unbuffered
}

func (t *ChanType) String() string {
	if t.Buffer > 0 {
		return fmt.Sprintf("chan %s (%d)", t.EltType, t.Buffer)
	}
	return fmt.Sprintf("chan %s", t.EltType)
}
func (t *ChanType) isType() {}

// FuncType represents (fn [ParamTypes] ReturnType)
type FuncType struct {
	ParamTypes []Type
	ReturnType Type
}

func (t *FuncType) String() string {
	var sb strings.Builder
	sb.WriteString("fn(")
	for i, p := range t.ParamTypes {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(p.String())
	}
	sb.WriteByte(')')
	sb.WriteByte(' ')
	sb.WriteString(t.ReturnType.String())
	return sb.String()
}
func (t *FuncType) isType() {}

// StructType represents a struct type
type StructType struct {
	Name   string
	Fields map[string]Type
}

func (t *StructType) String() string {
	return fmt.Sprintf("struct %s", t.Name)
}
func (t *StructType) isType() {}

// UnionType represents a union/sum type
type UnionType struct {
	Name string
	Alts map[string]Type // alternative name -> type
}

func (t *UnionType) String() string {
	return fmt.Sprintf("union %s", t.Name)
}
func (t *UnionType) isType() {}

// ============================================================================
// Helper Functions
// ============================================================================

// Equal checks if two types are equal
func Equal(a, b Type) bool {
	if a == nil || b == nil {
		return a == b
	}

	switch ta := a.(type) {
	case *NamedType:
		tb, ok := b.(*NamedType)
		return ok && ta.Name == tb.Name
	case *SliceType:
		tb, ok := b.(*SliceType)
		return ok && Equal(ta.EltType, tb.EltType)
	case *ArrayType:
		tb, ok := b.(*ArrayType)
		return ok && ta.Size == tb.Size && Equal(ta.EltType, tb.EltType)
	case *MapType:
		tb, ok := b.(*MapType)
		return ok && Equal(ta.KeyType, tb.KeyType) && Equal(ta.ValType, tb.ValType)
	case *ChanType:
		tb, ok := b.(*ChanType)
		return ok && Equal(ta.EltType, tb.EltType) && ta.Buffer == tb.Buffer
	case *FuncType:
		tb, ok := b.(*FuncType)
		if !ok || len(ta.ParamTypes) != len(tb.ParamTypes) {
			return false
		}
		for i := range ta.ParamTypes {
			if !Equal(ta.ParamTypes[i], tb.ParamTypes[i]) {
				return false
			}
		}
		return Equal(ta.ReturnType, tb.ReturnType)
	case *StructType:
		tb, ok := b.(*StructType)
		return ok && ta.Name == tb.Name
	}

	return false
}

// IsAssignableTo checks if type `from` can be assigned to type `to`
func IsAssignableTo(from, to Type) bool {
	return Equal(from, to)
}

// ParseTypeString parses a type name to a type
func ParseTypeString(name string) Type {
	switch name {
	case "int":
		return Int
	case "float":
		return Float
	case "string":
		return String
	case "bool":
		return Bool
	case "byte":
		return Byte
	case "rune":
		return Rune
	case "nil":
		return Nil
	default:
		return &NamedType{Name: name}
	}
}
