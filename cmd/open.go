package cmd

import (
	"github.com/gunwooko/devx/internal/app"
	"github.com/spf13/cobra"
)

var openAgent string

var openCmd = &cobra.Command{
	Use:   "open <name>",
	Short: "Open a project in tmux",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return app.OpenProject(app.OpenOptions{
			ConfigPath:    cfgPath,
			Name:          args[0],
			AgentOverride: openAgent,
		})
	},
}

func init() {
	openCmd.Flags().StringVarP(&openAgent, "agent", "a", "", "override the configured agent for this new session")
	rootCmd.AddCommand(openCmd)
}
