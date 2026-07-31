#!/usr/bin/env bash
# DOSBox oracle 包裝（CLAUDE.md §4：DOSBox 實測是最高位階的 oracle）。
#
#   tools/dosbox.sh probe                     跑 play.bat，等間隔截三張
#   tools/dosbox.sh run "GRT.EXE" "wait:8;shot:a;key:Return;wait:3;shot:b"
#
# 沿用現成的 dosbox image（含 Xvfb / xdotool / ImageMagick），不自建。
# 原版素材唯讀：每次把 game/ 複製一份到可寫的 workplace/dosbox/drive_c，不動原始檔。
# 截圖輸出在 workplace/dosbox/shots/。
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="${DSDS_DOSBOX_IMAGE:-fd2-dosbox-screenshot-local}"
WORK="$ROOT/workplace/dosbox"
ORIG="$ROOT/workplace/orig/game"

mkdir -p "$WORK/drive_c" "$WORK/shots"

sync_game() {
  rm -rf "$WORK/drive_c"
  mkdir -p "$WORK/drive_c"
  cp -r "$ORIG/." "$WORK/drive_c/"
  chmod -R u+w "$WORK/drive_c"
}

cmd="${2:-play.bat}"
timeline="${3:-wait:6;shot:01-boot;wait:6;shot:02-after;wait:8;shot:03-later}"
cycles="${4:-fixed 12000}"

case "${1:-probe}" in
  probe|run)
    sync_game
    docker run --rm \
      -v "$WORK/drive_c:/game" \
      -v "$WORK/shots:/shots" \
      -v "$ROOT/tools/dosbox_runner.sh:/runner.sh:ro" \
      --entrypoint bash \
      "$IMAGE" /runner.sh "$cmd" "$timeline" "$cycles"
    echo "[dosbox.sh] 截圖 → $WORK/shots"
    ;;
  *) sed -n '2,9p' "$0"; exit 1 ;;
esac
