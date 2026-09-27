#!/usr/bin/env bash
# 從原創純 Go 程序音樂產生 Modern Ogg：WAV 渲染 → libvorbis → manifest。
#
#   tools/make_modern_ogg.sh [輸出目錄，預設 assets/music/modern]
#
# 渲染與轉檔全部在同一個 dsds-go:1.25 容器內執行（Docker-only，
# --network none，一次性 --rm）；主機只做雜湊與 manifest 組裝。
# 來源是 internal/ui/audio 的原創程序作曲（六首 cue；8 類效果音不轉），
# 不含原版 .MUS/.TIM 旋律、不含外部音源，provenance 見輸出 manifest.json。
# 正式作曲署名／盲聽／人耳混音 QA 仍未完成（見 docs/design/32），
# manifest 版本字串如實標 tech-preview。
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT="${1:-$ROOT/assets/music/modern}"
IMAGE="${DSDS_GO_IMAGE:-dsds-go:1.25}"
RATE=44100

case "$OUT" in
"$ROOT"/*) ;;
*) echo "[ogg] 輸出目錄須在工作樹內：$OUT" >&2; exit 2 ;;
esac
mkdir -p "$OUT"
expected="$(id -u):$(id -g)"
owner="$(stat -c '%u:%g' "$OUT")"
if [[ "$owner" != "$expected" ]]; then
  echo "[ogg] 拒絕寫入非目前使用者擁有的目錄：$OUT" >&2; exit 1
fi

CACHE="$ROOT/workplace/.gocache"
mkdir -p "$CACHE/pkg" "$CACHE/build"
OUT_REL="${OUT#"$ROOT"/}"

docker run --rm --network none --memory 2g --cpus 2 --pids-limit 512 \
  -v "$ROOT:/work" \
  -v "$CACHE/pkg:/go/pkg" \
  -v "$CACHE/build:/.cache/go-build" \
  -u "$expected" -w /work \
  -e GOCACHE=/.cache/go-build -e GOFLAGS=-buildvcs=false \
  "$IMAGE" bash -eu -o pipefail -c '
    out_rel="'"$OUT_REL"'"
    rate="'"$RATE"'"
    mkdir -p "$out_rel" /tmp/oggwav
    for track in scene strategy battle-1 battle-2 wall final; do
      go run ./cmd/modern_audio -out /tmp/oggwav -sample-rate "$rate" \
        -track "$track" -effects=false >/dev/null
    done
    enc() { ffmpeg -hide_banner -loglevel error -i "$1" \
      -c:a libvorbis -q:a 4 -ar "$rate" -y "$2"; }
    enc /tmp/oggwav/scene.wav    "$out_rel/modern_scene.ogg"
    enc /tmp/oggwav/strategy.wav "$out_rel/modern_strategy.ogg"
    enc /tmp/oggwav/battle-1.wav "$out_rel/modern_battle_a.ogg"
    enc /tmp/oggwav/battle-2.wav "$out_rel/modern_battle_b.ogg"
    enc /tmp/oggwav/wall.wav     "$out_rel/modern_story.ogg"
    enc /tmp/oggwav/final.wav    "$out_rel/modern_final.ogg"
    rm -rf /tmp/oggwav
    ffmpeg -hide_banner -loglevel error -version | head -1
  '

# 主機只做雜湊與 manifest 組裝（不碰音訊位元組）
python3 - "$OUT" <<'EOF'
import hashlib, json, sys
out = sys.argv[1]
tracks = {}
for cue in ["modern_scene", "modern_strategy", "modern_battle_a",
            "modern_battle_b", "modern_story", "modern_final"]:
    fn = cue + ".ogg"
    with open(f"{out}/{fn}", "rb") as f:
        digest = hashlib.sha256(f.read()).hexdigest()
    tracks[cue] = {
        "file": fn,
        "loop_start": 0,
        "loop_length": 0,
        "author": "great-era-remake procedural generator (pure-Go original "
                  "composition; formal composer credit and blind-listening QA pending)",
        "license": "LicenseRef-RRSAL-1.0 (see repo LICENSE)",
        "sha256": digest,
    }
manifest = {"schema": 1, "version": "modern-music-tech-preview-2026-09-27",
            "tracks": tracks}
with open(f"{out}/manifest.json", "w", encoding="utf-8") as f:
    json.dump(manifest, f, ensure_ascii=False, indent=2)
    f.write("\n")
print(f"[ogg] 寫出 {out}/manifest.json，六首 cue")
EOF
