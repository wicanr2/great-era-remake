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
IMAGE="${DSDS_GO_IMAGE:-dsds-go:1.25}"
CACHE="$ROOT/workplace/.gocache"

# mount 整個 /go/pkg（不只 pkg/mod），否則 sumdb 寫不進去、go get 會失敗
mkdir -p "$CACHE/pkg" "$CACHE/build"

# Ebiten 走 CGO，需要 X11/OpenGL 標頭檔，官方 golang image 沒有 → 自建一份
if ! docker image inspect "$IMAGE" >/dev/null 2>&1; then
  echo "[go.sh] 建 image $IMAGE（含 X11/OpenGL 開發套件）…" >&2
  docker build -f "$ROOT/tools/go.dockerfile" -t "$IMAGE" "$ROOT/tools" >&2
fi

exec docker run --rm -i \
  -v "$ROOT:/work" \
  -v "$CACHE/pkg:/go/pkg" \
  -v "$CACHE/build:/.cache/go-build" \
  -u "$(id -u):$(id -g)" \
  -w /work \
  -e GOCACHE=/.cache/go-build \
  -e GOFLAGS=-buildvcs=false \
  "$IMAGE" go "$@"
