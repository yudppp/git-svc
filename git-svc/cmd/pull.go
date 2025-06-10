package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/example/git-svc/pkg/utils" // Adjust module path
	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull <dir-path>",
	Short: "Pull latest changes for a service worktree",
	Long: `Performs a 'git pull --ff-only' in the worktree associated
with the given directory path (symlink).
The <dir-path> must be a symbolic link managed by git-svc.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dirPathArg := args[0]

		// Get symlink absolute path
		symlinkAbsPath, err := filepath.Abs(dirPathArg)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for '%s': %w", dirPathArg, err)
		}

		// Verify it's a symlink
		isLink, err := utils.IsSymlink(symlinkAbsPath)
		if err != nil {
			return fmt.Errorf("failed to check if '%s' is a symlink: %w", dirPathArg, err)
		}
		if !isLink {
			return fmt.Errorf("path '%s' is not a symbolic link. 'pull' command works on symlinks created by 'git-svc init'", dirPathArg)
		}

		// Resolve the symlink to get the path it points to within the worktree
		resolvedSymlinkTargetAbsPath, err := utils.ResolveSymlink(symlinkAbsPath)
		if err != nil {
			return fmt.Errorf("failed to resolve symlink '%s': %w", dirPathArg, err)
		}
		fmt.Printf("Symlink '%s' points to '%s'\n", dirPathArg, resolvedSymlinkTargetAbsPath)

		// Determine the root of the Git worktree containing resolvedSymlinkTargetAbsPath
		// Running `git rev-parse --show-toplevel` inside the resolved path will give the worktree's root.
		// This command needs to be run in a directory that is part of a git repository (the worktree).
		// If resolvedSymlinkTargetAbsPath is a file, we should use its directory.
		// However, `git-svc init` creates symlinks to directories. So, resolvedSymlinkTargetAbsPath should be a dir.
		worktreeRepoRoot, err := utils.RunGitCommandInDir(resolvedSymlinkTargetAbsPath, "rev-parse", "--show-toplevel")
		if err != nil {
			// Attempt to provide a more specific error message if the target path doesn't exist.
			// This could happen if the worktree was removed manually after the symlink was created.
			if _, statErr := filepath.EvalSymlinks(resolvedSymlinkTargetAbsPath); statErr != nil {
                 return fmt.Errorf("failed to determine worktree root for '%s' (resolved from '%s'). The target path does not exist or is not accessible. It might have been removed. Error: %w", resolvedSymlinkTargetAbsPath, dirPathArg, err)
            }
			return fmt.Errorf("failed to determine worktree root for '%s' (resolved from '%s'). Is it a valid git worktree and accessible? Error: %w", resolvedSymlinkTargetAbsPath, dirPathArg, err)
		}
		fmt.Printf("Worktree root identified as: %s\n", worktreeRepoRoot)

		// Perform git pull --ff-only in the worktree root
		fmt.Printf("Running 'git pull --ff-only' in '%s'...\n", worktreeRepoRoot)
		// The output from RunGitCommandInDir includes both stdout and stderr in the error if command fails.
		// If successful, output is stdout.
		output, err := utils.RunGitCommandInDir(worktreeRepoRoot, "pull", "--ff-only")
		if err != nil {
			// The error from RunGitCommandInDir already contains Stdout and Stderr details from the command.
			return fmt.Errorf("git pull --ff-only failed in '%s': %w", worktreeRepoRoot, err)
		}

		fmt.Println("Git pull successful.")
		if output != "" {
			fmt.Println("Output:")
			fmt.Println(output) // Print stdout from the command
		}

		return nil
	},
}

func init() {
	RootCmd.AddCommand(pullCmd)
}
