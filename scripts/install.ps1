#Requires -Version 5.1
$ErrorActionPreference = "Stop"

$Root = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$BinDir = Join-Path $Root "bin"
$Dest = Join-Path $env:LOCALAPPDATA "Programs\clx"

function Require-Go {
    if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
        Write-Error "go is not on PATH (need Go 1.26+)"
    }
}

function Get-LdFlags {
    $version = "dev"
    $commit = "unknown"
    try { $version = git -C $Root describe --tags --always 2>$null } catch {}
    try { $commit = git -C $Root rev-parse --short HEAD 2>$null } catch {}
    return "-s -w -X github.com/alibaba40core/clx/internal/cliversion.Version=$version -X github.com/alibaba40core/clx/internal/cliversion.Commit=$commit"
}

function Build-Binaries {
    New-Item -ItemType Directory -Force -Path $BinDir | Out-Null
    $ldflags = Get-LdFlags
    Push-Location $Root
    try {
        go run ./cmd/genrules
        go build -trimpath -tags lite -ldflags $ldflags -o (Join-Path $BinDir "clx.exe") ./cmd/clx
        go build -trimpath -ldflags $ldflags -o (Join-Path $BinDir "clx-ai.exe") ./cmd/clx-ai
        go build -trimpath -ldflags $ldflags -o (Join-Path $BinDir "clxmax.exe") ./cmd/clxmax
    } finally {
        Pop-Location
    }
}

function Stop-DestProcesses {
    foreach ($name in @("clx", "clxmax")) {
        Get-Process -Name $name -ErrorAction SilentlyContinue | ForEach-Object {
            if ($_.Path -and $_.Path.StartsWith($Dest, [StringComparison]::OrdinalIgnoreCase)) {
                Write-Host "stopping $name (PID $($_.Id)) so install can replace $($_.Path)"
                Stop-Process -Id $_.Id -Force -ErrorAction Stop
            }
        }
    }
    Start-Sleep -Milliseconds 300
}

function Install-Binary {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Name
    )

    $src = Join-Path $BinDir "$Name.exe"
    $dst = Join-Path $Dest "$Name.exe"
    $stale = "$dst.old"

    if (Test-Path $stale) {
        Remove-Item -Force $stale -ErrorAction SilentlyContinue
    }

    if (Test-Path $dst) {
        try {
            Copy-Item -Force $src $dst
            return
        } catch {
            Write-Host "could not overwrite locked $dst; trying rename-then-copy"
        }

        if (Test-Path $stale) {
            Remove-Item -Force $stale -ErrorAction SilentlyContinue
        }
        Move-Item -Force $dst $stale
    }

    Copy-Item -Force $src $dst
    Remove-Item -Force $stale -ErrorAction SilentlyContinue
}

function Install-ToDest {
    New-Item -ItemType Directory -Force -Path $Dest | Out-Null
    Stop-DestProcesses
    Install-Binary -Name "clx"
    Install-Binary -Name "clx-ai"
    Install-Binary -Name "clxmax"
}

function Add-ToUserPath {
    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($userPath -notlike "*$Dest*") {
        $newPath = if ($userPath) { "$userPath;$Dest" } else { $Dest }
        [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
        $env:Path = "$env:Path;$Dest"
        Write-Host "added $Dest to user PATH (restart shell if needed)"
    }
}

Require-Go
Build-Binaries
Install-ToDest
Add-ToUserPath

# Stale binaries in the repo root shadow "clx" when the shell cwd is the clone (Windows
# searches .\clx.exe before PATH). Remove them so make install fixes alias/config subcommands.
foreach ($stale in @("clx.exe", "clx-ai.exe", "clxmax.exe", "clx", "clx-ai", "clxmax")) {
    $p = Join-Path $Root $stale
    if (Test-Path $p) {
        Remove-Item -Force $p
        Write-Host "removed stale $p (was shadowing installed clx in this directory)"
    }
}

Write-Host "installed clx, clx-ai (internal), and clxmax to $Dest"
$installed = Join-Path $Dest "clx.exe"
& $installed --version
& $installed doctor
$onPath = (Get-Command clx -ErrorAction SilentlyContinue).Source
if ($onPath -and ($onPath -ne $installed)) {
    Write-Warning "another clx is first on PATH: $onPath"
    Write-Warning "from this repo directory use: $installed <args>  (or remove the shadowing binary)"
}
