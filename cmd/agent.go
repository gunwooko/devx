package cmd

import (
	"github.com/gunwooko/devx/internal/app"
	"github.com/spf13/cobra"
)

var agentCmd = &cobra.Command{
	Use:   "agent <name> <agent>",
	Short: "Change a project's default AI agent",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return app.SetAgent(cfgPath, args[0], args[1], cmd.OutOrStdout())
	},
}

func init() {
	rootCmd.AddCommand(agentCmd)
}
