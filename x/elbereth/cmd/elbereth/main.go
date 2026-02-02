package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/qjcg/arcadia/x/elbereth/internal/ast"
	"github.com/qjcg/arcadia/x/elbereth/internal/codegen"
	"github.com/qjcg/arcadia/x/elbereth/internal/expander"
	"github.com/qjcg/arcadia/x/elbereth/internal/lexer"
	"github.com/qjcg/arcadia/x/elbereth/internal/parser"
	"github.com/qjcg/arcadia/x/elbereth/internal/repl"
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
			fmt.Println("Usage: elbereth check <file_or_dir>")
			os.Exit(1)
		}
		path := os.Args[2]
		info, err := os.Stat(path)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		if info.IsDir() {
			checkDir(path)
		} else {
			checkFile(path)
		}

	case "build":
		if len(os.Args) < 3 {
			fmt.Println("Usage: elbereth build <file_or_dir> [-o output]")
			os.Exit(1)
		}
		flagSet := flag.NewFlagSet("build", flag.ContinueOnError)
		output := flagSet.String("o", "", "output file")
		flagSet.Parse(os.Args[3:])

		path := os.Args[2]
		info, err := os.Stat(path)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		if info.IsDir() {
			buildDir(path)
		} else {
			buildFile(path, *output)
		}

	case "run":
		if len(os.Args) < 3 {
			fmt.Println("Usage: elbereth run <file>")
			os.Exit(1)
		}
		runFile(os.Args[2])

	case "test":
		if len(os.Args) < 3 {
			fmt.Println("Usage: elbereth test <file_or_dir> [go test flags]")
			os.Exit(1)
		}
		var files []string
		var flags []string
		for _, arg := range os.Args[2:] {
			if strings.HasSuffix(arg, ".elb") {
				files = append(files, arg)
			} else if info, err := os.Stat(arg); err == nil && info.IsDir() {
				// Find all .elb files in dir
				filepath.Walk(arg, func(path string, info os.FileInfo, err error) error {
					if err == nil && !info.IsDir() && strings.HasSuffix(path, ".elb") {
						files = append(files, path)
					}
					return nil
				})
			} else {
				flags = append(flags, arg)
			}
		}
		testFiles(files, flags)

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

func checkDir(dir string) {
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".elb") {
			checkFile(path)
		}
		return nil
	})
}

func buildDir(dir string) {
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".elb") {
			buildFile(path, "")
		}
		return nil
	})
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
	absPath, err := filepath.Abs(filename)
	if err != nil {
		fmt.Printf("Error getting absolute path: %v\n", err)
		os.Exit(1)
	}

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

	// Infer package name from directory if not declared
	if prog.Package == "" {
		prog.Package = filepath.Base(filepath.Dir(absPath))
		if prog.Package == "." || prog.Package == "/" {
			prog.Package = "main"
		}
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

	// If no output specified, use the filename base
	if output == "" {
		output = strings.TrimSuffix(filepath.Base(filename), ".elb")
	}

	// Write to a .go file in the same directory
	goFileName := strings.TrimSuffix(absPath, ".elb") + ".go"
	err = os.WriteFile(goFileName, []byte(goCode), 0o644)
	if err != nil {
		fmt.Printf("Error writing Go file: %v\n", err)
		os.Exit(1)
	}
	// No defer removal here if we want to keep the Go files for the module system

	// Compile using go build
	// If it's package main, we build an executable.
	// Otherwise, we just make sure it compiles.
	var cmd *exec.Cmd
	if prog.Package == "main" {
		cmd = exec.Command("go", "build", "-o", output, goFileName)
	} else {
		fmt.Printf("Compiled package %s to %s\n", prog.Package, goFileName)
		return
	}
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

func testFiles(filenames []string, testFlags []string) {
	var allItems []ast.Node

	for _, filename := range filenames {
		data, err := os.ReadFile(filename)
		if err != nil {
			fmt.Printf("Error reading file %s: %v\n", filename, err)
			os.Exit(1)
		}

		lex := lexer.New(string(data))
		p := parser.New(lex)
		prog, err := p.Parse()
		if err != nil {
			fmt.Printf("Parse error in %s: %v\n", filename, err)
			os.Exit(1)
		}
		allItems = append(allItems, prog.Items...)
	}

	prog := &ast.Program{Items: allItems}

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
	args := []string{"test", "-v", tmpFileName}
	args = append(args, testFlags...)
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
