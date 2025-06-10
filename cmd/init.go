package cmd

import (
	"github.com/spf13/cobra"
       "github.com/yudppp/git-svc/svc"
)

var initCmd = &cobra.Command{
	Use:   "init <dir> <branch>",
	Short: "add worktree and link directory",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return svc.Init(args[0], args[1], worktreeRoot)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
