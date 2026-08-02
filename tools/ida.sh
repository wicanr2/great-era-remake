#!/usr/bin/env bash
# IDA Pro 9.4 headless 包裝（CLAUDE.md §4.1：反組譯一律先用 IDA，Ghidra 只當交叉驗證）。
#
#   tools/ida.sh analyze WAR.EXE          產 .i64 + .asm（-A -B）
#   tools/ida.sh query tools/foo.idc WAR.EXE.i64 [IDC 參數...]
#       從同名 EXE 建立一次性 DB，匯出到 workplace/ida/user-output/
#   tools/ida.sh script tools/foo.idc WAR.EXE.i64 [IDC 參數...]
#       query 的相容別名
#   tools/ida.sh raw idat -A -B WAR.EXE   直接下 idat 參數
#
# 工作目錄固定為 workplace/ida/（gitignore）。原版素材唯讀，複製後才分析。
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="ida-pro-9.4-ver3"
WORK="$ROOT/workplace/ida"

mkdir -p "$WORK"

run() {
  docker run --rm \
    --network none \
    --memory 2g \
    --cpus 2 \
    --pids-limit 256 \
    -u "$(id -u):$(id -g)" \
    -v "$WORK:/work" \
    -v "$ROOT/tools:/work/tools:ro" \
    -w /work \
    "$IMAGE" "$@"
}

cleanup_query_dir() {
  local dir="${1:-}"
  [ -n "$dir" ] && [ -d "$dir" ] || return 0
  docker run --rm \
    --network none --memory 128m --cpus 1 --pids-limit 32 \
    -u "$(id -u):$(id -g)" \
    -v "$dir:/task" alpine:3.20 sh -lc 'find /task -mindepth 1 -delete'
  rmdir "$dir"
}

query() {
  local idc="${1:?query 需要 IDC 路徑}"
  local db="${2:?query 需要 .i64 名稱}"
  shift 2

  local base exe
  base="$(basename "$db")"
  exe="${base%.i64}"
  [ "$exe" != "$base" ] || { echo "[ida.sh] query 的 DB 必須是 .i64" >&2; exit 2; }
  [ -f "$WORK/$exe" ] || {
    echo "[ida.sh] 找不到可重建 DB 的 $WORK/$exe" >&2
    exit 2
  }
  [ -f "$ROOT/$idc" ] || { echo "[ida.sh] $ROOT/$idc 不存在" >&2; exit 2; }

  local query_dir
  query_dir="$(mktemp -d /tmp/dsds-ida-query-XXXXXX)"
  trap 'cleanup_query_dir "$query_dir"' EXIT

  docker run --rm \
    --network none --memory 128m --cpus 1 --pids-limit 32 \
    -u "$(id -u):$(id -g)" \
    -v "$WORK/$exe:/input/$exe:ro" -v "$query_dir:/query" \
    alpine:3.20 cp "/input/$exe" "/query/$exe"

  docker run --rm \
    --network none --memory 2g --cpus 2 --pids-limit 256 \
    -u "$(id -u):$(id -g)" \
    -v "$query_dir:/work" -w /work \
    "$IMAGE" idat -A -B "$exe"

  local script_spec arg
  script_spec="-S/work/tools/$(basename "$idc")"
  for arg in "$@"; do script_spec+=" $arg"; done
  docker run --rm \
    --network none --memory 2g --cpus 2 --pids-limit 256 \
    -u "$(id -u):$(id -g)" \
    -v "$query_dir:/work" -v "$ROOT/tools:/work/tools:ro" -w /work \
    "$IMAGE" idat -A "$script_spec" "$base"

  docker run --rm \
    --network none --memory 128m --cpus 1 --pids-limit 32 \
    -u "$(id -u):$(id -g)" -v "$WORK:/out" \
    alpine:3.20 mkdir -p /out/user-output
  docker run --rm \
    --network none --memory 128m --cpus 1 --pids-limit 32 \
    -u "$(id -u):$(id -g)" \
    -v "$query_dir:/query:ro" -v "$WORK/user-output:/out" \
    alpine:3.20 sh -lc \
    'find /query -maxdepth 1 -type f -name "*.txt" -exec cp {} /out/ \;'
  docker run --rm \
    --network none --memory 128m --cpus 1 --pids-limit 32 \
    -u "$(id -u):$(id -g)" -v "$WORK/user-output:/out:ro" \
    alpine:3.20 sh -lc 'find /out -maxdepth 1 -type f -exec stat -c "%u:%g %a %n" {} \;'
  cleanup_query_dir "$query_dir"
  trap - EXIT
}

case "${1:-}" in
  analyze)
    bin="${2:?用法: tools/ida.sh analyze <檔名（需已在 workplace/ida/）>}"
    [ -f "$WORK/$bin" ] || { echo "[ida.sh] $WORK/$bin 不存在" >&2; exit 1; }
    echo "[ida.sh] SHA-256: $(sha256sum "$WORK/$bin")"
    run idat -A -B "$bin"
    ;;
  query|script)
    idc="${2:?用法: tools/ida.sh script <idc 路徑> <db>}"
    db="${3:?用法: tools/ida.sh script <idc 路徑> <db>}"
    shift 3
    query "$idc" "$db" "$@"
    ;;
  raw)
    shift
    for arg in "$@"; do
      case "$arg" in
        *.i64)
          echo "[ida.sh] 拒絕直接對 .i64 執行 raw；請用 query 從同名 EXE 建立一次性 DB" >&2
          exit 2
          ;;
      esac
    done
    run "$@"
    ;;
  *)
    sed -n '2,12p' "$0"
    exit 1
    ;;
esac
