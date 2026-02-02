#!/usr/bin/env bash
set -euo pipefail

# healthcheck.sh
# Checks frontend `/health` and backend `/api/health` endpoints.
# Usage:
#   FRONTEND_URL=... BACKEND_URL=... ./healthcheck.sh
#   or: ./healthcheck.sh [FRONTEND_URL] [BACKEND_URL]

FRONTEND_URL="${FRONTEND_URL:-http://localhost:80/health}"
BACKEND_URL="${BACKEND_URL:-http://localhost:80/api/health}"
TIMEOUT="${TIMEOUT:-5}"

usage() {
  cat <<EOF
Usage: $0 [FRONTEND_URL] [BACKEND_URL]

Environment overrides:
  FRONTEND_URL  default: $FRONTEND_URL
  BACKEND_URL   default: $BACKEND_URL
  TIMEOUT       default: $TIMEOUT (seconds)

Returns exit code 0 if both endpoints respond successfully, non-zero otherwise.
EOF
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

if [ $# -ge 1 ] && [ -n "$1" ]; then
  FRONTEND_URL="$1"
fi
if [ $# -ge 2 ] && [ -n "$2" ]; then
  BACKEND_URL="$2"
fi

failed=0
check() {
  local name="$1" url="$2"
  if curl -sS --fail --max-time "$TIMEOUT" "$url" >/dev/null 2>&1; then
    printf "%s: OK (%s)\n" "$name" "$url"
  else
    printf "%s: FAIL (%s)\n" "$name" "$url" >&2
    failed=1
  fi
}

check "Frontend" "$FRONTEND_URL"
check "Backend" "$BACKEND_URL"

exit $failed
