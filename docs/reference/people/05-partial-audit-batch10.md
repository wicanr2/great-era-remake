# `partial` 池來源與骨架稽核：batch10（最後 6 筆，來自 sample）

> 執行日：2026-08-02。輸入：`facts-sample.json`、`bios-sample.md`、
> `people.json` 與 `docs/design/21-biography-authoring-spec.md` §0、§2、§3、§6、§23.2–§23.4。
>
> `batch10` 是本次稽核的收束編號，並非既有的 `facts-batch10.json`。
> 最後六筆位於規格試寫用的 `facts-sample.json`；故本批的品質閘目標是 `sample`。
> 範圍仍只做**結構、來源代號、成文存在性及隔離視圖**的機械稽核；不重新連線抓取
> 史料、不重判史實，也不改寫既有小傳。

## 1. 結論

最後六筆舊 `partial` 均已完成成文、槽位與可畫性檢查：

- 從 `people.json` 的 96 筆舊 `partial` 扣除 batch01–batch09 已稽核的 90 筆後，
  剩餘**恰為 6 筆**：#93、#353、#356、#367、#368、#383；本批已逐一稽核。
- 6/6 的 `id` 與遊戲姓名均通過 `bio_gate.py sample`；6/6 均有一篇成文。
- 欄位與衝突陳述實際使用的來源代號，以及每筆 `verified_by`，均在 sample 的 `sources`
  表宣告；懸空代號 0，且沒有未錨定的 `verified_by`。
- 完整骨架的信心度為 `medium` 4、`low` 2；舊 `partial` 是歷史分類，不能覆蓋
  這一輪逐欄來源鏈的結果。

**完成後再次以同一集合計數：舊 `partial` 待稽核為 0。** 這只代表「96 筆已完成
本類機械稽核」，不代表來源內容已重新上網覆核或 schema 遷移完成。

## 2. 逐筆結果

| id | 人物 | 目前信心度 | 骨架中的已登錄來源 | 任職／事件 | 衝突／缺口 | 結論 |
|---:|---|---|---|---:|---|---|
| 93 | 王獻臣 | `low` | `S-GPOST`、`S-WPZH` | 8／2 | 1／2 | 通過；同名接合風險維持 `low`。 |
| 353 | 葉肇 | `medium` | `S-GPOST`、`S-WPZH`、`S-1937NJ`、`S-GDLQW` | 7／6 | 1／1 | 通過；多來源仍保留未解分歧。 |
| 356 | 劉膺古 | `medium` | `S-GPOST`、`S-WPZH`、`S-CN-ENCY` | 11／3 | 2／1 | 通過；第三來源已逐欄錨定。 |
| 367 | 許鴻林 | `low` | `S-GPOST`、`S-CN-ENCY` | 3／0 | 0／3 | 通過；資料稀少未被補造成完整經歷。 |
| 368 | 于鎮河 | `medium` | `S-GPOST`、`S-WPZH`、`S-CN-ENCY`、`S-TYDAO` | 4／5 | 1／2 | 通過；來源分歧與身分資料可回查。 |
| 383 | 張淦 | `medium` | `S-GPOST`、`S-WPZH`、`S-CN-ENCY` | 5／4 | 1／3 | 通過；來源鏈與成文一致。 |

表中的任職／事件、衝突／缺口數均直接由 `facts-sample.json` 的
`posts`／`events`、`conflicts`／`gaps` 陣列計得；不是以小傳篇幅推測。

## 3. 隔離視圖與可重現檢查

在專案既有 Docker Python 工具鏈執行：

```sh
./tools/py.sh tools/bio_skeleton.py sample --audit
./tools/py.sh tools/bio_gate.py sample
```

結果如下：

- `sample` 的完整隔離視圖扣住 `note: 26`、`says: 27`，共 53 處自由文字。
  sample 同時含 4 筆不屬於舊 `partial` 的規格試寫；本批六人的人物資料本身包含
  `note: 12`、`says: 12`。無論哪一層，成文視圖都不會直接看到舊 `note` 或來源原話。
- `bio_gate.py sample`：10 篇、最大 330 格（上限 340）、正文與顯示欄位不可畫字皆 0、
  `id`／`name_ingame` 全符、禁用詞無阻擋項。
- 額外的 JSON 結構稽核：本批六人的欄位與衝突陳述 `src`、`verified_by` 懸空代號均為 0，
  且 `verified_by` 全數可在對應人物的欄位級來源引用中找到。

## 4. 跨批次結算與未涵蓋事項

1. `bio_gate.py --cross` 仍報 #16 孫連仲、#21 商震各在 batchC2／batchC3 重複。
   這兩組是既有跨批次成文重複，**仍需人工合併**；本批沒有改動它們。
2. 這六筆及前九批的舊 `note` 都仍是 legacy schema；規格 §23.4 明定不回頭改，避免在
   沒有原始來源重讀的情況下錯分成 `note_fact` 或 `note_source`。
3. 本稽核沒有連線重抓 gpost 或 MediaWiki，故不把來源可及性、內容正確性或最新頁面狀態
   誤報成已驗證。`partial` 稽核待辦歸零，不會消除 124 筆 Tier A 未量測與個別來源複核的工作。

