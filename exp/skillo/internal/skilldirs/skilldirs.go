package skilldirs

import (
	"os"
	"path/filepath"
)

// Sources holds paths for both skillo scopes.
type Sources struct {
	Primary          string // primary skills dir (project or user-default)
	Secondary        string // user skills dir
	HomeDir          string
	UserSkilloDir    string // ~/.config/skillo/ — always set
	ProjectSkilloDir string // <git-root>/.skillo/ — empty if not in a repo
}

// Detect returns the skill sources based on the current working directory.
func Detect(homeDir, cwd string) *Sources {
	userSkills := UserSkillsDir(homeDir)
	userSkillo := UserSkilloDir(homeDir)

	projectRoot := findGitRoot(cwd)
	if projectRoot != "" {
		projectSkills := filepath.Join(projectRoot, ".agents", "skills")
		return &Sources{
			Primary:          projectSkills,
			Secondary:        userSkills,
			HomeDir:          homeDir,
			UserSkilloDir:    userSkillo,
			ProjectSkilloDir: ProjectSkilloDir(projectRoot),
		}
	}

	return &Sources{
		Primary:          userSkills,
		Secondary:        userSkills,
		HomeDir:          homeDir,
		UserSkilloDir:    userSkillo,
		ProjectSkilloDir: "",
	}
}

// DefaultDir returns the primary skills directory for the given cwd.
func DefaultDir(homeDir, cwd string) string {
	return Detect(homeDir, cwd).Primary
}

// UserSkillsDir returns the user-level skills extraction directory.
func UserSkillsDir(homeDir string) string {
	return filepath.Join(homeDir, ".agents", "skills")
}

// UserSkilloDir returns the user-level skillo config directory.
func UserSkilloDir(homeDir string) string {
	return filepath.Join(homeDir, ".config", "skillo")
}

// ProjectSkilloDir returns the project-level skillo directory.
func ProjectSkilloDir(root string) string {
	return filepath.Join(root, ".skillo")
}

// SkilloDirs returns the non-empty skillo directories in order:
// project first (if set), then user.
func SkilloDirs(s *Sources) []string {
	var dirs []string
	if s.ProjectSkilloDir != "" {
		dirs = append(dirs, s.ProjectSkilloDir)
	}
	if s.UserSkilloDir != "" {
		dirs = append(dirs, s.UserSkilloDir)
	}
	return dirs
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
