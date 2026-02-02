package minimal

import (
	"strings"

	"github.com/qjcg/arcadia/x/elbereth/internal/ast"
	"github.com/qjcg/arcadia/x/elbereth/internal/lang"
)

// MinimalLanguage is a sample DSL where every line is a string to be printed.
type MinimalLanguage struct{}

func (l *MinimalLanguage) Parse(input string) (*ast.Program, error) {
	lines := strings.Split(input, "\n")
	prog := &ast.Program{
		Lang:  "minimal",
		Items: []ast.Node{},
	}

	mainBody := []ast.Expr{}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#lang") {
			continue
		}

		// Wrap the line in a println call
		mainBody = append(mainBody, &ast.FuncCall{
			Func: &ast.Symbol{Name: "println"},
			Args: []ast.Expr{&ast.StringLit{Value: trimmed}},
		})
	}

	if len(mainBody) == 0 {
		mainBody = append(mainBody, &ast.FuncCall{
			Func: &ast.Symbol{Name: "println"},
			Args: []ast.Expr{&ast.StringLit{Value: "Minimal DSL: No content found"}},
		})
	}

	// Add the main function
	prog.Items = append(prog.Items, &ast.Defn{
		Name: "main",
		Body: mainBody,
	})

	return prog, nil
}

func (l *MinimalLanguage) Expand(prog *ast.Program) error {
	// No expansion needed for this simple language
	return nil
}

func init() {
	lang.Register("minimal", &MinimalLanguage{})
}
