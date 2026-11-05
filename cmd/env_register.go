package cmd

import (
	"fmt"
	"path/filepath"
	"rpy/internal/env"

	"github.com/spf13/cobra"
)

var registerActivate bool
var registerShell string

var envRegisterCmd = &cobra.Command{
	Use:   "register <name|local>",
	Short: "Save the project environment selection in .rpy-env",
	Long:  "Save a shared environment for this project. Register local to clear the saved selection and return to .venv.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		root, err := env.ProjectRoot()
		if err != nil {
			return err
		}
		var activation string
		if registerActivate {
			activation, err = env.SelectionActivationCommand(name, registerShell, false)
			if err != nil {
				return err
			}
		}

		path := filepath.Join(root, ".rpy-env")
		if name == "local" {
			if err := env.UseLocalEnv(); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "Using local .venv by default in %s (cleared %s)\n", root, path)
		} else {
			if err := env.UseSharedEnv(name); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "Saved %s to %s\n", name, path)
		}
		if registerActivate {
			fmt.Fprintln(cmd.OutOrStdout(), activation)
		}
		return nil
	},
}

func init() {
	envCmd.AddCommand(envRegisterCmd)
	envRegisterCmd.Flags().BoolVar(&registerActivate, "activate", false, "Activate the saved environment after registering it")
	envRegisterCmd.Flags().StringVar(&registerShell, "shell", "", "Shell to target when activating")
}
