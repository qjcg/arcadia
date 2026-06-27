package steps

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
)

type ScaffoldState struct {
	tmpDir      string
	lastError   error
	projectType string
	projectName string
}

func NewScaffoldState() *ScaffoldState {
	tmpDir, err := os.MkdirTemp("", "pavona-test-")
	if err != nil {
		panic(err)
	}
	return &ScaffoldState{
		tmpDir: tmpDir,
	}
}

func (s *ScaffoldState) Cleanup() {
	os.RemoveAll(s.tmpDir)
}

func (s *ScaffoldState) runPavona(args ...string) error {
	// Allow override via env var (set by test runner)
	if bin := os.Getenv("PAVONA_BIN"); bin != "" {
		cmd := exec.Command(bin, args...)
		cmd.Dir = s.tmpDir
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("%s: %s", err, string(out))
		}
		return nil
	}

	binary, err := os.Executable()
	if err != nil {
		return err
	}

	// Look for pavona next to the test binary, then in PATH
	pavonaBin := filepath.Join(filepath.Dir(binary), "pavona")
	if _, err := os.Stat(pavonaBin); os.IsNotExist(err) {
		var err error
		pavonaBin, err = exec.LookPath("pavona")
		if err != nil {
			return fmt.Errorf("pavona binary not found (set PAVONA_BIN): %w", err)
		}
	}

	cmd := exec.Command(pavonaBin, args...)
	cmd.Dir = s.tmpDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err, string(out))
	}
	return nil
}

func (s *ScaffoldState) iScaffoldATypeNamed(projectType, name string) error {
	s.projectType = projectType
	s.projectName = name
	s.lastError = s.runPavona("new", projectType, name)
	return nil
}

func (s *ScaffoldState) theProjectShouldExist(name string) error {
	dir := filepath.Join(s.tmpDir, name)
	info, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return godog.ErrPending
	}
	return nil
}

func (s *ScaffoldState) fileShouldContain(filePath, expectedSubstring string) error {
	content, err := os.ReadFile(filepath.Join(s.tmpDir, filePath))
	if err != nil {
		return err
	}
	if !strings.Contains(string(content), expectedSubstring) {
		return godog.ErrPending
	}
	return nil
}

func (s *ScaffoldState) fileShouldExist(filePath string) error {
	_, err := os.Stat(filepath.Join(s.tmpDir, filePath))
	return err
}

func (s *ScaffoldState) theProjectShouldCompile() error {
	projectDir := filepath.Join(s.tmpDir, s.projectName)

	tidy := exec.Command("go", "mod", "tidy")
	tidy.Dir = projectDir
	if out, err := tidy.CombinedOutput(); err != nil {
		return fmt.Errorf("go mod tidy failed: %w\n%s", err, out)
	}

	build := exec.Command("go", "build", "-o", "/dev/null", ".")
	build.Dir = projectDir
	if out, err := build.CombinedOutput(); err != nil {
		return fmt.Errorf("go build failed: %w\n%s", err, out)
	}

	return nil
}

func (s *ScaffoldState) iShouldGetAnError(expected string) error {
	if s.lastError == nil {
		return fmt.Errorf("expected error containing %q, but got none", expected)
	}
	if !strings.Contains(s.lastError.Error(), expected) {
		return fmt.Errorf("expected error containing %q, got %q", expected, s.lastError.Error())
	}
	return nil
}

func RegisterScaffoldSteps(ctx *godog.ScenarioContext) {
	s := NewScaffoldState()

	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		s = NewScaffoldState()
		return ctx, nil
	})

	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		s.Cleanup()
		return ctx, nil
	})

	ctx.Step(`^I scaffold a "([^"]*)" named "([^"]*)"$`,
		func(projectType, name string) error {
			s.projectType = projectType
			s.projectName = name
			s.lastError = s.runPavona("new", projectType, name)
			return nil
		})
	ctx.Step(`^the project "([^"]*)" should exist$`,
		func(name string) error {
			dir := filepath.Join(s.tmpDir, name)
			info, err := os.Stat(dir)
			if err != nil {
				return err
			}
			if !info.IsDir() {
				return fmt.Errorf("%s is not a directory", dir)
			}
			return nil
		})
	ctx.Step(`^"([^"]*)" should contain "([^"]*)"$`,
		func(filePath, expectedSubstring string) error {
			content, err := os.ReadFile(filepath.Join(s.tmpDir, filePath))
			if err != nil {
				return err
			}
			if !strings.Contains(string(content), expectedSubstring) {
				return fmt.Errorf("%s does not contain %q", filePath, expectedSubstring)
			}
			return nil
		})
	ctx.Step(`^"([^"]*)" should exist$`,
		func(filePath string) error {
			_, err := os.Stat(filepath.Join(s.tmpDir, filePath))
			return err
		})
	ctx.Step(`^the project should compile$`,
		func() error {
			projectDir := filepath.Join(s.tmpDir, s.projectName)
			tidy := exec.Command("go", "mod", "tidy")
			tidy.Dir = projectDir
			if out, err := tidy.CombinedOutput(); err != nil {
				return fmt.Errorf("go mod tidy failed: %w\n%s", err, out)
			}
			build := exec.Command("go", "build", "-o", "/dev/null", ".")
			build.Dir = projectDir
			if out, err := build.CombinedOutput(); err != nil {
				return fmt.Errorf("go build failed: %w\n%s", err, out)
			}
			return nil
		})
	ctx.Step(`^I should get an error about unknown project type$`,
		func() error {
			if s.lastError == nil {
				return fmt.Errorf("expected error containing %q, but got none", "unknown")
			}
			if !strings.Contains(s.lastError.Error(), "unknown") {
				return fmt.Errorf("expected error containing %q, got %q", "unknown", s.lastError.Error())
			}
			return nil
		})
	ctx.Step(`^I scaffold a "([^"]*)" named "([^"]*)" with pages "([^"]*)"$`,
		func(projectType, name, pages string) error {
			s.projectType = projectType
			s.projectName = name
			// Split pages on commas (the step definition passes them as a single string)
			pageArgs := strings.Split(pages, ",")
			args := []string{"new", projectType, name}
			for _, p := range pageArgs {
				args = append(args, "--pages", strings.TrimSpace(p))
			}
			s.lastError = s.runPavona(args...)
			return nil
		})
	ctx.Step(`^I scaffold a "([^"]*)" named "([^"]*)" with format "([^"]*)" and pages "([^"]*)"$`,
		func(projectType, name, format, pages string) error {
			s.projectType = projectType
			s.projectName = name
			pageArgs := strings.Split(pages, ",")
			args := []string{"new", projectType, name, "--format", format}
			for _, p := range pageArgs {
				args = append(args, "--pages", strings.TrimSpace(p))
			}
			s.lastError = s.runPavona(args...)
			return nil
		})
	ctx.Step(`^I scaffold a "([^"]*)" named "([^"]*)" with format "([^"]*)"$`,
		func(projectType, name, format string) error {
			s.projectType = projectType
			s.projectName = name
			s.lastError = s.runPavona("new", projectType, name, "--format", format)
			return nil
		})
	ctx.Step(`^a directory called "([^"]*)"$`,
		func(name string) error {
			return os.MkdirAll(filepath.Join(s.tmpDir, name), 0o755)
		})
	ctx.Step(`^I should get an error about the directory already existing$`,
		func() error {
			if s.lastError == nil {
				return fmt.Errorf("expected error about directory existing, but got none")
			}
			if !strings.Contains(s.lastError.Error(), "already exists") {
				return fmt.Errorf("expected error containing %q, got %q", "already exists", s.lastError.Error())
			}
			return nil
		})
}
