package types

import (
	"strings"

	"gopkg.in/yaml.v3"
)

// StringList is a []string that accepts both a single YAML scalar (comma-separated)
// and a YAML sequence when unmarshaling.
type StringList []string

func (s *StringList) UnmarshalYAML(value *yaml.Node) error {
	var single string
	if err := value.Decode(&single); err == nil {
		parts := strings.Split(single, ",")
		result := make(StringList, len(parts))
		for i, p := range parts {
			result[i] = strings.TrimSpace(p)
		}
		*s = result
		return nil
	}
	var multi []string
	if err := value.Decode(&multi); err != nil {
		return err
	}
	*s = multi
	return nil
}

type Skill struct {
	Name          string            `yaml:"name"`
	Description   string            `yaml:"description"`
	License       string            `yaml:"license,omitempty"`
	Compatibility string            `yaml:"compatibility,omitempty"`
	Metadata      map[string]string `yaml:"metadata,omitempty"`
	AllowedTools  StringList        `yaml:"allowed-tools,omitempty"`
}
