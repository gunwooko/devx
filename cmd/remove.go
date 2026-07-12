package cmd

import (
	"github.com/gunwooko/devx/internal/app"
	"github.com/spf13/cobra"
)

var removeForce bool

var removeCmd = &cobra.Command{
	Use:     "remove <name>",
	Aliases: []string{"rm"},
	Short:   "Remove a project from devx without deleting its files",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return app.RemoveProject(cfgPath, args[0], removeForce, cmd.OutOrStdout())
	},
}

func init() {
	removeCmd.Flags().BoolVarP(&removeForce, "force", "f", false, "stop an active tmux session before removing")
	rootCmd.AddCommand(removeCmd)
}
