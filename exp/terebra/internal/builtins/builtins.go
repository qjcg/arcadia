package builtins

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Handler func(args []string, stdin io.Reader, stdout, stderr io.Writer) int

type Registry struct {
	cmds map[string]Handler
}

func New() *Registry {
	r := &Registry{cmds: make(map[string]Handler)}
	r.register("cd", cdHandler)
	r.register("pwd", pwdHandler)
	r.register("echo", echoHandler)
	r.register("exit", exitHandler)
	r.register("help", helpHandler)
	r.register("type", typeHandler)
	r.register("which", whichHandler)
	r.register("ls", lsHandler)
	r.register("export", exportHandler)
	r.register("unset", unsetHandler)
	r.register("set", setHandler)
	return r
}

func (r *Registry) Register(name string, h Handler) {
	r.cmds[name] = h
}

func (r *Registry) register(name string, h Handler) {
	r.cmds[name] = h
}

func (r *Registry) Lookup(name string) (Handler, bool) {
	h, ok := r.cmds[name]
	return h, ok
}

func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.cmds))
	for n := range r.cmds {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

func cdHandler(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	dir := os.Getenv("HOME")
	if len(args) > 0 {
		dir = args[0]
	}
	if dir == "" {
		dir = "/"
	}
	if err := os.Chdir(dir); err != nil {
		fmt.Fprintf(stderr, "cd: %v\n", err)
		return 1
	}
	wd, err := os.Getwd()
	if err == nil {
		os.Setenv("PWD", wd)
	}
	return 0
}

func pwdHandler(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "pwd: %v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, wd)
	return 0
}

func echoHandler(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fmt.Fprintln(stdout, strings.Join(args, " "))
	return 0
}

func exitHandler(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	code := 0
	if len(args) > 0 {
		if _, err := fmt.Sscanf(args[0], "%d", &code); err != nil {
			fmt.Fprintf(stderr, "exit: invalid exit code: %s\n", args[0])
			code = 1
		}
	}
	os.Exit(code)
	return 0
}

func helpHandler(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fmt.Fprint(stdout, "Terebra -- auger shell\n\n")
	fmt.Fprint(stdout, "Built-in commands:\n")
	fmt.Fprint(stdout, "  cd [dir]     Change directory (defaults to $HOME)\n")
	fmt.Fprint(stdout, "  pwd          Print working directory\n")
	fmt.Fprint(stdout, "  echo [args]  Print arguments\n")
	fmt.Fprint(stdout, "  exit [code]  Exit the shell\n")
	fmt.Fprint(stdout, "  help         Show this help\n")
	fmt.Fprint(stdout, "  type [cmd]   Show how a command would be interpreted\n")
	fmt.Fprint(stdout, "  which [cmd]  Locate a command in PATH\n")
	fmt.Fprint(stdout, "  ls [path]    List directory contents\n")
	fmt.Fprint(stdout, "  export       Set or list environment variables\n")
	fmt.Fprint(stdout, "  unset        Remove a variable\n")
	fmt.Fprint(stdout, "  set          List shell variables\n")
	fmt.Fprint(stdout, "  jobs         List background jobs\n")
	fmt.Fprint(stdout, "  fg [job]     Bring a job to foreground\n")
	fmt.Fprint(stdout, "  bg [job]     Continue a stopped job in background\n")
	fmt.Fprint(stdout, "  drill        Drill into structured data\n")
	fmt.Fprint(stdout, "\n")
	fmt.Fprint(stdout, "Pipes and redirects:\n")
	fmt.Fprint(stdout, "  cmd1 | cmd2         Pipe stdout of cmd1 to stdin of cmd2\n")
	fmt.Fprint(stdout, "  cmd1 |> cmd2        Auger pipe: parse output as CUE, pipe to cmd2\n")
	fmt.Fprint(stdout, "  cmd > file          Redirect stdout to file\n")
	fmt.Fprint(stdout, "  cmd >> file         Append stdout to file\n")
	fmt.Fprint(stdout, "  cmd < file          Read stdin from file\n")
	fmt.Fprint(stdout, "  cmd 2> file         Redirect stderr to file\n")
	fmt.Fprint(stdout, "  cmd 2>> file        Append stderr to file\n")
	fmt.Fprint(stdout, "  cmd 2>&1            Redirect stderr to stdout\n")
	fmt.Fprint(stdout, "\n")
	fmt.Fprint(stdout, "Job control:\n")
	fmt.Fprint(stdout, "  cmd &               Run command in background\n")
	fmt.Fprint(stdout, "\n")
	fmt.Fprint(stdout, "Variables:\n")
	fmt.Fprint(stdout, "  $VAR                Expand variable\n")
	fmt.Fprint(stdout, "  ${VAR}              Expand variable with braces\n")
	fmt.Fprint(stdout, "  $?                  Last exit code\n")
	fmt.Fprint(stdout, "  $$                  Shell PID\n")
	fmt.Fprint(stdout, "\n")
	fmt.Fprint(stdout, "Quoting:\n")
	fmt.Fprint(stdout, "  'single quotes'     Preserve all characters literally\n")
	fmt.Fprint(stdout, "  \"double quotes\"     Preserve most characters, allow $ and `\n")
	fmt.Fprint(stdout, "  \\escape             Escape next character\n")
	return 0
}

func typeHandler(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	for _, arg := range args {
		if isBuiltin(arg) {
			fmt.Fprintf(stdout, "%s is a shell builtin\n", arg)
			continue
		}
		if path := findInPath(arg); path != "" {
			fmt.Fprintf(stdout, "%s is %s\n", arg, path)
			continue
		}
		fmt.Fprintf(stderr, "type: %s: not found\n", arg)
		return 1
	}
	return 0
}

func whichHandler(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	for _, arg := range args {
		path := findInPath(arg)
		if path != "" {
			fmt.Fprintln(stdout, path)
			continue
		}
		fmt.Fprintf(stderr, "which: %s: not found in PATH\n", arg)
		return 1
	}
	return 0
}

func lsHandler(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	dir := "."
	if len(args) > 0 {
		dir = args[0]
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(stderr, "ls: %v\n", err)
		return 1
	}

	var names []string
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() {
			name = name + "/"
		}
		names = append(names, name)
	}

	sort.Strings(names)
	for _, name := range names {
		fmt.Fprintln(stdout, name)
	}
	return 0
}

func exportHandler(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		for _, env := range os.Environ() {
			fmt.Fprintln(stdout, env)
		}
		return 0
	}
	for _, arg := range args {
		if before, after, ok := strings.Cut(arg, "="); ok {
			os.Setenv(before, after)
		}
	}
	return 0
}

func unsetHandler(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	for _, arg := range args {
		os.Unsetenv(arg)
	}
	return 0
}

func setHandler(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	for _, env := range os.Environ() {
		fmt.Fprintln(stdout, env)
	}
	return 0
}

func isBuiltin(name string) bool {
	switch name {
	case "cd", "pwd", "echo", "exit", "help", "type", "which", "ls",
		"export", "unset", "set", "jobs", "fg", "bg", "drill":
		return true
	}
	return false
}

func findInPath(name string) string {
	path := os.Getenv("PATH")
	for _, dir := range filepath.SplitList(path) {
		full := filepath.Join(dir, name)
		if fi, err := os.Stat(full); err == nil && !fi.IsDir() {
			return full
		}
	}
	return ""
}

func FindInPath(name string) string {
	return findInPath(name)
}
