# Phase 3 — Safety Hardening

> **Status:** Not started. Prerequisites complete; Phase 3 proper next.
>
> This doc is both the implementation plan **and** the live tracker. Flip
> checkboxes (`[ ]` → `[x]`) and append to the **Update log** at the bottom as
> work lands. Cross-reference `doc/architecture.md` §3.7–§3.9 (Risk, Policy,
> Execution) and §6 (Phases) for the architectural source of truth.

---

## Status snapshot

| Sub-phase | Scope | Status | Last commit | Notes |
|-----------|-------|--------|-------------|-------|
| **P.1** | Prerequisites: YAML decoder, config drift, auto_execute fix, cache secret guard | ✅ Done | — | Unblocks Phase 3 start; see update log. |
| **3.1** | Risk engine hardening | ✅ Done | — | Destructive patterns extended; RequiresConfirmation removed. |
| **3.2** | Policy allow-list + argv-aware matching | ✅ Done | — | Argv block match, allow-list, mtime reload. |
| **3.3** | Access levels (Safe / Moderate / Full) | ✅ Done | — | `access_level` in policy.yaml; gated in policy.Check. |

Legend: ⬜ Not started · 🟡 In progress · ✅ Done · 🛑 Blocked

---

## TL;DR

Phase 3 completes the safety layer that architecture describes but Phase 1.6
only stubbed. **Much of the UX-facing safety already shipped in Phase 2.8**
(dry-run preview, confirmations, `clx safety`, mode×risk matrix) — do **not**
re-implement those.

**What remains:**

- Harden `internal/risk` classification (especially now that AI command
  generation can emit arbitrary argv)
- Implement policy `allowed`-list semantics and argv-aware block matching
- Add policy **access levels** (`Safe` / `Moderate` / `Full`) as a separate
  axis from safety **mode** (`low` / `medium` / `high` / `custom`)

**What does NOT land:** aliases (Phase 3.5), shell hooks, session memory
expansion, embeddings.

---

## Already shipped (Phase 2.8 — do not redo)

| Capability | Where |
|------------|-------|
| Safety mode×risk matrix (`DecideSafetyAction`) | `internal/config/safety.go` |
| `clx safety show/set` subcommand | `cmd/clx/` |
| Dry-run preview line | `internal/pipeline/safety.go` |
| User confirmation prompt | `internal/pipeline/confirm.go` |
| Preview-then-confirm in high mode | `internal/pipeline/run.go` |
| Executor hard-fail without risk/policy | `internal/executor/run.go` |
| BlockYes: high mode cannot skip confirm via `-y` | `internal/config/safety.go` |
| BlockYes wins over `auto_execute` (prerequisite fix) | `internal/config/safety.go` |

---

## Two safety axes (do not conflate)

| Concept | Values | Source | Status |
|---------|--------|--------|--------|
| **Safety mode** | `low` / `medium` / `high` / `custom` | `config.yaml` / `clx safety` | ✅ Phase 2.8 |
| **Policy access level** | `safe` / `moderate` / `full` | `~/.clx/policies/policy.yaml` | ✅ Phase 3.3 |

Safety mode controls UX gates (explain / preview / confirm) after risk
classification. Access level controls which command categories may execute at
all. Both run in the pipeline; neither bypasses the other.

---

## Foundation already in place

| Capability | Where |
|------------|-------|
| `risk.Assess` heuristic (low/medium/high verbs, git/docker subverbs) | `internal/risk/assess.go` |
| `policy.Check` block-list (argv token match) | `internal/policy/check.go` |
| Policy file load + cache | `internal/policy/load.go` |
| Default policy template | `internal/config/templates/policy.yaml` |
| Pipeline order: risk → policy → safety → exec | `internal/pipeline/run.go` |
| AI argv + intent validation before gates | `internal/pipeline/run.go`, `aicommand.go` |

---

## Locked decisions (do not relitigate without doc update)

| # | Decision | Rationale |
|---|----------|-----------|
| **D1** | **Phase 2.8 safety UX is frozen.** Phase 3 extends classification and policy, not the confirm/dry-run matrix. | Avoid relitigating shipped behavior. |
| **D2** | **`BlockYes` wins over `auto_execute` and `-y`.** High mode always confirms medium/high-risk commands. | Architecture §3.9 invariant; fixed in prerequisite P.1. |
| **D3** | **Intent cache skips secret-shaped params** (no Put). Re-resolve via AI instead of redacting cached values. | Redaction would corrupt cache hits; `executor.ContainsSecret` is the single matcher. |
| **D4** | **`features.ai_command_generation` defaults on.** Heuristic path is opt-out; Medium/High always confirm. | Deliberate product choice per architecture §3.10. |

---

## Phase 3.1 — Risk engine hardening

> **Goal:** Replace the Phase 1.6 heuristic stub with a maintainable classifier
> that keeps pace with AI-generated argv.

- [x] Audit and extend destructive-pattern lists for AI command generation path
- [x] Confirm risk classifies rule-rendered `gen.Argv`, not shell-host wrapper tokens
- [x] Decide fate of `RiskAssessment.RequiresConfirmation` (likely remove; safety matrix owns confirm)
- [x] Expand test matrix for cross-shell destructive patterns
- [x] Remove "Phase 1.6 heuristic stub" comment when done

---

## Phase 3.2 — Policy engine hardening

> **Goal:** Replace naive `strings.Contains` block matching with argv-aware checks;
> implement the `allowed` list from policy.yaml.

- [x] Design argv-aware block matching (avoid `format` false positives, close spacing bypasses)
- [x] Implement `allowed`-list semantics (architecture §3.8)
- [x] Policy reload or invalidation strategy (currently `sync.Once` cache)
- [x] Adversarial test suite mirroring risk tests
- [x] Update default policy template documentation

---

## Phase 3.3 — Access levels

> **Goal:** Safe / Moderate / Full access levels from architecture §3.8.

- [x] Add `access_level` (or equivalent) to policy config schema
- [x] Wire access level into `policy.Check` after block/allow lists
- [x] CLI or config path to set access level
- [x] Tests: Safe = explain-only, Moderate = read-only auto-allow, Full = most ops
- [x] Reconcile architecture §3.8 docs with implementation

---

## Cross-cutting non-negotiables

| # | Constraint |
|---|------------|
| C1 | Pipeline order unchanged: Generate → Risk → Policy → Safety → DryRun → Confirm → Exec |
| C2 | `executor.Run` must hard-fail without risk + policy |
| C3 | No `exec.Command` outside `internal/executor` |
| C4 | Policy and risk changes must cover AI-generated argv path |
| C5 | Secret-shaped values never persisted to intent cache |

---

## Risks and open questions

| # | Item | Owner | Status |
|---|------|-------|--------|
| R1 | Policy matching strategy: argv token match vs normalized command string | — | Resolved: argv subsequence |
| R2 | Access level config shape: policy.yaml field vs config.yaml | — | Resolved: `policy.yaml` `access_level` |
| R3 | Legacy `safety.level` key in config (`apply.go` TODO) — deprecate and warn? | TBD | Open |
| R4 | Risk pattern maintenance cadence with AI command generation enabled | TBD | Open |

---

## Update log

Append one line per merged commit. Format: `YYYY-MM-DD · <task id> · <short summary> · <commit sha>`.

```
2026-05-30 · P.1 · Phase 3 prerequisites: YAML inline comments, config drift, auto_execute fix, cache secret guard · (pending)
2026-05-30 · 3.1 · risk: harden destructive-pattern classifier and drop vestigial RequiresConfirmation · 8036768
2026-05-30 · 3.2 · policy: replace substring block matching with argv-aware token matching · 2de873a
2026-05-30 · 3.3 · policy: implement allow-list semantics and mtime-based reload · 09ae8f9
2026-05-30 · 3.4 · policy: add safe/moderate/full access levels gating execution · (pending)
```

---

## How to use this doc

1. Pick the next unchecked task in the active sub-phase.
2. Implement it in one commit on `development`.
3. Flip the box, add a line to **Update log**, push.
4. When a sub-phase completes, flip its row in **Status snapshot** to ✅.
5. If a decision changes, update **Locked decisions** in the same commit.
