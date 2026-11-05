package env

import (
	"fmt"
	"os"
	"strings"
)

func PrintUseAndActivateCommand(selection, shell string) error {
	command, err := SelectionActivationCommand(selection, shell, true)
	if err == nil {
		fmt.Println(command)
	}
	return err
}

func SelectionActivationCommand(selection, shell string, temporary bool) (string, error) {
	info, name, err := selectionActivationInfo(selection)
	if err != nil {
		return "", err
	}
	return selectionSnippet(info, name, shell, temporary)
}

func selectionSnippet(info Info, name, shell string, temporary bool) (string, error) {
	activation, err := ActivateCommand(info, shell)
	if err != nil {
		return "", err
	}
	normalized := normalizeShell(shell)
	if normalized != "fish" && normalized != "sh" && normalized != "zsh" {
		return activation, nil
	}
	project := ""
	if temporary {
		project, err = ProjectRoot()
		if err != nil {
			return "", err
		}
	}
	if normalized == "fish" {
		state := "set -e RPY_ENV_SELECTION; set -e RPY_ENV_PROJECT"
		if temporary {
			state = fmt.Sprintf("set -gx RPY_ENV_SELECTION %s; set -gx RPY_ENV_PROJECT %s",
				quoteShell(name, normalized), quoteShell(project, normalized))
		}
		return fmt.Sprintf("%s; and begin; %s; set -gx RPY_AUTO_VENV %s; true; end",
			activation, state, quoteShell(info.Root, normalized)), nil
	}
	state := "unset RPY_ENV_SELECTION RPY_ENV_PROJECT"
	if temporary {
		state = fmt.Sprintf("export RPY_ENV_SELECTION=%s RPY_ENV_PROJECT=%s",
			quoteShell(name, normalized), quoteShell(project, normalized))
	}
	return fmt.Sprintf("%s && { %s; export RPY_AUTO_VENV=%s; }",
		activation, state, quoteShell(info.Root, normalized)), nil
}

func quoteShell(value, shell string) string {
	if shell == "fish" {
		return "'" + strings.NewReplacer("\\", "\\\\", "'", "\\'").Replace(value) + "'"
	}
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func selectionActivationInfo(selection string) (Info, string, error) {
	if selection == "local" {
		info, err := ProjectEnvInfo()
		if err != nil {
			return Info{}, "", err
		}
		if !info.Exists {
			return Info{}, "", fmt.Errorf("local environment does not exist at %s", info.Root)
		}
		return info, "local", nil
	}

	info, err := SharedEnvInfo(selection)
	if err != nil {
		return Info{}, "", err
	}
	if !info.Exists {
		return Info{}, "", fmt.Errorf("shared environment %q does not exist at %s", selection, info.Root)
	}
	return info, info.Name, nil
}

func activationScriptPathForSelection(info Info, shell string) (string, error) {
	path := activationScriptPath(info, normalizeShell(shell))
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("activation script not found at %s", path)
		}
		return "", err
	}
	return path, nil
}
