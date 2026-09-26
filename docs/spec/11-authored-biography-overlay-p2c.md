# SPEC-11：人物自撰傳記執行期增量 overlay（P2c）

狀態：**READY**（2026-08-10）

本規格把使用者要求的「人物自傳串上去」拆成一個可回復的增量切片：把研究稿中
已完成、且目前執行期 `people.json` 尚無正文的 61 篇自撰傳記接入 `PeopleDB`。
它不覆寫研究資料，也不改寫目前已有的 326 篇產品正文。

## 1. 邊界與資料權威

- 上游研究資料仍是 `docs/reference/people/` 的 facts／bios 批次。
- `docs/reference/people/people.json` 與 `translations/zh-Hant/people.json` 都是
  既有基準；本切片不得覆寫或自動潤稿。
- 新增產品輸入 `translations/zh-Hant/people-authored.json`，由
  `tools/gen_authored_bios.py --overlay` 可重生。這是**只填空白欄位**的 overlay，
  不是第二份完整人物資料庫。
- `DESIGN-22` 的「將 387 篇全部合併回 `people.json`」仍是 DRAFT；本規格只授權
  61 篇 additive overlay，不解除該文件的完整寫回禁令。

## 2. overlay schema

```json
{
  "schema_version": "1",
  "language": "zh-Hant",
  "overlay_scope": "new-biographies-only",
  "generated_from": "versioned research batches",
  "people": [
    {
      "id": 13,
      "name_ingame": "…",
      "bio_zh": "…",
      "confidence": "low",
      "provenance": {
        "facts": "docs/reference/people/facts-…json",
        "facts_sha256": "64 個十六進位字元",
        "bios": "docs/reference/people/bios-…md",
        "bios_sha256": "64 個十六進位字元",
        "bio_sha256": "64 個十六進位字元"
      }
    }
  ]
}
```

固定契約：`schema_version=1`、`language` 必須等於 locale、`overlay_scope` 必須是
`new-biographies-only`，且 `people` 恰有 61 筆、ID 不重複、排序為遞增。#274「無省長」
與 30 位 `unknown` 不得出現在 overlay；每筆正文非空、`confidence` 只能是
`high`／`medium`／`low`，五個 provenance 雜湊欄位都必須存在且是 64 位十六進位字串。

## 3. 載入與 fail-closed

`i18n.LoadPeople` 先載入既有 locale `people.json`，再尋找同目錄的
`people-authored.json`：

1. 檔案不存在是相容的「尚未安裝 overlay」狀態，維持既有 326 篇功能。
2. 檔案存在但 JSON、schema、語系、ID、姓名、正文、信心度或 provenance 任一項不符，
   整份 `LoadPeople` 失敗；不可靜默跳過 overlay 或只套用前半批。
3. 每個 overlay ID 在 base `people.json` 必須存在，且 `name_ingame` 完全相同。
4. base `bio_zh` 必須是空字串；若 overlay 嘗試覆蓋既有正文，直接失敗。成功後只填
   `Biography` 與 `Confidence`，其他人物欄位保持 base 原值。
5. overlay 套用後，執行期預期有 387 篇非空正文；30 位 unknown 仍維持空白並由既有
   「查無可靠傳記記載」fallback 呈現。

非繁中 locale 沒有此檔案時不報錯；未來英文／日文要有自傳資料，另立語系規格，不能
把繁中正文當成翻譯完成的假象。

## 4. 產生器與可重現性

`tools/gen_authored_bios.py --check` 維持原本 417／387／30／61 的研究品質閘。
新增 `--overlay --output translations/zh-Hant/people-authored.json`：只輸出
`newly_available` 的 61 筆，保留每筆 provenance；同一批輸入連跑兩次，輸出必須位元組
相同。產物不得指向 base `people.json`，不得由人工直接編輯。

## 5. 驗收

- 產生器 `--check` 與 `--overlay` 全過，overlay ID 數為 61。
- Go `internal/i18n` 測試證明：套用後 387 篇非空、#274／unknown 仍空、既有正文
  未被替換；缺檔相容、壞檔失敗即關閉。
- 人物自傳既有 B 鍵、翻頁、滑鼠／觸控入口不改；新增正文至少抽驗短／中／長篇與
  `high`／`medium`／`low`。
- `tools/deny_scan.sh --all`、`git diff --check` 與 Docker 容器清理完成。
