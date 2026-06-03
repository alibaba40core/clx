$ErrorActionPreference = "Continue"
$clx = Join-Path $env:LOCALAPPDATA "Programs\clx\clx.exe"
$outFile = Join-Path $PSScriptRoot "..\doc\development\gen-test-round4-results.json"

$ruleTests = @(
    @{ n=1; phrase="what time is it"; wantIntent="current_date" },
    @{ n=2; phrase="current user"; wantIntent="current_user" },
    @{ n=3; phrase="show current directory"; wantIntent="current_dir" },
    @{ n=4; phrase="echo Round4"; wantIntent="print_text" },
    @{ n=5; phrase="files changed today"; wantIntent="find_modified_today" },
    @{ n=6; phrase="find file pom.xml"; wantIntent="find_file" },
    @{ n=7; phrase="locate settings.ini in etc"; wantIntent="find_file" },
    @{ n=8; phrase="grep WARNING system.log"; wantIntent="search_text_in_file" },
    @{ n=9; phrase="cat package.json"; wantIntent="view_file" },
    @{ n=10; phrase="touch deploy.marker"; wantIntent="create_empty_file" },
    @{ n=11; phrase="remove directory staging"; wantIntent="remove_dir" },
    @{ n=12; phrase="copy backup.zip to archive.zip"; wantIntent="copy_file" },
    @{ n=13; phrase="move inbox.csv to processed.csv"; wantIntent="move_file" },
    @{ n=14; phrase="files larger than 50MB"; wantIntent="find_large_files" },
    @{ n=15; phrase="du -sh data"; wantIntent="disk_usage" },
    @{ n=16; phrase="docker logs api --tail 50"; wantIntent="docker_logs" },
    @{ n=17; phrase="tracert cloudflare.com"; wantIntent="traceroute_host" },
    @{ n=18; phrase="ping localhost"; wantIntent="ping_host" },
    @{ n=19; phrase="git diff src/main.go"; wantIntent="git_diff_path" },
    @{ n=20; phrase="list directory tmp"; wantIntent="list_dir" }
)

$nlTests = @(
    @{ n=1; phrase="show installed windows hotfixes" },
    @{ n=2; phrase="list processes listening on port 8080" },
    @{ n=3; phrase="stop all notepad processes" },
    @{ n=4; phrase="show dns server addresses for active adapters" },
    @{ n=5; phrase="list enabled optional windows features" },
    @{ n=6; phrase="show free space on the C drive" },
    @{ n=7; phrase="find files modified in the last seven days" },
    @{ n=8; phrase="show the powershell execution policy" },
    @{ n=9; phrase="list all scheduled tasks on this system" },
    @{ n=10; phrase="show ip configuration for all network adapters" },
    @{ n=11; phrase="clear the arp cache on this computer" },
    @{ n=12; phrase="show the path to the temp directory" },
    @{ n=13; phrase="list services set to start automatically" },
    @{ n=14; phrase="show cpu model and number of cores" },
    @{ n=15; phrase="find hidden files in the current folder" },
    @{ n=16; phrase="show the last 20 entries in the security event log" },
    @{ n=17; phrase="check whether hyper-v is enabled on this pc" },
    @{ n=18; phrase="show the installed dotnet runtime version" },
    @{ n=19; phrase="list exe files in the program files folder" },
    @{ n=20; phrase="start the windows time service" }
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
        wantIntent = $wantIntent
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
