package shell

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/qjcg/arcadia/exp/terebra/internal/parser"
	"github.com/qjcg/arcadia/exp/terebra/internal/script"
)

func TestApplyTempEnv(t *testing.T) {
	s := newTestShell()
	cmd := &parser.Command{Name: "FOO=bar", Args: []string{"echo", "hi"}}
	restore, ok := s.applyTempEnv(cmd)
	if !ok {
		t.Fatal("expected temp env applied")
	}
	if cmd.Name != "echo" {
		t.Fatalf("expected cmd name 'echo', got %q", cmd.Name)
	}
	if len(cmd.Args) != 1 {
		t.Fatalf("expected 1 remaining arg, got %d", len(cmd.Args))
	}
	restore()
	if _, existed := s.vars["FOO"]; existed {
		t.Fatal("expected FOO restored/removed")
	}
}

func TestApplyTempEnvNoEquals(t *testing.T) {
	s := newTestShell()
	cmd := &parser.Command{Name: "echo", Args: []string{"hi"}}
	_, ok := s.applyTempEnv(cmd)
	if ok {
		t.Fatal("expected not applied for plain command")
	}
}

func TestApplyTempEnvOnlyAssignment(t *testing.T) {
	s := newTestShell()
	cmd := &parser.Command{Name: "FOO=bar"}
	_, ok := s.applyTempEnv(cmd)
	if ok {
		t.Fatal("expected not applied with no command after")
	}
}

func TestApplyTempEnvExistingVar(t *testing.T) {
	s := newTestShell()
	s.vars["FOO"] = "old"
	cmd := &parser.Command{Name: "FOO=new", Args: []string{"echo"}}
	restore, ok := s.applyTempEnv(cmd)
	if !ok {
		t.Fatal("expected applied")
	}
	if s.getVar("FOO") != "new" {
		t.Fatalf("expected FOO=new, got %q", s.getVar("FOO"))
	}
	restore()
	if s.getVar("FOO") != "old" {
		t.Fatalf("expected FOO restored to 'old', got %q", s.getVar("FOO"))
	}
}

func TestHandleAssignmentVar(t *testing.T) {
	s := newTestShell()
	if !s.handleAssignment("x=5", nil) {
		t.Fatal("expected assignment handled")
	}
	if s.getVar("x") != "5" {
		t.Fatalf("expected x=5, got %q", s.getVar("x"))
	}
}

func TestHandleAssignmentNotAssignment(t *testing.T) {
	s := newTestShell()
	if s.handleAssignment("echo", []string{"hi"}) {
		t.Fatal("expected false for non-assignment")
	}
}

func TestExpandAlias(t *testing.T) {
	s := newTestShell()
	s.aliases["ll"] = "echo hi"
	name, args := s.expandAlias("ll", []string{"world"})
	if name != "echo" {
		t.Fatalf("expected alias expanded to 'echo', got %q", name)
	}
	if strings.Join(args, ",") != "hi,world" {
		t.Fatalf("expected args hi,world, got %q", strings.Join(args, ","))
	}
}

func TestExpandAliasNone(t *testing.T) {
	s := newTestShell()
	name, args := s.expandAlias("git", []string{"status"})
	if name != "git" {
		t.Fatalf("expected unchanged name, got %q", name)
	}
	if len(args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(args))
	}
}

func TestHandleExplain(t *testing.T) {
	s := newTestShell()
	var out strings.Builder
	s.Stdout = &out
	if !s.handleExplain("--explain", []string{"echo", "hi"}) {
		t.Fatal("expected explain handled")
	}
	if !strings.Contains(out.String(), "# would execute:") {
		t.Fatalf("expected explain output, got %q", out.String())
	}
}

func TestHandleExplainNotExplain(t *testing.T) {
	s := newTestShell()
	if s.handleExplain("echo", []string{"hi"}) {
		t.Fatal("expected not explain")
	}
}

func TestExecuteScriptEmpty(t *testing.T) {
	s := newTestShell()
	if err := s.ExecuteScript(&parser.Script{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecuteScriptChainingAnd(t *testing.T) {
	s := newTestShell()
	// echo a && echo b — both should run
	script := &parser.Script{
		Pipelines: []*parser.Pipeline{
			{Commands: []*parser.Command{{Name: "true"}}},
			{Commands: []*parser.Command{{Name: "true"}}},
		},
		Ops: []parser.ChainingOp{parser.ChainingAnd},
	}
	if err := s.ExecuteScript(script); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLevenshtein(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"", "", 0},
		{"a", "", 1},
		{"", "a", 1},
		{"kitten", "sitting", 3},
		{"flaw", "lawn", 2},
		{"same", "same", 0},
	}
	for _, c := range cases {
		if got := levenshtein(c.a, c.b); got != c.want {
			t.Errorf("levenshtein(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestSuggestCommand(t *testing.T) {
	s := newTestShell()
	// "ech" is close to "echo" builtin
	if got := s.suggestCommand("ech"); got != "echo" {
		t.Fatalf("expected 'echo', got %q", got)
	}
}

func TestSuggestCommandNoMatch(t *testing.T) {
	s := newTestShell()
	if got := s.suggestCommand("zzzzzz"); got != "" {
		t.Fatalf("expected empty suggestion, got %q", got)
	}
}

func TestExpandCmdSubstNoSubst(t *testing.T) {
	s := newTestShell()
	if got := s.expandCmdSubst("plain"); got != "plain" {
		t.Fatalf("expected 'plain', got %q", got)
	}
}

func TestExpandCmdSubst(t *testing.T) {
	s := newTestShell()
	got := s.expandCmdSubst("$(echo hi)")
	if got != "hi" {
		t.Fatalf("expected 'hi', got %q", got)
	}
}

func TestOpenRedirectsStdout(t *testing.T) {
	s := newTestShell()
	dir := t.TempDir()
	file := dir + "/out.txt"
	cmd := &parser.Command{
		Redirects: []*parser.Redirect{{Type: parser.RedirectStdout, File: file}},
	}
	in, out, _, closeFn, err := s.openRedirects(cmd, strings.NewReader(""), &strings.Builder{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if in == nil {
		t.Fatal("expected stdin reader")
	}
	if out == nil {
		t.Fatal("expected stdout writer")
	}
	closeFn()
}

func TestOpenRedirectsStdin(t *testing.T) {
	s := newTestShell()
	dir := t.TempDir()
	file := dir + "/in.txt"
	if err := os.WriteFile(file, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := &parser.Command{
		Redirects: []*parser.Redirect{{Type: parser.RedirectStdin, File: file}},
	}
	in, _, _, closeFn, err := s.openRedirects(cmd, strings.NewReader(""), &strings.Builder{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if in == nil {
		t.Fatal("expected stdin reader")
	}
	closeFn()
}

func TestOpenRedirectsError(t *testing.T) {
	s := newTestShell()
	cmd := &parser.Command{
		Redirects: []*parser.Redirect{{Type: parser.RedirectStdout, File: "/nonexistent/dir/out.txt"}},
	}
	_, _, _, _, err := s.openRedirects(cmd, strings.NewReader(""), &strings.Builder{}, nil)
	if err == nil {
		t.Fatal("expected error for bad redirect path")
	}
}

func TestOpenRedirectsHeredoc(t *testing.T) {
	s := newTestShell()
	cmd := &parser.Command{
		Redirects: []*parser.Redirect{{Type: parser.RedirectHeredoc, File: "EOF", Content: "hello"}},
	}
	in, _, _, closeFn, err := s.openRedirects(cmd, strings.NewReader(""), &strings.Builder{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if in == nil {
		t.Fatal("expected stdin reader for heredoc")
	}
	closeFn()
}

func TestExecutePipedCommands(t *testing.T) {
	s := newTestShell()
	var out strings.Builder
	s.Stdout = &out
	cmds := []*parser.Command{
		{Name: "echo", Args: []string{"hello"}},
		{Name: "cat"},
	}
	err := s.executePipedCommands(cmds, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "hello") {
		t.Fatalf("expected 'hello' in output, got %q", out.String())
	}
}

func TestOpenRedirectsAppend(t *testing.T) {
	s := newTestShell()
	dir := t.TempDir()
	file := dir + "/log.txt"
	cmd := &parser.Command{Redirects: []*parser.Redirect{{Type: parser.RedirectAppend, File: file}}}
	_, out, _, closeFn, err := s.openRedirects(cmd, strings.NewReader(""), &strings.Builder{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected append writer")
	}
	closeFn()
}

func TestOpenRedirectsBoth(t *testing.T) {
	s := newTestShell()
	dir := t.TempDir()
	file := dir + "/both.txt"
	cmd := &parser.Command{Redirects: []*parser.Redirect{{Type: parser.RedirectBoth, File: file}}}
	_, out, errOut, closeFn, err := s.openRedirects(cmd, strings.NewReader(""), &strings.Builder{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil || errOut == nil {
		t.Fatal("expected both writers")
	}
	closeFn()
}

func TestOpenRedirectsStderrToOut(t *testing.T) {
	s := newTestShell()
	cmd := &parser.Command{Redirects: []*parser.Redirect{{Type: parser.RedirectStderrToStdout, File: "1"}}}
	_, out, errOut, _, err := s.openRedirects(cmd, strings.NewReader(""), &strings.Builder{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errOut != out {
		t.Fatal("expected stderr to be redirected to stdout")
	}
}

func TestOpenRedirectsStderr(t *testing.T) {
	s := newTestShell()
	dir := t.TempDir()
	file := dir + "/err.txt"
	cmd := &parser.Command{Redirects: []*parser.Redirect{{Type: parser.RedirectStderr, File: file}}}
	_, _, errOut, closeFn, err := s.openRedirects(cmd, strings.NewReader(""), &strings.Builder{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errOut == nil {
		t.Fatal("expected stderr writer")
	}
	closeFn()
}

func TestOpenRedirectsStderrAppend(t *testing.T) {
	s := newTestShell()
	dir := t.TempDir()
	file := dir + "/err.log"
	cmd := &parser.Command{Redirects: []*parser.Redirect{{Type: parser.RedirectStderrAppend, File: file}}}
	_, _, errOut, closeFn, err := s.openRedirects(cmd, strings.NewReader(""), &strings.Builder{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errOut == nil {
		t.Fatal("expected stderr append writer")
	}
	closeFn()
}

func TestOpenRedirectsBothAppend(t *testing.T) {
	s := newTestShell()
	dir := t.TempDir()
	file := dir + "/both.log"
	cmd := &parser.Command{Redirects: []*parser.Redirect{{Type: parser.RedirectBothAppend, File: file}}}
	_, out, errOut, closeFn, err := s.openRedirects(cmd, strings.NewReader(""), &strings.Builder{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil || errOut == nil {
		t.Fatal("expected both writers")
	}
	closeFn()
}

func TestExecuteCommandAlias(t *testing.T) {
	s := newTestShell()
	s.aliases["gg"] = "echo hello"
	cmd := &parser.Command{Name: "gg"}
	err := s.ExecuteCommand(cmd, strings.NewReader(""), &strings.Builder{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// alias expanded to echo hello; echo is a builtin -> exit 0
	if s.getExitCode() != 0 {
		t.Fatalf("expected exit 0, got %d", s.getExitCode())
	}
}

func TestExecuteCommandFunction(t *testing.T) {
	s := newTestShell()
	// Define a function whose body runs a pipeline (echo builtin)
	out := &strings.Builder{}
	body := []script.Stmt{
		&script.CommandStmt{Pipeline: &parser.Pipeline{Commands: []*parser.Command{{Name: "echo", Args: []string{"fn"}}}}},
	}
	s.SetFuncDef("myfunc", body)
	s.Stdout = out
	cmd := &parser.Command{Name: "myfunc"}
	if err := s.ExecuteCommand(cmd, strings.NewReader(""), out, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "fn") {
		t.Fatalf("expected function output 'fn', got %q", out.String())
	}
}

func TestExecuteCommandExportSkipsCmdSubst(t *testing.T) {
	s := newTestShell()
	cmd := &parser.Command{Name: "export", Args: []string{"FOO=$(echo hi)"}}
	err := s.ExecuteCommand(cmd, strings.NewReader(""), &strings.Builder{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// export should preserve the $(...) literal
	if s.getVar("FOO") != "$(echo hi)" {
		t.Fatalf("expected FOO to preserve cmdsubst, got %q", s.getVar("FOO"))
	}
}

func TestExecuteCommandBuiltinErrorExit(t *testing.T) {
	s := newTestShell()
	// 'set -z' returns exit 1
	cmd := &parser.Command{Name: "set", Args: []string{"-z"}}
	err := s.ExecuteCommand(cmd, strings.NewReader(""), &strings.Builder{}, nil)
	if err == nil {
		t.Fatal("expected error for failed builtin")
	}
	if s.getExitCode() != 1 {
		t.Fatalf("expected exit 1, got %d", s.getExitCode())
	}
}

func TestExecuteCommandDebugTrace(t *testing.T) {
	s := newTestShell()
	s.debug = true
	s.Stderr = &strings.Builder{}
	cmd := &parser.Command{Name: "echo", Args: []string{"hi"}}
	if err := s.ExecuteCommand(cmd, strings.NewReader(""), &strings.Builder{}, s.Stderr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(s.Stderr.(*strings.Builder).String(), "+ echo") {
		t.Fatalf("expected debug trace, got %q", s.Stderr.(*strings.Builder).String())
	}
}

func TestRunExternalCommandSuccess(t *testing.T) {
	s := newTestShell()
	var out, errb bytes.Buffer
	s.Stderr = &errb
	err := s.runExternalCommand("true", nil, strings.NewReader(""), &out, &errb)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.getExitCode() != 0 {
		t.Fatalf("expected exit 0, got %d", s.getExitCode())
	}
}

func TestRunExternalCommandOutput(t *testing.T) {
	s := newTestShell()
	var out bytes.Buffer
	err := s.runExternalCommand("echo", []string{"hello"}, strings.NewReader(""), &out, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "hello") {
		t.Fatalf("expected 'hello' in output, got %q", out.String())
	}
}

func TestRunExternalCommandNotFound(t *testing.T) {
	s := newTestShell()
	err := s.runExternalCommand("nonexistent-cmd-xyz123", nil, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error")
	}
	if s.getExitCode() != 127 {
		t.Fatalf("expected exit 127, got %d", s.getExitCode())
	}
}

func TestRunExternalCommandFails(t *testing.T) {
	s := newTestShell()
	var errb bytes.Buffer
	s.Stderr = &errb
	err := s.runExternalCommand("false", nil, strings.NewReader(""), &bytes.Buffer{}, &errb)
	if err == nil {
		t.Fatal("expected error for failing command")
	}
	if s.getExitCode() == 0 {
		t.Fatalf("expected non-zero exit code")
	}
}

func TestRunExternalCommandExitCode(t *testing.T) {
	s := newTestShell()
	err := s.runExternalCommand("sh", []string{"-c", "exit 7"}, strings.NewReader(""), &bytes.Buffer{}, s.Stderr)
	if err == nil {
		t.Fatal("expected error")
	}
	if s.getExitCode() != 7 {
		t.Fatalf("expected exit 7, got %d", s.getExitCode())
	}
}

func TestRunExternalCommandSuggestion(t *testing.T) {
	s := newTestShell()
	s.Stderr = &bytes.Buffer{}
	err := s.runExternalCommand("ecoh", nil, strings.NewReader(""), &bytes.Buffer{}, s.Stderr)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "did you mean") {
		t.Fatalf("expected suggestion in error, got %v", err)
	}
}

func TestRegisterStoppedJob(t *testing.T) {
	s := newTestShell()
	s.Stderr = &bytes.Buffer{}
	cmd := exec.Command("true")
	err := s.registerStoppedJob(cmd, "true")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(s.jobs))
	}
	if s.jobs[0].State != JobStopped {
		t.Fatalf("expected stopped state, got %v", s.jobs[0].State)
	}
}

func TestReportWaitResultGenericError(t *testing.T) {
	s := newTestShell()
	err := s.reportWaitResult(errors.New("boom"))
	if err == nil {
		t.Fatal("expected error")
	}
	if s.getExitCode() != 1 {
		t.Fatalf("expected exit 1, got %d", s.getExitCode())
	}
}

func TestReportWaitResultNil(t *testing.T) {
	s := newTestShell()
	if err := s.reportWaitResult(nil); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if s.getExitCode() != 0 {
		t.Fatalf("expected exit 0, got %d", s.getExitCode())
	}
}
