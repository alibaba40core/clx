#!/usr/bin/env bash
# CLX shell hook (V1: explain-only). Source from ~/.bashrc or ~/.zshrc:
#   source /path/to/clx-hook.sh
#
# Forwards unknown commands to `clx --explain` so you can review the translation
# before running anything. Does not auto-execute (security).

clx_translate() {
  if ! command -v clx >/dev/null 2>&1; then
    echo "clx: not found on PATH" >&2
    return 127
  fi
  clx --explain "$@"
}

# Bash: optional command_not_found_handler (bash 4.1+)
if [ -n "${BASH_VERSION:-}" ]; then
  command_not_found_handle() {
    clx_translate "$1" "${@:2}"
    return 127
  }
fi
