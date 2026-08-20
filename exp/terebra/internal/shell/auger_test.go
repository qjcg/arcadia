package shell

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/qjcg/arcadia/exp/terebra/internal/parser"
)

func TestExecuteAugerPipelineSimple(t *testing.T) {
	s := newTestShell()
	pipe := &parser.Pipeline{
		Commands: []*parser.Command{
			{Name: "echo", Args: []string{"a: 1"}},
			{Name: "cat"},
		},
		Connects: []parser.ConnectType{parser.ConnectAuger},
	}
	path := redirectStdoutToFile(t, s)
	if err := s.executeAugerPipeline(pipe); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := readFileWaits(t, path, "a:", 2000); !strings.Contains(got, "a:") {
		t.Fatalf("expected a: in output, got %q", got)
	}
}

func TestExecuteAugerPipelineTooFew(t *testing.T) {
	s := newTestShell()
	pipe := &parser.Pipeline{
		Commands: []*parser.Command{{Name: "echo"}},
		Connects: []parser.ConnectType{},
	}
	if err := s.executeAugerPipeline(pipe); err == nil {
		t.Fatal("expected error for single-command auger pipe")
	}
}

func TestExecuteAugerPipelineInvalidCUE(t *testing.T) {
	s := newTestShell()
	pipe := &parser.Pipeline{
		Commands: []*parser.Command{
			{Name: "echo", Args: []string{"not valid cue {"}},
			{Name: "cat"},
		},
		Connects: []parser.ConnectType{parser.ConnectAuger},
	}
	path := redirectStdoutToFile(t, s)
	if err := s.executeAugerPipeline(pipe); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := readFileWaits(t, path, "not valid cue", 2000); !strings.Contains(got, "not valid cue") {
		t.Fatalf("expected raw passthrough, got %q", got)
	}
}

func TestExecuteAugerWithEncoderJSON(t *testing.T) {
	s := newTestShell()
	pipe := &parser.Pipeline{
		Commands: []*parser.Command{
			{Name: "echo", Args: []string{`a: 1`}},
		},
		Encoder: "json",
	}
	var out bytes.Buffer
	s.Stdout = &out
	if err := s.executeAugerWithEncoder(pipe); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), `"a"`) {
		t.Fatalf("expected JSON output, got %q", out.String())
	}
}

func TestExecuteAugerWithEncoderCUE(t *testing.T) {
	s := newTestShell()
	pipe := &parser.Pipeline{
		Commands: []*parser.Command{
			{Name: "echo", Args: []string{`a: 1`}},
		},
		Encoder: "cue",
	}
	var out bytes.Buffer
	s.Stdout = &out
	if err := s.executeAugerWithEncoder(pipe); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "a:") {
		t.Fatalf("expected cue output, got %q", out.String())
	}
}

func TestExecuteAugerWithEncoderInvalidCUE(t *testing.T) {
	s := newTestShell()
	pipe := &parser.Pipeline{
		Commands: []*parser.Command{
			{Name: "echo", Args: []string{"not valid {"}},
		},
		Encoder: "json",
	}
	var out bytes.Buffer
	s.Stdout = &out
	if err := s.executeAugerWithEncoder(pipe); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "not valid") {
		t.Fatalf("expected raw output, got %q", out.String())
	}
}

func cleanPromptShell() *Shell {
	s := &Shell{
		vars:     map[string]string{},
		readonly: map[string]bool{},
		aliases:  map[string]string{},
		arrays:   map[string][]string{},
		assoc:    map[string]map[string]string{},
	}
	s.Stdout = &bytes.Buffer{}
	s.Stderr = &bytes.Buffer{}
	return s
}

func TestPromptPS1Subst(t *testing.T) {
	s := cleanPromptShell()
	s.vars["PS1"] = "{{.}}> "
	got := s.prompt()
	if !strings.Contains(got, "> ") {
		t.Fatalf("expected '> ' tail, got %q", got)
	}
}

func TestPromptPS1ExitMark(t *testing.T) {
	s := cleanPromptShell()
	s.vars["PS1"] = "{{.}}{{exit}}"
	s.setExitCode(1)
	got := s.prompt()
	if !strings.Contains(got, "!") {
		t.Fatalf("expected '!' exit mark, got %q", got)
	}
}

func TestPromptPS1MultiLine(t *testing.T) {
	s := cleanPromptShell()
	s.vars["PS1"] = "line1\nline2"
	got := s.prompt()
	if !strings.Contains(got, "line2") {
		t.Fatalf("expected last line as prompt, got %q", got)
	}
}

func TestPromptDefault(t *testing.T) {
	s := cleanPromptShell()
	os.Unsetenv("PS1")
	got := s.prompt()
	if !strings.Contains(got, "trb:") {
		t.Fatalf("expected trb: default prompt, got %q", got)
	}
	os.Setenv("PS1", "")
}
