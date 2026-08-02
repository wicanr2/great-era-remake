# `partial` 池來源與骨架稽核：batch04（10 筆）

> 執行日：2026-08-02。輸入：`facts-batch04.json`、`bios-batch04.md`、
> `people.json` 與 `docs/design/21-biography-authoring-spec.md` §0、§2、§3、§6、§23.2–§23.4。
>
> 範圍：本批十人都在舊 `people.json` 標作 `partial`。本檔完成的是
> **結構、來源代號、成文存在性及隔離視圖**的機械稽核；不重新連線抓取史料、
> 不重判史實，也不改寫既有小傳。

## 1. 結論

batch04 的十筆均可閉合為「來源鏈、骨架、成文及可畫性一致」：

- 10/10 的 `id` 與遊戲姓名均通過 `bio_gate.py`；10/10 均有一篇成文。
- 所有欄位內出現的來源代號，以及每筆 `verified_by`，均在本批 `sources` 表宣告；
  懸空代號 0，且沒有未錨定的 `verified_by`。
- 完整骨架的信心度為 `high` 6、`medium` 2、`low` 2；舊 `partial` 是歷史分類，
  不能覆蓋這一輪逐欄來源鏈的結果。
- 117 個舊 `note` 與 37 段 `conflicts[].statements[].says` 均會被
  `bio_skeleton.py` 的成文視圖扣住，維持著作權隔離。依規格 §23.4，**不回頭遷移或
  改寫**這些既有查證紀錄。

本批來源鏈以 `S-GPOST`、`S-WPZH` 為主；#138 王麒另有 `S-WPDIS`。
三者均在批次來源表登錄，且皆有欄位級引用，沒有以來源表的存在取代欄位證據。

## 2. 逐筆結果

| id | 人物 | 目前信心度 | 骨架中的已登錄來源 | 任職／事件 | 衝突／缺口 | 結論 |
|---:|---|---|---|---:|---|---|
| 111 | 馮紹閔 | `low` | `S-GPOST`、`S-WPZH` | 12／5 | 2／3 | 通過；同名與身分缺口未被硬接合。 |
| 116 | 馬祥斌 | `medium` | `S-GPOST`、`S-WPZH` | 6／4 | 3／4 | 通過；多項分歧均留在衝突欄。 |
| 120 | 蔣鎮臣 | `high` | `S-GPOST`、`S-WPZH` | 6／4 | 2／4 | 通過；交叉來源均有欄位錨點。 |
| 121 | 唐福山 | `high` | `S-GPOST`、`S-WPZH` | 6／4 | 2／4 | 通過；生卒與任職材料分欄追溯。 |
| 123 | 劉寶題 | `high` | `S-GPOST`、`S-WPZH` | 7／3 | 3／4 | 通過；來源分歧未被靜默裁決。 |
| 127 | 陳修爵 | `low` | `S-GPOST`、`S-WPZH` | 4／5 | 2／5 | 通過；有限材料維持 `low`。 |
| 130 | 劉俊 | `high` | `S-GPOST`、`S-WPZH` | 5／3 | 1／4 | 通過；未解資料不以推定補足。 |
| 133 | 蔣啟鳳 | `high` | `S-GPOST`、`S-WPZH` | 6／0 | 0／4 | 通過；事件為空不以背景敘述替代。 |
| 136 | 李鳳翔 | `high` | `S-GPOST`、`S-WPZH` | 4／3 | 0／4 | 通過；成文與結構來源一致。 |
| 138 | 王麒 | `medium` | `S-GPOST`、`S-WPZH`、`S-WPDIS` | 5／0 | 2／5 | 通過；第三來源已逐欄錨定。 |

表中的任職／事件、衝突／缺口數均直接由 `facts-batch04.json` 的
`posts`／`events`、`conflicts`／`gaps` 陣列計得；不是以小傳篇幅推測。

## 3. 可重現檢查

在專案既有 Docker Python 工具鏈執行：

```sh
./tools/py.sh tools/bio_skeleton.py batch04 --audit
./tools/py.sh tools/bio_gate.py batch04
```

結果如下：

- 隔離視圖扣住 `note: 117`、`says: 37`，共 154 處自由文字；成文流程不會直接看到
  舊 `note` 或來源原話。
- `bio_gate.py batch04`：10 篇、最大 332 格（上限 340）、正文與顯示欄位不可畫字皆 0、
  `id`／`name_ingame` 全符、禁用詞 0。
- 額外的 JSON 結構稽核：欄位 `src` 與 `verified_by` 的懸空代號均為 0，且
  `verified_by` 全數可在該人物的欄位級來源引用中找到。

## 4. 未涵蓋事項

1. 這十筆的舊 `note` 仍是 legacy schema；規格 §23.4 明定舊批不回頭改，避免在沒有
   原始來源重讀的情況下錯分成 `note_fact` 或 `note_source`。
2. 本稽核沒有連線重抓 gpost 或 MediaWiki，故不把來源可及性、內容正確性或最新頁面狀態
   誤報成已驗證；該工作需依 §26.4 單一作業存取 gpost 的限制另行安排。
3. 完成本批後，舊 `partial` 母體尚有 56 筆未經同類稽核。這是審核待辦，不是 56 位
   未立傳人物。

