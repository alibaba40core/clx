# Shared helpers for CLX generation test runners.
$script:GenTestClx = Join-Path $env:LOCALAPPDATA "Programs\clx\clx.exe"

function Run-Explain([string]$phrase) {
    $tmpOut = [System.IO.Path]::GetTempFileName()
    $tmpErr = [System.IO.Path]::GetTempFileName()
    $p = Start-Process -FilePath $script:GenTestClx -ArgumentList "--explain", $phrase -NoNewWindow -PassThru -Wait -RedirectStandardOutput $tmpOut -RedirectStandardError $tmpErr
    $out = [IO.File]::ReadAllText($tmpOut)
    $err = [IO.File]::ReadAllText($tmpErr)
    Remove-Item $tmpOut, $tmpErr -Force -ErrorAction SilentlyContinue
    return @{ out = $out; err = $err; exit = $p.ExitCode }
}

function Parse-Explain($round, $suite, $n, $phrase, $wantIntent, $r) {
    $out = $r.out
    $err = $r.err
    $intent = ""
    if ($out -match "(?m)^Intent:\s+(.+)$") { $intent = [string]$Matches[1].Trim() }
    $source = ""
    if ($out -match "(?m)^Source:\s+(\S+)") { $source = [string]$Matches[1] }
    $cmd = ""
    if ($out -match "(?m)^Command:\s+(.+)$") { $cmd = [string]$Matches[1].Trim() }
    [PSCustomObject]@{
        round = $round
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

function Invoke-GenerationTestSuite {
    param(
        [int]$Round,
        [string]$OutFile,
        [array]$RuleTests,
        [array]$NlTests
    )
    $results = New-Object System.Collections.Generic.List[object]
    foreach ($t in $RuleTests) {
        Write-Host "R${Round} rule $($t.n): $($t.phrase)"
        $r = Run-Explain $t.phrase
        $wi = $null
        if ($t.wantIntent) { $wi = $t.wantIntent }
        $results.Add((Parse-Explain $Round "rule" $t.n $t.phrase $wi $r)) | Out-Null
        $results | ConvertTo-Json -Depth 4 | Set-Content -Encoding utf8 $OutFile
    }
    foreach ($t in $NlTests) {
        Write-Host "R${Round} nl $($t.n): $($t.phrase)"
        $r = Run-Explain $t.phrase
        $results.Add((Parse-Explain $Round "nl" $t.n $t.phrase $null $r)) | Out-Null
        $results | ConvertTo-Json -Depth 4 | Set-Content -Encoding utf8 $OutFile
    }
    Write-Host "Done R${Round}. Results: $OutFile"
    return $results
}
