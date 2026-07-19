package shell

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
	rl       *readline.Instance
	interp   *script.Interpreter
	debug    bool // set -x debug mode
	fuzzy    bool // Ctrl+R fuzzy search flag
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

	return sh.interp.ParseAndExec(content)
}

func (s *Shell) Repl() error {
	histPath := ""
	if home := os.Getenv("HOME"); home != "" {
		histPath = filepath.Join(home, historyFile)
	}

	cfg := &readline.Config{
		Prompt:              s.prompt(),
		HistoryFile:         histPath,
		HistoryLimit:        1000,
		AutoComplete:        s.completer(),
		Stdin:               os.Stdin,
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
		rl.SetPrompt(s.prompt())

		line, err := rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt {
				continue
			}
			if err == io.EOF {
				fmt.Fprintln(s.Stdout)
				return nil
			}
			return err
		}

		line = strings.TrimSpace(line)

		// Handle fuzzy search triggered by Ctrl+R
		if s.fuzzy {
			s.fuzzy = false
			if line != "" {
				s.doFuzzySearch(line)
			}
			continue
		}

		if line == "" {
			continue
		}

		pipe, err := parser.ParseScript(line)
		if err != nil {
			fmt.Fprintf(s.Stderr, "trb: parse error: %v\n", err)
			continue
		}

		if err := s.ExecuteScript(pipe); err != nil {
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
		return r, false // consume the key
	}
	return r, true
}

// doFuzzySearch searches history entries containing the query and inserts the selected match.
func (s *Shell) doFuzzySearch(query string) {
	lines := s.readHistory()
	// Filter for entries containing the query (case-insensitive fuzzy match)
	ql := strings.ToLower(query)
	var matches []string
	seen := make(map[string]bool)
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if trimmed == "" || seen[trimmed] {
			continue
		}
		if strings.Contains(strings.ToLower(trimmed), ql) {
			matches = append(matches, trimmed)
			seen[trimmed] = true
		}
	}

	if len(matches) == 0 {
		fmt.Fprintf(s.Stderr, "\n(fuzzy-search): no matches for %q\n", query)
		return
	}

	if len(matches) == 1 {
		s.rl.Operation.SetBuffer(matches[0])
		s.rl.Operation.Refresh()
		return
	}

	// Show numbered list of matches on stderr
	fmt.Fprintf(s.Stderr, "\n(fuzzy-search) %q: %d matches\n", query, len(matches))
	showCount := min(len(matches), 20)
	for i, m := range matches[:showCount] {
		fmt.Fprintf(s.Stderr, "  %2d: %s\n", i+1, m)
	}
	if len(matches) > 20 {
		fmt.Fprintf(s.Stderr, "  ... (%d more)\n", len(matches)-20)
	}

	// Read selection from stdin using the readline instance (handles raw mode correctly)
	fmt.Fprintf(s.Stderr, "select (1-%d or Enter for 1): ", len(matches))
	var sel int
	// Set prompt to empty for the selection read
	s.rl.SetPrompt("")
	selLine, err := s.rl.Readline()
	if err != nil {
		// Default to first match
		s.rl.Operation.SetBuffer(matches[0])
		s.rl.Operation.Refresh()
		return
	}
	selStr := strings.TrimSpace(selLine)
	if selStr == "" {
		sel = 1
	} else {
		fmt.Sscanf(selStr, "%d", &sel)
	}
	if sel < 1 || sel > len(matches) {
		sel = 1
	}
	s.rl.Operation.SetBuffer(matches[sel-1])
	s.rl.Operation.Refresh()
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
