package cmd

import (
	"git-svc/svc"
	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean <dir>",
	Short: "remove symlink and worktree",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return svc.Clean(args[0], worktreeRoot)
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
