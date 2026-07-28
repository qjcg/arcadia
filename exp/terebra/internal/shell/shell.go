package shell

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"github.com/chzyer/readline"
	"github.com/qjcg/arcadia/exp/terebra/internal/builtins"
	"github.com/qjcg/arcadia/exp/terebra/internal/drill"
	"github.com/qjcg/arcadia/exp/terebra/internal/parser"
	"github.com/qjcg/arcadia/exp/terebra/internal/plugin"
	"github.com/qjcg/arcadia/exp/terebra/internal/script"
)

const historyFile = ".terebra_history"

type Shell struct {
	Stdin    io.Reader
	Stdout   io.Writer
	Stderr   io.Writer
	builtins *builtins.Registry
	vars     map[string]string
	readonly map[string]bool
	aliases  map[string]string
	arrays   map[string][]string
	assoc    map[string]map[string]string
	funcs    map[string][]script.Stmt
	plugins  *plugin.Registry
	jobs     []*Job
	jobSeq   int
	exitCode int
	mu       sync.Mutex
	rl       *readline.Instance
	interp   *script.Interpreter
	debug    bool // set -x debug mode
	fuzzy    bool // Ctrl+R fuzzy search flag

	// stdin pipe for interrupting readline
	stdinR          *os.File // read end of pipe (readline reads from this)
	stdinW          *os.File // write end of pipe (goroutine writes to this)
	stdinDone       chan struct{}
	stdinNeedsReset bool // set when stdin pipe was paused for an external command
}

func (s *Shell) setExitCode(code int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.exitCode = code
}

func (s *Shell) getExitCode() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.exitCode
}

func New() *Shell {
	s := &Shell{
		Stdin:    os.Stdin,
		Stdout:   os.Stdout,
		Stderr:   os.Stderr,
		builtins: builtins.New(),
		vars:     make(map[string]string),
		readonly: make(map[string]bool),
		aliases:  make(map[string]string),
		arrays:   make(map[string][]string),
		assoc:    make(map[string]map[string]string),
		funcs:    make(map[string][]script.Stmt),
		plugins:  plugin.New(),
	}
	// Set default PS1 if not already set
	if s.getVar("PS1") == "" && os.Getenv("PS1") == "" {
		s.vars["PS1"] = "trb:{{.}}{{exit}}"
	}
	s.interp = script.NewInterpreter(s)
	// Register shell-aware builtins that need access to the Shell struct
	s.builtins.Register("export", func(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
		return s.builtinExport(args, stdout, stderr)
	})
	s.builtins.Register("unset", func(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
		return s.builtinUnset(args, stdout, stderr)
	})
	s.builtins.Register("set", func(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
		return s.builtinSet(args, stdout, stderr)
	})
	s.builtins.Register("jobs", func(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
		return builtinJobs(args, stdout, stderr, s)
	})
	s.builtins.Register("fg", func(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
		return builtinFg(args, stdout, stderr, s)
	})
	s.builtins.Register("bg", func(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
		return builtinBg(args, stdout, stderr, s)
	})
	s.builtins.Register("drill", drill.Handler)
	s.builtins.Register("cue", builtins.CueHandler())
	s.builtins.Register("plugin", plugin.PluginBuiltin(s.plugins))
	s.builtins.Register("state", func(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
		return s.builtinState(args, stdin, stdout, stderr)
	})
	s.builtins.Register("source", func(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
		return s.builtinSource(args, stdin, stdout, stderr)
	})
	s.builtins.Register("alias", func(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
		return s.builtinAlias(args, stdout, stderr)
	})
	s.builtins.Register("unalias", func(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
		return s.builtinUnalias(args, stdout, stderr)
	})
	s.builtins.Register("history", func(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
		return s.builtinHistory(args, stdout, stderr)
	})
	s.builtins.Register("readonly", func(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
		return s.builtinReadonly(args, stdout, stderr)
	})
	s.builtins.Register("exec", func(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
		return s.builtinExec(args, stdout, stderr)
	})
	s.builtins.Register("exit", func(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
		return s.builtinExit(args, stdout, stderr)
	})
	// Load ~/.terebrarc if it exists
	s.loadRc()
	return s
}

func Run() error {
	sh := New()
	return sh.Repl()
}

// RunScript executes a script file.
func RunScript(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("cannot execute %q: %v", path, err)
	}
	return RunScriptFromString(string(data))
}

// RunScriptFromString executes a script string.
func RunScriptFromString(content string) error {
	sh := New()
	// Skip shebang line if present
	if strings.HasPrefix(content, "#!") {
		idx := strings.IndexByte(content, '\n')
		if idx >= 0 {
			content = content[idx+1:]
		} else {
			content = ""
		}
	}

	// If the script spans multiple lines, use the scripting interpreter
	// (the shell parser treats newlines as whitespace, not as command separators).
	if strings.Contains(content, "\n") {
		return sh.interp.ParseAndExec(content)
	}

	// Try parsing as shell script first (supports ; && || chaining)
	script, err := parser.ParseScript(content)
	if err == nil && len(script.Pipelines) > 0 {
		err = sh.ExecuteScript(script)
		if errors.Is(err, errExitShell) {
			if sh.getExitCode() != 0 {
				os.Exit(sh.getExitCode())
			}
			return nil
		}
		return err
	}

	// Fall back to the scripting language interpreter
	return sh.interp.ParseAndExec(content)
}

func (s *Shell) Repl() error {
	histPath := ""
	if home := os.Getenv("HOME"); home != "" {
		histPath = filepath.Join(home, historyFile)
	}

	// Set up stdin pipe so we can interrupt readline on Ctrl+R
	s.setupStdinPipe()

	cfg := &readline.Config{
		Prompt:              "", // set below before first Readline
		HistoryFile:         histPath,
		HistoryLimit:        1000,
		AutoComplete:        s.completer(),
		Stdin:               s.stdinR,
		Stdout:              os.Stdout,
		Stderr:              os.Stderr,
		InterruptPrompt:     "^C\n",
		EOFPrompt:           "exit\n",
		HistorySearchFold:   true,
		FuncFilterInputRune: s.filterInput,
	}

	rl, err := readline.NewEx(cfg)
	if err != nil {
		return fmt.Errorf("readline: %v", err)
	}
	s.rl = rl
	defer rl.Close()

	for {
		// If the stdin pipe was paused for an external command, reset it
		// before the next readline call so readline reads from a fresh
		// pipe, not the stale (write-end-closed) one.
		if s.stdinNeedsReset {
			s.stdinNeedsReset = false
			rl.Close()
			s.closeStdinPipe()
			s.setupStdinPipe()
			cfg.Stdin = s.stdinR
			newRL, err := readline.NewEx(cfg)
			if err != nil {
				return fmt.Errorf("readline: %v", err)
			}
			s.rl = newRL
			rl = newRL
		}

		rl.SetPrompt(s.prompt())

		line, err := rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt {
				continue
			}
			if err == io.EOF {
				if s.fuzzy {
					// Ctrl+R triggered: close old readline, launch TUI
					s.fuzzy = false
					s.closeStdinPipe()
					rl.Close()

					result := s.runFuzzySearch()

					// Set up new pipe and readline
					s.setupStdinPipe()
					cfg.Stdin = s.stdinR
					newRL, err := readline.NewEx(cfg)
					if err != nil {
						return fmt.Errorf("readline: %v", err)
					}
					s.rl = newRL
					rl = newRL

					if result != "" {
						rl.Operation.SetBuffer(result)
					}
					continue
				}
				fmt.Fprintln(s.Stdout)
				return nil
			}
			return err
		}

		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		pipe, err := parser.ParseScript(line)
		if err != nil {
			fmt.Fprintf(s.Stderr, "trb: parse error: %v\n", err)
			continue
		}

		// Fill in heredoc content from interactive input if needed
		if s.rl != nil {
			s.fillHeredocs(pipe)
		}

		if err := s.ExecuteScript(pipe); err != nil {
			if errors.Is(err, errExitShell) {
				os.Exit(s.exitCode)
			}
			fmt.Fprintf(s.Stderr, "trb: error: %v\n", err)
			s.exitCode = 1
		} else {
			s.exitCode = 0
		}
	}
}

func (s *Shell) filterInput(r rune) (rune, bool) {
	// Ctrl+R (0x12) triggers fuzzy search
	if r == 0x12 {
		s.fuzzy = true
		// Write a Ctrl+D (EOF) to the pipe so readline returns immediately
		if s.stdinW != nil {
			s.stdinW.Write([]byte{0x04})
		}
		return r, false // consume the key
	}
	return r, true
}

// setupStdinPipe creates a pipe and starts a goroutine that copies
// from os.Stdin into the pipe. Readline reads from the read end of
// the pipe so we can interrupt it by closing the pipe.
func (s *Shell) setupStdinPipe() {
	r, w, err := os.Pipe()
	if err != nil {
		// Fall back to direct stdin
		s.stdinR = os.Stdin
		s.stdinW = nil
		s.stdinDone = nil
		return
	}
	s.stdinR = r
	s.stdinW = w
	s.stdinDone = make(chan struct{})
	s.Stdin = r // redirect all reads through the pipe so the goroutine is the sole os.Stdin reader

	go func() {
		defer close(s.stdinDone)
		buf := make([]byte, 4096)
		for {
			// Use a helper goroutine to read so we can detect
			// when the pipe is closed (via s.stdinDone)
			type readResult struct {
				n   int
				err error
			}
			ch := make(chan readResult, 1)
			go func() {
				n, err := os.Stdin.Read(buf)
				ch <- readResult{n, err}
			}()

			select {
			case r := <-ch:
				if r.err != nil {
					return
				}
				if _, err := w.Write(buf[:r.n]); err != nil {
					return
				}
			case <-s.stdinDone:
				return
			}
		}
	}()
}

// pauseStdinPipe stops the stdin pipe goroutine and restores direct stdin
// without closing the read end (so readline can be reused after resuming).
func (s *Shell) pauseStdinPipe() {
	if s.stdinW != nil {
		s.stdinW.Close()
		s.stdinW = nil
	}
	if s.stdinDone != nil {
		// Close stdin to unblock the goroutine's os.Stdin.Read(buf) so it
		// can exit immediately instead of waiting for the user to type.
		os.Stdin.Close()
		<-s.stdinDone
		s.stdinDone = nil
		// Reopen stdin from the terminal.
		if f, err := os.OpenFile("/dev/tty", os.O_RDONLY, 0); err == nil {
			os.Stdin = f
		}
	}
	s.Stdin = os.Stdin
}

// closeStdinPipe closes the stdin pipe and waits for the copy goroutine
// to exit.
func (s *Shell) closeStdinPipe() {
	if s.stdinW != nil {
		s.stdinW.Close()
	}
	if s.stdinDone != nil {
		// Close stdin to unblock the goroutine's os.Stdin.Read(buf) so it
		// can exit immediately instead of waiting for the user to type.
		os.Stdin.Close()
		<-s.stdinDone
		// Reopen stdin from the terminal.
		if f, err := os.OpenFile("/dev/tty", os.O_RDONLY, 0); err == nil {
			os.Stdin = f
		}
	}
	if s.stdinR != nil && s.stdinR != os.Stdin {
		s.stdinR.Close()
	}
	s.stdinR = nil
	s.stdinW = nil
	s.stdinDone = nil
	s.Stdin = os.Stdin // restore direct stdin after pipe is closed
}

// readHistory reads the history file and returns all lines.
func (s *Shell) readHistory() []string {
	home := os.Getenv("HOME")
	if home == "" {
		return nil
	}
	histPath := filepath.Join(home, historyFile)
	data, err := os.ReadFile(histPath)
	if err != nil {
		return nil
	}
	lines := strings.Split(string(data), "\n")
	// Remove trailing empty line
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	// Reverse so most recent is first
	for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
		lines[i], lines[j] = lines[j], lines[i]
	}
	return lines
}

// loadRc loads ~/.terebrarc if it exists, executing it as a script.
func (s *Shell) loadRc() {
	home := os.Getenv("HOME")
	if home == "" {
		return
	}
	rcPath := filepath.Join(home, ".terebrarc")
	data, err := os.ReadFile(rcPath)
	if err != nil {
		return // file doesn't exist or can't be read — skip silently
	}
	s.interp.ParseAndExec(string(data))
}

// fillHeredocs reads heredoc content from the user when it wasn't provided
// in the input line (e.g. interactive REPL usage).
func (s *Shell) fillHeredocs(script *parser.Script) {
	for _, pipe := range script.Pipelines {
		for _, cmd := range pipe.Commands {
			for _, redir := range cmd.Redirects {
				if (redir.Type == parser.RedirectHeredoc || redir.Type == parser.RedirectHeredocDash) && redir.Content == "" {
					var content strings.Builder
					delimiter := redir.File
					for {
						s.rl.SetPrompt("> ")
						line, err := s.rl.Readline()
						if err != nil {
							break
						}
						trimmed := line
						if redir.Type == parser.RedirectHeredocDash {
							trimmed = strings.TrimLeft(line, "\t")
						}
						if trimmed == delimiter {
							break
						}
						content.WriteString(line)
						content.WriteByte('\n')
					}
					redir.Content = content.String()
				}
			}
		}
	}
}

func (s *Shell) prompt() string {
	// Check for PS1 customization — check shell vars first, then env
	ps1 := s.getVar("PS1")
	if ps1 == "" {
		ps1 = os.Getenv("PS1")
	}
	if ps1 != "" {
		// Expand $(...) command substitutions first
		ps1 = s.expandPromptCmd(ps1)
		// Then process template vars
		if strings.Contains(ps1, "{{.}}") {
			wd, _ := os.Getwd()
			home := os.Getenv("HOME")
			display := wd
			if after, ok := strings.CutPrefix(display, home); ok {
				display = "~" + after
			}
			mark := "$"
			if s.exitCode != 0 {
				mark = "!"
			}
			ps1 = strings.ReplaceAll(ps1, "{{.}}", display)
			ps1 = strings.ReplaceAll(ps1, "{{exit}}", mark)
			ps1 = strings.ReplaceAll(ps1, "{{exitcode}}", fmt.Sprintf("%d", s.exitCode))
		}
		// Handle multi-line prompts: print preceding lines to stderr,
		// return only the last line as the readline prompt.
		if idx := strings.LastIndex(ps1, "\n"); idx >= 0 {
			fmt.Fprint(s.Stderr, ps1[:idx], "\r\n")
			ps1 = ps1[idx+1:]
		}
		ps1 = strings.TrimRight(ps1, "\n\r")
		// Strip ANSI escape codes so readline computes cursor position
		// correctly. Trim trailing spaces so the appended space is the
		// only one — users who include a trailing space in their PS1
		// won't get a double space.
		ps1 = strings.TrimRight(stripANSI(ps1), " ")
		return ps1 + " "
	}
	wd, _ := os.Getwd()
	home := os.Getenv("HOME")
	display := wd
	if after, ok := strings.CutPrefix(display, home); ok {
		display = "~" + after
	}
	mark := "$"
	if s.exitCode != 0 {
		mark = "!"
	}
	return fmt.Sprintf("trb:%s%s ", display, mark)
}

// stripANSI removes ANSI escape codes (both CSI and OSC sequences) from s.
func stripANSI(s string) string {
	var result strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\033' {
			if i+1 < len(s) && s[i+1] == '[' {
				// CSI sequence: \033[...<letter>
				i += 2
				for i < len(s) && !(s[i] >= 'A' && s[i] <= 'Z' || s[i] >= 'a' && s[i] <= 'z') {
					i++
				}
				if i < len(s) {
					i++
				}
				continue
			}
			if i+1 < len(s) && s[i+1] == ']' {
				// OSC sequence: \033]...\007 or \033]...\033\\
				i += 2
				for i < len(s) {
					if s[i] == '\007' {
						i++
						break
					}
					if s[i] == '\033' && i+1 < len(s) && s[i+1] == '\\' {
						i += 2
						break
					}
					i++
				}
				continue
			}
		}
		result.WriteByte(s[i])
		i++
	}
	return result.String()
}

// expandPromptCmd finds $(...) in input, executes each command via sh -c,
// and replaces the expression with the command's stdout.
func (s *Shell) expandPromptCmd(input string) string {
	if !strings.Contains(input, "$(") {
		return input
	}
	var result strings.Builder
	i := 0
	for i < len(input) {
		if input[i] == '$' && i+1 < len(input) && input[i+1] == '(' {
			i += 2 // skip $(
			// Find matching ), handling nesting
			depth := 1
			start := i
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
			// Execute the command via sh -c
			cmd := exec.Command("sh", "-c", cmdStr)
			output, err := cmd.Output()
			if err == nil {
				out := strings.TrimRight(string(output), "\n\r")
				result.WriteString(out)
			} // on error, substitute nothing
			continue
		}
		result.WriteByte(input[i])
		i++
	}
	return result.String()
}

// builtinExit handles the exit builtin. It sets the shell's exit code and
// returns -1 as a sentinel that the executor interprets as "exit the shell".
// builtinExec handles the exec builtin. It replaces the current process with
// the given command using syscall.Exec, which never returns on success.
func (s *Shell) builtinExec(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "exec: missing command")
		return 1
	}

	argv0 := ""
	cmdArgs := args
	if args[0] == "-a" {
		if len(args) < 2 {
			fmt.Fprintln(stderr, "exec: option requires an argument: -a")
			return 1
		}
		argv0 = args[1]
		cmdArgs = args[2:]
	}

	if len(cmdArgs) == 0 {
		fmt.Fprintln(stderr, "exec: missing command")
		return 1
	}

	command := cmdArgs[0]
	path, err := exec.LookPath(command)
	if err != nil {
		s.exitCode = 127
		fmt.Fprintf(stderr, "exec: %v\n", err)
		return 127
	}

	argv := cmdArgs
	if argv0 != "" {
		argv = make([]string, len(cmdArgs))
		argv[0] = argv0
		copy(argv[1:], cmdArgs[1:])
	}

	envv := os.Environ()

	if err := syscall.Exec(path, argv, envv); err != nil {
		s.exitCode = 126
		fmt.Fprintf(stderr, "exec: %v\n", err)
		return 126
	}

	return 0
}

func (s *Shell) builtinExit(args []string, stdout, stderr io.Writer) int {
	code := 0
	if len(args) > 0 {
		if _, err := fmt.Sscanf(args[0], "%d", &code); err != nil {
			fmt.Fprintf(stderr, "exit: invalid exit code: %s\n", args[0])
			code = 1
		}
	}
	s.exitCode = code
	return -1
}
