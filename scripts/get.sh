#!/usr/bin/env bash
# CLX download installer (macOS / Linux).
#
# Fetches prebuilt clx, clx-ai (internal worker), and clxmax from GitHub Releases —
# no Go toolchain and no source checkout required. Downloads are verified against the
# published checksums.txt (SHA-256) before anything is installed.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/alibaba40core/clx/main/scripts/get.sh | bash
#
# Environment overrides:
#   CLX_VERSION      Release tag to install (default: latest), e.g. v1.0.2
#   CLX_INSTALL_DIR  Install destination (default: /usr/local/bin if writable, else ~/.local/bin)
set -euo pipefail

REPO="alibaba40core/clx"
VERSION="${CLX_VERSION:-latest}"

log() { printf 'clx-install: %s\n' "$*" >&2; }
err() { printf 'clx-install: error: %s\n' "$*" >&2; exit 1; }

need() { command -v "$1" >/dev/null 2>&1 || err "missing required tool: $1"; }

detect_os() {
  local os
  os="$(uname -s)"
  case "${os}" in
    Linux*)  echo "linux" ;;
    Darwin*) echo "darwin" ;;
    *)       err "unsupported OS: ${os} (use scripts/get.ps1 on Windows)" ;;
  esac
}

detect_arch() {
  local arch
  arch="$(uname -m)"
  case "${arch}" in
    x86_64 | amd64)  echo "amd64" ;;
    arm64 | aarch64) echo "arm64" ;;
    *)               err "unsupported architecture: ${arch}" ;;
  esac
}

download() {
  # download <url> <dest> ; returns non-zero on HTTP failure
  local url="$1" dest="$2"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "${url}" -o "${dest}"
  elif command -v wget >/dev/null 2>&1; then
    wget -qO "${dest}" "${url}"
  else
    err "need curl or wget to download"
  fi
}

sha256_of() {
  local file="$1"
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "${file}" | awk '{print $1}'
  elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "${file}" | awk '{print $1}'
  else
    err "need sha256sum or shasum to verify download"
  fi
}

choose_install_dir() {
  if [[ -n "${CLX_INSTALL_DIR:-}" ]]; then
    mkdir -p "${CLX_INSTALL_DIR}"
    echo "${CLX_INSTALL_DIR}"
    return
  fi
  if [[ -w /usr/local/bin ]]; then
    echo "/usr/local/bin"
    return
  fi
  mkdir -p "${HOME}/.local/bin"
  echo "${HOME}/.local/bin"
}

main() {
  need uname
  need tar
  need awk

  local os arch asset base tmp
  os="$(detect_os)"
  arch="$(detect_arch)"
  asset="clx_${os}_${arch}.tar.gz"

  if [[ "${VERSION}" == "latest" ]]; then
    base="https://github.com/${REPO}/releases/latest/download"
  else
    base="https://github.com/${REPO}/releases/download/${VERSION}"
  fi

  tmp="$(mktemp -d)"
  trap 'rm -rf "${tmp}"' EXIT

  log "downloading ${asset} (${VERSION})"
  download "${base}/${asset}" "${tmp}/${asset}" \
    || err "could not download ${base}/${asset} (does the release exist for ${os}/${arch}?)"

  log "verifying checksum"
  download "${base}/checksums.txt" "${tmp}/checksums.txt" \
    || err "could not download checksums.txt"

  local want got
  want="$(awk -v a="${asset}" '$2 == a || $2 == "*"a {print $1}' "${tmp}/checksums.txt" | head -n1)"
  [[ -n "${want}" ]] || err "no checksum entry for ${asset}"
  got="$(sha256_of "${tmp}/${asset}")"
  [[ "${want}" == "${got}" ]] || err "checksum mismatch for ${asset} (want ${want}, got ${got})"

  log "extracting"
  tar -xzf "${tmp}/${asset}" -C "${tmp}"

  local dest
  dest="$(choose_install_dir)"
  install -m 0755 "${tmp}/clx" "${dest}/clx" 2>/dev/null || { cp "${tmp}/clx" "${dest}/clx" && chmod 0755 "${dest}/clx"; }
  local installed="clx"
  if [[ -f "${tmp}/clx-ai" ]]; then
    install -m 0755 "${tmp}/clx-ai" "${dest}/clx-ai" 2>/dev/null || { cp "${tmp}/clx-ai" "${dest}/clx-ai" && chmod 0755 "${dest}/clx-ai"; }
    installed="clx and clx-ai"
  fi
  if [[ -f "${tmp}/clxmax" ]]; then
    install -m 0755 "${tmp}/clxmax" "${dest}/clxmax" 2>/dev/null || { cp "${tmp}/clxmax" "${dest}/clxmax" && chmod 0755 "${dest}/clxmax"; }
    installed="${installed} and clxmax"
  fi

  log "installed ${installed} to ${dest}"
  if [[ ":${PATH}:" != *":${dest}:"* ]]; then
    log "note: ${dest} is not on your PATH — add this to your shell profile:"
    log "  export PATH=\"${dest}:\$PATH\""
  fi

  "${dest}/clx" --version || true
}

main "$@"
