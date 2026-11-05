#!/usr/bin/env python3
"""Build and install rpy in ~/.local/bin."""

import os
from pathlib import Path
import shlex
import shutil
import subprocess
import sys
import tempfile


def main():
    repo = Path(__file__).resolve().parent
    go = shutil.which("go")
    if go is None:
        print("Go is required. Install the version specified in go.mod or newer.", file=sys.stderr)
        return 1

    bin_dir = Path.home() / ".local" / "bin"
    target = bin_dir / "rpy"
    try:
        bin_dir.mkdir(parents=True, exist_ok=True)
        # Build beside the destination so a failed build leaves the old binary intact.
        with tempfile.TemporaryDirectory(prefix=".rpy-build-", dir=bin_dir) as temp:
            binary = Path(temp) / "rpy"
            subprocess.run([go, "build", "-o", str(binary), "."], cwd=repo, check=True)
            os.replace(binary, target)
    except (OSError, subprocess.CalledProcessError) as exc:
        print(f"Could not install rpy: {exc}", file=sys.stderr)
        return 1

    print(f"Installed {target}")
    found = shutil.which("rpy")
    if found is None or Path(found).resolve() != target.resolve():
        if found:
            print(f"Your PATH currently finds {found} first.")
        shell = Path(os.environ.get("SHELL", "")).name
        path = shlex.quote(str(bin_dir))
        if shell == "fish":
            print(f"Run: fish_add_path --move {path}")
        else:
            print(f'Add this to your shell startup file: export PATH={path}:"$PATH"')
    else:
        print("Your PATH resolves to the installed binary.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
