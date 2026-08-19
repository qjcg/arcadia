package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	testscript.Main(m, map[string]func(){
		"terebra": main,
	})
}

func TestTerebra(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata",
	})
}

func TestRunVersion(t *testing.T) {
	var out, errb bytes.Buffer
	code := run([]string{"--version"}, &out, &errb)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(out.String(), "terebra") {
		t.Fatalf("expected version output, got %q", out.String())
	}
}

func TestRunVersionShort(t *testing.T) {
	var out, errb bytes.Buffer
	if code := run([]string{"-v"}, &out, &errb); code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
}

func TestRunHelp(t *testing.T) {
	var out, errb bytes.Buffer
	if code := run([]string{"--help"}, &out, &errb); code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(out.String(), "Usage:") {
		t.Fatalf("expected usage output, got %q", out.String())
	}
}

func TestRunUnknownFlag(t *testing.T) {
	var out, errb bytes.Buffer
	if code := run([]string{"--bogus"}, &out, &errb); code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(errb.String(), "unknown flag") {
		t.Fatalf("expected unknown flag error, got %q", errb.String())
	}
}

func TestRunDashCMissingArg(t *testing.T) {
	var out, errb bytes.Buffer
	if code := run([]string{"-c"}, &out, &errb); code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
}

func TestRunDashCScript(t *testing.T) {
	var out, errb bytes.Buffer
	if code := run([]string{"-c", "echo hi"}, &out, &errb); code != 0 {
		t.Fatalf("expected exit 0, got %d (stderr: %s)", code, errb.String())
	}
}

func TestRunExplain(t *testing.T) {
	var out, errb bytes.Buffer
	// Output goes to the shell's own os.Stdout, not the passed writer,
	// so we only assert on the exit code here.
	if code := run([]string{"--explain", "echo", "hi"}, &out, &errb); code != 0 {
		t.Fatalf("expected exit 0, got %d (stderr: %s)", code, errb.String())
	}
}

func TestRunExplainMissingArg(t *testing.T) {
	var out, errb bytes.Buffer
	if code := run([]string{"--explain"}, &out, &errb); code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
}
