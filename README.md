# CLX

**Write commands once. Run anywhere.**

CLX is an AI-powered cross-platform command intelligence layer for developers. It understands intent, translates commands between shell environments, and executes them safely across Windows, Linux, macOS, PowerShell, CMD, Bash, Zsh, and WSL.

## Status

> **v1.0.0 (shipped)** — Phases 1–4 + 3.5: rules-first pipeline with AI fallback (Ollama, OpenAI, Gemini), intent cache, risk/policy gates, safety presets, aliases, and `clx init`. Manage settings with `clx config` / `clx safety` / `clx policy` / `clx cache` / `clx alias` (API keys encrypted at rest). Provider HTTP 429 surfaces as a clear rate-limit message. Common NL phrases (e.g. `what is my ip`) hit built-in rules before AI. Flags: `--provider`, `--explain`, `--dry-run`, `-y`. See [doc/architecture.md](doc/architecture.md), [doc/provider-config.md](doc/provider-config.md), and [CHANGELOG.md](CHANGELOG.md).

**CLI subcommands:** `doctor`, `init`, `config`, `safety`, `policy`, `alias`, `cache` — run `clx <cmd> help` for each.

**`clxmax`:** Version 2 / Phase 5 (in development) — advanced reasoning binary; stub today (`clxmax --version` only). See [doc/phase-5.md](doc/phase-5.md).

## What CLX does

```bash
# On PowerShell — CLX translates and runs the native equivalent
clx grep errors logs.txt
# → Select-String "errors" logs.txt

# Natural language works too
clx find all files modified today
# → Get-ChildItem ... (Windows) or find ... (Linux)
```

**How it resolves input (v1):**

1. **Rules first** — built-in YAML intents (filesystem, git, docker, networking, …) match most common phrases with no LLM call.
2. **Cache** — repeats of prior AI resolutions are stored under `~/.clx/cache/`.
3. **AI intent** — when rules miss, the configured provider maps natural language to a known intent + params.
4. **AI command generation** — when intent still misses (`features.ai_command_generation`, on by default), the provider may return a structured `argv`; it still passes validation, risk, policy, and confirmation before exec.

**Session memory** — follow-up commands in the same shell session can use recent context (`memory.enabled` in config). Session data is bounded and stored under `~/.clx/sessions/`; it is not a long-term knowledge base.

**Aliases** — shortcuts in `~/.clx/aliases.yaml` expand before intent resolution and go through the same safety gates as any other input:

```bash
clx alias set gst "git status"
clx gst
```

## What CLX is not (v1)

- Not an autonomous coding agent or multi-step planner (`clxmax` will add planning in v2).
- Not a full repo or IDE assistant.
- Not a plugin marketplace (yet).

Focus: **trustable cross-platform command abstraction** — fast, reliable, safe, portable.

## Requirements

| Need | Notes |
|------|--------|
| **Build from source** | Go **1.26+** (`go.mod`). There is no `brew` / `winget` / curl installer yet — use `make build` or `make install`. |
| **Local AI (optional)** | Default provider is **Ollama** (`http://localhost:11434`). Install [Ollama](https://ollama.com) and pull a model (e.g. `qwen3:1.7b`) for offline use. |
| **Cloud AI (optional)** | OpenAI or Gemini API keys via `clx config set providers.openai.api_key` (see [doc/provider-config.md](doc/provider-config.md)). |
| **Azure** | Listed in config/help but **not implemented** in v1 — use `ollama`, `openai`, or `gemini`. |

## Getting started

After building or installing:

```bash
make build

# First run creates ~/.clx/; wizard configures provider and safety
clx init

# Detect OS, shell, and tools → ~/.clx/system_profile.json
clx doctor

# Preview a translation (no execution)
clx --explain grep errors logs.txt

# Run with confirmation (or -y to skip confirm when policy allows)
clx grep errors logs.txt
```

On Windows, prefer `bin\clx.exe` or ensure `bin\` is on PATH after `make build`. See [Install and avoiding a stale binary](#install-and-avoiding-a-stale-binary) below.

**AI providers:** default is Ollama. Switch with `clx config provider use openai` (or `gemini`) and set API keys as in [doc/provider-config.md](doc/provider-config.md). Override per run: `clx --provider openai …`.

## Architecture

See [doc/architecture.md](doc/architecture.md) for the full V1 architecture, component contracts, and implementation phases.

## Project structure

```
cmd/clx/          CLI entrypoint (v1.0.0)
cmd/clxmax/       Version 2 reasoning binary (stub in v1)
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
./bin/clx --version          # Unix
# bin\clx.exe --version      # Windows

# Interactive setup + environment profile
clx init
clx doctor

# Dev install (build + copy to PATH)
make install

# Local bootstrap without touching real ~/.clx/
make bootstrap-local
```

### Install and avoiding a stale binary

- **Install today:** from source only (`make build` / `make install`). Packaged installs (brew, winget) are not shipped yet.
- **Preferred:** `make build` then run `./bin/clx` (or `make install` to copy into your user PATH).
- **Stale binary trap:** A `clx.exe` in the **repo root** (from an old `go build` in `.`) can appear **before** `bin/` on PATH when your shell cwd is the repo — you may run weeks-old code while `bin/clx` is current. Use `where clx` (Windows) or `which clx` (Unix) to see what runs; remove stray root copies or run `make clean` (deletes root `clx.exe` / `clxmax.exe` and `bin/`).
- **Check version:** `clx --version` should report **1.0.0** on a fresh `make build` (unless overridden). Rebuild after pulling: `go build -o bin/clx.exe ./cmd/clx` on Windows.
- **Clean artifacts:** `make clean` removes `bin/` and any `clx.exe` / `clxmax.exe` in the repo root (see `Makefile` `clean` target).
- Root `*.exe` files are gitignored but still execute if present on disk — delete them when troubleshooting version mismatches.

```bash
# Translate and run (prompts [Y/n] unless -y; behavior depends on safety.mode)
clx grep errors logs.txt

# Preview without executing
clx --explain grep errors logs.txt
clx pwd
clx --dry-run pwd

# Execute when policy and safety mode allow (often after lowering dry_run or using -y)
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
├── aliases.yaml              # user-global shortcuts (clx alias)
├── system_profile.json
├── cache/
│   ├── intents.json          # AI intent cache
│   ├── explanations.json
│   └── commands.json         # AI command-generation cache
├── sessions/                 # session-scoped memory (bounded)
├── policies/
│   └── policy.yaml           # block/allow lists, access_level
├── rules/                    # optional user rule overrides
├── skills/                   # optional user skill pack overrides
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
