package env

import (
	"fmt"
	"strings"
)

func PrintShellInit(shell string) error {
	script, err := ShellInitScript(shell)
	if err != nil {
		return err
	}

	fmt.Println(script)
	return nil
}

func PrintShellInstallHint(shell string) error {
	hint, err := ShellInstallHint(shell)
	if err != nil {
		return err
	}

	fmt.Println(hint)
	return nil
}

func ShellInitScript(shell string) (string, error) {
	switch normalizeShell(shell) {
	case "sh":
		return bashShellInit(), nil
	case "zsh":
		return zshShellInit(), nil
	case "fish":
		return fishShellInit(), nil
	default:
		return "", fmt.Errorf("unsupported shell %q; supported shells are bash, zsh, and fish", shell)
	}
}

func bashShellInit() string {
	return strings.TrimSpace(`
# rpy shell integration
export RPY_SHELL_INIT=1

rpy() {
  if [ "$#" -ge 3 ] && [ "$1" = "env" ] && [ "$2" = "use" ]; then
    local _rpy_use_activate=""
    local _rpy_use_args=()
    local arg
    for arg in "${@:3}"; do
      if [ "$arg" = "--activate" ]; then
        _rpy_use_activate=1
        continue
      fi
      _rpy_use_args+=("$arg")
    done
    if [ -n "$_rpy_use_activate" ]; then
      if [ "${#_rpy_use_args[@]}" -ne 1 ]; then
        command rpy env use "${_rpy_use_args[@]}" --activate
        return $?
      fi
      unset RPY_ENV_PROJECT
      export RPY_ENV_SELECTION="${_rpy_use_args[0]}"
      local _rpy_activate_path
      _rpy_activate_path="$(command rpy env activate --shell bash --path --current)" || return $?
      if [ -n "${VIRTUAL_ENV:-}" ] && [ "${VIRTUAL_ENV}" != "$(dirname "$(dirname "$_rpy_activate_path")")" ]; then
        deactivate >/dev/null 2>&1 || true
      fi
      . "$_rpy_activate_path"
      export RPY_AUTO_VENV="$(dirname "$(dirname "$_rpy_activate_path")")"
      return $?
    fi
    command rpy env use "${_rpy_use_args[@]}" || return $?
    unset RPY_ENV_SELECTION
    _rpy_autoenv
    return $?
  fi

  if [ "$#" -ge 2 ] && [ "$1" = "env" ] && { [ "$2" = "activate" ] || [ "$2" = "register" ]; }; then
    local arg
    for arg in "${@:3}"; do
      case "$arg" in
        -h|--help|--help=*|--path|--path=*) command rpy "$@"; return $? ;;
      esac
    done
    local _rpy_snippet
    _rpy_snippet="$(command rpy "$@" --shell bash)" || return $?
    eval "$_rpy_snippet"
    return $?
  fi

  command rpy "$@"
}

_rpy_autoenv() {
  local _rpy_activate_path=""
  _rpy_activate_path="$(command rpy env activate --shell bash --path --current 2>/dev/null || true)"

  if [ -n "${RPY_AUTO_VENV:-}" ] && [ "${VIRTUAL_ENV:-}" = "$RPY_AUTO_VENV" ] && [ -z "$_rpy_activate_path" ]; then
    deactivate >/dev/null 2>&1 || true
    unset RPY_AUTO_VENV
  fi

  if [ -n "$_rpy_activate_path" ]; then
    local _rpy_activate_root
    _rpy_activate_root="$(dirname "$(dirname "$_rpy_activate_path")")"
    if [ "${VIRTUAL_ENV:-}" != "$_rpy_activate_root" ]; then
      if [ -n "${VIRTUAL_ENV:-}" ]; then
        deactivate >/dev/null 2>&1 || true
      fi
      . "$_rpy_activate_path"
      export RPY_AUTO_VENV="$_rpy_activate_root"
    fi
  fi
}

case ";${PROMPT_COMMAND:-};" in
  *";_rpy_autoenv;"*) ;;
  *) PROMPT_COMMAND="_rpy_autoenv${PROMPT_COMMAND:+;$PROMPT_COMMAND}" ;;
esac

_rpy_autoenv
`)
}

func fishShellInit() string {
	return strings.TrimSpace(`
# rpy shell integration
set -gx RPY_SHELL_INIT 1

function rpy
    if test (count $argv) -ge 3; and test "$argv[1]" = "env"; and test "$argv[2]" = "use"
        set -l use_activate 0
        set -l passthrough
        for arg in $argv[3..-1]
            if test "$arg" = "--activate"
                set use_activate 1
                continue
            end
            set passthrough $passthrough $arg
        end
        if test $use_activate -eq 1
            if test (count $passthrough) -ne 1
                command rpy env use $passthrough --activate
                return $status
            end
            set -e RPY_ENV_PROJECT
            set -gx RPY_ENV_SELECTION "$passthrough[1]"
            set -l activate_path (command rpy env activate --shell fish --path --current)
            or return $status
            set -l activate_root (dirname (dirname "$activate_path"))
            if set -q VIRTUAL_ENV; and test "$VIRTUAL_ENV" != "$activate_root"
                deactivate >/dev/null 2>/dev/null
            end
            source "$activate_path"
            set -gx RPY_AUTO_VENV "$activate_root"
            return $status
        end
        command rpy env use $passthrough
        or return $status
        set -e RPY_ENV_SELECTION
        __rpy_autoenv
        return $status
    end

    if test (count $argv) -ge 2; and test "$argv[1]" = "env"
        if contains -- "$argv[2]" activate register
            for arg in $argv[3..-1]
                switch "$arg"
                    case -h --help '--help=*' --path '--path=*'
                        command rpy $argv
                        return $status
                end
            end
            set -l snippet (command rpy $argv --shell fish)
            or return $status
            eval "$snippet"
            return $status
        end
    end

    command rpy $argv
end

function __rpy_autoenv --on-variable PWD
    set -l activate_path (command rpy env activate --shell fish --path --current 2>/dev/null)

    if set -q RPY_AUTO_VENV; and test "$VIRTUAL_ENV" = "$RPY_AUTO_VENV"; and test -z "$activate_path"
        deactivate >/dev/null 2>/dev/null
        set -e RPY_AUTO_VENV
    end

    if test -n "$activate_path"
        set -l activate_root (dirname (dirname "$activate_path"))
        if not set -q VIRTUAL_ENV; or test "$VIRTUAL_ENV" != "$activate_root"
            if set -q VIRTUAL_ENV
                deactivate >/dev/null 2>/dev/null
            end
            source "$activate_path"
            set -gx RPY_AUTO_VENV "$activate_root"
        end
    end
end

__rpy_autoenv
`)
}

func zshShellInit() string {
	return strings.TrimSpace(`
# rpy shell integration
export RPY_SHELL_INIT=1

function rpy() {
  if [ "$#" -ge 3 ] && [ "$1" = "env" ] && [ "$2" = "use" ]; then
    local _rpy_use_activate=""
    local _rpy_use_args=()
    local arg
    for arg in "${@:3}"; do
      if [ "$arg" = "--activate" ]; then
        _rpy_use_activate=1
        continue
      fi
      _rpy_use_args+=("$arg")
    done
    if [ -n "$_rpy_use_activate" ]; then
      if [ "${#_rpy_use_args[@]}" -ne 1 ]; then
        command rpy env use "${_rpy_use_args[@]}" --activate
        return $?
      fi
      unset RPY_ENV_PROJECT
      export RPY_ENV_SELECTION="${_rpy_use_args[0]}"
      local _rpy_activate_path
      _rpy_activate_path="$(command rpy env activate --shell zsh --path --current)" || return $?
      if [ -n "${VIRTUAL_ENV:-}" ] && [ "${VIRTUAL_ENV}" != "$(dirname "$(dirname "$_rpy_activate_path")")" ]; then
        deactivate >/dev/null 2>&1 || true
      fi
      . "$_rpy_activate_path"
      export RPY_AUTO_VENV="$(dirname "$(dirname "$_rpy_activate_path")")"
      return $?
    fi
    command rpy env use "${_rpy_use_args[@]}" || return $?
    unset RPY_ENV_SELECTION
    _rpy_autoenv
    return $?
  fi

  if [ "$#" -ge 2 ] && [ "$1" = "env" ] && { [ "$2" = "activate" ] || [ "$2" = "register" ]; }; then
    local arg
    for arg in "${@:3}"; do
      case "$arg" in
        -h|--help|--help=*|--path|--path=*) command rpy "$@"; return $? ;;
      esac
    done
    local _rpy_snippet
    _rpy_snippet="$(command rpy "$@" --shell zsh)" || return $?
    eval "$_rpy_snippet"
    return $?
  fi

  command rpy "$@"
}

_rpy_autoenv() {
  local _rpy_activate_path=""
  _rpy_activate_path="$(command rpy env activate --shell zsh --path --current 2>/dev/null || true)"

  if [ -n "${RPY_AUTO_VENV:-}" ] && [ "${VIRTUAL_ENV:-}" = "$RPY_AUTO_VENV" ] && [ -z "$_rpy_activate_path" ]; then
    deactivate >/dev/null 2>&1 || true
    unset RPY_AUTO_VENV
  fi

  if [ -n "$_rpy_activate_path" ]; then
    local _rpy_activate_root
    _rpy_activate_root="$(dirname "$(dirname "$_rpy_activate_path")")"
    if [ "${VIRTUAL_ENV:-}" != "$_rpy_activate_root" ]; then
      if [ -n "${VIRTUAL_ENV:-}" ]; then
        deactivate >/dev/null 2>&1 || true
      fi
      . "$_rpy_activate_path"
      export RPY_AUTO_VENV="$_rpy_activate_root"
    fi
  fi
}

autoload -Uz add-zsh-hook
if ! (( ${chpwd_functions[(I)_rpy_autoenv]} )); then
  add-zsh-hook chpwd _rpy_autoenv
fi

_rpy_autoenv
`)
}

func ShellInstallPath(shell string) (string, error) {
	switch normalizeShell(shell) {
	case "sh":
		return "~/.bashrc", nil
	case "zsh":
		return "~/.zshrc", nil
	case "fish":
		return "~/.config/fish/config.fish", nil
	default:
		return "", fmt.Errorf("unsupported shell %q; supported shells are bash, zsh, and fish", shell)
	}
}

func ShellInstallLine(shell string) (string, error) {
	switch normalizeShell(shell) {
	case "sh":
		return `eval "$(rpy shell-init --shell bash)"`, nil
	case "zsh":
		return `eval "$(rpy shell-init --shell zsh)"`, nil
	case "fish":
		return `rpy shell-init --shell fish | source`, nil
	default:
		return "", fmt.Errorf("unsupported shell %q; supported shells are bash, zsh, and fish", shell)
	}
}

func ShellInstallHint(shell string) (string, error) {
	path, err := ShellInstallPath(shell)
	if err != nil {
		return "", err
	}

	line, err := ShellInstallLine(shell)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Add this line to %s:\n\n%s", path, line), nil
}
