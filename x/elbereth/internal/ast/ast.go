package ast

import "fmt"

// Position represents a location in source code
type Position struct {
	Line   int
	Column int
}

// Node is the base interface for all AST nodes
type Node interface {
	Pos() Position
	String() string
}

// Expr represents any expression
type Expr interface {
	Node
	exprNode()
}

// Type represents a type annotation
type Type interface {
	Node
	typeNode()
}

// ============================================================================
// Literals
// ============================================================================

// IntLit represents an integer literal
type IntLit struct {
	Loc   Position
	Value int64
}

func (n *IntLit) exprNode()      {}
func (n *IntLit) Pos() Position  { return n.Loc }
func (n *IntLit) String() string { return fmt.Sprintf("%d", n.Value) }

// FloatLit represents a float literal
type FloatLit struct {
	Loc   Position
	Value float64
}

func (n *FloatLit) exprNode()      {}
func (n *FloatLit) Pos() Position  { return n.Loc }
func (n *FloatLit) String() string { return fmt.Sprintf("%f", n.Value) }

// StringLit represents a string literal
type StringLit struct {
	Loc   Position
	Value string
}

func (n *StringLit) exprNode()      {}
func (n *StringLit) Pos() Position  { return n.Loc }
func (n *StringLit) String() string { return fmt.Sprintf(`"%s"`, n.Value) }

// KeywordLit represents a keyword literal
type KeywordLit struct {
	Loc   Position
	Value string // without the :
}

func (n *KeywordLit) exprNode()      {}
func (n *KeywordLit) Pos() Position  { return n.Loc }
func (n *KeywordLit) String() string { return fmt.Sprintf(":%s", n.Value) }

// BoolLit represents a boolean literal
type BoolLit struct {
	Loc   Position
	Value bool
}

func (n *BoolLit) exprNode()      {}
func (n *BoolLit) Pos() Position  { return n.Loc }
func (n *BoolLit) String() string { return fmt.Sprintf("%v", n.Value) }

// NilLit represents a nil literal
type NilLit struct {
	Loc Position
}

func (n *NilLit) exprNode()      {}
func (n *NilLit) Pos() Position  { return n.Loc }
func (n *NilLit) String() string { return "nil" }

// VectorLit represents a vector [a b c]
type VectorLit struct {
	Loc  Position
	Elts []Expr
}

func (n *VectorLit) exprNode()      {}
func (n *VectorLit) Pos() Position  { return n.Loc }
func (n *VectorLit) String() string { return fmt.Sprintf("[%v]", n.Elts) }

// MapLit represents a map {:key value}
type MapLit struct {
	Loc   Position
	Pairs []Pair
}

type Pair struct {
	Key   Expr
	Value Expr
}

func (n *MapLit) exprNode()      {}
func (n *MapLit) Pos() Position  { return n.Loc }
func (n *MapLit) String() string { return fmt.Sprintf("{%v}", n.Pairs) }

// ============================================================================
// Identifiers & Symbols
// ============================================================================

// Symbol represents a symbol/identifier
type Symbol struct {
	Loc  Position
	Name string
}

func (n *Symbol) exprNode()      {}
func (n *Symbol) Pos() Position  { return n.Loc }
func (n *Symbol) String() string { return n.Name }

// ============================================================================
// Function-related
// ============================================================================

// FuncLit represents an anonymous function (fn [args] body)
type FuncLit struct {
	Loc    Position
	Params []*Param
	Body   []Expr // implied do block
}

type Param struct {
	Name     string
	Type     Type // nil if not specified
	Variadic bool
}

func (n *FuncLit) exprNode()      {}
func (n *FuncLit) Pos() Position  { return n.Loc }
func (n *FuncLit) String() string { return fmt.Sprintf("(fn %v %v)", n.Params, n.Body) }

// FuncCall represents a function call (f arg1 arg2)
type FuncCall struct {
	Loc  Position
	Func Expr
	Args []Expr
}

func (n *FuncCall) exprNode()      {}
func (n *FuncCall) Pos() Position  { return n.Loc }
func (n *FuncCall) String() string { return fmt.Sprintf("(%v %v)", n.Func, n.Args) }

// ============================================================================
// Special Forms
// ============================================================================

// IfExpr represents (if cond then else)
type IfExpr struct {
	Loc  Position
	Cond Expr
	Then Expr
	Else Expr // can be nil
}

func (n *IfExpr) exprNode()      {}
func (n *IfExpr) Pos() Position  { return n.Loc }
func (n *IfExpr) String() string { return fmt.Sprintf("(if %v %v %v)", n.Cond, n.Then, n.Else) }

// DoExpr represents (do expr1 expr2 ...)
type DoExpr struct {
	Loc   Position
	Exprs []Expr
}

func (n *DoExpr) exprNode()      {}
func (n *DoExpr) Pos() Position  { return n.Loc }
func (n *DoExpr) String() string { return fmt.Sprintf("(do %v)", n.Exprs) }

// LetExpr represents (let [x 1 y 2] expr)
type LetExpr struct {
	Loc      Position
	Bindings []*Binding
	Body     []Expr
}

type Binding struct {
	Name string
	Init Expr
	Type Type // nil if not specified
}

func (n *LetExpr) exprNode()      {}
func (n *LetExpr) Pos() Position  { return n.Loc }
func (n *LetExpr) String() string { return fmt.Sprintf("(let [...] %v)", n.Body) }

// SelectExpr represents (select [chan val] body ...)
type SelectExpr struct {
	Loc   Position
	Cases []SelectCase
}

type SelectCase struct {
	Chan    Expr // nil if default
	Binding string
	Body    []Expr
}

func (n *SelectExpr) exprNode()      {}
func (n *SelectExpr) Pos() Position  { return n.Loc }
func (n *SelectExpr) String() string { return fmt.Sprintf("(select ...)") }

// RecurExpr represents (recur expr1 expr2 ...)
type RecurExpr struct {
	Loc  Position
	Args []Expr
}

func (n *RecurExpr) exprNode()      {}
func (n *RecurExpr) Pos() Position  { return n.Loc }
func (n *RecurExpr) String() string { return fmt.Sprintf("(recur ...)") }

// LoopExpr represents (loop [bindings] body)
type LoopExpr struct {
	Loc      Position
	Bindings []*Binding
	Body     []Expr
}

func (n *LoopExpr) exprNode()      {}
func (n *LoopExpr) Pos() Position  { return n.Loc }
func (n *LoopExpr) String() string { return fmt.Sprintf("(loop ...)") }

// MatchExpr represents (match val pattern1 body1 ...)
type MatchExpr struct {
	Loc   Position
	Val   Expr
	Cases []MatchCase
}

type MatchCase struct {
	Pattern Expr
	Body    Expr
}

func (n *MatchExpr) exprNode()      {}
func (n *MatchExpr) Pos() Position  { return n.Loc }
func (n *MatchExpr) String() string { return fmt.Sprintf("(match %v ...)", n.Val) }

// QuoteExpr represents (quote expr) or 'expr
type QuoteExpr struct {
	Loc  Position
	Expr Expr
}

func (n *QuoteExpr) exprNode()      {}
func (n *QuoteExpr) Pos() Position  { return n.Loc }
func (n *QuoteExpr) String() string { return fmt.Sprintf("'%v", n.Expr) }

// BackquoteExpr represents `expr
type BackquoteExpr struct {
	Loc  Position
	Expr Expr
}

func (n *BackquoteExpr) exprNode()      {}
func (n *BackquoteExpr) Pos() Position  { return n.Loc }
func (n *BackquoteExpr) String() string { return fmt.Sprintf("`%v", n.Expr) }

// UnquoteExpr represents ,expr
type UnquoteExpr struct {
	Loc  Position
	Expr Expr
}

func (n *UnquoteExpr) exprNode()      {}
func (n *UnquoteExpr) Pos() Position  { return n.Loc }
func (n *UnquoteExpr) String() string { return fmt.Sprintf(",%v", n.Expr) }

// UnquoteSpliceExpr represents ,@expr
type UnquoteSpliceExpr struct {
	Loc  Position
	Expr Expr
}

func (n *UnquoteSpliceExpr) exprNode()      {}
func (n *UnquoteSpliceExpr) Pos() Position  { return n.Loc }
func (n *UnquoteSpliceExpr) String() string { return fmt.Sprintf(",@%v", n.Expr) }

// ============================================================================
// Definitions (Top-level)
// ============================================================================

// Def represents (def name value) or (def name type value)
type Def struct {
	Loc   Position
	Name  string
	Type  Type // can be nil
	Value Expr
}

func (n *Def) Pos() Position  { return n.Loc }
func (n *Def) String() string { return fmt.Sprintf("(def %s %v)", n.Name, n.Value) }

// Defn represents (defn name [params] body)
type Defn struct {
	Loc        Position
	Name       string
	Params     []*Param
	ReturnType Type // can be nil
	Body       []Expr
}

func (n *Defn) Pos() Position  { return n.Loc }
func (n *Defn) String() string { return fmt.Sprintf("(defn %s %v %v)", n.Name, n.Params, n.Body) }

// Deftype represents (deftype Name {field Type ...})
type Deftype struct {
	Loc      Position
	Name     string
	Params   []string   // Generics like :T
	Fields   []*Field   // For structs
	Variants []*Variant // For sum types
}

type Field struct {
	Name    string
	Type    Type
	Default Expr // can be nil
}

type Variant struct {
	Name string
	Type Type // nil if no associated data
}

func (n *Deftype) Pos() Position  { return n.Loc }
func (n *Deftype) String() string { return fmt.Sprintf("(deftype %s ...)", n.Name) }

// Defmacro represents (defmacro name [params] body)
type Defmacro struct {
	Loc    Position
	Name   string
	Params []*Param
	Body   []Expr
}

func (n *Defmacro) Pos() Position  { return n.Loc }
func (n *Defmacro) String() string { return fmt.Sprintf("(defmacro %s ...)", n.Name) }

// Deftest represents (deftest name body...)
type Deftest struct {
	Loc  Position
	Name string
	Body []Expr
}

func (n *Deftest) Pos() Position  { return n.Loc }
func (n *Deftest) String() string { return fmt.Sprintf("(deftest %s ...)", n.Name) }

// Defbenchmark represents (defbenchmark name [b] body...)
type Defbenchmark struct {
	Loc    Position
	Name   string
	BParam string // name of the *testing.B parameter
	Body   []Expr
}

func (n *Defbenchmark) Pos() Position  { return n.Loc }
func (n *Defbenchmark) String() string { return fmt.Sprintf("(defbenchmark %s ...)", n.Name) }

// Defexample represents (defexample name body...)
type Defexample struct {
	Loc  Position
	Name string
	Body []Expr
}

func (n *Defexample) Pos() Position  { return n.Loc }
func (n *Defexample) String() string { return fmt.Sprintf("(defexample %s ...)", n.Name) }

// Package represents (package name)
type Package struct {
	Loc  Position
	Name string
}

func (n *Package) Pos() Position  { return n.Loc }
func (n *Package) String() string { return fmt.Sprintf("(package %s)", n.Name) }

// Import represents (import "path") or (import [alias "path"])
type Import struct {
	Loc   Position
	Specs []ImportSpec
}

type ImportSpec struct {
	Path  string
	Alias string // empty if no alias
}

func (n *Import) Pos() Position { return n.Loc }
func (n *Import) String() string {
	res := "(import"
	for _, spec := range n.Specs {
		if spec.Alias != "" {
			res += fmt.Sprintf(" [%s %q]", spec.Alias, spec.Path)
		} else {
			res += fmt.Sprintf(" %q", spec.Path)
		}
	}
	res += ")"
	return res
}

// ============================================================================
// Types
// ============================================================================

// NamedType represents a named type like :int or :string
type NamedType struct {
	Loc  Position
	Name string
}

func (n *NamedType) typeNode()      {}
func (n *NamedType) Pos() Position  { return n.Loc }
func (n *NamedType) String() string { return n.Name }

// SliceType represents :[T]
type SliceType struct {
	Loc     Position
	EltType Type
}

func (n *SliceType) typeNode()      {}
func (n *SliceType) Pos() Position  { return n.Loc }
func (n *SliceType) String() string { return fmt.Sprintf("[%v]", n.EltType) }

// MapType represents :{K V}
type MapType struct {
	Loc     Position
	KeyType Type
	ValType Type
}

func (n *MapType) typeNode()      {}
func (n *MapType) Pos() Position  { return n.Loc }
func (n *MapType) String() string { return fmt.Sprintf("{%v %v}", n.KeyType, n.ValType) }

// ChanType represents (chan T) or (chan T n)
type ChanType struct {
	Loc     Position
	EltType Type
	Buffer  int64 // 0 for unbuffered
}

func (n *ChanType) typeNode()      {}
func (n *ChanType) Pos() Position  { return n.Loc }
func (n *ChanType) String() string { return fmt.Sprintf("(chan %v)", n.EltType) }

// FuncType represents (fn [ParamTypes] ReturnType)
type FuncType struct {
	Loc        Position
	ParamTypes []Type
	ReturnType Type
}

func (n *FuncType) typeNode()      {}
func (n *FuncType) Pos() Position  { return n.Loc }
func (n *FuncType) String() string { return "fn" }

// ============================================================================
// Program (root node)
// ============================================================================

// Program represents an entire Elbereth program
type Program struct {
	Loc     Position
	Package string // inferred or declared via (defmodule)
	Items   []Node // mix of Def, Defn, Deftype, Defmacro, or Expr at top level
}

func (n *Program) Pos() Position  { return n.Loc }
func (n *Program) String() string { return fmt.Sprintf("(program %v items)", len(n.Items)) }
