# CLX lite / clx-ai startup split

CLX ships two binaries in the same install directory:

| Binary | Role | On PATH |
|--------|------|---------|
| `clx` | Lite front: rules-first hot path, subcommands | Yes |
| `clx-ai` | Full pipeline with AI providers (internal) | No |

Users always invoke `clx`. When rules, memory, and cache miss and AI is enabled, the lite front execs `clx-ai` with the same argv (argv-only, no shell).

## Build

```bash
make build   # bin/clx (lite), bin/clx-ai, bin/clxmax
```

- `clx`: `go build -tags=lite ./cmd/clx`
- `clx-ai`: `go build ./cmd/clx-ai` (full pipeline, `!lite` tag)
- Tests: `go test ./...` (full) + `go test -tags=lite ./internal/pipeline/ ./internal/clxsidecar/ ./cmd/clx/`

## Windows subprocess benchmarks (dev machine, unsigned)

Measured with `scripts/profile-startup.ps1` (3 runs, warm). Times dominated by Windows Defender scanning unsigned PE files.

### Before (monolithic 7.56 MB `clx.exe`)

| Phase | Median |
|-------|--------|
| `--version` | ~4740 ms |
| Rule hit (`pwd --explain`) | ~4430 ms |
| Rule miss | ~4500 ms |

### After (lite 4.05 MB `clx.exe` + 7.33 MB `clx-ai.exe`)

| Phase | Median |
|-------|--------|
| `--version` | ~3060 ms |
| Rule hit (`pwd --explain`) | ~3360 ms |
| Rule miss (`--provider none`) | ~3140 ms |
| NL rule hit | ~3560 ms |

### Binary sizes

| Artifact | Size |
|----------|------|
| `clx` (lite) | 4.05 MB |
| `clx-ai` (worker) | 7.33 MB |
| Monolith (previous) | 7.56 MB |

**Rule-path improvement:** ~30–35% faster median spawn on this machine. AI miss path pays lite spawn + worker spawn (not benchmarked here; rare path).

In-process logic remains fast (microseconds for rule resolve); remaining gap vs Linux CI budgets is environmental (AV), not pipeline code.

## Delegation rules (lite front)

Delegate to `clx-ai` when:

- Provider is not `none` (`aiEnabled`), or
- `features.ai_command_generation` is on and resolver miss would use AI

Otherwise handle miss locally (no double spawn for `--provider none` rule misses).

Missing worker:

```
AI support is not installed (clx-ai not found next to clx): ...
Reinstall CLX or run make install.
```

## Install layout

Both binaries live in the same directory (e.g. `%LOCALAPPDATA%\Programs\clx\` on Windows). Only that directory is added to PATH.
