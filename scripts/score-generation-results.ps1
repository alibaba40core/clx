$ErrorActionPreference = "Stop"
$root = Split-Path $PSScriptRoot -Parent
$docDir = Join-Path $root "doc\development"

# Post-fix baselines: overall pass count per round (rule, nl, overall)
$baselines = @{
    1 = @{ rule = 20; nl = 18; overall = 38 }
    2 = @{ rule = 20; nl = 20; overall = 40 }
    3 = @{ rule = 20; nl = 20; overall = 40 }
    4 = @{ rule = 20; nl = 20; overall = 40 }
}

# R1 NL known acceptable failures (not regressions)
$expectedFails = @(
    @{ round = 1; suite = "nl"; n = 12 }
    @{ round = 1; suite = "nl"; n = 18 }
)

function Test-Pass($row) {
    $hasCmd = -not [string]::IsNullOrWhiteSpace($row.command)
    if ($row.exit -ne 0 -or -not $hasCmd) { return $false }
    if ($row.suite -eq "rule") {
        if ($row.source -ne "Rule") { return $false }
        if ($row.wantIntent) {
            $intent = [string]$row.intent
            if ($intent -eq "ai-generated command") { return $false }
            if ($intent -ne $row.wantIntent) { return $false }
        }
        return $true
    }
    # nl
    return $true
}

function Is-ExpectedFail($row) {
    foreach ($ef in $expectedFails) {
        if ($row.round -eq $ef.round -and $row.suite -eq $ef.suite -and $row.n -eq $ef.n) {
            return $true
        }
    }
    return $false
}

$allPath = Join-Path $docDir "gen-test-all-results.json"
if (-not (Test-Path $allPath)) {
    $merged = New-Object System.Collections.Generic.List[object]
    foreach ($r in 1..4) {
        $p = Join-Path $docDir "gen-test-round$r-results.json"
        if (Test-Path $p) {
            (Get-Content $p -Raw | ConvertFrom-Json) | ForEach-Object { $merged.Add($_) | Out-Null }
        }
    }
    $merged | ConvertTo-Json -Depth 4 | Set-Content -Encoding utf8 $allPath
}

$rows = Get-Content $allPath -Raw | ConvertFrom-Json
$scored = New-Object System.Collections.Generic.List[object]
$regressions = New-Object System.Collections.Generic.List[object]

foreach ($row in $rows) {
    $pass = Test-Pass $row
    $expectedFail = Is-ExpectedFail $row
    $status = if ($pass) { "pass" } elseif ($expectedFail) { "expected_fail" } else { "fail" }
    $isRegression = $false
    if ($status -eq "fail") {
        $isRegression = $true
        $regressions.Add([PSCustomObject]@{
            round = $row.round
            suite = $row.suite
            n = $row.n
            input = $row.input
            source = $row.source
            intent = $row.intent
            exit = $row.exit
            stderr = $row.stderr
            command = $row.command
        }) | Out-Null
    }
    $scored.Add([PSCustomObject]@{
        round = $row.round
        suite = $row.suite
        n = $row.n
        input = $row.input
        status = $status
        regression = $isRegression
        source = $row.source
        intent = $row.intent
        exit = $row.exit
        command = $row.command
        stderr = $row.stderr
    }) | Out-Null
}

$byRound = @{}
foreach ($r in 1..4) {
    $subset = $scored | Where-Object { $_.round -eq $r }
    $rulePass = ($subset | Where-Object { $_.suite -eq "rule" -and $_.status -eq "pass" }).Count
    $nlPass = ($subset | Where-Object { $_.suite -eq "nl" -and ($_.status -eq "pass" -or $_.status -eq "expected_fail") }).Count
    $overallPass = ($subset | Where-Object { $_.status -eq "pass" -or $_.status -eq "expected_fail" }).Count
    $bl = $baselines[$r]
    $byRound[$r] = [PSCustomObject]@{
        round = $r
        rulePass = $rulePass
        ruleBaseline = $bl.rule
        nlPass = $nlPass
        nlBaseline = $bl.nl
        overallPass = $overallPass
        overallBaseline = $bl.overall
        meetsBaseline = ($rulePass -ge $bl.rule) -and ($nlPass -ge $bl.nl) -and ($overallPass -ge $bl.overall)
    }
}

$totalPass = ($scored | Where-Object { $_.status -eq "pass" }).Count
$totalExpected = ($scored | Where-Object { $_.status -eq "expected_fail" }).Count
$totalOverall = ($scored | Where-Object { $_.status -eq "pass" -or $_.status -eq "expected_fail" }).Count
$unexpectedFail = ($scored | Where-Object { $_.status -eq "fail" }).Count

$meta = $null
$metaPath = Join-Path $docDir "gen-test-run-meta.json"
$runAt = (Get-Date).ToString("o")
$commit = ""
$clxVersion = ""
if (Test-Path $metaPath) {
    $meta = Get-Content $metaPath -Raw | ConvertFrom-Json
    if ($meta.runAt) { $runAt = [string]$meta.runAt }
    if ($meta.commit) { $commit = [string]$meta.commit }
    if ($meta.clxVersion) { $clxVersion = [string]$meta.clxVersion }
}

function Get-SpotRow($round, $suite, $n) {
    $hit = $scored | Where-Object { $_.round -eq $round -and $_.suite -eq $suite -and $_.n -eq $n } | Select-Object -First 1
    if (-not $hit) { return $null }
    return [PSCustomObject]@{
        label = "R$round $suite #$n"
        status = $hit.status
        source = $hit.source
        intent = $hit.intent
        command = $hit.command
        stderr = $hit.stderr
    }
}

$spotList = @(
    (Get-SpotRow 1 "nl" 2),
    (Get-SpotRow 2 "rule" 10),
    (Get-SpotRow 2 "rule" 11),
    (Get-SpotRow 2 "rule" 16),
    (Get-SpotRow 2 "rule" 17),
    (Get-SpotRow 2 "rule" 6),
    (Get-SpotRow 3 "rule" 1),
    (Get-SpotRow 4 "rule" 15)
) | Where-Object { $_ -ne $null }

$roundList = @(
    $byRound[[int]1],
    $byRound[[int]2],
    $byRound[[int]3],
    $byRound[[int]4]
)

$summary = [PSCustomObject]@{
    runAt = $runAt
    commit = $commit.TrimEnd(',')
    clxVersion = $clxVersion
    totalTests = [int]$scored.Count
    totalPass = [int]$totalPass
    totalExpectedFail = [int]$totalExpected
    totalOverall = [int]$totalOverall
    totalBaseline = 158
    unexpectedFailCount = [int]$unexpectedFail
    regressionCount = [int]$regressions.Count
    rounds = $roundList
    regressions = @($regressions.ToArray())
    spotChecks = $spotList
}

$summary | ConvertTo-Json -Depth 6 | Set-Content -Encoding utf8 (Join-Path $docDir "gen-test-score-summary.json")

Write-Host "Score summary: $totalOverall / $($scored.Count) (pass=$totalPass expected_fail=$totalExpected)"
Write-Host "Regressions: $($regressions.Count)"
foreach ($rd in $summary.rounds) {
    Write-Host "R$($rd.round): rule $($rd.rulePass)/$($rd.ruleBaseline) nl $($rd.nlPass)/$($rd.nlBaseline) overall $($rd.overallPass)/$($rd.overallBaseline) baseline_ok=$($rd.meetsBaseline)"
}
