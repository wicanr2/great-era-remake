# 高解析度 Modern UI 與 Android 版規劃

> **狀態：DRAFT（H1-a／H1-b／H1-c＋H2 流程頁已接；H3 僅保留技術基準；H4「全域指揮室」
> 的軍事衛星作戰室 M1 已接，其他 H4 流程頁與 Android 實機仍待）**
>
> 日期：2026-08-11
>
> 範圍：Modern 呈現層、可重用的高解析版面、Android 技術驗證與觸控尺寸。
> 本文件只做規劃與決策閘門；三平台桌面封裝由獨立發行流程處理，不藉由版面工作
> 修改遊戲規則、戰鬥或存檔格式。

## 0. 本輪結論先行

目前畫面看起來「太擠、字太大」，根因不是視窗不夠大，而是現有 Modern UI
仍以原版 640×350 當**唯一邏輯畫布**：`cmd/dsds` 的 `scale = 2` 只把視窗
放大到 1280×700，`Layout()` 仍回傳 640×350，16×15 點陣字與原版欄距也被
整數放大。這條路徑要保留給 `retro` 與目前可玩的 Modern shell，不能直接
稱為高解析度版。

高解析度版應是另一個只供 Modern 使用的設計畫布：以可重排的面板、較小的
相對字級、可再散布比例字型與密度感知觸控區重新排版；規則資料、動作識別碼、
人物自傳資料與 `.DT1`／`.DT2` 寫回完全共用。Android 以這個 Modern 畫布為
基準，不把 640×350 的原味版面硬塞到手機上。

本輪 Docker 截圖驗收已留下四張可版本化的成果：[Modern 高解析地圖／資訊卡](../images/high-modern-province-26.png)、
[Modern 高解析政略指令](../images/high-modern-menu-26.png)、[Modern 高解析戰鬥 HUD](../images/high-modern-battle-26.png)與
[Modern 高解析人物自傳](../images/high-modern-biography-058-01.png)。四張圖只作展示證據，
不代表原版逐像素 parity；現有 `docs/images/province-26.png`／`menu-26.png` 復古成果
保持不變。

### 2026-08-11 H3 UX 決策（已退回視覺驗收）：任務導向策略儀表板

使用者明確要求高解析 Modern 更符合現在策略遊戲。本輪採用「地圖為戰略上下文、
儀表板為本回合決策中心」：全寬頂列提供省份／日期，左側 744×592 地圖卡只放戰場，
右側 464×592 儀表板集中司令、省長、黃金／糧食／兵力／將領、後勤與兩個既有入口。
政略卡片改成序號＋圖示＋短標籤，戰鬥 HUD 則分離摘要、戰況與命令。排除的方案是
「只把原版地圖加側欄放大」。這是可見的 remake 差異；Action、資料、規則與存檔 bytes
完全共用，沒有自行發明任務、快捷鍵或新的戰鬥規則。

使用者已確認 **1280×720、16:9、橫向** 作為第一個 Modern／Android 設計
畫布。寬高決策已落定；H1-a 地圖／資訊卡、H1-b 指令／設定／自傳／敘事、H1-c
戰鬥 HUD，以及 H2 政略／查閱流程頁的 remake 玩家路徑已接通。Modern 的 16×16 GEMF
比例字型 atlas 與英／日 387 篇 coverage 已接入；逐篇人審、Android 實機／音訊
驗收、Modern Ogg provenance 與發行包仍未完成，故本文件維持 DRAFT，不能把畫面
接通誤宣稱為 READY 或原版 parity。

### 2026-08-11 H4 UX 決定：全域指揮室（採用 A）

使用者已從可丟棄的三方向版面原型中選定 **A「全域指揮室」**；「省份作戰板」與
「沉浸式地圖」不作為根層資訊架構。新版 Modern 以持續可見的導覽列、中央戰略地圖
工作區、依選取省份／戰區變化的檢視器，以及底部指令／事件列構成一條正常玩家路徑。
它只重新安排既有 `Action`、資料與畫面，不新增任務、快捷規則、成本、存檔欄位或假功能。

人物自傳在 Modern 版改為「人物檔案」：從將領或省份脈絡開啟，先呈現姓名、職務、
時期與可靠度，再閱讀可分頁的小傳。人物照片屬於額外的文化詮釋層，而非原版資料；
沒有逐檔確認身分與再利用權利前，畫面使用中性的檔案卡後備畫面（fallback），不能以 AI 臉孔或
未證實圖片補洞。H3 的幾何、正常 `Action` 接線、滑鼠／觸控與資料提供層（provider）保留為
技術證據，但其六卡視覺不再是 H4 的設計目標。

### 2026-08-12 H4 視覺語言：軍事衛星作戰室（已採用）

使用者已在可丟棄的 H4 原型後選定「**軍事衛星作戰室**」。因此「檔案式指揮室」與
「編輯檔案館」不再作為 H4 主畫面的視覺方向。Modern 主畫面將採深藍黑戰情底色、冷白文字、
青藍資訊／連線、琥珀焦點與橘紅主要指令；中央是受掃描格線保護的地圖窗，右側是省份軍情
檢視器，底部是既有指令與事件的操作列。它不是衛星照片、真實情報或新遊戲系統，而是純 Go
幾何與完整 Modern theme group 構成的 remake 呈現差異。

第一個可實作切片見 [`SPEC-41`](../spec/41-modern-satellite-command-center-h4-m1.md)：保留既有
`OpenCommands`／`OpenNarrative` action、六角格幾何、人物肖像權利 gate 與儲存邊界；H3 的
淺色六卡儀表板只留作歷史技術回歸，不能再充當 H4 截圖或推廣片畫面。

其後的 [`SPEC-42`](../spec/42-modern-command-matrix-h4-m1.md) 已把 15 項既有政略指令改為
5×3 衛星指令矩陣；它只換高解析的閱讀順序與卡片密度，`Selection(1..15)`、鍵盤、滑鼠、
觸控與規則處理均保持同一條路徑。

## 1. 已知事實與限制

| 項目 | 目前證據 | 對規劃的意義 |
|---|---|---|
| 原味主遊戲畫布 | `internal/ui/render` 的 `ModeBGIW/ModeBGIH = 640/350`，並有逐像素測試 | `retro` 不改；高解析 Modern 必須是另一條呈現路徑 |
| 現有桌面放大 | `cmd/dsds` `scale = 2`、視窗 1280×700，但 `Layout()` 仍回傳 640×350 | 這只是視窗縮放，不是高解析渲染或寬版排版 |
| Modern 現況 | 主題、地圖、部隊圖示、HUD、語系、人物自傳與敘事畫廊已在 640×350 可玩 | 可重用語意資料與 provider；幾何需抽離硬編碼 |
| 輸入現況 | `internal/ui/actions.SurfaceToLogical` 先處理等比縮放／黑邊；滑鼠與觸控共用 `Action` | 高解析版只換座標基準，不另造規則入口 |
| 觸控契約 | 現有程式圖形按鈕多為 48×48 邏輯像素；Android 48 dp、安全區、密度仍待實機驗證 | 不能把邏輯像素直接冒稱 Android dp；要在 adapter 做密度驗收 |
| 高 DPI 能力 | Ebiten 的 `Game.Layout` 可回傳高於外部尺寸的畫布，並可使用 `DeviceScaleFactor()` | 可做真正高解析 render，但需明確區分設計畫布與裝置像素 |
| Android 綁定 | 官方 `ebitenmobile bind -target android` 產生 `.aar`；文件列最低 Android SDK 16 | 先做工具鏈 smoke，再決定專案封裝與最低版本 |
| cgo 邊界 | 官方 Ebiten 平台表將 Android 標成需要 cgo；本專案 `cmd/`／`internal/` 仍禁止 `import "C"`／`#cgo` | Android feasibility gate 必須記錄「作者程式無 cgo」與「官方產物可能需要 cgo」的差異，不可偷偷宣稱純 Go Android 包已完成 |
| 原版字型 | 倚天字模由玩家執行期提供，不可進版控或安裝包 | 高解析 Modern 必須另找可再散布、涵蓋繁中／英／日的字型；未清權前只做 placeholder／測試，不進發行 |
| 音訊 | OPL2／AdLib 與 Modern Ogg runtime 已是純 Go；現代音樂內容與 Android 真機音訊仍是 release gate | 高解析規劃不重新打開 SDFA parity；音訊在 Android 階段另做裝置驗證 |

官方參考：

- [Ebitengine Mobile](https://ebitengine.org/en/documents/mobile.html)：`ebitenmobile bind`、`.aar` 與 Android SDK 需求。
- [Ebitengine FAQ：高 DPI](https://ebitengine.org/en/documents/faq.html)：`Layout` 與 `DeviceScaleFactor()` 的高 DPI 模式。
- [Ebitengine 平台列表](https://github.com/hajimehoshi/ebiten#platforms)：Android 的 cgo 條件。

## 2. 目標與非目標

### 2.1 目標

1. `retro` 仍能以 640×350 原味路徑執行；Modern 高解析版與它可在遊戲中切換，
   不重啟、不改遊戲狀態。
2. Modern 使用獨立設計畫布與響應式版面，降低文字擁擠；人物自傳、英／日語系、
   白話用語與新聞／史事入口仍走現有 catalog，不複製第二套內容。
3. 桌面滑鼠、鍵盤、Android 觸控都只產生同一個 `internal/ui/actions.Action`；
   點擊不直接修改規則或存檔。
4. 高解析度版的命中區由同一份版面配置產生，按鈕在 Android 上達到至少 48 dp，
   並處理黑邊、安全區、旋轉、背景恢復與系統返回鍵。
5. 不散布原版執行檔、資料、倚天字型、原版音樂或其旋律衍生物；Modern 資產與
   字型授權狀態可逐項追溯。

### 2.2 非目標

- 不把 640×350 `retro` 逐像素比對改成 16:9。
- 不把目前的 H0 等比橋接（或 `scale = 2`）冒充最終高解析 Modern；橋接只用來
  先驗證 1280×720 Surface、F3 切換與座標安全區。
- 不重寫戰鬥規則、AI、撤退／部署時機、`.DT1`／`.DT2` 未解欄位或存檔格式。
- 不在沒有字型、圖像、音樂授權與 Android 實機證據前建立正式發行包。
- 不因完成高解析 UI 而宣稱 Android APK／AAB、Android 真機或各桌面平台真機驗收完成；
  桌面包由獨立發行流程產生與記錄。

## 3. 根層決策：Modern 設計畫布

### 選項 A（建議）：1280×720、16:9、橫向

- 手機橫向與桌面預覽最常見，安全區與影片／商店素材容易共用。
- 比現有 640×350 多出約 2 倍像素面積，能把左側資訊卡、地圖、指令列與自傳
  內容分開，不必把每個字畫成兩倍大的點陣。
- 以 1280×720 作設計座標，不代表每台裝置都渲染成 1280×720 實體像素；實際
  `Layout` 仍依 Surface 與裝置比例等比縮放，超出的上下／左右區域進安全區。
- 16:10 平板與瀏海裝置透過 `SafeInsets` 保留內容，不拉伸文字；必要時讓地圖
  維持比例、面板向外延伸。

### 選項 B：1280×800、16:10、橫向

- 上下空間較充裕，適合平板與人物自傳長文；可少做一層垂直捲動／分頁。
- 16:9 手機需要上下安全區或裁切策略；商店截圖、桌面視窗與現有 1280×700
  預覽要多一套對照，第一版驗證成本較高。

### 已採用的根層決策（2026-08-11）

使用者選定 **選項 A**。選項 B（1280×800）不列入第一版基準，但未來平板仍可透過
`SafeInsets` 承接上下空間，不另造一套遊戲幾何。F3 是桌面快捷鍵；沒有實體鍵盤的
Android 入口是「其他選項 → 解析度切換」，兩者共用同一個偏好 setter。

## 4. 建議架構（確認畫布後才進入 READY）

### 4.1 畫布與渲染

新增與 Ebiten 無關的呈現資料（名稱可在 prototype 再定）：

```text
DesignSurface {
    Width, Height       // 1280×720 或使用者選定的基準
    SafeInsets           // 上／右／下／左，設計座標
    Density              // 僅供觸控與字級換算，不改規則
}
ModernLayout {
    map, status, command, message, biography, narrative, battle
}
```

- `internal/ui/render` 仍可在無頭環境合成 `image.RGBA`；先把 `ModeBGIW`、
  `ModeBGIH` 與 640 相關的幾何集中到 `DesignSurface`／layout metrics，避免每個
  `Draw*` 函式各自乘倍率。
- `internal/ui/theme` 只提供語意色彩、圖示與元件樣式；規則層不依賴 Modern。
- `retro` 仍建立固定 640×350 `Canvas`；目前 H0 以安全等比橋接把 frame 放入
  1280×720 Surface，並將觸控座標反算回 640×350。H1 才替換成真正的 Modern
  高解析 layout／字型；切換只替換 provider／layout，不重載遊戲資料，也不碰存檔 bytes。
- 現代地圖、部隊圖示、HUD 圖示優先使用程式圖形或已清權的高解析資產；原版
  `.TPC`／`.RGB` 不被放大後寫回 Modern 資產。

### 4.5 H1-a／H3 已接的策略儀表板切片

- `internal/ui/layout/modern.go` 固定 A 方案 1280×720、四側 24 px 安全區、64 px 頂列、
  左側 744 px 地圖卡與右側 464 px 策略儀表板；地圖、六張摘要卡（先顯示剩餘指令）、後勤與兩個既有
  入口均由 renderer 與 pointer 共用同一份矩形。地圖卡不再被入口按鈕覆蓋。
- H1-b 已追加 `ModernPageLayout`、`ModernCommandLayout` 與 `ModernOptionLayout`：
  政略 15 卡、顯示／解析度／其他設定卡、自傳與敘事頁都直接使用 1280×720 設計座標；
導覽與卡片命中區維持至少 48 px。Modern 預設使用 `GEMF` 比例字距 atlas，人物
自傳沿用既有 `bioPage` 並以可量測換行放大字與留白，Eten 只作玩家自備字庫的
繁中／復古 fallback，避免高解析頁與翻頁狀態分叉。
- `internal/ui/render/modern_high.go` 以 3/2 顯示倍率畫 14×14 六角格，保留奇數欄
  下移 18 px；地形／鐵路仍由既有 `Theme` 索引提供。
- `cmd/dsds` 在 Modern 地圖頁使用 H1-a renderer，在指令／設定／自傳／敘事頁使用
  H1-b renderer；戰鬥頁已接 H1-c renderer／命中區，原版未閉合的細節仍標在規則／
  oracle 邊界。
- `cmd/screenshot -theme modern -resolution high -battle` 現在以同一份 `BattleSim`
  部署快照產生 1280×720 H1-c 展示圖；它只整理 readonly panel／語系資料，不執行
  攻擊、AI 或存檔寫回，避免展示工具另造一條規則路徑。
- H2 已追加 `ModernFlowLayout`／`DrawModernFlowSurface`：其餘政略、查閱、政策、
  外交、數字輸入與確認頁不再把 640×350 frame 放大到高解析畫布，而是使用同一套
  高解析卡片／3×4 鍵盤／導覽命中區；`Selection`、`Digit`、`Confirm` 與存檔／規則
  接點保持不變。
- M0 已追加 `internal/ui/mobile/viewport.go`：以實際 Surface、density 與系統安全區
  反算 1280×720 設計座標，並在不改視覺矩形的前提下擴大 48 dp 觸控命中區；這層是
  純 Go，尚未等同 `ebitenmobile` binding 或 Android 實機驗收。
- M0 另追加 `internal/ui/mobile/lifecycle.go`／[`SPEC-39`](../spec/39-modern-mobile-lifecycle-m0.md)：
  純 Go gate 去重 pause／resume／destroy，背景切回會重置輸入並只對音訊產生一次
  pause／resume／close effect；它不直接派送 `Action`，也不冒稱 Android binding 完成。
- `cmd/screenshot -theme modern -resolution high` 目前可產生地圖／資訊卡與政略 3×5
  指令頁預覽；H1-b／H1-c 的文件／設定／戰鬥頁由正常玩家路徑、layout／renderer
  測試與 Docker＋Xvfb 回歸驗收，暫不把暫存 PNG 放進 repo。
- 加上 `-biography -province 26` 可由同一份 PeopleDB 產生人物自傳首頁／末頁；展示圖只放在
  `docs/images/` 白名單，工具不另造人物資料或規則路徑。
- H1-a 不嵌入原版字型；Modern 文字使用可再散布 `GEMF` atlas，缺少或損壞時才
  使用玩家自備字庫並保留缺字診斷；幾何／地圖預覽不會因原版字型缺席而失敗。

### 4.2 字型與排版

- `typography=bitmap` 的原味軸維持現況；Modern 高解析預設使用 `GEMF` 1-bit
  proportional provider，行高、字距與 CJK／日文 fallback 由 layout metrics 決定；
  這是純 Go raster atlas，不冒稱向量字型或原版字型。
- 倚天字型永遠只從玩家自備原版目錄執行期讀取，不進 APK、AAR 或 GitHub。
  Modern atlas 已有可再散布來源／hash／重生紀錄與英／日 overlay coverage；正式
  發行仍須將 OFL notice、商標聲明與各平台 license 檢查列入 release checklist。
- 長文（人物自傳、新聞 caption）採可量測換行與分頁；不再沿用 20 px 全形槽位
  硬塞現代白話。原典文句、現代白話與人物來源 metadata 仍由語系資料供應。
- `typography=vector` 與 `layout=original` 仍視為非法組合；載入時要明確診斷，
  不做靜默升級。

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

## 5. 分期與交付閘門

| 階段 | 內容 | 交付／停止條件 |
|---|---|---|
| H0 決策與可丟棄 prototype | 1280×720 畫布決策已確認；目前已接 F3／偏好／顯示設定、Surface 安全區與 640×350 frame 橋接 | 先驗證切換、黑邊與輸入不漂移；橋接不等於 H1 完成 |
| H1 Modern 高解析 renderer | **H1-a 地圖／資訊卡、H1-b 指令／設定／自傳／敘事、H1-c 戰鬥 HUD 已接；GEMF 比例 atlas 與英／日 coverage 已接**；人審與 Android 實機仍待 | retro golden、Modern 語意測試與存檔 round-trip 不變 |
| H2 共用指標與觸控 | **Desktop design-surface／流程卡／3×4 鍵盤已接**；Android 密度／旋轉仍待實機 | 同一 action 的鍵盤／滑鼠／觸控結果與 bytes 相同 |
| H3 Android 技術驗證 | `ebitenmobile bind` 最小 `.aar`、橫向 Activity、背景恢復、系統返回、`audio=off` | M0 viewport 已完成；cgo／NDK 邊界、最低版本與實機 Surface 量測仍待，失敗則停在技術報告 |
| H4 語系與敘事 | **GEMF／OFL provenance 與 387 篇 machine-draft coverage 已接**；逐篇人審、日文禁則／長文截圖仍待 | `bio_status`、source fallback、日文假名與長文截圖 gate |
| H5 音訊與裝置 QA | Modern Ogg／OPL2 在 Android 真機抽樣；背景、耳機、低端裝置與無音訊降級 | 技術解碼與人耳／授權分開記錄；未完成不進發行包 |
| H6 發行 | 三平台桌面包、Android APK／AAB、商店素材與推廣片 clearance | 桌面包可獨立於 Android 建置；Android APK／AAB 與公開發行仍待其專屬 gate |

每一階段都必須保留三種狀態：remake 已驗證、原版 oracle 未知、可選 polish。
高解析畫面好看不代表原版 parity；也不能因 Android shell 可啟動就宣稱觸控與音訊
實機驗收完成。

## 6. 驗收清單（規劃版）

- `retro` 640×350 golden 與現有原版資產拒絕掃描不變。
- H1-a 已有 Modern 1280×720 地圖／資訊卡可比較預覽；H1-b／H1-c／H2 已接政略／設定／
  自傳／敘事／戰鬥／一般流程正常路徑；完整 gate 仍需長文截圖、逐篇人審與 Android
  實機量測，不能把 renderer 通過寫成原版 parity。
- 640×350 與高解析 Modern 切換前後，遊戲狀態、`Action`、`.DT1`／`.DT2` writer
  輸出與人物翻頁結果一致。
- 觸控命中測試覆蓋按鈕中央／邊界／相鄰空隙／黑邊／安全區；Android 實機再量測
  48 dp，而不是只跑桌面單元測試。
- Android 背景切回不重複派送 action；系統返回等同 `common.back`；F10 離開語意不變。
- Modern 字型、圖示、音訊與影片的作者／授權／雜湊 metadata 可追溯；原版資料與
  倚天字型不進 APK、AAB、AAR 或公開 repo。
- 所有建置／測試／抓圖仍在 Docker；每批工作後清理本輪容器，且 `git diff --check`
  與 `tools/check_no_cgo.sh` 維持通過。

## 7. 待辦與決策紀錄

| 項目 | 狀態 | 說明 |
|---|---|---|
| Modern 設計畫布 | **已確認 A** | 1280×720、16:9、橫向；B 暫不作第一版基準 |
| Android 預設方向 | 目前規劃橫向 | 依畫布 prototype 與實機安全區再鎖定；直向不列入第一版 |
| Modern 字型 | **GEMF atlas 已接；OFL 發行 notice 待補** | 不使用倚天；來源／hash／重生命令見 `docs/licenses/modern-font-atlas.md`，英／日 overlay coverage 已由測試鎖定 |
| Android cgo／NDK 邊界 | 技術閘門 | 作者程式維持 no-cgo；官方 binding 的 cgo 需求需實測與明示 |
| Android 最低版本／封裝 | 待技術驗證 | 官方文件的 SDK 16 只作參考，不等同本專案支援承諾 |
| 三平台桌面包 | 發行流程中 | 解析度切換程式已接；封裝／各平台驗收不與 Android gate 混為一談 |
