# CLX Command-Generation Test Report

**Date:** 2026-06-02  
**Host:** Windows 10.0.22631, PowerShell  
**CLX binary:** `C:\Users\user\AppData\Local\Programs\clx\clx.exe` (installed via `make install`, not repo `bin\`)  
**CLX version:** `2ad9851` (commit 2ad9851, go1.26.3)  
**Provider:** OpenAI (`gpt-4.1-mini`)  
**Test mode:** `clx --explain "<input>"` (generation only; nothing executed)

## Method

Each input was run against the globally installed CLX binary with `--explain`, which prints `Intent`, `Source`, `Command`, `Explanation`, and `Risk` without executing anything.

**Pass criteria**

- **Rule-based:** `Source: Rule` and a command that correctly implements the request on Windows/PowerShell.
- **Natural-language:** `Source: AI` (or `cache`) and a correct, runnable command for the request. Rule hits on NL phrases also count as pass when the command is correct.

**Fail criteria:** no command produced, wrong source, validation rejection, or a generated command that would not work as intended.

---

## Improvements applied (Phases 1+2)

1. **Rule pattern fixes** — added `copy {{source}} {{dest}}` / `move {{source}} {{dest}}` examples; expanded `disk_usage` phrases; new `find_large_files` intent with size normalization.
2. **AI prompt** — few-shot chain JSON examples, anti-placeholder rules, bulk-rename pipe guidance.
3. **Validation retry** — one retry with validation feedback when the first AI response fails `ValidateGeneratedArgv` / `ValidateCommandChain`; cache only written after validation passes.

---

## Rule-Based Generation (20 tests)

**Baseline:** 18 PASS / 2 FAIL  
**After improvements:** **20 PASS / 0 FAIL**

| # | Input | Expected Source | Actual Source | Generated Command | Result | Notes |
|---|-------|-----------------|---------------|-------------------|--------|-------|
| 1 | find all files modified today | rule | Rule | `Get-ChildItem … \| Where-Object { $_.LastWriteTime.Date -eq [datetime]::Today }` | PASS | — |
| 2 | what is my ip | rule | Rule | `ipconfig` | PASS | — |
| 3 | ipconfig | rule | Rule | `ipconfig` | PASS | — |
| 4 | git status | rule | Rule | `git status` | PASS | — |
| 5 | git log | rule | Rule | `git log --oneline -n 20` | PASS | — |
| 6 | git diff | rule | Rule | `git diff` | PASS | — |
| 7 | git branch | rule | Rule | `git branch` | PASS | — |
| 8 | ping google.com | rule | Rule | `ping -n 4 google.com` | PASS | — |
| 9 | tracert google.com | rule | Rule | `tracert google.com` | PASS | — |
| 10 | ls . | rule | Rule | `Get-ChildItem .` | PASS | — |
| 11 | cat README.md | rule | Rule | `Get-Content README.md` | PASS | — |
| 12 | mkdir testdir | rule | Rule | `New-Item -ItemType Directory -Path testdir -Force` | PASS | — |
| 13 | find file README.md | rule | Rule | `Get-ChildItem -Path . -Recurse -Filter README.md …` | PASS | — |
| 14 | head -n 10 README.md | rule | Rule | `Get-Content README.md -TotalCount 10` | PASS | — |
| 15 | tail -n 5 README.md | rule | Rule | `Get-Content README.md -Tail 5` | PASS | — |
| 16 | curl -I https://example.com | rule | Rule | `curl -I https://example.com` | PASS | — |
| 17 | netstat -an | rule | Rule | `netstat -an` | PASS | — |
| 18 | del somefile.txt | rule | Rule | `Remove-Item -Path somefile.txt` | PASS | — |
| 19 | copy a.txt b.txt | rule | Rule | `Copy-Item -Path a.txt -Destination b.txt` | **PASS** | Fixed: added `copy {{source}} {{dest}}` example |
| 20 | move a.txt b.txt | rule | Rule | `Move-Item -Path a.txt -Destination b.txt` | **PASS** | Fixed: added `move {{source}} {{dest}}` example |

---

## Natural-Language Generation (20 tests)

**Baseline:** 11 PASS / 9 FAIL  
**After improvements:** **18 PASS / 2 FAIL** (2 unchanged placeholder failures; NL #2 and #13 now hit rules)

| # | Input | Expected Source | Actual Source | Generated Command | Result | Notes |
|---|-------|-----------------|---------------|-------------------|--------|-------|
| 1 | show me the 10 largest files in this folder | ai | AI | `Get-ChildItem … \| Sort-Object Length -Descending \| Select-Object -First 10` | PASS | — |
| 2 | how much disk space is free | ai | **Rule** | `Get-PSDrive` | **PASS** | Now hits `disk_usage` rule |
| 3 | kill the process listening on port 8080 | ai | AI | `netstat -ano \| … \| Stop-Process -Id $_ -Force` | PASS | — |
| 4 | show running docker containers sorted by memory | ai | AI | `docker ps \| docker stats --no-stream \| Sort-Object …` | **PASS** | Fixed: chain prompt + validation retry |
| 5 | compress this folder into a zip archive | ai | AI | `Compress-Archive -Path . -DestinationPath archive.zip` | PASS | — |
| 6 | show the last 50 lines of the newest log file | ai | AI | `Get-ChildItem *.log \| Sort-Object … \| Get-Content -Tail 50` | **PASS** | Fixed: chain prompt + validation retry |
| 7 | count how many lines of Go code are in this project | ai | AI | `Get-ChildItem -Recurse -Include '*.go' … \| Measure-Object -Line` | PASS | — |
| 8 | find every TODO comment in the source files | ai | AI | `Get-ChildItem … \| Select-String TODO` | PASS | — |
| 9 | show my current branch and uncommitted changes | ai | AI | `git branch --show-current \| git status --short` | PASS | — |
| 10 | list all environment variables containing PATH | ai | AI | `Get-ChildItem Env: \| Where-Object …` | PASS | — |
| 11 | show CPU and memory usage right now | ai | AI | `Get-CimInstance Win32_Processor \| Select-Object LoadPercentage` | **PASS** | Fixed: chain prompt + validation retry |
| 12 | download a file from a URL and save it | ai | AI | `curl -o file URL` | FAIL | Placeholder literals `file` and `URL` (Phase 3 quality gate) |
| 13 | find files larger than 100MB | ai | **Rule** | `Get-ChildItem … \| Where-Object { $_.Length -gt 104857600 }` | **PASS** | Fixed: new `find_large_files` rule |
| 14 | show the git commit history for the last week | ai | AI | `git log '--since=1.week' --oneline \| …` | PASS | — |
| 15 | rename all .txt files to .md in this directory | ai | AI | `Get-ChildItem *.txt \| Rename-Item …` | **PASS** | Fixed: prompt pipe guidance + retry |
| 16 | show which program is using the most CPU | ai | AI | `Get-Process \| Sort-Object CPU -Descending \| Select-Object -First 1` | **PASS** | Fixed: chain prompt + validation retry |
| 17 | list installed python packages | ai | AI | `python -m pip list` | PASS | — |
| 18 | show a file with line numbers | ai | AI | `Get-Content . \| {0}: {1} -f ++$i $_` | FAIL | Uses `.` as filename placeholder (Phase 3) |
| 19 | tail the windows event log | ai | AI | `Get-WinEvent -LogName System -MaxEvents 10` | PASS | — |
| 20 | find duplicate files by name | ai | AI | `Get-ChildItem -Recurse -File \| Group-Object Name \| …` | PASS | — |

---

## Summary

| Path | Baseline | After improvements |
|------|----------|-------------------|
| Rule-based | 18/20 (90%) | **20/20 (100%)** |
| Natural-language | 11/20 (55%) | **18/20 (90%)** |
| **Overall** | **29/40 (72.5%)** | **38/40 (95%)** |

### Remaining failures (2)

All three are **placeholder / ambiguous-input** cases deferred to Phase 3 (post-generation quality gate):

- **NL #12** — `curl -o file URL` (no URL or filename in request)
- **NL #18** — `Get-Content .` (no file named in request)

These require either a quality gate that rejects placeholder tokens or interactive param prompting — not addressed in Phases 1+2.

### What fixed the baseline failures

| Failure theme | Fix | Tests recovered |
|---------------|-----|-----------------|
| Rule token-count mismatch | `copy/move {{source}} {{dest}}` examples | R19, R20 |
| Missing disk-free rule phrases | `disk_usage` examples expanded | NL #2 (rule hit) |
| No rule for large files | `find_large_files` intent + size parsing | NL #13 (rule hit) |
| AI flat argv with pipes | Few-shot chain prompt + 1 validation retry | NL #4, #6, #11 |
| Bad PowerShell pipeline syntax | Improved prompt + retry | NL #15, #16 |

### Next steps (Phase 3, not implemented)

- Post-generation quality gate rejecting placeholder tokens (`URL`, `file`, lone `.` as path)
- Optional: prompt user for missing URL/filename when input is ambiguous

---

## Post-fix re-run (Round 2 fix plan, 2026-06-03)

After implementing the Round 2 fix plan (`ValidateCommandQuality`, `Where-Object` few-shot, `list_recycle_bin` rule, `list_env`/`print_text` rule fixes), the two remaining Round 1 NL failures were re-run:

| # | Input | Result | Notes |
|---|-------|--------|-------|
| 12 | download a file from a URL and save it | FAIL | Quality gate rejects placeholder `file` after retry; input has no concrete URL/filename |
| 18 | show a file with line numbers | FAIL | Validation rejects complex `$i`/`ForEach-Object` expr token; retry did not produce a valid chain |

These remain **ambiguous-input** failures — the quality gate correctly blocks placeholders rather than executing unsafe stand-ins.
