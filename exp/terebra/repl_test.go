package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/creack/pty"
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
	bin := buildBinary(t)
	cmd := exec.Command(bin)
	f, err := pty.Start(cmd)
	if err != nil {
		t.Fatalf("pty.Start: %v", err)
	}
	r := &repl{t: t, cmd: cmd, pty: f, done: make(chan struct{})}
	go func() {
		defer close(r.done)
		buf := make([]byte, 8192)
		for {
			n, err := f.Read(buf)
			if n > 0 {
				r.mu.Lock()
				r.acc = append(r.acc, buf[:n]...)
				r.mu.Unlock()
			}
			if err != nil {
				return
			}
		}
	}()
	return r
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

	if !r.waitFor("❯", 5*time.Second) {
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

	if !r.waitFor("❯", 5*time.Second) {
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
	if !r.waitFor("❯", 5*time.Second) {
		t.Fatalf("no prompt; output so far: %q", r.snapshot())
	}
}
