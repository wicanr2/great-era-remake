# `partial` 池來源與骨架稽核：batch07（10 筆）

> 執行日：2026-08-02。輸入：`facts-batch07.json`、`bios-batch07.md`、
> `people.json` 與 `docs/design/21-biography-authoring-spec.md` §0、§2、§3、§6、§23.2–§23.4。
>
> 範圍：本批十人都在舊 `people.json` 標作 `partial`。本檔完成的是
> **結構、來源代號、成文存在性及隔離視圖**的機械稽核；不重新連線抓取史料、
> 不重判史實，也不改寫既有小傳。

## 1. 結論

batch07 的十筆均可閉合為「來源鏈、骨架、成文及可畫性一致」：

- 10/10 的 `id` 與遊戲姓名均通過 `bio_gate.py`；10/10 均有一篇成文。
- 欄位與衝突陳述實際使用的來源代號，以及每筆 `verified_by`，均在本批 `sources`
  表宣告；懸空代號 0，且沒有未錨定的 `verified_by`。
- 完整骨架的信心度為 `high` 5、`medium` 5；舊 `partial` 是歷史分類，不能覆蓋
  這一輪逐欄來源鏈的結果。
- 142 個舊 `note` 與 33 段 `conflicts[].statements[].says` 均會被
  `bio_skeleton.py` 的成文視圖扣住，維持著作權隔離。依規格 §23.4，**不回頭遷移或
  改寫**這些既有查證紀錄。

本批除 `S-GPOST`、`S-WPZH` 外，#251 鍾祖培、#386 張光瑋與 #394 汪之斌使用
`S-WPNR`；三筆均在批次來源表登錄且均有欄位級錨點。沒有需要另立來源例外的個案。

## 2. 逐筆結果

| id | 人物 | 目前信心度 | 骨架中的已登錄來源 | 任職／事件 | 衝突／缺口 | 結論 |
|---:|---|---|---|---:|---|---|
| 323 | 張達 | `high` | `S-GPOST`、`S-WPZH` | 8／2 | 2／3 | 通過；時序衝突均可回查。 |
| 324 | 張瑞貴 | `high` | `S-GPOST`、`S-WPZH` | 7／5 | 1／4 | 通過；身分與事件材料分欄追溯。 |
| 325 | 李振球 | `high` | `S-GPOST`、`S-WPZH` | 5／2 | 1／4 | 通過；交叉來源均有欄位錨點。 |
| 397 | 黃濤 | `medium` | `S-GPOST`、`S-WPZH` | 4／8 | 1／4 | 通過；嚴格分群後維持 `medium`。 |
| 398 | 黃國樑 | `medium` | `S-GPOST`、`S-WPZH` | 7／5 | 1／5 | 通過；跨系統職務未被錯接。 |
| 251 | 鍾祖培 | `medium` | `S-GPOST`、`S-WPZH`、`S-WPNR` | 4／5 | 1／5 | 通過；第三來源已逐欄錨定。 |
| 382 | 黎行恕 | `medium` | `S-GPOST`、`S-WPZH` | 5／5 | 2／5 | 通過；衝突未被壓成確定斷言。 |
| 386 | 張光瑋 | `medium` | `S-GPOST`、`S-WPZH`、`S-WPNR` | 8／3 | 2／5 | 通過；異寫材料仍保留分群紀錄。 |
| 311 | 陶廣 | `high` | `S-GPOST`、`S-WPZH` | 9／5 | 2／5 | 通過；來源鏈與成文一致。 |
| 394 | 汪之斌 | `high` | `S-GPOST`、`S-WPZH`、`S-WPNR` | 7／5 | 2／6 | 通過；地名分歧保留原始索引。 |

表中的任職／事件、衝突／缺口數均直接由 `facts-batch07.json` 的
`posts`／`events`、`conflicts`／`gaps` 陣列計得；不是以小傳篇幅推測。

## 3. 可重現檢查

在專案既有 Docker Python 工具鏈執行：

```sh
./tools/py.sh tools/bio_skeleton.py batch07 --audit
./tools/py.sh tools/bio_gate.py batch07
```

結果如下：

- 隔離視圖扣住 `note: 142`、`says: 33`，共 175 處自由文字；成文流程不會直接看到
  舊 `note` 或來源原話。
- `bio_gate.py batch07`：10 篇、最大 328.5 格（上限 340）、正文與顯示欄位不可畫字皆 0、
  `id`／`name_ingame` 全符、禁用詞 0。
- 額外的 JSON 結構稽核：欄位與衝突陳述的 `src`、`verified_by` 懸空代號均為 0，且
  `verified_by` 全數可在對應人物的欄位級來源引用中找到。

## 4. 未涵蓋事項

1. 這十筆的舊 `note` 仍是 legacy schema；規格 §23.4 明定舊批不回頭改，避免在沒有
   原始來源重讀的情況下錯分成 `note_fact` 或 `note_source`。
2. 本稽核沒有連線重抓 gpost 或 MediaWiki，故不把來源可及性、內容正確性或最新頁面狀態
   誤報成已驗證；`S-WPNR` 的內容也只按既有骨架的來源分級使用。
3. 完成本批後，舊 `partial` 母體尚有 26 筆未經同類稽核。這是審核待辦，不是 26 位
   未立傳人物。

