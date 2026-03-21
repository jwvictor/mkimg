#!/bin/bash
set -euo pipefail

BINARY_NAME="mkimg"
INSTALL_DIR="/usr/local/bin"
SKILL_NAME="mkimg"
SKILL_DIR="$HOME/.claude/skills/$SKILL_NAME"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

echo "Building $BINARY_NAME..."
cd "$SCRIPT_DIR"
go build -o "$BINARY_NAME" .

echo "Installing binary to $INSTALL_DIR/$BINARY_NAME..."
sudo install -m 755 "$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
rm -f "$BINARY_NAME"

echo "Installing skill to $SKILL_DIR..."
mkdir -p "$SKILL_DIR"
cp "$SCRIPT_DIR/skill/SKILL.md" "$SKILL_DIR/SKILL.md"

echo "Done! '$BINARY_NAME' is on your PATH and the Claude Code skill is installed."
