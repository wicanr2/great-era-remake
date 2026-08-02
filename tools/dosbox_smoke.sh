#!/usr/bin/env bash
# DOSBox oracle 最小 smoke：證明 Xvfb 焦點與按鍵注入仍可從主選單進到載入選單。
# 畫面語意仍須人工檢視；本腳本只提供可重跑的機械 gate。
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SHOTS="$ROOT/workplace/dosbox/shots"

bash "$ROOT/tools/dosbox.sh" run "WAR.EXE" \
  "wait:10;shot:smoke-0-title;key:Return;wait:2;shot:smoke-1-after-return;key:2;wait:2;shot:smoke-2-load-choice;key:Return;wait:2;shot:smoke-3-save-menu"

docker run --rm --network none \
  --memory 128m --cpus 1 --pids-limit 64 \
  -u "$(id -u):$(id -g)" \
  -v "$SHOTS:/shots:ro" \
  alpine:3.20 sh -eu -c '
    for f in \
      /shots/smoke-0-title.png \
      /shots/smoke-1-after-return.png \
      /shots/smoke-2-load-choice.png \
      /shots/smoke-3-save-menu.png
    do
      test -s "$f" || { echo "[dosbox-smoke] 缺少或空白：$f" >&2; exit 1; }
    done

    h0=$(sha256sum /shots/smoke-0-title.png | cut -d " " -f 1)
    h1=$(sha256sum /shots/smoke-1-after-return.png | cut -d " " -f 1)
    h2=$(sha256sum /shots/smoke-2-load-choice.png | cut -d " " -f 1)
    h3=$(sha256sum /shots/smoke-3-save-menu.png | cut -d " " -f 1)
    expected_title=ead5b9243736cd7d455d8f4810ea810b27c53eeb2b00b9b45f5cae139554875b
    expected_transition=7d459121c1731da65d550d2a0252c66c7bc609b9efb2789aa1e47d2b7594d638
    expected_menu_cursor_off=582963e04bc3f0686717775b47c07d1da2c7088e88a63ad56e727537c2096629
    expected_menu_cursor_on=5ee11b2371482026358ae9cb64fbf7167a249152d20f263c0d32a2d892304b3d

    test "$h0" = "$h1" || {
      echo "[dosbox-smoke] Return 前後不應換頁：$h0 != $h1" >&2
      exit 1
    }
    test "$h1" != "$h2" || {
      echo "[dosbox-smoke] 按 2 後畫面未變，輸入可能失效" >&2
      exit 1
    }
    test "$h2" != "$h3" || {
      echo "[dosbox-smoke] 按 Return 後畫面未變，輸入可能失效" >&2
      exit 1
    }
    test "$h0" = "$expected_title" || {
      echo "[dosbox-smoke] 主選單與已人工確認的 oracle 不同：$h0" >&2
      exit 1
    }
    test "$h2" = "$expected_transition" || {
      echo "[dosbox-smoke] 載入轉場與已人工確認的 oracle 不同：$h2" >&2
      exit 1
    }
    test "$h3" = "$expected_menu_cursor_off" || test "$h3" = "$expected_menu_cursor_on" || {
      echo "[dosbox-smoke] 進度選單與已人工確認的 oracle 不同：$h3" >&2
      exit 1
    }

    echo "[dosbox-smoke] PASS"
    printf "  title=%s\n  load-transition=%s\n  save-menu=%s\n" "$h0" "$h2" "$h3"
  '

echo "[dosbox-smoke] 請人工檢視：$SHOTS/smoke-3-save-menu.png"
