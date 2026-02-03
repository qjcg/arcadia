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
	"github.com/qjcg/arcadia/x/elbereth/internal/lang"
	_ "github.com/qjcg/arcadia/x/elbereth/internal/lang/epdsl"
	_ "github.com/qjcg/arcadia/x/elbereth/internal/lang/minimal"
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
			buildPackage(path, *output)
		} else {
			// Even if it's a file, we treat it as part of its package
			buildPackage(filepath.Dir(path), *output)
		}

	case "run":
		if len(os.Args) < 3 {
			fmt.Println("Usage: elbereth run <file_or_dir> [args...]")
			os.Exit(1)
		}
		runPackage(os.Args[2], os.Args[3:]...)

	case "init":
		if len(os.Args) < 3 {
			fmt.Println("Usage: elbereth init <module-name>")
			os.Exit(1)
		}
		runGoCommand("mod", "init", os.Args[2])

	case "mod":
		if len(os.Args) < 3 {
			fmt.Println("Usage: elbereth mod <command> [args...]")
			os.Exit(1)
		}
		runGoCommand("mod", os.Args[2:]...)

	case "get", "tidy":
		runGoCommand(cmd, os.Args[2:]...)

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
  init <module>          Initialize a new Elbereth module (runs go mod init)
  mod <cmd>              Run a Go module command (e.g., tidy, vendor)
  get <pkg>              Add a dependency to the module (runs go get)
  tidy                   Tidy module dependencies (runs go mod tidy)
  check <path>           Check syntax of an Elbereth file or directory
  build <path> [-o out]  Compile Elbereth package to a binary or Go package
  run <path>             Compile and run an Elbereth package
  test <path>            Compile and run tests in an Elbereth package
  gen <path>             Generate Go code from an Elbereth file
  repl                   Start an interactive REPL
  version                Print the version of Elbereth
  help                   Show this help message

Examples:
  elbereth init github.com/user/myproject
  elbereth tidy
  elbereth build .
  elbereth run main.elb
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

func buildPackage(dir string, output string) {
	// First, transcode all .elb files in the directory to .go
	files, err := os.ReadDir(dir)
	if err != nil {
		fmt.Printf("Error reading directory: %v\n", err)
		os.Exit(1)
	}

	var hasElb bool
	var isMain bool
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".elb") {
			hasElb = true
			pkg, err := transcodeFile(filepath.Join(dir, f.Name()))
			if err != nil {
				fmt.Printf("Error transcoding %s: %v\n", f.Name(), err)
				os.Exit(1)
			}
			if pkg == "main" {
				isMain = true
			}
		}
	}

	if !hasElb {
		// Maybe it's a Go-only package or sub-package path, pass to Go directly
	}

	// Now run go build on the package
	var args []string
	args = append(args, "build")
	if output != "" {
		args = append(args, "-o", output)
	}
	args = append(args, ".")

	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		if isMain || hasElb {
			fmt.Printf("Go build error: %v\n", err)
			os.Exit(1)
		}
	}

	if isMain {
		if output == "" {
			output = filepath.Base(dir)
			if output == "." || output == "/" {
				output = "a.out"
			}
		}
		fmt.Printf("Built executable: %s/%s\n", dir, output)
	} else {
		fmt.Printf("Compiled package in %s\n", dir)
	}
}

func runPackage(path string, args ...string) {
	dir := path
	info, err := os.Stat(path)
	if err == nil && !info.IsDir() {
		dir = filepath.Dir(path)
	}

	// Transcode all .elb files in the package directory
	files, _ := os.ReadDir(dir)
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".elb") {
			_, err := transcodeFile(filepath.Join(dir, f.Name()))
			if err != nil {
				fmt.Printf("Error transcoding %s: %v\n", f.Name(), err)
				os.Exit(1)
			}
		}
	}

	// Run using go run . [args...]
	goArgs := append([]string{"run", "."}, args...)
	cmd := exec.Command("go", goArgs...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		os.Exit(1)
	}
}

func parseWithLang(data []byte) (*ast.Program, error) {
	lex := lexer.New(string(data))
	p := parser.New(lex)
	prog, err := p.Parse()
	if err != nil {
		return nil, err
	}

	if prog.Lang != "" {
		l, err := lang.Get(prog.Lang)
		if err != nil {
			return nil, err
		}
		return l.Parse(string(data))
	}

	return prog, nil
}

func transcodeFile(filename string) (string, error) {
	absPath, _ := filepath.Abs(filename)
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}

	prog, err := parseWithLang(data)
	if err != nil {
		return "", err
	}

	if prog.Package == "" {
		// Default to directory name, but force 'main' if it's a standalone run
		// or if it's a special DSL.
		prog.Package = filepath.Base(filepath.Dir(absPath))
		if prog.Package == "." || prog.Package == "/" || prog.Lang != "" {
			prog.Package = "main"
		}
	}

	gen := codegen.New()
	if prog.Lang != "" {
		l, _ := lang.Get(prog.Lang)
		if err := l.Expand(prog); err != nil {
			return "", err
		}
	} else {
		ex := expander.New()
		ex.Expand(prog)
	}
	goCode, err := gen.Generate(prog)
	if err != nil {
		return "", err
	}

	goFileName := strings.TrimSuffix(absPath, ".elb") + ".go"
	err = os.WriteFile(goFileName, []byte(goCode), 0o644)
	return prog.Package, err
}

func runGoCommand(command string, args ...string) {
	fullArgs := append([]string{command}, args...)
	cmd := exec.Command("go", fullArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		os.Exit(1)
	}
}

func testFiles(filenames []string, testFlags []string) {
	var allItems []ast.Node
	var firstLang string

	for i, filename := range filenames {
		data, err := os.ReadFile(filename)
		if err != nil {
			fmt.Printf("Error reading file %s: %v\n", filename, err)
			os.Exit(1)
		}

		prog, err := parseWithLang(data)
		if err != nil {
			fmt.Printf("Parse error in %s: %v\n", filename, err)
			os.Exit(1)
		}
		if i == 0 {
			firstLang = prog.Lang
		}
		allItems = append(allItems, prog.Items...)
	}

	prog := &ast.Program{Items: allItems, Lang: firstLang}

	gen := codegen.New()
	gen.SetTestMode(true)
	if prog.Lang != "" {
		l, _ := lang.Get(prog.Lang)
		if err := l.Expand(prog); err != nil {
			fmt.Printf("Expansion error: %v\n", err)
			os.Exit(1)
		}
	} else {
		ex := expander.New()
		if err := ex.Expand(prog); err != nil {
			fmt.Printf("Expansion error: %v\n", err)
			os.Exit(1)
		}
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
	// If the file is in a package, we might need more care.
	// For now, this handles simple cases.
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

	prog, err := parseWithLang(data)
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
		os.Exit(1)
	}

	gen := codegen.New()
	if prog.Lang != "" {
		l, _ := lang.Get(prog.Lang)
		if err := l.Expand(prog); err != nil {
			fmt.Printf("Expansion error: %v\n", err)
			os.Exit(1)
		}
	} else {
		ex := expander.New()
		if err := ex.Expand(prog); err != nil {
			fmt.Printf("Expansion error: %v\n", err)
			os.Exit(1)
		}
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
