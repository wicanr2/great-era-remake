#!/usr/bin/env bash
# 以同一個 Docker／Xvfb 視窗錄製 A2 M2 民國測繪版與人物肖像的連續玩家路徑。
#
#   bash tools/capture_h4_promo.sh [輸出目錄]
#
# 輸出必須位於 dist-all/ 之下，且腳本拒絕覆蓋既有影片。原版資料與字庫只
# 唯讀掛載到錄影容器；最終交付只含 MP4、技術 QA JSON、執行紀錄與雜湊，
# 不帶原版檔案、字庫、WAV 或任何可抽取音源。
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="${DSDS_PROMO_IMAGE:-dsds-go:1.25}"
OUT="${1:-$ROOT/dist-all/2026-08-12-a2-m2-portraits/promo}"
CACHE="$ROOT/workplace/.gocache"
OVERWRITE="${DSDS_PROMO_OVERWRITE:-0}"

if [[ "$OUT" != /* ]]; then
	OUT="$ROOT/$OUT"
fi
case "$OUT" in
	"$ROOT"/dist-all/*) ;;
	*) echo "[promo] 輸出必須位於 dist-all/：$OUT" >&2; exit 2 ;;
esac
case "$OVERWRITE" in 0|1) ;; *) echo "[promo] DSDS_PROMO_OVERWRITE 只能是 0 或 1" >&2; exit 2 ;; esac

mkdir -p "$OUT" "$CACHE/pkg" "$CACHE/build"
expected="$(id -u):$(id -g)"
for path in "$OUT" "$CACHE" "$CACHE/pkg" "$CACHE/build"; do
	owner="$(stat -c '%u:%g' "$path")"
	if [[ "$owner" != "$expected" ]]; then
		echo "[promo] 拒絕寫入非目前使用者擁有的路徑：$path（$owner，預期 $expected）" >&2
		exit 1
	fi
done

video="great-era-remake-a2-m2-portrait-playthrough.mp4"
video_only="great-era-remake-a2-m2-portrait-video-only.mp4"
audio_qa="modern-audio-qa.json"
run_log="capture-game.log"
sums="SHA256SUMS-a2-m2-portrait.txt"
for name in "$video" "$video_only" "$audio_qa" "$run_log" "$sums"; do
	if [[ -e "$OUT/$name" && "$OVERWRITE" != 1 ]]; then
		echo "[promo] 拒絕覆蓋既有產物：$OUT/$name（若確認需要，設 DSDS_PROMO_OVERWRITE=1）" >&2
		exit 5
	fi
done

STAGE="$(mktemp -d "$OUT/.dsds-promo-stage.XXXXXX")"
stage_base="$(basename "$STAGE")"
cleanup_stage() {
	docker run --rm --network none --memory 256m --cpus 0.5 --pids-limit 64 \
		-v "$OUT:/dist:rw" -u "$expected" "$IMAGE" bash -eu -c '
			case "$1" in .dsds-promo-stage.*) ;; *) exit 2 ;; esac
			rm -rf -- "/dist/$1"
		' bash "$stage_base" >/dev/null 2>&1 || true
}
trap cleanup_stage EXIT

docker run --rm --network none --memory 2g --cpus 2 --pids-limit 512 \
	-v "$ROOT:/src:ro" \
	-v "$CACHE/pkg:/go/pkg:rw" \
	-v "$CACHE/build:/.cache/go-build:rw" \
	-v "$STAGE:/stage:rw" \
	-v "$OUT:/dist:rw" \
	-u "$expected" -w /src \
	-e GOCACHE=/.cache/go-build -e GOFLAGS='-buildvcs=false -mod=readonly' \
	"$IMAGE" bash -eu -o pipefail -c '
		/usr/local/go/bin/go build -trimpath -buildvcs=false -o /stage/dsds ./cmd/dsds
		/usr/local/go/bin/go run ./cmd/modern_audio \
			-out /stage/audio -report /stage/audio/qa-report.json -track main-theme -effects=false

		Xvfb :99 -screen 0 1280x720x24 -nolisten tcp >/stage/xvfb.log 2>&1 &
		xvfb_pid=$!
		game_pid=""
		ffmpeg_pid=""
		cleanup() {
			if [ -n "$ffmpeg_pid" ] && kill -0 "$ffmpeg_pid" 2>/dev/null; then kill -INT "$ffmpeg_pid" 2>/dev/null || true; fi
			if [ -n "$game_pid" ] && kill -0 "$game_pid" 2>/dev/null; then kill "$game_pid" 2>/dev/null || true; fi
			if kill -0 "$xvfb_pid" 2>/dev/null; then kill "$xvfb_pid" 2>/dev/null || true; fi
		}
		trap cleanup EXIT
		export DISPLAY=:99
		# 無家目錄的數值 UID 在容器中會使 X11 client 不斷輸出無意義的
		# .Xauthority 警告；錄影不需要 X 授權，但提供空的私有檔案可讓診斷
		# 紀錄只保留真正與遊戲相關的訊息。
		export HOME=/stage
		: > /stage/.Xauthority
		sleep 1
		kill -0 "$xvfb_pid"

		/stage/dsds -game /src/workplace/orig/game -eten /src/workplace/eten \
			-locale /src/translations/zh-Hant -province 36 -theme retro -resolution original \
			-audio off -save /stage/SAVE\(1\).DT1 -prefs /stage/prefs.json >/stage/game.log 2>&1 &
		game_pid=$!
		for _ in $(seq 1 40); do
			win="$(xdotool search --onlyvisible --name ".*" 2>/dev/null | tail -n 1 || true)"
			[ -n "$win" ] && break
			sleep 0.25
		done
		test -n "${win:-}"
		sleep 2

		ffmpeg -y -f x11grab -video_size 1280x720 -framerate 30 -i :99.0 -an \
			-c:v libx264 -preset veryfast -threads 2 -crf 20 -pix_fmt yuv420p -movflags +faststart \
			/stage/video-only.mp4 >/stage/ffmpeg-capture.log 2>&1 &
		ffmpeg_pid=$!
		sleep 2

		press() {
			# Ebiten 在 Xvfb 下只可靠接收目前取得焦點的鍵盤事件；直接指定
			# --window 的 synthetic key event 可能被 X11 遞送到舊視窗佇列，導致
			# 影片看似在按鍵、實際卻停在復古畫面。先把焦點交給同一個遊戲視窗，
			# 再送全域鍵盤事件，並保留足夠長的按下時間供 IsKeyJustPressed 擷取。
			xdotool windowfocus "$win" 2>/dev/null || true
			sleep 0.15
			xdotool keydown "$1"
			sleep 0.22
			xdotool keyup "$1"
			sleep 0.60
		}
		press_fast() {
			xdotool windowfocus "$win" 2>/dev/null || true
			sleep 0.08
			xdotool keydown "$1"
			sleep 0.12
			xdotool keyup "$1"
			sleep 0.20
		}
		click_at() {
			xdotool windowfocus "$win" 2>/dev/null || true
			sleep 0.15
			xdotool mousemove --window "$win" "$1" "$2"
			sleep 0.15
			xdotool mousedown 1
			sleep 0.25
			xdotool mouseup 1
			sleep 0.75
		}

		# 同一條真實玩家路徑：復古地圖 → Modern → 高解析 A2 M2 → 查閱／人物
		# 檔案 → 轉往可重播的河南攻防 → 滑鼠進攻擊選單與結束回合。
		sleep 2
		press F2
		sleep 2
		press F3
		sleep 2
		click_at 1040 592
		sleep 2
		click_at 220 356
		sleep 2
		click_at 395 489
		sleep 2
		press Return
		sleep 2
		press b
		sleep 4
		press b
		press Escape
		press Escape
		press Escape
		press Escape
		for _ in $(seq 1 17); do press_fast Left; done
		sleep 2
		press Return
		press a
		sleep 4
		click_at 1028 366
		sleep 3
		press Escape
		click_at 1172 578
		sleep 5

		# ffmpeg 收到 SIGINT 會正確寫完 MP4 的 moov atom，但慣例上回傳
		# 非零。先等它完成封箱，再以實際檔案與 ffprobe 當成功判定，不能讓
		# set -e 在正常停止時中斷並留下無法播放的暫存 MP4。
		kill -INT "$ffmpeg_pid"
		wait "$ffmpeg_pid" || true
		ffmpeg_pid=""
		test -s /stage/video-only.mp4

		# 只有 51 秒的音樂可循環；畫面必須維持單次連續錄影，不能跟著
		# 無限循環，否則 -shortest 會永遠等不到結尾。
		ffmpeg -y -i /stage/video-only.mp4 -stream_loop -1 \
			-i /stage/audio/main-theme.wav -map 0:v:0 -map 1:a:0 -shortest \
			-c:v copy -c:a aac -b:a 192k -movflags +faststart /stage/playthrough.mp4 \
			>/stage/ffmpeg-mux.log 2>&1
		test -s /stage/playthrough.mp4

		install -m 0644 /stage/video-only.mp4 /dist/$2
		install -m 0644 /stage/playthrough.mp4 /dist/$1
		install -m 0644 /stage/audio/qa-report.json /dist/$3
		install -m 0644 /stage/game.log /dist/$4
		(
			cd /dist
			sha256sum "$1" "$2" "$3" "$4"
		) > /dist/$5

		ffprobe -v error -select_streams v:0 -show_entries stream=codec_name,width,height,r_frame_rate \
			-of default=nokey=1:noprint_wrappers=1 /dist/$1 > /stage/video-probe.txt
		grep -qx h264 /stage/video-probe.txt
		grep -qx 1280 /stage/video-probe.txt
		grep -qx 720 /stage/video-probe.txt
		grep -qx 30/1 /stage/video-probe.txt
		ffprobe -v error -select_streams a:0 -show_entries stream=codec_name,sample_rate,channels \
			-of default=nokey=1:noprint_wrappers=1 /dist/$1 > /stage/audio-probe.txt
		grep -qx aac /stage/audio-probe.txt
		grep -qx 48000 /stage/audio-probe.txt
		grep -qx 2 /stage/audio-probe.txt
	' bash "$video" "$video_only" "$audio_qa" "$run_log" "$sums"

for path in "$OUT/$video" "$OUT/$video_only" "$OUT/$audio_qa" "$OUT/$run_log" "$OUT/$sums"; do
	test -s "$path"
	owner="$(stat -c '%u:%g' "$path")"
	if [[ "$owner" != "$expected" ]]; then
		echo "[promo] 產物擁有權錯誤：$path（$owner，預期 $expected）" >&2
		exit 1
	fi
done

echo "[promo] 完成：$OUT/$video"
echo "[promo] 同一視窗連續錄影；Modern 配樂為本機技術預覽，未納入 release archive。"
