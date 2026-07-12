package cmd

import (
	"github.com/gunwooko/devx/internal/app"
	"github.com/spf13/cobra"
)

var (
	createPath   string
	createAgent  string
	createNoGit  bool
	createNoOpen bool
	createYes    bool
)

var createCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create and register a new project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return app.CreateProject(app.CreateOptions{
			ConfigPath: cfgPath,
			Name:       args[0],
			Path:       createPath,
			Agent:      createAgent,
			InitGit:    !createNoGit,
			Open:       !createNoOpen,
			AssumeYes:  createYes,
		})
	},
}

func init() {
	createCmd.Flags().StringVarP(&createPath, "path", "p", "", "project path (default: <defaultProjectsDir>/<name>)")
	createCmd.Flags().StringVarP(&createAgent, "agent", "a", "", "AI agent: claude, codex, or none")
	createCmd.Flags().BoolVar(&createNoGit, "no-git", false, "do not initialize a Git repository")
	createCmd.Flags().BoolVar(&createNoOpen, "no-open", false, "create without opening the project")
	createCmd.Flags().BoolVarP(&createYes, "yes", "y", false, "accept defaults without prompting")
	rootCmd.AddCommand(createCmd)
}
