package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const compileTemplate = `package main

import (
	"os"
	"github.com/qjcg/arcadia/exp/terebra/internal/shell"
)

func main() {
	script := %s
	if err := shell.RunScriptFromString(script); err != nil {
		os.Exit(1)
	}
}
`

// buildCmd compiles a .trb script into a standalone Go binary.
func buildCmd(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: terebra build <script.trb> [output]")
		return 1
	}

	scriptPath := args[0]
	outputPath := ""
	if len(args) > 1 {
		outputPath = args[1]
	}

	// Read the script
	data, err := os.ReadFile(scriptPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "build: %v\n", err)
		return 1
	}

	// Determine output name
	if outputPath == "" {
		base := filepath.Base(scriptPath)
		ext := filepath.Ext(base)
		outputPath = strings.TrimSuffix(base, ext)
	}

	// Create temp directory for the build
	tmpDir, err := os.MkdirTemp("", "terebra-build-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "build: %v\n", err)
		return 1
	}
	defer os.RemoveAll(tmpDir)

	// Escape the script for embedding in Go source
	escaped := fmt.Sprintf("%q", string(data))

	// Generate the Go source
	goSource := fmt.Sprintf(compileTemplate, escaped)
	goPath := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(goPath, []byte(goSource), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "build: %v\n", err)
		return 1
	}

	// Initialize a Go module
	goMod := `module terebra_build

go 1.26.4

require github.com/qjcg/arcadia/exp/terebra v0.0.0

replace github.com/qjcg/arcadia/exp/terebra => ` + "`" + getModuleRoot() + "`" + `
`
	modPath := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(modPath, []byte(goMod), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "build: %v\n", err)
		return 1
	}

	// Run go mod tidy
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = tmpDir
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "build: go mod tidy failed: %v\n", err)
		return 1
	}

	// Build the binary
	buildCmd := exec.Command("go", "build", "-o", outputPath, ".")
	buildCmd.Dir = tmpDir
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "build: compilation failed: %v\n", err)
		return 1
	}

	fmt.Fprintf(os.Stderr, "build: compiled %s -> %s\n", scriptPath, outputPath)
	return 0
}

// getModuleRoot returns the absolute path to the terebra module root.
func getModuleRoot() string {
	// Try to find the module root from the executable path
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	// Walk up from the executable to find go.mod
	dir := filepath.Dir(exe)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			// Check if it's the terebra module
			data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
			if err == nil && strings.Contains(string(data), "github.com/qjcg/arcadia/exp/terebra") {
				return dir
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "."
}
