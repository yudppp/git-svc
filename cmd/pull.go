package cmd

import (
	"github.com/spf13/cobra"
	"github.com/yudppp/git-svc/svc"
)

var pullCmd = &cobra.Command{
	Use:   "pull <dir>",
	Short: "update worktree linked from dir",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return svc.Pull(args[0], worktreeRoot)
	},
}

func init() {
	rootCmd.AddCommand(pullCmd)
}
