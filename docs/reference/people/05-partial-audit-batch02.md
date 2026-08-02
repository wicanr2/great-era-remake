# `partial` 池來源與骨架稽核：batch02（10 筆）

> 執行日：2026-08-02。輸入：`facts-batch02.json`、`bios-batch02.md`、
> `people.json` 與 `docs/design/21-biography-authoring-spec.md` §0、§2、§3、§6、§23.2–§23.4。
>
> 範圍：本批的十人都在舊 `people.json` 標作 `partial`。本檔完成的是
> **結構、來源代號、成文存在性及隔離視圖**的機械稽核；不重新連線抓取史料、
> 不重判史實，也不改寫既有小傳。

## 1. 結論

batch02 的十筆均可閉合為「來源鏈、骨架、成文及可畫性一致」：

- 10/10 的 `id` 與遊戲姓名均通過 `bio_gate.py`；10/10 均有一篇成文。
- 所有欄位內出現的來源代號，以及每筆 `verified_by`，均在本批 `sources` 表宣告；
  懸空代號 0。
- 重新依完整骨架彙總的信心度為 `high` 4、`medium` 3、`low` 3；舊的
  `partial` 是歷史分類，不能覆蓋這一輪逐欄來源鏈的結果。
- 112 個舊 `note` 與 24 段 `conflicts[].statements[].says` 均會被
  `bio_skeleton.py` 的成文視圖扣住，維持著作權隔離。依規格 §23.4，**不回頭遷移或
  改寫**這些既有查證紀錄。

本檔只以 `facts-batch02.json` 已登錄的 `S-GPOST`／`S-WPZH` 為來源鏈證據。
其中 #38 李群僅有 `S-WPZH`，故維持 `low`，沒有把「條目存在」錯當成足以升級的互證。

## 2. 逐筆結果

| id | 人物 | 目前信心度 | 骨架中的已登錄來源 | 任職／事件 | 衝突／缺口 | 結論 |
|---:|---|---|---|---:|---|---|
| 10 | 馮軼裴 | `high` | `S-GPOST`、`S-WPZH` | 9／7 | 1／4 | 通過；兩類來源均有欄位錨點。 |
| 22 | 豐玉璽 | `high` | `S-GPOST`、`S-WPZH` | 6／2 | 1／3 | 通過；同名與時序問題留在衝突欄。 |
| 27 | 謝慕韓 | `low` | `S-GPOST`、`S-WPZH` | 4／0 | 2／4 | 通過；候選無法安全接合時維持 `low`。 |
| 28 | 朱世貴 | `high` | `S-GPOST`、`S-WPZH` | 6／3 | 0／4 | 通過；未以缺少身分欄補造敘述。 |
| 29 | 張言傳 | `low` | `S-GPOST`、`S-WPZH` | 6／1 | 1／4 | 通過；有限材料與身分缺口均如實保留。 |
| 38 | 李群 | `low` | `S-WPZH` | 2／1 | 0／4 | 通過；沒有把單一 Tier B 線索升格為互證。 |
| 40 | 練炳章 | `medium` | `S-GPOST`、`S-WPZH` | 4／5 | 1／4 | 通過；衝突未被壓平成確定斷言。 |
| 43 | 胡謙 | `high` | `S-GPOST`、`S-WPZH` | 12／3 | 2／4 | 通過；大量職務仍保留來源與衝突索引。 |
| 46 | 席液池 | `medium` | `S-GPOST`、`S-WPZH` | 5／9 | 1／4 | 通過；事件材料與職務材料分欄記錄。 |
| 47 | 萬選才 | `medium` | `S-GPOST`、`S-WPZH` | 9／4 | 2／5 | 通過；來源內部差異沒有被靜默裁決。 |

表中的任職／事件、衝突／缺口數均直接由 `facts-batch02.json` 的
`posts`／`events`、`conflicts`／`gaps` 陣列計得；不是以小傳篇幅推測。

## 3. 可重現檢查

在專案既有 Docker Python 工具鏈執行：

```sh
./tools/py.sh tools/bio_skeleton.py batch02 --audit
./tools/py.sh tools/bio_gate.py batch02
```

結果如下：

- 隔離視圖扣住 `note: 112`、`says: 24`，共 136 處自由文字；成文流程不會直接看到
  舊 `note` 或來源原話。
- `bio_gate.py batch02`：10 篇、最大 328.5 格（上限 340）、正文與顯示欄位不可畫字皆 0、
  `id`／`name_ingame` 全符、禁用詞 0。
- 額外的 JSON 結構稽核：十筆欄位來源與 `verified_by` 均是批次 `sources` 表內的代號，
  懸空代號 0。

## 4. 未涵蓋事項

1. 這十筆的舊 `note` 仍是 legacy schema；規格 §23.4 明定舊批不回頭改，避免在沒有
   原始來源重讀的情況下錯分成 `note_fact` 或 `note_source`。
2. 本稽核沒有連線重抓 gpost 或 MediaWiki，故不把來源可及性、內容正確性或最新頁面狀態
   誤報成已驗證；該工作需依 §26.4 單一作業存取 gpost 的限制另行安排。
3. 完成本批後，舊 `partial` 母體尚有 76 筆未經同類稽核。這是審核待辦，不是 76 位
   未立傳人物。

