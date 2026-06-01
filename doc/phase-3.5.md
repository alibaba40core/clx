# Phase 3.5 — Aliases

> **Status:** Complete.
>
> Tracker for global aliases in `~/.clx/aliases.yaml`. See `doc/architecture.md` §3.16.

---

## Status snapshot

| Sub-phase | Scope | Status | Notes |
|-----------|-------|--------|-------|
| **3.5.1** | Alias store + config paths | ✅ Done | `internal/aliases` |
| **3.5.2** | Parser expansion | ✅ Done | Single-level, before classify |
| **3.5.3** | CLI + collision warnings + e2e | ✅ Done | `clx alias` |

---

## Tasks

- [x] Alias store: load/set/remove/list, atomic writes, max_aliases
- [x] `config.AliasesPath`, bootstrap template
- [x] Parser expands first token before classification
- [x] `clx alias set/list/rm` with collision warning at set time
- [x] Security e2e: alias value still passes risk/policy

---

## Update log

```
2026-05-30 · 3.5.1 · aliases: add persistent store and config paths for ~/.clx/aliases.yaml · 3ca1a77
2026-05-30 · 3.5.2 · parser: expand first-token aliases before classification · 1057f68
2026-05-30 · 3.5.3 · cmd: add clx alias subcommand with collision warnings and safety e2e · 0362e4d
```
