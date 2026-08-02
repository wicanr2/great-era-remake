# `partial` 池來源與骨架稽核：batch08（10 筆）

> 執行日：2026-08-02。輸入：`facts-batch08.json`、`bios-batch08.md`、
> `people.json` 與 `docs/design/21-biography-authoring-spec.md` §0、§2、§3、§6、§23.2–§23.4。
>
> 範圍：本批十人都在舊 `people.json` 標作 `partial`。本檔完成的是
> **結構、來源代號、成文存在性及隔離視圖**的機械稽核；不重新連線抓取史料、
> 不重判史實，也不改寫既有小傳。

## 1. 結論

batch08 的十筆均已完成成文、槽位與可畫性檢查：

- 10/10 的 `id` 與遊戲姓名均通過 `bio_gate.py`；10/10 均有一篇成文。
- 欄位與衝突陳述實際使用的來源代號全數在本批 `sources` 表宣告，故**懸空來源 0**。
- 完整骨架的信心度為 `high` 6、`medium` 3、`low` 1；舊 `partial` 是歷史分類，
  不能覆蓋這一輪逐欄來源鏈的結果。
- 224 個舊 `note` 與 41 段 `conflicts[].statements[].says` 都會被
  `bio_skeleton.py` 的成文視圖扣住，維持著作權隔離。

### 一筆 identity 來源索引例外

#387 陳鼎勳的 `verified_by` 含 `S-ROSTER`，其用途僅是原版名冊的 identity 旁證；
現行 `identity.disambiguation` 沒有欄位級 `src`，故該來源不能由通用欄位掃描直接抓到。
它不是懸空代號，也未被當作史實陳述或信心度升級依據。#297 李家銓的 `S-ROSTER`
則已在結構化衝突欄中錨定。

本批另有 `S-RMRB`、`S-WPLIST`、`S-WPDIS`、`S-WPCOM` 等來源；皆已在來源表登錄。
其中清單型來源未因「列有姓名」而單獨計作信心度互證，仍以骨架中的逐欄來源分級為準。

## 2. 逐筆結果

| id | 人物 | 目前信心度 | 骨架中的已登錄來源 | 任職／事件 | 衝突／缺口 | 結論 |
|---:|---|---|---|---:|---|---|
| 267 | 阮玄武 | `high` | `S-GPOST`、`S-WPZH` | 16／4 | 2／3 | 通過；高材料量仍保留衝突索引。 |
| 268 | 鮑剛 | `medium` | `S-GPOST`、`S-WPZH`、`S-WPLIST` | 13／4 | 3／3 | 通過；清單來源未獨立抬升評等。 |
| 279 | 孫桐萱 | `high` | `S-GPOST`、`S-WPZH`、`S-RMRB` | 17／7 | 2／3 | 通過；生卒與職務分欄追溯。 |
| 280 | 曹福林 | `high` | `S-GPOST`、`S-WPZH` | 11／5 | 2／3 | 通過；來源分歧未被靜默裁決。 |
| 297 | 李家銓 | `low` | `S-GPOST`、`S-WPZH`、`S-ROSTER` | 16／1 | 1／3 | 通過；遊戲寫法與人名接合維持待考。 |
| 387 | 陳鼎勳 | `high` | `S-GPOST`、`S-WPZH`、`S-WPLIST`、`S-ROSTER`（identity） | 16／3 | 2／4 | 通過；identity 來源索引例外，見上節。 |
| 262 | 楊勝治 | `high` | `S-GPOST`、`S-WPZH` | 7／5 | 1／3 | 通過；交叉來源均有欄位錨點。 |
| 370 | 劉效曾 | `medium` | `S-GPOST`、`S-WPZH` | 8／3 | 2／5 | 通過；資料缺口如實保留。 |
| 265 | 陳焯 | `high` | `S-GPOST`、`S-WPZH`、`S-WPDIS` | 15／4 | 2／4 | 通過；第三來源可逐欄追溯。 |
| 283 | 郜子舉 | `medium` | `S-GPOST`、`S-WPZH`、`S-WPCOM` | 12／4 | 2／4 | 通過；異寫與來源層級均有記帳。 |

表中的任職／事件、衝突／缺口數均直接由 `facts-batch08.json` 的
`posts`／`events`、`conflicts`／`gaps` 陣列計得；不是以小傳篇幅推測。

## 3. 可重現檢查

在專案既有 Docker Python 工具鏈執行：

```sh
./tools/py.sh tools/bio_skeleton.py batch08 --audit
./tools/py.sh tools/bio_gate.py batch08
```

結果如下：

- 隔離視圖扣住 `note: 224`、`says: 41`，共 265 處自由文字；成文流程不會直接看到
  舊 `note` 或來源原話。
- `bio_gate.py batch08`：10 篇、最大 338 格（上限 340）、正文與顯示欄位不可畫字皆 0、
  `id`／`name_ingame` 全符、禁用詞 0。
- 額外的 JSON 結構稽核：欄位與衝突陳述的 `src` 懸空代號 0；#387 的例外只涉及
  identity 旁證的索引位置，沒有把 `S-ROSTER` 當欄位級事實。

## 4. 未涵蓋事項

1. 這十筆的舊 `note` 仍是 legacy schema；規格 §23.4 明定舊批不回頭改，避免在沒有
   原始來源重讀的情況下錯分成 `note_fact` 或 `note_source`。
2. 本稽核沒有連線重抓 gpost 或 MediaWiki，故不把來源可及性、內容正確性或最新頁面狀態
   誤報成已驗證；清單型來源亦不取代個人身分分群。
3. 完成本批後，舊 `partial` 母體尚有 16 筆未經同類稽核。這是審核待辦，不是 16 位
   未立傳人物。

