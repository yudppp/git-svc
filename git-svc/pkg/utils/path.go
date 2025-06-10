package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// GetWorktreePath calculates the full path to a specific worktree.
// worktreeBaseDir is the root directory for all worktrees (e.g., ".worktrees").
// branchName is the name of the branch, also used as the worktree directory name.
func GetWorktreePath(worktreeBaseDir, branchName string) (string, error) {
	absWorktreeBaseDir, err := filepath.Abs(worktreeBaseDir)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for worktree base %s: %w", worktreeBaseDir, err)
	}
	return filepath.Join(absWorktreeBaseDir, branchName), nil
}

// GetRelativeSymlinkTarget calculates the relative path for a symlink.
// It computes the path from the symlink's own parent directory to the target path.
//
// symlinkLocationAbs: Absolute path where the symlink will be created (e.g., /repo/packages/service-a).
// targetPathAbs: Absolute path the symlink should point to (e.g., /repo/.worktrees/branch/packages/service-a).
func GetRelativeSymlinkTarget(symlinkLocationAbs string, targetPathAbs string) (string, error) {
	symlinkParentDir := filepath.Dir(symlinkLocationAbs)
	relativeTarget, err := filepath.Rel(symlinkParentDir, targetPathAbs)
	if err != nil {
		return "", fmt.Errorf("failed to calculate relative path from %s to %s: %w", symlinkParentDir, targetPathAbs, err)
	}
	return relativeTarget, nil
}

// IsSymlink checks if the given path is a symbolic link.
func IsSymlink(path string) (bool, error) {
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil // Doesn't exist, so not a symlink
		}
		return false, err // Other error
	}
	return info.Mode()&os.ModeSymlink != 0, nil
}

// ResolveSymlink resolves a symbolic link to its absolute target path.
// If the path is not a symlink, it returns the absolute path of the original path.
// If the path doesn't exist, it returns an error.
func ResolveSymlink(path string) (string, error) {
    absPath, err := filepath.Abs(path)
    if err != nil {
        return "", fmt.Errorf("failed to get absolute path for %s: %w", path, err)
    }

    lstatInfo, err := os.Lstat(absPath)
    if err != nil {
        return "", fmt.Errorf("lstat failed for %s: %w", absPath, err)
    }

    if lstatInfo.Mode()&os.ModeSymlink != 0 {
        // It's a symlink, read it
        resolved, err := os.Readlink(absPath)
        if err != nil {
            return "", fmt.Errorf("failed to readlink %s: %w", absPath, err)
        }
        // If the link is relative, make it absolute with respect to the link's parent directory
        if !filepath.IsAbs(resolved) {
            resolved = filepath.Join(filepath.Dir(absPath), resolved)
        }
        return filepath.Clean(resolved), nil
    }
    // Not a symlink, just return its (now confirmed) absolute path
    return absPath, nil
}

// GetGitRepoRoot finds the root directory of the git repository containing the current path.
// It does this by running `git rev-parse --show-toplevel`.
func GetGitRepoRoot() (string, error) {
    output, err := RunGitCommand("rev-parse", "--show-toplevel")
    if err != nil {
        return "", fmt.Errorf("failed to find git repository root: %w. Ensure you are in a git repository", err)
    }
    return output, nil
}
