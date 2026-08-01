#!/usr/bin/env bash
# 在容器內跑：Xvfb + DOSBox + 依 timeline 送鍵截圖。
# 由 tools/dosbox.sh 掛進現成的 dosbox image 執行（不自建 image）。
#
#   $1  autoexec 要跑的指令（如 "play.bat" 或 "GRT.EXE"）
#   $2  timeline，分號分隔：wait:秒 / key:鍵 / type:文字 / shot:名稱
#   $3  cycles（預設 fixed 12000）
#
# /game = 可寫的遊戲副本；/shots = 截圖輸出。原版目錄不掛進來。
set -uo pipefail

cmd="${1:-play.bat}"
timeline="${2:-wait:6;shot:01-boot;wait:6;shot:02-after;wait:6;shot:03-later}"
cycles="${3:-fixed 12000}"
export DISPLAY=:99

mkdir -p /shots /shots/captures
Xvfb :99 -screen 0 1024x768x24 -nolisten tcp &
xvfb_pid=$!
for _ in $(seq 1 50); do [[ -S /tmp/.X11-unix/X99 ]] && break; sleep 0.1; done
[[ -S /tmp/.X11-unix/X99 ]] || { echo "Xvfb 沒起來" >&2; exit 1; }

# [雷] Xvfb 起來時**沒有鍵盤映射表**（`xmodmap -pk` 輸出 0 行）。
# 沒有 keymap，xdotool 送出的鍵盤事件產不出有效的 keycode，
# DOSBox 收到也解讀不了——畫面完全沒反應，而且 xdotool 不會報錯。
setxkbmap -display :99 us 2>/dev/null || true
keymap_rows=$(xmodmap -display :99 -pk 2>/dev/null | wc -l)
echo "[probe] keymap 載入 ${keymap_rows} 列"

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

# 等 DOSBox 視窗出現
window=""
for _ in $(seq 1 40); do
    window=$(xdotool search --name "DOSBox" 2>/dev/null | head -1 || true)
    [[ -n "$window" ]] && break
    sleep 0.25
done
[[ -n "$window" ]] || { echo "找不到 DOSBox 視窗" >&2; cat /tmp/dosbox.log >&2; exit 1; }

# focus_window 把滑鼠移進 DOSBox 視窗中央。
#
# 沒有 WM 時 X 的輸入焦點是 PointerRoot（跟著滑鼠），所以這一步就是
# 「把鍵盤焦點交給 DOSBox」。windowfocus 當備援，有 WM 的環境下也能用。
focus_window() {
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
        shot)
            # 視窗幾何要跟截圖一起記：DOSBox 視窗大小 = 遊戲當下的顯示模式，
            # 光看截圖內容 trim 出來的 bounding box 會被黑邊誤導。
            geom=$(xdotool getwindowgeometry --shell "$window" 2>/dev/null \
                   | tr '\n' ' ' || true)
            echo "[probe] shot $arg  ($geom)"
            import -window root "/shots/${arg}.png"
            ;;
        *) echo "未知步驟: $step" >&2; exit 2 ;;
    esac
done

echo "=== DOSBox log ==="
cat /tmp/dosbox.log
echo "=== 遊戲目錄異動（磁片檢查／存檔痕跡）==="
find /game -newer /tmp/dsds.conf -type f 2>/dev/null | head -20
