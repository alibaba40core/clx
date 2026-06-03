$ErrorActionPreference = "Continue"
$clx = Join-Path $env:LOCALAPPDATA "Programs\clx\clx.exe"
$outFile = Join-Path $PSScriptRoot "..\doc\development\gen-test-round3-results.json"

$ruleTests = @(
    @{ n=1; phrase="free disk space"; wantIntent="disk_usage" },
    @{ n=2; phrase="ll"; wantIntent="list_dir" },
    @{ n=3; phrase="ls src"; wantIntent="list_dir" },
    @{ n=4; phrase="type app.config"; wantIntent="view_file" },
    @{ n=5; phrase="show settings.json"; wantIntent="view_file" },
    @{ n=6; phrase="locate nginx.conf in config"; wantIntent="find_file" },
    @{ n=7; phrase="list files modified today"; wantIntent="find_modified_today" },
    @{ n=8; phrase="remove old.log"; wantIntent="remove_file" },
    @{ n=9; phrase="delete file temp.dat"; wantIntent="remove_file" },
    @{ n=10; phrase="mkdir backups"; wantIntent="make_dir" },
    @{ n=11; phrase="create directory logs"; wantIntent="make_dir" },
    @{ n=12; phrase="cp config.ini config.bak"; wantIntent="copy_file" },
    @{ n=13; phrase="move draft.txt final.txt"; wantIntent="move_file" },
    @{ n=14; phrase="git log -n 5"; wantIntent="git_log" },
    @{ n=15; phrase="ping 8.8.8.8"; wantIntent="ping_host" },
    @{ n=16; phrase="curl -I https://httpbin.org/get"; wantIntent="curl_url" },
    @{ n=17; phrase="print environment"; wantIntent="list_env" },
    @{ n=18; phrase="where is git"; wantIntent="which_command" },
    @{ n=19; phrase="head -n 20 access.log"; wantIntent="head_file" },
    @{ n=20; phrase="tail -n 50 error.log"; wantIntent="tail_file" }
)

$nlTests = @(
    @{ n=1; phrase="show the windows version and build number" },
    @{ n=2; phrase="list all stopped windows services" },
    @{ n=3; phrase="show top 5 processes by memory usage" },
    @{ n=4; phrase="find empty files in this directory tree" },
    @{ n=5; phrase="show the arp table on this machine" },
    @{ n=6; phrase="test if port 22 is open on github.com" },
    @{ n=7; phrase="show the system timezone" },
    @{ n=8; phrase="list files changed in the last 24 hours" },
    @{ n=9; phrase="show disk read and write performance counters" },
    @{ n=10; phrase="what is my public ip address" },
    @{ n=11; phrase="list local user accounts on this computer" },
    @{ n=12; phrase="find all running processes named chrome" },
    @{ n=13; phrase="show the total size of the windows directory" },
    @{ n=14; phrase="show active connections to port 80" },
    @{ n=15; phrase="show the most recent error in the system event log" },
    @{ n=16; phrase="display monitor resolution and refresh rate" },
    @{ n=17; phrase="list mapped network drives" },
    @{ n=18; phrase="find all pdf files in the downloads folder" },
    @{ n=19; phrase="show battery health or wear level" },
    @{ n=20; phrase="restart the print spooler service" }
)

function Run-Explain([string]$phrase) {
    $tmpOut = [System.IO.Path]::GetTempFileName()
    $tmpErr = [System.IO.Path]::GetTempFileName()
    $p = Start-Process -FilePath $clx -ArgumentList "--explain", $phrase -NoNewWindow -PassThru -Wait -RedirectStandardOutput $tmpOut -RedirectStandardError $tmpErr
    $out = [IO.File]::ReadAllText($tmpOut)
    $err = [IO.File]::ReadAllText($tmpErr)
    Remove-Item $tmpOut, $tmpErr -Force -ErrorAction SilentlyContinue
    return @{ out = $out; err = $err; exit = $p.ExitCode }
}

function Parse-Explain($suite, $n, $phrase, $wantIntent, $r) {
    $out = $r.out
    $err = $r.err
    $intent = ""
    if ($out -match "(?m)^Intent:\s+(.+)$") { $intent = [string]$Matches[1].Trim() }
    $source = ""
    if ($out -match "(?m)^Source:\s+(\S+)") { $source = [string]$Matches[1] }
    $cmd = ""
    if ($out -match "(?m)^Command:\s+(.+)$") { $cmd = [string]$Matches[1].Trim() }
    [PSCustomObject]@{
        suite = $suite
        n = $n
        input = $phrase
        intent = $intent
        source = $source
        command = $cmd
        exit = $r.exit
        stderr = ($err -replace '\s+', ' ').Trim()
    }
}

$results = New-Object System.Collections.Generic.List[object]
foreach ($t in $ruleTests) {
    Write-Host "rule $($t.n): $($t.phrase)"
    $r = Run-Explain $t.phrase
    $results.Add((Parse-Explain "rule" $t.n $t.phrase $t.wantIntent $r)) | Out-Null
    $results | ConvertTo-Json -Depth 4 | Set-Content -Encoding utf8 $outFile
}
foreach ($t in $nlTests) {
    Write-Host "nl $($t.n): $($t.phrase)"
    $r = Run-Explain $t.phrase
    $results.Add((Parse-Explain "nl" $t.n $t.phrase $null $r)) | Out-Null
    $results | ConvertTo-Json -Depth 4 | Set-Content -Encoding utf8 $outFile
}

Write-Host "Done. Results: $outFile"
