package shell

import (
	"fmt"
	"io"
	"os"
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
	arrays   map[string][]string
	assoc    map[string]map[string]string
	funcs    map[string][]script.Stmt
	plugins  *plugin.Registry
	jobs     []*Job
	jobSeq   int
	exitCode int
	rl       *readline.Instance
	interp   *script.Interpreter
}

func New() *Shell {
	s := &Shell{
		Stdin:    os.Stdin,
		Stdout:   os.Stdout,
		Stderr:   os.Stderr,
		builtins: builtins.New(),
		vars:     make(map[string]string),
		arrays:   make(map[string][]string),
		assoc:    make(map[string]map[string]string),
		funcs:    make(map[string][]script.Stmt),
		plugins:  plugin.New(),
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
	s.builtins.Register("plugin", plugin.PluginBuiltin(s.plugins))
	s.builtins.Register("state", func(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
		return s.builtinState(args, stdin, stdout, stderr)
	})
	s.builtins.Register("source", func(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
		return s.builtinSource(args, stdin, stdout, stderr)
	})
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
		if line == "" {
			continue
		}

		pipe, err := parser.Parse(line)
		if err != nil {
			fmt.Fprintf(s.Stderr, "trb: parse error: %v\n", err)
			continue
		}

		if err := s.ExecutePipeline(pipe); err != nil {
			fmt.Fprintf(s.Stderr, "trb: error: %v\n", err)
			s.exitCode = 1
		} else {
			s.exitCode = 0
		}
	}
}

func (s *Shell) filterInput(r rune) (rune, bool) {
	// Allow Ctrl-C to pass through
	return r, true
}

func (s *Shell) prompt() string {
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
