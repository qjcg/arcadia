package repl

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"elbereth/internal/ast"
	"elbereth/internal/codegen"
	"elbereth/internal/lexer"
	"elbereth/internal/parser"
)

type REPL struct {
	input  io.Reader
	output io.Writer
}

func New(input io.Reader, output io.Writer) *REPL {
	if input == nil {
		input = os.Stdin
	}
	if output == nil {
		output = os.Stdout
	}
	return &REPL{
		input:  input,
		output: output,
	}
}

func (r *REPL) Run() error {
	scanner := bufio.NewScanner(r.input)
	gen := codegen.New()
	context := &evalContext{
		gen: gen,
	}

	fmt.Fprintln(r.output, "Elbereth REPL v0.1.0")
	fmt.Fprintln(r.output, "Type (exit) to quit, (help) for help")
	fmt.Fprintln(r.output, "")

	for {
		fmt.Fprint(r.output, "elb> ")
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if line == "(exit)" {
			fmt.Fprintln(r.output, "Goodbye!")
			break
		}

		if line == "(help)" {
			r.printHelp()
			continue
		}

		if err := r.eval(line, context); err != nil {
			fmt.Fprintf(r.output, "Error: %v\n", err)
		}
	}

	return scanner.Err()
}

func (r *REPL) eval(input string, _ *evalContext) error {
	// Lex and parse the input
	lex := lexer.New(input)
	p := parser.New(lex)
	prog, err := p.Parse()
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	if len(prog.Items) == 0 {
		return nil
	}

	// For now, just show AST representation
	for _, item := range prog.Items {
		r.printNode(item)
	}

	return nil
}

func (r *REPL) printNode(node ast.Node) {
	switch n := node.(type) {
	case *ast.IntLit:
		fmt.Fprintf(r.output, "%d\n", n.Value)
	case *ast.FloatLit:
		fmt.Fprintf(r.output, "%f\n", n.Value)
	case *ast.StringLit:
		fmt.Fprintf(r.output, "\"%s\"\n", n.Value)
	case *ast.BoolLit:
		fmt.Fprintf(r.output, "%v\n", n.Value)
	case *ast.NilLit:
		fmt.Fprintln(r.output, "nil")
	case *ast.KeywordLit:
		fmt.Fprintf(r.output, ":%s\n", n.Value)
	case *ast.Symbol:
		fmt.Fprintf(r.output, "%s\n", n.Name)
	case *ast.VectorLit:
		fmt.Fprintf(r.output, "%v\n", r.vectorToString(n))
	case *ast.MapLit:
		fmt.Fprintf(r.output, "%v\n", r.mapToString(n))
	case *ast.Defn:
		fmt.Fprintf(r.output, "#<function %s>\n", n.Name)
	case *ast.Def:
		fmt.Fprintf(r.output, "#<var %s>\n", n.Name)
	case *ast.FuncCall:
		fmt.Fprintf(r.output, "%v\n", n)
	case *ast.FuncLit:
		fmt.Fprintf(r.output, "#<lambda>\n")
	default:
		fmt.Fprintf(r.output, "%v\n", n)
	}
}

func (r *REPL) vectorToString(vec *ast.VectorLit) string {
	var parts []string
	for _, item := range vec.Elts {
		parts = append(parts, r.exprToString(item))
	}
	return "[" + strings.Join(parts, " ") + "]"
}

func (r *REPL) mapToString(m *ast.MapLit) string {
	var parts []string
	for _, pair := range m.Pairs {
		parts = append(parts, r.exprToString(pair.Key)+" "+r.exprToString(pair.Value))
	}
	return "{" + strings.Join(parts, " ") + "}"
}

func (r *REPL) exprToString(node ast.Expr) string {
	switch n := node.(type) {
	case *ast.IntLit:
		return fmt.Sprintf("%d", n.Value)
	case *ast.FloatLit:
		return fmt.Sprintf("%f", n.Value)
	case *ast.StringLit:
		return fmt.Sprintf("\"%s\"", n.Value)
	case *ast.BoolLit:
		return fmt.Sprintf("%v", n.Value)
	case *ast.NilLit:
		return "nil"
	case *ast.KeywordLit:
		return fmt.Sprintf(":%s", n.Value)
	case *ast.Symbol:
		return n.Name
	case *ast.VectorLit:
		return r.vectorToString(n)
	case *ast.MapLit:
		return r.mapToString(n)
	case *ast.FuncCall:
		return n.String()
	case *ast.FuncLit:
		return "#<lambda>"
	default:
		return fmt.Sprintf("%v", n)
	}
}

func (r *REPL) printHelp() {
	help := `Elbereth REPL Commands:
  (exit)          - Exit the REPL
  (help)          - Show this help message

Examples:
  (+ 1 2)         - Arithmetic
  (defn f [x] x)  - Define a function
  (def x 42)      - Define a variable

For more information, see the documentation.
`
	fmt.Fprint(r.output, help)
}

type evalContext struct {
	gen *codegen.Generator
}
