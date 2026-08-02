# `partial` 池來源與骨架稽核：batch06（10 筆）

> 執行日：2026-08-02。輸入：`facts-batch06.json`、`bios-batch06.md`、
> `people.json` 與 `docs/design/21-biography-authoring-spec.md` §0、§2、§3、§6、§23.2–§23.4。
>
> 範圍：本批十人都在舊 `people.json` 標作 `partial`。本檔完成的是
> **結構、來源代號、成文存在性及隔離視圖**的機械稽核；不重新連線抓取史料、
> 不重判史實，也不改寫既有小傳。

## 1. 結論

batch06 的十筆均已完成成文、槽位與可畫性檢查：

- 10/10 的 `id` 與遊戲姓名均通過 `bio_gate.py`；10/10 均有一篇成文。
- 欄位與衝突陳述實際使用的來源代號全數在本批 `sources` 表宣告，故**懸空來源 0**。
- 完整骨架的信心度為 `high` 4、`medium` 4、`low` 2；舊 `partial` 是歷史分類，
  不能覆蓋這一輪逐欄來源鏈的結果。
- 131 個舊 `note` 與 34 段 `conflicts[].statements[].says` 都會被
  `bio_skeleton.py` 的成文視圖扣住，維持著作權隔離。

### 兩筆零命中不是 Tier A 證據

#185 許寶祥與 #187 毛思義的 `verified_by` 列有 `S-GPOST`，但該來源在兩筆中都是
重跑後確認的**零命中查詢結果**，僅記於 `identity.disambiguation`／`gaps`，沒有可寫入
`posts` 的任免令。因此不會出現在欄位級 `src`；這不是懸空代號，也不能被當作 Tier A
正面佐證。兩人分別維持 `low` 與 `medium`，沒有因「已查 gpost」而升級。

其他來源（`S-WPZH`、#219 的 `S-WPDIS`、#189／#264 的 `S-WPJPMA`，以及 #185 的
identity 旁證 `S-ROSTER`）都在批次來源表登錄；`S-ROSTER` 僅用於確認槽位對應，
不承載史實陳述。

## 2. 逐筆結果

| id | 人物 | 目前信心度 | 骨架中的已登錄來源 | 任職／事件 | 衝突／缺口 | 結論 |
|---:|---|---|---|---:|---|---|
| 185 | 許寶祥 | `low` | `S-WPZH`、`S-ROSTER`；`S-GPOST` 為零命中紀錄 | 1／0 | 2／4 | 通過；材料無法形成序列，維持 `low`。 |
| 201 | 鄭澤生 | `high` | `S-GPOST`、`S-WPZH` | 9／2 | 0／5 | 通過；交叉來源均有欄位錨點。 |
| 215 | 張九卿 | `high` | `S-GPOST`、`S-WPZH` | 10／1 | 1／6 | 通過；來源分歧與缺口均保留。 |
| 219 | 李桂林 | `medium` | `S-GPOST`、`S-WPZH`、`S-WPDIS` | 10／1 | 2／7 | 通過；第三來源可逐欄追溯。 |
| 228 | 梁忠甲 | `high` | `S-GPOST`、`S-WPZH` | 9／3 | 2／6 | 通過；同名與時序問題未被靜默裁決。 |
| 187 | 毛思義 | `medium` | `S-WPZH`；`S-GPOST` 為零命中紀錄 | 1／2 | 1／6 | 通過；零命中明確與正面史實分開。 |
| 188 | 武衍周 | `medium` | `S-GPOST`、`S-WPZH` | 6／0 | 1／6 | 通過；事件為空不補造背景敘述。 |
| 189 | 王振 | `low` | `S-GPOST`、`S-WPZH`、`S-WPJPMA` | 3／0 | 2／6 | 通過；多來源仍不足以安全接合，維持 `low`。 |
| 205 | 袁振清 | `medium` | `S-GPOST`、`S-WPZH` | 4／4 | 1／6 | 通過；事件與任職材料分欄記錄。 |
| 264 | 范熙績 | `high` | `S-GPOST`、`S-WPZH`、`S-WPJPMA` | 14／3 | 2／8 | 通過；高材料量仍保留衝突與取捨。 |

表中的任職／事件、衝突／缺口數均直接由 `facts-batch06.json` 的
`posts`／`events`、`conflicts`／`gaps` 陣列計得；不是以小傳篇幅推測。

## 3. 可重現檢查

在專案既有 Docker Python 工具鏈執行：

```sh
./tools/py.sh tools/bio_skeleton.py batch06 --audit
./tools/py.sh tools/bio_gate.py batch06
```

結果如下：

- 隔離視圖扣住 `note: 131`、`says: 34`，共 165 處自由文字；成文流程不會直接看到
  舊 `note` 或來源原話。
- `bio_gate.py batch06`：10 篇、最大 319.5 格（上限 340）、正文與顯示欄位不可畫字皆 0、
  `id`／`name_ingame` 全符、禁用詞 0。
- 額外的 JSON 結構稽核：欄位與衝突陳述的 `src` 懸空代號 0；上述兩筆例外只涉及
  零命中查詢的紀錄位置，並未把 `S-GPOST` 當欄位級事實。

## 4. 未涵蓋事項

1. 這十筆的舊 `note` 仍是 legacy schema；規格 §23.4 明定舊批不回頭改，避免在沒有
   原始來源重讀的情況下錯分成 `note_fact` 或 `note_source`。
2. 本稽核沒有連線重抓 gpost 或 MediaWiki，故不把來源可及性、內容正確性或最新頁面狀態
   誤報成已驗證；零命中也只是當時已重跑的查證紀錄，不能外推成網站永久無資料。
3. 完成本批後，舊 `partial` 母體尚有 36 筆未經同類稽核。這是審核待辦，不是 36 位
   未立傳人物。

