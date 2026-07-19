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
	outputPath, err = filepath.Abs(outputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "build: %v\n", err)
		return 1
	}

	// Escape the script for embedding in Go source
	escaped := fmt.Sprintf("%q", string(data))

	// Generate the Go source
	goSource := fmt.Sprintf(compileTemplate, escaped)

	// Create a temp file in the module root so internal packages are accessible
	tmpFile, err := os.CreateTemp(".", "terebra-build-*.go")
	if err != nil {
		fmt.Fprintf(os.Stderr, "build: %v\n", err)
		return 1
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	if err := os.WriteFile(tmpPath, []byte(goSource), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "build: %v\n", err)
		return 1
	}

	// Build the binary within the module root
	buildCmd := exec.Command("go", "build", "-o", outputPath, tmpPath)
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "build: compilation failed: %v\n", err)
		return 1
	}

	fmt.Fprintf(os.Stderr, "build: compiled %s -> %s\n", scriptPath, outputPath)
	return 0
}
