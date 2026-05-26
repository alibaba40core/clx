#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN_DIR="${ROOT}/bin"

: "${MAX_BIN_SIZE_BYTES:=$((20 * 1024 * 1024))}"

failures=0

log() {
  printf '[budgets] %s\n' "$*"
}

fail() {
  log "$1 -- FAIL"
  failures=$((failures + 1))
}

pass() {
  log "$1 -- OK"
}

resolve_bin() {
  local os
  os="$(uname -s)"
  case "${os}" in
    MINGW* | MSYS* | CYGWIN*)
      echo "${BIN_DIR}/clx.exe"
      ;;
    *)
      echo "${BIN_DIR}/clx"
      ;;
  esac
}

format_bytes() {
  awk -v b="$1" 'BEGIN { printf "%.2f MB", b / (1024 * 1024) }'
}

check_binary_size() {
  local bin size limit
  bin="$(resolve_bin)"
  if [[ ! -f "${bin}" ]]; then
    fail "binary size: ${bin} not found (run make build first)"
    return
  fi
  size="$(wc -c < "${bin}" | tr -d '[:space:]')"
  limit="$(format_bytes "${MAX_BIN_SIZE_BYTES}")"
  if (( size > MAX_BIN_SIZE_BYTES )); then
    fail "binary size: $(format_bytes "${size}") (limit ${limit})"
    return
  fi
  pass "binary size: $(format_bytes "${size}") (limit ${limit})"
}

main() {
  cd "${ROOT}"
  unset GOFLAGS || true
  make build
  check_binary_size
  if (( failures > 0 )); then
    exit 1
  fi
}

main "$@"
