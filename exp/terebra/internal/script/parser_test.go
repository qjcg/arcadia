package script

import (
	"testing"
)

func TestParseSimpleCommand(t *testing.T) {
	script, err := Parse("echo hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(script.Stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(script.Stmts))
	}
	_, ok := script.Stmts[0].(*CommandStmt)
	if !ok {
		t.Fatalf("expected CommandStmt, got %T", script.Stmts[0])
	}
}

func TestParseMultipleCommands(t *testing.T) {
	script, err := Parse("echo hello\nls -la\npwd")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(script.Stmts) != 3 {
		t.Fatalf("expected 3 statements, got %d", len(script.Stmts))
	}
}

func TestParseIf(t *testing.T) {
	input := "if true\nthen\n  echo yes\nfi"
	script, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(script.Stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(script.Stmts))
	}
	ifStmt, ok := script.Stmts[0].(*IfStmt)
	if !ok {
		t.Fatalf("expected IfStmt, got %T", script.Stmts[0])
	}
	if len(ifStmt.Then) != 1 {
		t.Fatalf("expected 1 statement in then body, got %d", len(ifStmt.Then))
	}
}

func TestParseIfElse(t *testing.T) {
	input := "if false\nthen\n  echo yes\nelse\n  echo no\nfi"
	script, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ifStmt, ok := script.Stmts[0].(*IfStmt)
	if !ok {
		t.Fatalf("expected IfStmt")
	}
	if len(ifStmt.Else) != 1 {
		t.Fatalf("expected 1 statement in else body, got %d", len(ifStmt.Else))
	}
}

func TestParseIfElif(t *testing.T) {
	input := "if false\nthen\n  echo a\nelif true\nthen\n  echo b\nelse\n  echo c\nfi"
	script, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ifStmt, ok := script.Stmts[0].(*IfStmt)
	if !ok {
		t.Fatalf("expected IfStmt")
	}
	if len(ifStmt.ElseIf) != 1 {
		t.Fatalf("expected 1 elif, got %d", len(ifStmt.ElseIf))
	}
	if len(ifStmt.Else) != 1 {
		t.Fatalf("expected 1 else, got %d", len(ifStmt.Else))
	}
}

func TestParseFor(t *testing.T) {
	input := "for i in 1 2 3\ndo\n  echo $i\ndone"
	script, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	forStmt, ok := script.Stmts[0].(*ForStmt)
	if !ok {
		t.Fatalf("expected ForStmt, got %T", script.Stmts[0])
	}
	if forStmt.Var != "i" {
		t.Errorf("expected var 'i', got %q", forStmt.Var)
	}
	if len(forStmt.Words) != 3 {
		t.Errorf("expected 3 words, got %d", len(forStmt.Words))
	}
	if len(forStmt.Body) != 1 {
		t.Errorf("expected 1 body statement, got %d", len(forStmt.Body))
	}
}

func TestParseWhile(t *testing.T) {
	input := "while true\ndo\n  echo looping\ndone"
	script, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, ok := script.Stmts[0].(*WhileStmt)
	if !ok {
		t.Fatalf("expected WhileStmt")
	}
}

func TestParseUntil(t *testing.T) {
	input := "until false\ndo\n  echo running\ndone"
	script, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, ok := script.Stmts[0].(*UntilStmt)
	if !ok {
		t.Fatalf("expected UntilStmt")
	}
}

func TestParseFuncDef(t *testing.T) {
	input := "function hello {\n  echo hi\n}"
	script, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fd, ok := script.Stmts[0].(*FuncDef)
	if !ok {
		t.Fatalf("expected FuncDef, got %T", script.Stmts[0])
	}
	if fd.Name != "hello" {
		t.Errorf("expected name 'hello', got %q", fd.Name)
	}
	if len(fd.Body) != 1 {
		t.Errorf("expected 1 body statement, got %d", len(fd.Body))
	}
}

func TestParseFuncDefAlt(t *testing.T) {
	input := "hello() {\n  echo hi\n}"
	script, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fd, ok := script.Stmts[0].(*FuncDef)
	if !ok {
		t.Fatalf("expected FuncDef, got %T", script.Stmts[0])
	}
	if fd.Name != "hello" {
		t.Errorf("expected name 'hello', got %q", fd.Name)
	}
}

func TestParseNestedIf(t *testing.T) {
	input := "if true\nthen\n  if false\n  then\n    echo nested\n  fi\nfi"
	script, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(script.Stmts) != 1 {
		t.Fatalf("expected 1 statement")
	}
	outer, ok := script.Stmts[0].(*IfStmt)
	if !ok {
		t.Fatalf("expected IfStmt")
	}
	if len(outer.Then) != 1 {
		t.Fatalf("expected 1 then statement")
	}
	_, ok = outer.Then[0].(*IfStmt)
	if !ok {
		t.Fatalf("expected nested IfStmt")
	}
}

func TestParseEmpty(t *testing.T) {
	script, err := Parse("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(script.Stmts) != 0 {
		t.Fatalf("expected 0 statements, got %d", len(script.Stmts))
	}
}

func TestParseComments(t *testing.T) {
	script, err := Parse("# comment\necho hi\n# another comment")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(script.Stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(script.Stmts))
	}
	if _, ok := script.Stmts[0].(*CommandStmt); !ok {
		t.Fatalf("expected CommandStmt")
	}
}
