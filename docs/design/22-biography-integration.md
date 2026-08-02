# DESIGN-22 人物自傳可追溯整合提案

狀態：**DRAFT（流程提案；尚未解除 DESIGN-21 §9 的寫回禁令）**  
日期：2026-08-02

2026-08-03 實作進度：`tools/gen_authored_bios.py` 已完成本提案第 2、3 節的
唯讀／暫存輸出階段。`--check` 會失敗即關閉地驗證 417 筆骨架、387 篇正文、
30 筆 unknown、#274 佔位排除、姓名／信心度一致、可畫性、格數、禁用詞與來源
SHA-256；`--output /tmp/...` 連跑兩次可產生位元組相同的索引。**尚未建立版控內的
`authored-bios.json`，也未接 `gen_people.py` 或寫回產品資料**，所以本文件仍是 DRAFT。

## 0. 現況與問題

- 417/417 個人物槽位已有唯一事實骨架。
- 30 位經查證後為 `unknown`，依 DESIGN-21 不立傳。
- 排除 `unknown` 後，387/387 位已有唯一成文；跨批次重複 0、未做 0。
- 執行期權威來源 `docs/reference/people/people.json` 仍只有 326 篇非空
  `bio_zh`，所以產品缺口是 **61 篇整合**，不是 61 篇重寫。

直接把 `bios-batch*.md` 複製進 `people.json` 會丟失「這篇來自哪一份骨架、哪一版
成文、是否通過品質閘」的資訊，也違反 DESIGN-21 §9 的 DRAFT 禁令。本提案只定義
可重生、可稽核的流程；未經使用者裁決不得執行產品寫回。

## 1. 建議裁決

採 DESIGN-21 V1 所傾向的「**研究稿獨立保存，由產生器合併**」：

```text
facts-<batch>.json + bios-<batch>.md
                  │ 嚴格解析、品質閘、雜湊
                  ▼
authored-bios.json（可重生的來源索引，不手改）
                  │ 只覆蓋明列欄位
people.json ──────┴─→ translations/zh-Hant/people.json
```

不整批覆寫原研究 `people.json`，讓舊版 326 篇仍可比較；發行資料由
`tools/gen_people.py` 合併。這項裁決仍需使用者確認後，才能把本文件升為 READY。

## 2. 嚴格擷取契約

新增 `tools/gen_authored_bios.py`，只讀所有成對的 `facts-<batch>.json` 與
`bios-<batch>.md`：

1. 每個 facts ID 必須只有一個批次擁有；#274「無省長」不得進人物輸出。
2. `confidence == unknown` 必須沒有正文；其餘每位必須恰有一個
   `## #<id> <name_ingame>` 標題及恰有一段正文。
3. 標題 ID、姓名與 facts、原 `people.json` 三方須完全一致。
4. 正文必須通過現有格數、可畫性與禁用詞門檻；產生器不得自行修字或潤稿。
5. 任一缺檔、重複、額外正文、姓名漂移或未通過門檻均失敗即關閉
   （fail-closed），不得留下部分輸出。

## 3. `authored-bios.json` 最小 schema

```jsonc
{
  "schema_version": "1",
  "generated_from": "versioned research batches",
  "people": [
    {
      "id": 13,
      "name_ingame": "蔡仲笏",
      "bio_zh": "……",
      "confidence": "low",
      "provenance": {
        "facts": "docs/reference/people/facts-batchU2.json",
        "facts_sha256": "…",
        "bios": "docs/reference/people/bios-batchU2.md",
        "bios_sha256": "…"
      }
    }
  ]
}
```

輸出順序固定為人物 ID；JSON 序列化固定縮排與尾端換行。`generated` 不放即時時間，
避免相同輸入產生不同結果。來源雜湊採整檔 SHA-256，另可加正文 SHA-256 供逐篇比較。

## 4. 合併白名單

第一階段只允許覆蓋：

- `bio_zh`
- `confidence`（取 facts 的人物層自撰信心度）

`name_common`、`courtesy`、`highest_post`、`faction` 等欄位雖也可能有新版資料，
但各批 facts schema 的欄位形狀不完全相同；沒有獨立欄位對照與測試前不得順手覆蓋。
這些欄位另列第二階段遷移，不得用猜測式轉換。

## 5. 發行合併與帳本

`tools/gen_people.py` 合併時應：

1. 先載入舊 `people.json`，再依 ID 套用 `authored-bios.json` 白名單欄位。
2. 檢查 30 位 `unknown` 仍為空、#274 仍排除、387 位正文皆非空。
3. 沿用既有可畫字正規化，並把正規化前後與來源正文雜湊寫入帳本。
4. 在發行檔頂層加入 `authored_source_sha256`；不得把研究來源網址或抓取原文打包。
5. 以暫存檔完整驗證後再原子替換三份產物，任何錯誤不得只更新其中一份。

## 6. 驗收證據

- `bio_gate.py --all` 與 `--cross` 全過。
- `gen_authored_bios.py --check` 證明工作樹產物可由研究批次位元組重生。
- 唯一骨架 417、`unknown` 30、唯一正文 387、執行期正文 387。
- 新增 61 篇逐 ID 清單；既有 326 篇合併前後正文 SHA-256 相同，除非另有明列遷移。
- 387 篇全部通過 `textlayout` 分頁及倚天可畫性測試。
- GUI 至少抽驗新增的短、中、最長正文，以及 `low`、`medium`、`high` 各一篇。
- 產生器連跑兩次後 `git diff` 為空，證明輸出可重現。

## 7. 回復方式

發行資料完全由來源重生；回復時移除合併輸入並重跑產生器即可，不需手工從
`people.json` 挑回舊段落。研究批次與舊 `people.json` 始終保留，方便逐篇比較與勘誤。

## 8. 待使用者裁決

唯一會改變資料權威關係的決定：是否採用本提案的「獨立來源索引＋產生器合併」，
而不是直接覆寫 `docs/reference/people/people.json`。在裁決前，本文件維持 DRAFT，
不建立或寫入 `authored-bios.json`。

裁決前可安全執行：

```sh
tools/py.sh tools/gen_authored_bios.py --check
tools/py.sh tools/gen_authored_bios.py --output /tmp/authored-bios.json
```
