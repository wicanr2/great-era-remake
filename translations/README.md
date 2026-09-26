# 語系資料

`AGENTS.md` §6 的兩條要求：

> 所有玩家看得到的文字都在語系資料檔，不進 Go 原始碼。
> 原版的字模索引機制到 remake 就結束，重寫版用真正的字串表 + 完整字型。

所以這個目錄放的是**執行期載入**的語系資料，換一個目錄就換一種語言，
不必重編（`internal/i18n`）。

```
translations/
  glossary.md              譯名表，唯一真相（品質閘要檢查譯名漂移）
  zh-Hant/glyphtext.json   繁中母本 ← 從原版字模還原出來的，不是重寫的
  zh-Hant/wording.json     原典／現代白話的穩定語意鍵
  zh-Hant/narrative.json   17 張 NEWSDATA 來源圖像的 caption／fallback catalog
  zh-Hant/people.json      可發行人物資料 ← tools/gen_people.py
  zh-Hant/people-authored.json  新增 61 篇自撰傳記 overlay ← tools/gen_authored_bios.py --overlay
  zh-Hant/people-unmapped.md  人物文字正規化帳本
  en/                  英文 UI／省名與 387 篇 machine-draft 人物 overlay
  ja/                  日文漢字優先 UI／省名與 387 篇 machine-draft 人物 overlay
  shared/roster-slots.json 三期將領槽位 → 人物 id（不隨語系改變）
```

## `glyphtext.json` 的來源與格式

繁中母本**不是翻譯，是還原**：51 個 `.15` 字模檔逐格拿去跟倚天字庫
做 byte-exact 比對（`docs/formats/01-glyph-text.md`）。
`tools/gen_locale.py` 把反查結果轉成這份表。

```json
{
  "language": "zh-Hant",
  "files": {
    "3.15": {
      "slot_width": 3,
      "entries": [
        {"text": "嫩江省", "raw": "嫩江省"},
        {"text": "西藏",   "raw": "西　藏"}
      ]
    }
  }
}
```

| 欄位 | 用途 |
|---|---|
| `slot_width` | 原版的槽寬（每個詞條佔幾個字模格）|
| `text` | 去掉排版空白 —— **remake 用這個** |
| `raw` | 保留原版排版空白（全形空格）—— 還原原版畫面用 |

`raw` 存在的理由：原版的「西藏」在畫面上是**三格分散排版**「西　藏」，
因為那個表的槽寬是 3。remake 用比例字排版之後那個空白沒有意義，
但保存專案要留著它。

索引一律 **1-based**，與原版的字模索引一致（`docs/re/24`）。

## 可信度

- 51 檔 6,174 個字模，倚天命中 4,799、空白填充 1,374、**例外 1**
  （`1.15` 的逗號，倚天「，」左移 3px 下移 2 列的版本，人工判讀後補上）。
- `tools/dup_glyph.py` 證實倚天字庫**沒有任何兩個碼點共用同一字形**，
  所以 byte-exact 反查是單射的，不會取到錯的字
  （`docs/formats/01-glyph-text.md` §5b）。
- 槽寬照 `docs/formats/01-glyph-text.md` §4 的表，分兩級證據：
  46 檔 `auto`（排版模式推出，符合率 100%）、5 檔 `content`（語意判定，較弱）。
  `tools/gen_locale.py` 會**驗整除**——槽寬錯會讓整份詞條錯位而不報錯。

## 產生英文版／日文版

目前的可重現語系包由 `tools/gen_locale_packs.py` 產生：

```text
tools/py.sh tools/gen_locale_packs.py
```

它會驗證 154 個穩定語意鍵與 `%d` 格式參數，產生 `translations/en`、
`translations/ja` 的 wording、39 省名稱與 417 人物 base 資料；base 保留繁中
來源正文，翻譯由同目錄的 `people-biography-overlay.json` 另行套用。執行時以 `-locale`
選擇語系，例如 `-locale translations/en`；非繁中語系會自動走 semantic wording
路徑，不再偷偷繪製繁中原版命令字模。

人物姓名保留遊戲的歷史寫法。本輪接受固定詞表產生的英／日 machine-draft：每包
有 387 筆可追溯 overlay，`bio_language` 與 `bio_status=machine-draft` 會由
runtime 套用；未命中的專名保留原文並明示需要人審。產生器與 runtime 會驗證 387
筆 ID、姓名、來源／譯文 SHA-256、狀態及 reviewer metadata；失敗即整包拒絕。正式
發行仍需 `--release` gate 的 human-reviewed，日文假名字型與比例版面亦仍待後續。

`narrative.json` 是另一個 17 筆、以 `NEWSDATA.DAT#<id>` 為來源索引的唯讀 catalog。
目前只有 #0／#9 有可追溯的繁中／英／日 caption；其餘 15 筆以
`status=source-image` 保留原圖，畫廊顯示 fallback「本文未解出」。這不是把未知新聞
翻成機器句子，也不會改變事件規則；完整 schema 與玩家路徑見
[`docs/spec/31-narrative-gallery-m1.md`](../docs/spec/31-narrative-gallery-m1.md)。

繁中自傳另需使用者提供的倚天完整字庫：

```text
tools/go.sh run ./cmd/dsds -eten workplace/eten
```

字庫檔不進版控；`tools/deny_scan.sh` 會拒絕 `STDFONT`／`SPCFONT` 等檔名。

研究批次的自撰小傳可先做不寫回 base `people.json` 的完整性檢查：

```text
tools/py.sh tools/gen_authored_bios.py --check
```

目前會證實 417 筆骨架、387 篇正文、30 筆 `unknown` 與 61 篇待整合。要產生目前
執行期缺少的 61 篇，可使用：

```text
tools/py.sh tools/gen_authored_bios.py --overlay \
  --output translations/zh-Hant/people-authored.json
```

`internal/i18n.LoadPeople` 只把 overlay 填入 base 空白欄位；既有 326 篇不會被覆蓋。
完整 387 篇回寫 `people.json` 的 DESIGN-22 仍為 DRAFT，overlay 產物則由
`docs/spec/11-authored-biography-overlay-p2c.md`（READY）管理。

⚠️ 英文版不是把中文換掉就好：640×350 的版面是照全形字排的，
換成比例字後字寬、行高、對話框都要重算（`AGENTS.md` §6）。
**排版層必須先抽離**，否則英文一定溢出。`internal/ui/textlayout` 已完成第一階段的
繁中半格量測、標點禁則與分頁；比例向量字、英文斷詞與畫面接線仍未完成。

⚠️ 譯名一律走 `glossary.md`。人名有 22 筆遊戲寫法與通行寫法不同
（閰錫山／閻錫山…），**繁中母本保留原版寫法**，英日文版才用通行寫法
（`docs/reference/people/02-status.md` §2）。
