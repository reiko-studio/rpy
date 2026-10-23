# rpy — Reiko's Python

`rpy` is a Python runtime and environment manager written in Go. It builds
CPython from source, manages local and shared virtual environments, and runs
commands and `pip` in the environment selected for each project.

Part of **Reiko Studio**.

It brings together source builds like [pyenv](https://github.com/pyenv/pyenv),
project-local environments like [uv](https://docs.astral.sh/uv/), and named
environments shared across projects like
[Anaconda](https://www.anaconda.com/docs/getting-started/working-with-conda/environments).
It uses ordinary Python virtual environments and `pip`, with shell integration
for activation.

It was built around the way its creator, [leogabac](https://github.com/leogabac),
uses Python for research and day-to-day projects.
Some projects need their own dependencies; others use the same scientific
Python setup across several repositories. Keeping a separate environment for
each of those repositories means installing and maintaining the same packages
several times.

With `rpy`, you can choose either approach for each project. The workflow stays
the same: choose a Python interpreter, create an environment, and use it to
install packages and run your code.

## Features

- Managed Python runtimes: `~/.rpy/pythons/<version>`
- Local project environments: `./.venv`
- Shared named environments: `~/.rpy/envs/<name>`
- Shell activation and automatic activation when entering a project

Each project uses one environment at a time. By default, that is the `.venv`
in the project directory. When you select a shared environment, `rpy` records
its name in a small `.rpy-env` file, so you do not need to choose it again each
time you return to the project.

A shared environment is useful when several projects need the same packages.
Changes to it are visible to all projects that use it, so a local `.venv` is a
better fit when dependencies need to evolve independently.

## Requirements

This tool is deliberately Linux-first.

Building `rpy` requires Go 1.26.2 or newer. Managed Python installs are built
from upstream CPython source, so you also need the system build dependencies
for the Python version you want to install.

**For Ubuntu:**
```sh
sudo apt update; sudo apt install make build-essential libssl-dev zlib1g-dev \
libbz2-dev libreadline-dev libsqlite3-dev curl git \
libncursesw5-dev xz-utils tk-dev libxml2-dev libxmlsec1-dev libffi-dev liblzma-dev
```

**For Fedora 22 and above:**
```sh
dnf install make gcc patch zlib-devel bzip2 bzip2-devel readline-devel sqlite sqlite-devel openssl-devel tk-devel libffi-devel xz-devel libuuid-devel gdbm-libs libnsl2
```

**For Arch:**
```sh
sudo pacman -S --needed base-devel openssl zlib xz tk zstd
```

## Install

Build from the repository root and put the binary on your `PATH`:

```sh
go build -o rpy .
mkdir -p ~/.local/bin
install -m 0755 rpy ~/.local/bin/rpy
```

Make sure `~/.local/bin` is on your `PATH`, so you can call `rpy` from any
project directory. The Go compiler is needed to build the binary; the resulting
executable can then be used directly.

## A First Environment

For a first project, you can use a Python interpreter that is already installed.
From the project directory, create a local environment:

```sh
rpy env create --python python3
```

This creates `./.venv` using `python3` from your `PATH`. You can now install a
package and run Python inside that environment:

```sh
rpy pip install numpy
rpy run python -c "import numpy; print(numpy.__version__)"
```

Both `rpy pip` and `rpy run` use the environment selected for the project,
so these commands work without activating it in your shell.

### Sharing an environment

Suppose several research projects use the same NumPy, pandas, and Matplotlib
setup. You can create a named environment once, then select it for each project:

```sh
rpy env create research --python python3
rpy env use research
rpy pip install numpy pandas matplotlib
```

In another project directory, select the same environment:

```sh
cd ../other-project
rpy env use research
rpy run python -c "import numpy; print(numpy.__version__)"
```

Both projects now use the packages installed in `research`. Each selection is
saved in that project's `.rpy-env` file. To return a project to its own `.venv`,
run `rpy env use local`; create the local environment first if it does not exist.

To keep the selection file out of version control, add it to `.gitignore`:

```sh
printf '\n.rpy-env\n' >> .gitignore
```

## Shell Integration

If you prefer to call `python` and other installed commands directly, shell
integration handles activation for you. Add the appropriate line to your shell
startup file, then reload it or open a new shell:

```sh
# ~/.bashrc
eval "$(rpy shell-init --shell bash)"
```

```sh
# ~/.zshrc
eval "$(rpy shell-init --shell zsh)"
```

```fish
# ~/.config/fish/config.fish
rpy shell-init --shell fish | source
```

With integration loaded, `rpy env activate` updates the current shell to use
the selected environment. Entering a project directory also activates its
environment automatically.

```sh
rpy env activate
rpy env use research --activate
```

`rpy env use <name>` saves the project selection. Adding `--activate` instead
makes a temporary selection for the current shell without writing `.rpy-env`.

Without shell integration, `rpy env activate` prints an activation command.
For a one-off activation in bash or zsh, evaluate that output:

```sh
eval "$(rpy env activate)"
```

## Python Runtimes

Install a managed runtime from upstream CPython source:

```sh
rpy py install 3.12.3
rpy env create --python 3.12.3
```

Installation downloads the source from `python.org`, then runs `./configure`,
`make`, and `make install`. The installed interpreter lives under
`~/.rpy/pythons/3.12.3`. Building Python can take a while and requires the system
libraries listed above.

You do not need a managed runtime to use `rpy`. When creating an environment,
`--python` also accepts an interpreter on `PATH` or an explicit path:

```sh
rpy env create --python python3
rpy env create --python /usr/bin/python3
```

List, locate, or remove runtimes:

```sh
rpy py list
rpy py which 3.12.3
rpy py which python3
rpy py remove 3.12.3
```

## Manage Environments

Create a shared environment and select it in one step:

```sh
rpy env create research --python 3.12.3 --use
```

If you need to recreate an environment, use `--force`. This replaces the
existing environment, including its installed packages:

```sh
rpy env create --python 3.12.3 --force
rpy env create research --python 3.12.3 --force
```

To check which environment a project is using, `rpy env info` shows its
metadata and `rpy env path` prints its directory. `rpy env list` shows the local
and shared environments available to the project:

```sh
rpy env info
rpy env path
rpy env list
```

When an environment is no longer needed, remove it by name. Use `local` for
the current project's `.venv`:

```sh
rpy env remove local
rpy env remove research
```

## Run Commands and Manage Packages

`rpy run` puts the selected environment on `PATH` for a command. This is useful
for scripts and tools you want to run without changing your current shell:

```sh
rpy run python --version
rpy run pytest
```

Package installation goes through the environment's Python interpreter:

```sh
rpy pip install -r requirements.txt
rpy pip list
rpy pip freeze
```

`rpy pip` runs `pip` through the selected environment's interpreter.
`rpy freeze` is a shortcut for `rpy pip freeze`.

## Diagnostics

If something is not working as expected, start with the built-in checks:

```sh
rpy doctor
```

This checks storage, Python availability, shell integration, the selected
environment, and source-build tools (`make` and `cc`). It can help identify a
missing environment or an incomplete local setup.

## Current Limitations

`rpy` currently focuses on Python runtimes and environment selection. Packages
are managed through `pip`; there is no dependency sync workflow, lockfile, or
environment spec format yet.

- Linux-first, with Python built from source.
- System build dependencies must be installed separately.
- Removing a shared environment does not check other projects for references.

[MIT license](LICENSE)
