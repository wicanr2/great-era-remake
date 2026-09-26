#!/usr/bin/env bash
# 確認遊戲與規則原始碼沒有自行引入 cgo。
#
# 這個檢查只針對本專案的 cmd/ 與 internal/；Ebiten 的平台後端是外部依賴，
# 不把依賴內部的 GLFW cgo 檔案誤算成遊戲原始碼。
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

files=$(find cmd internal -type f -name '*.go' -print)
hits=$(printf '%s\n' "$files" | xargs grep -nE \
  '(^[[:space:]]*import[[:space:]]*"C"|^[[:space:]]*"C"[[:space:]]*$|^[[:space:]]*//[[:space:]]*#cgo)' \
  2>/dev/null || true)

if [ -n "$hits" ]; then
	printf '%s\n' "$hits" | sed 's/^/[no-cgo] ✗ /'
	exit 1
fi

printf '[no-cgo] cmd/ 與 internal/ 沒有 import "C" 或 #cgo ✓\n'
