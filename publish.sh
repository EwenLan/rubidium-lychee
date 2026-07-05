#!/bin/bash
set -euo pipefail

# Package the client into a ZIP whose root directly contains start.sh
# (matches 任务书 10.1: ZIP 根目录必须直接包含 start.sh).

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

OUTPUT="$SCRIPT_DIR/publish.zip"
rm -f "$OUTPUT"

# Build in a staging dir so start.sh lands at the zip root (no nested folder).
STAGING="$(mktemp -d)"
trap 'rm -rf "$STAGING"' EXIT

# Required root files.
cp start.sh "$STAGING/"
cp go.mod "$STAGING/"
[ -f go.sum ] && cp go.sum "$STAGING/"

# Go source files (excluding tests and local-only cmd/ tools), preserving
# directory structure. cmd/mockserver is a local test tool, not client code.
find . -type f -name "*.go" \
  ! -name "*_test.go" \
  ! -path "./vendor/*" \
  ! -path "./.git/*" \
  ! -path "./cmd/*" \
  -print0 | while IFS= read -r -d '' f; do
  rel="${f#./}"
  mkdir -p "$STAGING/$(dirname "$rel")"
  cp "$f" "$STAGING/$rel"
done

# Vendored deps (required for offline `go run` with third-party imports).
if [ -d vendor ]; then
  cp -r vendor "$STAGING/"
fi

# Zip from staging so files sit at root.
( cd "$STAGING" && zip -r -q "$OUTPUT" . )

echo "Created $OUTPUT"
ls -lh "$OUTPUT"
