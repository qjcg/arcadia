package steps

import (
	"path/filepath"

	"github.com/cucumber/godog"
)

func RegisterListSteps(ctx *godog.ScenarioContext, state *PavonaState) {
	ctx.Step(`^I run pavona with "([^"]+)"$`, func(flags string) error {
		return state.runPavonaWithFlags(flags)
	})
	ctx.Step(`^the output should contain "([^"]+)"$`, func(expected string) error {
		return state.outputShouldContain(expected)
	})
}

func RegisterHydrateSteps(ctx *godog.ScenarioContext, state *PavonaState) {
	ctx.Step(`^I hydrate the "([^"]+)" template with name "([^"]+)"$`, func(template, name string) error {
		return state.hydrateTemplate(template, name, "")
	})
	ctx.Step(`^I hydrate the "([^"]+)" template into that directory$`, func(template string) error {
		return state.hydrateTemplate(template, "test-project", state.existingDir)
	})
	ctx.Step(`^the output directory should contain "([^"]+)"$`, func(path string) error {
		return state.outputDirShouldContain(path)
	})
	ctx.Step(`^"([^"]+)" should contain "([^"]+)"$`, func(filePath, expected string) error {
		return state.fileShouldContain(filePath, expected)
	})
}

func RegisterCustomSteps(ctx *godog.ScenarioContext, state *PavonaState) {
	ctx.Step(`^a custom template with config\.cue and main\.go\.tmpl$`, func() error {
		return state.createCustomTemplate()
	})
	ctx.Step(`^I hydrate the custom template with name "([^"]+)"$`, func(name string) error {
		return state.hydrateCustomTemplate(name)
	})
}

func RegisterErrorSteps(ctx *godog.ScenarioContext, state *PavonaState) {
	ctx.Step(`^I run pavona with "([^"]+)", "([^"]+)"$`, func(flag, value string) error {
		_, err := state.runPavona(flag, value)
		_ = err // we care about exit code, captured in state
		return nil
	})
	ctx.Step(`^an existing non-empty output directory$`, func() error {
		return state.createExistingDir()
	})
	ctx.Step(`^the output should contain "([^"]+)"$`, func(expected string) error {
		return state.outputShouldContain(expected)
	})
}

func RegisterVersionSteps(ctx *godog.ScenarioContext, state *PavonaState) {
	ctx.Step(`^I run pavona with "([^"]+)"$`, func(flags string) error {
		return state.runPavonaWithFlags(flags)
	})
	ctx.Step(`^the output should contain "([^"]+)"$`, func(expected string) error {
		return state.outputShouldContain(expected)
	})
}

// runPavonaWithFlags splits a space-separated flag string and runs pavona.
func (s *PavonaState) runPavonaWithFlags(flags string) error {
	// Parse flags respecting quoted strings
	args := parseArgs(flags)
	_, err := s.runPavona(args...)
	return err
}

// parseArgs splits a string into arguments, respecting double-quoted tokens.
func parseArgs(s string) []string {
	var args []string
	var current []byte
	inQuote := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '"' {
			inQuote = !inQuote
			continue
		}
		if c == ' ' && !inQuote {
			if len(current) > 0 {
				args = append(args, string(current))
				current = nil
			}
			continue
		}
		current = append(current, c)
	}
	if len(current) > 0 {
		args = append(args, string(current))
	}
	return args
}

// outputShouldContain checks that lastOutput contains a substring.
func (s *PavonaState) outputShouldContain(expected string) error {
	return containsStr(s.lastOutput, expected)
}

// outputDirShouldContain checks that outputDir has a given file.
func (s *PavonaState) outputDirShouldContain(path string) error {
	if !s.fileExists(path) {
		return errf("expected %q to exist in output directory, but it does not", path)
	}
	return nil
}

// fileShouldContain checks that a file in outputDir contains a substring.
func (s *PavonaState) fileShouldContain(filePath, expected string) error {
	content, err := s.readFile(filePath)
	if err != nil {
		return errf("could not read %q: %w", filePath, err)
	}
	return containsStr(content, expected)
}

// hydrateTemplate runs pavona to hydrate a built-in template with quiet mode.
func (s *PavonaState) hydrateTemplate(template, name, outputDir string) error {
	if outputDir == "" {
		outputDir = filepath.Join(s.tmpDir, name)
	}
	s.outputDir = outputDir
	_, err := s.runPavona("-t", template, "-o", outputDir, "-n", name, "-q")
	return err
}

// hydrateCustomTemplate hydrates the custom template directory.
func (s *PavonaState) hydrateCustomTemplate(name string) error {
	s.outputDir = filepath.Join(s.tmpDir, name)
	_, err := s.runPavona("-t", s.customDir, "-o", s.outputDir, "-n", name, "-q")
	return err
}
