# Android 呈現、輸入與平台邊界

狀態：DRAFT（平台接線與實機驗收）。本檔只保存 Android 契約；桌面呈現以 [SPEC-48](../spec/48-hd-release-integration.md)為準，現況只在 [CONTEXT.md](../../CONTEXT.md)，待辦只在 [Issue #6](https://github.com/wicanr2/great-era-remake/issues/6)。

舊桌面方案及過去的進度表已移出目前文件，完整歷史見 [Git 版本](https://github.com/wicanr2/great-era-remake/blob/fae384579b8f63e974f5c0c6f6b701f800ce6b36/docs/design/33-high-resolution-modern-android.md)。以下保留原有 Android 畫布、輸入、密度、生命週期、cgo 與驗收要求，未更動平台支持承諾。

## 畫布

Android 沿用 1280×720 邏輯畫布。

### 已採用的根層決策（2026-08-11）

使用者選定 **選項 A**。選項 B（1280×800）不列入第一版基準，但未來平板仍可透過
`SafeInsets` 承接上下空間，不另造一套遊戲幾何。F3 是桌面快捷鍵；沒有實體鍵盤的
Android 入口是「其他選項 → 解析度切換」，兩者共用同一個偏好 setter。


## 輸入與密度

### 4.3 輸入與密度

1. 實體 Surface 座標先扣除 letterbox／安全區，再以等比矩陣反算到
   `DesignSurface`；黑邊點擊忽略。
2. 命中區由 `ModernLayout` 產生，與繪製元件使用同一個矩形；不能另維護一份
   640×350 座標表。
3. Android 以密度獨立像素驗收 48 dp；若 Ebiten 平台回傳的是裝置像素，adapter
   必須以實測 `density` 換算後再決定命中區，不能把目前 48 邏輯像素直接宣稱為
   48 dp。
4. 單指按下／抬起仍只派送一次 action；拖曳超過門檻取消；多點觸控與系統返回鍵
   都 fail-closed，不得重複執行存檔或戰鬥指令。
5. 地圖與戰鬥六角格保留既有幾何與 action ID；高解析只改顯示尺寸與安全命中區。


## 平台接線

### 4.4 Ebiten／Android 邊界

- 桌面與 Android 共用 `Game`、規則層、`Action` 與 Modern layout；平台層只負責
  Surface、生命週期、輸入與音訊裝置。
- 官方 Android 路徑是 `ebitenmobile bind` 產生 `.aar`，不是把桌面 `main` 直接
  交叉編譯成 APK。第一個 Android milestone 應先建立最小 `mobile` package，呼叫
  `mobile.SetGame`，不呼叫桌面 `RunGame`。
- 本專案作者程式碼仍維持 no-cgo（`tools/check_no_cgo.sh`）；但官方文件把
  Android 列為 cgo-required，因此要在工具鏈 smoke 階段記錄生成 binding 的 cgo／
  NDK 需求。若「任何產物都不得有 cgo」是不可退讓的硬條件，Android 正式封裝在
  此邊界前停住，不能用桌面 Linux／Xvfb 通過冒稱完成。
- 最低 Android 版本、NDK、Gradle／Android Studio 專案與旋轉策略都列為後續
  子決策；不在尚未確定畫布前一起鎖死。

#### 2026-08-11 工具鏈 smoke 勘誤

目前可沿用的 `rich2-go-android:20260809` 實際包含 `/go/bin/ebitenmobile`、Android
platform `android-35`、NDK `27.2.12479018`、`javac` 與 `adb`；`gomobile` 與
`ndk-build` 不在 image。先前交接中「image 缺 `ebitenmobile`／SDK／NDK」的說法已由
這次容器盤點訂正，不能再當成目前事實。

隔離執行 `ebitenmobile bind -target android -androidapi 35` 時，工具本身可啟動並取得
所需 `gomobile` 模組；但直接指定目前的 `./cmd/dsds` 會 fail-closed：
`binding "main" package .../cmd/dsds is not supported`。這不是遊戲規則或高解析 renderer
失敗，而是入口形狀不符合 mobile binding。下一個 H3 子任務必須抽出不屬於 `main` 的
`mobileapp` package，提供 `ebitenmobile` 可接受的 `Game`／初始化邊界，再另做 AAR
smoke；本輪沒有新增假 adapter，也沒有把容器暫存 AAR 當成發行成果。

作者程式仍通過 `tools/check_no_cgo.sh`；上述官方 binding 的 cgo／NDK 是平台產物邊界，
不等於在 `cmd/`／`internal/` 寫入 `import "C"` 或 `#cgo`。若未來要求所有最終產物也
不得經過 cgo，H3 應在這個邊界明確停住並回報，而不是繞過規則宣稱 Android 完成。


## 生命週期

純 Go viewport 與生命週期契約分別見 [SPEC-38](../spec/38-modern-mobile-viewport-m0.md)及 [SPEC-39](../spec/39-modern-mobile-lifecycle-m0.md)。平台層仍須實際接上 pause、resume、destroy、安全區、音訊與輸入重置，不以桌面容器驗證替代。

## 驗收要求

- 觸控命中測試覆蓋按鈕中央／邊界／相鄰空隙／黑邊／安全區；Android 實機再量測
  48 dp，而不是只跑桌面單元測試。
- Android 背景切回不重複派送 action；系統返回等同 `common.back`；F10 離開語意不變。
- Modern 字型、圖示、音訊與影片的作者／授權／雜湊 metadata 可追溯；原版資料與
  倚天字型不進 APK、AAB、AAR 或公開 repo。
- 所有建置／測試／抓圖仍在 Docker；每批工作後清理本輪容器，且 `git diff --check`
  與 `tools/check_no_cgo.sh` 維持通過。


## 技術參考

[官方 mobile 文件](https://ebitengine.org/en/documents/mobile.html)、[高 DPI 文件](https://ebitengine.org/en/documents/faq.html)及 [平台列表](https://github.com/hajimehoshi/ebiten#platforms)是平台語意來源。歷史工具鏈查驗只是當時記錄，最低 Android 版本與封裝方式仍待技術驗證。
