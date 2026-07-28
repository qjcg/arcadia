package shell

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/qjcg/arcadia/exp/terebra/internal/parser"
)

type JobState string

const (
	JobRunning JobState = "running"
	JobDone    JobState = "done"
	JobStopped JobState = "stopped"
)

type Job struct {
	ID     int
	Cmd    *exec.Cmd
	State  JobState
	Line   string
	waitCh chan error
}

func (s *Shell) addJob(cmd *exec.Cmd, line string) *Job {
	s.jobSeq++
	job := &Job{
		ID:     s.jobSeq,
		Cmd:    cmd,
		State:  JobRunning,
		Line:   line,
		waitCh: make(chan error, 1),
	}
	go func() {
		job.waitCh <- cmd.Wait()
	}()
	s.jobs = append(s.jobs, job)
	return job
}

func (s *Shell) removeJob(job *Job) {
	for i, j := range s.jobs {
		if j == job {
			s.jobs = append(s.jobs[:i], s.jobs[i+1:]...)
			return
		}
	}
}

func (s *Shell) findJob(id int) *Job {
	for _, j := range s.jobs {
		if j.ID == id {
			return j
		}
	}
	return nil
}

func (s *Shell) formatJobs() []string {
	var result []string
	for _, j := range s.jobs {
		state := string(j.State)
		if j.Cmd.Process != nil {
			state = fmt.Sprintf("%s (pid %d)", state, j.Cmd.Process.Pid)
		}
		result = append(result, fmt.Sprintf("[%d] %s  %s", j.ID, state, j.Line))
	}
	return result
}

// executeBackground runs a command in the background and returns immediately.
func (s *Shell) executeBackground(cmd *parser.Command, stdin io.Reader, stdout io.Writer) error {
	// Check builtins - can't run them in background directly
	if _, ok := s.builtins.Lookup(cmd.Name); ok {
		// For builtins, just run them synchronously
		return fmt.Errorf("cannot run builtin %q in background", cmd.Name)
	}

	path, err := exec.LookPath(cmd.Name)
	if err != nil {
		return fmt.Errorf("command not found: %s", cmd.Name)
	}

	expandedArgs := make([]string, len(cmd.Args))
	for i, arg := range cmd.Args {
		expandedArgs[i] = s.expandVars(arg)
	}

	extCmd := exec.Command(path, expandedArgs...)
	extCmd.Stdin = stdin
	extCmd.Stdout = stdout
	extCmd.Stderr = s.Stderr
	extCmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	in, out, errOut, closeFn, err := s.openRedirects(cmd, stdin, stdout, nil)
	if err != nil {
		return err
	}
	if closeFn != nil {
		defer closeFn()
	}
	extCmd.Stdin = in
	extCmd.Stdout = out
	extCmd.Stderr = errOut

	var line strings.Builder
	line.WriteString(cmd.Name)
	for _, arg := range cmd.Args {
		line.WriteString(" " + arg)
	}

	if err := extCmd.Start(); err != nil {
		return err
	}

	job := s.addJob(extCmd, line.String())
	fmt.Fprintf(s.Stdout, "[%d] %d\n", job.ID, extCmd.Process.Pid)

	go func() {
		<-job.waitCh
		job.State = JobDone
		fmt.Fprintf(s.Stdout, "\n[%d] done  %s\n", job.ID, line.String())
		s.removeJob(job)
	}()

	return nil
}

// builtinJobs implements the jobs builtin.
func builtinJobs(args []string, stdout, stderr io.Writer, s *Shell) int {
	for _, line := range s.formatJobs() {
		fmt.Fprintln(stdout, line)
	}
	return 0
}

// builtinFg implements the fg builtin.
func builtinFg(args []string, stdout, stderr io.Writer, s *Shell) int {
	if len(s.jobs) == 0 {
		fmt.Fprintln(stderr, "fg: no current job")
		return 1
	}

	var job *Job
	if len(args) > 0 {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Fprintf(stderr, "fg: invalid job id: %s\n", args[0])
			return 1
		}
		job = s.findJob(id)
		if job == nil {
			fmt.Fprintf(stderr, "fg: job [%d] not found\n", id)
			return 1
		}
	} else {
		job = s.jobs[len(s.jobs)-1]
	}

	fmt.Fprintf(stdout, "%s\n", job.Line)

	if job.Cmd.Process == nil {
		fmt.Fprintln(stderr, "fg: job not running")
		return 1
	}

	// If the job was stopped, send SIGCONT to resume it
	if job.State == JobStopped {
		job.Cmd.Process.Signal(syscall.SIGCONT)
		job.State = JobRunning
	}

	// Set up a temporary SIGINT/SIGTSTP listener to forward signals
	// to the child process. The child is in its own process group
	// (background jobs use Setpgid), so the terminal sends SIGTSTP
	// to the shell's group — we must catch it here and forward it.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTSTP)
	defer signal.Stop(sig)

	select {
	case err := <-job.waitCh:
		// Job completed
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				fmt.Fprintf(stderr, "command exited with code %d\n", exitErr.ExitCode())
			} else {
				fmt.Fprintf(stderr, "fg: %v\n", err)
			}
		}
		job.State = JobDone
		s.removeJob(job)

	case received := <-sig:
		switch received {
		case syscall.SIGINT:
			// Ctrl+C pressed — forward SIGINT to the child
			job.Cmd.Process.Signal(syscall.SIGINT)
			<-job.waitCh
			// Drain any remaining signal so readline's handler
			// doesn't see a stale interrupt on the next Readline.
			select {
			case <-sig:
			default:
			}
			job.State = JobDone
			s.removeJob(job)

		case syscall.SIGTSTP:
			// ^Z pressed — forward SIGSTOP to the child
			job.Cmd.Process.Signal(syscall.SIGSTOP)
			job.State = JobStopped
			fmt.Fprintf(stdout, "\n[%d] stopped  %s\n", job.ID, job.Line)
			// Don't wait for waitCh — the process was stopped, not
			// exited. The addJob goroutine's Wait() call will
			// eventually return when the process is continued and
			// exits.
		}
	}

	return 0
}

// builtinBg implements the bg builtin.
func builtinBg(args []string, stdout, stderr io.Writer, s *Shell) int {
	if len(s.jobs) == 0 {
		fmt.Fprintln(stderr, "bg: no current job")
		return 1
	}

	var job *Job
	if len(args) > 0 {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Fprintf(stderr, "bg: invalid job id: %s\n", args[0])
			return 1
		}
		job = s.findJob(id)
		if job == nil {
			fmt.Fprintf(stderr, "bg: job [%d] not found\n", id)
			return 1
		}
	} else {
		job = s.jobs[len(s.jobs)-1]
	}

	if job.State == JobStopped {
		if job.Cmd.Process != nil {
			job.Cmd.Process.Signal(syscall.SIGCONT)
		}
		job.State = JobRunning
		fmt.Fprintf(stdout, "[%d] %s\n", job.ID, job.Line)
		return 0
	}

	fmt.Fprintln(stderr, "bg: job is not stopped")
	return 1
}

// cleanupJobs removes all completed jobs from the list.
func (s *Shell) cleanupJobs() {
	var active []*Job
	for _, j := range s.jobs {
		if j.State != JobDone {
			active = append(active, j)
		}
	}
	s.jobs = active
}
