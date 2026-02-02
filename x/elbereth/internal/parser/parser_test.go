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

func TestParsePackage(t *testing.T) {
	input := `(package myutils)`
	l := lexer.New(input)
	p := New(l)
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if prog.Package != "myutils" {
		t.Errorf("expected program package 'myutils', got %q", prog.Package)
	}

	if len(prog.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(prog.Items))
	}

	pkg, ok := prog.Items[0].(*ast.Package)
	if !ok {
		t.Fatalf("expected *ast.Package, got %T", prog.Items[0])
	}

	if pkg.Name != "myutils" {
		t.Errorf("expected package name 'myutils', got %q", pkg.Name)
	}
}

func TestParseImport(t *testing.T) {
	input := `(import "fmt" "math" [time "time"])`
	l := lexer.New(input)
	p := New(l)
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if len(prog.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(prog.Items))
	}

	imp, ok := prog.Items[0].(*ast.Import)
	if !ok {
		t.Fatalf("expected *ast.Import, got %T", prog.Items[0])
	}

	if len(imp.Specs) != 3 {
		t.Fatalf("expected 3 specs, got %d", len(imp.Specs))
	}

	if imp.Specs[0].Path != "fmt" {
		t.Errorf("expected spec 0 path 'fmt', got %q", imp.Specs[0].Path)
	}

	if imp.Specs[1].Path != "math" {
		t.Errorf("expected spec 1 path 'math', got %q", imp.Specs[1].Path)
	}

	if imp.Specs[2].Alias != "time" || imp.Specs[2].Path != "time" {
		t.Errorf("expected spec 2 [time \"time\"], got [%s %q]", imp.Specs[2].Alias, imp.Specs[2].Path)
	}
}
