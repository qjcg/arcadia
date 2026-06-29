package git

import (
	"os"
	"os/exec"
	"reflect"
	"testing"
)

func runGit(t *testing.T, dir string, args ...string) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\nOutput: %s", args, err, string(out))
	}
}

func setupGitRepo(t *testing.T) string {
	tmpDir := t.TempDir()

	runGit(t, tmpDir, "init")
	runGit(t, tmpDir, "config", "user.email", "test@example.com")
	runGit(t, tmpDir, "config", "user.name", "Test User")
	runGit(t, tmpDir, "config", "commit.gpgsign", "false")
	runGit(t, tmpDir, "commit", "--allow-empty", "-m", "initial commit")

	return tmpDir
}

func TestLatestTag(t *testing.T) {
	tmpDir := setupGitRepo(t)

	// No tags yet
	tag, err := LatestTag(tmpDir, ".")
	if err != nil || tag != "" {
		t.Errorf("expected no tag, got %q, err %v", tag, err)
	}

	runGit(t, tmpDir, "tag", "v1.0.0")
	runGit(t, tmpDir, "tag", "x/mod/v2.0.0")
	runGit(t, tmpDir, "commit", "--allow-empty", "-m", "another commit")
	runGit(t, tmpDir, "tag", "v1.1.0")

	tag, err = LatestTag(tmpDir, ".")
	if err != nil {
		t.Fatal(err)
	}
	if tag != "v1.1.0" {
		t.Errorf("expected v1.1.0, got %q", tag)
	}

	tag, err = LatestTag(tmpDir, "x/mod")
	if err != nil {
		t.Fatal(err)
	}
	if tag != "x/mod/v2.0.0" {
		t.Errorf("expected x/mod/v2.0.0, got %q", tag)
	}
}

func TestCommitsSince(t *testing.T) {
	tmpDir := setupGitRepo(t)

	// Create some commits
	if err := os.MkdirAll(tmpDir+"/x/mod", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tmpDir+"/x/mod/file.txt", []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, tmpDir, "add", ".")
	runGit(t, tmpDir, "commit", "-m", "feat: added file in x/mod")
	runGit(t, tmpDir, "tag", "x/mod/v1.0.0")

	runGit(t, tmpDir, "commit", "--allow-empty", "-m", "fix: root fix")
	runGit(t, tmpDir, "commit", "--allow-empty", "-m", "feat: another root feat")

	if err := os.WriteFile(tmpDir+"/x/mod/file.txt", []byte("hello world"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, tmpDir, "add", ".")
	runGit(t, tmpDir, "commit", "-m", "fix: module fix")

	// Test commits since tag for root
	var commits []string
	var err error
	runGit(t, tmpDir, "tag", "v1.0.0")
	runGit(t, tmpDir, "commit", "--allow-empty", "-m", "feat: after v1.0.0")

	commits, err = CommitsSince(tmpDir, "v1.0.0", ".", nil)
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{"feat: after v1.0.0"}
	if !reflect.DeepEqual(commits, expected) {
		t.Errorf("expected %v, got %v", expected, commits)
	}

	// Test commits since tag for module
	commits, err = CommitsSince(tmpDir, "x/mod/v1.0.0", "x/mod", nil)
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
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Logf("failed to restore directory: %v", err)
		}
	}()

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

func TestCommitsSince_NonExistentTag(t *testing.T) {
	tmpDir := setupGitRepo(t)

	// Create a tag first
	runGit(t, tmpDir, "commit", "--allow-empty", "-m", "first commit")
	runGit(t, tmpDir, "tag", "v0.1.0")

	// Create more commits
	runGit(t, tmpDir, "commit", "--allow-empty", "-m", "second commit")
	runGit(t, tmpDir, "commit", "--allow-empty", "-m", "third commit")

	// Non-existent tag should fail (git returns error for unknown tag in range)
	_, err := CommitsSince(tmpDir, "nonexistent-tag", ".", nil)
	if err == nil {
		t.Error("expected error for non-existent tag, got nil")
	}
}

func TestCommitsSince_WithExcludePaths(t *testing.T) {
	tmpDir := setupGitRepo(t)

	// Create directories
	if err := os.MkdirAll(tmpDir+"/vendor", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(tmpDir+"/internal/pkg", 0o755); err != nil {
		t.Fatal(err)
	}

	// Create files in different dirs and commit
	if err := os.WriteFile(tmpDir+"/main.go", []byte("main"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tmpDir+"/vendor/dep.go", []byte("dep"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tmpDir+"/internal/pkg/util.go", []byte("util"), 0o644); err != nil {
		t.Fatal(err)
	}

	runGit(t, tmpDir, "add", ".")
	runGit(t, tmpDir, "commit", "-m", "feat: add features")

	// Create a baseline tag
	runGit(t, tmpDir, "tag", "v0.1.0")

	// Modify files again
	if err := os.WriteFile(tmpDir+"/main.go", []byte("main v2"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tmpDir+"/vendor/dep.go", []byte("dep v2"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tmpDir+"/internal/pkg/util.go", []byte("util v2"), 0o644); err != nil {
		t.Fatal(err)
	}

	runGit(t, tmpDir, "add", ".")
	runGit(t, tmpDir, "commit", "-m", "feat: update features")

	// Test excluding vendor - when checking commits after v0.1.0 with vendor excluded
	// The commit touches main.go (include), vendor/dep.go (exclude), internal/pkg/util.go (include)
	// So the exclude should filter it out partially, but since it's a single commit,
	// the behavior depends on whether ALL files are excluded or ANY
	commits, err := CommitsSince(tmpDir, "v0.1.0", ".", []string{"vendor"})
	if err != nil {
		t.Fatal(err)
	}

	// The excludePaths logic checks each commit - if any file is NOT in excluded paths, include the commit
	// So this commit should still be included since main.go and internal/pkg/util.go are not excluded
	if len(commits) != 1 {
		t.Errorf("expected 1 commit (commit touches non-excluded files), got %d: %v", len(commits), commits)
	}
}

func TestLatestTag_NonExistentPath(t *testing.T) {
	tmpDir := setupGitRepo(t)

	// No tags at all - should return empty
	tag, err := LatestTag(tmpDir, "nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if tag != "" {
		t.Errorf("expected no tag, got %q", tag)
	}
}
