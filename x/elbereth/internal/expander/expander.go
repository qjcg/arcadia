package expander

import (
	"fmt"

	"elbereth/internal/ast"
)

// Expander expands macros in an Elbereth AST
type Expander struct {
	macros map[string]*ast.Defmacro
}

// New creates a new expander
func New() *Expander {
	return &Expander{
		macros: make(map[string]*ast.Defmacro),
	}
}

// Expand expands all macros in the program
func (e *Expander) Expand(prog *ast.Program) error {
	// First pass: collect macro definitions
	for _, item := range prog.Items {
		if m, ok := item.(*ast.Defmacro); ok {
			e.macros[m.Name] = m
		}
	}

	// Second pass: expand macro calls
	var newItems []ast.Node
	for _, item := range prog.Items {
		if _, ok := item.(*ast.Defmacro); ok {
			continue
		}

		expanded, err := e.expandNode(item, 0)
		if err != nil {
			return err
		}
		if expanded != nil {
			newItems = append(newItems, expanded)
		}
	}

	prog.Items = newItems
	return nil
}

func (e *Expander) expandNode(node ast.Node, depth int) (ast.Node, error) {
	if depth > 1000 {
		return nil, fmt.Errorf("maximum expansion depth exceeded at node %T", node)
	}
	if node == nil {
		return nil, nil
	}

	switch n := node.(type) {
	case *ast.Defn:
		var newBody []ast.Expr
		for _, expr := range n.Body {
			expanded, err := e.expandExpr(expr, depth+1)
			if err != nil {
				return nil, err
			}
			newBody = append(newBody, expanded)
		}
		n.Body = newBody
		return n, nil

	case *ast.Deftest:
		var newBody []ast.Expr
		for _, expr := range n.Body {
			expanded, err := e.expandExpr(expr, depth+1)
			if err != nil {
				return nil, err
			}
			newBody = append(newBody, expanded)
		}
		n.Body = newBody
		return n, nil

	case *ast.Defbenchmark:
		var newBody []ast.Expr
		for _, expr := range n.Body {
			expanded, err := e.expandExpr(expr, depth+1)
			if err != nil {
				return nil, err
			}
			newBody = append(newBody, expanded)
		}
		n.Body = newBody
		return n, nil

	case *ast.Defexample:
		var newBody []ast.Expr
		for _, expr := range n.Body {
			expanded, err := e.expandExpr(expr, depth+1)
			if err != nil {
				return nil, err
			}
			newBody = append(newBody, expanded)
		}
		n.Body = newBody
		return n, nil

	case *ast.Def:
		expanded, err := e.expandExpr(n.Value, depth+1)
		if err != nil {
			return nil, err
		}
		n.Value = expanded
		return n, nil

	case ast.Expr:
		return e.expandExpr(n, depth+1)

	default:
		return n, nil
	}
}

func (e *Expander) expandExpr(expr ast.Expr, depth int) (ast.Expr, error) {
	if depth > 1000 {
		return nil, fmt.Errorf("maximum macro expansion depth exceeded at expression %T", expr)
	}
	if expr == nil {
		return nil, nil
	}

	switch n := expr.(type) {
	case *ast.FuncCall:
		// Check if it's a macro call
		if sym, ok := n.Func.(*ast.Symbol); ok {
			if macro, ok := e.macros[sym.Name]; ok {
				expanded, err := e.expandMacroCall(macro, n, depth+1)
				if err != nil {
					return nil, err
				}
				// Recursively expand the result of the macro expansion
				return e.expandExpr(expanded, depth+1)
			}
		}

		// Recursively expand function and arguments
		newFunc, err := e.expandExpr(n.Func, depth+1)
		if err != nil {
			return nil, err
		}
		var newArgs []ast.Expr
		for _, arg := range n.Args {
			expanded, err := e.expandExpr(arg, depth+1)
			if err != nil {
				return nil, err
			}
			newArgs = append(newArgs, expanded)
		}
		return e.reifyFuncCall(&ast.FuncCall{Loc: n.Loc, Func: newFunc, Args: newArgs}), nil

	case *ast.IfExpr:
		cond, err := e.expandExpr(n.Cond, depth+1)
		if err != nil {
			return nil, err
		}
		then, err := e.expandExpr(n.Then, depth+1)
		if err != nil {
			return nil, err
		}
		var elseExpr ast.Expr
		if n.Else != nil {
			elseExpr, err = e.expandExpr(n.Else, depth+1)
			if err != nil {
				return nil, err
			}
		}
		return &ast.IfExpr{Loc: n.Loc, Cond: cond, Then: then, Else: elseExpr}, nil

	case *ast.DoExpr:
		var newExprs []ast.Expr
		for _, ex := range n.Exprs {
			expanded, err := e.expandExpr(ex, depth+1)
			if err != nil {
				return nil, err
			}
			newExprs = append(newExprs, expanded)
		}
		return &ast.DoExpr{Loc: n.Loc, Exprs: newExprs}, nil

	case *ast.LetExpr:
		var newBindings []*ast.Binding
		for _, b := range n.Bindings {
			expanded, err := e.expandExpr(b.Init, depth+1)
			if err != nil {
				return nil, err
			}
			newBindings = append(newBindings, &ast.Binding{Name: b.Name, Init: expanded, Type: b.Type})
		}
		var newBody []ast.Expr
		for _, ex := range n.Body {
			expanded, err := e.expandExpr(ex, depth+1)
			if err != nil {
				return nil, err
			}
			newBody = append(newBody, expanded)
		}
		return &ast.LetExpr{Loc: n.Loc, Bindings: newBindings, Body: newBody}, nil

	case *ast.LoopExpr:
		var newBindings []*ast.Binding
		for _, b := range n.Bindings {
			expanded, err := e.expandExpr(b.Init, depth+1)
			if err != nil {
				return nil, err
			}
			newBindings = append(newBindings, &ast.Binding{Name: b.Name, Init: expanded, Type: b.Type})
		}
		var newBody []ast.Expr
		for _, ex := range n.Body {
			expanded, err := e.expandExpr(ex, depth+1)
			if err != nil {
				return nil, err
			}
			newBody = append(newBody, expanded)
		}
		return &ast.LoopExpr{Loc: n.Loc, Bindings: newBindings, Body: newBody}, nil

	case *ast.VectorLit:
		var newElts []ast.Expr
		for _, ex := range n.Elts {
			expanded, err := e.expandExpr(ex, depth+1)
			if err != nil {
				return nil, err
			}
			newElts = append(newElts, expanded)
		}
		return &ast.VectorLit{Loc: n.Loc, Elts: newElts}, nil

	case *ast.MapLit:
		var newPairs []ast.Pair
		for _, pair := range n.Pairs {
			k, err := e.expandExpr(pair.Key, depth+1)
			if err != nil {
				return nil, err
			}
			v, err := e.expandExpr(pair.Value, depth+1)
			if err != nil {
				return nil, err
			}
			newPairs = append(newPairs, ast.Pair{Key: k, Value: v})
		}
		return &ast.MapLit{Loc: n.Loc, Pairs: newPairs}, nil

	case *ast.BackquoteExpr:
		// When we hit a backquote in normal code (not inside a macro template),
		// it should be fully expanded (unquoted parts substituted).
		// For now, we expand it with no bindings.
		return e.expandBackquote(n.Expr, nil, depth+1)

	case *ast.UnquoteExpr:
		return e.expandExpr(n.Expr, depth+1)

	case *ast.UnquoteSpliceExpr:
		return e.expandExpr(n.Expr, depth+1)

	case *ast.RecurExpr:
		var newArgs []ast.Expr
		for _, arg := range n.Args {
			expanded, err := e.expandExpr(arg, depth+1)
			if err != nil {
				return nil, err
			}
			newArgs = append(newArgs, expanded)
		}
		return &ast.RecurExpr{Loc: n.Loc, Args: newArgs}, nil

	default:
		return n, nil
	}
}

func (e *Expander) expandMacroCall(macro *ast.Defmacro, call *ast.FuncCall, depth int) (ast.Expr, error) {
	bindings := make(map[string]interface{})

	for i, param := range macro.Params {
		if param.Variadic {
			bindings[param.Name] = call.Args[i:]
			break
		}
		if i < len(call.Args) {
			bindings[param.Name] = call.Args[i]
		}
	}

	if len(macro.Body) == 0 {
		return &ast.NilLit{}, nil
	}

	template := macro.Body[len(macro.Body)-1]

	// If template is backquoted, expand its content with bindings.
	if bq, ok := template.(*ast.BackquoteExpr); ok {
		return e.expandBackquote(bq.Expr, bindings, depth+1)
	}

	// Otherwise return template as is
	return template, nil
}

func (e *Expander) expandBackquote(expr ast.Expr, bindings map[string]interface{}, depth int) (ast.Expr, error) {
	if depth > 1000 {
		return nil, fmt.Errorf("maximum expansion depth exceeded in backquote")
	}

	switch n := expr.(type) {
	case *ast.UnquoteExpr:
		// If it's a symbol in bindings, return the bound AST node
		if sym, ok := n.Expr.(*ast.Symbol); ok {
			if val, ok := bindings[sym.Name]; ok {
				if astNode, ok := val.(ast.Expr); ok {
					return astNode, nil
				}
			}
		}
		// If not in bindings, return expression as is (or keep unquote if supporting nested)
		return n.Expr, nil

	case *ast.FuncCall:
		var newArgs []ast.Expr
		for _, arg := range n.Args {
			if splice, ok := arg.(*ast.UnquoteSpliceExpr); ok {
				if sym, ok := splice.Expr.(*ast.Symbol); ok {
					if val, ok := bindings[sym.Name]; ok {
						if list, ok := val.([]ast.Expr); ok {
							newArgs = append(newArgs, list...)
							continue
						}
					}
				}
			}
			expanded, err := e.expandBackquote(arg, bindings, depth+1)
			if err != nil {
				return nil, err
			}
			newArgs = append(newArgs, expanded)
		}

		newFunc, err := e.expandBackquote(n.Func, bindings, depth+1)
		if err != nil {
			return nil, err
		}

		return e.reifyFuncCall(&ast.FuncCall{Loc: n.Loc, Func: newFunc, Args: newArgs}), nil

	case *ast.VectorLit:
		var newElts []ast.Expr
		for _, elt := range n.Elts {
			if splice, ok := elt.(*ast.UnquoteSpliceExpr); ok {
				if sym, ok := splice.Expr.(*ast.Symbol); ok {
					if val, ok := bindings[sym.Name]; ok {
						if list, ok := val.([]ast.Expr); ok {
							newElts = append(newElts, list...)
							continue
						}
					}
				}
			}
			expanded, err := e.expandBackquote(elt, bindings, depth+1)
			if err != nil {
				return nil, err
			}
			newElts = append(newElts, expanded)
		}
		return &ast.VectorLit{Loc: n.Loc, Elts: newElts}, nil

	case *ast.IfExpr:
		cond, err := e.expandBackquote(n.Cond, bindings, depth+1)
		if err != nil {
			return nil, err
		}
		then, err := e.expandBackquote(n.Then, bindings, depth+1)
		if err != nil {
			return nil, err
		}
		var elseExpr ast.Expr
		if n.Else != nil {
			elseExpr, err = e.expandBackquote(n.Else, bindings, depth+1)
			if err != nil {
				return nil, err
			}
		}
		return &ast.IfExpr{Loc: n.Loc, Cond: cond, Then: then, Else: elseExpr}, nil

	case *ast.DoExpr:
		var newExprs []ast.Expr
		for _, ex := range n.Exprs {
			if splice, ok := ex.(*ast.UnquoteSpliceExpr); ok {
				if sym, ok := splice.Expr.(*ast.Symbol); ok {
					if val, ok := bindings[sym.Name]; ok {
						if list, ok := val.([]ast.Expr); ok {
							newExprs = append(newExprs, list...)
							continue
						}
					}
				}
			}
			expanded, err := e.expandBackquote(ex, bindings, depth+1)
			if err != nil {
				return nil, err
			}
			newExprs = append(newExprs, expanded)
		}
		return &ast.DoExpr{Loc: n.Loc, Exprs: newExprs}, nil

	case *ast.LetExpr:
		var newBindings []*ast.Binding
		for _, b := range n.Bindings {
			expanded, err := e.expandBackquote(b.Init, bindings, depth+1)
			if err != nil {
				return nil, err
			}
			newBindings = append(newBindings, &ast.Binding{Name: b.Name, Init: expanded, Type: b.Type})
		}
		var newBody []ast.Expr
		for _, ex := range n.Body {
			if splice, ok := ex.(*ast.UnquoteSpliceExpr); ok {
				if sym, ok := splice.Expr.(*ast.Symbol); ok {
					if val, ok := bindings[sym.Name]; ok {
						if list, ok := val.([]ast.Expr); ok {
							newBody = append(newBody, list...)
							continue
						}
					}
				}
			}
			expanded, err := e.expandBackquote(ex, bindings, depth+1)
			if err != nil {
				return nil, err
			}
			newBody = append(newBody, expanded)
		}
		return &ast.LetExpr{Loc: n.Loc, Bindings: newBindings, Body: newBody}, nil

	default:
		return n, nil
	}
}

func (e *Expander) reifyFuncCall(call *ast.FuncCall) ast.Expr {
	if sym, ok := call.Func.(*ast.Symbol); ok {
		switch sym.Name {
		case "if":
			if len(call.Args) >= 2 {
				var elseExpr ast.Expr
				if len(call.Args) >= 3 {
					elseExpr = call.Args[2]
				}
				return &ast.IfExpr{Loc: call.Loc, Cond: call.Args[0], Then: call.Args[1], Else: elseExpr}
			}
		case "do":
			return &ast.DoExpr{Loc: call.Loc, Exprs: call.Args}
		case "let":
			if len(call.Args) >= 1 {
				if v, ok := call.Args[0].(*ast.VectorLit); ok {
					var bindings []*ast.Binding
					for i := 0; i+1 < len(v.Elts); i += 2 {
						if nameSym, ok := v.Elts[i].(*ast.Symbol); ok {
							bindings = append(bindings, &ast.Binding{Name: nameSym.Name, Init: v.Elts[i+1]})
						}
					}
					return &ast.LetExpr{Loc: call.Loc, Bindings: bindings, Body: call.Args[1:]}
				}
			}
		}
	}
	return call
}
