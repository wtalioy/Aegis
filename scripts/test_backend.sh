#!/bin/sh

set -eu

check_no_legacy_backend_refs() {
  legacy_pkg_files="$(find pkg -name '*.go' -print 2>/dev/null || true)"
  if [ -n "$legacy_pkg_files" ]; then
    echo "legacy backend packages remain under pkg/:"
    echo "$legacy_pkg_files"
    return 1
  fi

  legacy_imports="$(rg -n '"aegis/pkg/' internal cmd tests || true)"
  if [ -n "$legacy_imports" ]; then
    echo "legacy backend imports remain in active code:"
    echo "$legacy_imports"
    return 1
  fi

  legacy_shared_api="$(find internal/shared/api -name '*.go' -print 2>/dev/null || true)"
  if [ -n "$legacy_shared_api" ]; then
    echo "legacy shared API transport types remain:"
    echo "$legacy_shared_api"
    return 1
  fi

  legacy_routes="$(rg -n '/api(/|$)' internal/platform/http cmd tests frontend/src | rg -v '/api/v1' || true)"
  if [ -n "$legacy_routes" ]; then
    echo "legacy /api/ route references remain in active code:"
    echo "$legacy_routes"
    return 1
  fi
}

check_no_legacy_backend_refs
mkdir -p /tmp/aegis-gocache
GOCACHE=/tmp/aegis-gocache go test ./...
check_no_legacy_backend_refs
