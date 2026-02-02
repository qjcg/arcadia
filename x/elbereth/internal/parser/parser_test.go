package parser

import (
	"testing"

	"github.com/qjcg/arcadia/x/elbereth/internal/ast"
	"github.com/qjcg/arcadia/x/elbereth/internal/lexer"
)

func TestParseDefn(t *testing.T) {
	input := `(defn add [x y] (+ x y))`
	l := lexer.New(input)
	p := New(l)
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if len(prog.Items) != 1 {
		t.Fatalf("expected 1 top-level item, got %d", len(prog.Items))
	}

	defn, ok := prog.Items[0].(*ast.Defn)
	if !ok {
		t.Fatalf("expected *ast.Defn, got %T", prog.Items[0])
	}

	if defn.Name != "add" {
		t.Errorf("expected name 'add', got %q", defn.Name)
	}

	if len(defn.Params) != 2 {
		t.Errorf("expected 2 params, got %d", len(defn.Params))
	}

	if len(defn.Body) != 1 {
		t.Errorf("expected body length 1, got %d", len(defn.Body))
	}
}

func TestParseDeftype(t *testing.T) {
	input := `(deftype Person {name string age int})`
	l := lexer.New(input)
	p := New(l)
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	deftype, ok := prog.Items[0].(*ast.Deftype)
	if !ok {
		t.Fatalf("expected *ast.Deftype, got %T", prog.Items[0])
	}

	if deftype.Name != "Person" {
		t.Errorf("expected name 'Person', got %q", deftype.Name)
	}

	if len(deftype.Fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(deftype.Fields))
	}
}

func TestParseSumType(t *testing.T) {
	input := `(deftype Result
  (:ok int)
  (:err string))`
	l := lexer.New(input)
	p := New(l)
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	deftype, ok := prog.Items[0].(*ast.Deftype)
	if !ok {
		t.Fatalf("expected *ast.Deftype, got %T", prog.Items[0])
	}

	if len(deftype.Variants) != 2 {
		t.Errorf("expected 2 variants, got %d", len(deftype.Variants))
	}
}
