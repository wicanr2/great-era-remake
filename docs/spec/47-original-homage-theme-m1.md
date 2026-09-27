# SPEC-47：原版致敬高解析 Modern 主題 M1–M2

> **狀態：READY（2026-09-27）**
> 前置：`SPEC-45`（M1–M3 管線）、theme 合約（`internal/ui/theme/theme.go`）、
> `RetroStyle`（米黃紙／暗紅墨／寶藍焦點）、原版裝飾遮罩契約
>（`DecodeOriginalOrnaments`：MENU1..5／WARMENU）、GEMF 比例字型。
> 使用者決策（2026-09-27，confirmed-user-decision）：否決戰棋主題（含配色），
> 新主題整套重做——忠於 1992 原版 16 色 BGI 美學的高解析重繪。
> `retro` 原味模式完全不變；戰棋產物（SPEC-45、wargame- 成果圖）保留為歷史。

## 1. 設計語言：高解析原版味

不是仿古濾鏡，而是把原版視覺語彙用高解析重畫一遍：

- **紙墨**：`RetroStyle` 的米黃紙（`#FFFFA2` 系）、暗紅墨（`#AE0000` 系）、
  寶藍焦點（`#0000AA` 系）即為新 Modern 色組基準；A2 土黃與戰棋銅章退場。
- **邊框**：程式化重繪 MENU1..5／WARMENU 的花框語彙（雙線＋轉角紋＋帶狀飾），
  clean-room 繪製，不直接拿原版 TPC 位元組當正式視覺；原版遮罩保留為
  差異比對參考（`drawModernOrnamentMask` 管線沿用）。
- **部隊圖**：回到 NEWICON 剪影語彙（人形／戰車／騎兵／火炮剪影＋紅藍攻守），
  32×17／透明／索引順序不變，僅以高解析重繪邊緣（目的尺寸 4 倍超採）。
- **地圖**：16 色 BGI 地形對照（平原米黃、森林綠、水域藍、山脈棕）的高解析版；
  測繪紙感退場。
- **標題字**：宋體（C 分工不變）；印章、勳章、回紋角等戰棋裝飾全數移除。

## 2. 合約邊界（只換呈現，不碰規則）

- 仍是 `retro`／`modern` 雙 mode，`UIStyle`／`Theme`／`HighResolutionTheme`
  介面不變；整套重做 = 再次替換 Modern provider 的實作內容。
- 不改：15 指令語意、`Selection`、命令成本、地圖資料、`TileKind`、鐵路接線、
  戰鬥規則、存檔、pointer／touch 矩形、語系文字、AI、自傳 gate。
- 排版沿用既有度量（GEMF advance 即繪製 advance，`ModernBiographyPageCount`
  唯一入口）；新邊框只吃 text-safe 剩餘空間。
- 高解析圖一律目的尺寸 4 倍 RGBA 繪製後降採樣；禁 emoji、外部
  webfont、執行期 SVG parser、`cgo`、原版素材衍生。

## 3. 切片與驗收

- **M1 色組＋邊框**：Modern `Style()`／調色盤轉原版三色基準；花框 primitive
  取代回紋角（`modern_ornaments.go`）。驗收：theme／render 測試＋政略頁截圖。
- **M2 部隊＋地形＋頁面**：18 剪影重繪、22 地形 BGI 對照、各頁套框。
  驗收：戰鬥／自傳截圖＋`internal/ui/...` 全綠＋deny／no-cgo。
- 每一切片：Docker 測試 → Xvfb 1280×720 截圖目檢 → 閘門 → 登記。
  wargame- 成果圖保留為歷史，不得冒充新主題驗收；M2 結束後重包
  `0.1.3-homage` 並更新 `docs/images/homage-*.png`。

## 4. 明確排除

- 新玩法、命令更名、規則／存檔／命中區改動。
- 正式 Ogg、Android 實機——仍是獨立 gate。
- 原版 oracle／parity 不因換皮重開（SPEC-30 維持）。
