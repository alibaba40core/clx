# CLX

**Write commands once. Run anywhere.**

CLX is an AI-powered cross-platform command intelligence layer for developers. It understands intent, translates commands between shell environments, and executes them safely across Windows, Linux, macOS, PowerShell, CMD, Bash, Zsh, and WSL.

## Status

> **V1 polish (Phases 1–4 + 3.5)** — Rules-first pipeline with AI fallback (Ollama, OpenAI, Gemini), intent cache, risk/policy gates, safety presets, aliases, and `clx init`. Manage settings with `clx config` / `clx safety` / `clx policy` / `clx cache` / `clx alias` (API keys encrypted at rest). Provider HTTP 429 surfaces as a clear rate-limit message. Common NL phrases (e.g. `what is my ip`) hit built-in rules before AI. Flags: `--provider`, `--explain`, `--dry-run`, `-y`. See [doc/architecture.md](doc/architecture.md) and [doc/provider-config.md](doc/provider-config.md).

**CLI subcommands:** `doctor`, `init`, `config`, `safety`, `policy`, `alias`, `cache` — run `clx <cmd> help` for each.

**`clxmax`:** planned advanced reasoning binary; not shipped in V1 (`clx --version` only today).

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
cmd/clxmax/       Advanced reasoning mode (future; stub in tree)
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

### Install and avoiding a stale binary

- **Preferred:** `make build` then run `./bin/clx` (or `make install` to copy into your user PATH).
- **Stale binary trap:** A `clx.exe` in the **repo root** (from an old `go build` in `.`) can appear **before** `bin/` on PATH when your shell cwd is the repo — you may run weeks-old code while `bin/clx` is current. Use `where clx` (Windows) or `which clx` (Unix) to see what runs; remove stray root copies or run `make clean` (deletes root `clx.exe` / `clxmax.exe` and `bin/`).
- **Check version:** `clx --version` should match your latest build; rebuild after pulling: `go build -o bin/clx.exe ./cmd/clx` on Windows.

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

### Shell integration (explain-only)

`execution.shell_integration: true` in config only enables a **hint** when rules miss; it does **not** intercept or auto-run commands in your shell. Optional hooks ([`scripts/clx-hook.ps1`](scripts/clx-hook.ps1), [`scripts/clx-hook.sh`](scripts/clx-hook.sh)) forward input to `clx --explain` so you review the translation before running anything yourself. Auto-execution from hooks is intentionally out of scope (safe-command-execution contract). Install snippets: `clx init` or `clx config` + embedded instructions.

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

Manage settings from the CLI: `clx config` (providers, features, cache, memory, execution, logging), `clx safety`, `clx policy`, `clx cache status|clear`, `clx alias`. API keys are encrypted at rest in `config.yaml`. See [doc/provider-config.md](doc/provider-config.md).

## Customizing rules

Built-in rules and skills ship **inside the `clx` binary** (from `internal/builtin/`), so `clx` works from any directory without the source tree on disk.

To override or extend behavior, add YAML under:

- `~/.clx/rules/*.yaml` — same format as built-in rule files (`intent:` or `intents:` list)
- `~/.clx/skills/<pack>/intents.yaml` — same format as built-in skill packs

If a user rule uses the same `intent` name as a built-in rule, **your definition wins**. Malformed overlay files are skipped with a warning; built-ins still load.

## License

MIT — see [LICENSE](LICENSE).
