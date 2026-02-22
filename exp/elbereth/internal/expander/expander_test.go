package expander

import (
	"testing"

	"github.com/qjcg/arcadia/exp/elbereth/internal/ast"
	"github.com/qjcg/arcadia/exp/elbereth/internal/lexer"
	"github.com/qjcg/arcadia/exp/elbereth/internal/parser"
)

func TestMacroExpansion(t *testing.T) {
	input := `
(defmacro unless [cond & body]
  ` + "`" + `(if (not ,cond)
    (do ,@body)))

(defn test-unless [x]
  (unless (> x 10)
    (println "small")))
`
	l := lexer.New(input)
	p := parser.New(l)
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	ex := New()
	err = ex.Expand(prog)
	if err != nil {
		t.Fatalf("Expand error: %v", err)
	}

	// Find test-unless function
	var testUnless *ast.Defn
	for _, item := range prog.Items {
		if d, ok := item.(*ast.Defn); ok && d.Name == "test-unless" {
			testUnless = d
			break
		}
	}

	if testUnless == nil {
		t.Fatal("test-unless not found")
	}

	// Check if body contains 'if' instead of 'unless'
	if len(testUnless.Body) != 1 {
		t.Fatalf("expected 1 expr in body, got %d", len(testUnless.Body))
	}

	// The expander should have turned the macro call into an IfExpr
	ifCall, ok := testUnless.Body[0].(*ast.IfExpr)
	if !ok {
		// If it's still a FuncCall, check why
		if fc, ok := testUnless.Body[0].(*ast.FuncCall); ok {
			t.Fatalf("expected *ast.IfExpr, got *ast.FuncCall for func %v", fc.Func)
		}
		t.Fatalf("expected *ast.IfExpr, got %T", testUnless.Body[0])
	}

	// Check condition is (not (> x 10))
	notCall, ok := ifCall.Cond.(*ast.FuncCall)
	if !ok {
		t.Fatalf("expected *ast.FuncCall for cond, got %T", ifCall.Cond)
	}
	if sym, ok := notCall.Func.(*ast.Symbol); !ok || sym.Name != "not" {
		t.Errorf("expected 'not' call, got %v", notCall.Func)
	}
}
