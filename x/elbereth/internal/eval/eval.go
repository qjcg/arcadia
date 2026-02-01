package eval

import (
	"fmt"
	"math"

	"elbereth/internal/ast"
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
	result := "["
	for i, el := range v.Elements {
		if i > 0 {
			result += " "
		}
		result += el.String()
	}
	result += "]"
	return result
}

// MapVal represents a map
type MapVal struct {
	Pairs map[string]Value
}

func (v MapVal) TypeName() string { return "map" }
func (v MapVal) String() string {
	result := "{"
	first := true
	for k, val := range v.Pairs {
		if !first {
			result += " "
		}
		result += fmt.Sprintf("%s %s", k, val.String())
		first = false
	}
	result += "}"
	return result
}

// NodeVal represents an AST node as a value
type NodeVal struct {
	Node ast.Node
}

func (v NodeVal) TypeName() string { return "node" }
func (v NodeVal) String() string   { return v.Node.String() }

// Evaluator evaluates Elbereth expressions
type Evaluator struct {
	globals map[string]Value
}

// New creates a new evaluator
func New() *Evaluator {
	return &Evaluator{
		globals: make(map[string]Value),
	}
}

// Eval evaluates an expression
func (e *Evaluator) Eval(expr ast.Expr) (Value, error) {
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
		if val, ok := e.globals[node.Name]; ok {
			return val, nil
		}
		return nil, fmt.Errorf("undefined symbol: %s", node.Name)

	case *ast.VectorLit:
		var elements []Value
		for _, elt := range node.Elts {
			val, err := e.Eval(elt)
			if err != nil {
				return nil, err
			}
			elements = append(elements, val)
		}
		return VectorVal{Elements: elements}, nil

	case *ast.MapLit:
		pairs := make(map[string]Value)
		for _, pair := range node.Pairs {
			keyVal, err := e.Eval(pair.Key)
			if err != nil {
				return nil, err
			}
			val, err := e.Eval(pair.Value)
			if err != nil {
				return nil, err
			}
			pairs[keyVal.String()] = val
		}
		return MapVal{Pairs: pairs}, nil

	case *ast.FuncCall:
		return e.evalFuncCall(node)

	case *ast.QuoteExpr:
		// For now, quote just returns a string representation
		return StringVal{Value: node.Expr.String()}, nil

	default:
		return nil, fmt.Errorf("unsupported expression type: %T", expr)
	}
}

func (e *Evaluator) evalFuncCall(call *ast.FuncCall) (Value, error) {
	// Handle built-in functions
	if sym, ok := call.Func.(*ast.Symbol); ok {
		switch sym.Name {
		case "+":
			return e.builtinAdd(call)
		case "-":
			return e.builtinSub(call)
		case "*":
			return e.builtinMul(call)
		case "/":
			return e.builtinDiv(call)
		case "%":
			return e.builtinMod(call)
		case "==":
			return e.builtinEq(call)
		case "!=":
			return e.builtinNe(call)
		case "<":
			return e.builtinLt(call)
		case "<=":
			return e.builtinLte(call)
		case ">":
			return e.builtinGt(call)
		case ">=":
			return e.builtinGte(call)
		case "println":
			return e.builtinPrintln(call)
		case "print":
			return e.builtinPrint(call)
		case "len":
			return e.builtinLen(call)
		}
	}

	return nil, fmt.Errorf("unknown function")
}

func (e *Evaluator) builtinAdd(call *ast.FuncCall) (Value, error) {
	if len(call.Args) == 0 {
		return IntVal{Value: 0}, nil
	}

	result, err := e.Eval(call.Args[0])
	if err != nil {
		return nil, err
	}

	for i := 1; i < len(call.Args); i++ {
		arg, err := e.Eval(call.Args[i])
		if err != nil {
			return nil, err
		}

		switch r := result.(type) {
		case IntVal:
			switch a := arg.(type) {
			case IntVal:
				result = IntVal{Value: r.Value + a.Value}
			case FloatVal:
				result = FloatVal{Value: float64(r.Value) + a.Value}
			default:
				return nil, fmt.Errorf("cannot add %s and %s", r.TypeName(), a.TypeName())
			}
		case FloatVal:
			switch a := arg.(type) {
			case IntVal:
				result = FloatVal{Value: r.Value + float64(a.Value)}
			case FloatVal:
				result = FloatVal{Value: r.Value + a.Value}
			default:
				return nil, fmt.Errorf("cannot add %s and %s", r.TypeName(), a.TypeName())
			}
		case StringVal:
			switch a := arg.(type) {
			case StringVal:
				result = StringVal{Value: r.Value + a.Value}
			default:
				return nil, fmt.Errorf("cannot add %s and %s", r.TypeName(), a.TypeName())
			}
		default:
			return nil, fmt.Errorf("cannot add %s", r.TypeName())
		}
	}

	return result, nil
}

func (e *Evaluator) builtinSub(call *ast.FuncCall) (Value, error) {
	if len(call.Args) < 2 {
		return nil, fmt.Errorf("- requires at least 2 arguments")
	}

	result, err := e.Eval(call.Args[0])
	if err != nil {
		return nil, err
	}

	for i := 1; i < len(call.Args); i++ {
		arg, err := e.Eval(call.Args[i])
		if err != nil {
			return nil, err
		}

		switch r := result.(type) {
		case IntVal:
			switch a := arg.(type) {
			case IntVal:
				result = IntVal{Value: r.Value - a.Value}
			case FloatVal:
				result = FloatVal{Value: float64(r.Value) - a.Value}
			default:
				return nil, fmt.Errorf("cannot subtract %s from %s", a.TypeName(), r.TypeName())
			}
		case FloatVal:
			switch a := arg.(type) {
			case IntVal:
				result = FloatVal{Value: r.Value - float64(a.Value)}
			case FloatVal:
				result = FloatVal{Value: r.Value - a.Value}
			default:
				return nil, fmt.Errorf("cannot subtract %s from %s", a.TypeName(), r.TypeName())
			}
		default:
			return nil, fmt.Errorf("cannot subtract from %s", r.TypeName())
		}
	}

	return result, nil
}

func (e *Evaluator) builtinMul(call *ast.FuncCall) (Value, error) {
	if len(call.Args) == 0 {
		return IntVal{Value: 1}, nil
	}

	result, err := e.Eval(call.Args[0])
	if err != nil {
		return nil, err
	}

	for i := 1; i < len(call.Args); i++ {
		arg, err := e.Eval(call.Args[i])
		if err != nil {
			return nil, err
		}

		switch r := result.(type) {
		case IntVal:
			switch a := arg.(type) {
			case IntVal:
				result = IntVal{Value: r.Value * a.Value}
			case FloatVal:
				result = FloatVal{Value: float64(r.Value) * a.Value}
			default:
				return nil, fmt.Errorf("cannot multiply %s and %s", r.TypeName(), a.TypeName())
			}
		case FloatVal:
			switch a := arg.(type) {
			case IntVal:
				result = FloatVal{Value: r.Value * float64(a.Value)}
			case FloatVal:
				result = FloatVal{Value: r.Value * a.Value}
			default:
				return nil, fmt.Errorf("cannot multiply %s and %s", r.TypeName(), a.TypeName())
			}
		default:
			return nil, fmt.Errorf("cannot multiply %s", r.TypeName())
		}
	}

	return result, nil
}

func (e *Evaluator) builtinDiv(call *ast.FuncCall) (Value, error) {
	if len(call.Args) < 2 {
		return nil, fmt.Errorf("/ requires at least 2 arguments")
	}

	result, err := e.Eval(call.Args[0])
	if err != nil {
		return nil, err
	}

	for i := 1; i < len(call.Args); i++ {
		arg, err := e.Eval(call.Args[i])
		if err != nil {
			return nil, err
		}

		switch r := result.(type) {
		case IntVal:
			switch a := arg.(type) {
			case IntVal:
				if a.Value == 0 {
					return nil, fmt.Errorf("division by zero")
				}
				result = IntVal{Value: r.Value / a.Value}
			case FloatVal:
				if a.Value == 0 {
					return nil, fmt.Errorf("division by zero")
				}
				result = FloatVal{Value: float64(r.Value) / a.Value}
			default:
				return nil, fmt.Errorf("cannot divide %s and %s", r.TypeName(), a.TypeName())
			}
		case FloatVal:
			switch a := arg.(type) {
			case IntVal:
				if a.Value == 0 {
					return nil, fmt.Errorf("division by zero")
				}
				result = FloatVal{Value: r.Value / float64(a.Value)}
			case FloatVal:
				if a.Value == 0 {
					return nil, fmt.Errorf("division by zero")
				}
				result = FloatVal{Value: r.Value / a.Value}
			default:
				return nil, fmt.Errorf("cannot divide %s and %s", r.TypeName(), a.TypeName())
			}
		default:
			return nil, fmt.Errorf("cannot divide %s", r.TypeName())
		}
	}

	return result, nil
}

func (e *Evaluator) builtinMod(call *ast.FuncCall) (Value, error) {
	if len(call.Args) != 2 {
		return nil, fmt.Errorf("%% requires exactly 2 arguments")
	}

	left, err := e.Eval(call.Args[0])
	if err != nil {
		return nil, err
	}

	right, err := e.Eval(call.Args[1])
	if err != nil {
		return nil, err
	}

	l, ok := left.(IntVal)
	if !ok {
		return nil, fmt.Errorf("left operand to %% must be int, got %s", left.TypeName())
	}

	r, ok := right.(IntVal)
	if !ok {
		return nil, fmt.Errorf("right operand to %% must be int, got %s", right.TypeName())
	}

	if r.Value == 0 {
		return nil, fmt.Errorf("division by zero")
	}

	return IntVal{Value: l.Value % r.Value}, nil
}

func (e *Evaluator) builtinEq(call *ast.FuncCall) (Value, error) {
	if len(call.Args) != 2 {
		return nil, fmt.Errorf("== requires exactly 2 arguments")
	}

	left, err := e.Eval(call.Args[0])
	if err != nil {
		return nil, err
	}

	right, err := e.Eval(call.Args[1])
	if err != nil {
		return nil, err
	}

	result := e.valuesEqual(left, right)
	return BoolVal{Value: result}, nil
}

func (e *Evaluator) builtinNe(call *ast.FuncCall) (Value, error) {
	if len(call.Args) != 2 {
		return nil, fmt.Errorf("!= requires exactly 2 arguments")
	}

	left, err := e.Eval(call.Args[0])
	if err != nil {
		return nil, err
	}

	right, err := e.Eval(call.Args[1])
	if err != nil {
		return nil, err
	}

	result := !e.valuesEqual(left, right)
	return BoolVal{Value: result}, nil
}

func (e *Evaluator) builtinLt(call *ast.FuncCall) (Value, error) {
	if len(call.Args) != 2 {
		return nil, fmt.Errorf("< requires exactly 2 arguments")
	}

	left, err := e.Eval(call.Args[0])
	if err != nil {
		return nil, err
	}

	right, err := e.Eval(call.Args[1])
	if err != nil {
		return nil, err
	}

	cmp, err := e.compare(left, right)
	if err != nil {
		return nil, err
	}

	return BoolVal{Value: cmp < 0}, nil
}

func (e *Evaluator) builtinLte(call *ast.FuncCall) (Value, error) {
	if len(call.Args) != 2 {
		return nil, fmt.Errorf("<= requires exactly 2 arguments")
	}

	left, err := e.Eval(call.Args[0])
	if err != nil {
		return nil, err
	}

	right, err := e.Eval(call.Args[1])
	if err != nil {
		return nil, err
	}

	cmp, err := e.compare(left, right)
	if err != nil {
		return nil, err
	}

	return BoolVal{Value: cmp <= 0}, nil
}

func (e *Evaluator) builtinGt(call *ast.FuncCall) (Value, error) {
	if len(call.Args) != 2 {
		return nil, fmt.Errorf("> requires exactly 2 arguments")
	}

	left, err := e.Eval(call.Args[0])
	if err != nil {
		return nil, err
	}

	right, err := e.Eval(call.Args[1])
	if err != nil {
		return nil, err
	}

	cmp, err := e.compare(left, right)
	if err != nil {
		return nil, err
	}

	return BoolVal{Value: cmp > 0}, nil
}

func (e *Evaluator) builtinGte(call *ast.FuncCall) (Value, error) {
	if len(call.Args) != 2 {
		return nil, fmt.Errorf(">= requires exactly 2 arguments")
	}

	left, err := e.Eval(call.Args[0])
	if err != nil {
		return nil, err
	}

	right, err := e.Eval(call.Args[1])
	if err != nil {
		return nil, err
	}

	cmp, err := e.compare(left, right)
	if err != nil {
		return nil, err
	}

	return BoolVal{Value: cmp >= 0}, nil
}

func (e *Evaluator) valuesEqual(left, right Value) bool {
	switch l := left.(type) {
	case IntVal:
		r, ok := right.(IntVal)
		return ok && l.Value == r.Value
	case FloatVal:
		r, ok := right.(FloatVal)
		return ok && math.Abs(l.Value-r.Value) < 1e-10
	case StringVal:
		r, ok := right.(StringVal)
		return ok && l.Value == r.Value
	case BoolVal:
		r, ok := right.(BoolVal)
		return ok && l.Value == r.Value
	case NilVal:
		_, ok := right.(NilVal)
		return ok
	}
	return false
}

func (e *Evaluator) compare(left, right Value) (int, error) {
	switch l := left.(type) {
	case IntVal:
		switch r := right.(type) {
		case IntVal:
			if l.Value < r.Value {
				return -1, nil
			} else if l.Value > r.Value {
				return 1, nil
			}
			return 0, nil
		case FloatVal:
			lf := float64(l.Value)
			if lf < r.Value {
				return -1, nil
			} else if lf > r.Value {
				return 1, nil
			}
			return 0, nil
		}
	case FloatVal:
		switch r := right.(type) {
		case IntVal:
			rf := float64(r.Value)
			if l.Value < rf {
				return -1, nil
			} else if l.Value > rf {
				return 1, nil
			}
			return 0, nil
		case FloatVal:
			if l.Value < r.Value {
				return -1, nil
			} else if l.Value > r.Value {
				return 1, nil
			}
			return 0, nil
		}
	case StringVal:
		r, ok := right.(StringVal)
		if !ok {
			return 0, fmt.Errorf("cannot compare string and %s", right.TypeName())
		}
		if l.Value < r.Value {
			return -1, nil
		} else if l.Value > r.Value {
			return 1, nil
		}
		return 0, nil
	}
	return 0, fmt.Errorf("cannot compare %s and %s", left.TypeName(), right.TypeName())
}

func (e *Evaluator) builtinPrintln(call *ast.FuncCall) (Value, error) {
	for i, arg := range call.Args {
		val, err := e.Eval(arg)
		if err != nil {
			return nil, err
		}
		if i > 0 {
			fmt.Print(" ")
		}
		fmt.Print(val.String())
	}
	fmt.Println()
	return NilVal{}, nil
}

func (e *Evaluator) builtinPrint(call *ast.FuncCall) (Value, error) {
	for i, arg := range call.Args {
		val, err := e.Eval(arg)
		if err != nil {
			return nil, err
		}
		if i > 0 {
			fmt.Print(" ")
		}
		fmt.Print(val.String())
	}
	return NilVal{}, nil
}

func (e *Evaluator) builtinLen(call *ast.FuncCall) (Value, error) {
	if len(call.Args) != 1 {
		return nil, fmt.Errorf("len requires exactly 1 argument")
	}

	val, err := e.Eval(call.Args[0])
	if err != nil {
		return nil, err
	}

	switch v := val.(type) {
	case StringVal:
		return IntVal{Value: int64(len(v.Value))}, nil
	case VectorVal:
		return IntVal{Value: int64(len(v.Elements))}, nil
	case MapVal:
		return IntVal{Value: int64(len(v.Pairs))}, nil
	default:
		return nil, fmt.Errorf("len requires string, vector, or map, got %s", v.TypeName())
	}
}
