package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"

	"elbereth/internal/codegen"
	"elbereth/internal/expander"
	"elbereth/internal/lexer"
	"elbereth/internal/parser"
	"elbereth/internal/repl"
)

const Version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "-h", "--help", "help":
		printHelp()

	case "check":
		if len(os.Args) < 3 {
			fmt.Println("Usage: elbereth check <file>")
			os.Exit(1)
		}
		checkFile(os.Args[2])

	case "build":
		if len(os.Args) < 3 {
			fmt.Println("Usage: elbereth build <file> [-o output]")
			os.Exit(1)
		}
		flagSet := flag.NewFlagSet("build", flag.ContinueOnError)
		output := flagSet.String("o", "", "output file")
		flagSet.Parse(os.Args[3:])

		buildFile(os.Args[2], *output)

	case "run":
		if len(os.Args) < 3 {
			fmt.Println("Usage: elbereth run <file>")
			os.Exit(1)
		}
		runFile(os.Args[2])

	case "test":
		if len(os.Args) < 3 {
			fmt.Println("Usage: elbereth test <file> [go test flags]")
			os.Exit(1)
		}
		testFile(os.Args[2], os.Args[3:])

	case "gen":
		if len(os.Args) < 3 {
			fmt.Println("Usage: elbereth gen <file>")
			os.Exit(1)
		}
		genCode(os.Args[2])

	case "repl":
		r := repl.New(nil, nil)
		if err := r.Run(); err != nil {
			fmt.Printf("REPL error: %v\n", err)
			os.Exit(1)
		}

	case "version":
		fmt.Printf("elbereth version %s\n", Version)

	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println(`Elbereth - A practical Lisp dialect that compiles to Go

Usage: elbereth <command> [args]

Commands:
  check <file>           Check syntax of an Elbereth file
  build <file> [-o out]  Compile an Elbereth file to a binary
  run <file>             Compile and run an Elbereth file
  test <file>            Compile and run tests in an Elbereth file
  gen <file>             Generate Go code from an Elbereth file
  repl                   Start an interactive REPL
  version                Print the version of Elbereth
  help                   Show this help message

Examples:
  elbereth check hello.elb
  elbereth build hello.elb -o hello
  elbereth run hello.elb
  elbereth gen hello.elb
  elbereth repl`)
}

func checkFile(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	lex := lexer.New(string(data))
	p := parser.New(lex)
	prog, err := p.Parse()
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("OK: %d top-level items\n", len(prog.Items))
}

func buildFile(filename string, output string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	lex := lexer.New(string(data))
	p := parser.New(lex)
	prog, err := p.Parse()
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
		os.Exit(1)
	}

	gen := codegen.New()
	ex := expander.New()
	if err := ex.Expand(prog); err != nil {
		fmt.Printf("Expansion error: %v\n", err)
		os.Exit(1)
	}
	goCode, err := gen.Generate(prog)
	if err != nil {
		fmt.Printf("Code generation error: %v\n", err)
		os.Exit(1)
	}

	if output == "" {
		output = "a.out"
	}

	// Write to temporary Go file
	tmpFileName := "/tmp/elbereth_" + randomString() + ".go"
	err = os.WriteFile(tmpFileName, []byte(goCode), 0o644)
	if err != nil {
		fmt.Printf("Error writing Go file: %v\n", err)
		os.Exit(1)
	}
	defer os.Remove(tmpFileName)

	// Compile using go build
	cmd := exec.Command("go", "build", "-o", output, tmpFileName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Compilation error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Built: %s\n", output)
}

func runFile(filename string) {
	output := "/tmp/elbereth_run_" + randomString()
	buildFile(filename, output)

	// Run the compiled binary
	cmd := exec.Command(output)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		os.Exit(1)
	}

	os.Remove(output)
}

func testFile(filename string, testFlags []string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	lex := lexer.New(string(data))
	p := parser.New(lex)
	prog, err := p.Parse()
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
		os.Exit(1)
	}

	gen := codegen.New()
	gen.SetTestMode(true)
	ex := expander.New()
	if err := ex.Expand(prog); err != nil {
		fmt.Printf("Expansion error: %v\n", err)
		os.Exit(1)
	}
	goCode, err := gen.Generate(prog)
	if err != nil {
		fmt.Printf("Code generation error: %v\n", err)
		os.Exit(1)
	}

	// Write to temporary Go test file
	tmpFileName := "/tmp/elbereth_test_" + randomString() + "_test.go"
	err = os.WriteFile(tmpFileName, []byte(goCode), 0o644)
	if err != nil {
		fmt.Printf("Error writing Go file: %v\n", err)
		os.Exit(1)
	}
	defer os.Remove(tmpFileName)

	// Run using go test
	args := append([]string{"test", "-v", tmpFileName}, testFlags...)
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		os.Exit(1)
	}
}

func genCode(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	lex := lexer.New(string(data))
	p := parser.New(lex)
	prog, err := p.Parse()
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
		os.Exit(1)
	}

	gen := codegen.New()
	ex := expander.New()
	if err := ex.Expand(prog); err != nil {
		fmt.Printf("Expansion error: %v\n", err)
		os.Exit(1)
	}
	goCode, err := gen.Generate(prog)
	if err != nil {
		fmt.Printf("Code generation error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(goCode)
}

func randomString() string {
	return fmt.Sprintf("%d", os.Getpid())
}
