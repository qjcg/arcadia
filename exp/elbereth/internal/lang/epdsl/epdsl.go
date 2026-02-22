package epdsl

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/qjcg/arcadia/exp/elbereth/internal/ast"
	"github.com/qjcg/arcadia/exp/elbereth/internal/lang"
)

// EpDSL is a sample language that parses infix arithmetic expressions.
type EpDSL struct{}

func (l *EpDSL) Parse(input string) (*ast.Program, error) {
	lines := strings.Split(input, "\n")
	prog := &ast.Program{
		Lang:  "epdsl",
		Items: []ast.Node{},
	}

	mainBody := []ast.Expr{}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#lang") {
			continue
		}

		// Simple infix parser for demonstration
		expr, err := parseInfix(trimmed)
		if err != nil {
			return nil, err
		}

		// Wrap in println to see result
		mainBody = append(mainBody, &ast.FuncCall{
			Func: &ast.Symbol{Name: "println"},
			Args: []ast.Expr{&ast.StringLit{Value: trimmed + " = "}, expr},
		})
	}

	prog.Items = append(prog.Items, &ast.Defn{
		Name: "main",
		Body: mainBody,
	})

	return prog, nil
}

func parseInfix(s string) (ast.Expr, error) {
	// Very naive parser for演示 purposes
	// Supports: num + num * num ...
	tokens := strings.Fields(s)
	if len(tokens) == 0 {
		return &ast.NilLit{}, nil
	}

	return parseAdditive(tokens)
}

func parseAdditive(tokens []string) (ast.Expr, error) {
	lhs, rem, err := parseMultiplicative(tokens)
	if err != nil {
		return nil, err
	}

	for len(rem) > 0 && rem[0] == "+" {
		rhs, nextRem, err := parseMultiplicative(rem[1:])
		if err != nil {
			return nil, err
		}
		lhs = &ast.FuncCall{
			Func: &ast.Symbol{Name: "+"},
			Args: []ast.Expr{lhs, rhs},
		}
		rem = nextRem
	}
	return lhs, nil
}

func parseMultiplicative(tokens []string) (lhs ast.Expr, rem []string, err error) {
	lhs, rem, err = parsePrimary(tokens)
	if err != nil {
		return nil, nil, err
	}

	for len(rem) > 0 && rem[0] == "*" {
		rhs, nextRem, err := parsePrimary(rem[1:])
		if err != nil {
			return nil, nil, err
		}
		lhs = &ast.FuncCall{
			Func: &ast.Symbol{Name: "*"},
			Args: []ast.Expr{lhs, rhs},
		}
		rem = nextRem
	}
	return lhs, rem, nil
}

func parsePrimary(tokens []string) (ast.Expr, []string, error) {
	if len(tokens) == 0 {
		return nil, nil, fmt.Errorf("unexpected EOF")
	}
	val, err := strconv.ParseInt(tokens[0], 10, 64)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid number: %s", tokens[0])
	}
	return &ast.IntLit{Value: val}, tokens[1:], nil
}

func (l *EpDSL) Expand(prog *ast.Program) error {
	return nil
}

func init() {
	lang.Register("epdsl", &EpDSL{})
}
