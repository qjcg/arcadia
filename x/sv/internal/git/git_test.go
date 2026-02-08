package git

import (
	"os"
	"os/exec"
	"testing"
)

func setupGitRepo(t *testing.T) string {
	tmpDir, err := os.MkdirTemp("", "sv-git-test")
	if err != nil {
		t.Fatal(err)
	}

	runGit := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("git %v failed: %v", args, err)
		}
	}

	runGit("init")
	runGit("config", "user.email", "test@example.com")
	runGit("config", "user.name", "Test User")
	runGit("commit", "--allow-empty", "-m", "initial commit")

	return tmpDir
}

func TestLatestTag(t *testing.T) {
	tmpDir := setupGitRepo(t)
	defer os.RemoveAll(tmpDir)

	origWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origWd)

	runGit := func(args ...string) {
		cmd := exec.Command("git", args...)
		if err := cmd.Run(); err != nil {
			t.Fatalf("git %v failed: %v", args, err)
		}
	}

	// No tags yet
	tag, err := LatestTag(tmpDir, ".")
	if err != nil || tag != "" {
		t.Errorf("expected no tag, got %q, err %v", tag, err)
	}

	runGit("tag", "v1.0.0")
	runGit("tag", "x/mod/v2.0.0")

	tag, _ = LatestTag(tmpDir, ".")
	if tag != "v1.0.0" {
		t.Errorf("expected v1.0.0, got %q", tag)
	}

	tag, _ = LatestTag(tmpDir, "x/mod")
	if tag != "x/mod/v2.0.0" {
		t.Errorf("expected x/mod/v2.0.0, got %q", tag)
	}
}
