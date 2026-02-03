package repl

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/chzyer/readline"

	"github.com/qjcg/arcadia/x/elbereth/internal/ast"
	"github.com/qjcg/arcadia/x/elbereth/internal/codegen"
	"github.com/qjcg/arcadia/x/elbereth/internal/eval"
	"github.com/qjcg/arcadia/x/elbereth/internal/expander"
	"github.com/qjcg/arcadia/x/elbereth/internal/lexer"
	"github.com/qjcg/arcadia/x/elbereth/internal/parser"
)

type Mode int

const (
	Interpreted Mode = iota
	Compiled
)

type REPL struct {
	input       io.Reader
	output      io.Writer
	mode        Mode
	history     []ast.Node // Accumulation of definitions
	importSpecs []ast.ImportSpec
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
		mode:   Interpreted,
	}
}

func (r *REPL) SetMode(mode Mode) {
	r.mode = mode
}

func (r *REPL) Run() error {
	rl, err := readline.New("elb> ")
	if err != nil {
		return err
	}
	defer rl.Close()

	evaluator := eval.New()
	ex := expander.New()

	fmt.Fprintln(r.output, "Elbereth REPL v0.1.0")
	fmt.Fprintln(r.output, "Commands: (exit), (help), (mode compiled), (mode interpreted)")
	fmt.Fprintln(r.output, "")

	for {
		prompt := "elb> "
		if r.mode == Compiled {
			prompt = "elb[c]> "
		}
		rl.SetPrompt(prompt)

		line, err := rl.Readline()
		if err != nil { // io.EOF or readline.ErrInterrupt
			if err == readline.ErrInterrupt {
				continue
			}
			break
		}

		line = strings.TrimSpace(line)
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

		if line == "(mode compiled)" {
			r.mode = Compiled
			fmt.Fprintln(r.output, "Switched to Compiled mode (experimental)")
			continue
		}

		if line == "(mode interpreted)" {
			r.mode = Interpreted
			fmt.Fprintln(r.output, "Switched to Interpreted mode")
			continue
		}

		if r.mode == Compiled {
			if err := r.evalCompiled(line, ex); err != nil {
				fmt.Fprintf(r.output, "Error: %v\n", err)
			}
			continue
		}

		if err := r.eval(line, evaluator, ex); err != nil {
			fmt.Fprintf(r.output, "Error: %v\n", err)
		}
	}

	return nil
}

func (r *REPL) eval(input string, evaluator *eval.Evaluator, ex *expander.Expander) error {
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

	// Expand macros
	if err := ex.Expand(prog); err != nil {
		return fmt.Errorf("macro expansion error: %w", err)
	}

	// Evaluate and print results
	for _, item := range prog.Items {
		val, err := evaluator.EvalTop(item)
		if err != nil {
			return err
		}
		if val != nil {
			fmt.Fprintln(r.output, val.String())
		}
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

func (r *REPL) evalCompiled(input string, ex *expander.Expander) error {
	lex := lexer.New(input)
	p := parser.New(lex)
	prog, err := p.Parse()
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	if len(prog.Items) == 0 {
		return nil
	}

	// Expand macros
	if err := ex.Expand(prog); err != nil {
		return fmt.Errorf("macro expansion error: %w", err)
	}

	hasExpr := false
	var lastExpr ast.Expr

	for _, item := range prog.Items {
		switch n := item.(type) {
		case *ast.Import, *ast.Def, *ast.Defn, *ast.Deftype, *ast.Defmacro:
			r.history = append(r.history, n)
		case ast.Expr:
			hasExpr = true
			lastExpr = n
		}
	}

	if !hasExpr {
		// If it's just definitions, we might want to check if they compile
		return r.checkCompilation()
	}

	// Create a temp program to evaluate the last expression
	evalProg := &ast.Program{
		Items: append([]ast.Node{}, r.history...),
	}

	// Wrap last expression in println if it's not already
	wrappedExpr := &ast.FuncCall{
		Func: &ast.Symbol{Name: "println"},
		Args: []ast.Expr{lastExpr},
	}

	evalProg.Items = append(evalProg.Items, &ast.Defn{
		Name: "main",
		Body: []ast.Expr{wrappedExpr},
	})

	gen := codegen.New()
	goCode, err := gen.Generate(evalProg)
	if err != nil {
		return fmt.Errorf("codegen error: %w", err)
	}

	// Write to temp file and run
	tmpDir, err := os.MkdirTemp("", "elbereth-repl-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	tmpFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(tmpFile, []byte(goCode), 0o644); err != nil {
		return err
	}

	// Need a go.mod if we want any imports to work reliably?
	// Actually, if it's stdlib it should be fine.
	// For now let's assume stdlib.

	cmd := exec.Command("go", "run", "main.go")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("runtime error: %s\n%w", string(output), err)
	}

	fmt.Fprint(r.output, string(output))
	return nil
}

func (r *REPL) checkCompilation() error {
	evalProg := &ast.Program{
		Items: append([]ast.Node{}, r.history...),
	}
	evalProg.Items = append(evalProg.Items, &ast.Defn{
		Name: "main",
		Body: []ast.Expr{&ast.NilLit{}},
	})

	gen := codegen.New()
	_, err := gen.Generate(evalProg)
	return err
}

func (r *REPL) printHelp() {
	help := `Elbereth REPL Commands:
  (exit)          - Exit the REPL
  (help)          - Show this help message
  (mode compiled) - Switch to compiled mode (supports all Go imports)
  (mode interpreted) - Switch to interpreted mode (faster feedback)

Examples:
  (import "math/rand")
  (rand/Int63)
`
	fmt.Fprint(r.output, help)
}
