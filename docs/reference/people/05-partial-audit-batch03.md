# `partial` 池來源與骨架稽核：batch03（10 筆）

> 執行日：2026-08-02。輸入：`facts-batch03.json`、`bios-batch03.md`、
> `people.json` 與 `docs/design/21-biography-authoring-spec.md` §0、§2、§3、§6、§23.2–§23.4。
>
> 範圍：本批十人都在舊 `people.json` 標作 `partial`。本檔完成的是
> **結構、來源代號、成文存在性及隔離視圖**的機械稽核；不重新連線抓取史料、
> 不重判史實，也不改寫既有小傳。

## 1. 結論

batch03 的十筆均已完成成文、槽位與可畫性檢查：

- 10/10 的 `id` 與遊戲姓名均通過 `bio_gate.py`；10/10 均有一篇成文。
- 欄位內實際使用的來源代號全數在本批 `sources` 表宣告，故**懸空來源 0**。
- 完整骨架的信心度為 `high` 4、`medium` 5、`low` 1；舊 `partial` 是歷史分類，
  不能覆蓋這一輪逐欄來源鏈的結果。
- 106 個舊 `note` 與 36 段 `conflicts[].statements[].says` 都會被
  `bio_skeleton.py` 的成文視圖扣住，維持著作權隔離。

### 本批唯一的來源追溯例外

#106 張國威的 `verified_by` 列有 `S-GPOST`，但該筆的 `life`、`education`、
`posts`、`events`、`ending`、`conflicts` 與 `gaps` 沒有任何欄位實際引用這個代號；
欄位材料只有 `S-WPZH`。`S-GPOST` 本身已在批次來源表宣告，因此這不是懸空代號，
但它是**一筆未落在欄位上的驗證宣稱**，不能當作 #106 的欄位級 Tier A 證據。
本稽核依限制不改既有骨架，僅將此例外留作後續原始來源重讀時的修正起點。

## 2. 逐筆結果

| id | 人物 | 目前信心度 | 骨架中的已登錄來源 | 任職／事件 | 衝突／缺口 | 結論 |
|---:|---|---|---|---:|---|---|
| 51 | 王鴻恩 | `medium` | `S-GPOST`、`S-WPZH` | 15／3 | 2／4 | 通過；時序衝突留在骨架。 |
| 56 | 孫長勝 | `medium` | `S-GPOST`、`S-WPZH` | 10／5 | 2／4 | 通過；期別外材料未被誤當期內事實。 |
| 60 | 楊源濬 | `medium` | `S-GPOST`、`S-WPZH` | 6／5 | 2／4 | 通過；異寫與來源分歧皆保留。 |
| 62 | 孫建業 | `high` | `S-GPOST`、`S-WPZH` | 8／2 | 1／4 | 通過；交叉來源均有欄位錨點。 |
| 69 | 余蔭森 | `high` | `S-GPOST`、`S-WPZH` | 9／3 | 2／4 | 通過；未解處不以推定填補。 |
| 70 | 馬濟 | `high` | `S-GPOST`、`S-WPZH` | 10／6 | 2／4 | 通過；身分資料與職務序列分欄追溯。 |
| 81 | 李鴻翥 | `low` | `S-GPOST`、`S-WPZH` | 7／2 | 1／4 | 通過；同名／接合風險維持 `low`。 |
| 82 | 袁家驥 | `medium` | `S-GPOST`、`S-WPZH` | 7／2 | 1／4 | 通過；資料空缺未被補造。 |
| 106 | 張國威 | `medium` | 欄位為 `S-WPZH`；`verified_by` 另列 `S-GPOST` | 5／4 | 2／5 | 通過但有未錨定的 `verified_by`，見上節。 |
| 108 | 謝鴻勳 | `high` | `S-GPOST`、`S-WPZH` | 8／6 | 2／4 | 通過；衝突與缺口均可回查。 |

表中的任職／事件、衝突／缺口數均直接由 `facts-batch03.json` 的
`posts`／`events`、`conflicts`／`gaps` 陣列計得；不是以小傳篇幅推測。

## 3. 可重現檢查

在專案既有 Docker Python 工具鏈執行：

```sh
./tools/py.sh tools/bio_skeleton.py batch03 --audit
./tools/py.sh tools/bio_gate.py batch03
```

結果如下：

- 隔離視圖扣住 `note: 106`、`says: 36`，共 142 處自由文字；成文流程不會直接看到
  舊 `note` 或來源原話。
- `bio_gate.py batch03`：10 篇、最大 334.5 格（上限 340）、正文與顯示欄位不可畫字皆 0、
  `id`／`name_ingame` 全符、禁用詞 0。
- 額外的 JSON 結構稽核：欄位 `src` 的懸空代號 0；`verified_by` 全都已在來源表宣告。
  但 #106 的 `S-GPOST` 未出現在任何該筆欄位 `src`，故另列為未錨定驗證宣稱。

## 4. 未涵蓋事項

1. 這十筆的舊 `note` 仍是 legacy schema；規格 §23.4 明定舊批不回頭改，避免在沒有
   原始來源重讀的情況下錯分成 `note_fact` 或 `note_source`。
2. 本稽核沒有連線重抓 gpost 或 MediaWiki，故不把來源可及性、內容正確性或最新頁面狀態
   誤報成已驗證；特別是 #106 的 `S-GPOST` 是否應保留，需在單一 gpost 作業輪次重讀。
3. 完成本批後，舊 `partial` 母體尚有 66 筆未經同類稽核。這是審核待辦，不是 66 位
   未立傳人物。

