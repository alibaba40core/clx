# CLX

**Write commands once. Run anywhere.**

CLX is an AI-powered cross-platform command intelligence layer for developers. It understands intent, translates commands between shell environments, and executes them safely across Windows, Linux, macOS, PowerShell, CMD, Bash, Zsh, and WSL.

## Status

> **Phase 2.3 — OpenAI + intent cache** — Rules-miss natural-language inputs resolve through Ollama (default `qwen3:1.7b`) or OpenAI when configured, with optional `providers.fallback` chain (primary down → fallback on infrastructure errors only). Repeat inputs hit a file-backed intent cache at `~/.clx/cache/intents.json`. Schema-constrained JSON output, `Engine.ValidateResolved`, and the full safety pipeline unchanged. Flags: `--provider` (`ollama` / `openai`; `azure` stub), `--explain`, `--dry-run`, `-y`. Includes Phase 1.1–2.3.

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
internal/builtin/ Embedded built-in rules and skills (YAML, shipped in binary)
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

```bash
# Translate and run (prompts [Y/n] unless -y)
clx grep errors logs.txt

# Preview (default: dry-run from config; --dry-run forces preview)
clx --explain grep errors logs.txt
clx pwd
clx --dry-run pwd

# Execute on Windows (after safety.dry_run: false in ~/.clx/config.yaml)
clx -y pwd
```

Host scripts are assembled only from rule templates with validated parameters; CLX does not pass your raw shell input to `powershell -Command` or `cmd /c`.

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
├── rules/          # optional user rule overrides
├── skills/         # optional user skill pack overrides
└── logs/
```

See [configs/config.example.yaml](configs/config.example.yaml) for the default config template.

## Customizing rules

Built-in rules and skills ship **inside the `clx` binary** (from `internal/builtin/`), so `clx` works from any directory without the source tree on disk.

To override or extend behavior, add YAML under:

- `~/.clx/rules/*.yaml` — same format as built-in rule files (`intent:` or `intents:` list)
- `~/.clx/skills/<pack>/intents.yaml` — same format as built-in skill packs

If a user rule uses the same `intent` name as a built-in rule, **your definition wins**. Malformed overlay files are skipped with a warning; built-ins still load.

## License

MIT — see [LICENSE](LICENSE).
