package steps

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

var (
	stepsDir    string
	projectRoot string
	FeaturesDir string
)

func init() {
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		stepsDir = filepath.Dir(filename)
		FeaturesDir = filepath.Dir(stepsDir)
		projectRoot = filepath.Dir(FeaturesDir)
	}
}

type State struct {
	binPath      string
	homeDir      string
	testDir      string
	storedDir    string
	lastOutput   string
	lastExitCode int
}

func (s *State) buildBinary() error {
	if s.binPath != "" {
		return nil
	}
	s.binPath = filepath.Join(s.testDir, "skillo")
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

func (s *State) runApp(args ...string) (string, error) {
	cmd := exec.Command(s.binPath, args...)
	cmd.Env = append(os.Environ(), "HOME="+s.homeDir)
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

func (s *State) resetState() {
	s.lastOutput = ""
	s.lastExitCode = 0
	s.storedDir = ""
}
