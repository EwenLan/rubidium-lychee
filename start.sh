#!/bin/bash
set -e

if [ "$#" -ne 3 ]; then
  echo "Usage: $0 <playerId> <host> <port>" >&2
  exit 1
fi

PLAYER_ID="$1"
HOST="$2"
PORT="$3"

# Resolve script directory so `go run .` works regardless of CWD.
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

go run . --playerId="${PLAYER_ID}" --host="${HOST}" --port="${PORT}"
