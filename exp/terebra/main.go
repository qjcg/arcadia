package main

import (
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"

	"github.com/qjcg/arcadia/exp/terebra/internal/engine"
	"github.com/qjcg/arcadia/exp/terebra/internal/shell"
)

// scriptExit maps a script engine error to a process exit code,
// printing unexpected errors and honoring explicit exit statuses.
func scriptExit(err error, stderr io.Writer) int {
	if err != nil {
		if code, ok := engine.ExitCode(err); ok {
			return code
		}
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

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
		return scriptExit(engine.RunScriptFromString(args[1]), stderr)

	case "--version", "-v":
		fmt.Fprintf(stdout, "terebra %s\n", getVersion())
		return 0

	case "--explain":
		// Dry-run: show what the command would do
		if len(args) < 2 {
			fmt.Fprintln(stderr, "--explain requires a command")
			return 1
		}
		fmt.Fprintf(stdout, "# would execute: %s\n",
			strings.Join(args[1:], " "))
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
			return scriptExit(engine.RunScript(args[0]), stderr)
		}
		fmt.Fprintf(stderr, "unknown flag: %s\n", args[0])
		return 1
	}
}
