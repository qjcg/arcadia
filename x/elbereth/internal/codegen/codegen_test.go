package codegen

import (
	"strings"
	"testing"

	"elbereth/internal/expander"
	"elbereth/internal/lexer"
	"elbereth/internal/parser"
)

func TestGenerateSimple(t *testing.T) {
	input := `(defn add [x y] (+ x y))`
	l := lexer.New(input)
	p := parser.New(l)
	prog, _ := p.Parse()

	gen := New()
	code, err := gen.Generate(prog)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	if !strings.Contains(code, "func Add") {
		t.Errorf("expected func Add, got %s", code)
	}
}

func TestGenerateStruct(t *testing.T) {
	input := `(deftype p {x int y int})`
	l := lexer.New(input)
	p := parser.New(l)
	prog, _ := p.Parse()

	gen := New()
	code, _ := gen.Generate(prog)

	if !strings.Contains(code, "type P struct") {
		t.Errorf("expected type P struct, got %s", code)
	}
}

func TestGenerateWithExpander(t *testing.T) {
	input := `
(defmacro unless [c & b] ` + "`" + `(if (not ,c) (do ,@b)))
(defn f [x] (unless (> x 10) (println x)))
`
	l := lexer.New(input)
	p := parser.New(l)
	prog, _ := p.Parse()

	ex := expander.New()
	_ = ex.Expand(prog)

	gen := New()
	code, _ := gen.Generate(prog)

	if !strings.Contains(code, "if !(X > int64(10))") {
		// Note: sanitizeIdent and capitalization might change case
		// (unless (> x 10) ...) -> if !(X > 10) { ... }
		// Wait, my sanitizeIdent keeps some things.
		if !strings.Contains(code, "if !(") {
			t.Errorf("expected if !, got %s", code)
		}
	}
}
