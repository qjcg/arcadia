package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/creack/pty"
)

func startReplWithHome(t *testing.T, home string) *repl {
	t.Helper()
	bin := buildBinary(t)
	cmd := exec.Command(bin)
	// Build a hermetic environment: drop HOME and PS1 so the REPL uses the
	// temp home and its default prompt rather than the developer's shell.
	env := make([]string, 0, len(os.Environ())+1)
	for _, kv := range os.Environ() {
		if strings.HasPrefix(kv, "HOME=") || strings.HasPrefix(kv, "PS1=") {
			continue
		}
		env = append(env, kv)
	}
	cmd.Env = append(env, "HOME="+home)
	f, err := pty.Start(cmd)
	if err != nil {
		t.Fatalf("pty.Start: %v", err)
	}
	pty.Setsize(f, &pty.Winsize{Rows: 24, Cols: 80})
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

func TestReplCtrlRFuzzySearch(t *testing.T) {
	for _, tc := range []struct{ name, key string }{
		{"enter executes", "\r"},
		{"tab recalls", "\t"},
		{"ctrl-g cancels", "\x07"},
	} {
		t.Run(tc.name, func(t *testing.T) { testReplCtrlR(t, tc.key) })
	}
}

func testReplCtrlR(t *testing.T, key string) {
	t.Helper()
	home := t.TempDir()
	marker := filepath.Join(home, "ctrlr_marker")
	hist := "touch " + marker + "\necho OTHER_BETA\n"
	if err := os.WriteFile(filepath.Join(home, ".terebra_history"), []byte(hist), 0o644); err != nil {
		t.Fatal(err)
	}

	r := startReplWithHome(t, home)
	defer r.close()

	if !r.waitFor("trb:", 5*time.Second) {
		t.Fatalf("no prompt; output: %q", r.snapshot())
	}

	r.send("\x12") // Ctrl+R
	if !r.waitFor("ctrlr_marker", 5*time.Second) {
		t.Fatalf("fuzzy search UI did not appear; output: %q", r.snapshot())
	}

	start := len(r.snapshot())
	r.send("ctrlr_marker" + key)
	if key != "\r" {
		deadline := time.Now().Add(5 * time.Second)
		for !strings.Contains(string(r.snapshot()[start:]), "trb:") {
			if time.Now().After(deadline) {
				t.Fatalf("did not return to prompt: %q", r.snapshot())
			}
			time.Sleep(20 * time.Millisecond)
		}
		if _, err := os.Stat(marker); !os.IsNotExist(err) {
			t.Fatalf("command ran without Enter: %v", err)
		}
		if key == "\x07" {
			r.send("echo CANCELLED-OK\r")
			if !r.waitFor("\nCANCELLED-OK\r\n", 5*time.Second) {
				t.Fatalf("cancel did not leave a clean prompt: %q", r.snapshot())
			}
			return
		}
		r.send("\r")
	}

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(marker); err == nil {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("selected history command did not execute; output: %q", r.snapshot())
}
