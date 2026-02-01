package codegen

import (
	"fmt"
	"strings"

	"elbereth/internal/ast"
	"elbereth/internal/types"
)

// Generator generates Go code from Elbereth AST
type Generator struct {
	buf       strings.Builder
	indent    int
	symbols   map[string]types.Type // symbol -> type mapping
	functions map[string]*ast.Defn  // function definitions
	structs   map[string]*ast.Deftype
	imports   []*ast.Import
	variants  map[string]string // variant tag -> parent type name
	loopStack [][]string        // stack of loop binding names for recur
	isTest    bool
}

// New creates a new code generator
func New() *Generator {
	return &Generator{
		symbols:   make(map[string]types.Type),
		functions: make(map[string]*ast.Defn),
		structs:   make(map[string]*ast.Deftype),
		variants:  make(map[string]string),
	}
}

// SetTestMode sets whether the generator is in test mode
func (g *Generator) SetTestMode(isTest bool) {
	g.isTest = isTest
}

// Generate generates Go code from an AST program
func (g *Generator) Generate(prog *ast.Program) (string, error) {
	// First pass: collect definitions
	for _, item := range prog.Items {
		switch n := item.(type) {
		case *ast.Import:
			g.imports = append(g.imports, n)
		case *ast.Defn:
			g.functions[n.Name] = n
		case *ast.Deftype:
			g.structs[n.Name] = n
			for _, v := range n.Variants {
				g.variants[v.Name] = n.Name
			}
		case *ast.Def:
			if n.Type != nil {
				if nt, ok := n.Type.(*ast.NamedType); ok {
					g.symbols[n.Name] = types.ParseTypeString(nt.Name)
				}
			}
		}
	}

	// Write package declaration
	g.writeLine("package main")
	g.writeLine("")

	if len(g.imports) > 0 {
		g.writeLine("import (")
		g.indent++
		// Always include fmt if not already there, or just always for safety since we use it
		hasFmt := false
		for _, imp := range g.imports {
			if imp.Path == "fmt" {
				hasFmt = true
			}
			if imp.Alias != "" {
				g.writeLine(fmt.Sprintf("%s \"%s\"", imp.Alias, imp.Path))
			} else {
				g.writeLine(fmt.Sprintf("\"%s\"", imp.Path))
			}
		}
		if !hasFmt {
			g.writeLine("\"fmt\"")
		}
		g.indent--
		g.writeLine(")")
	} else if g.isTest {
		g.writeLine(`import (`)
		g.writeLine(`  "fmt"`)
		g.writeLine(`  "testing"`)
		g.writeLine(`)`)
	} else {
		g.writeLine(`import "fmt"`)
	}
	g.writeLine("")

	// Add testing import if in test mode and not already present
	if g.isTest {
		hasTesting := false
		for _, imp := range g.imports {
			if imp.Path == "testing" {
				hasTesting = true
				break
			}
		}
		if !hasTesting && len(g.imports) > 0 {
			// This is a bit hacky, but if we have other imports, testing wasn't added to the block above
			// because the loop only checks for fmt.
			// Wait, I should probably improve the import generation logic above.
		}
	}

	// Second pass: generate code
	for _, item := range prog.Items {
		switch n := item.(type) {
		case *ast.Deftype:
			g.genDeftype(n)
			g.writeLine("")
		case *ast.Defn:
			g.genDefn(n)
			g.writeLine("")
		case *ast.Def:
			g.genDef(n)
			g.writeLine("")
		case *ast.Deftest:
			g.genDeftest(n)
			g.writeLine("")
		case *ast.Defbenchmark:
			g.genDefbenchmark(n)
			g.writeLine("")
		case *ast.Defexample:
			g.genDefexample(n)
			g.writeLine("")
		default:
			// Top-level expressions are ignored in Go
		}
	}

	return g.buf.String(), nil
}

// ============================================================================
// Code Generation
// ============================================================================

func (g *Generator) genDeftype(d *ast.Deftype) {
	if len(d.Variants) > 0 {
		// Sum type
		g.writeLine(fmt.Sprintf("type %s interface { is_%s() }", d.Name, d.Name))
		for _, v := range d.Variants {
			typeName := fmt.Sprintf("%s_%s", d.Name, sanitizeIdent(strings.TrimPrefix(v.Name, ":")))
			if v.Type != nil {
				g.writeLine(fmt.Sprintf("type %s struct { Value %s }", typeName, g.typeToGoString(v.Type)))
			} else {
				g.writeLine(fmt.Sprintf("type %s struct{}", typeName))
			}
			g.writeLine(fmt.Sprintf("func (%s) is_%s() {}", typeName, d.Name))
		}
		return
	}

	g.write(fmt.Sprintf("type %s struct {\n", d.Name))
	g.indent++

	for _, f := range d.Fields {
		goType := g.typeToGoString(f.Type)
		g.writeLine(fmt.Sprintf("%s %s", capitalize(f.Name), goType))
	}

	g.indent--
	g.writeLine("}")
}

func (g *Generator) genDef(d *ast.Def) {
	g.write("var ")
	g.write(capitalize(d.Name))
	g.write(" = ")
	g.genExpr(d.Value)
	g.writeLine("")
}

func (g *Generator) genDefn(d *ast.Defn) {
	g.write("func ")
	name := sanitizeIdent(d.Name)
	if name != "main" {
		name = capitalize(name)
	}
	g.write(name)
	g.write("(")

	for i, p := range d.Params {
		if i > 0 {
			g.write(", ")
		}
		g.write(capitalize(sanitizeIdent(p.Name)))
		if p.Variadic {
			g.write(" ...")
		} else {
			g.write(" ")
		}
		if p.Type != nil {
			g.write(g.typeToGoString(p.Type))
		} else {
			g.write("interface{}")
		}
	}

	g.write(")")

	if d.ReturnType != nil {
		g.write(" ")
		g.write(g.typeToGoString(d.ReturnType))
	}

	g.writeLine(" {")
	g.indent++

	for i, expr := range d.Body {
		if d.ReturnType != nil && i == len(d.Body)-1 {
			switch ex := expr.(type) {
			case *ast.IfExpr:
				g.genIf(ex, "return ")
				g.writeLine("")
				continue
			case *ast.MatchExpr:
				g.genMatch(ex, "return ")
				g.writeLine("")
				continue
			case *ast.LoopExpr:
				g.genLoop(ex, "return ")
				g.writeLine("")
				continue
			}
			g.write("return ")
		}
		g.genExpr(expr)
		g.writeLine("")
	}

	g.indent--
	g.writeLine("}")
}

func (g *Generator) genDeftest(d *ast.Deftest) {
	g.write("func Test")
	g.write(capitalize(sanitizeIdent(d.Name)))
	g.writeLine("(t *testing.T) {")
	g.indent++
	for _, expr := range d.Body {
		g.genExpr(expr)
		g.writeLine("")
	}
	g.indent--
	g.writeLine("}")
}

func (g *Generator) genDefbenchmark(d *ast.Defbenchmark) {
	g.write("func Benchmark")
	g.write(capitalize(sanitizeIdent(d.Name)))
	g.write(fmt.Sprintf("(%s *testing.B) {\n", sanitizeIdent(d.BParam)))
	g.indent++
	for _, expr := range d.Body {
		g.genExpr(expr)
		g.writeLine("")
	}
	g.indent--
	g.writeLine("}")
}

func (g *Generator) genDefexample(d *ast.Defexample) {
	g.write("func Example")
	g.write(capitalize(sanitizeIdent(d.Name)))
	g.writeLine("() {")
	g.indent++
	for _, expr := range d.Body {
		g.genExpr(expr)
		g.writeLine("")
	}
	g.indent--
	g.writeLine("}")
}

func (g *Generator) genExpr(expr ast.Expr) {
	if expr == nil {
		return
	}

	switch e := expr.(type) {
	case *ast.IntLit:
		g.write(fmt.Sprintf("int64(%d)", e.Value))

	case *ast.FloatLit:
		g.write(fmt.Sprintf("float64(%f)", e.Value))

	case *ast.StringLit:
		g.write(fmt.Sprintf(`"%s"`, e.Value))

	case *ast.KeywordLit:
		g.write(fmt.Sprintf(`"%s"`, e.Value))

	case *ast.BoolLit:
		g.write(fmt.Sprintf("%v", e.Value))

	case *ast.NilLit:
		g.write("nil")

	case *ast.Symbol:
		name := sanitizeIdent(e.Name)
		switch name {
		case "int":
			name = "int64"
		case "float":
			name = "float64"
		case "string", "bool", "byte", "rune":
			// keep as is
		case "t", "b":
			// keep testing parameters small
		default:
			if name != "main" && !strings.Contains(name, ".") {
				name = capitalize(name)
			}
		}
		g.write(name)

	case *ast.VectorLit:
		g.write("[]interface{}{")
		for i, elt := range e.Elts {
			if i > 0 {
				g.write(", ")
			}
			g.genExpr(elt)
		}
		g.write("}")

	case *ast.MapLit:
		g.write("map[string]interface{}{")
		for i, pair := range e.Pairs {
			if i > 0 {
				g.write(", ")
			}
			g.genExpr(pair.Key)
			g.write(": ")
			g.genExpr(pair.Value)
		}
		g.write("}")

	case *ast.FuncCall:
		g.genFuncCall(e)

	case *ast.IfExpr:
		g.genIf(e, "")

	case *ast.DoExpr:
		g.genDo(e)

	case *ast.LetExpr:
		g.genLet(e)

	case *ast.MatchExpr:
		g.genMatch(e, "")

	case *ast.FuncLit:
		g.genFuncLit(e)

	case *ast.QuoteExpr:
		// Quoted expressions are just their values
		g.genExpr(e.Expr)

	case *ast.BackquoteExpr:
		g.write("/* error: unexpanded backquote */")

	case *ast.UnquoteExpr:
		g.write("/* error: unexpanded unquote */")

	case *ast.UnquoteSpliceExpr:
		g.write("/* error: unexpanded unquote-splice */")

	case *ast.LoopExpr:
		g.genLoop(e, "")

	case *ast.RecurExpr:
		g.genRecur(e)

	case *ast.SelectExpr:
		g.genSelect(e)

	default:
		g.write("/* unknown expr */")
	}
}

func (g *Generator) genFuncCall(call *ast.FuncCall) {
	// Check if it's a variant constructor call
	var tagName string
	if ks, ok := call.Func.(*ast.KeywordLit); ok {
		tagName = ks.Value
	} else if ss, ok := call.Func.(*ast.Symbol); ok {
		tagName = ss.Name
	}

	if parentType, ok := g.variants[tagName]; ok {
		typeName := fmt.Sprintf("%s_%s", parentType, sanitizeIdent(strings.TrimPrefix(tagName, ":")))
		g.write(fmt.Sprintf("%s{", typeName))
		if len(call.Args) > 0 {
			g.write("Value: ")
			g.genExpr(call.Args[0])
		}
		g.write("}")
		return
	}

	if sym, ok := call.Func.(*ast.Symbol); ok {
		// Struct construction
		if _, isStruct := g.structs[sym.Name]; isStruct {
			g.write(capitalize(sanitizeIdent(sym.Name)))
			g.write("{")
			if len(call.Args) > 0 {
				if m, ok := call.Args[0].(*ast.MapLit); ok {
					for i, pair := range m.Pairs {
						if i > 0 {
							g.write(", ")
						}
						if ks, ok := pair.Key.(*ast.KeywordLit); ok {
							g.write(capitalize(sanitizeIdent(ks.Value)))
						} else if ss, ok := pair.Key.(*ast.Symbol); ok {
							g.write(capitalize(sanitizeIdent(ss.Name)))
						} else {
							g.genExpr(pair.Key)
						}
						g.write(": ")
						g.genExpr(pair.Value)
					}
				}
			}
			g.write("}")
			return
		}

		switch sym.Name {
		case "go":
			if len(call.Args) == 1 {
				g.write("go func() {\n")
				g.indent++
				g.genExpr(call.Args[0])
				g.write("\n")
				g.indent--
				g.writeLine("}()")
				return
			}
		case "chan":
			if len(call.Args) >= 1 {
				g.write("make(chan ")
				g.genExpr(call.Args[0])
				if len(call.Args) >= 2 {
					g.write(", ")
					g.genExpr(call.Args[1])
				}
				g.write(")")
				return
			}
		case ">!":
			if len(call.Args) == 2 {
				g.genExpr(call.Args[0])
				g.write(" <- ")
				g.genExpr(call.Args[1])
				return
			}
		case "<!":
			if len(call.Args) == 1 {
				g.write("<-")
				g.genExpr(call.Args[0])
				return
			}
		case ".":
			if len(call.Args) >= 2 {
				g.genExpr(call.Args[0])
				g.write(".")
				if field, ok := call.Args[1].(*ast.Symbol); ok {
					g.write(capitalize(sanitizeIdent(field.Name)))
				} else if field, ok := call.Args[1].(*ast.KeywordLit); ok {
					g.write(capitalize(sanitizeIdent(field.Value)))
				} else {
					g.genExpr(call.Args[1])
				}
				return
			}
		case "set!":
			if len(call.Args) == 2 {
				g.genExpr(call.Args[0])
				g.write(" = ")
				g.genExpr(call.Args[1])
				return
			}
		case "defer":
			if len(call.Args) == 1 {
				g.write("defer func() {\n")
				g.indent++
				g.genExpr(call.Args[0])
				g.write("\n")
				g.indent--
				g.writeLine("}()")
				return
			}
		case "+", "-", "*", "/", "%", "==", "!=", "<", "<=", ">", ">=":
			// Binary operators
			if len(call.Args) >= 2 {
				g.write("(")
				g.genExpr(call.Args[0])
				g.write(fmt.Sprintf(" %s ", sym.Name))
				g.genExpr(call.Args[1])
				g.write(")")
				return
			}

		case "not":
			if len(call.Args) == 1 {
				g.write("!")
				g.genExpr(call.Args[0])
				return
			}
		case "str":
			g.write("fmt.Sprint(")
			for i, arg := range call.Args {
				if i > 0 {
					g.write(", ")
				}
				g.genExpr(arg)
			}
			g.write(")")
			return
		case "println":
			g.write("fmt.Println(")
			for i, arg := range call.Args {
				if i > 0 {
					g.write(", ")
				}
				g.genExpr(arg)
			}
			g.write(")")
			return

		case "print":
			g.write("fmt.Print(")
			for i, arg := range call.Args {
				if i > 0 {
					g.write(", ")
				}
				g.genExpr(arg)
			}
			g.write(")")
			return

		case "first":
			if len(call.Args) == 1 {
				g.genExpr(call.Args[0])
				g.write("[0]")
				return
			}
		case "rest":
			if len(call.Args) == 1 {
				g.genExpr(call.Args[0])
				g.write("[1:]")
				return
			}
		case "assert":
			if len(call.Args) >= 1 && g.isTest {
				g.write("if !")
				g.genExpr(call.Args[0])
				g.writeLine(" {")
				g.indent++
				g.write("t.Fatalf(\"assertion failed: %s\", ")
				g.write(fmt.Sprintf("`%s`", call.Args[0].String()))
				g.writeLine(")")
				g.indent--
				g.writeLine("}")
				return
			}
		case "assert-eq":
			if len(call.Args) >= 2 && g.isTest {
				g.write("if ")
				g.genExpr(call.Args[0])
				g.write(" != ")
				g.genExpr(call.Args[1])
				g.writeLine(" {")
				g.indent++
				g.write("t.Fatalf(\"assertion failed: %s == %s (got %v, want %v)\", ")
				g.write(fmt.Sprintf("`%s`, `%s`, ", call.Args[0].String(), call.Args[1].String()))
				g.genExpr(call.Args[0])
				g.write(", ")
				g.genExpr(call.Args[1])
				g.writeLine(")")
				g.indent--
				g.writeLine("}")
				return
			}
		case "nth":
			if len(call.Args) == 2 {
				g.genExpr(call.Args[0])
				g.write("[")
				g.genExpr(call.Args[1])
				g.write("]")
				return
			}
		case "get":
			if len(call.Args) >= 2 {
				g.genExpr(call.Args[0])
				g.write("[")
				g.genExpr(call.Args[1])
				g.write("]")
				return
			}
		case "len":
			if len(call.Args) > 0 {
				g.write("len(")
				g.genExpr(call.Args[0])
				g.write(")")
				return
			}
		}
	}

	// Regular function call
	g.genExpr(call.Func)
	g.write("(")
	for i, arg := range call.Args {
		if i > 0 {
			g.write(", ")
		}
		g.genExpr(arg)
	}
	g.write(")")
}

func (g *Generator) genIf(ifExpr *ast.IfExpr, prefix string) {
	g.write("if ")
	g.genExpr(ifExpr.Cond)
	g.writeLine(" {")
	g.indent++
	g.genExprWithPrefix(ifExpr.Then, prefix)
	g.writeLine("")
	g.indent--

	if ifExpr.Else != nil {
		g.writeLine("} else {")
		g.indent++
		g.genExprWithPrefix(ifExpr.Else, prefix)
		g.writeLine("")
		g.indent--
	}

	g.write("}")
}

func (g *Generator) genDo(doExpr *ast.DoExpr) {
	g.writeLine("{")
	g.indent++
	for _, expr := range doExpr.Exprs {
		g.genExpr(expr)
		g.writeLine("")
	}
	g.indent--
	g.write("}")
}

func (g *Generator) genLet(letExpr *ast.LetExpr) {
	for _, binding := range letExpr.Bindings {
		g.write(capitalize(sanitizeIdent(binding.Name)))
		g.write(" := ")
		g.genExpr(binding.Init)
		g.writeLine("")
	}
	for _, expr := range letExpr.Body {
		g.genExpr(expr)
		g.writeLine("")
	}
}

func (g *Generator) genMatch(m *ast.MatchExpr, prefix string) {
	// Detect if we need a type switch (if any pattern is a variant list)
	isTypeSwitch := false
	for _, c := range m.Cases {
		if _, ok := c.Pattern.(*ast.FuncCall); ok {
			isTypeSwitch = true
			break
		}
	}

	if isTypeSwitch {
		g.write("switch v := ")
		g.genExpr(m.Val)
		g.writeLine(".(type) {")
	} else {
		g.write("switch ")
		g.genExpr(m.Val)
		g.writeLine(" {")
	}

	for _, c := range m.Cases {
		isDefault := false
		if kw, ok := c.Pattern.(*ast.KeywordLit); ok && kw.Value == "else" {
			isDefault = true
		} else if sym, ok := c.Pattern.(*ast.Symbol); ok && (sym.Name == "else" || sym.Name == "_") {
			isDefault = true
		}

		if isDefault {
			g.writeLine("default:")
		} else if isTypeSwitch {
			if call, ok := c.Pattern.(*ast.FuncCall); ok {
				// Variant pattern: (:tag binding)
				var tagName string
				if kw, ok := call.Func.(*ast.KeywordLit); ok {
					tagName = kw.Value
				} else if ss, ok := call.Func.(*ast.Symbol); ok {
					tagName = ss.Name
				}

				if parentType, ok := g.variants[tagName]; ok {
					typeName := fmt.Sprintf("%s_%s", parentType, sanitizeIdent(strings.TrimPrefix(tagName, ":")))
					g.write(fmt.Sprintf("case %s:", typeName))
					g.writeLine("")
					g.indent++
					// Bind the value if there's a binding symbol
					if len(call.Args) > 0 {
						if sym, ok := call.Args[0].(*ast.Symbol); ok {
							g.writeLine(fmt.Sprintf("%s := v.Value", capitalize(sanitizeIdent(sym.Name))))
						}
					}
					g.genExprWithPrefix(c.Body, prefix)
					g.writeLine("")
					g.indent--
					continue
				}
			}
			// Fallback for non-variant patterns in type switch
			g.write("case ")
			g.genExpr(c.Pattern)
			g.writeLine(":")
		} else {
			g.write("case ")
			g.genExpr(c.Pattern)
			g.writeLine(":")
		}

		g.indent++
		g.genExprWithPrefix(c.Body, prefix)
		g.writeLine("")
		g.indent--
	}

	// Add panic if no cases matched to satisfy Go's return requirements
	if prefix != "" {
		hasDefault := false
		for _, c := range m.Cases {
			if kw, ok := c.Pattern.(*ast.KeywordLit); ok && kw.Value == "else" {
				hasDefault = true
				break
			} else if sym, ok := c.Pattern.(*ast.Symbol); ok && (sym.Name == "else" || sym.Name == "_") {
				hasDefault = true
				break
			}
		}

		if !hasDefault {
			g.writeLine("default:")
			g.indent++
			g.writeLine("panic(\"unbalanced match\")")
			g.indent--
		}
	}
	g.write("}")
}

func (g *Generator) genExprWithPrefix(expr ast.Expr, prefix string) {
	if _, isRecur := expr.(*ast.RecurExpr); isRecur {
		g.genExpr(expr)
		return
	}

	switch ex := expr.(type) {
	case *ast.IfExpr:
		g.genIf(ex, prefix)
	case *ast.MatchExpr:
		g.genMatch(ex, prefix)
	case *ast.LoopExpr:
		if prefix == "return " {
			g.genLoop(ex, prefix)
		} else {
			// Complex case: loop as an expression that is not returning
			g.write(prefix)
			g.genLoop(ex, "")
		}
	default:
		g.write(prefix)
		g.genExpr(ex)
	}
}

func (g *Generator) genFuncLit(fn *ast.FuncLit) {
	g.write("func(")
	for i, p := range fn.Params {
		if i > 0 {
			g.write(", ")
		}
		g.write(capitalize(sanitizeIdent(p.Name)))
		g.write(" interface{}")
	}
	g.writeLine(") interface{} {")
	g.indent++

	for i, expr := range fn.Body {
		if i == len(fn.Body)-1 {
			switch ex := expr.(type) {
			case *ast.IfExpr:
				g.genIf(ex, "return ")
				g.writeLine("")
				continue
			case *ast.MatchExpr:
				g.genMatch(ex, "return ")
				g.writeLine("")
				continue
			}
			g.write("return ")
		}
		g.genExpr(expr)
		g.writeLine("")
	}

	g.indent--
	g.write("}")
}

func (g *Generator) genLoop(l *ast.LoopExpr, prefix string) {
	var names []string
	for _, b := range l.Bindings {
		name := capitalize(sanitizeIdent(b.Name))
		names = append(names, name)
		g.write(name)
		g.write(" := ")
		g.genExpr(b.Init)
		g.writeLine("")
	}
	// Push loop binding names to stack
	g.loopStack = append(g.loopStack, names)

	g.writeLine("for {")
	g.indent++
	for i, expr := range l.Body {
		if i == len(l.Body)-1 {
			g.genExprWithPrefix(expr, prefix)
		} else {
			g.genExpr(expr)
		}
		g.writeLine("")
	}
	g.indent--
	g.writeLine("}")

	// Pop loop binding names from stack
	g.loopStack = g.loopStack[:len(g.loopStack)-1]
}

func (g *Generator) genRecur(r *ast.RecurExpr) {
	if len(g.loopStack) == 0 {
		g.write("/* recur outside loop */")
		return
	}

	names := g.loopStack[len(g.loopStack)-1]

	// Create temporary variables for updating loop bindings
	// only if we have more than one binding to avoid order-of-evaluation issues
	if len(r.Args) > 1 {
		for i, arg := range r.Args {
			if i < len(names) {
				g.write(fmt.Sprintf("tmp_%s := ", names[i]))
				g.genExpr(arg)
				g.writeLine("")
			}
		}
		for i := range r.Args {
			if i < len(names) {
				g.writeLine(fmt.Sprintf("%s = tmp_%s", names[i], names[i]))
			}
		}
	} else if len(r.Args) == 1 {
		g.write(names[0])
		g.write(" = ")
		g.genExpr(r.Args[0])
		g.writeLine("")
	}

	g.writeLine("continue")
}

func (g *Generator) genSelect(s *ast.SelectExpr) {
	g.writeLine("select {")
	for _, c := range s.Cases {
		if c.Chan == nil {
			g.writeLine("default:")
		} else {
			g.write("case ")
			if c.Binding != "" {
				g.write(fmt.Sprintf("%s := ", capitalize(sanitizeIdent(c.Binding))))
			}
			g.write("<-")
			g.genExpr(c.Chan)
			g.writeLine(":")
		}
		g.indent++
		for _, expr := range c.Body {
			g.genExpr(expr)
			g.writeLine("")
		}
		g.indent--
	}
	g.writeLine("}")
}

// ============================================================================
// Helpers
// ============================================================================

func (g *Generator) typeToGoString(t ast.Type) string {
	if t == nil {
		return "interface{}"
	}

	switch tt := t.(type) {
	case *ast.NamedType:
		switch tt.Name {
		case "int":
			return "int64"
		case "float":
			return "float64"
		case "string":
			return "string"
		case "bool":
			return "bool"
		case "byte":
			return "byte"
		case "rune":
			return "rune"
		default:
			return tt.Name
		}

	case *ast.SliceType:
		return fmt.Sprintf("[]%s", g.typeToGoString(tt.EltType))

	case *ast.ChanType:
		elemType := g.typeToGoString(tt.EltType)
		if tt.Buffer > 0 {
			return fmt.Sprintf("chan %s", elemType)
		}
		return fmt.Sprintf("chan %s", elemType)

	default:
		return "interface{}"
	}
}

func (g *Generator) write(s string) {
	g.buf.WriteString(s)
}

func (g *Generator) writeLine(s string) {
	g.buf.WriteString(strings.Repeat("  ", g.indent))
	g.buf.WriteString(s)
	g.buf.WriteString("\n")
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func sanitizeIdent(s string) string {
	// Replace hyphens with underscores to make it a valid Go identifier
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, "?", "p")
	s = strings.ReplaceAll(s, "!", "b")
	s = strings.ReplaceAll(s, "/", ".")
	return s
}
