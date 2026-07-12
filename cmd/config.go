package cmd

import (
	"github.com/gunwooko/devx/internal/app"
	"github.com/spf13/cobra"
)

var (
	configDefaultDir   string
	configDefaultAgent string
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show or update global devx configuration",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return app.Configure(app.ConfigureOptions{
			ConfigPath:   cfgPath,
			DefaultDir:   configDefaultDir,
			DefaultAgent: configDefaultAgent,
			Output:       cmd.OutOrStdout(),
		})
	},
}

func init() {
	configCmd.Flags().StringVar(&configDefaultDir, "default-dir", "", "default directory for newly created projects")
	configCmd.Flags().StringVar(&configDefaultAgent, "default-agent", "", "default agent: claude, codex, or none")
	rootCmd.AddCommand(configCmd)
}
