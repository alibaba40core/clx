#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN_DIR="${ROOT}/bin"
VERSION="$(git -C "${ROOT}" describe --tags --always 2>/dev/null || echo dev)"
COMMIT="$(git -C "${ROOT}" rev-parse --short HEAD 2>/dev/null || echo unknown)"
LDFLAGS="-s -w -X github.com/alibaba40core/clx/internal/cliversion.Version=${VERSION} -X github.com/alibaba40core/clx/internal/cliversion.Commit=${COMMIT}"

require_go() {
  if ! command -v go >/dev/null 2>&1; then
    echo "error: go is not on PATH (need Go 1.26+)" >&2
    exit 1
  fi
}

build_binaries() {
  mkdir -p "${BIN_DIR}"
  go run ./cmd/genrules
  go build -trimpath -tags=lite -ldflags="${LDFLAGS}" -o "${BIN_DIR}/clx" ./cmd/clx
  go build -trimpath -ldflags="${LDFLAGS}" -o "${BIN_DIR}/clx-ai" ./cmd/clx-ai
  go build -trimpath -ldflags="${LDFLAGS}" -o "${BIN_DIR}/clxmax" ./cmd/clxmax
}

install_dest() {
  if [[ -w /usr/local/bin ]]; then
    echo "/usr/local/bin"
    return
  fi
  mkdir -p "${HOME}/.local/bin"
  echo "${HOME}/.local/bin"
}

main() {
  require_go
  cd "${ROOT}"
  build_binaries

  DEST="$(install_dest)"
  cp "${BIN_DIR}/clx" "${DEST}/clx"
  cp "${BIN_DIR}/clx-ai" "${DEST}/clx-ai"
  cp "${BIN_DIR}/clxmax" "${DEST}/clxmax"
  chmod +x "${DEST}/clx" "${DEST}/clx-ai" "${DEST}/clxmax"

  # Remove stale repo-root binaries that shadow PATH when cwd is the clone.
  for stale in clx clxmax; do
    if [[ -f "${ROOT}/${stale}" ]]; then
      rm -f "${ROOT}/${stale}"
      echo "removed stale ${ROOT}/${stale} (was shadowing installed clx in this directory)"
    fi
  done

  echo "installed clx, clx-ai (internal), and clxmax to ${DEST}"
  if [[ "${DEST}" == "${HOME}/.local/bin" ]]; then
    echo "ensure ${DEST} is on your PATH"
  fi

  "${DEST}/clx" --version
  "${DEST}/clx" doctor
}

main "$@"
