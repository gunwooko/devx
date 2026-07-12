package cmd

import (
	"github.com/gunwooko/devx/internal/app"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show project and tmux session status",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return app.Status(cfgPath, cmd.OutOrStdout())
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
