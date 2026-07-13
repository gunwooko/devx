package cmd

import (
	"github.com/gunwooko/devx/internal/app"
	"github.com/spf13/cobra"
)

var (
	importAgent  string
	importDryRun bool
)

var importCmd = &cobra.Command{
	Use:   "import <dir>",
	Short: "Register every subdirectory as a project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return app.ImportProjects(app.ImportOptions{
			ConfigPath: cfgPath,
			Dir:        args[0],
			Agent:      importAgent,
			DryRun:     importDryRun,
			Output:     cmd.OutOrStdout(),
		})
	},
}

func init() {
	importCmd.Flags().StringVarP(&importAgent, "agent", "a", "", "AI agent for imported projects: claude, codex, gemini, opencode, none, or a custom agent")
	importCmd.Flags().BoolVarP(&importDryRun, "dry-run", "n", false, "show what would be imported without saving")
	rootCmd.AddCommand(importCmd)
}
