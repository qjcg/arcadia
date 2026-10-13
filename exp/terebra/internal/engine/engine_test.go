package engine_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/qjcg/arcadia/exp/terebra/internal/engine"
)

// runEval evaluates src on a fresh engine and returns stdout, stderr,
// and the runner error. The seam is engine.Eval: one interpreter,
// bash semantics from interp, terebra extras via the handler layer.
func runEval(t *testing.T, src string) (string, string, error) {
	t.Helper()
	var out, errOut bytes.Buffer
	e, err := engine.New(engine.Options{
		Stdout: &out,
		Stderr: &errOut,
	})
	if err != nil {
		t.Fatalf("engine.New: %v", err)
	}
	runErr := e.Eval(context.Background(), src)
	return out.String(), errOut.String(), runErr
}

func TestEval(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "should_run_bash_builtin_echo",
			src:  "echo hello",
			want: "hello\n",
		},
		{
			name: "should_honor_echo_dash_n_like_bash",
			src:  "echo -n hello",
			want: "hello",
		},
		{
			name: "should_expand_braces_like_terebra_did",
			src:  "echo {a,b}{1,2}",
			want: "a1 a2 b1 b2\n",
		},
		{
			name: "should_leave_quoted_braces_literal",
			src:  `echo '{a,b}'`,
			want: "{a,b}\n",
		},
		{
			name: "should_expand_variables_arrays_and_strings",
			src:  "A=hi; arr=(1 2 3); echo $A ${arr[1]} ${A^^}",
			want: "hi 2 HI\n",
		},
		{
			name: "should_pipe_between_commands",
			src:  "echo hello | tr a-z A-Z",
			want: "HELLO\n",
		},
		{
			name: "should_chain_with_and",
			src:  "true && echo yes || echo no",
			want: "yes\n",
		},
		{
			name: "should_substitute_commands",
			src:  "echo $(echo inner)",
			want: "inner\n",
		},
		{
			name: "should_run_auger_pipe_with_cue_formatting",
			src:  `echo "name: \"test\"" |> cat`,
			want: "name: \"test\"\n",
		},
		{
			name: "should_encode_auger_pipe_as_json",
			src:  `echo "name: \"test\"" |> json`,
			want: "{\"name\":\"test\"}\n",
		},
		{
			name: "should_pass_non_cue_through_auger_pipe_raw",
			src:  "echo 'hello there' |> cat",
			want: "hello there",
		},
		{
			name: "should_run_terebra_builtin_state",
			src:  "state",
			want: `version: "0.1.0"`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, errOut, err := runEval(t, tc.src)
			if err != nil {
				t.Errorf("Eval(%q) error: %v (stderr %q)",
					tc.src, err, errOut)
			}
			if !strings.Contains(out, tc.want) {
				t.Errorf("Eval(%q) stdout = %q, want it to contain %q",
					tc.src, out, tc.want)
			}
		})
	}
}

// TestRunScriptFromString covers the CLI script entry points: simple
// scripts, multi-line scripts, and shebang stripping.
func TestRunScriptFromString(t *testing.T) {
	for _, src := range []string{
		"echo hi",
		"echo a\necho b",
		"#!/bin/sh\necho a",
	} {
		if err := engine.RunScriptFromString(src); err != nil {
			t.Errorf("RunScriptFromString(%q) = %v, want nil", src, err)
		}
	}
}

// TestEvalSetEStopsScript is the set -e / trap recipe that replaces
// the retired try/catch/end blocks: errexit aborts and propagates.
func TestEvalSetEStopsScript(t *testing.T) {
	out, _, err := runEval(t, "set -e\nfalse\necho unreached")
	if err == nil {
		t.Fatal("Eval(set -e; false) = nil error, want exit status")
	}
	if code, ok := engine.ExitCode(err); !ok || code != 1 {
		t.Errorf("ExitCode(err) = %d, %v; want 1, true", code, ok)
	}
	if strings.Contains(out, "unreached") {
		t.Errorf("stdout = %q, must not contain 'unreached'", out)
	}
}

// TestEvalExitStatus pins the exit contract main.go relies on: a
// non-zero exit yields an error, and ExitCode maps it to the status.
func TestEvalExitStatus(t *testing.T) {
	_, _, err := runEval(t, "exit 3")
	if err == nil {
		t.Fatal("Eval(exit 3) = nil error, want exit status")
	}
	if code, ok := engine.ExitCode(err); !ok || code != 3 {
		t.Errorf("ExitCode(err) = %d, %v; want 3, true", code, ok)
	}
	_, _, err = runEval(t, "exit 0")
	if err != nil {
		t.Fatalf("Eval(exit 0) = %v, want nil", err)
	}
	if code, ok := engine.ExitCode(err); !ok || code != 0 {
		t.Errorf("ExitCode(nil) = %d, %v; want 0, true", code, ok)
	}
}
