package parser

import (
	"strings"
	"testing"
)

func TestRedirectTypeString(t *testing.T) {
	cases := map[RedirectType]string{
		RedirectStdout:         ">",
		RedirectStderr:         "2>",
		RedirectAppend:         ">>",
		RedirectStderrAppend:   "2>>",
		RedirectStdin:          "<",
		RedirectStderrToStdout: "2>&1",
		RedirectBoth:           "&>",
		RedirectBothAppend:     "&>>",
		RedirectHereString:     "<<<",
		RedirectHeredoc:        "<<",
		RedirectHeredocDash:    "<<-",
	}
	for rt, want := range cases {
		if got := rt.String(); got != want {
			t.Errorf("RedirectType(%d).String() = %q, want %q", rt, got, want)
		}
	}
	// Unknown type
	if got := RedirectType(999).String(); got != "RedirectType(999)" {
		t.Errorf("expected 'RedirectType(999)', got %q", got)
	}
}

func TestParseScriptSimple(t *testing.T) {
	l := NewLexer("echo hi")
	script, err := parseScript(l)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(script.Pipelines) != 1 {
		t.Fatalf("expected 1 pipeline, got %d", len(script.Pipelines))
	}
}

func TestParseScriptChaining(t *testing.T) {
	l := NewLexer("a && b || c")
	script, err := parseScript(l)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(script.Pipelines) != 3 {
		t.Fatalf("expected 3 pipelines, got %d", len(script.Pipelines))
	}
	if len(script.Ops) != 2 {
		t.Fatalf("expected 2 ops, got %d", len(script.Ops))
	}
	if script.Ops[0] != ChainingAnd || script.Ops[1] != ChainingOr {
		t.Fatalf("expected [And, Or], got %v", script.Ops)
	}
}

func TestParseScriptUnexpectedToken(t *testing.T) {
	l := NewLexer("a |")
	_, err := parseScript(l)
	if err == nil {
		t.Fatal("expected error for dangling pipe")
	}
}

func TestReadBalanced(t *testing.T) {
	l := NewLexer("$(echo hi) rest")
	var word strings.Builder
	var quoted []bool
	l.readBalanced(&word, &quoted, "$(", 1)
	if word.String() != "$(echo hi)" {
		t.Fatalf("expected '$(echo hi)', got %q", word.String())
	}
}

func TestReadBalancedNested(t *testing.T) {
	l := NewLexer("$(a $(b)) rest")
	var word strings.Builder
	var quoted []bool
	l.readBalanced(&word, &quoted, "$(", 1)
	if word.String() != "$(a $(b))" {
		t.Fatalf("expected nested balanced, got %q", word.String())
	}
}
