package cmd

import (
	"fmt"

	"git-svc/svc"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list managed symlinks",
	RunE: func(cmd *cobra.Command, args []string) error {
		m, err := svc.List(worktreeRoot)
		if err != nil {
			return err
		}
		for dir, br := range m {
			fmt.Printf("%s -> %s\n", dir, br)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
