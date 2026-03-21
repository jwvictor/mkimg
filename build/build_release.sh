#!/bin/bash
#
# Cross-compile mkimg binaries for GitHub releases.
#
# Usage:
#   ./build/build_release.sh <version>
#
# Example:
#   ./build/build_release.sh 0.1.0
#
# Produces tarballs in dist/ ready for upload to GitHub releases:
#   mkimg_0.1.0_darwin_arm64.tar.gz
#   mkimg_0.1.0_darwin_amd64.tar.gz
#   mkimg_0.1.0_linux_arm64.tar.gz
#   mkimg_0.1.0_linux_amd64.tar.gz

set -euo pipefail

MKIMG_REPO="${MKIMG_REPO:-$(cd "$(dirname "$0")/.." && pwd)}"
VERSION="${1:-}"
DIST_DIR="${MKIMG_REPO}/dist"

if [ -z "$VERSION" ]; then
  echo "Usage: $0 <version>"
  echo "Example: $0 0.1.0"
  exit 1
fi

# Platforms to build for
PLATFORMS=(
  "darwin/arm64"
  "darwin/amd64"
  "linux/arm64"
  "linux/amd64"
)

echo "Building mkimg v${VERSION} from ${MKIMG_REPO}"
echo ""

# Clean and create dist directory
rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"

cd "$MKIMG_REPO"

for platform in "${PLATFORMS[@]}"; do
  OS="${platform%/*}"
  ARCH="${platform#*/}"
  OUTPUT_NAME="mkimg_${VERSION}_${OS}_${ARCH}"
  BINARY_NAME="mkimg"

  echo "  Building ${OS}/${ARCH}..."

  # Cross-compile
  GOOS="$OS" GOARCH="$ARCH" CGO_ENABLED=0 \
    go build -ldflags="-s -w" \
    -o "${DIST_DIR}/${BINARY_NAME}" .

  # Create tarball
  (cd "$DIST_DIR" && tar -czf "${OUTPUT_NAME}.tar.gz" "$BINARY_NAME" && rm "$BINARY_NAME")

  echo "    -> dist/${OUTPUT_NAME}.tar.gz"
done

echo ""
echo "Done. Release artifacts in ${DIST_DIR}/:"
ls -lh "$DIST_DIR"/*.tar.gz
