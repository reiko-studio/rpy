package cmd

import (
	"fmt"
	"strings"

	"rpy/internal/tool"

	"github.com/spf13/cobra"
)

func init() {
	group := &cobra.Command{Use: "tool", Short: "Manage Python CLI tools in isolated environments"}
	var interpreter string
	var additional []string
	var extras []string
	install := &cobra.Command{
		Use: "install <package|wheel>", Short: "Install a Python CLI package",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			item, err := tool.Install(args[0], interpreter, extras, cmd.ErrOrStderr(), additional...)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Installed %s %s (%s) in ~/.local/bin\n", item.Name, item.Version, strings.Join(item.Commands, ", "))
			return nil
		},
	}
	install.Flags().StringArrayVar(&extras, "extra", nil, "Optional extra declared by the tool package (repeatable)")
	install.Flags().StringArrayVar(&additional, "with", nil, "Additional package requirement or wheel to install (repeatable)")
	install.Flags().StringVar(&interpreter, "python", "", "Python version, executable name, or path")
	group.AddCommand(install, &cobra.Command{
		Use: "list", Short: "List installed tool packages and their commands", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			items, err := tool.List()
			if err != nil {
				return err
			}
			for _, item := range items {
				fmt.Fprintf(cmd.OutOrStdout(), "%s %s (%s)\n", item.Name, item.Version, strings.Join(item.Commands, ", "))
			}
			return nil
		},
	}, &cobra.Command{
		Use: "uninstall <package>", Short: "Remove an installed tool by package name", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := tool.Uninstall(args[0]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Uninstalled %s\n", args[0])
			return nil
		},
	})
	rootCmd.AddCommand(group)
}
