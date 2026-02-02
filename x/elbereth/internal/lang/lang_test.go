package lang_test

import (
	"testing"

	"github.com/qjcg/arcadia/x/elbereth/internal/ast"
	"github.com/qjcg/arcadia/x/elbereth/internal/lang"
	_ "github.com/qjcg/arcadia/x/elbereth/internal/lang/minimal"
)

func TestRegistry(t *testing.T) {
	l, err := lang.Get("minimal")
	if err != nil {
		t.Fatalf("failed to get minimal language: %v", err)
	}

	prog, err := l.Parse("#lang minimal\nHello")
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if prog.Lang != "minimal" {
		t.Errorf("expected Lang minimal, got %s", prog.Lang)
	}

	if len(prog.Items) != 1 {
		t.Errorf("expected 1 item (main function), got %d", len(prog.Items))
	}
}

type mockLang struct {
	parsed bool
}

func (m *mockLang) Parse(input string) (*ast.Program, error) {
	m.parsed = true
	return &ast.Program{Lang: "mock"}, nil
}

func (m *mockLang) Expand(prog *ast.Program) error {
	return nil
}

func TestCustomRegistration(t *testing.T) {
	m := &mockLang{}
	lang.Register("mock", m)

	l, err := lang.Get("mock")
	if err != nil {
		t.Fatalf("failed to get mock language: %v", err)
	}

	_, err = l.Parse("anything")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if !m.parsed {
		t.Error("mock language Parse was not called")
	}
}
