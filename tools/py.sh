#!/usr/bin/env bash
# Python 執行包裝：一律 docker + uv venv，不裝進系統環境。
#
#   tools/py.sh --build              重建 image（改 pyproject.toml 後要跑）
#   tools/py.sh tools/glyph_lookup.py --help
#   tools/py.sh -m pytest
#
# venv 放在 workplace/.venv-py（gitignore），跨次執行持久化。
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="dsds-py:3.12"
VENV_VOL="$ROOT/workplace/.venv-py"

build() {
  docker build -f "$ROOT/tools/Dockerfile.py" -t "$IMAGE" "$ROOT/tools"
  mkdir -p "$VENV_VOL" "$ROOT/workplace/.uvcache"
  # 依 pyproject.toml 同步套件到持久化的 venv
  if [ -f "$ROOT/pyproject.toml" ]; then
    docker run --rm \
      -v "$ROOT:/work" -v "$VENV_VOL:/venv" \
      -v "$ROOT/workplace/.uvcache:/uvcache" \
      -u "$(id -u):$(id -g)" -w /work \
      -e HOME=/tmp -e UV_CACHE_DIR=/uvcache \
      "$IMAGE" uv sync --no-dev
  fi
}

if [ "${1:-}" = "--build" ]; then
  build
  echo "[py.sh] image $IMAGE 就緒，venv → $VENV_VOL"
  exit 0
fi

if ! docker image inspect "$IMAGE" >/dev/null 2>&1; then
  echo "[py.sh] image 不存在，先建一次…" >&2
  build
fi

mkdir -p "$VENV_VOL"
exec docker run --rm -i \
  -v "$ROOT:/work" \
  -v "$VENV_VOL:/venv" \
  -u "$(id -u):$(id -g)" \
  -w /work \
  -e HOME=/tmp \
  "$IMAGE" python3 "$@"
