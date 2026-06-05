# Provider configuration

CLX AI provider settings live in `~/.clx/config.yaml`. Use the CLI to view and update them — you do not need to edit the file by hand.

## Quick reference

```bash
# View provider settings (API keys masked)
clx config show

# Switch active provider
clx config provider use ollama
clx config provider use openai
clx config provider use gemini

# Set provider details
clx config set providers.ollama.host http://localhost:11434
clx config set providers.ollama.model qwen3:1.7b
clx config set providers.openai.model gpt-4.1-mini
clx config set providers.fallback openai

# Set API keys (interactive hidden prompt — no shell helpers needed)
clx config set providers.openai.api_key
clx config set providers.gemini.api_key
# Or explicitly:
clx config set providers.openai.api_key --stdin

# Migrate existing plaintext keys to encrypted form
clx config encrypt-secrets
```

Run `clx config help` for the full command list.

## Rules-only mode (no AI provider)

CLX does **not** require Ollama or any LLM. An AI provider is optional — it is only
consulted when the built-in rules and cache miss. To run fully offline with no AI:

```bash
# Disable the AI provider entirely (rules + cache only)
clx config provider use none      # or: clx config set provider none

# Per-run override without changing config
clx --provider none ls
```

In `none` mode a request that no rule matches reports `no matching rule for input`
instead of an AI-provider error. The first-run wizard's **"skip (rules only)"**
choice sets this for you. The shipped default provider is `ollama` (local), but it
is never mandatory — switch to `none` (or `openai`/`gemini`) at any time.

## Configurable paths

| Path | Purpose |
|------|---------|
| `provider` | Active provider (`none`, `ollama`, `openai`, `azure`, `gemini`). `none` = rules-only, no AI |
| `model` | Default model when provider block omits one |
| `providers.primary` | Optional chain primary (overrides `provider` when set) |
| `providers.fallback` | Optional fallback provider (infrastructure errors only) |
| `providers.timeout` | Provider HTTP timeout in seconds |
| `providers.ollama.host` | Ollama base URL |
| `providers.ollama.model` | Ollama model tag |
| `providers.openai.api_key` | OpenAI API key (encrypted at rest) |
| `providers.openai.model` | OpenAI model name |
| `providers.azure.endpoint` | Azure OpenAI endpoint |
| `providers.azure.api_key` | Azure API key (encrypted at rest) |
| `providers.azure.deployment` | Azure deployment name |
| `providers.gemini.api_key` | Gemini API key (encrypted at rest) |
| `providers.gemini.model` | Gemini model name (default: `gemini-2.0-flash`) |

### Features, cache, memory, execution, logging

Set via `clx config set <path> <value>` (see `clx config help`). Use `clx safety` for `safety.*` — not `config set`.

| Path | Type | Notes |
|------|------|-------|
| `features.explain` | bool | |
| `features.cache_commands` | bool | When false, pipeline skips cache read/write |
| `features.ai_command_generation` | bool | Hybrid AI argv fallback |
| `features.learning_mode` | bool | |
| `cache.max_entries` | int | Must be > 0 |
| `cache.ttl_days` | int | Must be > 0 |
| `cache.max_disk_bytes` | int | Must be > 0 |
| `memory.enabled` | bool | |
| `memory.max_entries_per_session` | int | |
| `memory.max_sessions` | int | |
| `memory.ttl_days` | int | |
| `execution.auto_execute` | bool | |
| `execution.timeout` | int | Seconds |
| `execution.shell_integration` | bool | |
| `logging.enabled` | bool | |
| `logging.level` | string | `debug`, `info`, `warn`, `error` |
| `aliases.max_aliases` | int | |

Inspect on-disk caches: `clx cache status` / `clx cache clear` (see `clx cache help`).

## Ollama on Windows with WSL

On Windows, `providers.ollama.host: http://localhost:11434` talks to the **Windows** loopback interface. If Ollama runs **inside WSL**, that URL often fails with "provider unavailable" even though `ollama list` works in the WSL terminal.

**Options (pick one):**

1. **Run Ollama on Windows** — install the Windows build and keep `localhost:11434`.
2. **Point CLX at the WSL host** — in WSL, run `hostname -I` (or check your distro's WSL IP) and set:
   ```bash
   clx config set providers.ollama.host http://<WSL-IP>:11434
   ```
   Ensure Ollama listens on `0.0.0.0` in WSL (not only `127.0.0.1` inside the VM).
3. **Run CLX inside WSL** — install `clx` in the same environment as Ollama so `localhost` matches.

When the active provider is Ollama on localhost and a connection fails, CLX prints a one-line reminder about the WSL host fix.

## Encrypted secrets

API keys are stored as `enc:v1:…` blobs in `config.yaml`, never in plaintext on disk.

- **Encryption key** is derived from OS + user identity (Windows MachineGuid, Linux `/etc/machine-id`, macOS hostname+user). If derivation fails, CLX creates `~/.clx/.secret-key` (mode `0600`) as fallback key material.
- **In memory**, keys are decrypted only during config load for provider use.
- **`clx config show`** and **`clx config get`** on secret paths always mask values (last 4 chars only).
- **Plaintext keys** in an existing config are re-encrypted automatically on the next `config set` or `config encrypt-secrets`.

Machine-bound encryption protects against casual file copy. It is not a substitute for full-disk encryption or OS access controls.

## Security practices

1. **Never commit `~/.clx/`** or local copies of `config.yaml` to git. The repo `.gitignore` blocks common accidental paths (`config.yaml`, `.clx/`, `.secret-key`).
2. **Set API keys interactively:** `clx config set providers.openai.api_key` — CLX prompts with hidden input; the key never appears on the command line or in shell history.
3. **Do not pass keys as arguments** — `clx config set providers.openai.api_key sk-...` is rejected.
4. **File permissions**: config and secret key files are written with mode `0600`; `~/.clx/` directories use `0700`.

## Run-time override

`--provider` on a normal `clx` invocation overrides the active provider for that run only and disables fallback (Phase 2 decision D10). All other provider settings still come from config.

```bash
clx --provider openai "find logs modified today"
```

## Alternate config file

```bash
clx --config /path/to/config.yaml ...
clx config show --config /path/to/config.yaml
```

Set `CLX_HOME` to relocate the entire runtime directory (default `~/.clx`).
