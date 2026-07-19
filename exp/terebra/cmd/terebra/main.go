package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/qjcg/arcadia/exp/terebra/internal/shell"
)

const version = "0.1.0"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "build":
			os.Exit(buildCmd(os.Args[2:]))

		case "-c":
			// Inline script
			if len(os.Args) < 3 {
				fmt.Fprintln(os.Stderr, "-c requires a script argument")
				os.Exit(1)
			}
			if err := shell.RunScriptFromString(os.Args[2]); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			return

		case "--version", "-v":
			fmt.Printf("terebra %s\n", version)
			return

		case "--explain":
			// Dry-run: show what the command would do
			if len(os.Args) < 3 {
				fmt.Fprintln(os.Stderr, "--explain requires a command")
				os.Exit(1)
			}
			// Build a script that the shell can parse: --explain <cmd> <args>
			var script strings.Builder
			script.WriteString("--explain ")
			script.WriteString(os.Args[2])
			for _, a := range os.Args[3:] {
				script.WriteString(" ")
				script.WriteString(a)
			}
			if err := shell.RunScriptFromString(script.String()); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			return

		case "help", "--help", "-h":
			fmt.Println("Usage: terebra [command] [script]")
			fmt.Println()
			fmt.Println("Commands:")
			fmt.Println("  terebra              Start the interactive REPL")
			fmt.Println("  terebra <script>     Execute a .trb script")
			fmt.Println("  terebra build <script> [output]  Compile a .trb script to a binary")
			fmt.Println("  terebra --version    Show version")
			fmt.Println("  terebra --help       Show this help")
			return

		default:
			// Script mode: execute the given file
			if !strings.HasPrefix(os.Args[1], "-") {
				if err := shell.RunScript(os.Args[1]); err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
				return
			}
			fmt.Fprintf(os.Stderr, "unknown flag: %s\n", os.Args[1])
			os.Exit(1)
		}
	}

	// REPL mode
	if err := shell.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
