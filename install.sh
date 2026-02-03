#!/usr/bin/env bash
set -euo pipefail

REPO="coffee-cup/peel"
BIN_NAME="peel"
INSTALL_DIR="/usr/local/bin"
FORCE=false

# Colors (disabled if not a terminal)
if [ -t 1 ] && command -v tput >/dev/null 2>&1; then
  GREEN=$(tput setaf 2)
  RED=$(tput setaf 1)
  YELLOW=$(tput setaf 3)
  BOLD=$(tput bold)
  RESET=$(tput sgr0)
else
  GREEN="" RED="" YELLOW="" BOLD="" RESET=""
fi

info()  { printf "%s%s%s\n" "$GREEN"  "$1" "$RESET"; }
warn()  { printf "%s%s%s\n" "$YELLOW" "$1" "$RESET"; }
error() { printf "%s%s%s\n" "$RED"    "$1" "$RESET" >&2; }
die()   { error "$1"; exit 1; }

usage() {
  cat <<EOF
${BOLD}Install ${BIN_NAME}${RESET}

Usage: install.sh [options]

Options:
  -b DIR   Install directory (default: ${INSTALL_DIR})
  -f       Skip confirmation prompt
  -h       Show this help
EOF
}

while getopts "b:fh" opt; do
  case $opt in
    b) INSTALL_DIR="$OPTARG" ;;
    f) FORCE=true ;;
    h) usage; exit 0 ;;
    *) usage; exit 1 ;;
  esac
done

# Detect OS
case "$(uname -s)" in
  Darwin) OS="darwin" ;;
  Linux)  OS="linux"  ;;
  *)      die "Unsupported OS: $(uname -s)" ;;
esac

# Detect arch
case "$(uname -m)" in
  x86_64)       ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *)            die "Unsupported architecture: $(uname -m)" ;;
esac

# Fetch latest version
info "Fetching latest release..."
VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
  | grep '"tag_name"' | sed -E 's/.*"v([^"]+)".*/\1/')
[ -n "$VERSION" ] || die "Could not determine latest version"

ARCHIVE="${BIN_NAME}_${VERSION}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/v${VERSION}/${ARCHIVE}"

info "Installing ${BIN_NAME} v${VERSION} (${OS}/${ARCH})"
echo "  ${BOLD}From:${RESET} ${URL}"
echo "  ${BOLD}To:${RESET}   ${INSTALL_DIR}/${BIN_NAME}"

# Confirm unless -f
if [ "$FORCE" = false ]; then
  printf "\nProceed? [y/N] "
  read -r answer
  case "$answer" in
    [yY]|[yY][eE][sS]) ;;
    *) warn "Aborted."; exit 0 ;;
  esac
fi

# Download and extract
TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

info "Downloading ${ARCHIVE}..."
curl -fsSL "$URL" -o "${TMPDIR}/${ARCHIVE}"

info "Extracting..."
tar -xzf "${TMPDIR}/${ARCHIVE}" -C "$TMPDIR"

# Install binary
if ! mkdir -p "$INSTALL_DIR" 2>/dev/null; then
  warn "Requires sudo to create ${INSTALL_DIR}"
  sudo mkdir -p "$INSTALL_DIR"
fi
if [ -w "$INSTALL_DIR" ]; then
  mv "${TMPDIR}/${BIN_NAME}" "${INSTALL_DIR}/${BIN_NAME}"
else
  warn "Requires sudo to write to ${INSTALL_DIR}"
  sudo mv "${TMPDIR}/${BIN_NAME}" "${INSTALL_DIR}/${BIN_NAME}"
fi
chmod +x "${INSTALL_DIR}/${BIN_NAME}"

info "Installed ${BIN_NAME} v${VERSION} to ${INSTALL_DIR}/${BIN_NAME}"
