package shell

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/qjcg/arcadia/exp/terebra/internal/cueutil"
	"github.com/qjcg/arcadia/exp/terebra/internal/expand"
	"github.com/qjcg/arcadia/exp/terebra/internal/parser"
)

func (s *Shell) ExecutePipeline(pipe *parser.Pipeline) error {
	if len(pipe.Commands) == 0 {
		return nil
	}

	// Check for auger pipes
	hasAuger := slices.Contains(pipe.Connects, parser.ConnectAuger)

	if hasAuger {
		return s.executeAugerPipeline(pipe)
	}

	if len(pipe.Commands) == 1 {
		cmd := pipe.Commands[0]
		if cmd.Background {
			return s.executeBackground(cmd, s.Stdin, s.Stdout)
		}
		return s.ExecuteCommand(cmd, s.Stdin, s.Stdout)
	}

	// Check if the last command in the pipeline is backgrounded
	last := pipe.Commands[len(pipe.Commands)-1]
	if last.Background {
		return fmt.Errorf("background not supported in pipelines yet")
	}

	return s.executePipedCommands(pipe.Commands)
}

func (s *Shell) executeAugerPipeline(pipe *parser.Pipeline) error {
	// For now, support a simple case: cmd1 |> cmd2
	// Execute cmd1, capture output, parse as CUE, pipe to cmd2
	cmds := pipe.Commands
	connects := pipe.Connects

	if len(cmds) < 2 {
		return fmt.Errorf("auger pipe requires at least 2 commands")
	}

	// Execute the first part of the pipeline (before the first auger pipe)
	// as a regular pipeline, capturing its output
	var buf bytes.Buffer
	if len(cmds) == 2 && len(connects) == 1 && connects[0] == parser.ConnectAuger {
		// Simple case: cmd1 |> cmd2
		if err := s.ExecuteCommand(cmds[0], s.Stdin, &buf); err != nil {
			return err
		}

		// Parse the output as CUE
		ctx := cueutil.NewContext()
		output := strings.TrimSpace(buf.String())
		if output == "" {
			return nil
		}

		v := cueutil.CompileString(ctx, output)
		if err := cueutil.Err(v); err != nil {
			// Not valid CUE - pass raw output to next command
			return s.ExecuteCommand(cmds[1], strings.NewReader(output), s.Stdout)
		}

		// Format the CUE value and pipe to the next command
		formatted, err := cueutil.FormatValue(v)
		if err != nil {
			return s.ExecuteCommand(cmds[1], strings.NewReader(output), s.Stdout)
		}

		return s.ExecuteCommand(cmds[1], strings.NewReader(formatted), s.Stdout)
	}

	// Complex case: mixed | and |> pipes
	// For now, just run as regular pipes
	return s.executePipedCommands(cmds)
}

func (s *Shell) ExecuteCommand(cmd *parser.Command, stdin io.Reader, stdout io.Writer) error {
	// Run expansion pipeline: brace expansion → variable expansion
	expandedName, expandedArgs := expand.ExpandCommand(cmd.Name, cmd.Args, s.expandVars)

	// Check for array assignment: name=(values ...)
	if s.tryArrayAssignment(expandedName, expandedArgs) {
		return nil
	}

	// Check for variable assignment: name=value
	if len(expandedArgs) == 0 && strings.Contains(expandedName, "=") && !strings.HasPrefix(expandedName, "-") {
		parts := strings.SplitN(expandedName, "=", 2)
		if len(parts) == 2 && isIdent(parts[0]) {
			s.setVar(parts[0], parts[1])
			return nil
		}
	}

	// Apply redirects to determine actual stdin/stdout/stderr
	in, out, errOut, closeFn, err := s.openRedirects(cmd, stdin, stdout)
	if err != nil {
		return err
	}
	if closeFn != nil {
		defer closeFn()
	}

	// Check builtins first
	if handler, ok := s.builtins.Lookup(expandedName); ok {
		exitCode := handler(expandedArgs, in, out, errOut)
		if exitCode != 0 {
			s.exitCode = exitCode
			return fmt.Errorf("command exited with code %d", exitCode)
		}
		s.exitCode = 0
		return nil
	}

	// Check user-defined functions
	if body, ok := s.funcs[expandedName]; ok {
		// Execute the function body
		for _, stmt := range body {
			if err := s.interp.ExecStmt(stmt); err != nil {
				return err
			}
		}
		return nil
	}

	// Look for external command
	path, err := exec.LookPath(expandedName)
	if err != nil {
		s.exitCode = 127
		return fmt.Errorf("command not found: %s", expandedName)
	}

	// Build the exec.Cmd
	extCmd := exec.Command(path, expandedArgs...)
	extCmd.Stdin = in
	extCmd.Stdout = out
	extCmd.Stderr = errOut

	if err := extCmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			s.exitCode = exitErr.ExitCode()
			return fmt.Errorf("command exited with code %d", exitErr.ExitCode())
		}
		s.exitCode = 1
		return err
	}

	s.exitCode = 0
	return nil
}

// openRedirects applies redirects and returns the effective stdin, stdout, stderr.
func (s *Shell) openRedirects(cmd *parser.Command, stdin io.Reader, stdout io.Writer) (io.Reader, io.Writer, io.Writer, func(), error) {
	in := stdin
	out := stdout
	errOut := s.Stderr
	var closers []io.Closer

	for _, redir := range cmd.Redirects {
		file := s.expandVars(redir.File)
		switch redir.Type {
		case parser.RedirectStdout:
			f, err := os.Create(file)
			if err != nil {
				return nil, nil, nil, nil, err
			}
			out = f
			closers = append(closers, f)

		case parser.RedirectStderr:
			f, err := os.Create(file)
			if err != nil {
				return nil, nil, nil, nil, err
			}
			errOut = f
			closers = append(closers, f)

		case parser.RedirectAppend:
			f, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
			if err != nil {
				return nil, nil, nil, nil, err
			}
			out = f
			closers = append(closers, f)

		case parser.RedirectStderrAppend:
			f, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
			if err != nil {
				return nil, nil, nil, nil, err
			}
			errOut = f
			closers = append(closers, f)

		case parser.RedirectStdin:
			f, err := os.Open(file)
			if err != nil {
				return nil, nil, nil, nil, err
			}
			in = f
			closers = append(closers, f)

		case parser.RedirectStderrToStdout:
			errOut = stdout
		}
	}

	closeFn := func() {
		for _, c := range closers {
			c.Close()
		}
	}
	return in, out, errOut, closeFn, nil
}

func (s *Shell) executePipedCommands(cmds []*parser.Command) error {
	n := len(cmds)
	pipes := make([]*os.File, 0, n-1)

	for i := 0; i < n-1; i++ {
		r, w, err := os.Pipe()
		if err != nil {
			return fmt.Errorf("pipe: %v", err)
		}
		pipes = append(pipes, r, w)
	}

	errCh := make(chan error, n)

	for i, cmd := range cmds {
		var stdin io.Reader
		var stdout io.Writer

		if i == 0 {
			stdin = s.Stdin
		} else {
			stdin = pipes[(i-1)*2]
		}

		if i == n-1 {
			stdout = s.Stdout
		} else {
			stdout = pipes[i*2+1]
		}

		go func(idx int, c *parser.Command, in io.Reader, out io.Writer) {
			errCh <- s.ExecuteCommand(c, in, out)
			if idx < n-1 {
				if w, ok := out.(*os.File); ok {
					w.Close()
				}
			}
		}(i, cmd, stdin, stdout)
	}

	var firstErr error
	for range n {
		if err := <-errCh; err != nil && firstErr == nil {
			firstErr = err
		}
	}

	for _, p := range pipes {
		p.Close()
	}

	return firstErr
}
