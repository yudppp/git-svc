package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/example/git-svc/pkg/utils" // Adjusted module path
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init <dir-path> <branch>",
	Short: "Initialize a new service worktree and symlink",
	Long: `Creates a git worktree for the specified branch and links it
to the given directory path.
Example: git-svc init packages/my-service feature-branch`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		dirPathArg := args[0] // User-provided dir-path
		branch := args[1]

		fmt.Printf("Initializing worktree for dir '%s' on branch '%s'\n", dirPathArg, branch)

		if dirPathArg == "" || branch == "" {
			return fmt.Errorf("dir-path and branch must not be empty")
		}

		repoRoot, err := utils.GetGitRepoRoot()
		if err != nil {
			// If not in a git repo, GetGitRepoRoot will provide a detailed error.
			// We might want to create a .git directory here to simulate a repo for testing.
			// For now, let's assume `run_in_bash_session` provides a git repo context or we handle it.
			// The testing environment for these tools usually has a dummy git repo.
			return fmt.Errorf("could not determine git repository root: %w. Are you in a git repository?", err)
		}

		// dirPathArg is the path where the symlink will be created.
		// It should be interpreted relative to the repo root if not absolute.
		absoluteSymlinkLocationPath := dirPathArg
		if !filepath.IsAbs(dirPathArg) {
			absoluteSymlinkLocationPath = filepath.Join(repoRoot, dirPathArg)
		}
        absoluteSymlinkLocationPath = filepath.Clean(absoluteSymlinkLocationPath)

        // For user messages, show path relative to repo root.
        displaySymlinkPath, err := filepath.Rel(repoRoot, absoluteSymlinkLocationPath)
        if err != nil {
            // Fallback if Rel fails (e.g. different drives on Windows), use the cleaned absolute path or original arg
             fmt.Printf("Warning: could not determine path relative to repo root for %s: %v. Using original path for messages.\n", absoluteSymlinkLocationPath, err)
             displaySymlinkPath = dirPathArg
        }
        if displaySymlinkPath == "" || displaySymlinkPath == "." { // if dirPathArg was repoRoot or "."
            displaySymlinkPath = filepath.Base(absoluteSymlinkLocationPath) // Use the last part of the path
        }


		// Worktree path (e.g., <repoRoot>/.worktrees/<branch>)
		// CurrentWorktreeRoot is now set by RootCmd's PersistentPreRunE
		absoluteWorktreeParentDir := filepath.Join(repoRoot, CurrentWorktreeRoot)
		absoluteActualWorktreePath, err := utils.GetWorktreePath(absoluteWorktreeParentDir, branch) // This returns <absoluteWorktreeParentDir>/<branch>
		if err != nil {
			return err
		}
		// For user messages, show path relative to repo root.
        displayWorktreePath, _ := filepath.Rel(repoRoot, absoluteActualWorktreePath)
        if displayWorktreePath == "" { displayWorktreePath = absoluteActualWorktreePath}


		fmt.Printf("Worktree will be at: %s\n", displayWorktreePath)

		// Create .worktrees directory (e.g. <repoRoot>/.worktrees) if it doesn't exist
		if err := os.MkdirAll(absoluteWorktreeParentDir, 0755); err != nil {
			return fmt.Errorf("failed to create worktree root directory %s: %w", absoluteWorktreeParentDir, err)
		}

		// Check existing path at the symlink location
		lstatInfo, lerr := os.Lstat(absoluteSymlinkLocationPath)
		if lerr == nil { // Path exists
			isSymlink, _ := utils.IsSymlink(absoluteSymlinkLocationPath) // Error from IsSymlink not critical here if lstat succeeded
			if lstatInfo.IsDir() && !isSymlink {
				return fmt.Errorf("path %s already exists and is a directory. Please remove it or choose a different path", displaySymlinkPath)
			}
			if err := os.Remove(absoluteSymlinkLocationPath); err != nil {
				return fmt.Errorf("failed to remove existing file/symlink at %s: %w", displaySymlinkPath, err)
			}
			fmt.Printf("Removed existing file/symlink at %s\n", displaySymlinkPath)
		} else if !os.IsNotExist(lerr) {
            return fmt.Errorf("failed to stat path %s: %w", displaySymlinkPath, lerr)
        }

		// Ensure parent directory for the symlink itself exists
		parentDirForSymlink := filepath.Dir(absoluteSymlinkLocationPath)
		if parentDirForSymlink != "" && parentDirForSymlink != "." {
			if err := os.MkdirAll(parentDirForSymlink, 0755); err != nil {
				return fmt.Errorf("failed to create parent directory %s for symlink: %w", parentDirForSymlink, err)
			}
		}

		worktreePathForGitCmd, err := filepath.Rel(repoRoot, absoluteActualWorktreePath)
		if err != nil {
			return fmt.Errorf("failed to calculate relative path for git worktree add command from %s to %s: %w", repoRoot, absoluteActualWorktreePath, err)
		}

		fmt.Printf("Running: git worktree add %s %s (from %s)\n", worktreePathForGitCmd, branch, repoRoot)
		_, gitErr := utils.RunGitCommandInDir(repoRoot, "worktree", "add", worktreePathForGitCmd, branch)
		if gitErr != nil {
            errMsg := gitErr.Error()
            if strings.Contains(errMsg, "is already a worktree") || strings.Contains(errMsg, "already exists") || strings.Contains(errMsg, "path conflicts with an existing branch") {
                 // Check if the target worktree directory is empty. If not, it might be a failed previous attempt.
                 isEmpty := true
                 if entries, readDirErr := os.ReadDir(absoluteActualWorktreePath); readDirErr == nil {
                    if len(entries) > 0 {
                        isEmpty = false
                    }
                 } else if !os.IsNotExist(readDirErr) {
                    // If we can't read the directory and it's not because it doesn't exist, then it's an issue.
                    isEmpty = false
                 }

                 // If the directory is not empty, or if git specific error, guide user.
                 if !isEmpty || strings.Contains(errMsg, "is already a worktree") || strings.Contains(errMsg, "path conflicts with an existing branch"){
                    return fmt.Errorf("git worktree add failed: worktree path '%s' may already exist and be in use or conflict with branch name. Consider `git worktree remove %s` if it's stale, or use a different branch/path. Details: %w", displayWorktreePath, displayWorktreePath, gitErr)
                 }
                 // If it exists but is empty, could be remnant of failed `git worktree add` without checkout. Try to remove it.
                 // This scenario is less common with `git worktree add` itself but could happen.
                 fmt.Printf("Worktree path %s exists but seems empty or git error is ambiguous, attempting to proceed with caution or clean up.\n", displayWorktreePath)
            }
			return fmt.Errorf("git worktree add failed for path %s, branch %s: %w", displayWorktreePath, branch, gitErr)
		}
		fmt.Printf("Successfully added worktree for branch '%s' at '%s'\n", branch, displayWorktreePath)

        symlinkTargetDirAbsolute := filepath.Join(absoluteActualWorktreePath, displaySymlinkPath)

		relativeSymlinkTargetStr, err := utils.GetRelativeSymlinkTarget(absoluteSymlinkLocationPath, symlinkTargetDirAbsolute)
		if err != nil {
			return fmt.Errorf("failed to calculate relative symlink target: %w", err)
		}

		if err := os.Symlink(relativeSymlinkTargetStr, absoluteSymlinkLocationPath); err != nil {
			return fmt.Errorf("failed to create symlink from %s to %s: %w. On Windows, this might require admin privileges or Developer Mode", displaySymlinkPath, relativeSymlinkTargetStr, err)
		}

		fmt.Printf("Successfully created symlink: %s -> %s\n", displaySymlinkPath, relativeSymlinkTargetStr)

        ignorePath := displaySymlinkPath
        if strings.Contains(displaySymlinkPath, "/") {
            ignorePath = filepath.Dir(displaySymlinkPath) + "/"
        }
		fmt.Printf("Hint: Consider adding '%s' to your .gitignore file (e.g., by adding '%s').\n", displaySymlinkPath, ignorePath)
		return nil
	},
}

func init() {
	RootCmd.AddCommand(initCmd)
}
