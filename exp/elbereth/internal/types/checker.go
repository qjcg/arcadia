package types

import (
	"slices"

	"github.com/qjcg/arcadia/exp/elbereth/internal/ast"
)

// Checker performs type inference and checking on an Elbereth AST
type Checker struct {
	types     map[ast.Expr]Type
	symbols   map[string]Type
	scopes    []map[string]Type
	functions map[string]*ast.Defn
	structs   map[string]*ast.Deftype
}

// NewChecker creates a new type checker
func NewChecker() *Checker {
	return &Checker{
		types:     make(map[ast.Expr]Type),
		symbols:   make(map[string]Type),
		functions: make(map[string]*ast.Defn),
		structs:   make(map[string]*ast.Deftype),
	}
}

// Check performs type checking on a program
func (c *Checker) Check(prog *ast.Program) (map[ast.Expr]Type, error) {
	// First pass: collect global definitions
	for _, item := range prog.Items {
		switch n := item.(type) {
		case *ast.Def:
			if n.Type != nil {
				if nt, ok := n.Type.(*ast.NamedType); ok {
					c.symbols[n.Name] = ParseTypeString(nt.Name)
				}
			}
		case *ast.Defn:
			c.functions[n.Name] = n
		case *ast.Deftype:
			c.structs[n.Name] = n
		}
	}

	// Second pass: infer types for everything
	for _, item := range prog.Items {
		var err error
		switch n := item.(type) {
		case ast.Expr:
			_, err = c.infer(n)
		case *ast.Defn:
			err = c.checkDefn(n)
		case *ast.Def:
			_, err = c.infer(n.Value)
		case *ast.Deftest:
			err = c.checkBody(n.Body)
		case *ast.Defbenchmark:
			c.pushScope()
			c.setInScope(n.BParam, Nil)
			err = c.checkBody(n.Body)
			c.popScope()
		case *ast.Defexample:
			err = c.checkBody(n.Body)
		}
		if err != nil {
			return nil, err
		}
	}

	return c.types, nil
}

func (c *Checker) checkBody(body []ast.Expr) error {
	for _, expr := range body {
		_, err := c.infer(expr)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Checker) checkDefn(d *ast.Defn) error {
	c.pushScope()
	defer c.popScope()

	for _, p := range d.Params {
		if p.Type != nil {
			c.setInScope(p.Name, c.astTypeToType(p.Type))
		} else {
			c.setInScope(p.Name, Nil) // fallback to nil/any if unknown
		}
	}

	for _, expr := range d.Body {
		_, err := c.infer(expr)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Checker) infer(expr ast.Expr) (Type, error) {
	if expr == nil {
		return Nil, nil
	}
	if t, ok := c.types[expr]; ok {
		return t, nil
	}

	var res Type
	var err error

	switch n := expr.(type) {
	case *ast.IntLit:
		res = Int
	case *ast.FloatLit:
		res = Float
	case *ast.StringLit:
		res = String
	case *ast.BoolLit:
		res = Bool
	case *ast.NilLit:
		res = Nil
	case *ast.Symbol:
		if t, ok := c.lookup(n.Name); ok {
			res = t
		} else {
			res = Nil // unknown
		}
	case *ast.FuncCall:
		res, err = c.inferFuncCall(n)
	case *ast.IfExpr:
		c.infer(n.Cond)
		t1, err1 := c.infer(n.Then)
		if err1 != nil {
			return nil, err1
		}
		if n.Else != nil {
			c.infer(n.Else)
		}
		res = t1
	case *ast.DoExpr:
		for _, ex := range n.Exprs {
			res, err = c.infer(ex)
		}
		if len(n.Exprs) == 0 {
			res = Nil
		}
	case *ast.VectorLit:
		for _, el := range n.Elts {
			c.infer(el)
		}
		res = Nil
	case *ast.MapLit:
		for _, p := range n.Pairs {
			c.infer(p.Key)
			c.infer(p.Value)
		}
		res = Nil
	case *ast.LetExpr:
		c.pushScope()
		for _, b := range n.Bindings {
			bt, _ := c.infer(b.Init)
			for _, name := range b.Names {
				c.setInScope(name, bt)
			}
		}
		for _, ex := range n.Body {
			res, err = c.infer(ex)
		}
		if len(n.Body) == 0 {
			res = Nil
		}
		c.popScope()
	case *ast.MatchExpr:
		c.infer(n.Val)
		for _, cs := range n.Cases {
			c.infer(cs.Pattern)
			res, err = c.infer(cs.Body)
		}
	case *ast.LoopExpr:
		c.pushScope()
		for _, b := range n.Bindings {
			bt, _ := c.infer(b.Init)
			for _, name := range b.Names {
				c.setInScope(name, bt)
			}
		}
		for _, ex := range n.Body {
			res, err = c.infer(ex)
		}
		c.popScope()
	case *ast.DoseqExpr:
		c.pushScope()
		c.infer(n.Coll)
		c.setInScope(n.Var, Nil) // Todo: infer from collection
		for _, ex := range n.Body {
			res, err = c.infer(ex)
		}
		c.popScope()
	case *ast.FuncLit:
		c.pushScope()
		for _, p := range n.Params {
			if p.Type != nil {
				c.setInScope(p.Name, c.astTypeToType(p.Type))
			} else {
				c.setInScope(p.Name, Nil)
			}
		}
		for _, ex := range n.Body {
			res, err = c.infer(ex)
		}
		c.popScope()
	case *ast.RecurExpr:
		for _, arg := range n.Args {
			c.infer(arg)
		}
		res = Nil
	case *ast.SelectExpr:
		for _, cs := range n.Cases {
			if cs.Chan != nil {
				c.infer(cs.Chan)
			}
			c.pushScope()
			if cs.Binding != "" {
				c.setInScope(cs.Binding, Nil)
			}
			for _, ex := range cs.Body {
				res, err = c.infer(ex)
			}
			c.popScope()
		}
	default:
		res = Nil
	}

	if err != nil {
		return nil, err
	}
	c.types[expr] = res
	return res, nil
}

func (c *Checker) inferFuncCall(call *ast.FuncCall) (Type, error) {
	if sym, ok := call.Func.(*ast.Symbol); ok {
		switch sym.Name {
		case "for":
			if len(call.Args) >= 3 {
				c.pushScope()
				if initV, ok := call.Args[0].(*ast.VectorLit); ok {
					for i := 0; i+1 < len(initV.Elts); i += 2 {
						if nameSym, ok := initV.Elts[i].(*ast.Symbol); ok {
							it, _ := c.infer(initV.Elts[i+1])
							c.setInScope(nameSym.Name, it)
						}
					}
				}
				for i := 1; i < len(call.Args); i++ {
					c.infer(call.Args[i])
				}
				c.popScope()
				return Nil, nil
			}
		}
	}

	for _, arg := range call.Args {
		c.infer(arg)
	}

	if sym, ok := call.Func.(*ast.Symbol); ok {
		switch sym.Name {
		case "+", "-", "*", "/", "%":
			if len(call.Args) > 0 {
				return c.infer(call.Args[0])
			}
		case "==", "!=", "<", "<=", ">", ">=":
			return Bool, nil
		case "println", "print":
			return Nil, nil
		case "len":
			return Int, nil
		case "chan":
			if len(call.Args) > 0 {
				var at ast.Type
				if st, ok := call.Args[0].(*ast.Symbol); ok {
					at = &ast.NamedType{Name: st.Name}
				}
				return &ChanType{EltType: c.astTypeToType(at)}, nil
			}
		}
		// Check user functions
		if fn, ok := c.functions[sym.Name]; ok {
			if fn.ReturnType != nil {
				return c.astTypeToType(fn.ReturnType), nil
			}
		}
	}
	return Nil, nil
}

func (c *Checker) astTypeToType(at ast.Type) Type {
	if at == nil {
		return Nil
	}
	switch t := at.(type) {
	case *ast.NamedType:
		return ParseTypeString(t.Name)
	case *ast.SliceType:
		return &SliceType{EltType: c.astTypeToType(t.EltType)}
	case *ast.ChanType:
		return &ChanType{EltType: c.astTypeToType(t.EltType), Buffer: t.Buffer}
	default:
		return Nil
	}
}

func (c *Checker) pushScope() {
	c.scopes = append(c.scopes, make(map[string]Type))
}

func (c *Checker) popScope() {
	c.scopes = c.scopes[:len(c.scopes)-1]
}

func (c *Checker) setInScope(name string, t Type) {
	if len(c.scopes) > 0 {
		c.scopes[len(c.scopes)-1][name] = t
	}
}

func (c *Checker) lookup(name string) (Type, bool) {
	for _, v := range slices.Backward(c.scopes) {
		if t, ok := v[name]; ok {
			return t, true
		}
	}
	if t, ok := c.symbols[name]; ok {
		return t, true
	}
	return nil, false
}
