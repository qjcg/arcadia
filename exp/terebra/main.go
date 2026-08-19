package main

import (
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"

	"github.com/qjcg/arcadia/exp/terebra/internal/shell"
)

func getVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		return info.Main.Version
	}
	return "(devel)"
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run dispatches a CLI invocation with args (excluding argv[0]) to the
// appropriate subcommand. It returns a process exit code.
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		// REPL mode
		if err := shell.Run(); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return 0
	}

	switch args[0] {
	case "build":
		return buildCmd(args[1:])

	case "-c":
		// Inline script
		if len(args) < 2 {
			fmt.Fprintln(stderr, "-c requires a script argument")
			return 1
		}
		if err := shell.RunScriptFromString(args[1]); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return 0

	case "--version", "-v":
		fmt.Fprintf(stdout, "terebra %s\n", getVersion())
		return 0

	case "--explain":
		// Dry-run: show what the command would do
		if len(args) < 2 {
			fmt.Fprintln(stderr, "--explain requires a command")
			return 1
		}
		// Build a script that the shell can parse: --explain <cmd> <args>
		var script strings.Builder
		script.WriteString("--explain ")
		script.WriteString(args[1])
		for _, a := range args[2:] {
			script.WriteString(" ")
			script.WriteString(a)
		}
		if err := shell.RunScriptFromString(script.String()); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return 0

	case "help", "--help", "-h":
		fmt.Fprintln(stdout, "Usage: terebra [command] [script]")
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, "Commands:")
		fmt.Fprintln(stdout, "  terebra              Start the interactive REPL")
		fmt.Fprintln(stdout, "  terebra <script>     Execute a .trb script")
		fmt.Fprintln(stdout, "  terebra build <script> [output]  Compile a .trb script to a binary")
		fmt.Fprintln(stdout, "  terebra --version    Show version")
		fmt.Fprintln(stdout, "  terebra --help       Show this help")
		return 0

	default:
		// Script mode: execute the given file
		if !strings.HasPrefix(args[0], "-") {
			if err := shell.RunScript(args[0]); err != nil {
				fmt.Fprintln(stderr, err)
				return 1
			}
			return 0
		}
		fmt.Fprintf(stderr, "unknown flag: %s\n", args[0])
		return 1
	}
}
