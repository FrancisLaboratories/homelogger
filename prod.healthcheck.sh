#!/usr/bin/env bash
set -euo pipefail

# healthcheck.sh
# Checks the API health endpoint served by the Go binary.
# Usage:
#   ./healthcheck.sh [URL]
#   URL default: http://localhost:3005/api/health

URL="${1:-http://localhost:3005/api/health}"
TIMEOUT="${TIMEOUT:-5}"

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  cat <<EOF
Usage: $0 [URL]

Environment overrides:
  TIMEOUT  default: $TIMEOUT (seconds)

Returns exit code 0 if the endpoint responds successfully, non-zero otherwise.
EOF
  exit 0
fi

exec curl -sS --fail --max-time "$TIMEOUT" "$URL" >/dev/null
