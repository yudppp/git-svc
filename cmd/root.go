package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var worktreeRoot string

var rootCmd = &cobra.Command{
	Use:   "git-svc",
	Short: "Manage git worktrees with symlinks",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(
		&worktreeRoot,
		"worktree-root",
		getenvDefault("GITSVC_WORKTREE_ROOT", ".worktrees"),
		"worktree root directory",
	)
}

func getenvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
