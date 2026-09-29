#!/usr/bin/env bash
# ==============================================================================
# OhioFiles CLI (ohio) One-Line Installer for macOS & Linux
# Usage: curl -fsSL https://ohiofiles.cloud/install.sh | bash
# ==============================================================================

set -e

RESET="\033[0m"
BOLD="\033[1m"
CYAN="\033[36m"
GREEN="\033[32m"
GRAY="\033[90m"
RED="\033[31m"

echo -e "\n${BOLD}${CYAN}⚡ OhioFiles CLI (ohio) Installer${RESET}"
echo -e "${GRAY}Fast, anonymous, login-free file sharing & TUI${RESET}\n"

# 1. Detect OS & Architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
  x86_64|amd64)
    ARCH_TARGET="amd64"
    ;;
  arm64|aarch64)
    ARCH_TARGET="arm64"
    ;;
  *)
    echo -e "${RED}✖ Unsupported architecture: $ARCH${RESET}"
    exit 1
    ;;
esac

case "$OS" in
  darwin|linux)
    ;;
  *)
    echo -e "${RED}✖ Unsupported operating system: $OS. For Windows, use install.ps1${RESET}"
    exit 1
    ;;
esac

# 2. Determine Installation Target Directory
INSTALL_DIR="/usr/local/bin"
if [ ! -w "$INSTALL_DIR" ]; then
  if [ "$(id -u)" -eq 0 ]; then
    mkdir -p "$INSTALL_DIR"
  else
    INSTALL_DIR="$HOME/.local/bin"
    mkdir -p "$INSTALL_DIR"
  fi
fi

# 3. Locate or Download Binary
LOCAL_BIN_SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" 2>/dev/null && pwd)"
LOCAL_CANDIDATE=""

if [ -f "$LOCAL_BIN_SCRIPT_DIR/cli/bin/ohio" ]; then
  LOCAL_CANDIDATE="$LOCAL_BIN_SCRIPT_DIR/cli/bin/ohio"
elif [ -f "$LOCAL_BIN_SCRIPT_DIR/bin/ohio" ]; then
  LOCAL_CANDIDATE="$LOCAL_BIN_SCRIPT_DIR/bin/ohio"
elif [ -f "$LOCAL_BIN_SCRIPT_DIR/ohio" ]; then
  LOCAL_CANDIDATE="$LOCAL_BIN_SCRIPT_DIR/ohio"
fi

if [ -n "$LOCAL_CANDIDATE" ]; then
  echo -e "  ${GRAY}• Installing from local build (${LOCAL_CANDIDATE})...${RESET}"
  cp -f "$LOCAL_CANDIDATE" "$INSTALL_DIR/ohio"
else
  # Remote release download
  DOWNLOAD_URL="https://api.ohiofiles.cloud/releases/latest/ohio-${OS}-${ARCH_TARGET}"
  echo -e "  ${GRAY}• Downloading OhioCLI for ${OS}/${ARCH_TARGET}...${RESET}"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$DOWNLOAD_URL" -o "$INSTALL_DIR/ohio" || {
      # Fallback build if go is installed
      if command -v go >/dev/null 2>&1 && [ -d "$LOCAL_BIN_SCRIPT_DIR/cli" ]; then
        echo -e "  ${GRAY}• Compiling native binary using Go...${RESET}"
        (cd "$LOCAL_BIN_SCRIPT_DIR/cli" && GOCACHE=/tmp/gocache GOPATH=/tmp/gopath go build -o "$INSTALL_DIR/ohio" ./cmd/ohfs)
      else
        echo -e "${RED}✖ Failed downloading pre-built binary.${RESET}"
        exit 1
      fi
    }
  else
    echo -e "${RED}✖ curl is required to install.${RESET}"
    exit 1
  fi
fi

chmod +x "$INSTALL_DIR/ohio"

# 4. Create alias / symlink for ohfs
ln -sf "$INSTALL_DIR/ohio" "$INSTALL_DIR/ohfs"

# 5. Check PATH
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
  echo -e "  ${GRAY}• Notice: $INSTALL_DIR is not in your PATH.${RESET}"
  echo -e "    Add it by running: export PATH=\"$INSTALL_DIR:\$PATH\""
fi

echo -e "\n${GREEN}${BOLD}✔ Successfully installed OhioCLI!${RESET}"
echo -e "  Command:  ${CYAN}${INSTALL_DIR}/ohio${RESET} (alias: ${CYAN}ohfs${RESET})"
echo -e "  Version:  ${BOLD}v2.1.0${RESET}"
echo -e "\n${BOLD}Quick Start:${RESET}"
echo -e "  ${CYAN}ohio tui${RESET}              Launch full-screen interactive dashboard"
echo -e "  ${CYAN}ohio -u file.txt${RESET}      Instantly upload a file"
echo -e "  ${CYAN}cat file | ohio${RESET}       Pipe stdin directly into an upload"
echo -e "  ${CYAN}ohio --help${RESET}           View all available commands\n"
