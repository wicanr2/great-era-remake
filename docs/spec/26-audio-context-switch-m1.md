# SPEC-26：情境音樂串接 M1

狀態：**READY**  
日期：2026-08-11

## 1. 目的

把 `MUS/TIM → 純 Go OPL2 風格 PCM → Ebiten audio.Player` 從開局單曲擴充成
情境曲目庫。曲目仍由玩家自備原版檔案提供；缺少非初始曲目時沿用目前曲目，
不讓音訊缺檔阻斷遊戲。

## 2. 曲目與切換

`internal/ui/audio.Track` 對應 `SCENE`、`STRATEGY`、`MAINTHEM`、`BATTLE1`、
`BATTLE2`、`BT02`、`WALL`、`FINAL`。`SCENE` 與其 `.MUS/.TIM` 是 `retro` 模式的
必要初始曲；其餘配對可選，解析與 PCM 產生在第一次 `PlayTrack` 時延遲執行。

`cmd/dsds` 每幀依畫面情境選曲：戰鬥依序嘗試 `BATTLE1`、`BATTLE2`、`BT02`，
其他畫面嘗試 `STRATEGY`，再退回 `SCENE`。所有鍵盤、滑鼠與觸控都只改變畫面狀態，
音訊在同一個 UI 入口同步；不把音量、曲目或 player 寫進規則層／存檔。

初始曲目與第一次成功渲染的曲目會留在 `Manager.pcmByTrack`；後續切換重用同一份
PCM，不在每次畫面重繪時重新解析原始 `MUS/TIM`。快取只存在執行期，不寫入存檔或
發行資產。

## 3. 邊界與差異

- `PlayTrack` 先成功解析／產生新 PCM 才停止舊 player；壞曲目保留原播放。
- 曲目切換先建立新 player，再以 bounded 18 frame（約 300 ms）交叉淡入淡出並關閉
  舊 player；這是 remake 聽感選擇。小節對齊仍未實作，切換新曲從 cue 零點開始。
- 合成器仍是 `internal/audio/opl2.RenderAdLib` 的 deterministic 純 Go 近似，不是
  `SDFA.EXE` 的 register parity，也不宣稱已完成 DOSBox 取樣；register parity 不再是
  玩家可播放路徑的阻塞條件。
- `opl2.ProgramRegisterWrites` 現可把 TIM 的 operator bitfield 轉成公開 YM3812
  靜態寄存器序列，供未來硬體／串流 adapter 使用；這是 strong inference，不代表
  SDFA 的實際寫入順序、延遲、頻率或 `An` 映射。
- `-audio=off` 不建立 `audio.Context`；作者程式碼不使用 `cgo`。

## 4. 驗收

1. fake context/player 測試確認切曲只保留一個 active player、缺曲目 fail-closed、
   切曲失敗不丟失目前曲目；3 frame fake tick 會關閉 outgoing player，實際 UI 映射
   使用 18 frame bounded crossfade。
2. Docker/Xvfb 下 `go test -count=1 ./...` 與 `cmd/dsds` 啟動路徑通過；
   no-cgo、deny scan、diff check 與 Windows 交叉建置照專案交接契約執行。
3. 完成宣告只涵蓋情境串接與 bounded crossfade，不涵蓋現代 Ogg、register parity、Android
   音訊裝置或三平台發行。
