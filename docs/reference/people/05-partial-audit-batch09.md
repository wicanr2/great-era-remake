# `partial` 池來源與骨架稽核：batch09（10 筆）

> 執行日：2026-08-02。輸入：`facts-batch09.json`、`bios-batch09.md`、
> `people.json` 與 `docs/design/21-biography-authoring-spec.md` §0、§2、§3、§6、§23.2–§23.4。
>
> 範圍：本批十人都在舊 `people.json` 標作 `partial`。本檔完成的是
> **結構、來源代號、成文存在性及隔離視圖**的機械稽核；不重新連線抓取史料、
> 不重判史實，也不改寫既有小傳。

## 1. 結論

batch09 的十筆均可閉合為「來源鏈、骨架、成文及可畫性一致」：

- 10/10 的 `id` 與遊戲姓名均通過 `bio_gate.py`；10/10 均有一篇成文。
- 欄位與衝突陳述實際使用的來源代號，以及每筆 `verified_by`，均在本批 `sources`
  表宣告；懸空代號 0，且沒有未錨定的 `verified_by`。
- 完整骨架的信心度為 `high` 7、`medium` 3；舊 `partial` 是歷史分類，不能覆蓋
  這一輪逐欄來源鏈的結果。
- 164 個舊 `note` 與 22 段 `conflicts[].statements[].says` 均會被
  `bio_skeleton.py` 的成文視圖扣住，維持著作權隔離。依規格 §23.4，**不回頭遷移或
  改寫**這些既有查證紀錄。

本批除 `S-GPOST`／`S-WPZH` 外，使用 `S-WPLIST`（#292、#343、#408、#388、#390、
#391、#396）及 #364 的 `S-PJSON`；皆已在來源表登錄並具欄位錨點。清單型來源不因
「列有姓名」而單獨計作信心度互證，仍以骨架中的逐欄來源分級為準。沒有需要另立例外的個案。

## 2. 逐筆結果

| id | 人物 | 目前信心度 | 骨架中的已登錄來源 | 任職／事件 | 衝突／缺口 | 結論 |
|---:|---|---|---|---:|---|---|
| 292 | 曾萬鍾 | `high` | `S-GPOST`、`S-WPZH`、`S-WPLIST` | 10／6 | 1／4 | 通過；生卒與任職分欄追溯。 |
| 343 | 孔令恂 | `high` | `S-GPOST`、`S-WPZH`、`S-WPLIST` | 11／1 | 1／4 | 通過；清單來源未單獨抬升評等。 |
| 364 | 陳瑞河 | `medium` | `S-GPOST`、`S-WPZH`、`S-PJSON` | 7／1 | 1／4 | 通過；第三來源可逐欄追溯。 |
| 402 | 黃正成 | `medium` | `S-GPOST`、`S-WPZH` | 5／2 | 1／5 | 通過；資料缺口如實保留。 |
| 408 | 鐘彬 | `high` | `S-GPOST`、`S-WPZH`、`S-WPLIST` | 12／3 | 2／5 | 通過；異寫與衝突均有索引。 |
| 384 | 劉尚志 | `high` | `S-GPOST`、`S-WPZH` | 8／3 | 1／4 | 通過；來源鏈與成文一致。 |
| 388 | 方靖 | `high` | `S-GPOST`、`S-WPZH`、`S-WPLIST` | 10／2 | 1／4 | 通過；未解內容未以推定補足。 |
| 390 | 柳際明 | `high` | `S-GPOST`、`S-WPZH`、`S-WPLIST` | 9／1 | 1／6 | 通過；時序爭議保持可回查。 |
| 391 | 宋肯堂 | `high` | `S-GPOST`、`S-WPZH`、`S-WPLIST` | 9／5 | 1／5 | 通過；事件與任職材料分欄記錄。 |
| 396 | 高卓東 | `medium` | `S-GPOST`、`S-WPZH`、`S-WPLIST` | 8／0 | 1／5 | 通過；事件為空不補造背景敘述。 |

表中的任職／事件、衝突／缺口數均直接由 `facts-batch09.json` 的
`posts`／`events`、`conflicts`／`gaps` 陣列計得；不是以小傳篇幅推測。

## 3. 可重現檢查

在專案既有 Docker Python 工具鏈執行：

```sh
./tools/py.sh tools/bio_skeleton.py batch09 --audit
./tools/py.sh tools/bio_gate.py batch09
```

結果如下：

- 隔離視圖扣住 `note: 164`、`says: 22`，共 186 處自由文字；成文流程不會直接看到
  舊 `note` 或來源原話。
- `bio_gate.py batch09`：10 篇、最大 340 格（上限 340）、正文與顯示欄位不可畫字皆 0、
  `id`／`name_ingame` 全符、禁用詞 0。
- 額外的 JSON 結構稽核：欄位與衝突陳述的 `src`、`verified_by` 懸空代號均為 0，且
  `verified_by` 全數可在對應人物的欄位級來源引用中找到。

## 4. 未涵蓋事項

1. 這十筆的舊 `note` 仍是 legacy schema；規格 §23.4 明定舊批不回頭改，避免在沒有
   原始來源重讀的情況下錯分成 `note_fact` 或 `note_source`。
2. 本稽核沒有連線重抓 gpost 或 MediaWiki，故不把來源可及性、內容正確性或最新頁面狀態
   誤報成已驗證；清單型來源亦不取代個人身分分群。
3. 完成本批後，舊 `partial` 母體尚有 6 筆未經同類稽核。這是審核待辦，不是 6 位
   未立傳人物。

