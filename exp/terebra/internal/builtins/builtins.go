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

	// Handle cd -
	if dir == "-" {
		dir = os.Getenv("OLDPWD")
		if dir == "" {
			fmt.Fprintln(stderr, "cd: OLDPWD not set")
			return 1
		}
	}

	// Save current directory as OLDPWD before changing
	oldwd, _ := os.Getwd()
	if oldwd != "" {
		os.Setenv("OLDPWD", oldwd)
	}

	if err := os.Chdir(dir); err != nil {
		fmt.Fprintf(stderr, "cd: %v\n", err)
		return 1
	}
	wd, err := os.Getwd()
	if err == nil {
		os.Setenv("PWD", wd)
	}

	// Print the directory when using cd - (like bash does)
	if len(args) > 0 && args[0] == "-" {
		fmt.Fprintln(stdout, wd)
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
	// Return a special code that signals the REPL to exit cleanly.
	// We use -1 as the sentinel (not a valid exit code).
	return -1
}

var builtinHelp = map[string]string{
	"cd":       "  cd [dir]       Change directory (defaults to $HOME)\n",
	"pwd":      "  pwd            Print working directory\n",
	"echo":     "  echo [args]    Print arguments\n",
	"exit":     "  exit [code]    Exit the shell\n",
	"help":     "  help [cmd]     Show help for a builtin or topic\n",
	"type":     "  type [cmd]     Show how a command would be interpreted\n",
	"which":    "  which [cmd]    Locate a command in PATH\n",
	"export":   "  export [name[=value]]  Set or list environment variables\n",
	"unset":    "  unset <name>   Remove a variable\n",
	"set":      "  set [-x|+x] [-o vi|emacs]  List vars or toggle debug/options\n",
	"exec":     "  exec [-a name] <command> [args...]  Replace shell with command\n",
	"readonly": "  readonly [name=value]  Mark variables as read-only\n",
	"alias":    "  alias [name=value]  Define or list command aliases\n",
	"unalias":  "  unalias <name>  Remove an alias\n",
	"history":  "  history [n]    Show command history\n",
	"source":   "  source <file>  Execute a script file\n",
	"jobs":     "  jobs           List background jobs\n",
	"fg":       "  fg [job]       Bring a job to foreground\n",
	"bg":       "  bg [job]       Continue a stopped job in background\n",
	"drill":    "  drill <sub>    Drill into structured data (cue, fs, proc, net)\n",
	"cue":      "  cue <sub>      CUE operations (eval, vet, export, def, fmt, trim)\n",
	"plugin":   "  plugin <sub>   Manage plugins (load, list)\n",
	"state":    "  state [save|load]  Export/import shell state as CUE\n",
}

func helpBuiltin(name string, stdout, stderr io.Writer) int {
	if help, ok := builtinHelp[name]; ok {
		fmt.Fprintf(stdout, "%s\n", help)
		return 0
	}
	// Check if it's a topic covered in the general help
	topics := map[string]string{
		"pipes":     "  cmd1 | cmd2           Pipe stdout to stdin\n  cmd1 |> cmd2          Auger pipe: pass CUE values between commands\n",
		"redirects": "  cmd > file            Redirect stdout to file\n  cmd >> file           Append stdout to file\n  cmd < file            Read stdin from file\n  cmd 2> file           Redirect stderr to file\n  cmd 2>> file          Append stderr to file\n  cmd 2>&1              Redirect stderr to stdout\n",
		"chaining":  "  cmd1 && cmd2          Run cmd2 only if cmd1 succeeds\n  cmd1 || cmd2          Run cmd2 only if cmd1 fails\n  cmd1 ; cmd2           Run cmd1 then cmd2 regardless\n",
		"quoting":   "  'single quotes'      Preserve all characters literally\n  \"double quotes\"      Preserve most, allow $, `, and \\\n  \\escape              Escape next character\n",
		"vars":      "  $VAR                 Expand variable\n  ${VAR}               Expand variable with braces\n  $?                   Last exit code\n  $$                   Shell PID\n  arr=(a b c)          Indexed array\n  arr[key]=val         Associative array\n  ${arr[idx]}          Array element\n  $arr[idx]            Array element (without braces)\n  ${#arr[@]}           Array length\n  ${!arr[@]}           Array keys\n",
		"strings":   "  ${var:off:len}       Substring\n  ${var/old/new}       Replace first match\n  ${var//old/new}      Replace all matches\n  ${var#pat}           Remove shortest prefix\n  ${var##pat}          Remove longest prefix\n  ${var%pat}           Remove shortest suffix\n  ${var%%pat}          Remove longest suffix\n  ${var^} ${var^^}     Uppercase first/all\n  ${var,} ${var,,}     Lowercase first/all\n",
		"expansion": "  $(cmd)               Command substitution\n  `cmd`                Backtick command substitution\n  $((expr))            Arithmetic expansion\n  {a,b,c}              Brace expansion\n  {1..5}               Numeric range\n  {a..f}               Alpha range\n  *.go                 Glob matching\n  **/go.mod            Recursive glob\n",
	}
	if topic, ok := topics[name]; ok {
		fmt.Fprint(stdout, topic)
		return 0
	}
	fmt.Fprintf(stderr, "help: no help for %q\n", name)
	return 1
}

func helpHandler(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) > 0 {
		return helpBuiltin(args[0], stdout, stderr)
	}

	fmt.Fprint(stdout, "Terebra -- auger shell\n\n")
	fmt.Fprint(stdout, "Built-in commands:\n")
	fmt.Fprint(stdout, "  cd [dir]       Change directory (defaults to $HOME)\n")
	fmt.Fprint(stdout, "  pwd            Print working directory\n")
	fmt.Fprint(stdout, "  echo [args]    Print arguments\n")
	fmt.Fprint(stdout, "  exec [-a name] <command> [args]  Replace shell with a command\n")
	fmt.Fprint(stdout, "  exit [code]    Exit the shell\n")
	fmt.Fprint(stdout, "  help           Show this help\n")
	fmt.Fprint(stdout, "  type [cmd]     Show how a command would be interpreted\n")
	fmt.Fprint(stdout, "  which [cmd]    Locate a command in PATH\n")
	fmt.Fprint(stdout, "  export         Set or list environment variables\n")
	fmt.Fprint(stdout, "  unset          Remove a variable\n")
	fmt.Fprint(stdout, "  set [-x|+x]    List vars or toggle debug mode\n")
	fmt.Fprint(stdout, "  readonly       Mark variables as read-only\n")
	fmt.Fprint(stdout, "  alias          Define or list command aliases\n")
	fmt.Fprint(stdout, "  unalias        Remove an alias\n")
	fmt.Fprint(stdout, "  history [n]    Show command history\n")
	fmt.Fprint(stdout, "  source <file>  Execute a script file\n")
	fmt.Fprint(stdout, "  jobs           List background jobs\n")
	fmt.Fprint(stdout, "  fg [job]       Bring a job to foreground\n")
	fmt.Fprint(stdout, "  bg [job]       Continue a stopped job in background\n")
	fmt.Fprint(stdout, "  drill          Drill into structured data (cue, fs, proc, net)\n")
	fmt.Fprint(stdout, "  cue <sub>      CUE operations (eval, vet, export, def, fmt, trim)\n")
	fmt.Fprint(stdout, "  plugin         Manage plugins (load, list)\n")
	fmt.Fprint(stdout, "  state          Export/import shell state as CUE\n")
	fmt.Fprint(stdout, "\n")
	fmt.Fprint(stdout, "Pipes and redirects:\n")
	fmt.Fprint(stdout, "  cmd1 | cmd2           Pipe stdout to stdin\n")
	fmt.Fprint(stdout, "  cmd1 |> cmd2          Auger pipe: pass CUE values between commands\n")
	fmt.Fprint(stdout, "  cmd1 |>json|cue|yaml  Encode CUE output as JSON/CUE/YAML\n")
	fmt.Fprint(stdout, "  cmd1 && cmd2          Run cmd2 only if cmd1 succeeds\n")
	fmt.Fprint(stdout, "  cmd1 || cmd2          Run cmd2 only if cmd1 fails\n")
	fmt.Fprint(stdout, "  cmd1 ; cmd2           Run cmd1 then cmd2 regardless\n")
	fmt.Fprint(stdout, "  cmd > file            Redirect stdout to file\n")
	fmt.Fprint(stdout, "  cmd >> file           Append stdout to file\n")
	fmt.Fprint(stdout, "  cmd < file            Read stdin from file\n")
	fmt.Fprint(stdout, "  cmd 2> file           Redirect stderr to file\n")
	fmt.Fprint(stdout, "  cmd 2>> file          Append stderr to file\n")
	fmt.Fprint(stdout, "  cmd 2>&1              Redirect stderr to stdout\n")
	fmt.Fprint(stdout, "  cmd << EOF            Heredoc: read stdin until delimiter\n")
	fmt.Fprint(stdout, "\n")
	fmt.Fprint(stdout, "Job control:\n")
	fmt.Fprint(stdout, "  cmd &                Run command in background\n")
	fmt.Fprint(stdout, "\n")
	fmt.Fprint(stdout, "Variables:\n")
	fmt.Fprint(stdout, "  $VAR                 Expand variable\n")
	fmt.Fprint(stdout, "  ${VAR}               Expand variable with braces\n")
	fmt.Fprint(stdout, "  ${var:off:len}       Substring\n")
	fmt.Fprint(stdout, "  ${var/old/new}       Replace first match\n")
	fmt.Fprint(stdout, "  ${var//old/new}      Replace all matches\n")
	fmt.Fprintf(stdout, "  ${var#pat}           Remove shortest prefix\n")
	fmt.Fprintf(stdout, "  ${var##pat}          Remove longest prefix\n")
	fmt.Fprintf(stdout, "  ${var%%%%pat}         Remove longest suffix\n")
	fmt.Fprintf(stdout, "  ${var%%pat}          Remove shortest suffix\n")
	fmt.Fprint(stdout, "  ${var^} ${var^^}     Uppercase first/all\n")
	fmt.Fprint(stdout, "  ${var,} ${var,,}     Lowercase first/all\n")
	fmt.Fprint(stdout, "  $?                   Last exit code\n")
	fmt.Fprint(stdout, "  $$                   Shell PID\n")
	fmt.Fprint(stdout, "  arr=(a b c)          Indexed array\n")
	fmt.Fprint(stdout, "  arr[key]=val         Associative array\n")
	fmt.Fprint(stdout, "  ${arr[idx]}          Array element\n")
	fmt.Fprint(stdout, "  ${#arr[@]}           Array length\n")
	fmt.Fprint(stdout, "  ${!arr[@]}           Array keys\n")
	fmt.Fprint(stdout, "\n")
	fmt.Fprint(stdout, "Expansion:\n")
	fmt.Fprint(stdout, "  $(cmd)               Command substitution\n")
	fmt.Fprint(stdout, "  `cmd`                Backtick command substitution\n")
	fmt.Fprint(stdout, "  $((expr))            Arithmetic expansion\n")
	fmt.Fprint(stdout, "  {a,b,c}              Brace expansion\n")
	fmt.Fprint(stdout, "  {1..5}               Numeric range\n")
	fmt.Fprint(stdout, "  {a..f}               Alpha range\n")
	fmt.Fprint(stdout, "  *.go                 Glob matching\n")
	fmt.Fprint(stdout, "  **/go.mod            Recursive glob\n")
	fmt.Fprint(stdout, "\n")
	fmt.Fprint(stdout, "Quoting:\n")
	fmt.Fprint(stdout, "  'single quotes'      Preserve all characters literally\n")
	fmt.Fprint(stdout, "  \"double quotes\"      Preserve most, allow $, `, and \\\n")
	fmt.Fprint(stdout, "  \\escape              Escape next character\n")
	fmt.Fprint(stdout, "\n")
	fmt.Fprint(stdout, "Shell options:\n")
	fmt.Fprint(stdout, "  set -x               Enable debug mode (trace commands)\n")
	fmt.Fprint(stdout, "  set +x               Disable debug mode\n")
	fmt.Fprint(stdout, "  set -o vi            Enable vi keybindings\n")
	fmt.Fprint(stdout, "  set -o emacs         Enable emacs keybindings (default)\n")
	fmt.Fprint(stdout, "\n")
	fmt.Fprint(stdout, "Prompt customization:\n")
	fmt.Fprint(stdout, "  PS1='template'       Custom prompt with {{.}} {{exit}} {{exitcode}}\n")
	fmt.Fprint(stdout, "  PS1='$(oh-my-posh)'  Use oh-my-posh or any command output\n")
	fmt.Fprint(stdout, "\n")
	fmt.Fprint(stdout, "Special keys:\n")
	fmt.Fprint(stdout, "  Ctrl+R               Fuzzy history search\n")
	fmt.Fprint(stdout, "  --explain <cmd>      Dry-run: show what a command would do\n")
	fmt.Fprint(stdout, "  $HOME/.terebrarc     Startup script sourced on launch\n")
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
	case "cd", "exec", "pwd", "echo", "exit", "help", "type", "which",
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
