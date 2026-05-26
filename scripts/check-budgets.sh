#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN_DIR="${ROOT}/bin"

: "${MAX_BIN_SIZE_BYTES:=$((20 * 1024 * 1024))}"
: "${MAX_COLDSTART_MS:=}"
: "${COLDSTART_RUNS:=11}"
: "${COLDSTART_WARMUP:=2}"

failures=0
CLX_HOME=""

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

default_coldstart_limit_ms() {
  local os
  os="$(uname -s)"
  case "${os}" in
    Linux*)
      echo 50
      ;;
    Darwin*)
      echo 75
      ;;
    MINGW* | MSYS* | CYGWIN*)
      echo 120
      ;;
    *)
      echo 50
      ;;
  esac
}

require_python3() {
  if ! command -v python3 >/dev/null 2>&1; then
    fail "cold start: python3 is required for portable timing"
    return 1
  fi
}

now_ns() {
  python3 -c 'import time; print(time.time_ns())'
}

run_version_once() {
  local bin="$1"
  CLX_HOME="${CLX_HOME}" "${bin}" --version >/dev/null
}

measure_cold_start_ms() {
  local bin start end elapsed_ns
  bin="$1"
  start="$(now_ns)"
  CLX_HOME="${CLX_HOME}" "${bin}" --version >/dev/null
  end="$(now_ns)"
  elapsed_ns=$((end - start))
  echo $((elapsed_ns / 1000000))
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

check_cold_start() {
  local bin limit_ms i median worst samples_line
  local -a samples

  if ! require_python3; then
    return
  fi

  if [[ -z "${MAX_COLDSTART_MS}" ]]; then
    MAX_COLDSTART_MS="$(default_coldstart_limit_ms)"
  fi

  bin="$(resolve_bin)"
  if [[ ! -x "${bin}" && ! -f "${bin}" ]]; then
    fail "cold start: ${bin} not found (run make build first)"
    return
  fi

  for ((i = 0; i < COLDSTART_WARMUP; i++)); do
    run_version_once "${bin}"
  done

  samples=()
  for ((i = 0; i < COLDSTART_RUNS; i++)); do
    samples+=("$(measure_cold_start_ms "${bin}")")
  done

  median="$(printf '%s\n' "${samples[@]}" | sort -n | awk -v n="${#samples[@]}" 'BEGIN { mid = int((n + 1) / 2) } NR == mid { print; exit }')"
  worst="$(printf '%s\n' "${samples[@]}" | sort -n | tail -n 1)"
  samples_line="$(IFS=, ; echo "${samples[*]}")"

  if (( median > MAX_COLDSTART_MS )); then
    fail "cold start (median of ${COLDSTART_RUNS}, ${COLDSTART_WARMUP} warmup): ${median} ms (worst ${worst} ms, limit ${MAX_COLDSTART_MS} ms; samples: ${samples_line})"
    return
  fi

  pass "cold start (median of ${COLDSTART_RUNS}, ${COLDSTART_WARMUP} warmup): ${median} ms (worst ${worst} ms, limit ${MAX_COLDSTART_MS} ms)"
}

print_deferred_budgets() {
  log "steady-state RSS: skipped (Phase 2 -- needs /usr/bin/time -v, Linux only)"
  log "peak RSS (AI path): skipped (no AI integration yet)"
  log "goroutines at idle: skipped (Phase 2 -- needs runtime probe)"
}

setup_clx_home() {
  CLX_HOME="$(mktemp -d)"
  trap 'rm -rf "${CLX_HOME}"' EXIT
}

build_release() {
  local version commit ldf
  if command -v make >/dev/null 2>&1; then
    make build
    return
  fi
  version="$(git -C "${ROOT}" describe --tags --always 2>/dev/null || echo dev)"
  commit="$(git -C "${ROOT}" rev-parse --short HEAD 2>/dev/null || echo unknown)"
  ldf="-s -w -X github.com/alibaba40core/clx/internal/cliversion.Version=${version} -X github.com/alibaba40core/clx/internal/cliversion.Commit=${commit}"
  mkdir -p "${BIN_DIR}"
  go build -trimpath -ldflags="${ldf}" -o "${BIN_DIR}/clx" ./cmd/clx
  go build -trimpath -ldflags="${ldf}" -o "${BIN_DIR}/clxmax" ./cmd/clxmax
}

main() {
  cd "${ROOT}"
  unset GOFLAGS || true
  build_release
  setup_clx_home
  check_binary_size
  check_cold_start
  print_deferred_budgets
  if (( failures > 0 )); then
    exit 1
  fi
}

main "$@"
