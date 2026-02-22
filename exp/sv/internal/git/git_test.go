package git

import (
	"os"
	"os/exec"
	"reflect"
	"testing"
)

func setupGitRepo(t *testing.T) string {
	tmpDir := t.TempDir()

	runGit := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = tmpDir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\nOutput: %s", args, err, string(out))
		}
	}

	runGit("init")
	runGit("config", "user.email", "test@example.com")
	runGit("config", "user.name", "Test User")
	runGit("config", "commit.gpgsign", "false")
	runGit("commit", "--allow-empty", "-m", "initial commit")

	return tmpDir
}

func TestLatestTag(t *testing.T) {
	tmpDir := setupGitRepo(t)

	runGit := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = tmpDir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\nOutput: %s", args, err, string(out))
		}
	}

	// No tags yet
	tag, err := LatestTag(tmpDir, ".")
	if err != nil || tag != "" {
		t.Errorf("expected no tag, got %q, err %v", tag, err)
	}

	runGit("tag", "v1.0.0")
	runGit("tag", "x/mod/v2.0.0")
	runGit("commit", "--allow-empty", "-m", "another commit")
	runGit("tag", "v1.1.0")

	tag, _ = LatestTag(tmpDir, ".")
	if tag != "v1.1.0" {
		t.Errorf("expected v1.1.0, got %q", tag)
	}

	tag, _ = LatestTag(tmpDir, "x/mod")
	if tag != "x/mod/v2.0.0" {
		t.Errorf("expected x/mod/v2.0.0, got %q", tag)
	}
}

func TestCommitsSince(t *testing.T) {
	tmpDir := setupGitRepo(t)

	runGit := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = tmpDir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\nOutput: %s", args, err, string(out))
		}
	}

	// Create some commits
	if err := os.MkdirAll(tmpDir+"/x/mod", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tmpDir+"/x/mod/file.txt", []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit("add", ".")
	runGit("commit", "-m", "feat: added file in x/mod")
	runGit("tag", "x/mod/v1.0.0")

	runGit("commit", "--allow-empty", "-m", "fix: root fix")
	runGit("commit", "--allow-empty", "-m", "feat: another root feat")

	if err := os.WriteFile(tmpDir+"/x/mod/file.txt", []byte("hello world"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit("add", ".")
	runGit("commit", "-m", "fix: module fix")

	// Test commits since tag for root
	var commits []string
	var err error
	runGit("tag", "v1.0.0")
	runGit("commit", "--allow-empty", "-m", "feat: after v1.0.0")

	commits, err = CommitsSince(tmpDir, "v1.0.0", ".")
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{"feat: after v1.0.0"}
	if !reflect.DeepEqual(commits, expected) {
		t.Errorf("expected %v, got %v", expected, commits)
	}

	// Test commits since tag for module
	commits, err = CommitsSince(tmpDir, "x/mod/v1.0.0", "x/mod")
	if err != nil {
		t.Fatal(err)
	}
	expected = []string{"fix: module fix"}
	if !reflect.DeepEqual(commits, expected) {
		t.Errorf("expected %v, got %v", expected, commits)
	}
}

func TestRoot(t *testing.T) {
	tmpDir := setupGitRepo(t)

	// Change to the temp directory and get root
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	root, err := Root()
	if err != nil {
		t.Fatalf("Root() failed: %v", err)
	}

	// Root should be the temp directory
	if root != tmpDir {
		t.Errorf("Root() = %q, want %q", root, tmpDir)
	}
}
