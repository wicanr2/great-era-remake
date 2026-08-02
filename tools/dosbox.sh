#!/usr/bin/env bash
# DOSBox oracle 包裝（CLAUDE.md §4：DOSBox 實測是最高位階的 oracle）。
#
#   tools/dosbox.sh probe                     跑 play.bat，等間隔截三張
#   tools/dosbox.sh run "GRT.EXE" "wait:8;shot:a;key:Return;wait:3;shot:b;snap:after"
#
# 沿用現成的 dosbox image（含 Xvfb / xdotool / ImageMagick），不自建。
# 原版素材唯讀掛載；runner 在容器內複製到可寫的 drive_c，不動原始檔。
# 截圖輸出在 workplace/dosbox/shots/。
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="${DSDS_DOSBOX_IMAGE:-fd2-dosbox-screenshot-local}"
WORK="$ROOT/workplace/dosbox"
ORIG="$ROOT/workplace/orig/game"

mkdir -p "$WORK/drive_c" "$WORK/shots"

for path in "$WORK" "$WORK/drive_c" "$WORK/shots"; do
  owner="$(stat -c '%u:%g' "$path")"
  expected="$(id -u):$(id -g)"
  if [[ "$owner" != "$expected" ]]; then
    echo "[dosbox.sh] 拒絕寫入非目前使用者擁有的路徑：$path（$owner，預期 $expected）" >&2
    exit 1
  fi
done

cmd="${2:-play.bat}"
timeline="${3:-wait:6;shot:01-boot;wait:6;shot:02-after;wait:8;shot:03-later}"
cycles="${4:-fixed 12000}"
limit="${DSDS_DOSBOX_TIMEOUT:-240}"

case "${1:-probe}" in
  probe|run)
    timeout --signal=TERM --kill-after=10s "$limit" \
    docker run --rm --network none \
      --memory 768m --cpus 1 --pids-limit 256 \
      -u "$(id -u):$(id -g)" \
      --read-only --tmpfs /tmp:rw,nosuid,nodev,size=64m \
      -e HOME=/tmp/dsds-home \
      -v "$ORIG:/orig:ro" \
      -v "$WORK/drive_c:/game" \
      -v "$WORK/shots:/shots" \
      -v "$ROOT/tools/dosbox_runner.sh:/runner.sh:ro" \
      --entrypoint bash \
      "$IMAGE" /runner.sh "$cmd" "$timeline" "$cycles"
    echo "[dosbox.sh] 截圖 → $WORK/shots"
    ;;
  *) sed -n '2,9p' "$0"; exit 1 ;;
esac
