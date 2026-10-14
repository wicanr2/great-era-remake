#!/usr/bin/env bash
# Docker-only 原創程序音樂預覽。正式作者、聽審與發行權利仍由 Issue #5 追蹤。
set -euo pipefail
audio_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
audio_out="${1:-$audio_root/assets/music/modern}"
audio_image="${DSDS_GO_IMAGE:-logh3-hd-preview:20261011-r1}"
audio_modules="${DSDS_GO_MODULES:-$audio_root/workplace/.gocache/pkg}"
test -d "$audio_modules"
test -f "$audio_root/tools/music/modern_ogg.py"
test -f "$audio_root/LICENSE"
# Only directory creation, ownership checks and Docker run on the host.
case "$audio_out" in /*) ;; *) audio_out="$PWD/$audio_out" ;; esac
case "$audio_out" in "$audio_root"/*) ;; *) echo '音訊輸出須在隔離工作樹內。' >&2; exit 2 ;; esac
mkdir -p -- "$audio_out"
audio_out="$(cd -- "$audio_out" && pwd)"
case "$audio_out" in "$audio_root"/*) ;; *) echo '音訊輸出須在隔離工作樹內。' >&2; exit 2 ;; esac
audio_owner="$(id -u):$(id -g)"
test "$(stat -c '%u:%g' "$audio_out")" = "$audio_owner"
exec timeout --foreground 240 docker run --rm --network none \
  --memory 3g --cpus 3 --pids-limit 256 --user "$audio_owner" \
  -v "$audio_root:/work:ro" -v "$audio_out:/audio-out:rw" \
  -v "$audio_modules:/go/pkg:ro" -w /work \
  -e GOCACHE=/tmp/audio-go-cache -e GOPROXY=off -e PYTHONUNBUFFERED=1 \
  --entrypoint /bin/bash "$audio_image" -c '
    set -euo pipefail
    Xvfb :97 -screen 0 1280x720x24 -nolisten tcp >/tmp/audio-xvfb.log 2>&1 &
    audio_xvfb_pid=$!
    trap '\''kill "$audio_xvfb_pid" 2>/dev/null || true'\'' EXIT
    export DISPLAY=:97
    python3 tools/music/modern_ogg.py /audio-out
  '
