package builtins

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCueHandlerNoArgs(t *testing.T) {
	var out, errb bytes.Buffer
	code := CueHandler()(nil, strings.NewReader(""), &out, &errb)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(errb.String(), "expected subcommand") {
		t.Fatalf("expected subcommand error, got %q", errb.String())
	}
}

func TestCueHandlerUnknownSubcommand(t *testing.T) {
	var out, errb bytes.Buffer
	code := CueHandler()([]string{"bogus"}, strings.NewReader(""), &out, &errb)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(errb.String(), "unknown subcommand") {
		t.Fatalf("expected unknown subcommand error, got %q", errb.String())
	}
}

func TestCueEvalFromStdin(t *testing.T) {
	var out, errb bytes.Buffer
	code := cueEval(nil, strings.NewReader("a: 1\n"), &out, &errb)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d (stderr: %s)", code, errb.String())
	}
	if !strings.Contains(out.String(), "a:") {
		t.Fatalf("expected formatted output, got %q", out.String())
	}
}

func TestCueEvalFromFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "x.cue")
	if err := os.WriteFile(file, []byte("b: 2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	code := cueEval([]string{file}, strings.NewReader(""), &out, &errb)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d (stderr: %s)", code, errb.String())
	}
	if !strings.Contains(out.String(), "b:") {
		t.Fatalf("expected formatted output, got %q", out.String())
	}
}

func TestCueEvalInvalid(t *testing.T) {
	var out, errb bytes.Buffer
	code := cueEval(nil, strings.NewReader("a: {"), &out, &errb)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
}

func TestCueEvalMissingFile(t *testing.T) {
	var out, errb bytes.Buffer
	code := cueEval([]string{"/nonexistent/x.cue"}, strings.NewReader(""), &out, &errb)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
}

func TestCueVetValid(t *testing.T) {
	dir := t.TempDir()
	data := filepath.Join(dir, "data.cue")
	schema := filepath.Join(dir, "schema.cue")
	if err := os.WriteFile(data, []byte("a: 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(schema, []byte("a: int\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	code := cueVet([]string{data, schema}, strings.NewReader(""), &out, &errb)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d (stderr: %s)", code, errb.String())
	}
	if !strings.Contains(out.String(), "ok") {
		t.Fatalf("expected ok output, got %q", out.String())
	}
}

func TestCueVetInvalid(t *testing.T) {
	dir := t.TempDir()
	data := filepath.Join(dir, "data.cue")
	schema := filepath.Join(dir, "schema.cue")
	if err := os.WriteFile(data, []byte("a: \"str\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(schema, []byte("a: int\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	code := cueVet([]string{data, schema}, strings.NewReader(""), &out, &errb)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(errb.String(), "cue vet:") {
		t.Fatalf("expected cue vet error, got %q", errb.String())
	}
}

func TestCueVetTooFewArgs(t *testing.T) {
	var out, errb bytes.Buffer
	code := cueVet([]string{"only"}, strings.NewReader(""), &out, &errb)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
}

func TestCueVetMissingDataFile(t *testing.T) {
	var out, errb bytes.Buffer
	code := cueVet([]string{"/nonexistent/data.cue", "/nonexistent/schema.cue"}, strings.NewReader(""), &out, &errb)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
}

func TestCueExportFromStdin(t *testing.T) {
	var out, errb bytes.Buffer
	code := cueExport(nil, strings.NewReader("a: 1\n"), &out, &errb)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d (stderr: %s)", code, errb.String())
	}
	if !strings.Contains(out.String(), `"a"`) {
		t.Fatalf("expected JSON output, got %q", out.String())
	}
}

func TestCueExportFromFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "x.cue")
	if err := os.WriteFile(file, []byte("a: 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	code := cueExport([]string{file}, strings.NewReader(""), &out, &errb)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d (stderr: %s)", code, errb.String())
	}
	if !strings.Contains(out.String(), `"a"`) {
		t.Fatalf("expected JSON output, got %q", out.String())
	}
}

func TestCueDefFromStdin(t *testing.T) {
	var out, errb bytes.Buffer
	code := cueDef(nil, strings.NewReader("a: 1\n"), &out, &errb)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d (stderr: %s)", code, errb.String())
	}
	if out.Len() == 0 {
		t.Fatal("expected output")
	}
}

func TestCueFmtFromStdin(t *testing.T) {
	var out, errb bytes.Buffer
	code := cueFmt(nil, strings.NewReader("a:1\n"), &out, &errb)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d (stderr: %s)", code, errb.String())
	}
	if !strings.Contains(out.String(), "a: 1") {
		t.Fatalf("expected formatted output, got %q", out.String())
	}
}

func TestCueFmtFromFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "x.cue")
	if err := os.WriteFile(file, []byte("a:1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	code := cueFmt([]string{file}, strings.NewReader(""), &out, &errb)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d (stderr: %s)", code, errb.String())
	}
	if !strings.Contains(out.String(), "a: 1") {
		t.Fatalf("expected formatted output, got %q", out.String())
	}
}

func TestCueTrimFromStdin(t *testing.T) {
	var out, errb bytes.Buffer
	code := cueTrim(nil, strings.NewReader("a: 1\n"), &out, &errb)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d (stderr: %s)", code, errb.String())
	}
	if !strings.Contains(out.String(), "a:") {
		t.Fatalf("expected output, got %q", out.String())
	}
}
