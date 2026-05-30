# Phase 4 — Advanced UX

> **Status:** Complete (4.1–4.4 + optional 4.5).

---

## Status snapshot

| Sub-phase | Scope | Status | Notes |
|-----------|-------|--------|-------|
| **4.1** | Memory package | ✅ Done | `internal/memory` |
| **4.2** | Session follow-ups | ✅ Done | Resolver + append |
| **4.3** | Shell integration | ✅ Done | Opt-in hook scripts |
| **4.4** | `clx init` wizard | ✅ Done | Interactive setup |
| **4.5** | Skill prompts (stretch) | ✅ Done | Per-pack `prompts.yaml` |

---

## Tasks

- [x] Session JSON store with bounds and redaction
- [x] Memory resolver first in chain; post-run append
- [x] Bash/PowerShell explain-only hooks
- [x] `clx init` wizard
- [x] (Stretch) per-domain skill prompts

---

## Update log

```
2026-05-30 · 4.1 · memory: add session-scoped JSON store with bounded command history · dd68b14
```
