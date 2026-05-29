# CLX — V1 Architecture

> **Write commands once. Run anywhere.**

CLX is a lightweight, AI-assisted cross-platform command intelligence layer. It understands user intent, translates commands between environments, and executes them safely across Windows CMD, PowerShell, Bash, Zsh, Linux, macOS, Ubuntu, and WSL.

This document defines the target V1 architecture. Runtime state lives in `~/.clx/`; the repo contains the engine, rules, skills, and shipped defaults.

---

## 1. Design principles

| Principle | Meaning |
|-----------|---------|
| **Rules first** | Most requests resolve via deterministic YAML rules — no LLM call |
| **AI fallback** | LLM invoked only for ambiguous, natural-language, or broken inputs |
| **Safe by default** | Dry-run, confirmation, risk classification, and policy enforcement |
| **Explainable** | Every translation shows intent, native command, and risk level |
| **Cross-platform** | Single Go binary; environment-aware command generation |
| **Lightweight** | Fast startup, minimal dependencies, session-scoped memory only |

Runtime footprint budgets (binary size, cold start) are enforced in CI via [`scripts/check-budgets.sh`](../scripts/check-budgets.sh); run `make budgets` locally.

---

## 2. High-level pipeline

```mermaid
flowchart TD
    user["User: clx grep errors logs.txt"] --> cli["CLI Entry (cmd/clx)"]
    cli --> parser["Parser - normalize input"]
    parser --> intent["Intent Resolver - rules first"]
    intent -->|"miss / ambiguous"| ai["AI Provider - Ollama / OpenAI / Azure"]
    ai --> intent
    intent --> envres["Environment Resolver - OS, shell, tools"]
    envres --> cap["Capability Engine - rg vs grep, fd vs find"]
    cap --> gen["Command Generator - render template"]
    gen --> risk["Risk Engine"]
    risk --> policy["Policy Engine - allow / block lists"]
    policy --> exec["Execution Engine - dry-run, confirm, execute"]
    exec --> out["Output + explanation"]

    subgraph support [Cross-cutting]
        cfg["Config Loader (~/.clx/config.yaml)"]
        prof["System Profile (~/.clx/system_profile.json)"]
        cache["Cache (~/.clx/cache/)"]
        mem["Session Memory (~/.clx/sessions/)"]
        skills["Skills (git, docker, k8s, fs, net)"]
        logs["Logger"]
    end

    cfg -.-> intent
    cfg -.-> exec
    prof -.-> envres
    prof -.-> ai
    cache -.-> intent
    mem -.-> intent
    skills -.-> intent
    logs -.-> cli
```

### Example flow

```
Input:  clx grep errors logs.txt
  ↓ Parser         → raw shell command detected
  ↓ Intent         → search_text_in_file { pattern: "errors", file: "logs.txt" }
  ↓ Environment    → { os: windows, shell: powershell, tools: [git, docker, node] }
  ↓ Capability     → Select-String preferred over findstr (PowerShell native)
  ↓ Generator      → Select-String "errors" logs.txt
  ↓ Risk           → { risk: low }
  ↓ Policy         → allowed
  ↓ Execution      → Execute? [Y/n]
```

---

## 3. Component contracts

Each component lives in `internal/<name>/` as a private Go package. Interfaces are defined per package; concrete implementations are wired in `cmd/clx`.

### 3.1 CLI Entry — `cmd/clx`, `cmd/clxmax`

**Responsibility:** Parse argv, load config, wire the pipeline, handle exit codes.

| Flag | Purpose |
|------|---------|
| `--explain` | Show intent + translation without executing |
| `--dry-run` | Preview command and affected resources |
| `--yes` / `-y` | Skip confirmation prompt |
| `--provider` | Override AI provider for this invocation |

**Binaries:**

- **`clx`** — Fast, rules-first, single-shot translation/execution.
- **`clxmax`** — Same engine with reasoning, multi-step planning, and clarification loops.

---

### 3.2 Parser — `internal/parser`

**Responsibility:** Normalize raw user input into a structured request.

**Input types:**

| Type | Example |
|------|---------|
| Raw shell command | `grep errors logs.txt` |
| Natural language | `find all files modified today` |
| Partial / broken shell | `locate help.txt` (on Windows) |
| CLX invocation | `clx grep errors logs.txt` |

**Output contract:**

```go
type Request struct {
    RawInput   string
    InputType  InputType   // Shell | NaturalLanguage | PartialShell | CLXInvocation
    Tokens     []string
    Args       map[string]string // args after "clx" prefix stripped
}
```

**Alias expansion (Phase 3.5+):** before classification, the parser consults `internal/aliases` and, on a hit for the first token, rewrites `RawInput` with the alias value before continuing. Expansion is single-level (no alias-of-alias chains). See [§3.16](#316-aliases--internalaliases).

---

### 3.3 Intent Resolver — `internal/intent`

**Responsibility:** Map input to a semantic intent and extracted parameters.

**Resolution order:**

1. Session memory (contextual follow-ups)
2. Rule engine (embedded `internal/builtin/rules/*.yaml`, `internal/builtin/skills/*/intents.yaml`; user overlays in `~/.clx/rules/`, `~/.clx/skills/`)
3. Cache lookup (`~/.clx/cache/`)
4. AI provider fallback

**Output contract:**

```go
type ResolvedIntent struct {
    Intent     string            // e.g. "find_file", "search_text_in_file"
    Params     map[string]string // e.g. { "filename": "help.txt" }
    Confidence float64
    Source     IntentSource      // Rule | Cache | AI | Memory
}
```

**Rule format** (store intents, not raw command mappings):

```yaml
intent: find_file
examples:
  - locate help.txt
  - find file help.txt
strategies:
  linux:
    primary: "find . -name {{filename}}"
  powershell:
    primary: "Get-ChildItem -Recurse -Filter {{filename}}"
  cmd:
    primary: "dir /s {{filename}}"
```

---

### 3.4 Environment Resolver — `internal/environment`

**Responsibility:** Detect and maintain the current machine profile.

**Detects:** OS, OS version, shell, shell version, terminal, package managers, installed tools, WSL status, key paths.

**Output contract:**

```go
type SystemProfile struct {
    OS               string
    OSVersion        string
    Shell            string
    ShellVersion     string
    Terminal         string
    PackageManagers  []string
    AvailableTools   []string
    WSLEnabled       bool
    Paths            map[string]string
}
```

Persisted to `~/.clx/system_profile.json`. Refreshed on install and on demand (`clx doctor`).

---

### 3.5 Capability Engine — `internal/capabilities`

**Responsibility:** Select the best available strategy for an intent given the environment.

**Example logic:**

```
IF rg installed → use rg over grep
IF fd installed → use fd over find
IF powershell   → prefer Select-String over findstr
```

**Output contract:**

```go
type Strategy struct {
    Tool       string   // e.g. "rg", "Select-String", "grep"
    Template   string   // command template with {{placeholders}}
    Priority   int
}
```

---

### 3.6 Command Generator — `internal/generator`

**Responsibility:** Render the selected strategy template with intent parameters into a final native command string.

**Output contract:**

```go
type GeneratedCommand struct {
    Command     string
    Shell       string   // target shell for execution
    Explanation string   // human-readable description
}
```

---

### 3.7 Risk Engine — `internal/risk`

**Responsibility:** Classify every generated command before execution.

**Risk levels:**

| Level | Examples |
|-------|---------|
| **Low** | `ls`, `grep`, `cat`, `git status` |
| **Medium** | installs, git push, docker run |
| **High** | recursive delete, format, shutdown |

**Output contract:**

```go
type RiskAssessment struct {
    Level                RiskLevel // Low | Medium | High
    Reason               string
    RequiresConfirmation bool
}
```

---

### 3.8 Policy Engine — `internal/policy`

**Responsibility:** Enforce user-defined allow/block lists and access levels.

**Access levels:**

| Level | Behavior |
|-------|----------|
| **Safe (0)** | Explain only — no execution |
| **Moderate (1)** | Read-only + safe ops auto-allowed |
| **Full (2)** | Most ops allowed; still blocks destructive OS-level commands |

**Policy file** (`~/.clx/policies/policy.yaml`):

```yaml
blocked:
  - "rm -rf /"
  - "shutdown"
  - "format"
allowed:
  - "git"
  - "docker"
  - "npm"
```

---

### 3.9 Execution Engine — `internal/executor`

**Responsibility:** Dry-run preview, user confirmation, timeout enforcement, shell-aware execution.

**Flow:**

```
Generate Command → Risk Scan → Policy Check → Dry Run Preview → User Confirmation → Execute
```

**Config knobs** (from `config.yaml`):

```yaml
execution:
  auto_execute: false
  timeout: 30
safety:
  mode: medium
  require_confirmation: true
  dry_run: true
```

---

### 3.10 AI Provider Layer — `internal/providers`

**Responsibility:** Pluggable LLM interface for intent resolution, explanation, and reasoning.

**Providers:**

| Provider | Package | Use case |
|----------|---------|----------|
| Ollama | `internal/providers/ollama` | Local-first, offline |
| OpenAI | `internal/providers/openai` | Cloud fallback |
| Azure | `internal/providers/azure` | Enterprise |

**Interface contract:**

```go
type Provider interface {
    ResolveIntent(ctx context.Context, req IntentRequest) (*ResolvedIntent, error)
    Explain(ctx context.Context, cmd GeneratedCommand) (string, error)
    Name() string
}
```

Every AI request automatically injects the system profile as grounding context.

---

### 3.11 Memory — `internal/memory`

**Responsibility:** Lightweight session-scoped context for follow-up commands.

**Stores:** previous commands, resolved translations, session preferences.

**Does NOT store:** filesystem knowledge, embeddings, long-term autonomous planning, user-global aliases (see [§3.16](#316-aliases--internalaliases)).

**Storage:** `~/.clx/sessions/<session_id>.json`

---

### 3.12 Skills — `internal/skills`

**Responsibility:** Load domain-specific intent capability packs at startup.

**Built-in skills:** git, docker, kubernetes, filesystem, networking.

**Skill structure:**

```
skills/git/
  intents.yaml    # domain intents and examples
  prompts.yaml    # AI prompt templates for this domain
```

---

### 3.13 Cache — `internal/cache`

**Responsibility:** Memoize resolved intents and AI translations to avoid repeated LLM calls.

**Storage:** `~/.clx/cache/`

**Cache key:** hash of (input + os + shell + installed_tools_hash)

---

### 3.14 Config — `internal/config`

**Responsibility:** Load, validate, and provide defaults for `~/.clx/config.yaml`.

```yaml
provider: ollama
# Default model for the active provider.
# Default Ollama: qwen3:1.7b (CPU). Quality tier: qwen3:4b (GPU).
model: qwen3:1.7b

providers:
  ollama:
    host: "http://localhost:11434"
    # Alternates: qwen3:4b (quality), qwen2.5:7b, llama3.1:8b
    model: qwen3:1.7b
  openai:
    api_key: ""
    model: gpt-4.1-mini
  azure:
    endpoint: ""
    api_key: ""
    deployment: ""

execution:
  auto_execute: false
  timeout: 30
  shell_integration: false

safety:
  mode: medium
  require_confirmation: true
  dry_run: true

features:
  explain: true
  cache_commands: true
  learning_mode: false

logging:
  enabled: true
  level: info
```

---

### 3.15 Logger — `internal/logging`

**Responsibility:** Structured logging to `~/.clx/logs/`. Levels: debug, info, warn, error.

---

### 3.16 Aliases — `internal/aliases`

**Responsibility:** Persistent, user-global input shortcuts. Distinct from `internal/memory` (which is session-scoped and ephemeral).

**Storage:** `~/.clx/aliases.yaml` (flat YAML, atomic tmp+rename writes).

**Schema:**

```yaml
schema_version: 1
aliases:
  - name: prd
    value: "cd /abc/clx/prd"
  - name: gst
    value: "git status"
```

**Lifecycle:**

| Operation | CLI | Behavior |
|---|---|---|
| Create | `clx alias set <name> "<value>"` | Warn if `<name>` collides with a known shell verb or built-in rule example head. `--force` to override. |
| List | `clx alias list` | Print all aliases. |
| Delete | `clx alias rm <name>` | Remove from `aliases.yaml`. |
| Invoke | `clx <name> [args]` | Parser-stage rewrite. Alias value replaces `<name>` in the input, then continues through the normal pipeline. |

**Pipeline placement:** alias expansion happens inside the parser (§3.2), **before** intent resolution. The expanded value flows through the full `intent → generator → risk → policy → executor` chain — aliases get **zero** privileged path and never bypass safety gates (per [`.cursor/rules/safe-command-execution.mdc`](../.cursor/rules/safe-command-execution.mdc)).

**Collision policy:** aliases always win over rules at runtime (shell-alias semantics). The warning at `alias set` time is the user's one chance to catch shadowing; once set, no invoke-time flag is required.

**Bounds (config-driven, default in `config.yaml` under `aliases:`):**

- `max_aliases: 256`
- Load lazily on first parser invocation, cache for the process lifetime, no background watcher (per [`.cursor/rules/runtime-footprint.mdc`](../.cursor/rules/runtime-footprint.mdc)).

**Does NOT do:** function-style aliases with positional substitution, regex aliases, alias-of-alias chains (resolved one level only — explicit by design to prevent infinite loops and surprise).

---

## 4. Runtime directory layout (`~/.clx/`)

Created on first run / install — **not committed to the repo.**

```
~/.clx/
├── config.yaml
├── aliases.yaml       # user-global alias shortcuts (see §3.16)
├── system_profile.json
├── cache/
├── memory/
├── sessions/
├── policies/
│   └── policy.yaml
├── rules/             # user rule overrides (*.yaml; same intent name wins over built-in)
├── skills/            # user skill pack overrides (*/intents.yaml)
└── logs/
```

User overlay files use the same YAML shapes as built-in rules. When the same `intent` name appears in both built-in and user content, **the user definition wins** (later merge order).

---

## 5. Repo directory layout

```
clx.ai/
├── cmd/
│   ├── clx/                  # main CLI binary
│   └── clxmax/               # advanced reasoning binary
├── internal/                 # private engine packages
│   ├── parser/
│   ├── intent/
│   ├── environment/
│   ├── capabilities/
│   ├── generator/
│   ├── risk/
│   ├── policy/
│   ├── executor/
│   ├── providers/
│   │   ├── ollama/
│   │   ├── openai/
│   │   └── azure/
│   ├── memory/
│   ├── aliases/
│   ├── cache/
│   ├── skills/               # skill loader (delegates to intent)
│   ├── builtin/              # embedded built-in rules + skills (//go:embed source)
│   │   ├── rules/
│   │   └── skills/
│   ├── config/
│   └── logging/
├── pkg/                      # future public/reusable libs
├── profiles/                 # example user/team profiles
├── policies/                 # default policy templates
├── configs/                  # example config.yaml
├── scripts/                  # install.sh, install.ps1
├── test/                     # integration & e2e tests
└── doc/                      # architecture & design docs
```

**Conventions:**

- Unit tests live beside source (`foo.go` + `foo_test.go`).
- `test/` is for cross-package integration and CLI end-to-end tests.
- `internal/` is private — engine API is free to evolve until V1 stabilizes.

---

## 6. Implementation phases

| Phase | Scope | Key packages |
|-------|-------|--------------|
| **Phase 1 — Core Engine** | Rules-first deterministic pipeline (no AI, no policy enforcement) — see [§6.1](#61-phase-1-sub-phases) for breakdown | `config`, `logging`, `environment`, `parser`, `intent` (rules path), `skills` (loader), `capabilities`, `generator`, `executor` (basic), `cmd/clx` |
| **Phase 2 — AI Integration** | Ollama + OpenAI providers, AI fallback, explanations | `providers/*`, `intent` (AI path), `cache` |
| **Phase 3 — Safety** | Risk engine, policy engine, dry-run, confirmations, access levels | `risk`, `policy`, `executor` (safety hooks) |
| **Phase 3.5 — Aliases** | Persistent user-global aliases in `~/.clx/aliases.yaml`. `clx alias set/list/rm` subcommand, parser-stage expansion (alias value flows through full risk/policy/exec chain), set-time collision warning against shell verbs and built-in rule example heads. No dependency on `internal/memory` or shell hooks — ships as a self-contained slice between safety and advanced UX. See [§3.16](#316-aliases--internalaliases). | `internal/aliases`, `internal/parser` (expansion hook), `cmd/clx` (`alias` subcommand) |
| **Phase 4 — Advanced UX** | Shell interception, auto-fix, session context, interactive `clx init` wizard | `memory`, `skills`, shell hooks |

### 6.1 Phase 1 sub-phases

Phase 1 is split into six dependency-ordered slices. Each slice is independently shippable, end-to-end testable, and unblocks the next.

| Sub-phase | Scope | Packages | Exit criteria |
|-----------|-------|----------|---------------|
| **1.1 — Foundation & Bootstrap** | Config schema + loader, structured logging, install scripts, first-run bootstrap of `~/.clx/` (config, dirs, logs). Default config baked in: `provider: ollama`, `safety.mode: medium`, `dry_run: true`. No interactive prompts. | `internal/config`, `internal/logging`, `scripts/install.sh`, `scripts/install.ps1` | `clx --version` works; first run creates the full `~/.clx/` tree with `config.yaml` from `configs/config.example.yaml`. |
| **1.2 — Environment Detection** | Detect OS, OS version, shell, shell version, terminal, package managers, installed tools, WSL state, key paths. Persist to `~/.clx/system_profile.json`. Ship `clx doctor` to refresh on demand. | `internal/environment` | `clx doctor` writes a complete, accurate `system_profile.json` on Windows (PowerShell + CMD), macOS, Linux, and WSL. |
| **1.3 — Parser** | Normalize raw input into a `Request`. Classify as `Shell`, `NaturalLanguage`, `PartialShell`, or `CLXInvocation`. Strip the `clx` prefix and tokenize args. | `internal/parser` | Unit tests pass for all four input types across representative samples. |
| **1.4 — Rules-First Intent Resolver** | YAML rule loader for built-in rules/skills (`internal/builtin/`, embedded in binary) and optional user overlays (`~/.clx/rules/`, `~/.clx/skills/`). Match input → `ResolvedIntent` with extracted params. Skill pack loader (loader only — no AI prompts yet). **No AI fallback, no cache, no memory.** | `internal/intent` (rule path), `internal/skills` (loader), `internal/builtin` | Seed rule set (e.g. `find_file`, `search_text_in_file`, `list_dir`, `current_dir`, `disk_usage`) resolves correctly with `Source: Rule`. |
| **1.5 — Capabilities & Generator** | Pick the best strategy for a resolved intent given the `SystemProfile` (e.g. `rg` over `grep`, `Select-String` over `findstr`). Render the chosen template into a final native command string. | `internal/capabilities`, `internal/generator` | `(ResolvedIntent + SystemProfile) → GeneratedCommand` produces the expected native command per shell for the seed rule set. |
| **1.6 — Basic Executor & CLI Wiring** | Shell-aware execution with timeout, risk/policy gates, dry-run and confirm. Direct `exec.CommandContext` for PATH binaries (`git`, `ping`, `findstr`). PowerShell cmdlets and CMD builtins deferred to **1.7** (explain-only on Windows for `pwd` / `ls` until then). | `internal/executor`, `internal/risk`, `internal/policy`, `cmd/clx` | `clx --explain` and argv-only exec for PATH tools on all three OSes; e2e matrix green. |
| **1.7 — Shell-native execution** | `ExecHost` on `GeneratedCommand`; validated script assembly in `internal/executor` (`BuildValidatedScript`, host resolution); dispatch via fixed host argv (`powershell -NoProfile -NonInteractive -Command`, `cmd /c`, `sh -c` fallback on Unix). Dry-run shows effective invocation via `FormatInvocation`. | `internal/generator`, `internal/executor` | `clx -y pwd` executes on Windows (Get-Location); `clx -y grep PAT FILE` runs Select-String; Linux/macOS direct argv unchanged. |

### 6.1.7 Phase 1.7 security contract

Shell hosts are allowed only when the script is built from rule-rendered argv: each token is metachar-checked, then joined with existing quoters in `internal/executor/quote.go`. User/AI raw strings never become `-Command` / `/c` input. `os/exec` stays in `internal/executor` only.

**Notes on what is intentionally NOT in Phase 1:**

- WSL-in-Windows routing to `wsl.exe`, arbitrary user `-Command` strings, CMD `cd` vs `pwd` semantics → later
- LLM provider selection and AI fallback → **Phase 2**
- Risk classification and access-level enforcement (Safe / Moderate / Full) → **Phase 3**
- Interactive setup wizard (`clx init`) → **Phase 4** (silent install with safe defaults is sufficient for Phase 1)
- Cache, session memory, `clxmax` reasoning binary → later phases

The full config schema (`configs/config.example.yaml`) ships in **1.1** even though several fields (`providers.*`, `safety.mode`, etc.) are not yet consumed. This avoids config migrations as later phases come online — they simply start reading fields that have been quietly present since 1.1.

---

## 7. What CLX is NOT (V1 scope guard)

- Not an autonomous coding agent
- Not a full repo AI assistant
- Not an IDE replacement
- Not a heavy AI workflow platform
- Not a plugin marketplace (yet)

Focus: **trustable cross-platform command abstraction** — fast, reliable, safe, portable.
