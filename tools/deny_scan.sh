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

n=$(printf '%s\n' "$hits" | sed '/^$/d' | wc -l)
total=$(printf '%s\n' "$files" | sed '/^$/d' | wc -l)

if [ "$n" -gt 0 ]; then
  echo "[deny-scan] 命中 $n 個不該進版控的檔案："
  printf '%s\n' "$hits" | sed 's/^/  ✗ /'
  exit 1
fi

echo "[deny-scan] 掃描 $total 個檔案，原版資產零命中 ✓"
