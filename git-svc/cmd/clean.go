package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/example/git-svc/pkg/utils" // Adjust module path
	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean <dir-path>",
	Short: "Clean up a service worktree and symlink",
	Long: `Removes the symbolic link and the associated git worktree.
The <dir-path> must be a symbolic link previously managed by git-svc.
This command will fail if the worktree contains uncommitted changes (unless a --force flag is implemented and used).`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dirPathArg := args[0]

		symlinkAbsPath, err := filepath.Abs(dirPathArg)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for '%s': %w", dirPathArg, err)
		}

		isLink, err := utils.IsSymlink(symlinkAbsPath)
		if err != nil {
			// This error typically means lstat failed for a reason other than NotExist
			return fmt.Errorf("failed to check if '%s' is a symlink: %w", dirPathArg, err)
		}
		if !isLink {
			if _, statErr := os.Lstat(symlinkAbsPath); os.IsNotExist(statErr) {
				return fmt.Errorf("path '%s' does not exist", dirPathArg)
			}
			return fmt.Errorf("path '%s' is not a symbolic link. 'clean' command works on symlinks created by 'git-svc init'", dirPathArg)
		}

		resolvedSymlinkTargetAbsPath, err := utils.ResolveSymlink(symlinkAbsPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not resolve symlink '%s': %v\n", dirPathArg, err)
			fmt.Println("This may indicate a broken or unreadable symlink.")
			fmt.Println("Attempting to remove symlink file itself...")
			if errRemove := os.Remove(symlinkAbsPath); errRemove != nil {
				return fmt.Errorf("failed to remove symlink file '%s' (it was also unresolvable): %w", dirPathArg, errRemove)
			}
			// Use fmt.Errorf to ensure it's treated as an error by the CLI runner, but it's a "successful partial cleanup" message
			return fmt.Errorf("symlink file '%s' removed, but it was unresolvable. Associated worktree (if any) was not removed. Please check manually", dirPathArg)
		}
		fmt.Printf("Symlink '%s' points to '%s'\n", dirPathArg, resolvedSymlinkTargetAbsPath)

        if _, statErr := os.Stat(resolvedSymlinkTargetAbsPath); os.IsNotExist(statErr) {
			 fmt.Fprintf(os.Stderr, "Warning: Resolved symlink target '%s' does not exist (broken symlink).\n", resolvedSymlinkTargetAbsPath)
			 fmt.Println("Attempting to remove symlink file...")
			 if errRemove := os.Remove(symlinkAbsPath); errRemove != nil {
				return fmt.Errorf("failed to remove symlink file '%s' (its target did not exist): %w", dirPathArg, errRemove)
			 }
			 return fmt.Errorf("symlink file '%s' removed. Its target directory did not exist, so no worktree could be identified or removed. Please check manually", dirPathArg)
        }

		actualWorktreePath, err := utils.RunGitCommandInDir(resolvedSymlinkTargetAbsPath, "rev-parse", "--show-toplevel")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not determine worktree root from '%s' (resolved from symlink '%s'): %v\n", resolvedSymlinkTargetAbsPath, dirPathArg, err)
			fmt.Println("This could mean the target directory is not part of a git repository or worktree.")
			fmt.Println("Attempting to remove symlink file...")
			if errRemove := os.Remove(symlinkAbsPath); errRemove != nil {
				return fmt.Errorf("failed to remove symlink file '%s' (and its target was not identifiable as a worktree): %w", dirPathArg, errRemove)
			}
			return fmt.Errorf("symlink file '%s' removed, but its target did not appear to be in a git worktree. Worktree not removed. Please check manually", dirPathArg)
		}
		actualWorktreePath = filepath.Clean(actualWorktreePath)
		fmt.Printf("Identified worktree root as: %s\n", actualWorktreePath)

        mainRepoRoot, err := utils.GetGitRepoRoot()
        if err != nil {
            return fmt.Errorf("critical: failed to get main repository root: %w. Cannot proceed with worktree removal safely", err)
        }
		mainRepoRoot = filepath.Clean(mainRepoRoot)

        if actualWorktreePath == mainRepoRoot {
            fmt.Fprintf(os.Stderr, "Error: Symlink '%s' (pointing to '%s') resolves to the main repository path '%s', not a separate worktree.\n", dirPathArg, resolvedSymlinkTargetAbsPath, actualWorktreePath)
            fmt.Println("To protect your main repository, only the symlink will be removed.")
            if errRemove := os.Remove(symlinkAbsPath); errRemove != nil {
				return fmt.Errorf("failed to remove symlink '%s': %w. Target was the main repository", dirPathArg, errRemove)
			}
            return fmt.Errorf("symlink '%s' removed. Target was the main repository, so no worktree was removed. This is a safeguard. Please verify the symlink if this was not intended", dirPathArg)
        }

		// Remove the symbolic link first
		fmt.Printf("Removing symlink '%s'...\n", dirPathArg)
		if err := os.Remove(symlinkAbsPath); err != nil {
			return fmt.Errorf("failed to remove symlink '%s': %w. Worktree at '%s' was NOT removed", dirPathArg, err, actualWorktreePath)
		}
		fmt.Printf("Symlink '%s' removed successfully.\n", dirPathArg)

		fmt.Printf("Attempting to remove worktree at '%s'...\n", actualWorktreePath)
		// Use the absolute path for `actualWorktreePath` when calling `git worktree remove`
		// The command should be run from mainRepoRoot.
		output, err := utils.RunGitCommandInDir(mainRepoRoot, "worktree", "remove", actualWorktreePath)
		if err != nil {
			errMsg := fmt.Sprintf("Failed to remove worktree at '%s'. Symlink '%s' was already removed.", actualWorktreePath, dirPathArg)

			// Extract Stderr from the utility function's error if available
			errStr := err.Error()
			stdErrOutput := ""
			if strings.Contains(errStr, "Stderr: ") {
				parts := strings.SplitN(errStr, "Stderr: ", 2)
				if len(parts) > 1 {
					stdErrOutput = strings.TrimSpace(parts[1])
				}
			} else {
                stdErrOutput = errStr // Use full error if "Stderr: " not found (e.g. if error is not from RunCommand)
            }


			if stdErrOutput != "" {
                 errMsg += fmt.Sprintf("\nDetails from git: %s", stdErrOutput)
            }


			if strings.Contains(stdErrOutput, "is not a working tree") || strings.Contains(stdErrOutput, "no such file or directory") { // Git can use "no such..."
                 errMsg += fmt.Sprintf("\nInfo: Git reported that '%s' is not a working tree or does not exist. It might have been removed already or was not correctly registered by git.", actualWorktreePath)
            } else if strings.Contains(stdErrOutput, "working tree has modifications") || strings.Contains(stdErrOutput, "untracked files") {
				errMsg += fmt.Sprintf("\nHint: The worktree may have uncommitted changes or untracked files. If you are sure, you can try 'git worktree remove --force \"%s\"' manually from '%s'.", actualWorktreePath, mainRepoRoot)
			} else if stdErrOutput == "" && err != nil { // If no specific Stderr found, but there was an error
                errMsg += fmt.Sprintf("\nError details: %v", err)
            }
			return fmt.Errorf(errMsg)
		}

		fmt.Printf("Worktree at '%s' removed successfully.\n", actualWorktreePath)
		if strings.TrimSpace(output) != "" {
			fmt.Println("Output from 'git worktree remove':")
			fmt.Println(output)
		}

		return nil
	},
}

func init() {
	RootCmd.AddCommand(cleanCmd)
}
