#!/usr/bin/env bash
# Go 執行包裝：一律 docker，不裝進系統環境。
#
#   tools/go.sh version
#   tools/go.sh test ./...
#   tools/go.sh build ./cmd/tpcdump
#
# module cache 放在 workplace/.gocache（gitignore），避免每次重抓。
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="golang:1.25-bookworm"
CACHE="$ROOT/workplace/.gocache"

mkdir -p "$CACHE/mod" "$CACHE/build"

exec docker run --rm -i \
  -v "$ROOT:/work" \
  -v "$CACHE/mod:/go/pkg/mod" \
  -v "$CACHE/build:/.cache/go-build" \
  -u "$(id -u):$(id -g)" \
  -w /work \
  -e GOCACHE=/.cache/go-build \
  -e GOFLAGS=-buildvcs=false \
  "$IMAGE" go "$@"
