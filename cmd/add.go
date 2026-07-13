package cmd

import (
	"github.com/gunwooko/devx/internal/app"
	"github.com/spf13/cobra"
)

var addAgent string

var addCmd = &cobra.Command{
	Use:   "add <name> <path>",
	Short: "Register an existing project",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return app.AddProject(app.AddOptions{
			ConfigPath: cfgPath,
			Name:       args[0],
			Path:       args[1],
			Agent:      addAgent,
			Output:     cmd.OutOrStdout(),
		})
	},
}

func init() {
	addCmd.Flags().StringVarP(&addAgent, "agent", "a", "", "AI agent: claude, codex, gemini, opencode, none, or a custom agent")
	rootCmd.AddCommand(addCmd)
}
