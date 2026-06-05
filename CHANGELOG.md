# Changelog

All notable changes to CLX are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## Versioning

- **v1.0.0** — Phases 1, 2, 3, 3.5, and 4 (`clx` binary).
- **v2.0.0** — Phase 5 (`clxmax` advanced reasoning); in development.

## [Unreleased] — 2.0.0

### Planned

- `clxmax` advanced reasoning: clarification loop, multi-step planning, sequenced execution, plan explainability.
- See [doc/phase-5.md](doc/phase-5.md) for the full Phase 5 tracker.

## [1.0.1] — 2026-06-05

Patch release: installers, rules-only mode, and bare `ls`/`dir` fix.

### Added

- Download-based installers (`scripts/get.sh`, `scripts/get.ps1`) and GitHub release pipeline (`.github/workflows/release.yml`); install guide in [doc/installation.md](doc/installation.md).
- First-class `none` (rules-only) provider so Ollama/AI is never mandatory; `clx config provider use none` and `--provider none`; init wizard "skip (rules only)" sets `provider: none`.

### Fixed

- Bare `clx ls` / `clx dir` resolve via built-in rules instead of falling through to AI and failing with "provider unavailable" when no provider is configured.

## [1.0.0] — 2026-05-31

First stable release. Ships through Phase 4 plus Phase 3.5 aliases and V1 polish.

### Added

- **Phase 1 — Core engine:** rules-first pipeline, environment detection (`clx doctor`), parser, intent resolver, capabilities, generator, argv-only executor with shell-native hosts.
- **Phase 2 — AI integration:** Ollama, OpenAI, and Gemini providers; AI intent fallback; intent/explanation/command caches; hybrid AI command generation (`features.ai_command_generation`).
- **Phase 3 — Safety:** risk engine, policy engine (block/allow lists, access levels), dry-run, safety mode matrix (`clx safety`), confirmations.
- **Phase 3.5 — Aliases:** persistent user-global aliases (`clx alias set/list/rm`), parser-stage expansion through full safety gates.
- **Phase 4 — Advanced UX:** session memory and follow-ups, opt-in explain-only shell hooks, interactive `clx init` wizard, per-skill AI prompt templates.
- **V1 polish:** `clx config` / `clx policy` / `clx cache` subcommands, NL rule examples, provider HTTP 429 rate-limit messages, encrypted API keys at rest.

### CLI surface

- Subcommands: `doctor`, `init`, `config`, `safety`, `policy`, `alias`, `cache`.
- Flags: `--explain`, `--dry-run`, `-y` / `--yes`, `--provider`, `--config`.
