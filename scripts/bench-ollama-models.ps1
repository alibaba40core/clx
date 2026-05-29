# Benchmark Ollama models for CLX NL intent resolution (same query each run).
param(
    [string]$Clx = (Join-Path $PSScriptRoot "..\bin\clx.exe"),
    [string]$Query = "show current directory"
)

$models = @("gemma3:270m", "qwen3:1.7b", "qwen3:4b")
$configPath = Join-Path $env:USERPROFILE ".clx\config.yaml"
$backup = Get-Content $configPath -Raw

function Set-Model($name) {
    $c = $backup -replace '(?m)^model:.*', "model: $name" `
        -replace '(?m)^(\s+model: ).*', "`${1}$name"
    Set-Content -Path $configPath -Value $c -NoNewline
}

Write-Host "Query: $Query"
Write-Host ("{0,-16} {1,8} {2}" -f "Model", "Sec", "Result")
Write-Host ("-" * 60)

foreach ($m in $models) {
    Set-Model $m
    $sw = [System.Diagnostics.Stopwatch]::StartNew()
    $out = & $Clx --explain $Query 2>&1 | Out-String
    $sw.Stop()
    $sec = [math]::Round($sw.Elapsed.TotalSeconds, 1)
    if ($LASTEXITCODE -eq 0 -and $out -match "Intent:\s+(\S+)") {
        $intent = $Matches[1]
        $src = if ($out -match "Source:\s+(\S+)") { $Matches[1] } else { "?" }
        Write-Host ("{0,-16} {1,8} {2} ({3})" -f $m, $sec, $intent, $src)
    } else {
        $err = ($out -split "`n" | Where-Object { $_ -match "error|rejected|timeout|untrusted|translate" } | Select-Object -First 1)
        if (-not $err) { $err = "exit $LASTEXITCODE" }
        Write-Host ("{0,-16} {1,8} FAIL: {2}" -f $m, $sec, $err.Trim())
    }
}

Set-Content -Path $configPath -Value $backup -NoNewline
Write-Host "`nConfig restored."
