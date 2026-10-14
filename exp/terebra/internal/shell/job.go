package shell

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"syscall"

	"mvdan.cc/sh/v3/interp"
)

// JobState is the state a job holds in the table.
type JobState string

const (
	JobRunning JobState = "running"
	JobStopped JobState = "stopped"
	JobDone    JobState = "done"
)

// Job is one external command tracked by the table.
type Job struct {
	ID    int
	Pid   int
	Pgid  int
	Line  string
	State JobState
}

// execJob is the engine's External handler: every external command
// runs in its own process group and is waited for with WUNTRACED, so
// Ctrl+Z stops it and returns control to the REPL (issue #12 layer).
func (s *Shell) execJob(ctx context.Context, args []string) error {
	hc := interp.HandlerCtx(ctx)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = hc.Stdin, hc.Stdout, hc.Stderr
	cmd.Dir = hc.Dir
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// The command's process group gets it SIGTTIN-stopped by the
	// kernel if it reads the terminal while the shell is foreground --
	// bash semantics. The stdin pump is paused by the REPL loop around
	// Eval (see withTerminal), never here: handlers for background
	// commands run on interp goroutines and must not touch it.

	if err := cmd.Start(); err != nil {
		return error(interp.ExitStatus(uint8(127)))
	}
	job := s.addJob(cmd.Process.Pid, strings.Join(args, " "))
	return s.waitJob(job, cmd.Process.Pid)
}

// waitJob consumes a job's state changes without reaping it until it
// exits. os/exec's Cmd.Wait cannot report stops, so the shell waits
// with Wait4 directly. A stop returns an exit status to the runner,
// which is the hook that hands control back to the REPL on Ctrl+Z.
func (s *Shell) waitJob(job *Job, pid int) error {
	for {
		var ws syscall.WaitStatus
		if _, err := syscall.Wait4(pid, &ws,
			syscall.WUNTRACED|syscall.WCONTINUED, nil); err != nil {
			return error(interp.ExitStatus(uint8(1)))
		}
		switch {
		case ws.Stopped():
			s.setJobState(job, JobStopped)
			return error(interp.ExitStatus(128 + uint8(ws.StopSignal())))
		case ws.Continued():
			s.setJobState(job, JobRunning)
		case ws.Exited():
			s.setJobState(job, JobDone)
			return error(interp.ExitStatus(uint8(ws.ExitStatus())))
		case ws.Signaled():
			s.setJobState(job, JobDone)
			return error(interp.ExitStatus(128 + uint8(ws.Signal())))
		}
	}
}

// printJobs lists the table like the jobs builtin.
func (s *Shell) printJobs(w io.Writer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, j := range s.jobs {
		if j.State == JobDone {
			continue
		}
		mark := " "
		if j == s.currentJob() {
			mark = "+"
		}
		state := "Running"
		if j.State == JobStopped {
			state = "Stopped"
		}
		fmt.Fprintf(w, "[%d]%s %s  %s\n", j.ID, mark, state, j.Line)
	}
}

// fgJob resumes a job and waits for its next state change, like the
// fg builtin: the job runs, the shell blocks.
func (s *Shell) fgJob(args []string) {
	j := s.pickJob(jobArg(args))
	if j == nil {
		fmt.Fprintln(s.Stderr, "fg: no current job")
		return
	}
	fmt.Fprintln(s.Stdout, j.Line)
	syscall.Kill(-j.Pgid, syscall.SIGCONT)
	s.waitJob(j, j.Pid)
}

// bgJob resumes a job without waiting, like the bg builtin.
func (s *Shell) bgJob(args []string) {
	j := s.pickJob(jobArg(args))
	if j == nil {
		fmt.Fprintln(s.Stderr, "bg: no current job")
		return
	}
	fmt.Fprintf(s.Stdout, "[%d] %s &\n", j.ID, j.Line)
	syscall.Kill(-j.Pgid, syscall.SIGCONT)
	s.setJobState(j, JobRunning)
}

func (s *Shell) addJob(pid int, line string) *Job {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobSeq++
	j := &Job{ID: s.jobSeq, Pid: pid, Pgid: pid, Line: line, State: JobRunning}
	s.jobs = append(s.jobs, j)
	return j
}

// currentJob returns the most recently added live job (the "+" job).
func (s *Shell) currentJob() *Job {
	var cur *Job
	for _, j := range s.jobs {
		if j.State != JobDone {
			cur = j
		}
	}
	return cur
}

// pickJob selects a job by id (0 = the current job).
func (s *Shell) pickJob(id int) *Job {
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
	return s.currentJob()
}

func (s *Shell) setJobState(job *Job, st JobState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	job.State = st
}

// jobArg parses a %N job spec (bare N accepted).
func jobArg(args []string) int {
	if len(args) == 0 {
		return 0
	}
	var id int
	if _, err := fmt.Sscanf(strings.TrimPrefix(args[0], "%"), "%d", &id); err != nil {
		return 0
	}
	return id
}
