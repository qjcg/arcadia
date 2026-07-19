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

	return s.executePipedCommands(pipe.Commands, pipe.Connects)
}

// ExecuteScript executes a parsed Script, handling &&, ||, ; chaining.
func (s *Shell) ExecuteScript(script *parser.Script) error {
	if len(script.Pipelines) == 0 {
		return nil
	}

	var lastErr error
	for i, pipe := range script.Pipelines {
		var err error
		// Execute a single pipeline
		if len(pipe.Commands) == 1 && !pipe.Commands[0].Background &&
			len(pipe.Connects) == 0 && len(pipe.Commands[0].Redirects) == 0 {
			// Simple command - use ExecuteCommand directly
			err = s.ExecuteCommand(pipe.Commands[0], s.Stdin, s.Stdout)
		} else {
			err = s.ExecutePipeline(pipe)
		}
		lastErr = err

		if i < len(script.Ops) {
			switch script.Ops[i] {
			case parser.ChainingThen:
				// Always continue regardless of exit code
				continue
			case parser.ChainingAnd:
				// Only continue if exit code is 0
				if s.exitCode != 0 {
					return err
				}
			case parser.ChainingOr:
				// Only continue if exit code is non-zero
				if s.exitCode == 0 && err == nil {
					return nil
				}
			}
		}
	}

	return lastErr
}

func (s *Shell) executeAugerPipeline(pipe *parser.Pipeline) error {
	// If the pipeline has an encoder, run the pipeline and encode the output
	if pipe.Encoder != "" {
		return s.executeAugerWithEncoder(pipe)
	}

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
	return s.executePipedCommands(cmds, pipe.Connects)
}

// executeAugerWithEncoder runs a pipeline and encodes the final CUE output.
func (s *Shell) executeAugerWithEncoder(pipe *parser.Pipeline) error {
	cmds := pipe.Commands

	// Execute the first command, capturing output
	var buf bytes.Buffer
	if err := s.ExecuteCommand(cmds[0], s.Stdin, &buf); err != nil {
		return err
	}

	// Parse as CUE
	ctx := cueutil.NewContext()
	output := strings.TrimSpace(buf.String())
	if output == "" {
		return nil
	}

	v := cueutil.CompileString(ctx, output)
	if err := cueutil.Err(v); err != nil {
		// Not valid CUE, output raw
		fmt.Fprint(s.Stdout, output)
		return nil
	}

	// Encode in the requested format
	switch pipe.Encoder {
	case "json":
		jsonBytes, err := cueutil.ToJSON(v)
		if err != nil {
			fmt.Fprint(s.Stdout, output)
			return nil
		}
		fmt.Fprintln(s.Stdout, string(jsonBytes))
	case "yaml":
		// For YAML, format as CUE then convert
		formatted, err := cueutil.FormatValue(v)
		if err != nil {
			fmt.Fprint(s.Stdout, output)
			return nil
		}
		fmt.Fprintln(s.Stdout, formatted)
	case "cue":
		formatted, err := cueutil.FormatValue(v)
		if err != nil {
			fmt.Fprint(s.Stdout, output)
			return nil
		}
		fmt.Fprintln(s.Stdout, formatted)
	}

	return nil
}

func (s *Shell) ExecuteCommand(cmd *parser.Command, stdin io.Reader, stdout io.Writer) error {
	// Handle temporary env var assignments: FOO=bar echo $FOO
	// Scan the command name and leading args for name=value patterns.
	// When followed by a command, the env vars are set only for that command.
	var restoreEnv []struct {
		name    string
		value   string
		existed bool
	}
	if strings.Contains(cmd.Name, "=") && !strings.HasPrefix(cmd.Name, "-") {
		parts := strings.SplitN(cmd.Name, "=", 2)
		if len(parts) == 2 && isIdent(parts[0]) && !strings.HasPrefix(parts[1], "(") {
			type envPair struct{ name, value string }
			var envs []envPair
			envs = append(envs, envPair{parts[0], parts[1]})
			var remainingArgs []string
			for _, arg := range cmd.Args {
				if strings.Contains(arg, "=") && !strings.HasPrefix(arg, "-") {
					p := strings.SplitN(arg, "=", 2)
					if len(p) == 2 && isIdent(p[0]) {
						envs = append(envs, envPair{p[0], p[1]})
						continue
					}
				}
				remainingArgs = append(remainingArgs, arg)
			}
			// Only treat as temporary if there's a command to run
			if len(remainingArgs) > 0 {
				for _, e := range envs {
					oldVal, existed := s.vars[e.name]
					restoreEnv = append(restoreEnv, struct {
						name    string
						value   string
						existed bool
					}{e.name, oldVal, existed})
					s.setVar(e.name, e.value)
					s.exportVar(e.name)
				}
				cmd.Name = remainingArgs[0]
				cmd.Args = remainingArgs[1:]
				defer func() {
					for _, re := range restoreEnv {
						if re.existed {
							s.vars[re.name] = re.value
						} else {
							delete(s.vars, re.name)
						}
						_ = os.Unsetenv(re.name)
						if re.existed && re.value != "" {
							os.Setenv(re.name, re.value)
						}
					}
				}()
			}
		}
	}

	// Run expansion pipeline: brace expansion → variable expansion
	expandedName, expandedArgs := expand.ExpandCommand(cmd.Name, cmd.Args, s.expandVars)

	// Debug trace
	if s.debug {
		var line strings.Builder
		line.WriteString(expandedName)
		for _, arg := range expandedArgs {
			line.WriteString(" " + arg)
		}
		fmt.Fprintf(s.Stderr, "+ %s\n", line.String())
	}

	// --explain flag: dry-run mode
	if expandedName == "--explain" || (len(expandedArgs) > 0 && expandedArgs[0] == "--explain") {
		actualName := expandedName
		if actualName == "--explain" {
			actualName = expandedArgs[0]
			expandedArgs = expandedArgs[1:]
		} else {
			expandedArgs = expandedArgs[1:]
		}
		fmt.Fprintf(s.Stdout, "# would execute: %s", actualName)
		for _, arg := range expandedArgs {
			fmt.Fprintf(s.Stdout, " %s", arg)
		}
		fmt.Fprintln(s.Stdout)
		return nil
	}

	// Expand aliases
	if alias, ok := s.aliases[expandedName]; ok {
		aliasParts := strings.Fields(alias)
		if len(aliasParts) > 0 {
			expandedName = aliasParts[0]
			expandedArgs = append(aliasParts[1:], expandedArgs...)
		}
	}

	// Check for array assignment: name=(values ...)
	if s.tryArrayAssignment(expandedName, expandedArgs) {
		return nil
	}

	// Check for variable assignment: name=value
	if len(expandedArgs) == 0 && strings.Contains(expandedName, "=") && !strings.HasPrefix(expandedName, "-") {
		parts := strings.SplitN(expandedName, "=", 2)
		if len(parts) == 2 && isIdent(parts[0]) {
			// Don't expand $(...) in PS1 — it's evaluated on prompt render
			s.setVar(parts[0], parts[1])
			return nil
		}
	}

	// Expand $(...) command substitution in name and args
	// Skip for export/readonly args — variable assignment values should
	// preserve $(...) for dynamic evaluation (e.g. PS1='$(oh-my-posh ...)')
	if expandedName == "export" || expandedName == "readonly" {
		expandedName = s.expandCmdSubst(expandedName)
		// Don't expand $(...) in the args — they're assignment values
	} else {
		expandedName = s.expandCmdSubst(expandedName)
		for i, arg := range expandedArgs {
			expandedArgs[i] = s.expandCmdSubst(arg)
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
		// Suggest similar commands
		if suggestion := s.suggestCommand(expandedName); suggestion != "" {
			return fmt.Errorf("command not found: %s (did you mean %s?)", expandedName, suggestion)
		}
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

		case parser.RedirectBoth:
			f, err := os.Create(file)
			if err != nil {
				return nil, nil, nil, nil, err
			}
			out = f
			errOut = f
			closers = append(closers, f)

		case parser.RedirectBothAppend:
			f, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
			if err != nil {
				return nil, nil, nil, nil, err
			}
			out = f
			errOut = f
			closers = append(closers, f)

		case parser.RedirectHeredoc, parser.RedirectHeredocDash, parser.RedirectHereString:
			// Heredoc: create a pipe with the content
			content := redir.Content
			if !redir.Quoted {
				// Expand variables in heredoc content
				content = s.expandVars(content)
			}
			// Write content to a pipe for stdin
			r, w, err := os.Pipe()
			if err != nil {
				return nil, nil, nil, nil, err
			}
			go func() {
				w.Write([]byte(content))
				w.Close()
			}()
			in = r
			closers = append(closers, r)
		}
	}

	closeFn := func() {
		for _, c := range closers {
			c.Close()
		}
	}
	return in, out, errOut, closeFn, nil
}

func (s *Shell) executePipedCommands(cmds []*parser.Command, connects []parser.ConnectType) error {
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

// levenshtein computes the Levenshtein distance between two strings.
func levenshtein(a, b string) int {
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}
	row := make([]int, lb+1)
	for i := range row {
		row[i] = i
	}
	for i := range la {
		prev := i + 1
		for j := range lb {
			cost := 1
			if a[i] == b[j] {
				cost = 0
			}
			val := min(row[j]+1, min(prev+1, row[j+1]+cost))
			row[j] = prev
			prev = val
		}
		row[lb] = prev
	}
	return row[lb]
}

// expandCmdSubst expands $(...) command substitutions in s.
// Each $(cmd) is executed via sh -c and replaced with the command's stdout.
func (s *Shell) expandCmdSubst(input string) string {
	if !strings.Contains(input, "$(") {
		return input
	}
	var result strings.Builder
	i := 0
	for i < len(input) {
		if input[i] == '$' && i+1 < len(input) && input[i+1] == '(' {
			i += 2 // skip $(
			start := i
			depth := 1
			for i < len(input) && depth > 0 {
				if input[i] == '(' {
					depth++
				} else if input[i] == ')' {
					depth--
				}
				if depth > 0 {
					i++
				}
			}
			cmdStr := input[start:i]
			if i < len(input) {
				i++ // skip )
			}
			cmd := exec.Command("sh", "-c", cmdStr)
			output, err := cmd.Output()
			if err == nil {
				result.WriteString(strings.TrimRight(string(output), "\n\r"))
			}
			continue
		}
		result.WriteByte(input[i])
		i++
	}
	return result.String()
}

// suggestCommand finds the closest matching builtin or PATH command.
func (s *Shell) suggestCommand(name string) string {
	best := ""
	bestDist := 3

	for _, cmd := range s.builtins.Names() {
		if dist := levenshtein(name, cmd); dist < bestDist && dist <= 2 {
			bestDist = dist
			best = cmd
		}
	}

	drillSubs := []string{"cue", "fs", "proc", "net"}
	for _, sub := range drillSubs {
		full := "drill " + sub
		if dist := levenshtein(name, full); dist < bestDist {
			bestDist = dist
			best = full
		}
	}

	return best
}
