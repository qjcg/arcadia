package eval

import (
	"testing"

	"github.com/qjcg/arcadia/x/elbereth/internal/lexer"
	"github.com/qjcg/arcadia/x/elbereth/internal/parser"
)

func TestEvalArithmetic(t *testing.T) {
	ev := New()
	tests := []struct {
		input string
		want  int64
	}{
		{"(+ 1 2 3)", 6},
		{"(- 10 3)", 7},
		{"(* 2 3 4)", 24},
		{"(/ 100 2)", 50},
		{"(% 10 3)", 1},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		prog, err := p.Parse()
		if err != nil {
			t.Fatalf("Parse error for %s: %v", tt.input, err)
		}
		res, err := ev.EvalTop(prog.Items[0])
		if err != nil {
			t.Fatalf("Eval error for %s: %v", tt.input, err)
		}
		if res.(IntVal).Value != tt.want {
			t.Errorf("got %v, want %v", res, tt.want)
		}
	}
}

func TestEvalDef(t *testing.T) {
	ev := New()
	l := lexer.New("(def x 42)")
	p := parser.New(l)
	prog, _ := p.Parse()
	_, _ = ev.EvalTop(prog.Items[0])

	l2 := lexer.New("x")
	p2 := parser.New(l2)
	prog2, _ := p2.Parse()
	res, _ := ev.EvalTop(prog2.Items[0])

	if res.(IntVal).Value != 42 {
		t.Errorf("expected 42, got %v", res)
	}
}

func TestEvalDefn(t *testing.T) {
	ev := New()
	l := lexer.New("(defn double [n] (* n 2))")
	p := parser.New(l)
	prog, _ := p.Parse()
	_, _ = ev.EvalTop(prog.Items[0])

	l2 := lexer.New("(double 21)")
	p2 := parser.New(l2)
	prog2, _ := p2.Parse()
	res, err := ev.EvalTop(prog2.Items[0])
	if err != nil {
		t.Fatalf("Eval error: %v", err)
	}

	if res.(IntVal).Value != 42 {
		t.Errorf("expected 42, got %v", res)
	}
}
