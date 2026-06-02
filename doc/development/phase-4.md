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
2026-05-30 · 4.2 · intent: wire memory resolver for session follow-ups and post-run append · eb5b6d6
2026-05-30 · 4.3 · scripts: add opt-in shell integration hooks for bash and powershell · 9d81ffb
2026-05-30 · 4.4 · cmd: add interactive clx init setup wizard · 13a9c2b
2026-05-30 · 4.5 · providers: load per-skill prompt templates for AI resolution · 85c9d7c
```
