#!/usr/bin/env bash
# 在容器內跑：Xvfb + DOSBox + 依 timeline 送鍵截圖。
# 由 tools/dosbox.sh 掛進現成的 dosbox image 執行（不自建 image）。
#
#   $1  autoexec 要跑的指令（如 "play.bat" 或 "GRT.EXE"）
#   $2  timeline，分號分隔：wait:秒 / key:鍵 / type:文字 / train:次數 /
#       shot:名稱 / snap:名稱
#   $3  cycles（預設 fixed 12000）
#
# /orig = 唯讀原版；/game = 可寫副本；/shots = 截圖輸出。
set -uo pipefail

cmd="${1:-play.bat}"
timeline="${2:-wait:6;shot:01-boot;wait:6;shot:02-after;wait:6;shot:03-later}"
cycles="${3:-fixed 12000}"
export DISPLAY=:99

[[ -d /orig && -d /game && -d /shots ]] || {
    echo "缺少 /orig、/game 或 /shots 掛載" >&2
    exit 1
}

# 每輪在容器內重建副本，避免舊存檔與截圖污染 oracle。掛載點本身不刪。
find /game -mindepth 1 -maxdepth 1 -exec rm -rf -- {} +
cp -R /orig/. /game/
chmod -R u+w /game
find /shots -maxdepth 1 -type f -name '*.png' -delete
rm -rf /shots/snapshots
mkdir -p /shots/captures /shots/snapshots /tmp/dsds-home

Xvfb :99 -screen 0 1024x768x24 -nolisten tcp &
xvfb_pid=$!
for _ in $(seq 1 50); do [[ -S /tmp/.X11-unix/X99 ]] && break; sleep 0.1; done
[[ -S /tmp/.X11-unix/X99 ]] || { echo "Xvfb 沒起來" >&2; exit 1; }

# docs/playtest/06 已訂正：舊報告的「keymap 0 列」是因為 xmodmap 不存在。
# 真正修復點是下方 focus_window()；這裡只明確載入並驗證 US keymap。
if ! command -v setxkbmap >/dev/null 2>&1; then
    echo "image 缺少 setxkbmap，無法建立可重現鍵盤環境" >&2
    exit 1
fi
setxkbmap -display :99 us
echo "[probe] keymap=$(setxkbmap -display :99 -query | tr '\n' ' ')"

cleanup() { kill "${dosbox_pid:-}" 2>/dev/null || true; kill "$xvfb_pid" 2>/dev/null || true; }
trap cleanup EXIT

cat >/tmp/dsds.conf <<EOF
[sdl]
fullscreen=false
output=surface
autolock=false
waitonerror=false
[dosbox]
machine=svga_s3
captures=/shots/captures
memsize=32
[render]
frameskip=0
aspect=false
scaler=none
[cpu]
core=auto
cputype=auto
cycles=${cycles}
[mixer]
nosound=true
[midi]
mpu401=none
[sblaster]
sbtype=none
[gus]
gus=false
[speaker]
pcspeaker=false
[joystick]
joysticktype=none
[dos]
xms=true
ems=true
umb=true
keyboardlayout=us
[autoexec]
mount c /game
c:
${cmd}
EOF

dosbox -conf /tmp/dsds.conf >/tmp/dosbox.log 2>&1 &
dosbox_pid=$!

# 等 DOSBox 視窗出現。DOSBox 切換顯示模式時可能重建 X11 視窗，因此後續
# 每次 focus／shot 也會重查，不能永遠握著啟動時的 window id。
window=""
find_window() {
    local found=""
    for _ in $(seq 1 40); do
        found=$(xdotool search --name "DOSBox" 2>/dev/null | tail -1 || true)
        if [[ -n "$found" ]]; then
            window="$found"
            return 0
        fi
        sleep 0.25
    done
    return 1
}
find_window || { echo "找不到 DOSBox 視窗" >&2; cat /tmp/dosbox.log >&2; exit 1; }

# focus_window 把滑鼠移進 DOSBox 視窗中央。
#
# 沒有 WM 時 X 的輸入焦點是 PointerRoot（跟著滑鼠），所以這一步就是
# 「把鍵盤焦點交給 DOSBox」。windowfocus 當備援，有 WM 的環境下也能用。
focus_window() {
    find_window || { echo "DOSBox 視窗已消失" >&2; return 1; }
    xdotool windowraise "$window" 2>/dev/null || true
    xdotool mousemove --window "$window" 320 175 2>/dev/null || true
    xdotool windowfocus "$window" 2>/dev/null || true
    sleep 0.3
}

IFS=';' read -ra steps <<< "$timeline"
for step in "${steps[@]}"; do
    action="${step%%:*}"; arg="${step#*:}"
    case "$action" in
        wait) echo "[probe] wait ${arg}s"; sleep "$arg" ;;
        # [雷] Xvfb 裡沒有 window manager，X 的輸入焦點是 **PointerRoot**
        # （`XGetInputFocus` 回傳 1）。那表示「焦點跟著滑鼠走」，
        # 而滑鼠預設停在 root window 上——所以 XTEST 送出的鍵盤事件
        # 全部進了 root，DOSBox 一個都收不到，畫面完全沒反應。
        #
        # `windowactivate` 也救不了：它需要 WM 支援 `_NET_ACTIVE_WINDOW`，
        # 沒有 WM 就會直接失敗。
        #
        # 修法是**把滑鼠移進視窗**——PointerRoot 模式下這就等於把焦點給它。
        key)  echo "[probe] key $arg"
              focus_window; xdotool key --clearmodifiers "$arg" ;;
        type) echo "[probe] type $arg"
              focus_window; xdotool type --clearmodifiers --delay 120 "$arg" ;;
        train)
            # 正常玩家路徑的「練兵」：1 → 3 → Return → y。遊戲會播放
            # 「十日操練中」動畫，下一組至少等 5 秒；不能把按鍵一次灌完，
            # 否則動畫期間的輸入會被吃掉（docs/playtest/16）。
            [[ "$arg" =~ ^[0-9]+$ && "$arg" -ge 1 && "$arg" -le 30 ]] || {
                echo "train 次數必須是 1..30：$arg" >&2
                exit 2
            }
            echo "[probe] train ${arg} 次"
            for _ in $(seq 1 "$arg"); do
                focus_window; xdotool key --clearmodifiers 1; sleep 1
                focus_window; xdotool key --clearmodifiers 3; sleep 1
                focus_window; xdotool key --clearmodifiers Return; sleep 1
                focus_window; xdotool key --clearmodifiers y; sleep 5
            done
            ;;
        shot)
            find_window || { echo "截圖時找不到 DOSBox 視窗" >&2; exit 1; }
            # 視窗幾何要跟截圖一起記：DOSBox 視窗大小 = 遊戲當下的顯示模式，
            # 光看截圖內容 trim 出來的 bounding box 會被黑邊誤導。
            geom=$(xdotool getwindowgeometry --shell "$window" 2>/dev/null \
                   | tr '\n' ' ' || true)
            echo "[probe] shot $arg  ($geom)"
            import -window root "/shots/${arg}.png"
            ;;
        snap)
            # 保存這一刻的狀態檔副本，供戰前／戰中 byte diff。原版仍在 /orig 唯讀掛載；
            # 這裡只複製 /game 的工作副本，而且每輪開跑已清除舊 snapshots。
            dest="/shots/snapshots/${arg}"
            mkdir -p "$dest"
            find /game -maxdepth 1 -type f \
              \( -iname '*.dt1' -o -iname '*.dt2' -o -iname 'mem_war.dat' \
                 -o -iname 'config.dat' -o -iname 'placd.sav' \) \
              -exec cp -p {} "$dest/" \;
            echo "[probe] snap $arg"
            find "$dest" -maxdepth 1 -type f -exec sha256sum {} \; | sort
            ;;
        *) echo "未知步驟: $step" >&2; exit 2 ;;
    esac
done

echo "=== DOSBox log ==="
cat /tmp/dosbox.log
echo "=== 遊戲目錄異動（磁片檢查／存檔痕跡）==="
find /game -newer /tmp/dsds.conf -type f 2>/dev/null | head -20
