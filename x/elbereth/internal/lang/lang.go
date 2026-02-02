package lang

import (
	"fmt"
	"sync"

	"github.com/qjcg/arcadia/x/elbereth/internal/ast"
)

// Language defines the interface for a domain-specific or general-purpose language.
// Just like Racket's #lang, a Language provides its own parser and expander.
type Language interface {
	// Parse takes the raw source code (after the #lang directive) and returns an AST.
	Parse(input string) (*ast.Program, error)
	// Expand performs language-specific transformations on the AST.
	Expand(prog *ast.Program) error
}

var (
	registry = make(map[string]Language)
	mu       sync.RWMutex
)

// Register adds a new language to the global registry.
func Register(name string, lang Language) {
	mu.Lock()
	defer mu.Unlock()
	registry[name] = lang
}

// Get retrieves a language by name from the registry.
func Get(name string) (Language, error) {
	mu.RLock()
	defer mu.RUnlock()
	lang, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown language: %s", name)
	}
	return lang, nil
}
