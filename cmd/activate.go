package cmd

import (
	"rpy/internal/env"

	"github.com/spf13/cobra"
)

var activateShell string
var activatePathOnly bool
var activateCurrent bool

var activateCmd = &cobra.Command{
	Use:   "activate [name|local]",
	Short: "Activate a project or named environment for this shell",
	Long: `Activate the registered environment or local .venv. Pass a name to override
the project selection for this shell without writing .rpy-env.

Because a child process cannot modify the parent shell environment, this
command prints the shell command you should evaluate, for example:

  eval "$(rpy env activate)"
  source "$(rpy env activate --path)"`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := ""
		if len(args) == 1 {
			name = args[0]
		}
		return env.PrintActivation(name, activateShell, activatePathOnly, activateCurrent)
	},
}

func init() {
	envCmd.AddCommand(activateCmd)
	activateCmd.Flags().StringVar(&activateShell, "shell", "", "Shell to target: bash, zsh, fish, powershell, pwsh, cmd")
	activateCmd.Flags().BoolVar(&activatePathOnly, "path", false, "Print only the activation script path")
	_ = activateCmd.Flags().MarkHidden("path")
	activateCmd.Flags().BoolVar(&activateCurrent, "current", false, "Respect a temporary shell selection")
	_ = activateCmd.Flags().MarkHidden("current")
}
