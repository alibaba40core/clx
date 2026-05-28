Review the security impact of the changes on @branch. CLX's job is to run shell commands generated from user and AI input — a single arg-quoting mistake or skipped gate is a remote code execution bug shipped to every user. Read [`.cursor/rules/safe-command-execution.mdc`](../rules/safe-command-execution.mdc) and [`.cursor/rules/memory-management.mdc`](../rules/memory-management.mdc) before starting; they are the trust boundary of the product.

## Argv discipline (hard ban)

- Any new `exec.Command(` outside `internal/executor/`? Forbidden — other packages call the executor interface.
- Any `exec.Command("sh", "-c", x)` / `exec.Command("bash", "-c", x)` / `exec.Command("powershell", "-Command", x)` / `exec.Command("cmd", "/C", x)` where `x` contains user, AI, or template-substituted content? Forbidden — RCE bug, must be rejected.
- Any string-concatenated command construction (e.g. `"git checkout " + branch`, `fmt.Sprintf("rm -rf %s", target)`)? Forbidden — use argv slices and template parameters.
- All subprocess execution uses `exec.CommandContext` with a timeout sourced from `config.execution.timeout`? No bare `exec.Command`.

## Pipeline gating (no skips, no reorders)

- Order strictly preserved: `Generate → Risk → Policy → DryRun → Confirm → Exec`.
- Executor hard-fails (not warns) if it receives a `GeneratedCommand` without a `RiskAssessment` attached?
- No new "fast path" that bypasses risk or policy "because the command looks safe"?
- New commands / rules / skills: every example resolves to a path that still goes through every gate, including when invoked via alias expansion (parser-stage rewrite must flow through the full chain — aliases get **zero** privileged path)?

## AI / provider output (untrusted by default)

- Every `ResolvedIntent` returned by `internal/providers/*` validated against the rule schema before any substitution? Intent name in the known set, params match declared types?
- No code path executes raw AI output (`provider.RawCompletion(...)` → `exec.*`)? The LLM never writes the final command string — it returns a resolved intent, the generator renders the template, the executor runs the argv.
- New provider added under `internal/providers/`? Stateless (no import of `internal/memory`), no hidden state, no background goroutines? Auth credentials sourced from config, never logged?
- Provider responses redacted before being persisted to session memory or logs?

## Path & input validation

- Any path that came from user, AI, or rule template substitution goes through `internal/executor.CleanAndValidatePath(p)`? It must reject `..`, null bytes, shell metacharacters, and absolute paths outside an explicit allowlist.
- Any new template parameter that takes a path / glob / pattern — validated before substitution? Substituted as **separate argv entries**, never joined into a single string?
- For Phase 1.7 shell-host scripts: each token metachar-checked, then joined via existing quoters in `internal/executor/quote.go`? User/AI raw strings never become `-Command` / `/c` input?

## Shell-quoting surface

- All POSIX / PowerShell / CMD quoting still lives in exactly one place: `internal/executor/quote.go` (`QuotePOSIX`, `QuotePowerShell`, `QuoteCmd`)?
- No new ad-hoc quoting helpers anywhere else in the tree?
- Shared table-driven test matrix in `quote_test.go` still covers every new case? Fuzz tests for each `Quote*` function still green?

## Secret hygiene

- Common token / API key / credential patterns redacted from argv, stdout, stderr, logs, **and** `internal/memory` before persistence or display?
- No secrets committed to repo (config examples, test fixtures, rule YAML)?
- New env vars or credentials read by providers / config: documented, not logged, not echoed in `--explain` / `--dry-run` output?
- Error messages and wrapped errors do not leak secrets or absolute paths from the user's machine?

## Memory & persistence

- All session writes still go through `internal/memory` — no other package touches `~/.clx/sessions/<id>.json` directly?
- New fields on `Session`? Must fit the allowed struct and ship a `schema_version` bump.
- Writes still atomic (tmp file + rename)? File permissions appropriate (0600 for anything containing user history)?
- No embeddings, vector stores, repo indexing, background watchers, daemons, or external databases introduced? No cross-session global knowledge store?
- Cache / alias / policy / rule file loads use `internal/config` for path resolution (never hardcoded `~/.clx/...`) and tolerate missing / malformed files without crashing?

## Policy & risk classification

- Risk levels for new commands set correctly? Destructive verbs (`rm -rf`, `format`, `shutdown`, recursive deletes, `dd`, force-pushes) classified High, with `RequiresConfirmation: true`?
- Policy engine still enforces user-defined allow/block lists? New commands respect `safety.mode` (Safe / Moderate / Full) semantics — Safe mode never executes, Moderate auto-allows only read-only + safe ops?
- Blocked patterns (`rm -rf /`, `shutdown`, `format`) still hard-blocked even when invoked via alias, skill, or AI-resolved intent?

## Dependencies (supply chain)

- Each new `go.mod` direct dep justified in the PR? Pinned to an exact version (not a floating range)?
- New dep expands network surface (HTTP client, gRPC, etc.) or subprocess surface (calls into other binaries)? Document and gate.
- Forbidden categories still excluded: ORMs, web frameworks, reflection-heavy serializers, anything pulling >5 transitive deps?
- `gosec` G204 (subprocess with variable) and G304 (file path from variable) treated as errors, not warnings? `depguard` clean?

## Cross-platform safety

- New rule / capability behaves safely on every supported shell? No POSIX-only assumption that a Windows user could hit a destructive default on?
- WSL routing (when added): no command crosses the boundary without re-validation in the target environment?
- File operations that succeed silently on Unix (case-sensitive FS, symlinks) handled deterministically on Windows?

## Tests targeted at the threat model

- New integration test in `test/` for any pipeline-gating change — wire a fake provider returning malicious output and assert the pipeline rejects it before exec?
- Fuzz coverage in `internal/executor/quote_test.go` extended for any new quoting case?
- E2E tests covering an attacker-controlled `ResolvedIntent`, a malicious alias value, and a malicious path parameter all blocked before reaching `exec`?

## Final sentinel

- After this change, can a crafted input — natural language, raw shell, alias, AI response, or rule param — reach `os/exec` without passing argv-only validation, risk, policy, and confirmation?

  If you cannot answer "no" with confidence and a code path to point at, **block the merge.**
