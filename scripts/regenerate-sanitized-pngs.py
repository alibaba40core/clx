#!/usr/bin/env python3
"""Regenerate presentation PNGs from sanitized terminal stubs."""

from __future__ import annotations

import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
IMG = ROOT / "doc" / "images"
RENDER = ROOT / "scripts" / "render-terminal-png.py"

STUBS: dict[str, tuple[str, str]] = {
    "clx-help-powershell.png": (
        "PowerShell",
        "PS C:\\Users> clx\nCLX — cross-platform command intelligence\nUsage: clx [flags] <input>\n  doctor, init, config, safety, policy, cache, alias",
    ),
    "powershell-exec-nl-ip.png": (
        "PowerShell",
        "PS C:\\Users> clx -y \"what is my ip\"\n\nWireless LAN adapter Wi-Fi:\n   Connection-specific DNS Suffix  . : example.local\n   IPv4 Address. . . . . . . . . . . : 192.0.2.1",
    ),
    "powershell-exec-grep-unix.png": (
        "PowerShell",
        "PS C:\\Users> clx -y grep CLX C:\\Users\\example-workspace\\clx\\README.md\n\nuser\\example-workspace\\clx\\README.md:1:# CLX",
    ),
    "powershell-doctor.png": (
        "PowerShell",
        "PS C:\\Users> clx doctor\nCLX environment profile written to C:\\Users\\user\\.clx\\system_profile.json\n  OS:       windows 10.0.22631\n  Home:     C:\\Users\\user",
    ),
    "cmd-doctor.png": (
        "CMD",
        "C:\\Users> clx doctor\nCLX environment profile written to C:\\Users\\user\\.clx\\system_profile.json\n  OS:       windows 10.0.22631",
    ),
    "powershell-config-show.png": (
        "PowerShell",
        "PS C:\\Users> clx config show\nproviders.openai.api_key: ****XXXX\nproviders.gemini.api_key: ****YYYY",
    ),
    "cmd-config-show.png": (
        "CMD",
        "C:\\Users> clx config show\nproviders.openai.api_key: ****XXXX\nproviders.gemini.api_key: ****YYYY",
    ),
    "powershell-config-help.png": (
        "PowerShell",
        "PS C:\\Users> clx config --help\nUsage: clx config [show|set|unset]",
    ),
    "cmd-config-help.png": (
        "CMD",
        "C:\\Users> clx config --help\nUsage: clx config [show|set|unset]",
    ),
    "powershell-policy-show.png": (
        "PowerShell",
        "PS C:\\Users> clx policy show\nPolicy file: C:\\Users\\user\\.clx\\policy.yaml\n  default_action: allow",
    ),
    "cmd-policy-show.png": (
        "CMD",
        "C:\\Users> clx policy show\nPolicy file: C:\\Users\\user\\.clx\\policy.yaml",
    ),
    "powershell-cache-status.png": (
        "PowerShell",
        "PS C:\\Users> clx cache status\n  intents: 12 entries\n    path: C:\\Users\\user\\.clx\\cache\\intents.json",
    ),
    "cmd-cache-status.png": (
        "CMD",
        "C:\\Users> clx cache status\n  intents: 12 entries\n    path: C:\\Users\\user\\.clx\\cache\\intents.json",
    ),
    "powershell-safety-show.png": (
        "PowerShell",
        "PS C:\\Users> clx safety show\nRisk engine: enabled\nConfirmation: medium+",
    ),
    "cmd-safety-show.png": (
        "CMD",
        "C:\\Users> clx safety show\nRisk engine: enabled",
    ),
    "powershell-init-help.png": (
        "PowerShell",
        "PS C:\\Users> clx init --help\nUsage: clx init [--provider openai|ollama|gemini]",
    ),
    "cmd-init-help.png": (
        "CMD",
        "C:\\Users> clx init --help\nUsage: clx init [--provider openai|ollama|gemini]",
    ),
    "powershell-git-status-explain.png": (
        "PowerShell",
        "PS C:\\Users> clx --explain git status\nIntent: git_status\nCommand: git status",
    ),
    "cmd-git-status-explain.png": (
        "CMD",
        "C:\\Users> clx --explain git status\nIntent: git_status\nCommand: git status",
    ),
    "powershell-ping-explain.png": (
        "PowerShell",
        "PS C:\\Users> clx --explain ping google.com\nIntent: ping_host\nCommand: Test-Connection google.com",
    ),
    "cmd-ping-explain.png": (
        "CMD",
        "C:\\Users> clx --explain ping google.com\nIntent: ping_host\nCommand: ping google.com",
    ),
    "powershell-docker-ps-explain.png": (
        "PowerShell",
        "PS C:\\Users> clx --explain docker ps\nIntent: docker_ps\nCommand: docker ps",
    ),
    "cmd-docker-ps-explain.png": (
        "CMD",
        "C:\\Users> clx --explain docker ps\nIntent: docker_ps\nCommand: docker ps",
    ),
    "powershell-version.png": (
        "PowerShell",
        "PS C:\\Users> clx --version\nclx version 1.0.2",
    ),
    "cmd-version.png": (
        "CMD",
        "C:\\Users> clx --version\nclx version 1.0.2",
    ),
    "powershell-grep-explain.png": (
        "PowerShell",
        "PS C:\\Users> clx --explain grep errors logs.txt\nIntent: grep_file\nCommand: Select-String errors logs.txt",
    ),
    "cmd-grep-explain.png": (
        "CMD",
        "C:\\Users> clx --explain grep errors logs.txt\nIntent: grep_file\nCommand: findstr errors logs.txt",
    ),
    "powershell-ls-explain.png": (
        "PowerShell",
        "PS C:\\Users> clx --explain ls .\nIntent: list_dir\nCommand: Get-ChildItem .",
    ),
    "cmd-ls-explain.png": (
        "CMD",
        "C:\\Users> clx --explain ls .\nIntent: list_dir\nCommand: dir .",
    ),
    "powershell-mkdir-explain.png": (
        "PowerShell",
        "PS C:\\Users> clx --explain mkdir demo-folder\nIntent: make_dir\nCommand: New-Item -ItemType Directory demo-folder",
    ),
    "cmd-mkdir-explain.png": (
        "CMD",
        "C:\\Users> clx --explain mkdir demo-folder\nIntent: make_dir\nCommand: mkdir demo-folder",
    ),
    "powershell-find-today.png": (
        "PowerShell",
        "PS C:\\Users> clx --explain find all files modified today\nIntent: find_modified_today\nCommand: Get-ChildItem -Recurse ...",
    ),
    "cmd-find-today.png": (
        "CMD",
        "C:\\Users> clx --explain find all files modified today\nIntent: find_modified_today\nCommand: forfiles ...",
    ),
    "powershell-ip-explain.png": (
        "PowerShell",
        "PS C:\\Users> clx --explain what is my ip\nIntent: show_ip\nCommand: ipconfig",
    ),
    "cmd-ip-explain.png": (
        "CMD",
        "C:\\Users> clx --explain what is my ip\nIntent: show_ip\nCommand: ipconfig",
    ),
    "powershell-pwd-dryrun.png": (
        "PowerShell",
        "PS C:\\Users> clx --dry-run pwd\nIntent: print_working_dir\nCommand: Get-Location (dry-run)",
    ),
    "cmd-pwd-dryrun.png": (
        "CMD",
        "C:\\Users> clx --dry-run pwd\nIntent: print_working_dir\nCommand: cd (dry-run)",
    ),
}


def main() -> int:
    tmp = ROOT / ".purge-stubs"
    tmp.mkdir(exist_ok=True)
    for name, (title, body) in STUBS.items():
        stub = tmp / name.replace(".png", ".txt")
        stub.write_text(body + "\n", encoding="utf-8")
        out = IMG / name
        subprocess.run(
            [sys.executable, str(RENDER), str(stub), str(out), "--title", title],
            check=True,
        )
    print(f"regenerated {len(STUBS)} PNGs")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
