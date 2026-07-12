package cmd

import (
	"fmt"
	"os"

	"github.com/gunwooko/devx/internal/app"
	"github.com/gunwooko/devx/internal/config"
	"github.com/spf13/cobra"
)

var (
	cfgPath string
	version = "dev"
)

var rootCmd = &cobra.Command{
	Use:   "devx",
	Short: "Manage AI coding projects and tmux sessions",
	Long: `devx creates, registers, opens, and manages projects that run inside tmux.

Each project can use a default AI CLI such as Claude Code or Codex.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	Args:          cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		if len(args) > 1 {
			return fmt.Errorf("expected one project name; got %d arguments", len(args))
		}
		return app.OpenProject(app.OpenOptions{
			ConfigPath: cfgPath,
			Name:       args[0],
			Output:     cmd.OutOrStdout(),
		})
	},
}

func Execute() {
	rootCmd.Version = version
	rootCmd.SetVersionTemplate("devx {{.Version}}\n")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func init() {
	defaultConfig := "<os config dir>/devx/config.json"
	if p, err := config.Path(""); err == nil {
		defaultConfig = p
	}
	rootCmd.PersistentFlags().StringVar(&cfgPath, "config", "", "config file path (default: "+defaultConfig+")")
}
