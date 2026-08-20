package shell

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/qjcg/arcadia/exp/terebra/internal/parser"
)

func newTestShell() *Shell {
	s := New()
	s.Stdout = &bytes.Buffer{}
	s.Stderr = &bytes.Buffer{}
	return s
}

// redirectStdoutToFile points the shell's stdout at a fresh temp file and
// returns its path. Subprocess and job goroutines write to the file, which
// avoids the data races that would come from sharing a bytes.Buffer with the
// copy goroutine os/exec spawns when a subprocess writes to an io.Writer.
func redirectStdoutToFile(t *testing.T, s *Shell) string {
	f, err := os.CreateTemp("", "terebra-out-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	s.Stdout = f
	return f.Name()
}

// redirectStderrToFile points the shell's stderr at a fresh temp file and
// returns its path, mirroring redirectStdoutToFile.
func redirectStderrToFile(t *testing.T, s *Shell) string {
	f, err := os.CreateTemp("", "terebra-err-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	s.Stderr = f
	return f.Name()
}

// readFileWaits reads path, retrying until it contains want or the timeout
// (in ms) elapses. Since subprocess output is flushed by a goroutine shortly
// after the child exits, a short poll makes the assertion deterministic.
func readFileWaits(t *testing.T, path, want string, ms int) string {
	deadline := time.Now().Add(time.Duration(ms) * time.Millisecond)
	var data []byte
	for {
		var err error
		data, err = os.ReadFile(path)
		if err == nil && strings.Contains(string(data), want) {
			return string(data)
		}
		if time.Now().After(deadline) {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if data != nil {
		return string(data)
	}
	if _, err := os.ReadFile(path); err != nil {
		t.Fatalf("failed to read output file %s: %v", path, err)
	}
	return ""
}

func TestAddJobAndFindJob(t *testing.T) {
	s := newTestShell()
	cmd := exec.Command("true")
	job := s.addJob(cmd, "true")
	if job.ID != 1 {
		t.Fatalf("expected job ID 1, got %d", job.ID)
	}
	if s.findJob(1) != job {
		t.Fatal("expected to find job by ID")
	}
	if s.findJob(99) != nil {
		t.Fatal("expected nil for unknown job ID")
	}
}

func TestRemoveJob(t *testing.T) {
	s := newTestShell()
	job := s.addJob(exec.Command("true"), "true")
	s.removeJob(job)
	if len(s.jobs) != 0 {
		t.Fatalf("expected 0 jobs after removal, got %d", len(s.jobs))
	}
}

func TestFormatJobs(t *testing.T) {
	s := newTestShell()
	s.addJob(exec.Command("true"), "echo hi")
	lines := s.formatJobs()
	if len(lines) != 1 {
		t.Fatalf("expected 1 job line, got %d", len(lines))
	}
	if !strings.Contains(lines[0], "echo hi") {
		t.Fatalf("expected job line to contain command, got %q", lines[0])
	}
}

func TestCleanupJobs(t *testing.T) {
	s := newTestShell()
	done := s.addJob(exec.Command("true"), "done")
	done.State = JobDone
	s.addJob(exec.Command("true"), "running")
	s.cleanupJobs()
	if len(s.jobs) != 1 {
		t.Fatalf("expected 1 active job after cleanup, got %d", len(s.jobs))
	}
}

func TestBuiltinJobs(t *testing.T) {
	s := newTestShell()
	s.addJob(exec.Command("true"), "echo hi")
	var out, errb bytes.Buffer
	code := builtinJobs(nil, &out, &errb, s)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(out.String(), "echo hi") {
		t.Fatalf("expected job in output, got %q", out.String())
	}
}

func TestBuiltinFgNoJobs(t *testing.T) {
	s := newTestShell()
	var out, errb bytes.Buffer
	code := builtinFg(nil, &out, &errb, s)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(errb.String(), "no current job") {
		t.Fatalf("expected no current job error, got %q", errb.String())
	}
}

func TestBuiltinFgInvalidID(t *testing.T) {
	s := newTestShell()
	s.addJob(exec.Command("true"), "true")
	var out, errb bytes.Buffer
	code := builtinFg([]string{"abc"}, &out, &errb, s)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(errb.String(), "invalid job id") {
		t.Fatalf("expected invalid job id error, got %q", errb.String())
	}
}

func TestBuiltinFgJobNotFound(t *testing.T) {
	s := newTestShell()
	s.addJob(exec.Command("true"), "true")
	var out, errb bytes.Buffer
	code := builtinFg([]string{"99"}, &out, &errb, s)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(errb.String(), "not found") {
		t.Fatalf("expected not found error, got %q", errb.String())
	}
}

func TestBuiltinBgNoJobs(t *testing.T) {
	s := newTestShell()
	var out, errb bytes.Buffer
	code := builtinBg(nil, &out, &errb, s)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
}

func TestBuiltinBgInvalidID(t *testing.T) {
	s := newTestShell()
	s.addJob(exec.Command("true"), "true")
	var out, errb bytes.Buffer
	code := builtinBg([]string{"abc"}, &out, &errb, s)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
}

func TestBuiltinBgJobNotFound(t *testing.T) {
	s := newTestShell()
	s.addJob(exec.Command("true"), "true")
	var out, errb bytes.Buffer
	code := builtinBg([]string{"99"}, &out, &errb, s)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
}

func TestBuiltinBgNotStopped(t *testing.T) {
	s := newTestShell()
	s.addJob(exec.Command("true"), "true")
	var out, errb bytes.Buffer
	code := builtinBg(nil, &out, &errb, s)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(errb.String(), "not stopped") {
		t.Fatalf("expected not stopped error, got %q", errb.String())
	}
}

func TestStripANSI(t *testing.T) {
	cases := map[string]string{
		"plain":                "plain",
		"\x1b[31mred\x1b[0m":   "red",
		"\x1b]0;title\x07rest": "rest",
		"a\x1b[1mb\x1b[0mc":    "abc",
		"\x1b]2;x\x1b\\tail":   "tail",
	}
	for in, want := range cases {
		if got := stripANSI(in); got != want {
			t.Errorf("stripANSI(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestExpandPromptCmdNoSubst(t *testing.T) {
	s := newTestShell()
	got := s.expandPromptCmd("hello world")
	if got != "hello world" {
		t.Fatalf("expected unchanged, got %q", got)
	}
}

func TestExpandPromptCmdSubst(t *testing.T) {
	s := newTestShell()
	got := s.expandPromptCmd("$(echo hi)")
	if got != "hi" {
		t.Fatalf("expected 'hi', got %q", got)
	}
}

func TestExpandPromptCmdMixed(t *testing.T) {
	s := newTestShell()
	got := s.expandPromptCmd("pre $(echo mid) post")
	if got != "pre mid post" {
		t.Fatalf("expected 'pre mid post', got %q", got)
	}
}

func TestGetVarFallbackToEnv(t *testing.T) {
	s := newTestShell()
	s.vars["FOO"] = "bar"
	if got := s.getVar("FOO"); got != "bar" {
		t.Fatalf("expected 'bar', got %q", got)
	}
}

func TestGetArrayVar(t *testing.T) {
	s := newTestShell()
	s.setArray("arr", []string{"a", "b", "c"})
	if got := s.getArrayVar("arr", "0"); got != "a" {
		t.Fatalf("expected 'a', got %q", got)
	}
	if got := s.getArrayVar("arr", "@"); got != "a b c" {
		t.Fatalf("expected 'a b c', got %q", got)
	}
	if got := s.getArrayVar("arr", "#"); got != "3" {
		t.Fatalf("expected '3', got %q", got)
	}
	if got := s.getArrayVar("arr", "99"); got != "" {
		t.Fatalf("expected empty for out-of-range, got %q", got)
	}
	if got := s.getArrayVar("missing", "0"); got != "" {
		t.Fatalf("expected empty for missing array, got %q", got)
	}
}

func TestGetAssocArrayVar(t *testing.T) {
	s := newTestShell()
	s.assoc["m"] = map[string]string{"k1": "v1", "k2": "v2"}
	if got := s.getArrayVar("m", "k1"); got != "v1" {
		t.Fatalf("expected 'v1', got %q", got)
	}
	if got := s.getArrayVar("m", "#"); got != "2" {
		t.Fatalf("expected '2', got %q", got)
	}
}

func TestGetArrayVarListKeys(t *testing.T) {
	s := newTestShell()
	s.assoc["m"] = map[string]string{"b": "1", "a": "2"}
	got := s.getArrayVar("m", "!@")
	if got != "a b" {
		t.Fatalf("expected 'a b', got %q", got)
	}
	if v := s.getArrayVar("missing", "!@"); v != "" {
		t.Fatalf("expected empty for missing assoc keys, got %q", v)
	}
	if v := s.getArrayVar("arr", "!@"); v != "" {
		t.Fatalf("expected empty keys for regular array, got %q", v)
	}
}

func TestExecuteBackgroundBuiltinError(t *testing.T) {
	s := newTestShell()
	cmd := &parser.Command{Name: "echo", Args: []string{"hi"}}
	if err := s.executeBackground(cmd, strings.NewReader(""), &bytes.Buffer{}); err == nil {
		t.Fatal("expected error for builtin in background")
	}
}

func TestExecuteBackgroundNotFound(t *testing.T) {
	s := newTestShell()
	cmd := &parser.Command{Name: "nonexistent-cmd-xyz"}
	if err := s.executeBackground(cmd, strings.NewReader(""), &bytes.Buffer{}); err == nil {
		t.Fatal("expected error for missing command")
	}
}

func TestExecuteBackgroundCommand(t *testing.T) {
	s := newTestShell()
	path := redirectStdoutToFile(t, s)
	cmd := &parser.Command{Name: "true"}
	if err := s.executeBackground(cmd, strings.NewReader(""), s.Stdout); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = path
}

func TestBuiltinFgWaitingJob(t *testing.T) {
	s := newTestShell()
	var out, errb bytes.Buffer
	cmd := exec.Command("sleep", "0.05")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	job := s.addJob(cmd, "sleep 0.05")
	code := builtinFg(nil, &out, &errb, s)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d (stderr: %s)", code, errb.String())
	}
	if job.State != JobDone {
		t.Fatalf("expected job done, got %v", job.State)
	}
}

func TestBuiltinFgJobNotRunning(t *testing.T) {
	s := newTestShell()
	// Build the job manually so no addJob goroutine runs on an unstarted command.
	cmd := exec.Command("true")
	s.jobSeq++
	job := &Job{ID: s.jobSeq, Cmd: cmd, State: JobRunning, Line: "true"}
	s.jobs = append(s.jobs, job)
	var out, errb bytes.Buffer
	code := builtinFg(nil, &out, &errb, s)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(errb.String(), "not running") {
		t.Fatalf("expected not running error, got %q", errb.String())
	}
}
