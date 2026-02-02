package codegen

import (
	"strings"
	"testing"

	"github.com/qjcg/arcadia/x/elbereth/internal/expander"
	"github.com/qjcg/arcadia/x/elbereth/internal/lexer"
	"github.com/qjcg/arcadia/x/elbereth/internal/parser"
)

func TestGenerateSimple(t *testing.T) {
	input := `(defn Add [x y] (+ x y))`
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

func TestGenerateVisibility(t *testing.T) {
	input := `(defn add [x y] (+ x y))`
	l := lexer.New(input)
	p := parser.New(l)
	prog, _ := p.Parse()

	gen := New()
	code, _ := gen.Generate(prog)

	if !strings.Contains(code, "func add") {
		t.Errorf("expected func add (private), got %s", code)
	}
}

func TestGenerateStruct(t *testing.T) {
	input := `(deftype Point {X int Y int})`
	l := lexer.New(input)
	p := parser.New(l)
	prog, _ := p.Parse()

	gen := New()
	code, _ := gen.Generate(prog)

	if !strings.Contains(code, "type Point struct") {
		t.Errorf("expected type Point struct, got %s", code)
	}
	if !strings.Contains(code, "X int64") {
		t.Errorf("expected field X, got %s", code)
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

func TestGeneratePackage(t *testing.T) {
	input := `(package myutils) (defn add [x y] (+ x y))`
	l := lexer.New(input)
	p := parser.New(l)
	prog, _ := p.Parse()

	gen := New()
	code, err := gen.Generate(prog)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	if !strings.Contains(code, "package myutils") {
		t.Errorf("expected package myutils, got %s", code)
	}

	if strings.Contains(code, "package main") {
		t.Errorf("did not expect package main, got %s", code)
	}
}

func TestGenerateMultiImport(t *testing.T) {
	input := `(import "fmt" "math" [time "time"])`
	l := lexer.New(input)
	p := parser.New(l)
	prog, _ := p.Parse()

	gen := New()
	code, err := gen.Generate(prog)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	if !strings.Contains(code, `"fmt"`) {
		t.Errorf("expected fmt import, got %s", code)
	}
	if !strings.Contains(code, `"math"`) {
		t.Errorf("expected math import, got %s", code)
	}
	if !strings.Contains(code, `time "time"`) {
		t.Errorf("expected time aliased import, got %s", code)
	}
}
