#!/bin/bash
set -euo pipefail

BINARY_NAME="mkimg"
INSTALL_DIR="/usr/local/bin"
SKILL_NAME="mkimg"
SKILL_DIR="$HOME/.claude/skills/$SKILL_NAME"

if [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
    echo "Removing binary from $INSTALL_DIR/$BINARY_NAME..."
    sudo rm -f "$INSTALL_DIR/$BINARY_NAME"
    echo "Binary removed."
else
    echo "Binary not found at $INSTALL_DIR/$BINARY_NAME (already removed?)."
fi

if [ -d "$SKILL_DIR" ]; then
    echo "Removing skill from $SKILL_DIR..."
    rm -rf "$SKILL_DIR"
    echo "Skill removed."
else
    echo "Skill not found at $SKILL_DIR (already removed?)."
fi

echo "Done! $BINARY_NAME has been uninstalled."
