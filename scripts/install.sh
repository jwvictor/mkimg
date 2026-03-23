#!/usr/bin/env sh
#
# mkimg installer
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/jwvictor/mkimg/main/scripts/install.sh | sh
#
# Detects OS and architecture, downloads the latest mkimg binary from
# GitHub releases, installs it, and sets up the AI agent skill file.

set -eu

REPO="jwvictor/mkimg"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="mkimg"
SKILL_NAME="mkimg"

# ── Helpers ──────────────────────────────────────────────────────

log()  { printf "  %s\n" "$*"; }
bold() { printf "\033[1m%s\033[0m\n" "$*"; }
err()  { printf "\033[31merror:\033[0m %s\n" "$*" >&2; exit 1; }

need() {
  command -v "$1" >/dev/null 2>&1 || err "required command not found: $1"
}

# ── Detect platform ─────────────────────────────────────────────

detect_os() {
  case "$(uname -s)" in
    Darwin*)  echo "darwin"  ;;
    Linux*)   echo "linux"   ;;
    *)        err "unsupported OS: $(uname -s)" ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64)  echo "amd64"  ;;
    arm64|aarch64) echo "arm64"  ;;
    *)             err "unsupported architecture: $(uname -m)" ;;
  esac
}

# ── Fetch latest version ────────────────────────────────────────

get_latest_version() {
  url="https://api.github.com/repos/${REPO}/releases/latest"
  version=$(curl -fsSL "$url" 2>/dev/null | grep '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')
  if [ -z "$version" ]; then
    err "could not determine latest version from GitHub releases"
  fi
  echo "$version"
}

# ── Install binary ──────────────────────────────────────────────

install_binary() {
  BINARY="$1"
  chmod +x "$BINARY"

  if [ -w "$INSTALL_DIR" ] || [ -w "$(dirname "$INSTALL_DIR")" ]; then
    cp "$BINARY" "${INSTALL_DIR}/${BINARY_NAME}"
    log "Binary installed to ${INSTALL_DIR}/${BINARY_NAME}"
  elif command -v sudo >/dev/null 2>&1; then
    sudo cp "$BINARY" "${INSTALL_DIR}/${BINARY_NAME}"
    log "Binary installed to ${INSTALL_DIR}/${BINARY_NAME} (via sudo)"
  else
    INSTALL_DIR="${HOME}/.local/bin"
    mkdir -p "$INSTALL_DIR"
    cp "$BINARY" "${INSTALL_DIR}/${BINARY_NAME}"
    log "Binary installed to ${INSTALL_DIR}/${BINARY_NAME}"
    case ":$PATH:" in
      *":${INSTALL_DIR}:"*) ;;
      *) log "Warning: ${INSTALL_DIR} is not in your PATH. Add it with:"
         log "  export PATH=\"${INSTALL_DIR}:\$PATH\""
         ;;
    esac
  fi
}

# ── Install skill ───────────────────────────────────────────────

install_skill() {
  VERSION="$1"
  SKILL_URL="https://raw.githubusercontent.com/${REPO}/${VERSION}/skill/SKILL.md"
  INSTALLED=0

  # Download skill file to a temp location first
  SKILL_TMP=$(mktemp)
  curl -fsSL "$SKILL_URL" -o "$SKILL_TMP" || {
    log "Warning: could not download skill file (mkimg will still work, but AI agents won't have the skill)"
    rm -f "$SKILL_TMP"
    return
  }

  # Install to ~/.claude/skills if ~/.claude exists
  if [ -d "${HOME}/.claude" ]; then
    SKILL_DIR="${HOME}/.claude/skills/${SKILL_NAME}"
    mkdir -p "$SKILL_DIR"
    cp "$SKILL_TMP" "${SKILL_DIR}/SKILL.md"
    log "Skill installed to ${SKILL_DIR}/SKILL.md"
    INSTALLED=1
  fi

  # Install to ~/.openclaw/skills if ~/.openclaw exists
  if [ -d "${HOME}/.openclaw" ]; then
    SKILL_DIR="${HOME}/.openclaw/skills/${SKILL_NAME}"
    mkdir -p "$SKILL_DIR"
    cp "$SKILL_TMP" "${SKILL_DIR}/SKILL.md"
    log "Skill installed to ${SKILL_DIR}/SKILL.md"
    INSTALLED=1
  fi

  # If neither exists, default to ~/.claude
  if [ "$INSTALLED" -eq 0 ]; then
    SKILL_DIR="${HOME}/.claude/skills/${SKILL_NAME}"
    mkdir -p "$SKILL_DIR"
    cp "$SKILL_TMP" "${SKILL_DIR}/SKILL.md"
    log "Skill installed to ${SKILL_DIR}/SKILL.md"
  fi

  rm -f "$SKILL_TMP"
}

# ── Main ────────────────────────────────────────────────────────

main() {
  bold "mkimg installer"
  echo ""

  need curl
  need tar

  OS=$(detect_os)
  ARCH=$(detect_arch)
  log "Detected platform: ${OS}/${ARCH}"

  VERSION=$(get_latest_version)
  VERSION_NUM=$(echo "$VERSION" | sed 's/^v//')
  log "Latest version: ${VERSION}"

  # Download binary
  TARBALL="mkimg_${VERSION_NUM}_${OS}_${ARCH}.tar.gz"
  URL="https://github.com/${REPO}/releases/download/${VERSION}/${TARBALL}"

  log "Downloading ${TARBALL}..."
  TMPDIR=$(mktemp -d)
  trap 'rm -rf "$TMPDIR"' EXIT

  curl -fsSL "$URL" -o "${TMPDIR}/${TARBALL}" || err "download failed — check that a release exists for ${OS}/${ARCH}"
  tar -xzf "${TMPDIR}/${TARBALL}" -C "$TMPDIR" || err "failed to extract archive"

  BINARY=$(find "$TMPDIR" -name "$BINARY_NAME" -type f | head -1)
  if [ -z "$BINARY" ]; then
    err "binary '${BINARY_NAME}' not found in archive"
  fi

  # Install binary
  install_binary "$BINARY"

  # Install skill file for AI agents
  install_skill "$VERSION"

  echo ""
  bold "mkimg ${VERSION} installed successfully"
  echo ""
  log "Get started:"
  log "  mkimg new my-ad --preset instagram-story"
  log "  mkimg layer add gradient --from \"#667eea\" --to \"#764ba2\""
  log "  mkimg render -o my-ad.png"
  echo ""
  log "AI agents (Claude Code, OpenClaw) can now use mkimg automatically."
  echo ""
}

main
