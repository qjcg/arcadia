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

// Config represents a parsed config.cue template definition.
type Config struct {
	Name        string
	Description string
	Variables   []Variable
}

// Variable defines a single template variable.
type Variable struct {
	Name     string
	Prompt   string
	Default  string
	Required bool
	Choices  []string
	Help     string
}

// ParseConfig reads and validates config.cue from a directory.
func ParseConfig(dir string) (*Config, error) {
	// Check config.cue exists
	configPath := filepath.Join(dir, "config.cue")
	if _, err := os.Stat(configPath); err != nil {
		return nil, fmt.Errorf("template must have a config.cue file: %w", err)
	}

	// Load the CUE instance from the template directory
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

	fields, err := variablesVal.Fields(cue.Definitions(false), cue.Hidden(false))
	if err != nil {
		return nil, fmt.Errorf("config.cue: reading variables: %w", err)
	}

	var varNames []string
	for fields.Next() {
		varNames = append(varNames, fields.Label())
	}
	sort.Strings(varNames)

	for _, name := range varNames {
		path := cue.ParsePath("variables." + name)
		v := inst.LookupPath(path)
		if v.Err() != nil {
			continue
		}

		variable := Variable{Name: name}

		if s, err := v.LookupPath(cue.ParsePath("prompt")).String(); err == nil {
			variable.Prompt = s
		}
		if s, err := v.LookupPath(cue.ParsePath("default")).String(); err == nil {
			variable.Default = s
		}
		if b, err := v.LookupPath(cue.ParsePath("required")).Bool(); err == nil {
			variable.Required = b
		}
		if s, err := v.LookupPath(cue.ParsePath("help")).String(); err == nil {
			variable.Help = s
		}

		// Parse choices
		choicesVal := v.LookupPath(cue.ParsePath("choices"))
		if choicesVal.Err() == nil {
			iter, err := choicesVal.List()
			if err == nil {
				for iter.Next() {
					if s, err := iter.Value().String(); err == nil {
						variable.Choices = append(variable.Choices, s)
					}
				}
			}
		}

		cfg.Variables = append(cfg.Variables, variable)
	}

	return cfg, nil
}

// ValidateValues checks that values satisfy the config constraints.
func ValidateValues(cfg *Config, values map[string]string) error {
	var errs []string
	for _, v := range cfg.Variables {
		val := values[v.Name]
		if v.Required && strings.TrimSpace(val) == "" {
			errs = append(errs, fmt.Sprintf("%s: value is required", v.Name))
		}
		if len(v.Choices) > 0 {
			found := false
			for _, c := range v.Choices {
				if val == c {
					found = true
					break
				}
			}
			if !found {
				errs = append(errs, fmt.Sprintf("%s: %q is not a valid choice (valid: %v)", v.Name, val, v.Choices))
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("validation errors:\n  %s", strings.Join(errs, "\n  "))
	}
	return nil
}
