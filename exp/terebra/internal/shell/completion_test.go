package shell

import (
	"strings"
	"testing"
)

func TestEvalArithmetic(t *testing.T) {
	s := newTestShell()
	cases := map[string]int{
		"":        0,
		"5":       5,
		"2+3":     5,
		"10-4":    6,
		"2*3":     6,
		"10/2":    5,
		"10%3":    1,
		"2+3*4":   14,
		"(2+3)*4": 20,
		"-5":      -5,
		"+5":      5,
		" 7 ":     7,
	}
	for in, want := range cases {
		if got := s.evalArithmetic(in); got != want {
			t.Errorf("evalArithmetic(%q) = %d, want %d", in, got, want)
		}
	}
}

func TestEvalArithmeticDivisionByZero(t *testing.T) {
	s := newTestShell()
	if got := s.evalArithmetic("5/0"); got != 0 {
		t.Fatalf("expected 0 for division by zero, got %d", got)
	}
}

func TestEvalArithmeticVariable(t *testing.T) {
	s := newTestShell()
	s.vars["x"] = "10"
	if got := s.evalArithmetic("x+5"); got != 15 {
		t.Fatalf("expected 15, got %d", got)
	}
}

func TestEvalFactorUnary(t *testing.T) {
	s := newTestShell()
	val, rest := s.evalFactor("-5")
	if val != -5 || rest != "" {
		t.Fatalf("expected (-5, ''), got (%d, %q)", val, rest)
	}
	val, _ = s.evalFactor("+5")
	if val != 5 {
		t.Fatalf("expected 5, got %d", val)
	}
}

func TestEvalFactorEmpty(t *testing.T) {
	s := newTestShell()
	val, rest := s.evalFactor("")
	if val != 0 || rest != "" {
		t.Fatalf("expected (0, ''), got (%d, %q)", val, rest)
	}
}

func TestExpandStringOpSubstring(t *testing.T) {
	s := newTestShell()
	s.vars["v"] = "hello world"
	var b strings.Builder
	if !s.expandStringOp("v:6", &b) {
		t.Fatal("expected substring op handled")
	}
	if b.String() != "world" {
		t.Fatalf("expected 'world', got %q", b.String())
	}
}

func TestExpandStringOpSubstringOffsetLength(t *testing.T) {
	s := newTestShell()
	s.vars["v"] = "hello world"
	var b strings.Builder
	s.expandStringOp("v:0:5", &b)
	if b.String() != "hello" {
		t.Fatalf("expected 'hello', got %q", b.String())
	}
}

func TestExpandStringOpSubstringNegativeOffset(t *testing.T) {
	s := newTestShell()
	s.vars["v"] = "hello"
	var b strings.Builder
	s.expandStringOp("v:-3", &b)
	if b.String() != "llo" {
		t.Fatalf("expected 'llo', got %q", b.String())
	}
}

func TestExpandStringOpSubstringOutOfRange(t *testing.T) {
	s := newTestShell()
	s.vars["v"] = "hi"
	var b strings.Builder
	s.expandStringOp("v:10", &b)
	if b.String() != "" {
		t.Fatalf("expected empty, got %q", b.String())
	}
}

func TestExpandStringOpRemovePrefix(t *testing.T) {
	s := newTestShell()
	s.vars["v"] = "foobar"
	var b strings.Builder
	s.expandStringOp("v#foo", &b)
	if b.String() != "bar" {
		t.Fatalf("expected 'bar', got %q", b.String())
	}
}

func TestExpandStringOpRemoveLongestPrefix(t *testing.T) {
	s := newTestShell()
	s.vars["v"] = "aab"
	var b strings.Builder
	s.expandStringOp("v##a", &b)
	if b.String() != "b" {
		t.Fatalf("expected 'b', got %q", b.String())
	}
}

func TestExpandStringOpRemoveSuffix(t *testing.T) {
	s := newTestShell()
	s.vars["v"] = "foobar"
	var b strings.Builder
	s.expandStringOp("v%bar", &b)
	if b.String() != "foo" {
		t.Fatalf("expected 'foo', got %q", b.String())
	}
}

func TestExpandStringOpRemoveLongestSuffix(t *testing.T) {
	s := newTestShell()
	s.vars["v"] = "abbb"
	var b strings.Builder
	s.expandStringOp("v%%b", &b)
	if b.String() != "a" {
		t.Fatalf("expected 'a', got %q", b.String())
	}
}

func TestExpandStringOpReplaceFirst(t *testing.T) {
	s := newTestShell()
	s.vars["v"] = "a-b-a"
	var b strings.Builder
	s.expandStringOp("v/a/x", &b)
	if b.String() != "x-b-a" {
		t.Fatalf("expected 'x-b-a', got %q", b.String())
	}
}

func TestExpandStringOpReplaceAll(t *testing.T) {
	s := newTestShell()
	s.vars["v"] = "a-b-a"
	var b strings.Builder
	s.expandStringOp("v//a/x", &b)
	if b.String() != "x-b-x" {
		t.Fatalf("expected 'x-b-x', got %q", b.String())
	}
}

func TestExpandStringOpUppercase(t *testing.T) {
	s := newTestShell()
	s.vars["v"] = "hello"
	var b strings.Builder
	s.expandStringOp("v^", &b)
	if b.String() != "Hello" {
		t.Fatalf("expected 'Hello', got %q", b.String())
	}
	b.Reset()
	s.expandStringOp("v^^", &b)
	if b.String() != "HELLO" {
		t.Fatalf("expected 'HELLO', got %q", b.String())
	}
}

func TestExpandStringOpLowercase(t *testing.T) {
	s := newTestShell()
	s.vars["v"] = "HELLO"
	var b strings.Builder
	s.expandStringOp("v,", &b)
	if b.String() != "hELLO" {
		t.Fatalf("expected 'hELLO', got %q", b.String())
	}
	b.Reset()
	s.expandStringOp("v,,", &b)
	if b.String() != "hello" {
		t.Fatalf("expected 'hello', got %q", b.String())
	}
}

func TestExpandStringOpUnhandled(t *testing.T) {
	s := newTestShell()
	var b strings.Builder
	if s.expandStringOp("plain", &b) {
		t.Fatal("expected unhandled for plain name")
	}
}

func TestExpandBracedRegularVar(t *testing.T) {
	s := newTestShell()
	s.vars["name"] = "world"
	var b strings.Builder
	i := s.expandBraced("${name}", 1, &b)
	if b.String() != "world" {
		t.Fatalf("expected 'world', got %q", b.String())
	}
	if i != len("${name}") {
		t.Fatalf("expected index %d, got %d", len("${name}"), i)
	}
}

func TestExpandBracedLength(t *testing.T) {
	s := newTestShell()
	s.vars["name"] = "hello"
	var b strings.Builder
	s.expandBraced("${#name}", 1, &b)
	if b.String() != "5" {
		t.Fatalf("expected '5', got %q", b.String())
	}
}

func TestExpandBracedArrayAccess(t *testing.T) {
	s := newTestShell()
	s.setArray("arr", []string{"a", "b"})
	var b strings.Builder
	s.expandBraced("${arr[1]}", 1, &b)
	if b.String() != "b" {
		t.Fatalf("expected 'b', got %q", b.String())
	}
}

func TestExpandBracedArrayAll(t *testing.T) {
	s := newTestShell()
	s.setArray("arr", []string{"a", "b"})
	var b strings.Builder
	s.expandBraced("${arr[@]}", 1, &b)
	if b.String() != "a b" {
		t.Fatalf("expected 'a b', got %q", b.String())
	}
}

func TestExpandBracedStringOp(t *testing.T) {
	s := newTestShell()
	s.vars["v"] = "hello"
	var b strings.Builder
	s.expandBraced("${v^}", 1, &b)
	if b.String() != "Hello" {
		t.Fatalf("expected 'Hello', got %q", b.String())
	}
}

func TestExpandVarsSimple(t *testing.T) {
	s := newTestShell()
	s.vars["x"] = "val"
	if got := s.expandVars("$x"); got != "val" {
		t.Fatalf("expected 'val', got %q", got)
	}
}

func TestExpandVarsNoDollar(t *testing.T) {
	s := newTestShell()
	if got := s.expandVars("plain"); got != "plain" {
		t.Fatalf("expected 'plain', got %q", got)
	}
}

func TestExpandVarsPid(t *testing.T) {
	s := newTestShell()
	got := s.expandVars("$$")
	if got == "$$" {
		t.Fatal("expected $$ to expand to pid")
	}
}

func TestExpandVarsExitCode(t *testing.T) {
	s := newTestShell()
	s.setExitCode(3)
	if got := s.expandVars("$?"); got != "3" {
		t.Fatalf("expected '3', got %q", got)
	}
}

func TestExpandVarsArithmetic(t *testing.T) {
	s := newTestShell()
	if got := s.expandVars("$((2+3))"); got != "5" {
		t.Fatalf("expected '5', got %q", got)
	}
}

func TestExpandVarsArrayNoBrace(t *testing.T) {
	s := newTestShell()
	s.setArray("arr", []string{"x", "y"})
	if got := s.expandVars("$arr[1]"); got != "y" {
		t.Fatalf("expected 'y', got %q", got)
	}
}

func TestTryArrayAssignmentElem(t *testing.T) {
	s := newTestShell()
	if !s.tryArrayAssignment("arr[0]=val", nil) {
		t.Fatal("expected array elem assignment handled")
	}
	if got := s.getArrayVar("arr", "0"); got != "val" {
		t.Fatalf("expected 'val', got %q", got)
	}
}

func TestTryArrayAssignmentParen(t *testing.T) {
	s := newTestShell()
	if !s.tryArrayAssignment("arr=(1", []string{"2", "3)"}) {
		t.Fatal("expected array assignment handled")
	}
	if got := s.getArrayVar("arr", "@"); got != "1 2 3" {
		t.Fatalf("expected '1 2 3', got %q", got)
	}
}

func TestTryArrayAssignmentSpaced(t *testing.T) {
	s := newTestShell()
	if !s.tryArrayAssignment("arr", []string{"=(", "1", "2)"}) {
		t.Fatal("expected spaced array assignment handled")
	}
	if got := s.getArrayVar("arr", "@"); got != "1 2" {
		t.Fatalf("expected '1 2', got %q", got)
	}
}

func TestTryArrayAssignmentAssoc(t *testing.T) {
	s := newTestShell()
	if !s.tryArrayAssignment("m=([k1]=v1", []string{"[k2]=v2)"}) {
		t.Fatal("expected assoc assignment handled")
	}
	if got := s.getArrayVar("m", "k1"); got != "v1" {
		t.Fatalf("expected 'v1', got %q", got)
	}
}

func TestTryArrayAssignmentNotIdent(t *testing.T) {
	s := newTestShell()
	if s.tryArrayAssignment("1bad", []string{"x"}) {
		t.Fatal("expected false for non-ident name")
	}
}

func TestIsIdent(t *testing.T) {
	cases := map[string]bool{
		"foo":  true,
		"_x":   true,
		"foo1": true,
		"1foo": false,
		"":     false,
		"a-b":  false,
	}
	for in, want := range cases {
		if got := isIdent(in); got != want {
			t.Errorf("isIdent(%q) = %v, want %v", in, got, want)
		}
	}
}
