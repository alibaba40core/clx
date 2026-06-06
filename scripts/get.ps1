#Requires -Version 5.1
# CLX download installer (Windows).
#
# Fetches prebuilt clx.exe, clx-ai.exe (internal worker), and clxmax.exe from GitHub
# Releases — no Go toolchain and no source checkout required. Downloads are verified
# against the published checksums.txt (SHA-256) before anything is installed, then the
# install dir is added to the user PATH.
#
# Usage:
#   irm https://raw.githubusercontent.com/alibaba40core/clx/main/scripts/get.ps1 | iex
#
# Environment overrides:
#   $env:CLX_VERSION      Release tag to install (default: latest), e.g. v1.0.2
#   $env:CLX_INSTALL_DIR  Install destination (default: %LOCALAPPDATA%\Programs\clx)
$ErrorActionPreference = "Stop"

$Repo = "alibaba40core/clx"
$Version = if ($env:CLX_VERSION) { $env:CLX_VERSION } else { "latest" }
$Dest = if ($env:CLX_INSTALL_DIR) { $env:CLX_INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA "Programs\clx" }

function Write-Log { param([string]$Message) Write-Host "clx-install: $Message" }

function Get-Arch {
    $arch = $env:PROCESSOR_ARCHITECTURE
    switch ($arch) {
        "AMD64" { return "amd64" }
        "ARM64" { return "arm64" }
        default { throw "unsupported architecture: $arch" }
    }
}

$arch = Get-Arch
$asset = "clx_windows_${arch}.zip"

if ($Version -eq "latest") {
    $base = "https://github.com/$Repo/releases/latest/download"
} else {
    $base = "https://github.com/$Repo/releases/download/$Version"
}

$tmp = Join-Path ([System.IO.Path]::GetTempPath()) ("clx-install-" + [System.Guid]::NewGuid().ToString("N"))
New-Item -ItemType Directory -Force -Path $tmp | Out-Null

try {
    $zipPath = Join-Path $tmp $asset
    $sumPath = Join-Path $tmp "checksums.txt"

    Write-Log "downloading $asset ($Version)"
    try {
        Invoke-WebRequest -Uri "$base/$asset" -OutFile $zipPath -UseBasicParsing
    } catch {
        throw "could not download $base/$asset (does the release exist for windows/$arch?)"
    }

    Write-Log "verifying checksum"
    Invoke-WebRequest -Uri "$base/checksums.txt" -OutFile $sumPath -UseBasicParsing

    $want = $null
    foreach ($line in Get-Content $sumPath) {
        $parts = $line -split '\s+', 2
        if ($parts.Count -eq 2) {
            $name = $parts[1].TrimStart('*').Trim()
            if ($name -eq $asset) { $want = $parts[0].Trim(); break }
        }
    }
    if (-not $want) { throw "no checksum entry for $asset" }

    $got = (Get-FileHash -Algorithm SHA256 -Path $zipPath).Hash.ToLower()
    if ($want.ToLower() -ne $got) {
        throw "checksum mismatch for $asset (want $want, got $got)"
    }

    Write-Log "extracting"
    Expand-Archive -Path $zipPath -DestinationPath $tmp -Force

    New-Item -ItemType Directory -Force -Path $Dest | Out-Null

    # Stop any running installed copies so we can overwrite them.
    foreach ($name in @("clx", "clxmax")) {
        Get-Process -Name $name -ErrorAction SilentlyContinue | ForEach-Object {
            if ($_.Path -and $_.Path.StartsWith($Dest, [StringComparison]::OrdinalIgnoreCase)) {
                Stop-Process -Id $_.Id -Force -ErrorAction SilentlyContinue
            }
        }
    }
    Start-Sleep -Milliseconds 300

    Copy-Item -Force (Join-Path $tmp "clx.exe") (Join-Path $Dest "clx.exe")
    if (Test-Path (Join-Path $tmp "clx-ai.exe")) {
        Copy-Item -Force (Join-Path $tmp "clx-ai.exe") (Join-Path $Dest "clx-ai.exe")
    }
    if (Test-Path (Join-Path $tmp "clxmax.exe")) {
        Copy-Item -Force (Join-Path $tmp "clxmax.exe") (Join-Path $Dest "clxmax.exe")
    }

    Write-Log "installed clx, clx-ai (internal), and clxmax to $Dest"

    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($userPath -notlike "*$Dest*") {
        $newPath = if ($userPath) { "$userPath;$Dest" } else { $Dest }
        [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
        $env:Path = "$env:Path;$Dest"
        Write-Log "added $Dest to user PATH (restart your shell to pick it up)"
    }

    & (Join-Path $Dest "clx.exe") --version
} finally {
    Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
}
