package cmd

import (
	"rpy/internal/env"

	"github.com/spf13/cobra"
)

var envUseActivate bool

var envUseCmd = &cobra.Command{
	Use:        "use <name|local>",
	Deprecated: "use env register <name> to save a selection or env activate <name> for this shell",
	Short:      "Select the environment for the current project",
	Long: `Select the environment for the current project.

Without --activate, this persists the selection in the project.
With --activate, the selection is temporary for the current shell session and does not write a project file.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if envUseActivate {
			return env.PrintUseAndActivateCommand(args[0], "")
		}

		if args[0] == "local" {
			if err := env.UseLocalEnv(); err != nil {
				return err
			}
		} else {
			if err := env.UseSharedEnv(args[0]); err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	envCmd.AddCommand(envUseCmd)
	envUseCmd.Flags().BoolVar(&envUseActivate, "activate", false, "Temporarily activate the selection for the current shell without writing a project file")
}
