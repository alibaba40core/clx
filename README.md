# CLX

**Write commands once. Run anywhere.**

CLX is an AI-powered cross-platform command intelligence layer for developers. It understands intent, translates commands between shell environments, and executes them safely across Windows, Linux, macOS, PowerShell, CMD, Bash, Zsh, and WSL.

## Status

> **Phase 1.2** — Environment detection: `clx doctor` writes `~/.clx/system_profile.json` (OS, shell, tools, package managers). Phase 1.1 foundation (`--version`, config, logging) is included.

## What CLX does

```bash
# On PowerShell — CLX translates and runs the native equivalent
clx grep errors logs.txt
# → Select-String "errors" logs.txt

# Natural language works too
clx find all files modified today
# → Get-ChildItem ... (Windows) or find ... (Linux)
```

## Architecture

See [doc/architecture.md](doc/architecture.md) for the full V1 architecture, component contracts, and implementation phases.

## Project structure

```
cmd/clx/          CLI entrypoint
cmd/clxmax/       Advanced reasoning mode
internal/         Engine packages (parser, intent, environment, ...)
rules/            Built-in intent rules (YAML)
skills/           Domain skill packs (git, docker, k8s, ...)
configs/          Example config templates
policies/         Default policy templates
test/             Integration and e2e tests
doc/              Design and architecture docs
```

## Quickstart

```bash
# Build
make build

# Print version (first run also creates ~/.clx/)
./bin/clx --version

# Detect environment → ~/.clx/system_profile.json
./bin/clx doctor

# Dev install (build + copy to PATH)
make install

# Local bootstrap without touching real ~/.clx/
make bootstrap-local
```

Command translation (`clx grep errors logs.txt`) lands in Phase 1.6.

`shell_version` in the profile is filled from environment variables when present (e.g. `POWERSHELL_VERSION`); otherwise it stays empty until a later phase adds subprocess probing.

## Configuration

Runtime config lives in `~/.clx/` (created on first run):

```
~/.clx/
├── config.yaml
├── system_profile.json
├── cache/
├── sessions/
├── policies/
└── logs/
```

See [configs/config.example.yaml](configs/config.example.yaml) for the default config template.

## License

MIT — see [LICENSE](LICENSE).
