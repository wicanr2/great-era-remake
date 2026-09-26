# Go 建置環境。遊戲本身是純 Go、沒有 cgo 原始碼；目前 Ebiten v2.8.8 的
# Linux desktop GLFW 後端仍需要 cgo 與 X11/OpenGL 開發標頭檔，官方 golang
# image 沒有這些，所以自建一份。
FROM golang:1.25-bookworm

RUN apt-get update && apt-get install -y --no-install-recommends \
      libx11-dev libxrandr-dev libxcursor-dev libxinerama-dev libxi-dev \
      libgl1-mesa-dev libxxf86vm-dev libasound2-dev pkg-config \
      xvfb imagemagick ffmpeg libgl1 libglx-mesa0 mesa-utils x11-utils xdotool \
 && install -d -m 1777 /tmp/.X11-unix \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /work
