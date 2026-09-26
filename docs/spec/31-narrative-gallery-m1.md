# SPEC-31 敘事／新聞來源圖像畫廊與語系 overlay

> 狀態：**READY（2026-08-11 已實作）**
> 範圍：`NEWSDATA.DAT` 的唯讀畫廊、語系 catalog、Modern UI 入口。這是 remake 新增的
> 文化保存功能，不改規則、回合、亂數或任何存檔 bytes。

## 1. 目的與證據邊界

`NEWSDATA.DAT` 由 17 張 `215×16` 的 BGI 橫幅組成。它們是預繪點陣圖，不是可直接以
Big5／倚天字模反查的文字串；目前反查 17/17 未得到可重生字串。因此本規格採
「原圖優先、已證實文字才翻譯」：

- 17 張來源圖像全部保留並以 `source_image` 索引。
- 目前只把 #0「調動軍隊至」與 #9「爆發大規模示威遊行」登記為已證實模板。
- 其餘 15 筆的 `status` 固定為 `source-image`，畫面顯示來源圖與「本文未解出」
  fallback；不得用 OCR 猜句子，也不得把 fallback 當成原版文字。
- 英文、日文與繁中都使用同一個 17 筆索引，翻譯只改 caption，不改事件觸發或規則。

這個邊界符合近似 remake 決策：未知的原版字型是證據限制，不是玩家可玩路徑的阻塞。
若日後取得新證據，只能在同一 schema 補上 provenance／caption，並保留舊的
`source-image` 歷史狀態。

## 2. 語系檔契約

每個語系目錄有 `narrative.json`：

```json
{
  "schema_version": "1",
  "language": "en",
  "source": "NEWSDATA.DAT",
  "fallback": "Original news bitmap (wording not yet decoded)",
  "entries": [
    {
      "id": 0,
      "status": "translated",
      "source_text": "調動軍隊至",
      "original": "Move troops to",
      "plain": "Move forces to",
      "source_image": "NEWSDATA.DAT#0"
    }
  ]
}
```

載入器 `internal/i18n.LoadNarrative` 失敗即關閉：schema、語系、來源、筆數、連續
ID、`source_image` 與 translated 的 `original`／`plain` 任一不符就拒絕整包。未知
條目由 `NarrativeCatalog.Text` 回傳 `ok=false`，呼叫端只能畫 fallback。

人物自傳是另一條資料鏈：英／日 387 篇 machine-draft overlay 仍標示
`bio_status=machine-draft`，繁中新增 61 篇走 additive overlay；本規格不把人物譯稿
誤混進新聞 catalog，也不把機器翻譯宣稱為人審出版稿。

## 3. 玩家路徑

1. 政略地圖右上顯示「新聞／史事」按鈕；沒有合法的 17 張來源圖時，入口不顯示。
2. 鍵盤按 `N`、滑鼠點擊或 Android 單指點擊都派送 `actions.OpenNarrative`，共用
   `layout.NarrativeButton` 命中矩形。
3. 畫廊每頁最多四張圖。`Space`／`PageDown`／下一頁箭頭前進，`PageUp`／上一頁
   箭頭後退，`Esc`／`N` 返回原畫面。
4. 切換 `wording=original|plain` 只換已證實 caption；圖片、頁數、索引與返回位置
   不變。Modern／retro 只換外殼色彩與邊界。

畫廊不寫回存檔，也不要求 `DOSBox` 長流程。`NEWSDATA.DAT` 解碼後另做尺寸 gate：
每張必須是 `215×16`，數量必須是 17；違反時整個入口停用並在 stderr 留下診斷。

## 4. 驗收

- `internal/i18n`：三語系 schema、#0／#9 translated、#1 source-image、Eten glyph
  覆蓋測試。
- `internal/ui/render`：Modern 資訊卡／勢力色帶與畫廊 renderer 測試。
- `cmd/dsds`：Docker Xvfb 正常啟動、地圖入口、頁面切換與返回路徑。
- `tools/deny_scan.sh --all`：不新增原版資產；來源圖像只來自玩家自備資料，不能
  進版控或發行包。

## 5. 已知限制與下一步

- #1–#8、#10–#16 的原版字型／填字機制仍是 `unknown`，不以本規格宣稱解出。
- 英／日正式人審譯稿、日文假名字型與比例字型是獨立 release gate。
- `AC.TPC` 戰鬥插圖與 320×200 開場尚未接入；不與新聞畫廊混為一談。

