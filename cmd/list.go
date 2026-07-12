package cmd

import (
	"github.com/gunwooko/devx/internal/app"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List registered projects",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return app.ListProjects(cfgPath, cmd.OutOrStdout())
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
