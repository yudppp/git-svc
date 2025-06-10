package svc

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const backupSuffix = ".gitsvc_backup"

// RepoRoot returns the absolute path of the repository root.
func RepoRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse failed: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// Init adds a worktree for the branch and links the directory to it.
// If create is true, a new branch is created with `-b branch` and optionally
// based on the provided base commitish.
func Init(dir, branch, base, root string, create, sparse bool) error {
	repoRoot, err := RepoRoot()
	if err != nil {
		return err
	}
	worktreePath := filepath.Join(repoRoot, root, branch)
	args := []string{"git", "worktree", "add"}
	if sparse {
		args = append(args, "--no-checkout")
	}
	if create {
		args = append(args, "-b", branch, worktreePath)
		if base != "" {
			args = append(args, base)
		}
	} else {
		args = append(args, worktreePath, branch)
	}
	if err := runCmd(repoRoot, args[0], args[1:]...); err != nil {
		return err
	}
	if sparse {
		if err := runCmd(worktreePath, "git", "sparse-checkout", "init", "--cone"); err != nil {
			return err
		}
		if err := runCmd(worktreePath, "git", "sparse-checkout", "set", dir); err != nil {
			return err
		}
		if err := runCmd(worktreePath, "git", "reset", "--hard", "HEAD"); err != nil {
			return err
		}
	}
	target := filepath.Join(worktreePath, dir)
	link := filepath.Join(repoRoot, dir)
	if fi, err := os.Lstat(link); err == nil {
		if fi.Mode()&os.ModeSymlink != 0 {
			if err := os.Remove(link); err != nil {
				return err
			}
		} else {
			backup := link + backupSuffix
			if _, err := os.Stat(backup); err == nil {
				return fmt.Errorf("backup already exists: %s", backup)
			} else if !os.IsNotExist(err) {
				return err
			}
			if err := os.Rename(link, backup); err != nil {
				return err
			}
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	rel, err := filepath.Rel(filepath.Dir(link), target)
	if err != nil {
		return err
	}
	return os.Symlink(rel, link)
}

// Pull updates the worktree linked from dir.
func Pull(dir, root string) error {
	repoRoot, err := RepoRoot()
	if err != nil {
		return err
	}
	branch, err := branchFromLink(filepath.Join(repoRoot, dir))
	if err != nil {
		return err
	}
	worktreePath := filepath.Join(repoRoot, root, branch)
	return runCmd(worktreePath, "git", "pull", "--ff-only")
}

// Clean removes the symlink and its worktree.
func Clean(dir, root string) error {
	repoRoot, err := RepoRoot()
	if err != nil {
		return err
	}
	link := filepath.Join(repoRoot, dir)
	branch, err := branchFromLink(link)
	if err != nil {
		return err
	}
	worktreePath := filepath.Join(repoRoot, root, branch)
	if err := runCmd(repoRoot, "git", "worktree", "remove", worktreePath); err != nil {
		return err
	}
	if err := os.Remove(link); err != nil && !os.IsNotExist(err) {
		return err
	}
	backup := link + backupSuffix
	if _, err := os.Stat(backup); err == nil {
		if err := os.Rename(backup, link); err != nil {
			return err
		}
	}
	return nil
}

// List returns map of dir -> branch.
func List(root string) (map[string]string, error) {
	repoRoot, err := RepoRoot()
	if err != nil {
		return nil, err
	}
	m := make(map[string]string)
	err = filepath.WalkDir(repoRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			if os.IsPermission(err) {
				// Ignore permission errors and continue walking
				return nil
			}
			return err
		}
		if d.Type()&os.ModeSymlink == 0 {
			return nil
		}
		target, err := filepath.EvalSymlinks(path)
		if err != nil {
			return nil
		}
		rel, err := filepath.Rel(filepath.Join(repoRoot, root), target)
		if err != nil || strings.Contains(rel, "..") {
			return nil
		}
		parts := strings.SplitN(rel, string(filepath.Separator), 2)
		if len(parts) == 0 {
			return nil
		}
		branch := parts[0]
		dirRel, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return nil
		}
		m[dirRel] = branch
		return nil
	})
	return m, err
}

func branchFromLink(link string) (string, error) {
	target, err := filepath.EvalSymlinks(link)
	if err != nil {
		return "", err
	}
	// target like /repo/.worktrees/branch/dir
	idx := strings.Index(target, string(filepath.Separator)+".worktrees"+string(filepath.Separator))
	if idx == -1 {
		return "", errors.New("link target not in worktrees")
	}
	rest := target[idx+len(string(filepath.Separator))+len(".worktrees")+1:]
	parts := strings.SplitN(rest, string(filepath.Separator), 2)
	if len(parts) == 0 {
		return "", errors.New("cannot parse branch")
	}
	return parts[0], nil
}

func runCmd(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
