# SPEC-45：民國戰棋 Modern 主題整套重做 M1–M3

> **狀態：READY（2026-09-27）**
> 前置：`SPEC-43`（A2 色組、C 字體分工）、`SPEC-44`（測繪地形、專屬指令圖示、
> 高解析管線）、theme 合約（`internal/ui/theme/theme.go`）、GEMF 比例字型。
> 使用者決策（2026-09-27，confirmed-user-decision）：否決 Codex 新版 UI，
> 新主題整套重做——1920s 民國風為底、戰棋桌遊感邊框／圖像／部隊、支援高解析。
> `retro` 原味模式完全不變。

## 1. 設計語言：民國戰棋

底為 1920 年代軍政文書質感，表為桌遊算子（counter）語彙：

- **紙墨**：暖黃公文紙面、深褐墨線、朱紅印章點綴、黛藍／藏青軍政色。沿用 A2
  明亮土黃紙面方向，收斂衛星深藍黑殘留。
- **邊框**：雙線回紋＋四角雲雷紋的程式化描邊 primitive（純 Go，目的尺寸繪製）；
  重要面板加銅章色內線。原版 `MENU1..5`／`WARMENU` 執行期遮罩退為次選裝飾，
  缺檔時程式化 fallback 即為正式視覺，不再以原版花框為驗收基準。
- **部隊圖**：算子風格——圓角方框＋兵種符號（步兵交叉步槍／騎兵馬頭／炮兵炮身／
  裝甲履帶菱形）＋攻守色＋右下角戰力數字；六朝向炮兵保留方向楔形。灰階下靠輪廓區分。
- **地圖**：保留測繪語彙（紙面、水系、山脈、鐵路），色溫向民國印刷品收斂，
  與新邊框同一次定稿，避免兩次換皮。
- **標題字**：宋體（既有 C 分工不變）；印章式小方塊作分類標籤，不取代語系文字。

## 2. 合約邊界（只換呈現，不碰規則）

- 不新增主題 mode：仍是 `retro`／`modern`，`UIStyle`／`Theme`／`HighResolutionTheme`
  介面不變；整套重做 = 替換 Modern provider 的實作內容。
- 不改：15 指令語意、`Selection`、命令成本、地圖資料、`TileKind`、鐵路接線、
  戰鬥規則、存檔、pointer／touch 矩形、語系文字、AI。
- 排版沿用既有度量：`modernBiographyAdvance`／`modernTextPixelWidth` 的 GEMF
  advance 即繪製 advance（`modern_pages.go:236-291,646-789`），頁數唯一入口
  `ModernBiographyPageCount`；新邊框只吃 text-safe 矩形剩餘空間，不得壓縮正文欄。
- 高解析圖一律目的尺寸 4 倍 RGBA 繪製後降採樣（沿 SPEC-44 §2）；禁 emoji、外部
  webfont、執行期 SVG parser、`cgo`、原版素材衍生。

## 3. 切片與檔案順序

- **M1 色組＋地形＋部隊**：`internal/ui/theme/modern.go`（色組）、地形 22／鐵路 21
  高解析繪製、18 部隊算子（含炮兵六朝向）。驗收：`theme` headless 尺寸／索引／透明
  測試＋39 省地圖截圖抽查。
- **M2 邊框＋頁面**：`modern_ornaments.go` 回紋邊框 primitive、按鈕／面板／
  `modern_pages.go` 各頁套用、`modern.go` 指令圖示收斂為算子語系。驗收：政略／
  指令／敘事／設定頁截圖＋命中區回歸。
- **M3 戰鬥＋人物＋推廣**：`modern_battle.go` HUD、`modern_high.go` 高解析外殼、
  人物檔案肖像欄框線、`cmd/screenshot` 成果圖更新、三平台重包。驗收：戰鬥頁、
  自傳頁截圖＋`dist-all/` 候選包。

每一切片：Docker 測試 → Xvfb 正常玩家截圖（1280×720，固定資料）→ deny／no-cgo／
diff check → 更新本規格驗收欄。舊 H4／A2 成果圖保留為歷史，不得冒充新主題驗收。

## 4. 明確排除

- 新玩法、命令更名、規則／存檔／命中區改動（見 SPEC-44 §4 同款禁令）。
- 正式 Ogg、人審譯稿、Android 實機——仍是各自獨立 gate，不在本規格內關閉。
- 原版 oracle／parity 不因換皮重開（SPEC-30 維持）。
