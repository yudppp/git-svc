package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yudppp/git-svc/svc"
)

var branchFlag string
var sparseFlag bool

var initCmd = &cobra.Command{
	Use:   "init <dir> [base]",
	Short: "add worktree and link directory",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := args[0]
		if branchFlag != "" {
			var base string
			if len(args) == 2 {
				base = args[1]
			}
			return svc.Init(dir, branchFlag, base, worktreeRoot, true, sparseFlag)
		}
		if len(args) < 2 {
			return fmt.Errorf("branch required")
		}
		return svc.Init(dir, args[1], "", worktreeRoot, false, sparseFlag)
	},
}

func init() {
	initCmd.Flags().StringVarP(&branchFlag, "branch", "b", "", "create new branch")
	initCmd.Flags().BoolVar(&sparseFlag, "sparse", false, "use git sparse-checkout")
	rootCmd.AddCommand(initCmd)
}
