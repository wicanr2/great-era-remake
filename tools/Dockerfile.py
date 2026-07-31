# 本專案專用的 Python 工具環境（uv venv，不污染系統）
# build: tools/py.sh --build
FROM python:3.12-slim-bookworm

RUN pip install --no-cache-dir uv

WORKDIR /work

# venv 建在容器內固定路徑，由 tools/py.sh 掛 volume 持久化
ENV UV_PROJECT_ENVIRONMENT=/venv \
    PATH="/venv/bin:$PATH" \
    PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1

CMD ["python3"]
