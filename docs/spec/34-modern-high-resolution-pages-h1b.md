# SPEC-34：Modern 高解析頁面 H1-b（指令／設定／自傳／敘事）

> 狀態：**READY（H1-b 已接玩家路徑）**　日期：2026-08-11

## 1. 範圍

H1-a 已完成地圖／資訊卡。本切片把同一個 1280×720 設計座標接到：

- 15 項政略指令卡片；
- 顯示設定、解析度與其他選項卡片；
- 人物自傳全文頁與分頁；
- NEWSDATA 敘事圖庫與分頁。

規則、Action ID、人物資料、語系來源與存檔 bytes 不變。戰鬥頁仍是下一個 H1-c，
不以本規格宣稱高解析戰鬥完成。

2026-08-11 的 H3 視覺收斂採現代策略遊戲的資訊層級：深藍色小型頂列保留頁面脈絡，
紅色細條只標示焦點；卡片以序號、圖示、短標籤分工，不再把原版點陣字等比例放大成
巨型標題。這是 Modern 外殼的排版差異，所有選擇仍回到既有 Action。

## 2. 幾何與輸入契約

`internal/ui/layout/modern.go` 的 `ModernPageLayout` 固定 24 px 安全區、頁首／內容／
頁尾，以及至少 48×48 的導覽目標；`ModernCommandLayout` 為 3×5 卡片，
`ModernOptionLayout` 為雙欄卡片。renderer 與 `cmd/dsds/pointer.go` 直接消費同一份
`Rect`，不再以 640×350 反算高解析頁。

命中仍轉成既有 `actions.Selection`、`Back`、`PreviousPage`、`NextPage`；拖曳門檻、
觸控按下／抬起與鍵盤語意維持既有契約。

每一個流程頁的標題必須引用它實際由哪張政略卡進入的語系詞條；流程 state 的歷史
分類順序不能當成選單序號。`cmd/dsds/modern_flow.go` 的 `strategyCommand*` 常數與
回歸測試鎖定此對應，避免「查閱」頁錯掛成「秘密行動」等玩家可見語意錯置。

## 3. 文字與資產邊界

頁面採隨程式散布的 `GEMF` 1-bit proportional atlas（來源／授權／重生見
`docs/licenses/modern-font-atlas.md`）；缺少／損壞時才退回玩家自備倚天字庫，且
renderer 保留缺字診斷。不把原版字型嵌入 repo 或 Android 包。人物自傳沿用
`textlayout.LayoutProportional` 與既有 `bioPage`，以 GEMF 像素字距計算換行、固定行數
分頁與留白，避免 page state 漂移；沒有 Eten 字庫時仍可由 embedded atlas 顯示；敘事圖像仍由通過格式驗證的
`NEWSDATA.DAT` 提供，未知翻譯使用 catalog fallback。

## 4. 驗收

- `internal/ui/layout`：頁首／內容／頁尾與 15 張指令卡不重疊，選項卡達 48 px；
- `internal/ui/render`：高解析 command／options geometry 與 fail-closed canvas；
- `cmd/dsds`（Docker＋Xvfb）：高解析指令／設定 pointer targets、人物／敘事分頁與
  原版路徑測試同時通過；
- 指令頁的 15 張卡、人物自傳的正文／可靠度／翻頁控制各有獨立留白與層級；不以
  靜態展示頁取代正常玩家路徑；
- `tools/check_no_cgo.sh`、`tools/deny_scan.sh --all` 與 `git diff --check` 通過；
- 英／日 387 篇 machine-draft overlay 的 glyph coverage 已由 `internal/assets/modernfont_test.go`
  鎖定；`internal/ui/render/modern_biography_overlay_test.go` 另抽測兩語系最長自傳的
  首／末頁與無 Eten fallback；逐篇人審、禁則／長文截圖、Android 密度／48 dp 實機、
  戰鬥 H1-c 與三平台包仍是後續 gate。
