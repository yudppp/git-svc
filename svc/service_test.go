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

	if err := Init("packages/a", "feature", "", ".worktrees", false, false); err != nil {
		t.Fatal(err)
	}

	if fi, err := os.Lstat("packages/a"); err != nil || fi.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("expected symlink at packages/a: %v", err)
	}
	if _, err := os.Stat("packages/a" + backupSuffix); err != nil {
		t.Fatalf("backup not created: %v", err)
	}

	if _, err := os.Stat(filepath.Join(repo, ".worktrees", "feature", "packages", "a", "file.txt")); err != nil {
		t.Fatalf("worktree file missing: %v", err)
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

	if fi, err := os.Lstat("packages/a"); err != nil || fi.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("directory not restored: %v", err)
	}
	if _, err := os.Stat("packages/a" + backupSuffix); !os.IsNotExist(err) {
		t.Fatalf("backup still exists")
	}
}

func TestInitNewBranchFromBase(t *testing.T) {
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
	if err := runCmd(repo, "git", "push", "-u", "origin", "master"); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	os.Chdir(repo)
	defer os.Chdir(cwd)

	if err := Init("packages/a", "feat-a", "origin/master", ".worktrees", true, false); err != nil {
		t.Fatal(err)
	}

	if fi, err := os.Lstat("packages/a"); err != nil || fi.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("expected symlink: %v", err)
	}
	if _, err := os.Stat(filepath.Join(repo, ".worktrees", "feat-a", "packages", "a", "file.txt")); err != nil {
		t.Fatalf("worktree file missing: %v", err)
	}
}

func TestInitSparse(t *testing.T) {
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

	if err := Init("packages/a", "feature", "", ".worktrees", false, true); err != nil {
		t.Fatal(err)
	}

	entries, err := os.ReadDir(filepath.Join(repo, ".worktrees", "feature", "packages"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "a" {
		t.Fatalf("sparse checkout not applied: %v", entries)
	}

	if err := Clean("packages/a", ".worktrees"); err != nil {
		t.Fatal(err)
	}
}
