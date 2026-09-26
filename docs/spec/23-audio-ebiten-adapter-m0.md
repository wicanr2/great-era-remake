# SPEC-23：Ebiten 音訊適配器 M0（可關閉、開局曲目）

狀態：**READY**  
日期：2026-08-10

## 1. 目的

把 SPEC-22 的純 Go PCM 接到遊戲外殼，完成第一個真實玩家可觀察的音訊垂直切片：
`-audio=retro` 讀取玩家自備的原版 `SCENE.MUS`／`SCENE.TIM`，經純 Go
`opl2.RenderAdLib`（AdLib／OPL2 直接播放入口）產生 stereo PCM，交給 Ebiten
`audio.Player` 播放；`-audio=off` 完全不建立
`audio.Context`，在無音效裝置／CI／截圖環境仍可正常啟動。

本切片不修改規則層，也不把原版檔案放進儲存庫或發行包。開局只接 `SCENE`，不宣稱
政略／戰鬥／過場／結局的完整曲目切換已完成。

## 2. 輸入與邊界

- `audio=off` 與 `audio=retro` 是明示的 CLI 值；未知值 fail-closed。
- `retro` 只接受呼叫端從唯讀 `gameDir` 讀出的 `SCENE.MUS`／`SCENE.TIM` bytes；
  適配器不自行搜尋、修改或複製原版資產。
- 音訊初始化延後至 `retro` 路徑；`off` 路徑不得呼叫 `audio.NewContext`。
- Ebiten／Oto 是外部平台依賴；本專案新增程式碼不得包含 `cgo`、`import "C"` 或
  `#cgo`。音訊核心仍留在無頭可測試的 `internal/audio`。
- `SCENE` 的實際情境對應為既有格式文件的強證據／假說，不在本規格升格；此處只標為
  「開局預覽曲目」，未來取得 DOSBox oracle 後可替換曲目選擇。

## 3. API 與生命週期

- `internal/ui/audio` 提供 `Mode`、`ParseMode`、`NewRetro`、`Start`、`Stop`、`Close`。
- `Manager` 只持有目前一個 player；重複 `Start` 先關閉舊 player，避免 Ebiten 同一
  source 被多個 player 播放。
- 播放器音量由呼叫端傳入 0..1；適配器不把它寫入遊戲存檔或規則亂數。
- 音訊建立／解析／PCM 失敗時，`run` 回傳具體錯誤；`off` 則不受音效裝置影響。

## 4. 驗收

1. `internal/ui/audio` 的 fake context/player 測試覆蓋：off 不初始化、未知 mode 拒絕、
   retro 播放／停止／重播只保留一個 player、資產解析錯誤與音量邊界。
2. Docker Xvfb 下既有 `go test -count=1 ./...` 綠；Windows `CGO_ENABLED=0` 仍可建置。
3. `tools/check_no_cgo.sh`、`tools/deny_scan.sh --all`、`git diff --check` 通過。
4. 本切片的完成宣告只涵蓋開局曲目與 off gate；不得推論曲目切換、原版 OPL2 parity、
   Android 音訊裝置或正式三平台發行已完成。

## 5. 非本切片範圍

- F3／設定頁音訊軸、prefs 持久化與 modern Ogg 音源。
- 進戰鬥、戰報、劇情、結局的曲目切換與淡入淡出。
- `SDFA.EXE` 解包、DOSBox register trace、`An` 真實映射與逐樣本比對。
