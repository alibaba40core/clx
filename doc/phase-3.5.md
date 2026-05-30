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
```
