// Package tool manages isolated Python CLI installations.
package tool

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"rpy/internal/python"

	"golang.org/x/sys/unix"
)

type Tool struct {
	Name     string
	Version  string
	Commands []string
	Env      string
}

var packageName = regexp.MustCompile(`^[A-Za-z0-9]([A-Za-z0-9._-]*[A-Za-z0-9])?$`)
var separators = regexp.MustCompile(`[-_.]+`)

func canonical(name string) string {
	return separators.ReplaceAllString(strings.ToLower(name), "-")
}

func storage() (root, bin string, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", err
	}
	return filepath.Join(home, ".rpy", "tools"), filepath.Join(home, ".local", "bin"), nil
}

func lock(root string) (*os.File, error) {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(filepath.Join(root, ".lock"), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	if err := unix.Flock(int(file.Fd()), unix.LOCK_EX); err != nil {
		file.Close()
		return nil, err
	}
	return file, nil
}

func Install(spec, requestedPython string, log io.Writer) (_ Tool, err error) {
	if !packageName.MatchString(spec) && !strings.HasSuffix(spec, ".whl") {
		return Tool{}, fmt.Errorf("expected a package name or local .whl file")
	}
	if strings.HasSuffix(spec, ".whl") {
		info, err := os.Stat(spec)
		if err != nil {
			return Tool{}, fmt.Errorf("local wheel: %w", err)
		}
		if !info.Mode().IsRegular() {
			return Tool{}, fmt.Errorf("wheel must be a regular file: %s", spec)
		}
		spec, err = filepath.Abs(spec)
		if err != nil {
			return Tool{}, err
		}
	}
	var interpreter python.Interpreter
	if requestedPython == "" {
		interpreter, err = python.ResolveDefault()
	} else {
		interpreter, err = python.Resolve(requestedPython)
	}
	if err != nil {
		return Tool{}, err
	}

	root, bin, err := storage()
	if err != nil {
		return Tool{}, err
	}
	guard, err := lock(root)
	if err != nil {
		return Tool{}, err
	}
	defer guard.Close()

	// Entry-point scripts embed the venv's absolute Python path. Keep the
	// environment where it was created, even after discovering its package name.
	venv, err := os.MkdirTemp(root, "venv-")
	if err != nil {
		return Tool{}, err
	}
	var linked []string
	defer func() {
		if err != nil {
			for _, name := range linked {
				err = errors.Join(err, removeLink(bin, venv, name))
			}
			err = errors.Join(err, os.RemoveAll(venv))
		}
	}()

	if err := run(log, interpreter.Path, "-m", "venv", venv); err != nil {
		return Tool{}, err
	}
	py := filepath.Join(venv, "bin", "python")
	report := filepath.Join(venv, "pip-report.json")
	if err := run(log, py, "-Im", "pip", "install", "--ignore-installed", "--report", report, "--", spec); err != nil {
		return Tool{}, err
	}
	query := exec.Command(py, "-I", "-c", inspectTool, report)
	query.Stderr = log
	output, err := query.Output()
	if err != nil {
		return Tool{}, fmt.Errorf("read installed package entry points: %w", err)
	}
	var item Tool
	if err := json.Unmarshal(output, &item); err != nil {
		return Tool{}, err
	}
	if !packageName.MatchString(item.Name) {
		return Tool{}, fmt.Errorf("invalid installed package name %q", item.Name)
	}
	if len(item.Commands) == 0 {
		return Tool{}, fmt.Errorf("package %s declares no CLI commands", item.Name)
	}
	item.Env = filepath.Base(venv)
	record := filepath.Join(root, canonical(item.Name)+".json")
	if _, err := os.Lstat(record); err == nil {
		return Tool{}, fmt.Errorf("tool %s is already installed; uninstall it before reinstalling", item.Name)
	} else if !os.IsNotExist(err) {
		return Tool{}, err
	}
	if err := os.MkdirAll(bin, 0o755); err != nil {
		return Tool{}, err
	}
	for _, name := range item.Commands {
		if !validCommand(name) {
			return Tool{}, fmt.Errorf("invalid command name %q", name)
		}
		target := filepath.Join(venv, "bin", name)
		if info, err := os.Stat(target); err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0o111 == 0 {
			return Tool{}, fmt.Errorf("package %s did not install an executable for %q", item.Name, name)
		}
		link := filepath.Join(bin, name)
		if err := os.Symlink(target, link); err != nil {
			if os.IsExist(err) {
				return Tool{}, fmt.Errorf("command %q already exists at %s", name, link)
			}
			return Tool{}, err
		}
		linked = append(linked, name)
	}
	if err := os.Remove(report); err != nil {
		return Tool{}, err
	}
	data, err := json.MarshalIndent(item, "", "  ")
	if err != nil {
		return Tool{}, err
	}
	// Publish a complete record for readers and avoid a partial JSON file.
	temp := filepath.Join(venv, "tool.json")
	if err := os.WriteFile(temp, data, 0o644); err != nil {
		return Tool{}, err
	}
	if err := os.Rename(temp, record); err != nil {
		return Tool{}, err
	}
	return item, nil
}

const inspectTool = `
import importlib.metadata as metadata
import json
import sys

with open(sys.argv[1], encoding="utf-8") as f:
    requested, = [item for item in json.load(f)["install"] if item["requested"]]
dist = metadata.distribution(requested["metadata"]["name"])
print(json.dumps({
    "Name": dist.metadata["Name"],
    "Version": dist.version,
    "Commands": sorted({ep.name for ep in dist.entry_points if ep.group in ("console_scripts", "gui_scripts")}),
}))
`

func run(log io.Writer, program string, args ...string) error {
	cmd := exec.Command(program, args...)
	cmd.Stdout, cmd.Stderr = log, log
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %s: %w", filepath.Base(program), strings.Join(args, " "), err)
	}
	return nil
}

func validCommand(name string) bool {
	return name != "" && name != "." && name != ".." && filepath.Base(name) == name
}

func readTool(path string) (Tool, error) {
	var item Tool
	data, err := os.ReadFile(path)
	if err != nil {
		return item, err
	}
	if err := json.Unmarshal(data, &item); err != nil {
		return item, fmt.Errorf("read %s: %w", path, err)
	}
	if !strings.HasPrefix(item.Env, "venv-") || filepath.Base(item.Env) != item.Env {
		return item, fmt.Errorf("invalid environment in %s", path)
	}
	for _, name := range item.Commands {
		if !validCommand(name) {
			return item, fmt.Errorf("invalid command name %q in %s", name, path)
		}
	}
	return item, nil
}

func List() ([]Tool, error) {
	root, _, err := storage()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(root)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var items []Tool
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		item, err := readTool(filepath.Join(root, entry.Name()))
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool { return canonical(items[i].Name) < canonical(items[j].Name) })
	return items, nil
}

func Uninstall(name string) error {
	if !packageName.MatchString(name) {
		return fmt.Errorf("invalid package name %q", name)
	}
	root, bin, err := storage()
	if err != nil {
		return err
	}
	guard, err := lock(root)
	if err != nil {
		return err
	}
	defer guard.Close()

	record := filepath.Join(root, canonical(name)+".json")
	item, err := readTool(record)
	if os.IsNotExist(err) {
		return fmt.Errorf("tool %q is not installed; use rpy tool list to see package names", name)
	}
	if err != nil {
		return err
	}
	venv := filepath.Join(root, item.Env)
	for _, command := range item.Commands {
		if err := removeLink(bin, venv, command); err != nil {
			return err
		}
	}
	if err := os.RemoveAll(venv); err != nil {
		return err
	}
	return os.Remove(record)
}

func removeLink(bin, venv, name string) error {
	link := filepath.Join(bin, name)
	target, err := os.Readlink(link)
	if err != nil {
		// The user may have removed or replaced the command independently.
		if info, statErr := os.Lstat(link); os.IsNotExist(statErr) || (statErr == nil && info.Mode()&os.ModeSymlink == 0) {
			return nil
		}
		return err
	}
	if target != filepath.Join(venv, "bin", name) {
		return nil
	}
	if err := os.Remove(link); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
