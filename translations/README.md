# 語系資料

`CLAUDE.md` §6 的兩條要求：

> 所有玩家看得到的文字都在語系資料檔，不進 Go 原始碼。
> 原版的字模索引機制到 remake 就結束，重寫版用真正的字串表 + 完整字型。

所以這個目錄放的是**執行期載入**的語系資料，換一個目錄就換一種語言，
不必重編（`internal/i18n`）。

```
translations/
  glossary.md              譯名表，唯一真相（品質閘要檢查譯名漂移）
  zh-Hant/glyphtext.json   繁中母本 ← 從原版字模還原出來的，不是重寫的
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

## 做英文版／日文版

1. 複製 `zh-Hant/glyphtext.json` 成 `en/glyphtext.json`，翻譯 `text` 欄位。
2. `raw` 欄位對非中文語系沒有意義，可以直接等於 `text`。
3. 跑起來：`tools/go.sh run ./cmd/dsds -locale translations/en`

⚠️ 英文版不是把中文換掉就好：640×350 的版面是照全形字排的，
換成比例字後字寬、行高、對話框都要重算（`CLAUDE.md` §6）。
**排版層必須先抽離**，否則英文一定溢出。那一層還沒做。

⚠️ 譯名一律走 `glossary.md`。人名有 22 筆遊戲寫法與通行寫法不同
（閰錫山／閻錫山…），**繁中母本保留原版寫法**，英日文版才用通行寫法
（`docs/reference/people/02-status.md` §2）。
