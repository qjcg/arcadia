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
}

// New creates a new code generator
func New() *Generator {
	return &Generator{
		symbols:   make(map[string]types.Type),
		functions: make(map[string]*ast.Defn),
		structs:   make(map[string]*ast.Deftype),
	}
}

// Generate generates Go code from an AST program
func (g *Generator) Generate(prog *ast.Program) (string, error) {
	// Write package declaration
	g.writeLine("package main")
	g.writeLine("")
	g.writeLine(`import "fmt"`)
	g.writeLine("")

	// First pass: collect definitions
	for _, item := range prog.Items {
		switch n := item.(type) {
		case *ast.Defn:
			g.functions[n.Name] = n
		case *ast.Deftype:
			g.structs[n.Name] = n
		case *ast.Def:
			if n.Type != nil {
				if nt, ok := n.Type.(*ast.NamedType); ok {
					g.symbols[n.Name] = types.ParseTypeString(nt.Name)
				}
			}
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
		g.write(p.Name)
		g.write(" ")
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
		g.write(fmt.Sprintf("%d", e.Value))

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
		// Capitalize for Go export rules, except for main
		if name != "main" {
			name = capitalize(name)
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
		g.genIf(e)

	case *ast.DoExpr:
		g.genDo(e)

	case *ast.LetExpr:
		g.genLet(e)

	case *ast.FuncLit:
		g.genFuncLit(e)

	case *ast.QuoteExpr:
		// Quoted expressions are just their values
		g.genExpr(e.Expr)

	default:
		g.write("/* unknown expr */")
	}
}

func (g *Generator) genFuncCall(call *ast.FuncCall) {
	if sym, ok := call.Func.(*ast.Symbol); ok {
		switch sym.Name {
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

func (g *Generator) genIf(ifExpr *ast.IfExpr) {
	g.write("if ")
	g.genExpr(ifExpr.Cond)
	g.writeLine(" {")
	g.indent++
	g.genExpr(ifExpr.Then)
	g.writeLine("")
	g.indent--

	if ifExpr.Else != nil {
		g.writeLine("} else {")
		g.indent++
		g.genExpr(ifExpr.Else)
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
		g.write(binding.Name)
		g.write(" := ")
		g.genExpr(binding.Init)
		g.writeLine("")
	}
	for _, expr := range letExpr.Body {
		g.genExpr(expr)
		g.writeLine("")
	}
}

func (g *Generator) genFuncLit(fn *ast.FuncLit) {
	g.write("func(")
	for i, p := range fn.Params {
		if i > 0 {
			g.write(", ")
		}
		g.write(p.Name)
		g.write(" interface{}")
	}
	g.writeLine(") interface{} {")
	g.indent++

	for _, expr := range fn.Body {
		g.genExpr(expr)
		g.writeLine("")
	}

	g.indent--
	g.write("}")
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
	return s
}
