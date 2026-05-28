Review the changes on @branch against CLX's architecture, runtime budgets, and safety contract. Read [`doc/architecture.md`](../../doc/architecture.md), [`.cursor/rules/safe-command-execution.mdc`](../rules/safe-command-execution.mdc), [`.cursor/rules/runtime-footprint.mdc`](../rules/runtime-footprint.mdc), and [`.cursor/rules/memory-management.mdc`](../rules/memory-management.mdc) before starting.

## Pipeline integrity

- Does the change preserve the canonical order `parser → intent → environment → capabilities → generator → risk → policy → executor`? Any new step added with a clear owner package, or shoehorned into the wrong stage?
- Are component boundaries respected? Each stage lives in its own `internal/<name>/` package; cross-package access goes through interfaces, not internals.
- Are new contracts (`Request`, `ResolvedIntent`, `GeneratedCommand`, `RiskAssessment`, etc.) extended in a backwards-compatible way, or do consumers in `cmd/clx` / `cmd/clxmax` and `test/` need updates landed in the same PR?
- For new rules under `internal/builtin/rules/` or `internal/builtin/skills/*/intents.yaml`: do they declare `intent`, `examples`, and per-shell `strategies` with `{{param}}` placeholders only — no inline command construction?
- For user-overlay precedence (`~/.clx/rules/`, `~/.clx/skills/`): does the merge order still let user definitions win over built-in for the same `intent` name?

## Cross-platform correctness

- Does the change behave correctly on Windows (PowerShell **and** CMD), macOS, Linux, and WSL? Any silent assumption that paths are POSIX, that `sh` exists, or that line endings are `\n`?
- Are new capability strategies declared per shell (`linux`, `powershell`, `cmd`, etc.) with a sensible fallback when the preferred tool isn't installed?
- Does `internal/environment` detection still produce a complete `SystemProfile` on every supported OS? Run / inspect `clx doctor` if anything in detection changed.
- Does Phase 1.7 shell-host dispatch still go through the fixed argv (`powershell -NoProfile -NonInteractive -Command`, `cmd /c`, `sh -c`) with the script built from rule-rendered, metachar-checked argv via `BuildValidatedScript`?

## Runtime footprint (CI-enforced budgets)

- Cold start (`clx --version`) still under 50 ms? Steady-state RSS under 30 MB on the rules path? Binary under 20 MB? Run `make budgets`.
- Any new `init()` side effects (network, FS scan, goroutine, config load) introduced? They are banned — initialization must be explicit and lazy.
- Any unbounded `map`, `slice`, channel, cache, or buffer added without a documented max size + eviction policy?
- Any `io.ReadAll` on an external/unknown source instead of `io.LimitReader` + `bufio.Scanner` with explicit `MaxScanTokenSize`?
- Hot-path allocations: are slice/builder capacities pre-sized? Are reusable buffers pooled via `sync.Pool`? Is `map[string]any` used in the parser / intent / generator hot path instead of typed structs?

## Dependencies & build

- Any new entry in `go.mod`? Each direct dep needs a one-line justification in the PR description. Forbidden categories: ORMs, web frameworks, reflection-heavy serializers, anything pulling >5 transitive deps.
- CGO still off? Release builds still use `-trimpath -ldflags="-s -w"`?
- `depguard`, `golangci-lint`, `gosec`, and `staticcheck` all clean? Custom `forbidigo` rule (no `exec.Command(` outside `internal/executor/`, no `sh -c` / `bash -c` / `powershell -Command` / `cmd /C` string literals) still satisfied?

## Resource hygiene

- Every `Open` / `Dial` / `http.Get` followed immediately by a `defer Close()` on the line after the error check?
- Every new goroutine owned by a `context.Context` or a `done` channel with a deterministic exit? Run `goleak` in tests for any package that spawns goroutines.
- Every long-running function takes `ctx context.Context` as its first parameter and honors `ctx.Done()` in loops?
- Every `exec.CommandContext` call uses the timeout from `config.execution.timeout`, not a hardcoded value?

## Memory contract

- All session persistence goes through `internal/memory`. No other package reads or writes `~/.clx/sessions/<id>.json` directly.
- New fields added to `Session`? They must fit the allowed struct (`SchemaVersion`, `SessionID`, `StartedAt`, `Commands`, `Aliases`, `Preferences`) and ship with a `schema_version` bump and a doc update.
- Writes still atomic (tmp file + rename) and bounded by `max_entries_per_session`, `max_sessions`, and TTL from `config.yaml`?
- No embeddings, vector stores, RAG indices, repo indexing, background watchers, or external databases (SQLite/BoltDB) introduced. AI providers (`internal/providers/*`) must remain stateless and never import `internal/memory`.
- Memory failure still degrades gracefully — the pipeline runs without it.

## CLI / config / runtime layout (the real "API surface")

- Flag changes on `cmd/clx` or `cmd/clxmax`: still backwards-compatible? Exit-code semantics preserved?
- New `~/.clx/` directory entry? Add it to the runtime layout section in `doc/architecture.md` §4.
- Config schema (`configs/config.example.yaml`) changes: full schema ships in 1.1 — quietly add fields rather than introducing migrations. Defaults safe (`safety.mode: medium`, `dry_run: true`, `auto_execute: false`)?
- New alias / skill behavior: alias expansion still single-level, still parser-stage, still flows through full `risk → policy → exec` chain?

## Tests

- `go test -race ./...` clean? Unit tests live beside source (`foo.go` + `foo_test.go`), cross-package and CLI e2e tests in `test/`.
- For hot paths (parser, intent, generator): `go test -bench=. -benchmem` showing no `allocs/op` regression.
- For anything touching `internal/executor`: fuzz tests in `quote_test.go` still pass for each `Quote*` function (POSIX, PowerShell, CMD)?
- For anything that spawns goroutines: `goleak` in test teardown?
- E2E matrix (Windows PS + CMD, macOS, Linux, WSL) green for changed rules/intents?

## Observability

- All logging routed through `internal/logging` (structured, leveled). No `fmt.Println` / `log.Println` on hot paths.
- Logs and any persisted state run through the secret-redaction layer before display or write.
- New error paths return wrapped errors (`fmt.Errorf("...: %w", err)`) with enough context to identify the failing stage.

## Scope guard

- Does the change drift into anything `doc/architecture.md` §7 explicitly excludes (autonomous coding agent, full repo AI assistant, IDE replacement, heavy AI workflow, plugin marketplace)? If yes, push back or scope down.

## Final sanity

- Auth flows, command-generation logic, AI-output handling, or anything touching `internal/executor`, `internal/risk`, or `internal/policy` changed? Run `/security-review` before merging.
