package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

var (
	buildOnce sync.Once
	binPath   string
	buildErr  error
)

// buildBinary compiles the terebra binary once and returns its path.
func buildBinary(t *testing.T) string {
	t.Helper()
	buildOnce.Do(func() {
		dir, err := os.MkdirTemp("", "terebra-test-*")
		if err != nil {
			buildErr = err
			return
		}
		binPath = filepath.Join(dir, "terebra")
		cmd := exec.Command("go", "build", "-o", binPath, ".")
		cmd.Dir = "."
		out, err := cmd.CombinedOutput()
		if err != nil {
			buildErr = err
			t.Logf("build output: %s", out)
		}
	})
	if buildErr != nil {
		t.Fatalf("build terebra: %v", buildErr)
	}
	return binPath
}

type repl struct {
	t    *testing.T
	cmd  *exec.Cmd
	pty  *os.File
	mu   sync.Mutex
	acc  []byte
	done chan struct{}
}

func startRepl(t *testing.T) *repl {
	t.Helper()
	// Hermetic like startReplWithHome: temp HOME, no ambient PS1, so
	// prompt assertions never depend on the developer's shell config.
	return startReplWithHome(t, t.TempDir())
}

func (r *repl) send(s string) {
	r.t.Helper()
	if _, err := r.pty.Write([]byte(s)); err != nil {
		r.t.Fatalf("pty write: %v", err)
	}
}

func (r *repl) waitFor(sub string, timeout time.Duration) bool {
	r.t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if bytes.Contains(r.snapshot(), []byte(sub)) {
			return true
		}
		time.Sleep(20 * time.Millisecond)
	}
	return bytes.Contains(r.snapshot(), []byte(sub))
}

// snapshot returns a copy of the accumulated output for safe reading.
func (r *repl) snapshot() []byte {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]byte(nil), r.acc...)
}

func (r *repl) close() {
	r.t.Helper()
	r.cmd.Process.Kill()
	select {
	case <-r.done:
	case <-time.After(2 * time.Second):
		r.t.Error("repl did not exit")
	}
	r.pty.Close()
}

func TestReplSingleEnterAfterExternalCommand(t *testing.T) {
	r := startRepl(t)
	defer r.close()

	if !r.waitFor("trb:", 5*time.Second) {
		t.Fatalf("no prompt; output so far: %q", r.snapshot())
	}

	r.send("echo FIRST\n")
	if !r.waitFor("FIRST", 3*time.Second) {
		t.Fatalf("builtin did not run on single Enter; output: %q", r.snapshot())
	}

	r.send("echo EXTERNAL\n")
	if !r.waitFor("EXTERNAL", 3*time.Second) {
		t.Fatalf("external command did not run on single Enter; output: %q", r.snapshot())
	}

	r.send("echo SECOND\n")
	if !r.waitFor("SECOND", 3*time.Second) {
		t.Fatalf("builtin after external command did not run on single Enter; output: %q", r.snapshot())
	}

	r.send("echo EXTERNAL2\n")
	if !r.waitFor("EXTERNAL2", 3*time.Second) {
		t.Fatalf("second external command did not run; output: %q", r.snapshot())
	}
	r.send("echo THIRD\n")
	if !r.waitFor("THIRD", 3*time.Second) {
		t.Fatalf("builtin after second external command did not run; output: %q", r.snapshot())
	}
}

func TestReplRapidInput(t *testing.T) {
	r := startRepl(t)
	defer r.close()

	if !r.waitFor("trb:", 5*time.Second) {
		t.Fatalf("no prompt; output so far: %q", r.snapshot())
	}

	r.send("echo EXTERNAL\necho B1\necho B2\n")
	if !r.waitFor("EXTERNAL", 3*time.Second) {
		t.Fatalf("external command in burst did not run; output: %q", r.snapshot())
	}
	if !r.waitFor("B1", 3*time.Second) {
		t.Fatalf("B1 in burst did not run; output: %q", r.snapshot())
	}
	if !r.waitFor("B2", 3*time.Second) {
		t.Fatalf("B2 in burst did not run; output: %q", r.snapshot())
	}
}

func TestReplPromptAppears(t *testing.T) {
	r := startRepl(t)
	defer r.close()
	if !r.waitFor("trb:", 5*time.Second) {
		t.Fatalf("no prompt; output so far: %q", r.snapshot())
	}
}

// TestReplPromptCmdSubstKeepsInputAlive is the regression test for
// the REPL hang: a startup file setting PS1 to a command substitution
// (the common oh-my-posh shape) must not starve the stdin pump, so
// typed input still reaches readline afterwards.
func TestReplPromptCmdSubstKeepsInputAlive(t *testing.T) {
	home := t.TempDir()
	rc := "export PS1='$(/bin/echo psub)> '\n"
	if err := os.WriteFile(filepath.Join(home, ".terebrarc"),
		[]byte(rc), 0o644); err != nil {
		t.Fatal(err)
	}
	r := startReplWithHome(t, home)
	defer r.close()

	if !r.waitFor("psub>", 5*time.Second) {
		t.Fatalf("prompt from cmdsubst PS1 never rendered; output: %q",
			r.snapshot())
	}
	r.send("echo ALIVE\n")
	if !r.waitFor("ALIVE", 3*time.Second) {
		t.Fatalf("input dead after cmdsubst prompt; output: %q", r.snapshot())
	}
}

// TestReplMultilineContinuation proves the REPL accumulates lines
// until the bash parser sees complete input, and that --explain works
// against the new parse path (issue #17 acceptance).
func TestReplMultilineContinuation(t *testing.T) {
	r := startRepl(t)
	defer r.close()
	if !r.waitFor("trb:", 5*time.Second) {
		t.Fatalf("no prompt; output so far: %q", r.snapshot())
	}

	r.send("if true; then\n")
	r.send("echo MULTI\n")
	r.send("fi\n")
	if !r.waitFor("MULTI", 3*time.Second) {
		t.Fatalf("multi-line block did not run; output: %q", r.snapshot())
	}

	r.send("--explain echo hello\n")
	if !r.waitFor("# would execute: echo hello", 3*time.Second) {
		t.Fatalf("--explain did not run; output: %q", r.snapshot())
	}
}
