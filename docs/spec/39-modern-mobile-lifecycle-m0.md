# SPEC-39：Modern 行動生命週期閘門（M0）

狀態：**READY／純 Go 契約已實作，Android binding 未宣稱完成**　日期：2026-08-11

## 目的

Android 的 `onPause`／`onResume`／`onDestroy` 必須先在平台無關層收束，避免
背景切回時把半按的滑鼠／觸控狀態重放成第二個 `Action`，或重複 pause／close
同一個音訊 player。原生 adapter 未來只負責轉送事件與執行 effect，不得另造
一份輸入或規則狀態機。

## 契約

- `internal/ui/mobile.LifecycleGate` 初始為 `PhaseActive`；只有 active 接受新輸入。
- `EventPause`：active → background，一次回傳 `PauseAudio`＋`ResetInput`；重複 pause
  是 no-op。
- `EventResume`：background → active，一次回傳 `ResumeAudio`＋`ResetInput`；重複 resume
  是 no-op。恢復後不重放背景期間觸控。
- `EventDestroy`：active／background → destroyed，一次回傳 `PauseAudio`、`CloseAudio`
  與 `ResetInput`；destroyed 是終態，後續事件全部 no-op。
- gate 不直接產生 `Action`、不保存遊戲規則或存檔；adapter 清除 pointer／touch 按下
  狀態後，仍交回既有正常玩家路徑。

## 完成證據與邊界

- `internal/ui/mobile/lifecycle.go`／`lifecycle_test.go`：重複事件去重、輸入接受條件、
  音訊 pause／resume／close 一次性與終態測試。
- 這只證明純 Go 的生命週期契約；`rich2-go-android:20260809` 已確認含
  `ebitenmobile`、Android SDK 35、NDK `27.2.12479018` 與 `adb`，但目前 `./cmd/dsds`
  是 `main` package，直接 bind 會被工具拒絕。因此仍未有 `.aar`／APK、真機旋轉、安全區
  實測、背景音訊或商店包；不能由本規格或桌面 Xvfb 綠燈冒稱完成。

## 2026-08-11 H3 工具鏈勘誤

隔離執行 `ebitenmobile bind -target android -androidapi 35` 的工具本身可啟動並取得
缺少的 `gomobile` 模組；對目前 `./cmd/dsds` 的結果是
`binding "main" package .../cmd/dsds is not supported`，不是 lifecycle 或 renderer
失敗。H3 的下一個實作邊界是抽出可綁定的非 `main` `mobileapp` package；作者程式仍維持
no-cgo，官方 binding 的 cgo／NDK 僅記在平台產物邊界。
