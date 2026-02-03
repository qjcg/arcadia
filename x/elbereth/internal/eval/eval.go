package eval

import (
	"fmt"
	"math"
	"strings"

	"github.com/qjcg/arcadia/x/elbereth/internal/ast"
)

// Value represents a runtime value
type Value interface {
	TypeName() string
	String() string
}

// IntVal represents an integer value
type IntVal struct {
	Value int64
}

func (v IntVal) TypeName() string { return "int" }
func (v IntVal) String() string   { return fmt.Sprintf("%d", v.Value) }

// FloatVal represents a float value
type FloatVal struct {
	Value float64
}

func (v FloatVal) TypeName() string { return "float" }
func (v FloatVal) String() string   { return fmt.Sprintf("%f", v.Value) }

// StringVal represents a string value
type StringVal struct {
	Value string
}

func (v StringVal) TypeName() string { return "string" }
func (v StringVal) String() string   { return v.Value }

// BoolVal represents a boolean value
type BoolVal struct {
	Value bool
}

func (v BoolVal) TypeName() string { return "bool" }
func (v BoolVal) String() string   { return fmt.Sprintf("%v", v.Value) }

// NilVal represents nil
type NilVal struct{}

func (v NilVal) TypeName() string { return "nil" }
func (v NilVal) String() string   { return "nil" }

// VectorVal represents a vector
type VectorVal struct {
	Elements []Value
}

func (v VectorVal) TypeName() string { return "vector" }
func (v VectorVal) String() string {
	var sb strings.Builder
	sb.WriteByte('[')
	for i, el := range v.Elements {
		if i > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(el.String())
	}
	sb.WriteByte(']')
	return sb.String()
}

// MapVal represents a map
type MapVal struct {
	Pairs map[string]Value
}

func (v MapVal) TypeName() string { return "map" }
func (v MapVal) String() string {
	var sb strings.Builder
	sb.WriteByte('{')
	first := true
	for k, val := range v.Pairs {
		if !first {
			sb.WriteByte(' ')
		}
		sb.WriteString(k)
		sb.WriteByte(' ')
		sb.WriteString(val.String())
		first = false
	}
	sb.WriteByte('}')
	return sb.String()
}

// NodeVal represents an AST node as a value
type NodeVal struct {
	Node ast.Node
}

func (v NodeVal) TypeName() string { return "node" }
func (v NodeVal) String() string   { return v.Node.String() }

// FuncVal represents a user-defined function
type FuncVal struct {
	Name   string
	Params []*ast.Param
	Body   []ast.Expr
	Env    *Env
}

func (v FuncVal) TypeName() string { return "function" }
func (v FuncVal) String() string   { return fmt.Sprintf("#<function %s>", v.Name) }

// BuiltinVal represents a built-in function
type BuiltinVal struct {
	Name string
	Fn   func(args []Value) (Value, error)
}

func (v BuiltinVal) TypeName() string { return "builtin" }
func (v BuiltinVal) String() string   { return fmt.Sprintf("#<builtin %s>", v.Name) }

// Env represents an evaluation environment
type Env struct {
	parent *Env
	vars   map[string]Value
}

func NewEnv(parent *Env) *Env {
	return &Env{
		parent: parent,
		vars:   make(map[string]Value),
	}
}

func (e *Env) Get(name string) (Value, bool) {
	if val, ok := e.vars[name]; ok {
		return val, true
	}
	if e.parent != nil {
		return e.parent.Get(name)
	}
	return nil, false
}

func (e *Env) Set(name string, val Value) {
	e.vars[name] = val
}

// Evaluator evaluates Elbereth expressions
type Evaluator struct {
	Global     *Env
	Namespaces map[string]*Env
}

// New creates a new evaluator
func New() *Evaluator {
	e := &Evaluator{
		Global:     NewEnv(nil),
		Namespaces: make(map[string]*Env),
	}
	e.registerBuiltins()
	return e
}

func (e *Evaluator) registerBuiltins() {
	e.Global.Set("+", BuiltinVal{Name: "+", Fn: e.wrapBinaryIntFloat(func(a, b float64) float64 { return a + b }, func(a, b int64) int64 { return a + b })})
	e.Global.Set("-", BuiltinVal{Name: "-", Fn: e.wrapBinaryIntFloat(func(a, b float64) float64 { return a - b }, func(a, b int64) int64 { return a - b })})
	e.Global.Set("*", BuiltinVal{Name: "*", Fn: e.wrapBinaryIntFloat(func(a, b float64) float64 { return a * b }, func(a, b int64) int64 { return a * b })})
	e.Global.Set("/", BuiltinVal{Name: "/", Fn: e.wrapBinaryIntFloat(func(a, b float64) float64 { return a / b }, func(a, b int64) int64 { return a / b })})
	e.Global.Set("%", BuiltinVal{Name: "%", Fn: func(args []Value) (Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("%% requires 2 arguments")
		}
		a, ok1 := args[0].(IntVal)
		b, ok2 := args[1].(IntVal)
		if !ok1 || !ok2 {
			return nil, fmt.Errorf("%% requires integers")
		}
		return IntVal{Value: a.Value % b.Value}, nil
	}})
	e.Global.Set("==", BuiltinVal{Name: "==", Fn: e.builtinEq})
	e.Global.Set("!=", BuiltinVal{Name: "!=", Fn: func(args []Value) (Value, error) {
		eq, err := e.builtinEq(args)
		if err != nil {
			return nil, err
		}
		return BoolVal{Value: !eq.(BoolVal).Value}, nil
	}})
	e.Global.Set("<", BuiltinVal{Name: "<", Fn: e.wrapBinaryComp(func(a, b float64) bool { return a < b }, func(a, b int64) bool { return a < b })})
	e.Global.Set(">", BuiltinVal{Name: ">", Fn: e.wrapBinaryComp(func(a, b float64) bool { return a > b }, func(a, b int64) bool { return a > b })})
	e.Global.Set("<=", BuiltinVal{Name: "<=", Fn: e.wrapBinaryComp(func(a, b float64) bool { return a <= b }, func(a, b int64) bool { return a <= b })})
	e.Global.Set(">=", BuiltinVal{Name: ">=", Fn: e.wrapBinaryComp(func(a, b float64) bool { return a >= b }, func(a, b int64) bool { return a >= b })})
	e.Global.Set("not", BuiltinVal{Name: "not", Fn: func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("not requires 1 argument")
		}
		return BoolVal{Value: !isTrue(args[0])}, nil
	}})
	e.Global.Set("println", BuiltinVal{Name: "println", Fn: e.builtinPrintln})
	e.Global.Set("len", BuiltinVal{Name: "len", Fn: func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("len requires 1 argument")
		}
		switch v := args[0].(type) {
		case StringVal:
			return IntVal{Value: int64(len(v.Value))}, nil
		case VectorVal:
			return IntVal{Value: int64(len(v.Elements))}, nil
		case MapVal:
			return IntVal{Value: int64(len(v.Pairs))}, nil
		default:
			return nil, fmt.Errorf("len not supported for %s", v.TypeName())
		}
	}})

	e.Global.Set("assert", BuiltinVal{Name: "assert", Fn: func(args []Value) (Value, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("assert requires at least 1 argument")
		}
		if !isTrue(args[0]) {
			msg := "assertion failed"
			if len(args) > 1 {
				msg = args[1].String()
			}
			return nil, fmt.Errorf("%s", msg)
		}
		return BoolVal{Value: true}, nil
	}})

	e.Global.Set("assert-eq", BuiltinVal{Name: "assert-eq", Fn: func(args []Value) (Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("assert-eq requires 2 arguments")
		}
		if !e.valuesEqual(args[0], args[1]) {
			return nil, fmt.Errorf("assertion failed: %v == %v", args[0], args[1])
		}
		return BoolVal{Value: true}, nil
	}})

	// Register fmt namespace
	fmtEnv := NewEnv(nil)
	fmtEnv.Set("Sprintf", BuiltinVal{Name: "fmt/Sprintf", Fn: func(args []Value) (Value, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("fmt/Sprintf requires at least 1 argument")
		}
		format, ok := args[0].(StringVal)
		if !ok {
			return nil, fmt.Errorf("fmt/Sprintf requires a format string")
		}
		var goArgs []any
		for _, arg := range args[1:] {
			// Basic conversion for now
			switch v := arg.(type) {
			case IntVal:
				goArgs = append(goArgs, v.Value)
			case FloatVal:
				goArgs = append(goArgs, v.Value)
			case StringVal:
				goArgs = append(goArgs, v.Value)
			case BoolVal:
				goArgs = append(goArgs, v.Value)
			default:
				goArgs = append(goArgs, v.String())
			}
		}
		return StringVal{Value: fmt.Sprintf(format.Value, goArgs...)}, nil
	}})
	fmtEnv.Set("Println", BuiltinVal{Name: "fmt/Println", Fn: e.builtinPrintln})
	e.Namespaces["fmt"] = fmtEnv

	// Pre-register some common namespaces
	mathEnv := NewEnv(nil)
	mathEnv.Set("Sqrt", BuiltinVal{Name: "math/Sqrt", Fn: func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("math/Sqrt requires 1 argument")
		}
		f, ok := toFloat(args[0])
		if !ok {
			return nil, fmt.Errorf("math/Sqrt requires a number")
		}
		return FloatVal{Value: math.Sqrt(f)}, nil
	}})
	mathEnv.Set("Pi", FloatVal{Value: math.Pi})
	e.Namespaces["math"] = mathEnv

	timeEnv := NewEnv(nil)
	timeEnv.Set("Now", BuiltinVal{Name: "time/Now", Fn: func(args []Value) (Value, error) {
		return StringVal{Value: "2026-02-01T12:00:00Z"}, nil // Mock time for now
	}})
	e.Namespaces["time"] = timeEnv
}

func toFloat(v Value) (float64, bool) {
	switch val := v.(type) {
	case IntVal:
		return float64(val.Value), true
	case FloatVal:
		return val.Value, true
	default:
		return 0, false
	}
}

// EvalTop evaluates a top-level node
func (e *Evaluator) EvalTop(node ast.Node) (Value, error) {
	switch n := node.(type) {
	case *ast.Def:
		val, err := e.Eval(n.Value, e.Global)
		if err != nil {
			return nil, err
		}
		e.Global.Set(n.Name, val)
		return val, nil

	case *ast.Defn:
		fn := FuncVal{
			Name:   n.Name,
			Params: n.Params,
			Body:   n.Body,
			Env:    e.Global,
		}
		e.Global.Set(n.Name, fn)
		return fn, nil

	case *ast.Package:
		// Just ignore for evaluation
		return NilVal{}, nil

	case *ast.Import:
		for _, spec := range n.Specs {
			name := spec.Path
			if spec.Alias != "" {
				name = spec.Alias
			}
			// If it's a known namespace, we're good
			if _, ok := e.Namespaces[spec.Path]; ok {
				if spec.Alias != "" {
					e.Namespaces[name] = e.Namespaces[spec.Path]
				}
				continue
			}
			// Otherwise, create an empty one to avoid errors later,
			// though it won't have any symbols.
			e.Namespaces[name] = NewEnv(nil)
		}
		return NilVal{}, nil

	case ast.Expr:
		return e.Eval(n, e.Global)

	default:
		return nil, fmt.Errorf("unsupported top-level node type: %T", node)
	}
}

// Eval evaluates an expression in an environment
func (e *Evaluator) Eval(expr ast.Expr, env *Env) (Value, error) {
	switch node := expr.(type) {
	case *ast.IntLit:
		return IntVal{Value: node.Value}, nil

	case *ast.FloatLit:
		return FloatVal{Value: node.Value}, nil

	case *ast.StringLit:
		return StringVal{Value: node.Value}, nil

	case *ast.BoolLit:
		return BoolVal{Value: node.Value}, nil

	case *ast.NilLit:
		return NilVal{}, nil

	case *ast.KeywordLit:
		return StringVal{Value: ":" + node.Value}, nil

	case *ast.Symbol:
		if strings.Contains(node.Name, "/") && node.Name != "/" {
			parts := strings.SplitN(node.Name, "/", 2)
			if parts[0] != "" {
				if ns, ok := e.Namespaces[parts[0]]; ok {
					if val, ok := ns.Get(parts[1]); ok {
						return val, nil
					}
					return nil, fmt.Errorf("symbol not found in namespace %s: %s", parts[0], parts[1])
				}
				return nil, fmt.Errorf("unknown namespace: %s", parts[0])
			}
		}
		if val, ok := env.Get(node.Name); ok {
			return val, nil
		}
		return nil, fmt.Errorf("undefined symbol: %s", node.Name)

	case *ast.VectorLit:
		var elements []Value
		for _, elt := range node.Elts {
			val, err := e.Eval(elt, env)
			if err != nil {
				return nil, err
			}
			elements = append(elements, val)
		}
		return VectorVal{Elements: elements}, nil

	case *ast.MapLit:
		pairs := make(map[string]Value)
		for _, pair := range node.Pairs {
			keyVal, err := e.Eval(pair.Key, env)
			if err != nil {
				return nil, err
			}
			val, err := e.Eval(pair.Value, env)
			if err != nil {
				return nil, err
			}
			pairs[keyVal.String()] = val
		}
		return MapVal{Pairs: pairs}, nil

	case *ast.FuncCall:
		return e.evalFuncCall(node, env)

	case *ast.IfExpr:
		cond, err := e.Eval(node.Cond, env)
		if err != nil {
			return nil, err
		}
		if isTrue(cond) {
			return e.Eval(node.Then, env)
		} else if node.Else != nil {
			return e.Eval(node.Else, env)
		}
		return NilVal{}, nil

	case *ast.DoExpr:
		var last Value = NilVal{}
		for _, ex := range node.Exprs {
			val, err := e.Eval(ex, env)
			if err != nil {
				return nil, err
			}
			last = val
		}
		return last, nil

	case *ast.LetExpr:
		childEnv := NewEnv(env)
		for _, b := range node.Bindings {
			val, err := e.Eval(b.Init, env)
			if err != nil {
				return nil, err
			}
			if len(b.Names) == 1 {
				childEnv.Set(b.Names[0], val)
			} else {
				// Basic support for vector destructuring in eval
				if vec, ok := val.(VectorVal); ok {
					for i, name := range b.Names {
						if i < len(vec.Elements) {
							childEnv.Set(name, vec.Elements[i])
						}
					}
				}
			}
		}
		var last Value = NilVal{}
		for _, ex := range node.Body {
			val, err := e.Eval(ex, childEnv)
			if err != nil {
				return nil, err
			}
			last = val
		}
		return last, nil

	case *ast.QuoteExpr:
		return StringVal{Value: node.Expr.String()}, nil

	case *ast.MatchExpr:
		val, err := e.Eval(node.Val, env)
		if err != nil {
			return nil, err
		}
		for _, c := range node.Cases {
			// Basic symbol/keyword match or else/_
			match := false
			if sym, ok := c.Pattern.(*ast.Symbol); ok && (sym.Name == "else" || sym.Name == "_") {
				match = true
			} else if kw, ok := c.Pattern.(*ast.KeywordLit); ok && kw.Value == "else" {
				match = true
			} else {
				patVal, err := e.Eval(c.Pattern, env)
				if err == nil && e.valuesEqual(val, patVal) {
					match = true
				}
			}

			if match {
				return e.Eval(c.Body, env)
			}
		}
		return nil, fmt.Errorf("no match found for value: %v", val)

	default:
		return nil, fmt.Errorf("unsupported expression type: %T", expr)
	}
}

func (e *Evaluator) evalFuncCall(call *ast.FuncCall, env *Env) (Value, error) {
	fnVal, err := e.Eval(call.Func, env)
	if err != nil {
		return nil, err
	}

	var args []Value
	for _, argExpr := range call.Args {
		val, err := e.Eval(argExpr, env)
		if err != nil {
			return nil, err
		}
		args = append(args, val)
	}

	switch fn := fnVal.(type) {
	case BuiltinVal:
		return fn.Fn(args)
	case FuncVal:
		return e.callUserFunc(fn, args)
	default:
		return nil, fmt.Errorf("not a function: %s", fnVal.String())
	}
}

func (e *Evaluator) callUserFunc(fn FuncVal, args []Value) (Value, error) {
	if len(args) < len(fn.Params) {
		return nil, fmt.Errorf("too few arguments for %s", fn.Name)
	}
	// TODO: variadic params

	callEnv := NewEnv(fn.Env)
	for i, p := range fn.Params {
		callEnv.Set(p.Name, args[i])
	}

	var last Value = NilVal{}
	for _, expr := range fn.Body {
		var err error
		last, err = e.Eval(expr, callEnv)
		if err != nil {
			return nil, err
		}
	}
	return last, nil
}

func isTrue(v Value) bool {
	switch val := v.(type) {
	case NilVal:
		return false
	case BoolVal:
		return val.Value
	default:
		return true
	}
}

// Builtin helpers
func (e *Evaluator) wrapBinaryIntFloat(fOp func(float64, float64) float64, iOp func(int64, int64) int64) func([]Value) (Value, error) {
	return func(args []Value) (Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("expected at least 2 arguments")
		}
		res := args[0]
		for i := 1; i < len(args); i++ {
			a := res
			b := args[i]
			if ai, ok := a.(IntVal); ok {
				if bi, ok := b.(IntVal); ok {
					res = IntVal{Value: iOp(ai.Value, bi.Value)}
				} else if bf, ok := b.(FloatVal); ok {
					res = FloatVal{Value: fOp(float64(ai.Value), bf.Value)}
				} else {
					return nil, fmt.Errorf("invalid operand types")
				}
			} else if af, ok := a.(FloatVal); ok {
				if bi, ok := b.(IntVal); ok {
					res = FloatVal{Value: fOp(af.Value, float64(bi.Value))}
				} else if bf, ok := b.(FloatVal); ok {
					res = FloatVal{Value: fOp(af.Value, bf.Value)}
				} else {
					return nil, fmt.Errorf("invalid operand types")
				}
			} else {
				return nil, fmt.Errorf("invalid operand types")
			}
		}
		return res, nil
	}
}

func (e *Evaluator) wrapBinaryComp(fOp func(float64, float64) bool, iOp func(int64, int64) bool) func([]Value) (Value, error) {
	return func(args []Value) (Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("expected 2 arguments")
		}
		a := args[0]
		b := args[1]
		if ai, ok := a.(IntVal); ok {
			if bi, ok := b.(IntVal); ok {
				return BoolVal{Value: iOp(ai.Value, bi.Value)}, nil
			} else if bf, ok := b.(FloatVal); ok {
				return BoolVal{Value: fOp(float64(ai.Value), bf.Value)}, nil
			}
		} else if af, ok := a.(FloatVal); ok {
			if bi, ok := b.(IntVal); ok {
				return BoolVal{Value: fOp(af.Value, float64(bi.Value))}, nil
			} else if bf, ok := b.(FloatVal); ok {
				return BoolVal{Value: fOp(af.Value, bf.Value)}, nil
			}
		}
		return nil, fmt.Errorf("invalid operand types for comparison")
	}
}

func (e *Evaluator) builtinEq(args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("== requires 2 arguments")
	}
	return BoolVal{Value: e.valuesEqual(args[0], args[1])}, nil
}

func (e *Evaluator) valuesEqual(left, right Value) bool {
	if left.TypeName() != right.TypeName() {
		return false
	}
	switch l := left.(type) {
	case IntVal:
		return l.Value == right.(IntVal).Value
	case FloatVal:
		return math.Abs(l.Value-right.(FloatVal).Value) < 1e-10
	case StringVal:
		return l.Value == right.(StringVal).Value
	case BoolVal:
		return l.Value == right.(BoolVal).Value
	case NilVal:
		return true
	case VectorVal:
		rv := right.(VectorVal)
		if len(l.Elements) != len(rv.Elements) {
			return false
		}
		for i := range l.Elements {
			if !e.valuesEqual(l.Elements[i], rv.Elements[i]) {
				return false
			}
		}
		return true
	}
	return false
}

func (e *Evaluator) builtinPrintln(args []Value) (Value, error) {
	for i, arg := range args {
		if i > 0 {
			fmt.Print(" ")
		}
		fmt.Print(arg.String())
	}
	fmt.Println()
	return NilVal{}, nil
}
