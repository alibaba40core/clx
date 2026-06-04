$ErrorActionPreference = "Continue"
$root = Split-Path $PSScriptRoot -Parent
$docDir = Join-Path $root "doc\development"

Push-Location $root
try {
    Write-Host "=== make install ==="
    & make install 2>&1 | Write-Host
    $verLine = & (Join-Path $env:LOCALAPPDATA "Programs\clx\clx.exe") --version 2>&1 | Out-String
    $commit = ""
    if ($verLine -match "commit\s+([0-9a-f]+)") { $commit = $Matches[1] }
    $meta = [PSCustomObject]@{
        runAt = (Get-Date).ToString("o")
        clxVersion = $verLine.Trim()
        commit = $commit
        host = [Environment]::OSVersion.VersionString
    }
    $meta | ConvertTo-Json | Set-Content -Encoding utf8 (Join-Path $docDir "gen-test-run-meta.json")

    $all = New-Object System.Collections.Generic.List[object]
    foreach ($r in 1..4) {
        Write-Host "`n=== Round $r ==="
        $script = Join-Path $PSScriptRoot "run-generation-tests-round$r.ps1"
        & powershell -ExecutionPolicy Bypass -File $script 2>&1 | Write-Host
        $json = Join-Path $docDir "gen-test-round$r-results.json"
        if (Test-Path $json) {
            $data = Get-Content $json -Raw | ConvertFrom-Json
            foreach ($item in $data) { $all.Add($item) | Out-Null }
        }
    }
    $all | ConvertTo-Json -Depth 4 | Set-Content -Encoding utf8 (Join-Path $docDir "gen-test-all-results.json")
    Write-Host "`nCombined: $(Join-Path $docDir 'gen-test-all-results.json') ($($all.Count) rows)"
}
finally {
    Pop-Location
}
