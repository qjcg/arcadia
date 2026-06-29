package steps

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// stepsDir is the directory containing this source file, computed at init.
var stepsDir string

// FeaturesDir is the parent of stepsDir, i.e. the features/ directory.
var FeaturesDir string

// projectRoot is the root of the pavona project.
var projectRoot string

func init() {
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		stepsDir = filepath.Dir(filename)
		FeaturesDir = filepath.Dir(stepsDir)
		projectRoot = filepath.Dir(FeaturesDir)
	}
}

// PavonaState holds the state shared across scenarios.
type PavonaState struct {
	binPath      string // path to the compiled pavona binary
	tmpDir       string // temp dir for test outputs
	lastOutput   string // last stdout/stderr from running pavona
	lastExitCode int    // last exit code
	outputDir    string // output directory for hydrate scenarios
	customDir    string // path to custom template testdata
	existingDir  string // pre-created directory for error scenarios
}

// buildBinary compiles the pavona binary once.
func (s *PavonaState) buildBinary() error {
	if s.binPath != "" {
		return nil
	}
	s.binPath = filepath.Join(s.tmpDir, "pavona")
	cmd := exec.Command("go", "build", "-o", s.binPath, ".")
	cmd.Dir = projectRoot
	cmd.Env = append(os.Environ(), "GOWORK=off")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	_ = out
	return nil
}

// runPavona executes the pavona binary with the given arguments.
func (s *PavonaState) runPavona(args ...string) (string, error) {
	cmd := exec.Command(s.binPath, args...)
	out, err := cmd.CombinedOutput()
	output := string(out)
	s.lastOutput = output
	s.lastExitCode = 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			s.lastExitCode = exitErr.ExitCode()
		} else {
			return output, err
		}
	}
	return output, nil
}

// fileExists checks if a file exists relative to the output directory.
func (s *PavonaState) fileExists(path string) bool {
	fullPath := filepath.Join(s.outputDir, path)
	_, err := os.Stat(fullPath)
	return err == nil
}

// readFile reads a file relative to the output directory.
func (s *PavonaState) readFile(path string) (string, error) {
	fullPath := filepath.Join(s.outputDir, path)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// reset clears state for a new scenario.
// containsStr returns an error if s does not contain substr.
func containsStr(s, substr string) error {
	if strings.Contains(s, substr) {
		return nil
	}
	return fmt.Errorf("expected output to contain %q, got:\n%s", substr, s)
}

// errf is a fmt.Errorf shorthand.
func errf(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}

func (s *PavonaState) reset() {
	s.lastOutput = ""
	s.lastExitCode = 0
	s.outputDir = ""
	s.customDir = ""
	s.existingDir = ""
}

// createCustomTemplate sets the customDir to the testdata/custom directory.
func (s *PavonaState) createCustomTemplate() error {
	s.customDir = filepath.Join(stepsDir, "testdata", "custom")
	if _, err := os.Stat(s.customDir); err != nil {
		return err
	}
	return nil
}

// createExistingDir creates a non-empty directory for error scenarios.
func (s *PavonaState) createExistingDir() error {
	dir := filepath.Join(s.tmpDir, "existing-output")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "dummy.txt"), []byte("content"), 0o644); err != nil {
		return err
	}
	s.existingDir = dir
	return nil
}
