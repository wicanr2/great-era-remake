#!/usr/bin/env bash
# 原版資產 deny-list 掃描（CLAUDE.md §9）。
#
#   tools/deny_scan.sh            掃 git 追蹤中 + 已暫存的檔案
#   tools/deny_scan.sh --all      連工作目錄未追蹤的一起掃
#
# 命中就以非 0 結束。commit 前與發行前都要跑。
set -uo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

# 原版二進位資產：一個 byte 都不該進版控
DENY_EXT='\.(TPC|RGB|MUS|TIM|GLB|GTB|DT1|DT2|SAV|CPS|BGI|EXE|exe)$|\.15$|^MAN[123]15$|(/|^)(TN15\.[0-9]+|[0-9]+\.15)$'
# 倚天字型：不得散布
DENY_FONT='(STDFONT|SPCFONT|ASCFONT|SPCFSUPP)\.'
# 整個原版素材區
DENY_DIR='^workplace/'

if [ "${1:-}" = "--all" ]; then
  files=$(git ls-files --cached --others --exclude-standard)
else
  files=$(git ls-files --cached)
  staged=$(git diff --cached --name-only 2>/dev/null)
  files=$(printf '%s\n%s\n' "$files" "$staged" | sort -u | sed '/^$/d')
fi

hits=$(printf '%s\n' "$files" | grep -E "$DENY_EXT|$DENY_FONT|$DENY_DIR" || true)

# ── 展示截圖的白名單（CLAUDE.md §9 的唯一例外）─────────────────────
#
# README 需要實機畫面才說得清楚進度，而那些畫面是由原版檔案算出來的，
# 必然含原版字模與美術。使用者 2026-08-01 明示放行，但**只放行一個位置**：
#
#     docs/images/*.png
#
# 其他任何地方冒出 PNG 一律報錯——包括 docs/ 的子目錄。這樣「例外」是
# 一份白名單而不是一個漏洞：要新增展示圖必須放進那個目錄，
# 想偷渡別的原版衍生物就會被擋下來。
stray_png=$(printf '%s\n' "$files" | grep -E '\.png$' | grep -vE '^docs/images/[^/]+\.png$' || true)

n=$(printf '%s\n' "$hits" | sed '/^$/d' | wc -l)
m=$(printf '%s\n' "$stray_png" | sed '/^$/d' | wc -l)
total=$(printf '%s\n' "$files" | sed '/^$/d' | wc -l)

if [ "$n" -gt 0 ]; then
  echo "[deny-scan] 命中 $n 個不該進版控的檔案："
  printf '%s\n' "$hits" | sed 's/^/  ✗ /'
  exit 1
fi

if [ "$m" -gt 0 ]; then
  echo "[deny-scan] 命中 $m 個位置錯誤的 PNG（展示截圖只准放 docs/images/）："
  printf '%s\n' "$stray_png" | sed 's/^/  ✗ /'
  echo "  → 若是 README 展示截圖，搬到 docs/images/；否則不該進版控。"
  exit 1
fi

shots=$(printf '%s\n' "$files" | grep -cE '^docs/images/[^/]+\.png$' || true)
echo "[deny-scan] 掃描 $total 個檔案，原版資產零命中 ✓（展示截圖 $shots 張，位置合規）"
