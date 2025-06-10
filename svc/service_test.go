package svc

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitPullCleanList(t *testing.T) {
	repo := t.TempDir()
	if err := runCmd(repo, "git", "init"); err != nil {
		t.Fatal(err)
	}
	remote := filepath.Join(repo, "remote.git")
	if err := runCmd("", "git", "init", "--bare", remote); err != nil {
		t.Fatal(err)
	}
	if err := runCmd(repo, "git", "remote", "add", "origin", remote); err != nil {
		t.Fatal(err)
	}
	// create sample dir
	if err := os.MkdirAll(filepath.Join(repo, "packages/a"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "packages/a/file.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := runCmd(repo, "git", "add", "."); err != nil {
		t.Fatal(err)
	}
	if err := runCmd(repo, "git", "commit", "-m", "init"); err != nil {
		t.Fatal(err)
	}
	if err := runCmd(repo, "git", "branch", "feature"); err != nil {
		t.Fatal(err)
	}
	if err := runCmd(repo, "git", "push", "-u", "origin", "master"); err != nil {
		t.Fatal(err)
	}
	if err := runCmd(repo, "git", "push", "-u", "origin", "feature"); err != nil {
		t.Fatal(err)
	}
	if err := runCmd(repo, "git", "branch", "--set-upstream-to=origin/feature", "feature"); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	os.Chdir(repo)
	defer os.Chdir(cwd)

	if err := Init("packages/a", "feature", ".worktrees"); err != nil {
		t.Fatal(err)
	}

	m, err := List(".worktrees")
	if err != nil {
		t.Fatal(err)
	}
	if m["packages/a"] != "feature" {
		t.Fatalf("expected mapping, got %v", m)
	}

	if err := Pull("packages/a", ".worktrees"); err != nil {
		t.Fatal(err)
	}

	if err := Clean("packages/a", ".worktrees"); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Lstat("packages/a"); !os.IsNotExist(err) {
		t.Fatalf("symlink not removed: %v", err)
	}
}
