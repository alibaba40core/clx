# Installing CLX

CLX ships two binaries — `clx` (v1) and `clxmax` (v2 stub). There are two ways to
install them:

1. **Build from source** — the **primary, recommended** method. Guarantees the
   binary matches the exact commit you have checked out and works on any platform
   with a Go toolchain.
2. **Download prebuilt binaries** — a **secondary** convenience path for users who
   just want a working `clx` without installing Go.

There is no `brew` / `winget` / `scoop` package yet.

---

## 1. Build from source (primary)

**Requirements:** Go **1.26+** (see `go.mod`), `git`, and `make`.

```bash
git clone https://github.com/alibaba40core/clx.git
cd clx

# Build + copy clx, clx-ai (internal), and clxmax into your user PATH
make install
```

`make install` runs the OS-appropriate dev installer:

- **macOS / Linux** → `scripts/install.sh` (installs to `/usr/local/bin` if
  writable, otherwise `~/.local/bin`).
- **Windows** → `scripts/install.ps1` (installs to
  `%LOCALAPPDATA%\Programs\clx` and adds it to your user PATH).

Both stamp the version/commit via `-ldflags`, remove stale repo-root binaries that
can shadow PATH, and run `clx --version` + `clx doctor` as a smoke test.

Prefer not to install globally? Just build and run from `bin/`:

```bash
make build
./bin/clx --version          # Unix
# bin\clx.exe --version      # Windows
```

> **Stale-binary trap:** a `clx.exe` left in the **repo root** from an old
> `go build .` can appear *before* `bin/` on PATH when your shell cwd is the repo.
> Use `where clx` (Windows) / `which clx` (Unix) to see what runs, and `make clean`
> to remove stray copies. See the README "Install and avoiding a stale binary".

---

## 2. Download prebuilt binaries (secondary)

These one-liners download `clx`, `clx-ai` (internal AI worker), and `clxmax` from the latest
[GitHub Release](https://github.com/alibaba40core/clx/releases) (currently **v1.0.2**), **verify the
SHA-256 checksum**, install them onto your PATH, and print the version. No Go
toolchain or source checkout required.

> **Run them in the right shell.** They are **not** interchangeable:
> - The `irm … | iex` line is **PowerShell only** (Windows PowerShell 5.1+ or
>   PowerShell 7). It will **not** work in Command Prompt (`cmd.exe`).
> - The `curl … | bash` line is for **bash** on macOS/Linux (or WSL / Git Bash on
>   Windows). It will **not** work in PowerShell or CMD.

### Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/alibaba40core/clx/main/scripts/get.ps1 | iex
```

### macOS / Linux (bash)

```bash
curl -fsSL https://raw.githubusercontent.com/alibaba40core/clx/main/scripts/get.sh | bash
```

### Options

Pin a specific release or change the install directory with environment variables:

```powershell
# Windows
$env:CLX_VERSION="v1.0.2"; $env:CLX_INSTALL_DIR="C:\tools\clx"; irm https://raw.githubusercontent.com/alibaba40core/clx/main/scripts/get.ps1 | iex
```

```bash
# macOS / Linux
CLX_VERSION=v1.0.2 CLX_INSTALL_DIR="$HOME/.local/bin" \
  curl -fsSL https://raw.githubusercontent.com/alibaba40core/clx/main/scripts/get.sh | bash
```

| Variable | Default | Purpose |
|---|---|---|
| `CLX_VERSION` | `latest` (resolves to current release, e.g. `v1.0.2`) | Release tag to install, e.g. `v1.0.2` |
| `CLX_INSTALL_DIR` | `/usr/local/bin` or `~/.local/bin` (Unix); `%LOCALAPPDATA%\Programs\clx` (Windows) | Install destination |

### How releases are produced

Prebuilt assets are published automatically by
[`.github/workflows/release.yml`](../.github/workflows/release.yml) whenever a
`v*` tag is pushed. Each release contains
`clx_<os>_<arch>.tar.gz` / `clx_windows_<arch>.zip` for
linux/darwin/windows × amd64/arm64 (includes `clx`, internal `clx-ai`, and `clxmax`), plus `checksums.txt`. If no release exists for
your platform, fall back to **Build from source**.

---

## Verify your install

```bash
clx --version     # should report the release/build version
clx doctor        # detects OS, shell, and tools → ~/.clx/system_profile.json
clx init          # first-run wizard: provider + safety setup
```

If `clx` isn't found after install, make sure the install directory is on your
PATH (restart your shell on Windows after the PowerShell installer adds it).
