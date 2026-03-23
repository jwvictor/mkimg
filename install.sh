#!/bin/bash
set -euo pipefail

BINARY_NAME="mkimg"
INSTALL_DIR="/usr/local/bin"
SKILL_NAME="mkimg"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

echo "Building $BINARY_NAME..."
cd "$SCRIPT_DIR"
go build -o "$BINARY_NAME" .

echo "Installing binary to $INSTALL_DIR/$BINARY_NAME..."
sudo install -m 755 "$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
rm -f "$BINARY_NAME"

echo "Installing skill..."
INSTALLED=0

# Install to ~/.claude/skills if ~/.claude exists
if [ -d "$HOME/.claude" ]; then
    SKILL_DIR="$HOME/.claude/skills/$SKILL_NAME"
    mkdir -p "$SKILL_DIR"
    cp "$SCRIPT_DIR/skill/SKILL.md" "$SKILL_DIR/SKILL.md"
    echo "  Skill installed to $SKILL_DIR/SKILL.md"
    INSTALLED=1
fi

# Install to ~/.openclaw/skills if ~/.openclaw exists
if [ -d "$HOME/.openclaw" ]; then
    SKILL_DIR="$HOME/.openclaw/skills/$SKILL_NAME"
    mkdir -p "$SKILL_DIR"
    cp "$SCRIPT_DIR/skill/SKILL.md" "$SKILL_DIR/SKILL.md"
    echo "  Skill installed to $SKILL_DIR/SKILL.md"
    INSTALLED=1
fi

# If neither exists, default to ~/.claude
if [ "$INSTALLED" -eq 0 ]; then
    SKILL_DIR="$HOME/.claude/skills/$SKILL_NAME"
    mkdir -p "$SKILL_DIR"
    cp "$SCRIPT_DIR/skill/SKILL.md" "$SKILL_DIR/SKILL.md"
    echo "  Skill installed to $SKILL_DIR/SKILL.md"
fi

echo "Done! '$BINARY_NAME' is on your PATH and the agent skill is installed."
