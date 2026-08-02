# `partial` 池來源與骨架稽核：batch01（10 筆）

> 執行日：2026-08-02。輸入：`facts-batch01.json`、`bios-batch01.md`、
> `people.json` 與 `docs/design/21-biography-authoring-spec.md` §0、§2、§3、§6、§23.2–§23.4。
>
> 範圍：這是舊 `partial` 母體的第一個**結構與來源鏈稽核批次**；不重查網站、
> 不重新判定史實，也不改寫既有小傳。故它能證明「成文可追到已登錄的骨架與來源代號」，
> 不能替代逐筆原始史料覆核。

## 1. 結論

batch01 的十筆皆可閉合為「來源鏈、骨架、成文及可畫性一致」：

- 10/10 的 `id` 與遊戲姓名均通過 `bio_gate.py`；10/10 均有一篇成文。
- 所有欄位內出現的來源代號均在本批 `sources` 表宣告；每筆的 `verified_by`
  也都只引用已宣告來源，沒有懸空代號。
- 9 筆為 `medium`、1 筆（#179 許琨）為 `low`；沒有以零星材料強行升為 `high`。
- 83 個舊 `note` 與 21 段 `conflicts[].statements[].says` 均會被
  `bio_skeleton.py` 的成文視圖扣住，維持著作權隔離。依規格 §23.4，**不回頭遷移或
  改寫**這些既有查證紀錄。

本批所使用的 Tier A／B／C 資訊只以 `facts-batch01.json` 已登錄的來源為準；
`S-SUMMARY` 僅見於 #263 的未採用線索，未列入該人物的 `verified_by`，也未用作
成文依據。

## 2. 逐筆結果

| id | 人物 | 目前信心度 | 骨架中的已登錄來源 | 任職／事件 | 衝突／缺口 | 結論 |
|---:|---|---|---|---:|---|---|
| 26 | 戴岳 | `medium` | `S-GPOST`、`S-WPZH`、兩筆 Tier C 地方文史 | 12／2 | 2／3 | 通過；身分欄的低信心來源未被抬高。 |
| 61 | 劉佐龍 | `medium` | `S-GPOST`、`S-WPZH` | 11／3 | 2／3 | 通過；生卒與籍貫未知，未以任官序列反推。 |
| 144 | 葉開鑫 | `medium` | `S-GPOST`、`S-WPZH` | 9／7 | 0／4 | 通過；期別外或未互證材料留在骨架。 |
| 179 | 許琨 | `low` | `S-GPOST`、`S-WPZH` | 8／6 | 1／3 | 通過；同名記錄未強行接合，維持 `low`。 |
| 263 | 曹萬順 | `medium` | `S-GPOST`、`S-WPZH`；另有未採用 `S-SUMMARY` | 8／6 | 2／3 | 通過；摘要線索未進 `verified_by`。 |
| 327 | 谷良民 | `medium` | `S-GPOST`、`S-WPZH` | 5／5 | 1／3 | 通過；來源分歧與未解時序均有記帳。 |
| 362 | 張際鵬 | `medium` | `S-GPOST`、`S-WPZH` | 6／3 | 1／3 | 通過；生年不因可計算而補填。 |
| 385 | 蘇祖馨 | `medium` | `S-GPOST`、`S-WPZH` | 3／2 | 0／3 | 通過；材料少而篇幅短，沒有補造經歷。 |
| 395 | 黃維綱 | `medium` | `S-GPOST`、`S-WPZH` | 5／3 | 0／3 | 通過；追贈資料與生卒推算明確分開。 |
| 413 | 邵百昌 | `medium` | `S-GPOST`、`S-WPZH` | 6／1 | 1／4 | 通過；異體字正規化和原字說明均留在骨架。 |

表中的任職／事件、衝突／缺口數均直接由 `facts-batch01.json` 的
`posts`／`events`、`conflicts`／`gaps` 陣列計得；不是以成文長度推測。

## 3. 可重現檢查

在專案既有 Docker Python 工具鏈執行：

```sh
./tools/py.sh tools/bio_skeleton.py batch01 --audit
./tools/py.sh tools/bio_gate.py batch01
```

結果如下：

- 隔離視圖扣住 `note: 83`、`says: 21`，共 104 處自由文字；因此成文流程不會直接看到
  舊 `note` 或來源原話。
- `bio_gate.py batch01`：10 篇、最大 331.5 格（上限 340）、正文與顯示欄位不可畫字皆 0、
  `id`／`name_ingame` 全符、禁用詞 0。
- 額外的 JSON 結構稽核：十筆欄位來源與 `verified_by` 均是批次 `sources` 表內的代號，
  懸空代號 0。

## 4. 未涵蓋事項

1. 這十筆的舊 `note` 仍是 legacy schema；規格 §23.4 明定舊批不回頭改，避免把查證
   紀錄在沒有原始來源重讀的情況下錯分成 `note_fact` 或 `note_source`。
2. 本稽核沒有連線重抓 gpost 或 MediaWiki，故不把來源可及性、內容正確性或最新頁面狀態
   誤報成已驗證；該工作需依 §26.4 單一作業存取 gpost 的限制另行安排。
3. `partial` 舊母體尚有 86 筆未經同類稽核。這是審核待辦，不是 86 位未立傳人物。

