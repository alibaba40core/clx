# Phase 3.5 — Aliases

> **Status:** In progress.
>
> Tracker for global aliases in `~/.clx/aliases.yaml`. See `doc/architecture.md` §3.16.

---

## Status snapshot

| Sub-phase | Scope | Status | Notes |
|-----------|-------|--------|-------|
| **3.5.1** | Alias store + config paths | ⬜ Not started | `internal/aliases` |
| **3.5.2** | Parser expansion | ⬜ Not started | Single-level, before classify |
| **3.5.3** | CLI + collision warnings + e2e | ⬜ Not started | `clx alias` |

---

## Tasks

- [ ] Alias store: load/set/remove/list, atomic writes, max_aliases
- [ ] `config.AliasesPath`, bootstrap template
- [ ] Parser expands first token before classification
- [ ] `clx alias set/list/rm` with collision warning at set time
- [ ] Security e2e: alias value still passes risk/policy

---

## Update log

```
```
