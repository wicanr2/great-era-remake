# 規格 14：modern HUD 資源／指令圖示 P3

狀態：**READY（P3 窄切片）**  
日期：2026-08-10

本規格把 modern 主題目前缺少的 HUD 輔助圖示接入政略面板與十五項指令選單。
圖示是理解操作的視覺輔助，不取代原典／現代白話文字，也不改變規則、索引、輸入或
存檔格式。資產全部由純 Go deterministic provider 產生，不新增 PNG／SVG，也不把
原版像素或字型打進版控。

## 1. 資產契約

`internal/ui/theme.HUDIconProvider` 提供兩組固定索引：

| 組別 | 索引 | 數量 | 語意來源 |
|---|---:|---:|---|
| 資源 | 0..5 | 6 | 黃金、糧食、彈藥、燃料、煤礦、鐵礦；沿用 `Province` 欄位與原版面板順序 |
| 指令 | 0..14 | 15 | 政略指令 1..15；沿用 `render.StrategyCommands` 的順序 |

每張圖是 16×16 索引圖：像素索引 0 透明，調色盤至少 16 色，圖像與調色盤必須同一
provider 回傳。圖示不得承載未證實的兵種、勢力或規則語意；圖形只作操作分類提示。
索引越界、空圖、尺寸錯誤或調色盤不足都必須 fail-closed。

## 2. 主題切換契約

- `theme.Modern` 必須一次生成完整的地形、鐵路、部隊、資源與指令圖示組；切換只交換
  已驗證的 provider 指標。
- `setThemePreference` 切至 modern 時，若缺 `HUDIconProvider` 必須拒絕並保留原主題／
  偏好；切回 retro 時可使用文字 fallback，不改原版像素路徑。
- `PanelData` 與指令 renderer 共用同一個 provider；不可在 `cmd/dsds` 另建第二份
  索引表或繪圖規則。

## 3. 畫面接線

- modern 政略左側面板在六項資源列顯示 16×16 圖示，原有語系標籤與數值仍保留。
- modern 政略指令頁在每一列顯示對應指令圖示；原典／白話文字仍是主要可讀內容。
- retro 主題維持既有原版字模排版，不畫 modern 圖示；這是明示的呈現差異，不是規則
  差異。
- 圖示不參與滑鼠／觸控命中；命中區仍由既有 `internal/ui/layout`／`actions` 契約提供。

## 4. 驗收

1. provider 測試涵蓋 6＋15 全索引、尺寸／透明／調色盤、越界與 deterministic 重建，
   並確認至少相鄰語意圖示有像素差異。
2. renderer 測試確認 modern panel／command page 會畫出圖示，retro fallback 不因
   provider 缺失而報錯；文字仍會畫出。
3. app theme 測試確認 modern 缺 provider 時 fail-closed，成功切換時圖示 provider
   與 battlefield／unit provider 一起交換。
4. Docker 內執行相關 Go／Xvfb 測試、`tools/check_no_cgo.sh`、`tools/deny_scan.sh --all`
   與 `git diff --check`；不得留下本輪容器。

## 5. 非目標

- 十勢力專屬色／徽號、向量字型、寬版 HUD、Android 48dp 重新排版與正式美術資產。
- 把圖示當作原版考證或改寫任何外交、戰鬥、經濟規則。
