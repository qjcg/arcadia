package codegen

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/qjcg/arcadia/exp/elbereth/internal/expander"
	"github.com/qjcg/arcadia/exp/elbereth/internal/lexer"
	"github.com/qjcg/arcadia/exp/elbereth/internal/parser"
)

func checkCompile(t *testing.T, code string) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "elbereth-codegen-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	goFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(goFile, []byte(code), 0o644); err != nil {
		t.Fatal(err)
	}

	// Initialize a dummy go.mod so go build works easily
	cmdMod := exec.Command("go", "mod", "init", "example.com/test")
	cmdMod.Dir = tmpDir
	if err := cmdMod.Run(); err != nil {
		t.Logf("go mod init failed (might be already in a module): %v", err)
	}

	cmd := exec.Command("go", "build", "-o", filepath.Join(tmpDir, "out"), goFile)
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build failed: %v\nOutput: %s\nCode:\n%s", err, string(output), code)
	}
}

func TestGenerate(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "simple function",
			input: `(package main) (defn Add [x y] (+ x y)) (defn main [] (Add 1 2))`,
		},
		{
			name:  "struct type",
			input: `(package main) (deftype Point {X int Y int}) (defn main [] (let [p (Point {:X 1 :Y 2})] (println p)))`,
		},
		{
			name: "macros",
			input: `(package main)
(defmacro unless [c & b] ` + "`" + `(if (not ,c) (do ,@b)))
(defn main [] (unless false (println "ok")))`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			prog, err := p.Parse()
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			ex := expander.New()
			if err := ex.Expand(prog); err != nil {
				t.Fatalf("Expand error: %v", err)
			}

			gen := New()
			code, err := gen.Generate(prog)
			if err != nil {
				t.Fatalf("Generate error: %v", err)
			}

			checkCompile(t, code)
		})
	}
}
