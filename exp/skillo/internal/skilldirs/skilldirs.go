package skilldirs

import (
	"os"
	"path/filepath"
)

// Sources holds the primary and secondary skill directories.
// Primary is the project-level dir if in a git repo, otherwise user-level.
// Secondary is always ~/.agents/skills/ and may equal Primary when outside a repo.
type Sources struct {
	Primary   string
	Secondary string
	HomeDir   string
}

// Detect returns the skill sources based on the current working directory.
func Detect(homeDir, cwd string) *Sources {
	userSkills := filepath.Join(homeDir, ".agents", "skills")

	projectRoot := findGitRoot(cwd)
	if projectRoot != "" {
		projectSkills := filepath.Join(projectRoot, ".agents", "skills")
		return &Sources{
			Primary:   projectSkills,
			Secondary: userSkills,
			HomeDir:   homeDir,
		}
	}

	return &Sources{
		Primary:   userSkills,
		Secondary: userSkills,
		HomeDir:   homeDir,
	}
}

// DefaultDir returns the primary skills directory for the given cwd.
func DefaultDir(homeDir, cwd string) string {
	return Detect(homeDir, cwd).Primary
}

// UserDir returns the user-level skills directory.
func UserDir(homeDir string) string {
	return filepath.Join(homeDir, ".agents", "skills")
}

func findGitRoot(dir string) string {
	d, err := filepath.Abs(dir)
	if err != nil {
		return ""
	}
	for {
		info, err := os.Stat(filepath.Join(d, ".git"))
		if err == nil && info.IsDir() {
			return d
		}
		parent := filepath.Dir(d)
		if parent == d {
			return ""
		}
		d = parent
	}
}
