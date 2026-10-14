// Package shell is terebra's interactive REPL around the interp
// engine: readline, prompt, completion, job control, and the REPL-only
// commands. All evaluation goes through internal/engine (ADR 0004),
// so the REPL, -c, and script mode share one parse-and-execute path.
package shell

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"cuelang.org/go/cue"
	"github.com/chzyer/readline"

	"github.com/qjcg/arcadia/exp/terebra/internal/cueutil"
	"github.com/qjcg/arcadia/exp/terebra/internal/engine"
)

const historyFile = ".terebra_history"

// Shell is the interactive REPL: one persistent engine plus the
// readline experience and the job table.
type Shell struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer

	engine   *engine.Engine
	aliases  map[string]string
	jobs     []*Job
	jobSeq   int
	exitCode int
	mu       sync.Mutex

	rl    *readline.Instance
	fuzzy bool // Ctrl+R fuzzy search flag

	// stdin pipe so we can interrupt readline (Ctrl+R, external
	// commands). The pipe is created once and reused.
	stdinR          *os.File // read end (readline reads from this)
	stdinW          *os.File // write end (copier goroutine writes to this)
	stdinDone       chan struct{}
	stopR, stopW    *os.File // self-pipe to stop the copier
	stdinNeedsReset bool
}

// New builds the REPL shell with its engine and job control.
func New() *Shell {
	s := &Shell{
		Stdin:   os.Stdin,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
		aliases: make(map[string]string),
	}
	e, err := engine.New(engine.Options{
		Stdin:  s.Stdin,
		Stdout: s.Stdout,
		Stderr: s.Stderr,
		// Job control owns external commands (issue #12 layer).
		External: s.execJob,
	})
	if err != nil {
		panic(err)
	}
	s.engine = e
	return s
}

// Run starts the interactive REPL.
func Run() error {
	sh := New()
	return sh.Repl()
}

// Repl is the readline loop: accumulate lines until input is
// complete, evaluate through the engine, repeat.
func (s *Shell) Repl() error {
	histPath := ""
	if home := os.Getenv("HOME"); home != "" {
		histPath = filepath.Join(home, historyFile)
	}

	// Set up stdin pipe so we can interrupt readline on Ctrl+R
	s.setupStdinPipe()

	rl, err := readline.NewEx(s.newReadlineConfig(histPath))
	if err != nil {
		return fmt.Errorf("readline: %v", err)
	}
	s.rl = rl
	defer rl.Close()

	s.loadRc()

	ctx := context.Background()
	var buf strings.Builder
	for {
		if buf.Len() > 0 {
			rl.SetPrompt("> ")
		} else {
			// Computing the prompt may run commands (PS1 command
			// substitutions), which pause the stdin pump so they can
			// read the terminal.
			rl.SetPrompt(s.prompt())
		}

		// The pump must be running before readline blocks on its pipe:
		// resume it after whatever ran above, not before.
		if s.stdinNeedsReset {
			s.stdinNeedsReset = false
			s.resumeStdinPipe()
		}

		line, err := rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt {
				buf.Reset()
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
					newRL, err := readline.NewEx(s.newReadlineConfig(histPath))
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
				s.closeStdinPipe()
				return nil
			}
			s.closeStdinPipe()
			return err
		}

		if buf.Len() > 0 {
			buf.WriteString("\n")
		}
		buf.WriteString(line)
		src := buf.String()
		if strings.TrimSpace(src) == "" {
			buf.Reset()
			continue
		}

		err = s.evalLine(ctx, src)
		if err != nil && engine.Incomplete(err) {
			// Unclosed quote or open block: read another line.
			continue
		}
		buf.Reset()
		switch {
		case err == nil:
			s.exitCode = 0
		case s.engine.Exited():
			code, _ := engine.ExitCode(err)
			s.closeStdinPipe()
			os.Exit(code)
		default:
			fmt.Fprintf(s.Stderr, "trb: %v\n", err)
			s.exitCode = 1
			if code, ok := engine.ExitCode(err); ok {
				s.exitCode = code
			}
		}
	}
}

// evalLine runs one REPL input chunk: REPL-only commands are served
// locally, everything else goes through the shared engine.
func (s *Shell) evalLine(ctx context.Context, line string) error {
	trimmed := strings.TrimSpace(line)
	if rest, ok := strings.CutPrefix(trimmed, "--explain"); ok {
		fmt.Fprintf(s.Stdout, "# would execute: %s\n", strings.TrimSpace(rest))
		return nil
	}
	fields := strings.Fields(trimmed)
	if len(fields) == 0 {
		return nil
	}
	args := splitWords(strings.TrimLeft(trimmed[len(fields[0]):], " \t"))
	switch fields[0] {
	case "alias":
		s.builtinAlias(args)
		return nil
	case "unalias":
		s.builtinUnalias(args)
		return nil
	case "jobs":
		s.printJobs(s.Stdout)
		return nil
	case "fg":
		s.fgJob(args)
		return nil
	case "bg":
		s.bgJob(args)
		return nil
	case "history":
		s.builtinHistory(args)
		return nil
	case "state":
		if len(args) > 0 && args[0] == "load" {
			return s.stateLoad(args[1:])
		}
	}
	return s.withTerminal(func() error {
		return s.engine.Eval(ctx, s.expandAlias(trimmed))
	})
}

// withTerminal runs fn with the stdin pump paused so child processes
// can read the terminal directly, then marks the pump for restart
// before the next Readline. Pump ownership stays on the REPL loop
// goroutine: background commands run on interp goroutines and must
// never pause it themselves, or readline starves.
func (s *Shell) withTerminal(fn func() error) error {
	s.pauseStdinPipe()
	s.stdinNeedsReset = true
	return fn()
}

// splitWords splits s on unquoted whitespace and strips one layer of
// quoting from each word, mirroring how bash parses builtin arguments
// as shell words. Adjacent quotes concatenate (`'it”s'` is `it's`).
func splitWords(s string) []string {
	var words []string
	var b strings.Builder
	var quote byte
	started := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case quote == '\'':
			if c == '\'' {
				quote = 0
			} else {
				b.WriteByte(c)
			}
		case quote == '"':
			switch c {
			case '"':
				quote = 0
			case '\\':
				if i+1 < len(s) {
					i++
				}
				b.WriteByte(s[i])
			default:
				b.WriteByte(c)
			}
		case c == '\'' || c == '"':
			quote = c
			started = true
		case c == '\\' && i+1 < len(s):
			i++
			b.WriteByte(s[i])
			started = true
		case c == ' ' || c == '\t':
			if started {
				words = append(words, b.String())
				b.Reset()
				started = false
			}
		default:
			b.WriteByte(c)
			started = true
		}
	}
	if started {
		words = append(words, b.String())
	}
	return words
}

// expandAlias expands a leading command word that matches an alias.
// Trailing-space aliases do not chain-expand; that is a documented
// limitation of the REPL layer.
func (s *Shell) expandAlias(line string) string {
	name, rest, hasRest := strings.Cut(line, " ")
	alias, ok := s.aliases[name]
	if !ok {
		return line
	}
	if hasRest {
		return alias + " " + rest
	}
	return alias
}

// builtinAlias defines or lists aliases (bash output format).
func (s *Shell) builtinAlias(args []string) {
	if len(args) == 0 {
		for name, val := range s.aliases {
			fmt.Fprintf(s.Stdout, "alias %s='%s'\n", name, val)
		}
		return
	}
	for _, arg := range args {
		name, val, ok := strings.Cut(arg, "=")
		if !ok {
			if val, ok := s.aliases[name]; ok {
				fmt.Fprintf(s.Stdout, "alias %s='%s'\n", name, val)
			}
			continue
		}
		s.aliases[name] = val
	}
}

// builtinUnalias removes aliases.
func (s *Shell) builtinUnalias(args []string) {
	for _, name := range args {
		delete(s.aliases, name)
	}
}

// builtinHistory prints recent history, most recent last.
func (s *Shell) builtinHistory(args []string) {
	lines := s.readHistory()
	n := len(lines)
	if len(args) > 0 {
		fmt.Sscanf(args[0], "%d", &n)
	}
	for i := min(n, len(lines)) - 1; i >= 0; i-- {
		fmt.Fprintf(s.Stdout, "%5d  %s\n", len(lines)-i, lines[i])
	}
}

// stateLoad restores variables from a state file into the engine.
func (s *Shell) stateLoad(args []string) error {
	path := "terebra-state.cue"
	if len(args) > 0 {
		path = args[0]
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	ctx := cueutil.NewContext()
	v := cueutil.CompileString(ctx, string(data))
	if err := cueutil.Err(v); err != nil {
		return err
	}
	vals := map[string]string{}
	cueutil.WalkFields(cueutil.LookupPath(v, "vars"),
		func(name string, val cue.Value) bool {
			if str, err := val.String(); err == nil {
				vals[name] = str
			}
			return true
		})
	s.engine.SetVars(vals)
	fmt.Fprintf(s.Stdout, "state loaded from %s\n", path)
	return nil
}

// captureCommandOutput evaluates a command and returns its stdout,
// used for PS1 command substitutions.
func (s *Shell) captureCommandOutput(cmdStr string) (string, error) {
	var out string
	err := s.withTerminal(func() error {
		var cErr error
		out, cErr = s.engine.Capture(context.Background(), cmdStr)
		return cErr
	})
	return strings.TrimRight(out, "\n"), err
}

// prompt renders PS1 (template vars and command substitutions).
func (s *Shell) prompt() string {
	// Check for PS1 customization — engine vars first, then env.
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
		// only one.
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

// getVar reads a shell variable from the engine, falling back to the
// process environment.
func (s *Shell) getVar(name string) string {
	if v, ok := s.engine.Vars()[name]; ok {
		return v.String()
	}
	return os.Getenv(name)
}

// expandPromptCmd evaluates $(...) command substitutions in a prompt.
func (s *Shell) expandPromptCmd(input string) string {
	var b strings.Builder
	for i := 0; i < len(input); {
		if strings.HasPrefix(input[i:], "$(") {
			depth := 1
			j := i + 2
			for j < len(input) && depth > 0 {
				switch input[j] {
				case '(':
					depth++
				case ')':
					depth--
				}
				j++
			}
			if depth == 0 {
				if out, err := s.captureCommandOutput(input[i+2 : j-1]); err == nil {
					b.WriteString(out)
				}
				i = j
				continue
			}
		}
		b.WriteByte(input[i])
		i++
	}
	return b.String()
}

// stripANSI removes ANSI escape codes (both CSI and OSC sequences).
func stripANSI(s string) string {
	var result strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\033' {
			if i+1 < len(s) && s[i+1] == '[' {
				// CSI sequence: \033[...<letter>
				j := i + 2
				for j < len(s) && !((s[j] >= 'A' && s[j] <= 'Z') || (s[j] >= 'a' && s[j] <= 'z')) {
					j++
				}
				i = j + 1
				continue
			}
			if i+1 < len(s) && s[i+1] == ']' {
				// OSC sequence: \033]...(\a or \033\\)
				j := i + 2
				for j < len(s) {
					if s[j] == '\a' {
						j++
						break
					}
					if s[j] == '\033' && j+1 < len(s) && s[j+1] == '\\' {
						j += 2
						break
					}
					j++
				}
				i = j
				continue
			}
			i++
			continue
		}
		result.WriteByte(s[i])
		i++
	}
	return result.String()
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

// newReadlineConfig creates a fresh readline.Config for the current
// stdin pipe.
func (s *Shell) newReadlineConfig(histPath string) *readline.Config {
	return &readline.Config{
		Prompt:              "",
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
}

// setupStdinPipe creates a pipe and starts a goroutine that copies
// from os.Stdin into the pipe. Readline reads from the read end so we
// can interrupt it without closing os.Stdin.
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
	s.Stdin = r // the copier is the sole os.Stdin reader
	s.startStdinCopier()
}

// startStdinCopier launches the goroutine that copies os.Stdin into
// the pipe. It can be paused and resumed without recreating the pipe.
func (s *Shell) startStdinCopier() {
	s.stdinDone = make(chan struct{})
	stopR, stopW, err := os.Pipe()
	if err != nil {
		stopR, stopW = nil, nil
	}
	s.stopR = stopR
	s.stopW = stopW
	w := s.stdinW

	go func() {
		defer close(s.stdinDone)
		buf := make([]byte, 4096)
		stdinFd := int(os.Stdin.Fd())
		for {
			// Wait for input on os.Stdin or the stop pipe. A terminal fd
			// is not unblocked by close() in raw mode, so the copier is
			// stopped via this self-pipe.
			rfds := &syscall.FdSet{}
			FD_SET(stdinFd, rfds)
			if stopR != nil {
				FD_SET(int(stopR.Fd()), rfds)
			}
			maxfd := stdinFd
			if stopR != nil && int(stopR.Fd()) > maxfd {
				maxfd = int(stopR.Fd())
			}
			n, err := syscall.Select(maxfd+1, rfds, nil, nil, nil)
			if err != nil {
				if err == syscall.EINTR {
					continue
				}
				return
			}
			if n == 0 {
				continue
			}
			if stopR != nil && FD_ISSET(int(stopR.Fd()), rfds) {
				return
			}
			if FD_ISSET(stdinFd, rfds) {
				n, err := os.Stdin.Read(buf)
				if err != nil {
					return
				}
				if _, err := w.Write(buf[:n]); err != nil {
					return
				}
			}
		}
	}()
}

// FD_SET sets the bit for fd in the fd set.
func FD_SET(fd int, set *syscall.FdSet) {
	set.Bits[fd/64] |= 1 << (uint(fd) % 64)
}

// FD_ISSET reports whether the bit for fd is set.
func FD_ISSET(fd int, set *syscall.FdSet) bool {
	return set.Bits[fd/64]&(1<<(uint(fd)%64)) != 0
}

// pauseStdinPipe stops the stdin copier and restores direct stdin so
// an external command can read the terminal. Buffered pipe input is
// preserved.
func (s *Shell) pauseStdinPipe() {
	if s.stdinDone != nil {
		if s.stopW != nil {
			s.stopW.Write([]byte{0})
		}
		<-s.stdinDone
		s.stdinDone = nil
	}
	s.Stdin = os.Stdin
}

// resumeStdinPipe restarts the stdin copier on the existing pipe.
func (s *Shell) resumeStdinPipe() {
	if s.stdinR != nil && s.stdinR != os.Stdin {
		s.startStdinCopier()
		s.Stdin = s.stdinR
	}
}

// closeStdinPipe closes the stdin pipe and waits for the copier.
func (s *Shell) closeStdinPipe() {
	if s.stdinDone != nil {
		if s.stopW != nil {
			s.stopW.Write([]byte{0})
		}
		<-s.stdinDone
		s.stdinDone = nil
	}
	if s.stdinW != nil {
		s.stdinW.Close()
	}
	if s.stdinR != nil && s.stdinR != os.Stdin {
		s.stdinR.Close()
	}
	if s.stopR != nil {
		s.stopR.Close()
	}
	if s.stopW != nil {
		s.stopW.Close()
	}
	s.stdinR = nil
	s.stdinW = nil
	s.stdinDone = nil
	s.stopR = nil
	s.stopW = nil
	s.Stdin = os.Stdin
}

// readHistory reads the history file, most recent first.
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
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
		lines[i], lines[j] = lines[j], lines[i]
	}
	return lines
}

// loadRc loads ~/.terebrarc if it exists, running each line through
// the same eval path as typed input so assignments (PS1) and aliases
// land where the REPL reads them.
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
	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		_ = s.evalLine(context.Background(), line)
	}
}
