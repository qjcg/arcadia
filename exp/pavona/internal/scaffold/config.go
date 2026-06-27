package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
)

// VariableType indicates how a variable is presented to the user.
type VariableType int

const (
	TypeFreeform VariableType = iota // free-form text input
	TypeChoice                       // select from a list of options
)

func (t VariableType) String() string {
	switch t {
	case TypeFreeform:
		return "freeform"
	case TypeChoice:
		return "choice"
	default:
		return "unknown"
	}
}

// Config represents a parsed config.cue template definition.
type Config struct {
	Name        string
	Description string
	Variables   []Variable
}

// Variable defines a single template variable, inferred from CUE types.
type Variable struct {
	Name     string
	Prompt   string
	Type     VariableType
	Default  string
	Required bool
	Choices  []string
}

// ParseConfig reads and validates config.cue from a directory.
func ParseConfig(dir string) (*Config, error) {
	configPath := filepath.Join(dir, "config.cue")
	if _, err := os.Stat(configPath); err != nil {
		return nil, fmt.Errorf("template must have a config.cue file: %w", err)
	}

	ctx := cuecontext.New()
	bis := load.Instances([]string{"."}, &load.Config{
		Dir: dir,
	})
	if len(bis) == 0 {
		return nil, fmt.Errorf("no CUE instances found in %s", dir)
	}

	inst := ctx.BuildInstance(bis[0])
	if inst.Err() != nil {
		return nil, fmt.Errorf("building CUE instance: %w", inst.Err())
	}

	cfg := &Config{}

	// Extract name
	nameVal := inst.LookupPath(cue.ParsePath("name"))
	if nameVal.Err() != nil {
		return nil, fmt.Errorf("config.cue: missing required field 'name': %w", nameVal.Err())
	}
	s, err := nameVal.String()
	if err != nil {
		return nil, fmt.Errorf("config.cue: field 'name' must be a string: %w", err)
	}
	cfg.Name = s

	// Extract description
	descVal := inst.LookupPath(cue.ParsePath("description"))
	if descVal.Err() != nil {
		return nil, fmt.Errorf("config.cue: missing required field 'description': %w", descVal.Err())
	}
	s, err = descVal.String()
	if err != nil {
		return nil, fmt.Errorf("config.cue: field 'description' must be a string: %w", err)
	}
	cfg.Description = s

	// Extract variables
	variablesVal := inst.LookupPath(cue.ParsePath("variables"))
	if variablesVal.Err() != nil {
		return nil, fmt.Errorf("config.cue: missing required field 'variables': %w", variablesVal.Err())
	}

	fields, err := variablesVal.Fields(cue.Optional(true), cue.Definitions(false), cue.Hidden(false))
	if err != nil {
		return nil, fmt.Errorf("config.cue: reading variables: %w", err)
	}

	// Collect field info with optionality from the iterator
	type fieldEntry struct {
		Name  string
		IsOpt bool
		Value cue.Value
	}
	var entries []fieldEntry
	for fields.Next() {
		sel := fields.Selector()
		entries = append(entries, fieldEntry{
			Name:  sel.Unquoted(),
			IsOpt: fields.IsOptional(),
			Value: fields.Value(),
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})

	for _, fe := range entries {
		variable := Variable{
			Name:     fe.Name,
			Required: !fe.IsOpt,
		}
		v := fe.Value

		// Extract doc comment as the prompt text
		docs := v.Doc()
		if len(docs) > 0 {
			prompt := strings.TrimSpace(docs[0].Text())
			if prompt != "" {
				variable.Prompt = prompt
			}
		}
		if variable.Prompt == "" {
			variable.Prompt = fe.Name
		}

		// Extract default value from CUE's *default syntax
		defVal, hasDefault := v.Default()
		if hasDefault {
			if s, err := defVal.String(); err == nil {
				variable.Default = s
			}
		}

		// Determine type: choice (disjunction of concrete string literals)
		// or freeform (allows any string)
		freeform, choices := extractTypeInfo(v)
		if freeform {
			variable.Type = TypeFreeform
		} else if len(choices) > 0 {
			variable.Type = TypeChoice
			variable.Choices = choices
		} else {
			variable.Type = TypeFreeform
		}

		cfg.Variables = append(cfg.Variables, variable)
	}

	return cfg, nil
}

// extractTypeInfo inspects a CUE value using the value-level expression API
// to determine if it allows any string (freeform) or only a finite set
// of concrete string literals (choice).
func extractTypeInfo(v cue.Value) (freeform bool, choices []string) {
	op, args := v.Expr()
	switch op {
	case cue.NoOp:
		// Single non-composite value
		if v.IncompleteKind() == cue.StringKind && !v.IsConcrete() {
			return true, nil // just "string"
		}
		if v.IsConcrete() {
			if s, err := v.String(); err == nil {
				return false, []string{s}
			}
		}
		// Some other concrete value or non-string type
		return true, nil

	case cue.OrOp:
		// Disjunction — e.g. "a" | "b" | "c" or string | *"default"
		hasString := false
		var literals []string
		for _, arg := range args {
			argFree, argLiterals := extractTypeInfo(arg)
			if argFree {
				hasString = true
			}
			literals = append(literals, argLiterals...)
		}
		if hasString {
			return true, nil
		}
		return false, literals

	default:
		return true, nil
	}
}
