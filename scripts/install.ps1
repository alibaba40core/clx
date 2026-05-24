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
        go build -trimpath -ldflags $ldflags -o (Join-Path $BinDir "clx.exe") ./cmd/clx
        go build -trimpath -ldflags $ldflags -o (Join-Path $BinDir "clxmax.exe") ./cmd/clxmax
    } finally {
        Pop-Location
    }
}

function Install-ToDest {
    New-Item -ItemType Directory -Force -Path $Dest | Out-Null
    Copy-Item -Force (Join-Path $BinDir "clx.exe") (Join-Path $Dest "clx.exe")
    Copy-Item -Force (Join-Path $BinDir "clxmax.exe") (Join-Path $Dest "clxmax.exe")
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

Write-Host "installed clx and clxmax to $Dest"
& (Join-Path $Dest "clx.exe") --version
& (Join-Path $Dest "clx.exe") doctor
