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

IFS=';' read -ra steps <<< "$timeline"
for step in "${steps[@]}"; do
    action="${step%%:*}"; arg="${step#*:}"
    case "$action" in
        wait) echo "[probe] wait ${arg}s"; sleep "$arg" ;;
        # [雷] DOSBox 忽略 XSendEvent，所以不能用 `xdotool key --window`——
        # 那條路徑送出去的事件會被丟掉，畫面完全沒反應。
        # 要走 XTEST（不帶 --window），而且視窗必須先 activate 而不只是 focus。
        key)  echo "[probe] key $arg"
              xdotool windowactivate --sync "$window" 2>/dev/null || xdotool windowfocus "$window"
              sleep 0.2; xdotool key --clearmodifiers "$arg" ;;
        type) echo "[probe] type $arg"
              xdotool windowactivate --sync "$window" 2>/dev/null || xdotool windowfocus "$window"
              sleep 0.2; xdotool type --clearmodifiers --delay 120 "$arg" ;;
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
