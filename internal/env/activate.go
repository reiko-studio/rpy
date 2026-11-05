package env

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func PrintActivation(name, shell string, pathOnly, current bool) error {
	temporary := name != ""
	var info Info
	var err error
	switch {
	case name != "":
		info, name, err = selectionActivationInfo(name)
	case current:
		info, err = requireCurrentEnv()
	default:
		name, err = registeredEnvName()
		if err == nil {
			if name == "" {
				name = "local"
			}
			info, _, err = selectionActivationInfo(name)
		}
	}
	if err != nil {
		return err
	}
	if pathOnly {
		path, err := activationScriptPathForSelection(info, shell)
		if err != nil {
			return err
		}
		fmt.Println(path)
		return nil
	}
	command, err := selectionSnippet(info, name, shell, temporary)
	if err == nil {
		fmt.Println(command)
	}
	return err
}

func ActivateCommand(info Info, shell string) (string, error) {
	normalized := normalizeShell(shell)
	script := activationScriptPath(info, normalized)

	if _, err := os.Stat(script); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("activation script not found at %s", script)
		}
		return "", err
	}

	switch normalized {
	case "fish":
		return "source " + quoteShell(script, normalized), nil
	case "powershell", "pwsh":
		return fmt.Sprintf("& %q", script), nil
	case "cmd":
		return script, nil
	default:
		return "source " + quoteShell(script, normalized), nil
	}
}

func activationScriptPath(info Info, shell string) string {
	if runtime.GOOS == "windows" {
		switch shell {
		case "powershell", "pwsh":
			return filepath.Join(info.BinDir, "Activate.ps1")
		case "cmd":
			return filepath.Join(info.BinDir, "activate.bat")
		default:
			return filepath.Join(info.BinDir, "activate")
		}
	}

	switch shell {
	case "fish":
		return filepath.Join(info.BinDir, "activate.fish")
	case "csh":
		return filepath.Join(info.BinDir, "activate.csh")
	default:
		return filepath.Join(info.BinDir, "activate")
	}
}

func normalizeShell(shell string) string {
	if shell == "" {
		shell = filepath.Base(os.Getenv("SHELL"))
	}

	shell = strings.ToLower(shell)
	switch shell {
	case "", "bash", "sh":
		return "sh"
	case "zsh":
		return "zsh"
	case "fish":
		return "fish"
	case "powershell", "pwsh":
		return "powershell"
	case "cmd", "cmd.exe":
		return "cmd"
	case "csh", "tcsh":
		return "csh"
	default:
		return shell
	}
}
