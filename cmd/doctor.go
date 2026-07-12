package cmd

import (
	"github.com/gunwooko/devx/internal/app"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check devx dependencies and configuration",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return app.Doctor(cfgPath, cmd.OutOrStdout())
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
