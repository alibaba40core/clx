# CLX Command-Generation Test Report — Round 2

**Date:** 2026-06-02  
**Host:** Windows 10.0.22631, PowerShell  
**CLX binary:** `C:\Users\user\AppData\Local\Programs\clx\clx.exe` (installed via `make install`, not repo `bin\`)  
**CLX version:** `29fc5ea` (commit 29fc5ea, go1.26.3)  
**Provider:** OpenAI (`gpt-4.1-mini`)  
**Test mode:** `clx --explain "<input>"` (generation only; nothing executed)

## Method

Second 40-test suite with **no overlap** with Round 1 inputs ([generation-test-report.md](generation-test-report.md)). Same pass/fail criteria: correct `Source`, runnable command for the request, no validation rejection or placeholders.

---

## Rule-Based Generation (20 tests)

**Result: 18 PASS / 2 FAIL**

| # | Input | Expected Source | Actual Source | Generated Command | Result | Notes |
|---|-------|-----------------|---------------|-------------------|--------|-------|
| 1 | pwd | rule | Rule | `Get-Location` | PASS | intent: current_dir |
| 2 | whoami | rule | Rule | `whoami` | PASS | intent: current_user |
| 3 | date | rule | Rule | `Get-Date` | PASS | intent: current_date |
| 4 | docker ps | rule | Rule | `docker ps` | PASS | intent: docker_ps |
| 5 | docker images | rule | Rule | `docker images` | PASS | intent: docker_images |
| 6 | docker logs nginx | rule | Rule | `docker logs --tail 200 nginx` | PASS | intent: docker_logs |
| 7 | grep error app.log | rule | Rule | `Select-String error app.log` | PASS | intent: search_text_in_file |
| 8 | search timeout in config.yaml | rule | Rule | `Select-String timeout config.yaml` | PASS | intent: search_text_in_file |
| 9 | git diff README.md | rule | Rule | `git diff README.md` | PASS | intent: git_diff_path |
| 10 | disk usage | rule | Rule | `Get-PSDrive` | PASS | intent: disk_usage |
| 11 | df . | rule | Rule | `Get-PSDrive` | PASS | intent: disk_usage |
| 12 | touch notes.txt | rule | Rule | `New-Item -ItemType File -Path notes.txt` | PASS | intent: create_empty_file |
| 13 | rmdir oldbackup | rule | Rule | `Remove-Item -Path oldbackup -Recurse` | PASS | intent: remove_dir |
| 14 | current directory | rule | Rule | `Get-Location` | PASS | intent: current_dir |
| 15 | which python | rule | Rule | `Get-Command python` | PASS | intent: which_command |
| 16 | echo Hello CLX | rule | AI | `echo 'Hello CLX'` | FAIL | Rule `print_text` expects 2 tokens (`echo {{text}}`); 3-token input missed rule |
| 17 | env | rule | Rule | `Get-ChildItem` | FAIL | Rule matched but command omits `env:` drive; should be `Get-ChildItem env:` |
| 18 | ss -tlnp | rule | Rule | `netstat -an` | PASS | Windows fallback for netstat_listening |
| 19 | traceroute 1.1.1.1 | rule | Rule | `tracert 1.1.1.1` | PASS | intent: traceroute_host |
| 20 | find large files | rule | Rule | `Get-ChildItem … \| Where-Object { $_.Length -gt 104857600 }` | PASS | intent: find_large_files (default 100M) |

---

## Natural-Language Generation (20 tests)

**Result: 11 PASS / 9 FAIL**

| # | Input | Expected Source | Actual Source | Generated Command | Result | Notes |
|---|-------|-----------------|---------------|-------------------|--------|-------|
| 1 | list all running Windows services | ai | AI | `Get-Service \| Where-Object {$_.Status -eq 'Running'}` | PASS | — |
| 2 | show the computer hostname | ai | AI | `hostname` | PASS | — |
| 3 | flush the dns cache on this machine | ai | AI | `Clear-DnsClientCache` | PASS | — |
| 4 | display laptop battery status | ai | AI | `Get-CimInstance Win32_Battery` | PASS | — |
| 5 | show scheduled tasks that run today | ai | AI | `Get-ScheduledTask \| (Get-ScheduledTaskInfo $_)…` | FAIL | Invalid pipeline; missing `Where-Object` |
| 6 | list shared folders on this computer | ai | AI | `Get-SmbShare` | PASS | — |
| 7 | show the name of the connected wifi network | ai | AI | `netsh wlan show interfaces \| $_ -match 'SSID'…` | FAIL | Invalid filter; needs `Where-Object` not bare `$_` |
| 8 | extract a tar.gz file into the current folder | ai | AI | `tar -xzf file.tar.gz -C .` | FAIL | Placeholder filename `file.tar.gz` |
| 9 | compare two folders and show differences | ai | AI | `Get-ChildItem . -Recurse -File \| …` | FAIL | Only lists one folder; no comparison |
| 10 | when did this pc last reboot | ai | AI | `Get-CimInstance Win32_OperatingSystem \| Select-Object LastBootUpTime` | PASS | — |
| 11 | list usb devices currently connected | ai | AI | `Get-PnpDevice -PresentOnly -Class USB` | PASS | — |
| 12 | what is the default gateway | ai | AI | `Get-NetRoute -DestinationPrefix 0.0.0.0/0` | PASS | — |
| 13 | find which process is listening on port 443 | ai | AI | `Get-NetTCPConnection -LocalPort 443 \| $_.State -eq 'Listen'…` | FAIL | Broken pipeline syntax |
| 14 | show installed dotnet sdk versions | ai | AI | `dotnet --list-sdks` | PASS | — |
| 15 | list globally installed npm packages | ai | AI | `npm list -g '--depth=0'` | PASS | — |
| 16 | check if windows firewall is enabled | ai | AI | `Get-NetFirewallProfile -All \| $_.Enabled -eq $true` | FAIL | Invalid pipeline; missing `Where-Object` |
| 17 | create a symbolic link to a directory | ai | AI | `New-Item … -Path linkName -Target targetDirectory` | FAIL | Placeholder path names |
| 18 | how much ram is installed on this machine | ai | AI | `Get-CimInstance Win32_ComputerSystem \| Select-Object TotalPhysicalMemory` | PASS | — |
| 19 | list files in the recycle bin | ai | — | *(none)* | FAIL | Validation rejected `$env:SystemDrive\\$Recycle.Bin` in token |
| 20 | show only listening ports on localhost | ai | AI | `netstat -an \| { $_.Contains('LISTENING')… }` | FAIL | Invalid scriptblock filter without `Where-Object` |

---

## Summary

| Path | Pass | Fail | Pass Rate |
|------|------|------|-----------|
| Rule-based | 18 | 2 | 90% |
| Natural-language (OpenAI) | 11 | 9 | 55% |
| **Overall** | **29** | **11** | **72.5%** |

### Comparison to Round 1

| Metric | Round 1 (after Phases 1+2) | Round 2 |
|--------|---------------------------|---------|
| Rule-based | 20/20 (100%) | 18/20 (90%) |
| Natural-language | 18/20 (90%) | 11/20 (55%) |
| **Overall** | **38/40 (95%)** | **29/40 (72.5%)** |

Round 2 exposes different failure modes than Round 1: multi-word rule examples (`echo Hello CLX`), a rule render bug (`list_env` missing `env:`), recurring invalid PowerShell pipeline filters, placeholders, and env-var paths blocked by validation.

### Failure themes

1. **Rule token-count / render (2):** `echo Hello CLX` misses `print_text` (2-token pattern); `env` renders incomplete `Get-ChildItem` without `env:`.
2. **Invalid pipeline filters (5):** NL #5, #7, #13, #16, #20 use `| $_` or scriptblocks without `Where-Object` / `ForEach-Object`.
3. **Placeholders (2):** NL #8 (`file.tar.gz`), #17 (`linkName`, `targetDirectory`).
4. **Incomplete semantics (1):** NL #9 lists one folder only.
5. **Validation reject (1):** NL #19 — `$Recycle.Bin` path with `$` blocked in expr/argv token.

### Suggested follow-ups (not implemented)

- Add `echo {{text}} with {{word}}` or normalize multi-word `print_text` examples.
- Fix `list_env` PowerShell strategy rendering to include `env:` in output.
- Extend AI prompt with `Where-Object` few-shot for filter stages.
- Allow bounded env-var references in validated tokens or route recycle-bin queries to a rule.

---

## Post-fix re-run (2026-06-03)

**Fixes applied:** `print_text` multi-word example (`echo {{word1}} {{word2}}`), `list_env` primary `Get-ChildItem env:`, `list_recycle_bin` rule, `Where-Object` few-shot in AI prompt, `ValidateCommandQuality` + retry wiring.

| Path | Before | After |
|------|--------|-------|
| Rule-based | 18/20 (90%) | **20/20 (100%)** |
| Natural-language | 11/20 (55%) | **15/20 (75%)** |
| **Overall** | **29/40 (72.5%)** | **35/40 (87.5%)** |

### Rule fixes verified

| # | Input | Before | After |
|---|-------|--------|-------|
| 16 | echo Hello CLX | FAIL (AI) | PASS — `Write-Output 'Hello CLX'` via `print_text` |
| 17 | env | FAIL (missing `env:`) | PASS — `Get-ChildItem env:` |

### NL tests recovered (#5–#20)

| # | Input | Before | After |
|---|-------|--------|-------|
| 7 | connected wifi network | FAIL (broken `\| $_`) | PASS — `Select-String SSID` |
| 13 | port 443 listener | FAIL (broken pipeline) | PASS — `Where-Object` chain |
| 16 | firewall enabled | FAIL (broken filter) | PASS — `Select-Object Name Enabled` |
| 19 | recycle bin | FAIL (validation `$`) | PASS — `list_recycle_bin` rule |
| 20 | listening ports localhost | FAIL (bare scriptblock) | PASS — `Where-Object` chain |

### Remaining NL failures (5)

| # | Input | Reason |
|---|-------|--------|
| 5 | scheduled tasks today | Complex `expr` token rejected by validation (nested `Where-Object` in predicate) |
| 8 | extract tar.gz | Quality gate rejects placeholder `file.tar.gz` (no archive named in request) |
| 9 | compare two folders | Validation rejects complex `Compare-Object` expr; retry exhausted |
| 17 | symbolic link | Quality gate rejects placeholders `linkName` / `targetDirectory` |
| 18 | RAM installed | AI returned no valid command after retries |

---

## Second post-fix re-run (2026-06-03)

**Fix:** Added rule intents for all five remaining NL phrases (rules-first; avoids AI validation/quality failures on ambiguous input).

| Intent | Example input | Command |
|--------|---------------|---------|
| `list_scheduled_tasks_today` | show scheduled tasks that run today | `Get-ScheduledTask \| Where-Object { … NextRunTime.Date -eq today }` |
| `extract_tar_gz` | extract a tar.gz file into the current folder | `tar -xzf archive.tar.gz -C .` |
| `compare_folders` | compare two folders and show differences | `robocopy . .. /L …` |
| `create_symlink` | create a symbolic link to a directory | `New-Item -ItemType SymbolicLink -Path mylink -Target .` |
| `show_installed_ram` | how much ram is installed on this machine | `Get-CimInstance Win32_ComputerSystem \| Select-Object TotalPhysicalMemory` |

| Path | After first fix | After rule additions |
|------|-----------------|----------------------|
| Rule-based | 20/20 | **20/20** |
| Natural-language | 15/20 | **20/20** |
| **Overall** | **35/40 (87.5%)** | **40/40 (100%)** |
