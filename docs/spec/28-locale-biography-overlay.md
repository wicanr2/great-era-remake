# SPEC-28：英／日人物自傳 overlay 與發行流程

狀態：**READY**（machine-draft 預覽；正式發行仍需 human-reviewed）  
日期：2026-08-10

本規格把「研究正文」與「翻譯正文」分開。它不解除 DESIGN-22／SPEC-11 對繁中
權威資料的保護；本輪接受 deterministic machine-draft 先行接入預覽，但不把它
改標成正式 `translated`。

## 1. 目前狀態與決策閘

- 繁中研究批次：417 個人物骨架、387 篇成文、30 筆 unknown；來源檔與正文雜湊
  已由 bio_gate 與 gen_authored_bios 驗證。
- 英／日語系：417 筆人物資料可載入，並各有 387 筆可追溯的
  `people-biography-overlay.json`；目前狀態為 `machine-draft`／`bio_language=en|ja`。
- 產生器使用固定片語詞表、日期規則與人物事實欄位，不呼叫外部翻譯服務；未命中的
  專名保留原文並在正文末尾明示，供逐篇人審修訂。
- `machine-draft` 只代表預覽鏈已接通，不代表 387 篇 human-reviewed；Modern `GEMF`
  atlas 已覆蓋英／日正文，但日文禁則、長文版面與正式逐篇審稿仍未完成。

## 2. 來源與產物拓樸

    docs/reference/people/facts-*.json + bios-*.md
            │ 研究稿／來源雜湊（只讀）
            ▼
    locale biography draft（en 或 ja，逐篇可追溯）
            │ review_status：machine-draft → human-reviewed
            ▼
    translations/en/people-biography-overlay.json
    translations/ja/people-biography-overlay.json
            │ schema／來源／版面／字型 gates
            ▼
    translations/{en,ja}/people.json（由產生器重生，不手改）

overlay 只覆蓋同一 ID 的 biography 欄位與語系狀態，不覆蓋姓名、派系、出生／
死亡年或研究信心度；那些欄位仍由 locale base people.json 管理。

## 3. overlay schema（草案）

    {
      "schema_version": "1",
      "language": "en",
      "overlay_scope": "translated-biographies",
      "translation_status": "machine-draft",
      "generated_from": [
        {
          "facts": "docs/reference/people/facts-batchU1.json",
          "facts_sha256": "64 位十六進位雜湊",
          "bios": "docs/reference/people/bios-batchU1.md",
          "bios_sha256": "64 位十六進位雜湊"
        }
      ],
      "people": [
        {
          "id": 13,
          "name_ingame": "人物原版姓名",
          "bio": "翻譯正文",
          "bio_language": "en",
          "bio_status": "machine-draft",
          "source_bio_sha256": "繁中正文雜湊",
          "translated_bio_sha256": "翻譯正文雜湊",
          "review": {
            "status": "unreviewed",
            "reviewer": "",
            "reviewed_at": ""
          }
        }
      ]
    }

日文將 language／bio_language 改為 ja；假名與日文標點不可因 schema 合法就視為
字型完成。所有 ID 必須遞增且只出現一次；人物數應為 387，#274「無省長」與
30 筆 unknown 不得進 overlay。每筆 name_ingame 必須與 base 完全相同，source
與 translated 雜湊必須可重算。

## 4. 正式流程與 fail-closed gates

1. **研究來源 gate：** 先通過 bio_gate、gen_authored_bios、來源 SHA-256 與
   387／30／#274 數量檢查。
2. **翻譯稿 gate：** 每篇保留 source_bio_sha256；缺稿、重複 ID、姓名漂移、
   空正文或格式參數錯誤即整份失敗。
3. **審稿 gate：** machine-draft 只能進預覽，不可進正式發行；正式包需要
   human-reviewed，並留下 reviewer／reviewed_at／勘誤紀錄。
4. **語系 overlay gate：** 由產生器以暫存檔合併，先檢查 387 筆與所有雜湊，再
   原子替換 locale people.json；不得直接手改 base 或研究來源。
5. **排版／字型 gate：** 英文比例字、斷詞、長文分頁；日文假名 glyph、比例字、
   禁則與行首行尾標點；各抽驗短／中／長文與 unknown fallback。
6. **發行 gate：** `bio_status=machine-draft` 只能進預覽；正式 `translated` 只能出現在已通過人審與版面 gate 的
   row；其餘保留 source-fallback 或 machine-draft，不得偽裝。

## 5. 與 runtime 的關係

PeopleDB 會在存在 overlay 時以 fail-closed 方式套用 machine-draft／human-reviewed
正文，仍沿用同一個 `Person.Biography` 欄位供 renderer 分頁，並讓
`BiographyLanguage`／`BiographyStatus` 反映實際 row；載入器遇到 schema、
語系、ID、姓名、雜湊或數量錯誤時整份 fail-closed。

## 6. 本輪決定與後續 gate

使用者已接受 machine translation 初稿，因此建立英／日兩份 387 篇 overlay，
狀態固定為 `machine-draft`、review 狀態為 `unreviewed`。後續逐篇校訂、日文假名
字型／比例版面與正式發行仍是獨立工作；`tools/locale_bio_gate.py --release`
會在任何一筆尚未 `human-reviewed` 時 fail-closed。
