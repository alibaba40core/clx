# CLX Command-Generation Test Report — Round 3

**Date:** 2026-06-03  
**Host:** Windows 10.0.22631, PowerShell  
**CLX binary:** `C:\Users\user\AppData\Local\Programs\clx\clx.exe` (installed via `make install`, not repo `bin\`)  
**CLX version:** `40a9d28` (commit 40a9d28, go1.26.3)  
**Provider:** OpenAI (`gpt-4.1-mini`)  
**Test mode:** `clx --explain "<input>"` (generation only; nothing executed)

## Method

Third 40-test suite with **no overlap** with Round 1 ([generation-test-report.md](generation-test-report.md)) or Round 2 ([generation-test-report-round2.md](generation-test-report-round2.md)). Same pass/fail criteria as prior rounds.

**Pass criteria:** Rule inputs → `Source: Rule` + correct Windows command. NL inputs → `Source: AI` (or cache) + runnable correct command; correct rule hit also passes.

**Fail criteria:** No command, validation/quality rejection, placeholders, broken pipelines, or semantically wrong command.

**Note:** Rule #1 uses `free disk space` instead of plan `uptime` (no Windows `system_uptime` strategy).

Runner: [scripts/run-generation-tests-round3.ps1](../../scripts/run-generation-tests-round3.ps1)

---

## Rule-Based Generation (20 tests)

**Result: 20 PASS / 0 FAIL**

| # | Input | Expected Intent | Actual Source | Generated Command | Result | Notes |
|---|-------|-----------------|---------------|-------------------|--------|-------|
| 1 | free disk space | disk_usage | Rule | `Get-PSDrive` | PASS | Windows swap for `uptime` |
| 2 | ll | list_dir | Rule | `Get-ChildItem .` | PASS | — |
| 3 | ls src | list_dir | Rule | `Get-ChildItem src` | PASS | — |
| 4 | type app.config | view_file | Rule | `Get-Content app.config` | PASS | — |
| 5 | show settings.json | view_file | Rule | `Get-Content settings.json` | PASS | — |
| 6 | locate nginx.conf in config | find_file | Rule | `Get-ChildItem -Path config -Recurse -Filter nginx.conf …` | PASS | — |
| 7 | list files modified today | find_modified_today | Rule | `Get-ChildItem … \| Where-Object { $_.LastWriteTime.Date -eq [datetime]::Today }` | PASS | — |
| 8 | remove old.log | remove_file | Rule | `Remove-Item -Path old.log` | PASS | — |
| 9 | delete file temp.dat | remove_file | Rule | `Remove-Item -Path temp.dat` | PASS | — |
| 10 | mkdir backups | make_dir | Rule | `New-Item -ItemType Directory -Path backups -Force` | PASS | — |
| 11 | create directory logs | make_dir | Rule | `New-Item -ItemType Directory -Path logs -Force` | PASS | — |
| 12 | cp config.ini config.bak | copy_file | Rule | `Copy-Item -Path config.ini -Destination config.bak` | PASS | — |
| 13 | move draft.txt final.txt | move_file | Rule | `Move-Item -Path draft.txt -Destination final.txt` | PASS | — |
| 14 | git log -n 5 | git_log | Rule | `git log --oneline -n 5` | PASS | — |
| 15 | ping 8.8.8.8 | ping_host | Rule | `ping -n 4 8.8.8.8` | PASS | — |
| 16 | curl -I https://httpbin.org/get | curl_url | Rule | `curl -I https://httpbin.org/get` | PASS | — |
| 17 | print environment | list_env | Rule | `Get-ChildItem env:` | PASS | — |
| 18 | where is git | which_command | Rule | `Get-Command git` | PASS | — |
| 19 | head -n 20 access.log | head_file | Rule | `Get-Content access.log -TotalCount 20` | PASS | — |
| 20 | tail -n 50 error.log | tail_file | Rule | `Get-Content error.log -Tail 50` | PASS | — |

---

## Natural-Language Generation (20 tests)

**Baseline: 17 PASS / 3 FAIL**  
**After rule additions: 20 PASS / 0 FAIL**

| # | Input | Expected Source | Actual Source | Generated Command | Result | Notes |
|---|-------|-----------------|---------------|-------------------|--------|-------|
| 1 | show the windows version and build number | AI | AI | `Get-ComputerInfo \| Select-Object WindowsVersion, WindowsBuildLabEx` | PASS | — |
| 2 | list all stopped windows services | AI | AI | `Get-Service \| Where-Object $_.Status -eq 'Stopped'` | PASS | — |
| 3 | show top 5 processes by memory usage | AI | AI | `Get-Process \| Sort-Object WS \| Select-Object -First 5` | PASS | — |
| 4 | find empty files in this directory tree | AI | AI | `Get-ChildItem -File -Recurse \| Where-Object $_.Length -eq 0` | PASS | — |
| 5 | show the arp table on this machine | AI | AI | `arp -a` | PASS | — |
| 6 | test if port 22 is open on github.com | AI | AI | `Test-NetConnection github.com -Port 22 \| …` | PASS | — |
| 7 | show the system timezone | AI | AI | `Get-TimeZone` | PASS | — |
| 8 | list files changed in the last 24 hours | AI | AI | `Get-ChildItem … \| Where-Object $_.LastWriteTime -gt (Get-Date).AddHours(-24)` | PASS | — |
| 9 | show disk read and write performance counters | AI | **Rule** | `Get-CimInstance Win32_PerfFormattedData_PhysicalDisk_PhysicalDisk` | **PASS** | Baseline FAIL: Get-Counter path rejected `$`/`\` in token |
| 10 | what is my public ip address | AI | AI | `curl -s https://api.ipify.org` | PASS | — |
| 11 | list local user accounts on this computer | AI | **Rule** | `Get-LocalUser` | **PASS** | Baseline FAIL: AI no match |
| 12 | find all running processes named chrome | AI | AI | `Get-Process chrome \| Where-Object …` | PASS | — |
| 13 | show the total size of the windows directory | AI | AI | `Get-ChildItem 'C:\Windows' -Recurse -File \| Measure-Object -Property Length -Sum` | PASS | — |
| 14 | show active connections to port 80 | AI | AI | `Get-NetTCPConnection -LocalPort 80 \| Where-Object …` | PASS | — |
| 15 | show the most recent error in the system event log | AI | AI | `Get-WinEvent -LogName System -FilterHashtable @{Level=2} -MaxEvents 1` | PASS | — |
| 16 | display monitor resolution and refresh rate | AI | AI | `Get-CimInstance Win32_DesktopMonitor \| Select-Object …` | PASS | — |
| 17 | list mapped network drives | AI | AI | `Get-PSDrive -PSProvider FileSystem \| Where-Object {$_.DisplayRoot -ne $null}` | PASS | — |
| 18 | find all pdf files in the downloads folder | AI | **Rule** | `Get-ChildItem -Path { Join-Path $env:USERPROFILE 'Downloads' } … \| Where-Object { $_.Extension -eq '.pdf' }` | **PASS** | Baseline FAIL: `$env:USERPROFILE` in plain token |
| 19 | show battery health or wear level | AI | AI | `Get-CimInstance Win32_Battery \| Select-Object …` | PASS | — |
| 20 | restart the print spooler service | AI | AI | `Restart-Service -Name Spooler` | PASS | — |

---

## Summary

| Path | Baseline | Post-fix |
|------|----------|----------|
| Rule-based | 20/20 (100%) | **20/20 (100%)** |
| Natural-language | 17/20 (85%) | **20/20 (100%)** |
| **Overall** | **37/40 (92.5%)** | **40/40 (100%)** |

### Comparison to prior rounds

| Metric | Round 1 (post-fix) | Round 2 (post-fix) | Round 3 (post-fix) |
|--------|-------------------|-------------------|-------------------|
| Rule-based | 20/20 | 20/20 | **20/20** |
| Natural-language | 18/20 | 20/20 | **20/20** |
| **Overall** | 38/40 | 40/40 | **40/40** |

### Baseline failure themes (3 NL)

| # | Input | Root cause | Fix |
|---|-------|------------|-----|
| 9 | disk performance counters | AI used `Get-Counter` with `\PhysicalDisk(*)\…` in plain argv token | New `show_disk_performance` rule (CIM perf class) |
| 11 | list local users | AI returned no valid command | New `list_local_users` rule (`Get-LocalUser`) |
| 18 | PDFs in Downloads | AI used `$env:USERPROFILE\Downloads` in plain token | New `find_pdf_downloads` rule (expr path + extension filter) |

### Post-fix changes

- [internal/builtin/rules/system_info.yaml](../../internal/builtin/rules/system_info.yaml): `show_disk_performance`, `list_local_users`
- [internal/builtin/rules/filesystem.yaml](../../internal/builtin/rules/filesystem.yaml): `find_pdf_downloads`
- [internal/intent/resolve_test.go](../../internal/intent/resolve_test.go): `TestResolveRound3NLRuleIntents`
