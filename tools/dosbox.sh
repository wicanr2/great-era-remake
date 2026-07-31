#!/usr/bin/env bash
# DOSBox oracle 包裝（CLAUDE.md §4：DOSBox 實測是最高位階的 oracle）。
#
#   tools/dosbox.sh probe            最小探路：跑起來、截一張圖就退出
#   tools/dosbox.sh run [conf]       用指定 conf 跑
#
# 原版素材唯讀：game/ 以 :ro 掛入，遊戲的寫入導到 workplace/dosbox/drive_c。
# 截圖輸出在 workplace/dosbox/capture/。
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="dsds-dosbox:x"
WORK="$ROOT/workplace/dosbox"
ORIG="$ROOT/workplace/orig/game"

mkdir -p "$WORK/drive_c" "$WORK/capture" "$WORK/conf"

build() {
  docker build -f "$ROOT/tools/Dockerfile.dosbox" -t "$IMAGE" "$ROOT/tools"
}

if ! docker image inspect "$IMAGE" >/dev/null 2>&1; then
  echo "[dosbox.sh] image 不存在，先建一次…" >&2
  build
fi

# 原版唯讀：每次把 game/ 複製一份到可寫的 drive_c（不動原始檔）
sync_game() {
  rsync -a --delete "$ORIG/" "$WORK/drive_c/" 2>/dev/null \
    || { rm -rf "$WORK/drive_c"; mkdir -p "$WORK/drive_c"; cp -r "$ORIG/." "$WORK/drive_c/"; }
  chmod -R u+w "$WORK/drive_c"
}

case "${1:-probe}" in
  probe)
    sync_game
    docker run --rm \
      -v "$WORK:/dos" \
      -u "$(id -u):$(id -g)" \
      -e HOME=/tmp \
      "$IMAGE" \
      dosbox -conf /dos/conf/probe.conf -exit
    echo "[dosbox.sh] 截圖 → $WORK/capture"
    ;;
  run)
    sync_game
    conf="${2:-/dos/conf/probe.conf}"
    docker run --rm -v "$WORK:/dos" -u "$(id -u):$(id -g)" -e HOME=/tmp \
      "$IMAGE" dosbox -conf "$conf"
    ;;
  build) build ;;
  *) sed -n '2,10p' "$0"; exit 1 ;;
esac
