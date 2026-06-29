package steps

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
)

func RegisterAllSteps(ctx *godog.ScenarioContext, s *State) {
	ctx.Step(`^a clean home directory$`, func() error {
		return s.setupCleanHome()
	})
	ctx.Step(`^a valid skill directory$`, func() error {
		return s.setupValidSkillDir()
	})
	ctx.Step(`^an empty directory$`, func() error {
		return s.setupEmptyDir()
	})
	ctx.Step(`^I run "skillo" with "([^"]+)"$`, func(args string) error {
		parts := splitArgs(args)
		_, err := s.runApp(parts...)
		return err
	})
	ctx.Step(`^I run "skillo" with "init --project"$`, func() error {
		_, err := s.runApp("init", "--project")
		return err
	})
	ctx.Step(`^it should succeed$`, func() error {
		if s.lastExitCode != 0 {
			return fmt.Errorf("expected exit code 0, got %d\noutput:\n%s", s.lastExitCode, s.lastOutput)
		}
		return nil
	})
	ctx.Step(`^it should fail$`, func() error {
		if s.lastExitCode == 0 {
			return fmt.Errorf("expected non-zero exit code, got 0\noutput:\n%s", s.lastOutput)
		}
		return nil
	})
	ctx.Step(`^the output should contain "([^"]+)"$`, func(expected string) error {
		if !strings.Contains(s.lastOutput, expected) {
			return fmt.Errorf("expected output to contain %q, got:\n%s", expected, s.lastOutput)
		}
		return nil
	})
	ctx.Step(`^the file "([^"]+)" should exist$`, func(path string) error {
		fullPath := filepath.Join(s.homeDir, path)
		_, err := os.Stat(fullPath)
		if err != nil {
			return fmt.Errorf("expected file %s to exist: %w", fullPath, err)
		}
		return nil
	})
	ctx.Step(`^the file "([^"]+)" should contain "([^"]+)"$`, func(path, expected string) error {
		fullPath := filepath.Join(s.homeDir, path)
		data, err := os.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("reading file %s: %w", fullPath, err)
		}
		if !strings.Contains(string(data), expected) {
			return fmt.Errorf("expected file %s to contain %q, got:\n%s", path, expected, string(data))
		}
		return nil
	})
}

func (s *State) setupCleanHome() error {
	homeDir := filepath.Join(s.testDir, "home")
	if err := os.MkdirAll(homeDir, 0o755); err != nil {
		return err
	}
	s.homeDir = homeDir
	return nil
}

func (s *State) setupValidSkillDir() error {
	skillDir := filepath.Join(s.testDir, "valid-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		return err
	}
	content := `---
name: test-skill
description: A test skill for validation
---

# Test Skill
`
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
		return err
	}
	s.storedDir = skillDir
	return nil
}

func (s *State) setupEmptyDir() error {
	emptyDir := filepath.Join(s.testDir, "empty")
	if err := os.MkdirAll(emptyDir, 0o755); err != nil {
		return err
	}
	s.storedDir = emptyDir
	return nil
}

func splitArgs(args string) []string {
	var parts []string
	for arg := range strings.FieldsSeq(args) {
		if arg != "" {
			parts = append(parts, arg)
		}
	}
	return parts
}
