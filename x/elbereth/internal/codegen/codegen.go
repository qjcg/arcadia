package codegen

import (
	"fmt"
	"strings"

	"github.com/qjcg/arcadia/x/elbereth/internal/ast"
	"github.com/qjcg/arcadia/x/elbereth/internal/lexer"
	"github.com/qjcg/arcadia/x/elbereth/internal/types"
)

// Generator generates Go code from Elbereth AST
type Generator struct {
	buf           strings.Builder
	indent        int
	symbols       map[string]types.Type // symbol -> type mapping
	inferredTypes map[ast.Expr]types.Type
	functions     map[string]*ast.Defn // function definitions
	structs       map[string]*ast.Deftype
	imports       []ast.ImportSpec
	variants      map[string]string // variant tag -> parent type name
	loopStack     [][]string        // stack of loop binding names for recur
	isTest        bool
	pkgNames      map[string]bool
	needsFmt      bool
}

// New creates a new code generator
func New() *Generator {
	return &Generator{
		symbols:   make(map[string]types.Type),
		functions: make(map[string]*ast.Defn),
		structs:   make(map[string]*ast.Deftype),
		variants:  make(map[string]string),
		pkgNames:  make(map[string]bool),
	}
}

func (g *Generator) checkFmtUsage(node ast.Node) {
	if node == nil || g.needsFmt {
		return
	}

	switch n := node.(type) {
	case *ast.Program:
		for _, item := range n.Items {
			g.checkFmtUsage(item)
		}
	case *ast.Defn:
		for _, expr := range n.Body {
			g.checkFmtUsage(expr)
		}
	case *ast.FuncCall:
		if sym, ok := n.Func.(*ast.Symbol); ok {
			if sym.Name == "println" || sym.Name == "print" || sym.Name == "str" {
				g.needsFmt = true
				return
			}
		}
		g.checkFmtUsage(n.Func)
		for _, arg := range n.Args {
			g.checkFmtUsage(arg)
		}
	case *ast.IfExpr:
		g.checkFmtUsage(n.Cond)
		g.checkFmtUsage(n.Then)
		g.checkFmtUsage(n.Else)
	case *ast.DoExpr:
		for _, expr := range n.Exprs {
			g.checkFmtUsage(expr)
		}
	case *ast.LetExpr:
		for _, b := range n.Bindings {
			g.checkFmtUsage(b.Init)
		}
		for _, expr := range n.Body {
			g.checkFmtUsage(expr)
		}
	case *ast.LoopExpr:
		for _, b := range n.Bindings {
			g.checkFmtUsage(b.Init)
		}
		for _, expr := range n.Body {
			g.checkFmtUsage(expr)
		}
	case *ast.DoseqExpr:
		g.checkFmtUsage(n.Coll)
		for _, expr := range n.Body {
			g.checkFmtUsage(expr)
		}
	case *ast.Deftest:
		for _, expr := range n.Body {
			g.checkFmtUsage(expr)
		}
	case *ast.Defbenchmark:
		for _, expr := range n.Body {
			g.checkFmtUsage(expr)
		}
	case *ast.Defexample:
		for _, expr := range n.Body {
			g.checkFmtUsage(expr)
		}
	}
}

// SetTestMode sets whether the generator is in test mode
func (g *Generator) SetTestMode(isTest bool) {
	g.isTest = isTest
}

// Generate generates Go code from an AST program
func (g *Generator) Generate(prog *ast.Program) (string, error) {
	// Run type check
	checker := types.NewChecker()
	inferred, err := checker.Check(prog)
	if err != nil {
		return "", err
	}
	g.inferredTypes = inferred

	// First pass: collect definitions and check fmt usage
	for _, item := range prog.Items {
		g.checkFmtUsage(item)
		switch n := item.(type) {
		case *ast.Import:
			for _, spec := range n.Specs {
				g.imports = append(g.imports, spec)
				if spec.Alias != "" {
					g.pkgNames[spec.Alias] = true
				} else {
					parts := strings.Split(spec.Path, "/")
					g.pkgNames[parts[len(parts)-1]] = true
				}
			}
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
	pkgName := "main"
	if prog.Package != "" {
		pkgName = prog.Package
	}
	g.writeLine(fmt.Sprintf("package %s", pkgName))
	g.writeLine("")

	if g.isTest {
		g.writeLine(`import (`)
		g.writeLine(`  "testing"`)
		hasFmt := false
		for _, imp := range g.imports {
			if imp.Path == "fmt" {
				hasFmt = true
			}
		}
		if g.needsFmt && !hasFmt {
			g.writeLine(`  "fmt"`)
		}

		for _, imp := range g.imports {
			if imp.Path == "testing" || imp.Path == "fmt" {
				continue
			}
			if imp.Alias != "" {
				g.writeLine(fmt.Sprintf("  %s \"%s\"", imp.Alias, imp.Path))
			} else {
				g.writeLine(fmt.Sprintf("  \"%s\"", imp.Path))
			}
		}
		g.writeLine(`)`)
	} else if len(g.imports) > 0 {
		g.writeLine("import (")
		g.indent++
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
		if g.needsFmt && !hasFmt {
			g.writeLine("\"fmt\"")
		}
		g.indent--
		g.writeLine(")")
	} else {
		if g.needsFmt {
			g.writeLine(`import "fmt"`)
		}
	}
	g.writeLine("")

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
		case *ast.DoseqExpr:
			g.genDoseq(n)
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
		case *ast.Package:
			// Package name handled in first pass
		case *ast.Import:
			// Imports handled in first pass
		case *ast.FuncCall:
			g.genExpr(n)
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
		name := sanitizeIdent(d.Name)
		g.writeLine(fmt.Sprintf("type %s interface { is_%s() }", name, name))
		for _, v := range d.Variants {
			vName := sanitizeIdent(strings.TrimPrefix(v.Name, ":"))
			typeName := fmt.Sprintf("%s_%s", name, vName)
			if v.Type != nil {
				g.writeLine(fmt.Sprintf("type %s struct { Value %s }", typeName, g.typeToGoString(v.Type)))
			} else {
				g.writeLine(fmt.Sprintf("type %s struct{}", typeName))
			}
			g.writeLine(fmt.Sprintf("func (%s) is_%s() {}", typeName, name))
		}
		return
	}

	g.write(fmt.Sprintf("type %s struct {\n", sanitizeIdent(d.Name)))
	g.indent++

	for _, f := range d.Fields {
		goType := g.typeToGoString(f.Type)
		g.writeLine(fmt.Sprintf("%s %s", sanitizeIdent(f.Name), goType))
	}

	g.indent--
	g.writeLine("}")
}

func (g *Generator) genDef(d *ast.Def) {
	name := sanitizeIdent(d.Name)
	g.genExprWithPrefix(d.Value, "var "+name+" = ")
	g.writeLine("")
}

func (g *Generator) genDefn(d *ast.Defn) {
	name := sanitizeIdent(d.Name)
	if name == "main" {
		// main must be lowercase for package main
	}

	g.write("func ")
	g.write(name)
	g.write("(")

	for i, p := range d.Params {
		if i > 0 {
			g.write(", ")
		}
		pName := sanitizeIdent(p.Name)
		g.write(pName)
		if p.Variadic {
			g.write(" ...")
		} else {
			g.write(" ")
		}
		if p.Type != nil {
			g.write(g.typeToGoString(p.Type))
		} else {
			g.write("any")
		}
	}

	g.write(")")

	isMain := name == "main"
	if d.ReturnType != nil {
		g.write(" ")
		g.write(g.typeToGoString(d.ReturnType))
	} else if !isMain {
		g.write(" any")
	}

	g.writeLine(" {")
	g.indent++

	for i, expr := range d.Body {
		if !isMain && i == len(d.Body)-1 {
			g.genExprWithPrefix(expr, "return ")
			g.writeLine("")
		} else {
			g.genExpr(expr)
			g.writeLine("")
		}
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
	// Output comment is required for examples to run in a standalone file?
	// Actually, just having an output comment is standard for example documentation.
	g.writeLine("// Output:")
	g.indent--
	g.writeLine("}")
}

func (g *Generator) genExprAsValue(expr ast.Expr) {
	switch expr.(type) {
	case *ast.IfExpr, *ast.MatchExpr, *ast.LoopExpr, *ast.DoExpr, *ast.DoseqExpr:
		g.writeLine("func() any {")
		g.indent++
		g.genExprWithPrefix(expr, "return ")
		g.writeLine("")
		g.indent--
		g.write("}()")
	default:
		g.genExpr(expr)
	}
}

func (g *Generator) genExpr(expr ast.Expr) {
	if expr == nil {
		return
	}

	switch e := expr.(type) {
	case *ast.IntLit:
		g.write(fmt.Sprintf("int64(%d)", e.Value))

	case *ast.FloatLit:
		g.write(fmt.Sprintf("%f", e.Value))

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
			isSpecial := name == "main" || name == "assert" || name == "assert_eq" ||
				name == "assert_true" || name == "assert_false" || name == "assert_err" ||
				name == "for" || name == "go" || name == "chan" || name == "defer" ||
				name == "select" || name == "recur" || name == "loop" || name == "let" ||
				name == "do" || name == "if" || name == "fn" || name == "setb"

			_ = isSpecial // keep for now
		}
		g.write(name)

	case *ast.VectorLit:
		g.write("[]any{")
		for i, elt := range e.Elts {
			if i > 0 {
				g.write(", ")
			}
			g.genExprAsValue(elt)
		}
		g.write("}")

	case *ast.MapLit:
		g.write("map[string]any{")
		for i, pair := range e.Pairs {
			if i > 0 {
				g.write(", ")
			}
			g.genExprAsValue(pair.Key)
			g.write(": ")
			g.genExprAsValue(pair.Value)
		}
		g.write("}")

	case *ast.FuncCall:
		g.genFuncCall(e)

	case *ast.IfExpr:
		g.genIf(e, "")

	case *ast.DoExpr:
		g.genDo(e, "")

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

	case *ast.DoseqExpr:
		g.genDoseq(e)

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
							g.write(sanitizeIdent(ks.Value))
						} else if ss, ok := pair.Key.(*ast.Symbol); ok {
							g.write(sanitizeIdent(ss.Name))
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
		case "addr", "&":
			if len(call.Args) == 1 {
				g.write("&(")
				g.genExpr(call.Args[0])
				g.write(")")
				return
			}
		case "go":
			if len(call.Args) >= 1 {
				g.write("go func() {\n")
				g.indent++
				for _, arg := range call.Args {
					g.genExpr(arg)
					g.writeLine("")
				}
				g.indent--
				g.writeLine("}()")
				return
			}
		case "defer":
			if len(call.Args) >= 1 {
				g.write("defer func() {\n")
				g.indent++
				for _, arg := range call.Args {
					g.genExpr(arg)
					g.writeLine("")
				}
				g.indent--
				g.writeLine("}()")
				return
			}
		case "chan":
			if len(call.Args) >= 1 {
				g.write("make(chan ")
				g.genExpr(call.Args[0])
				if len(call.Args) >= 2 {
					g.write(", int(")
					g.genExpr(call.Args[1])
					g.write(")")
				}
				g.write(")")
				return
			}
		case ">!":
			if len(call.Args) == 2 {
				g.write("func() any { ")
				g.genExpr(call.Args[0])
				g.write(" <- ")
				g.genExpr(call.Args[1])
				g.write("; return nil }()")
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
				// Handle Type Assertions (. val Type)
				// Improved detection: if the second arg is a symbol containing a dot (like html.Attribute)
				if ss, ok := call.Args[1].(*ast.Symbol); ok && (strings.Contains(ss.Name, ".") || strings.HasPrefix(ss.Name, "(") || isTypeSymbol(lexer.Token{Value: ss.Name})) && len(call.Args) == 2 {
					g.genExpr(call.Args[0])
					g.write(".(")
					g.write(sanitizeIdent(ss.Name))
					g.write(")")
					return
				}

				fieldName := ""
				if field, ok := call.Args[1].(*ast.Symbol); ok {
					fieldName = sanitizeIdent(field.Name)
				} else if field, ok := call.Args[1].(*ast.KeywordLit); ok {
					fieldName = sanitizeIdent(field.Value)
				}

				if fieldName == "N" {
					g.write("int64(")
					g.genExpr(call.Args[0])
					g.write(".N)")
				} else {
					g.genExpr(call.Args[0])
					g.write(".")
					if fieldName != "" {
						g.write(fieldName)
					} else {
						g.genExpr(call.Args[1])
					}
				}

				if len(call.Args) > 2 {
					g.write("(")
					for i := 2; i < len(call.Args); i++ {
						if i > 2 {
							g.write(", ")
						}
						// Heuristic: if it's sync.WaitGroup.Add, cast to int
						if fieldName == "Add" {
							g.write("int(")
							g.genExpr(call.Args[i])
							g.write(")")
						} else {
							g.genExpr(call.Args[i])
						}
					}
					g.write(")")
				} else {
					// Heuristic for 0-arg methods in Elbereth
					if fieldName == "Done" || fieldName == "Wait" || fieldName == "ReportAllocs" || fieldName == "Skip" || fieldName == "Helper" {
						g.write("()")
					}
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
		case "+", "-", "*", "/", "%":
			// Numeric operators
			if len(call.Args) >= 2 {
				g.write("(")
				for i, arg := range call.Args {
					if i > 0 {
						g.write(fmt.Sprintf(" %s ", sym.Name))
					}
					g.genExprWithNumericAssertion(arg)
				}
				g.write(")")
				return
			}

		case "==", "!=", "<", "<=", ">", ">=", "=":
			// Comparison operators
			if len(call.Args) >= 2 {
				g.write("(")
				op := sym.Name
				if op == "=" {
					op = "=="
				}

				// Try to detect if we should use numeric assertions
				isNumeric := false
				if t, ok := g.inferredTypes[call.Args[0]]; ok && t == types.Int {
					isNumeric = true
				} else if t, ok := g.inferredTypes[call.Args[1]]; ok && t == types.Int {
					isNumeric = true
				} else if _, ok := call.Args[0].(*ast.IntLit); ok {
					isNumeric = true
				} else if _, ok := call.Args[1].(*ast.IntLit); ok {
					isNumeric = true
				}

				if isNumeric {
					g.genExprWithNumericAssertion(call.Args[0])
					g.write(fmt.Sprintf(" %s ", op))
					g.genExprWithNumericAssertion(call.Args[1])
				} else {
					g.write("(")
					g.genExpr(call.Args[0])
					g.write(fmt.Sprintf(") %s (", op))
					g.genExpr(call.Args[1])
					g.write(")")
				}
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
			g.write("func() any { fmt.Println(")
			for i, arg := range call.Args {
				if i > 0 {
					g.write(", ")
				}
				g.genExpr(arg)
			}
			g.write("); return nil }()")
			return

		case "print":
			g.write("func() any { fmt.Print(")
			for i, arg := range call.Args {
				if i > 0 {
					g.write(", ")
				}
				g.genExpr(arg)
			}
			g.write("); return nil }()")
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
		case "last":
			if len(call.Args) == 1 {
				g.genExpr(call.Args[0])
				g.write("[len(")
				g.genExpr(call.Args[0])
				g.write(")-1]")
				return
			}
		case "for":
			if len(call.Args) >= 3 {
				g.write("for ")
				if v, ok := call.Args[0].(*ast.VectorLit); ok {
					for i := 0; i < len(v.Elts); i += 2 {
						if i > 0 {
							g.write(", ")
						}
						g.genExpr(v.Elts[i])
						g.write(" := ")
						g.genExpr(v.Elts[i+1])
					}
				}
				g.write("; ")
				g.genExpr(call.Args[1])
				g.write("; ")
				if v, ok := call.Args[2].(*ast.VectorLit); ok {
					for i := 0; i < len(v.Elts); i++ {
						if i > 0 {
							g.write(", ")
						}
						// This is a bit tricky, assuming the update corresponds to the init
						initVar := ""
						if initV, ok := call.Args[0].(*ast.VectorLit); ok && i*2 < len(initV.Elts) {
							if sym, ok := initV.Elts[i*2].(*ast.Symbol); ok {
								initVar = sanitizeIdent(sym.Name)
							}
						}
						if initVar != "" {
							g.write(initVar + " = ")
						}
						g.genExpr(v.Elts[i])
					}
				}
				g.writeLine(" {")
				g.indent++
				for _, expr := range call.Args[3:] {
					g.genExpr(expr)
					g.writeLine("")
				}
				g.indent--
				g.writeLine("}")
				return
			}
		case "assert":
			if len(call.Args) >= 1 && g.isTest {
				g.write("if !")
				g.genExpr(call.Args[0])
				needsCast := true
				if _, ok := call.Args[0].(*ast.BoolLit); ok {
					needsCast = false
				} else if fc, ok := call.Args[0].(*ast.FuncCall); ok {
					if sym, ok := fc.Func.(*ast.Symbol); ok {
						switch sym.Name {
						case "==", "!=", "<", "<=", ">", ">=", "and", "or", "not":
							needsCast = false
						}
					}
				}
				if needsCast {
					g.write(".(bool)")
				}
				g.writeLine(" {")
				g.indent++
				g.write("t.Fatalf(\"assertion failed: %s\", ")
				g.write("`" + call.Args[0].String() + "`")
				g.writeLine(")")
				g.indent--
				g.writeLine("}")
				return
			}
		case "assert-true":
			if len(call.Args) >= 1 && g.isTest {
				g.write("if !")
				g.genExpr(call.Args[0])
				needsCast := true
				if _, ok := call.Args[0].(*ast.BoolLit); ok {
					needsCast = false
				} else if fc, ok := call.Args[0].(*ast.FuncCall); ok {
					if sym, ok := fc.Func.(*ast.Symbol); ok {
						switch sym.Name {
						case "==", "!=", "<", "<=", ">", ">=", "and", "or", "not":
							needsCast = false
						}
					}
				}
				if needsCast {
					g.write(".(bool)")
				}
				g.writeLine(" {")
				g.indent++
				g.write("t.Fatalf(\"assertion failed: %s\", ")
				g.write("`" + call.Args[0].String() + "`")
				g.writeLine(")")
				g.indent--
				g.writeLine("}")
				return
			}
		case "assert-false":
			if len(call.Args) >= 1 && g.isTest {
				g.write("if ")
				g.genExpr(call.Args[0])
				needsCast := true
				if _, ok := call.Args[0].(*ast.BoolLit); ok {
					needsCast = false
				} else if fc, ok := call.Args[0].(*ast.FuncCall); ok {
					if sym, ok := fc.Func.(*ast.Symbol); ok {
						switch sym.Name {
						case "==", "!=", "<", "<=", ">", ">=", "and", "or", "not":
							needsCast = false
						}
					}
				}
				if needsCast {
					g.write(".(bool)")
				}
				g.writeLine(" {")
				g.indent++
				g.write("t.Fatalf(\"assertion failed: NOT %s\", ")
				g.write("`" + call.Args[0].String() + "`")
				g.writeLine(")")
				g.indent--
				g.writeLine("}")
				return
			}
		case "assert-err":
			if len(call.Args) >= 1 && g.isTest {
				// Result sums have :err variant.
				// We assume it's a sum type with is_Result() or similar if using Sum types.
				// Based on genDeftype, it generates Name interface and Name_Variant structs.
				g.write("if _, ok := ")
				g.genExpr(call.Args[0])
				g.write(".(interface{ is_err() }); !ok {")
				g.writeLine("")
				g.indent++
				g.write("t.Fatalf(\"assertion failed: expected error, got %v\", ")
				g.genExpr(call.Args[0])
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
				g.write("`" + call.Args[0].String() + "`, `" + call.Args[1].String() + "`, ")
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
				g.write("int(")
				g.genExpr(call.Args[1])
				g.write(")]")
				return
			}
		case "get":
			if len(call.Args) >= 2 {
				g.genExpr(call.Args[0])
				g.write("[")
				g.write("int(")
				g.genExpr(call.Args[1])
				g.write(")]")
				return
			}
		case "mod":
			if len(call.Args) == 2 {
				g.write("(")
				g.genExprWithNumericAssertion(call.Args[0])
				g.write(" % ")
				g.genExprWithNumericAssertion(call.Args[1])
				g.write(")")
				return
			}
		case "len":
			if len(call.Args) > 0 {
				g.write("len(")
				g.genExpr(call.Args[0])
				g.write(")")
				return
			}
		case "map":
			if len(call.Args) == 2 {
				// (map fn coll)
				g.write("func() []any {\n")
				g.indent++
				g.write("coll := ")
				g.genExpr(call.Args[1])
				g.writeLine("")
				g.writeLine("res := make([]any, len(any(coll).([]any)))")
				g.writeLine("for i, item := range any(coll).([]any) {")
				g.indent++
				g.write("res[i] = ")
				g.genExpr(call.Args[0])
				g.writeLine("(item)")
				g.indent--
				g.writeLine("}")
				g.writeLine("return res")
				g.indent--
				g.write("}()")
				return
			}
		case "filter":
			if len(call.Args) == 2 {
				// (filter pred coll)
				g.write("func() []any {\n")
				g.indent++
				g.write("coll := ")
				g.genExpr(call.Args[1])
				g.writeLine("")
				g.writeLine("res := make([]any, 0)")
				g.writeLine("for _, item := range any(coll).([]any) {")
				g.indent++
				g.write("if ")
				g.genExpr(call.Args[0])
				g.writeLine("(item).(bool) {")
				g.indent++
				g.writeLine("res = append(res, item)")
				g.indent--
				g.writeLine("}")
				g.indent--
				g.writeLine("}")
				g.writeLine("return res")
				g.indent--
				g.write("}()")
				return
			}
		case "reduce":
			if len(call.Args) == 3 {
				// (reduce fn init coll)
				g.write("func() any {\n")
				g.indent++
				g.write("coll := ")
				g.genExpr(call.Args[2])
				g.writeLine("")
				g.write("var acc any = ")
				g.genExpr(call.Args[1])
				g.writeLine("")
				g.writeLine("for _, item := range any(coll).([]any) {")
				g.indent++
				g.write("acc = ")
				g.genExpr(call.Args[0])
				g.writeLine("(acc, item)")
				g.indent--
				g.writeLine("}")
				g.writeLine("return acc")
				g.indent--
				g.write("}()")
				return
			}
		}
	}

	// Regular function call
	g.write("")
	if sym, ok := call.Func.(*ast.Symbol); ok {
		name := sanitizeIdent(sym.Name)
		isSpecial := name == "assert" || name == "assert_eq" || name == "assert_true" ||
			name == "assert_false" || name == "assert_err" || name == "for" ||
			name == "go" || name == "chan" || name == "defer" || name == "select" ||
			name == "recur" || name == "loop" || name == "let" || name == "do" ||
			name == "if" || name == "fn" || name == "setb"

		if !isSpecial && !g.pkgNames[name] && !strings.Contains(name, ".") {
			// name = capitalize(name)
		}

		// Check if it looks like a struct construction from another package
		// e.g., sync.WaitGroup
		if strings.Contains(name, ".") {
			parts := strings.Split(name, ".")
			if len(parts) == 2 && len(parts[1]) > 0 && parts[1][0] >= 'A' && parts[1][0] <= 'Z' {
				// It might be a struct construction or a function call.
				// In Elbereth, we use (Type {fields}) for struct construction.
				if len(call.Args) == 1 {
					if _, ok := call.Args[0].(*ast.MapLit); ok {
						g.write(name)
						g.write("{")
						if m, ok := call.Args[0].(*ast.MapLit); ok {
							for i, pair := range m.Pairs {
								if i > 0 {
									g.write(", ")
								}
								if ks, ok := pair.Key.(*ast.KeywordLit); ok {
									g.write(sanitizeIdent(ks.Value))
								} else if ss, ok := pair.Key.(*ast.Symbol); ok {
									g.write(sanitizeIdent(ss.Name))
								} else {
									g.genExpr(pair.Key)
								}
								g.write(": ")
								g.genExpr(pair.Value)
							}
						}
						g.write("}")
						return
					}
				}
			}
		}

		g.write(name)
	} else {
		g.genExpr(call.Func)
	}
	g.write("(")
	for i, arg := range call.Args {
		if i > 0 {
			g.write(", ")
		}
		g.genExprAsValue(arg)
	}
	g.write(")")
}

func (g *Generator) genIf(ifExpr *ast.IfExpr, prefix string) {
	if prefix != "" {
		g.writeLine("{")
		g.indent++
		g.write("if ")
	} else {
		g.write("if ")
	}
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
	} else if prefix != "" {
		g.writeLine("} else {")
		g.indent++
		g.write(prefix)
		g.writeLine("nil")
		g.indent--
	}

	g.write("}")
	if prefix != "" {
		g.writeLine("")
		g.indent--
		g.write("}")
	}
}

func (g *Generator) genDo(doExpr *ast.DoExpr, prefix string) {
	if prefix != "" {
		g.writeLine("{")
		g.indent++
		for i, expr := range doExpr.Exprs {
			if i == len(doExpr.Exprs)-1 {
				g.genExprWithPrefix(expr, prefix)
				g.writeLine("")
			} else {
				g.genExpr(expr)
				g.writeLine("")
			}
		}
		g.indent--
		g.write("}")
	} else {
		g.writeLine("{")
		g.indent++
		for _, expr := range doExpr.Exprs {
			g.genExpr(expr)
			g.writeLine("")
		}
		g.indent--
		g.write("}")
	}
}

func (g *Generator) genLet(letExpr *ast.LetExpr) {
	for _, binding := range letExpr.Bindings {
		prefix := ""
		for i, name := range binding.Names {
			if i > 0 {
				prefix += ", "
			}
			prefix += sanitizeIdent(name)
		}
		g.genExprWithPrefix(binding.Init, prefix+" := ")
		g.writeLine("")
	}
	for _, expr := range letExpr.Body {
		g.genExpr(expr)
		g.writeLine("")
	}
}

func (g *Generator) genDoseq(d *ast.DoseqExpr) {
	g.write("for _, ")
	g.write(sanitizeIdent(d.Var))
	g.write(" := range ")
	g.genExpr(d.Coll)
	g.writeLine(" {")
	g.indent++
	for _, expr := range d.Body {
		g.genExpr(expr)
		g.writeLine("")
	}
	g.indent--
	g.writeLine("}")
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

	if prefix != "" {
		g.writeLine("{")
		g.indent++
	}
	if isTypeSwitch {
		g.write("switch v := any(")
		g.genExpr(m.Val)
		g.writeLine(").(type) {")
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
						if sym, ok := call.Args[0].(*ast.Symbol); ok && sym.Name != "_" {
							g.writeLine(fmt.Sprintf("%s := v.Value", sanitizeIdent(sym.Name)))
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
	if prefix != "" {
		g.writeLine("")
		g.indent--
		g.write("}")
	}
}

func (g *Generator) genExprWithPrefix(expr ast.Expr, prefix string) {
	if _, isRecur := expr.(*ast.RecurExpr); isRecur {
		g.genExpr(expr)
		return
	}

	switch ex := expr.(type) {
	case *ast.IfExpr:
		if prefix == "return " {
			g.genIf(ex, prefix)
		} else {
			g.write(prefix)
			g.genExprAsValue(ex)
		}
	case *ast.MatchExpr:
		if prefix == "return " {
			g.genMatch(ex, prefix)
		} else {
			g.write(prefix)
			g.genExprAsValue(ex)
		}
	case *ast.LoopExpr:
		if prefix == "return " {
			g.genLoop(ex, prefix)
		} else {
			g.write(prefix)
			g.genExprAsValue(ex)
		}
	case *ast.DoExpr:
		if prefix == "return " {
			g.genDo(ex, prefix)
		} else {
			g.write(prefix)
			g.genExpr(ex)
		}
	case *ast.DoseqExpr:
		if prefix == "return " {
			g.genDoseq(ex)
			g.writeLine("return nil")
		} else {
			g.write(prefix)
			g.genDoseq(ex)
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
		g.write(sanitizeIdent(p.Name))
		g.write(" any")
	}
	g.writeLine(") any {")
	g.indent++

	for i, expr := range fn.Body {
		if i == len(fn.Body)-1 {
			g.genExprWithPrefix(expr, "return ")
			g.writeLine("")
		} else {
			g.genExpr(expr)
			g.writeLine("")
		}
	}

	g.indent--
	g.write("}")
}

func (g *Generator) genLoop(l *ast.LoopExpr, prefix string) {
	var names []string
	for _, b := range l.Bindings {
		name := sanitizeIdent(b.Names[0])
		names = append(names, name)
		g.genExprWithPrefix(b.Init, name+" := ")
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
			if c.Binding != "" && c.Binding != "_" {
				g.write(fmt.Sprintf("%s := ", sanitizeIdent(c.Binding)))
			}
			g.write("<-")
			g.genExpr(c.Chan)
			g.writeLine(":")
		}
		g.indent++
		if c.Binding != "" && c.Binding != "_" {
			g.writeLine(fmt.Sprintf("_ = %s", sanitizeIdent(c.Binding)))
		}
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

func (g *Generator) genExprWithNumericAssertion(expr ast.Expr) {
	if t, ok := g.inferredTypes[expr]; ok && t == types.Int {
		g.write("int64(")
		g.genExpr(expr)
		g.write(")")
		return
	}
	if _, ok := expr.(*ast.IntLit); ok {
		g.genExpr(expr)
		return
	}
	if call, ok := expr.(*ast.FuncCall); ok {
		if sym, ok := call.Func.(*ast.Symbol); ok && (sym.Name == "len" || sym.Name == "cap") {
			g.write("int64(")
			g.genExpr(expr)
			g.write(")")
			return
		}
	}
	g.write("any(")
	g.genExpr(expr)
	g.write(").(int64)")
}

func (g *Generator) typeToGoString(t ast.Type) string {
	if t == nil {
		return "any"
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
			return sanitizeIdent(tt.Name)
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
		return "any"
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

func isTypeSymbol(tok lexer.Token) bool {
	if tok.Type != lexer.TokenSymbol && tok.Type != lexer.TokenKeyword {
		return false
	}
	if strings.HasPrefix(tok.Value, "*") {
		return true
	}
	switch tok.Value {
	case "int", "float", "string", "bool", "byte", "rune":
		return true
	}
	// Capitalized names are often types in Go
	if len(tok.Value) > 0 && tok.Value[0] >= 'A' && tok.Value[0] <= 'Z' {
		return true
	}
	return false
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
	if s == "assert" || s == "assert_eq" {
		return s
	}
	return s
}
