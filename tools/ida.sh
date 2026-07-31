#!/usr/bin/env bash
# IDA Pro 9.4 headless 包裝（CLAUDE.md §4.1：反組譯一律先用 IDA，Ghidra 只當交叉驗證）。
#
#   tools/ida.sh analyze WAR.EXE          產 .i64 + .asm（-A -B）
#   tools/ida.sh script tools/foo.idc WAR.EXE.i64
#   tools/ida.sh raw idat -A -B WAR.EXE   直接下 idat 參數
#
# 工作目錄固定為 workplace/ida/（gitignore）。原版素材唯讀，複製後才分析。
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="ida-pro-9.4-ver2"
WORK="$ROOT/workplace/ida"

mkdir -p "$WORK"

run() {
  docker run --rm \
    -v "$WORK:/work" \
    -v "$ROOT/tools:/work/tools:ro" \
    -w /work \
    "$IMAGE" "$@"
}

case "${1:-}" in
  analyze)
    bin="${2:?用法: tools/ida.sh analyze <檔名（需已在 workplace/ida/）>}"
    [ -f "$WORK/$bin" ] || { echo "[ida.sh] $WORK/$bin 不存在" >&2; exit 1; }
    echo "[ida.sh] SHA-256: $(sha256sum "$WORK/$bin")"
    run idat -A -B "$bin"
    ;;
  script)
    idc="${2:?用法: tools/ida.sh script <idc 路徑> <db>}"
    db="${3:?用法: tools/ida.sh script <idc 路徑> <db>}"
    run idat -A "-S/work/tools/$(basename "$idc")" "$db"
    ;;
  raw)
    shift
    run "$@"
    ;;
  *)
    sed -n '2,12p' "$0"
    exit 1
    ;;
esac
