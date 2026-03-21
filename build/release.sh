#!/bin/bash
#
# Full mkimg release pipeline: build binaries, tag, push, create GitHub release.
#
# Usage:
#   ./build/release.sh <version>
#
# Example:
#   ./build/release.sh 0.1.0
#
# Prerequisites:
#   - gh auth login (GitHub CLI authenticated)
#   - Go installed (for cross-compilation)

set -euo pipefail

VERSION="${1:-}"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
MKIMG_REPO="${SCRIPT_DIR}/.."
DIST_DIR="${MKIMG_REPO}/dist"

if [ -z "$VERSION" ]; then
  echo "Usage: $0 <version>"
  echo "Example: $0 0.1.0"
  exit 1
fi

echo "=== Releasing mkimg v${VERSION} ==="
echo ""

# 1. Build cross-platform binaries
echo "Step 1: Building binaries..."
"${SCRIPT_DIR}/build_release.sh" "$VERSION"
echo ""

# 2. Tag the release
echo "Step 2: Tagging v${VERSION}..."
cd "$MKIMG_REPO"
git tag "v${VERSION}"
git push origin "v${VERSION}"
echo ""

# 3. Create GitHub release with binaries
echo "Step 3: Creating GitHub release..."
gh release create "v${VERSION}" "${DIST_DIR}"/*.tar.gz \
  --title "v${VERSION}" \
  --notes "Release v${VERSION}"
echo ""

echo "=== Done! ==="
echo "Release: https://github.com/jwvictor/mkimg/releases/tag/v${VERSION}"
