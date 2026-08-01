# Go 建置環境。Ebiten 走 CGO，需要 X11/OpenGL 的開發標頭檔，
# 官方 golang image 沒有這些，所以自建一份。
FROM golang:1.25-bookworm

RUN apt-get update && apt-get install -y --no-install-recommends \
      libx11-dev libxrandr-dev libxcursor-dev libxinerama-dev libxi-dev \
      libgl1-mesa-dev libxxf86vm-dev libasound2-dev pkg-config \
      xvfb imagemagick libgl1 libglx-mesa0 mesa-utils x11-utils \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /work
