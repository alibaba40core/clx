# CLX

**Write commands once. Run anywhere.**

CLX is an AI-powered cross-platform command intelligence layer for developers. It understands intent, translates commands between shell environments, and executes them safely across Windows, Linux, macOS, PowerShell, CMD, Bash, Zsh, and WSL.

## Status

> **Phase 1.1** — Foundation & bootstrap: `clx --version`, config loader, structured logging, and first-run `~/.clx/` setup.

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

# Dev install (build + copy to PATH)
make install

# Local bootstrap without touching real ~/.clx/
make bootstrap-local
```

Command translation (`clx grep errors logs.txt`) lands in Phase 1.6.

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
