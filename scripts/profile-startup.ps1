# Profile CLX startup phases on Windows.
param(
    [int]$Runs = 5,
    [string]$Bin = ""
)

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
if ($Bin -eq "") {
    $Bin = Join-Path $Root "bin\clx.exe"
}
if (-not (Test-Path $Bin)) {
    Write-Host "Building release binary..."
    Push-Location $Root
    go run ./cmd/genrules
    go build -trimpath -tags=lite -ldflags="-s -w" -o $Bin ./cmd/clx
    Pop-Location
}

$ClxHome = Join-Path $env:TEMP "clx-profile-startup"
if (-not (Test-Path $ClxHome)) {
    New-Item -ItemType Directory -Path $ClxHome | Out-Null
}

function Measure-CLX {
    param([string[]]$Args)
    $sw = [System.Diagnostics.Stopwatch]::StartNew()
    & $Bin @Args 2>$null | Out-Null
    $sw.Stop()
    return $sw.ElapsedMilliseconds
}

function Report-Phase {
    param([string]$Label, [string[]]$Args)
    $samples = @()
    for ($i = 0; $i -lt $Runs; $i++) {
        $samples += Measure-CLX -Args $Args
    }
    $sorted = $samples | Sort-Object
    $mid = [int](($sorted.Count - 1) / 2)
    $median = $sorted[$mid]
    $worst = $sorted[-1]
    Write-Host ("{0,-28} median={1,4} ms  worst={2,4} ms  samples={3}" -f $Label, $median, $worst, ($samples -join ","))
}

Write-Host "CLX startup profile (CLX_HOME=$ClxHome, runs=$Runs)"
Write-Host ""

$env:CLX_HOME = $ClxHome
Report-Phase "--version" @("--version")
Report-Phase "rule hit (pwd explain)" @("--provider", "none", "--explain", "pwd")
Report-Phase "rule miss" @("--provider", "none", "--explain", "unknown xyz command")
Report-Phase "NL rule hit" @("--provider", "none", "--explain", "show me disk usage")

Write-Host ""
Write-Host "Budget targets: --version < 120 ms (Windows CI); rule hit < 200 ms"
