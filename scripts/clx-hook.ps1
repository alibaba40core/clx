# CLX shell hook (V1: explain-only). Dot-source from $PROFILE:
#   . C:\path\to\clx-hook.ps1
#
# Forwards input to `clx --explain`. Does not auto-execute.

function Invoke-ClxTranslate {
    param(
        [Parameter(ValueFromRemainingArguments = $true)]
        [string[]]$Args
    )
    if (-not (Get-Command clx -ErrorAction SilentlyContinue)) {
        Write-Error "clx: not found on PATH"
        return 127
    }
    & clx --explain @Args
}

Set-Alias -Name clx_translate -Value Invoke-ClxTranslate -Scope Global -Force
