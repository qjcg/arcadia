// Package jobctl is the issue #12 spike: interactive job control over
// a persistent mvdan/sh interp.Runner.
//
// It exists to answer one question for ADR 0004: can terebra's
// jobs/fg/bg model survive when interp executes commands? The design
// it proves: interp runs semantics, a custom exec handler spawns every
// external command in its own process group and waits for it with
// WUNTRACED, and a terebra-owned job table does the bookkeeping. It is
// spike code: minimal, throwaway, and not wired into the shell.
package jobctl

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

// JobState mirrors the states a job can hold in the table.
type JobState string

const (
	Running JobState = "running"
	Stopped JobState = "stopped"
	Done    JobState = "done"
)

// Job is one external command tracked by the table.
type Job struct {
	ID    int
	Pid   int
	Pgid  int
	Line  string
	State JobState
}

// Shell is the spike's REPL: one persistent runner plus a job table.
type Shell struct {
	mu     sync.Mutex
	jobs   []*Job
	seq    int
	runner *interp.Runner
}

// New builds the spike shell with the job-control exec handler wired
// into a fresh persistent runner.
func New() (*Shell, error) {
	s := &Shell{}
	r, err := interp.New(
		interp.Interactive(true),
		interp.ExecHandlers(func(next interp.ExecHandlerFunc) interp.ExecHandlerFunc {
			return s.exec
		}),
	)
	if err != nil {
		return nil, err
	}
	s.runner = r
	return s, nil
}

// Exec runs one REPL line. Job-control words are dispatched by the
// shell itself; everything else goes through syntax + interp, the way
// the post-migration REPL will work.
func (s *Shell) Exec(ctx context.Context, line string) error {
	fields := strings.Fields(line)
	if len(fields) > 0 {
		switch fields[0] {
		case "jobs":
			return nil
		case "fg":
			s.fg(jobArg(fields))
			return nil
		case "bg":
			s.bg(jobArg(fields))
			return nil
		}
	}
	prog, err := syntax.NewParser().Parse(strings.NewReader(line), "")
	if err != nil {
		return err
	}
	return s.runner.Run(ctx, prog)
}

// Jobs renders the table like the jobs builtin would.
func (s *Shell) Jobs() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []string
	for _, j := range s.jobs {
		out = append(out,
			fmt.Sprintf("[%d] %s  %s (pid %d)", j.ID, j.State, j.Line, j.Pid))
	}
	return out
}

// Stop sends a stop signal to one job's process group, standing in for
// the terminal driver's Ctrl+Z.
func (s *Shell) Stop(id int) error {
	if j := s.find(id); j != nil {
		return syscall.Kill(-j.Pgid, syscall.SIGTSTP)
	}
	return fmt.Errorf("no job %d", id)
}

// Kill tears down one job's process group.
func (s *Shell) Kill(id int) error {
	if j := s.find(id); j != nil {
		return syscall.Kill(-j.Pgid, syscall.SIGKILL)
	}
	return fmt.Errorf("no job %d", id)
}

// exec is the job-control exec handler. Every external command lands
// here: own process group, WUNTRACED-aware wait, job-table updates.
// Stop notifications return an exit status to the runner instead of
// blocking it, which is what hands control back to the REPL on Ctrl+Z.
func (s *Shell) exec(ctx context.Context, args []string) error {
	hc := interp.HandlerCtx(ctx)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = hc.Stdin, hc.Stdout, hc.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		return error(interp.ExitStatus(uint8(127)))
	}
	job := s.add(cmd.Process.Pid, strings.Join(args, " "))
	return s.wait(job, cmd.Process.Pid)
}

// wait consumes state changes for a job's leader without reaping it
// until it exits. os/exec's Cmd.Wait cannot report stops, so the spike
// owns the wait directly.
func (s *Shell) wait(job *Job, pid int) error {
	for {
		var ws syscall.WaitStatus
		if _, err := syscall.Wait4(pid, &ws,
			syscall.WUNTRACED|syscall.WCONTINUED, nil); err != nil {
			return error(interp.ExitStatus(uint8(1)))
		}
		switch {
		case ws.Stopped():
			s.setState(job, Stopped)
			return error(interp.ExitStatus(128 + uint8(ws.StopSignal())))
		case ws.Continued():
			s.setState(job, Running)
		case ws.Exited():
			s.setState(job, Done)
			return error(interp.ExitStatus(uint8(ws.ExitStatus())))
		case ws.Signaled():
			s.setState(job, Done)
			return error(interp.ExitStatus(128 + uint8(ws.Signal())))
		}
	}
}

// fg resumes a job and waits for its next state change, like the fg
// builtin: the job runs, the shell blocks.
func (s *Shell) fg(id int) {
	j := s.pick(id)
	if j == nil {
		return
	}
	syscall.Kill(-j.Pgid, syscall.SIGCONT)
	s.wait(j, j.Pid)
}

// bg resumes a job without waiting, like the bg builtin.
func (s *Shell) bg(id int) {
	j := s.pick(id)
	if j == nil {
		return
	}
	syscall.Kill(-j.Pgid, syscall.SIGCONT)
	s.setState(j, Running)
}

func (s *Shell) add(pid int, line string) *Job {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq++
	j := &Job{ID: s.seq, Pid: pid, Pgid: pid, Line: line, State: Running}
	s.jobs = append(s.jobs, j)
	return j
}

func (s *Shell) find(id int) *Job {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, j := range s.jobs {
		if j.ID == id {
			return j
		}
	}
	return nil
}

// pick selects a job by id, or the last stopped job, or the last job.
func (s *Shell) pick(id int) *Job {
	s.mu.Lock()
	defer s.mu.Unlock()
	if id != 0 {
		for _, j := range s.jobs {
			if j.ID == id {
				return j
			}
		}
		return nil
	}
	var last *Job
	for _, j := range s.jobs {
		if j.State == Stopped {
			return j
		}
		last = j
	}
	return last
}

func (s *Shell) setState(job *Job, st JobState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	job.State = st
}

func jobArg(fields []string) int {
	if len(fields) < 2 {
		return 0
	}
	var id int
	if _, err := fmt.Sscanf(strings.TrimPrefix(fields[1], "%"), "%d", &id); err != nil {
		return 0
	}
	return id
}

// WaitState blocks until job id reaches the wanted state or the
// timeout expires. The spike's tests use it to synchronize with real
// signal delivery.
func (s *Shell) WaitState(id int, want JobState, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if j := s.find(id); j != nil && j.State == want {
			return true
		}
		time.Sleep(5 * time.Millisecond)
	}
	return false
}
