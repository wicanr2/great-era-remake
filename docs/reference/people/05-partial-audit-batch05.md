# `partial` 池來源與骨架稽核：batch05（10 筆）

> 執行日：2026-08-02。輸入：`facts-batch05.json`、`bios-batch05.md`、
> `people.json` 與 `docs/design/21-biography-authoring-spec.md` §0、§2、§3、§6、§23.2–§23.4。
>
> 範圍：本批十人都在舊 `people.json` 標作 `partial`。本檔完成的是
> **結構、來源代號、成文存在性及隔離視圖**的機械稽核；不重新連線抓取史料、
> 不重判史實，也不改寫既有小傳。

## 1. 結論

batch05 的十筆均已完成成文、槽位與可畫性檢查：

- 10/10 的 `id` 與遊戲姓名均通過 `bio_gate.py`；10/10 均有一篇成文。
- 欄位與衝突陳述實際使用的來源代號全數在本批 `sources` 表宣告，故**懸空來源 0**。
- 完整骨架的信心度為 `high` 6、`medium` 2、`low` 2；舊 `partial` 是歷史分類，
  不能覆蓋這一輪逐欄來源鏈的結果。
- 100 個舊 `note` 與 39 段 `conflicts[].statements[].says` 都會被
  `bio_skeleton.py` 的成文視圖扣住，維持著作權隔離。

### 身分辨識來源的 schema 例外

本批直魯聯軍人物使用 `S-ROSTER`（原版名冊順序）只作 identity 旁證，符合規格
§11.6「不進正文」的限制；但 `identity.disambiguation` 沒有欄位級 `src`，造成五筆
`verified_by` 的 `S-ROSTER`（#176、#177、#178、#181、#182）只能在該自由文字中回查。

#180 柴榮陞的 `S-GPOST` 亦只出現在 `identity.disambiguation`／`gaps`：它記錄的是五種
姓名寫法全數零命中的**查詢結果**，不是可放進 `posts` 的史實。這不是懸空代號，也不能被
錯當成欄位級 Tier A 證據。#183 劉志陞的 `S-ROSTER` 則已在 `conflicts` 的結構化
`src` 中錨定。

這六筆顯示的不是來源缺失，而是現行 schema 對 identity 來源的索引不足；本稽核依限制
不回頭改既有骨架，僅留下可重讀原始來源時的修正線索。

## 2. 逐筆結果

| id | 人物 | 目前信心度 | 骨架中的已登錄來源 | 任職／事件 | 衝突／缺口 | 結論 |
|---:|---|---|---|---:|---|---|
| 139 | 林忠 | `medium` | `S-GPOST`、`S-WPZH` | 7／2 | 2／4 | 通過；有限的一手資料未被誇大。 |
| 143 | 劉鉶 | `high` | `S-GPOST`、`S-WPZH` | 5／1 | 0／4 | 通過；欄位級來源均可追溯。 |
| 147 | 李燊 | `high` | `S-GPOST`、`S-WPZH` | 7／3 | 2／4 | 通過；來源分歧留在衝突欄。 |
| 176 | 程國瑞 | `high` | `S-GPOST`、`S-WPZH`、`S-ROSTER`（identity） | 8／8 | 2／4 | 通過；identity 來源索引例外，見上節。 |
| 177 | 方永昌 | `high` | `S-GPOST`、`S-WPZH`、`S-ROSTER`（identity） | 7／9 | 2／4 | 通過；identity 來源索引例外，見上節。 |
| 178 | 王棟 | `high` | `S-GPOST`、`S-WPZH`、`S-ROSTER`（identity） | 6／6 | 1／4 | 通過；同名分群沒有被壓成確定事實。 |
| 180 | 柴榮陞 | `low` | `S-WPZH`；`S-GPOST` 為零命中紀錄，`S-ROSTER` 為 identity 旁證 | 6／0 | 1／4 | 通過；候選未獲接合，維持 `low`。 |
| 181 | 吳奠卿 | `medium` | `S-GPOST`、`S-WPZH`、`S-ROSTER`（identity） | 4／3 | 1／4 | 通過；軍番號爭議沒有靜默裁決。 |
| 182 | 張宗輔 | `high` | `S-GPOST`、`S-WPZH`、`S-ROSTER`（identity） | 5／2 | 1／4 | 通過；時序衝突保持可回查。 |
| 183 | 劉志陞 | `low` | `S-GPOST`、`S-WPZH`、`S-ROSTER` | 8／7 | 2／4 | 通過；遊戲寫法與通行名的接合仍標為待考。 |

表中的任職／事件、衝突／缺口數均直接由 `facts-batch05.json` 的
`posts`／`events`、`conflicts`／`gaps` 陣列計得；不是以小傳篇幅推測。

## 3. 可重現檢查

在專案既有 Docker Python 工具鏈執行：

```sh
./tools/py.sh tools/bio_skeleton.py batch05 --audit
./tools/py.sh tools/bio_gate.py batch05
```

結果如下：

- 隔離視圖扣住 `note: 100`、`says: 39`，共 139 處自由文字；成文流程不會直接看到
  舊 `note` 或來源原話。
- `bio_gate.py batch05`：10 篇、最大 325 格（上限 340）、正文與顯示欄位不可畫字皆 0、
  `id`／`name_ingame` 全符、禁用詞 0。
- 額外的 JSON 結構稽核：欄位與衝突陳述的 `src` 懸空代號 0；上述六筆例外只涉及
  identity／零命中查詢的來源索引位置，並未把 `S-ROSTER` 用來主張史實。

## 4. 未涵蓋事項

1. 這十筆的舊 `note` 仍是 legacy schema；規格 §23.4 明定舊批不回頭改，避免在沒有
   原始來源重讀的情況下錯分成 `note_fact` 或 `note_source`。
2. 本稽核沒有連線重抓 gpost 或 MediaWiki，故不把來源可及性、內容正確性或最新頁面狀態
   誤報成已驗證；identity 的 `S-ROSTER` 旁證也不能取代外部史料。
3. 完成本批後，舊 `partial` 母體尚有 46 筆未經同類稽核。這是審核待辦，不是 46 位
   未立傳人物。

