package jobctl_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/qjcg/arcadia/exp/terebra/spike/jobctl"
)

// TestJobControlOverInterp is the issue #12 spike scenario: one
// external command is suspended (Ctrl+Z semantics), inspected with
// jobs, resumed with fg, suspended again, resumed with bg, and a
// backgrounded command never blocks the shell. Everything runs over a
// persistent interp.Runner with a job-control exec handler.
func TestJobControlOverInterp(t *testing.T) {
	if testing.Short() {
		t.Skip("spike drives real processes and signals")
	}

	sh, err := jobctl.New()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	done := make(chan error, 1)

	// A foreground command blocks the REPL while it runs.
	go func() { done <- sh.Exec(ctx, "sleep 30") }()
	if !waitJobs(t, sh, 1) {
		t.Fatal("foreground job never appeared in the table")
	}

	// Ctrl+Z: SIGTSTP to the job's process group stops it and the
	// handler hands control back to the shell instead of blocking.
	if err := sh.Stop(1); err != nil {
		t.Fatal(err)
	}
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Exec did not return after the job stopped")
	}
	if !sh.WaitState(1, jobctl.Stopped, time.Second) {
		t.Fatalf("job 1 did not stop: %v", sh.Jobs())
	}
	t.Logf("after Ctrl+Z: %v", sh.Jobs())

	// jobs shows the stopped job.
	if rows := strings.Join(sh.Jobs(), "\n"); !strings.Contains(rows, "stopped") {
		t.Fatalf("jobs does not list a stopped job: %q", rows)
	}

	// fg resumes the job and blocks the shell again; a second stop
	// hands control back.
	go func() { done <- sh.Exec(ctx, "fg") }()
	if !sh.WaitState(1, jobctl.Running, 5*time.Second) {
		t.Fatalf("fg did not resume job 1: %v", sh.Jobs())
	}
	if err := sh.Stop(1); err != nil {
		t.Fatal(err)
	}
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("fg did not return after the job stopped again")
	}
	if !sh.WaitState(1, jobctl.Stopped, time.Second) {
		t.Fatalf("job 1 did not stop the second time: %v", sh.Jobs())
	}
	t.Logf("after second Ctrl+Z: %v", sh.Jobs())

	// bg resumes without blocking the shell.
	if err := sh.Exec(ctx, "bg %1"); err != nil {
		t.Fatal(err)
	}
	if !sh.WaitState(1, jobctl.Running, 5*time.Second) {
		t.Fatalf("bg did not resume job 1: %v", sh.Jobs())
	}

	// A command with & must not block the shell at all, and must
	// still register in the job table.
	go func() { done <- sh.Exec(ctx, "sleep 30 &") }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("backgrounded command blocked the shell")
	}
	if !waitJobs(t, sh, 2) {
		t.Fatal("background job never appeared in the table")
	}
	t.Logf("final: %v", sh.Jobs())

	_ = sh.Kill(1)
	_ = sh.Kill(2)
}

func waitJobs(t *testing.T, sh *jobctl.Shell, n int) bool {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if len(sh.Jobs()) >= n {
			return true
		}
		time.Sleep(5 * time.Millisecond)
	}
	return false
}
