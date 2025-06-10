package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	// Assuming utils package is "github.com/example/git-svc/pkg/utils"
	// "github.com/example/git-svc/pkg/utils"
)

const defaultWorktreeRoot = ".worktrees"
const envVarWorktreeRoot = "GIT_SVC_WORKTREE_ROOT"

// CurrentWorktreeRoot holds the effective worktree root path determined at runtime.
// It can be absolute or relative (to be joined with repo root by commands).
var CurrentWorktreeRoot string

var RootCmd = &cobra.Command{
	Use:   "git-svc",
	Short: "git-svc is a CLI tool to manage microservices in a monorepo using git worktrees",
	Long: `git-svc simplifies developing services in a monorepo by automating
the creation and management of git worktrees and symbolic links.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Determine CurrentWorktreeRoot
		// 1. Flag
		userSetWorktreeRoot, _ := cmd.Flags().GetString("worktree-root")

		if cmd.Flags().Changed("worktree-root") {
			CurrentWorktreeRoot = userSetWorktreeRoot
		} else {
			// 2. Environment Variable
			envWorktreeRoot := os.Getenv(envVarWorktreeRoot)
			if envWorktreeRoot != "" {
				CurrentWorktreeRoot = envWorktreeRoot
			} else {
				// 3. Default
				CurrentWorktreeRoot = defaultWorktreeRoot
			}
		}

        CurrentWorktreeRoot = filepath.Clean(CurrentWorktreeRoot)
		// No print statements here for cleaner output by default.
		// The effect of this variable will be seen in where worktrees are created or listed from.
		return nil
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		// Cobra's default error handling (Execute() returning an error)
		// will print the error to Stderr. No need for duplicative printing here.
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentFlags().String("worktree-root", defaultWorktreeRoot, fmt.Sprintf("Root directory for git-svc worktrees. If relative, it's based on the git repository root. (default: %s, env: %s)", defaultWorktreeRoot, envVarWorktreeRoot))
}
