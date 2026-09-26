# SPEC-19：`.DT1` runtime 尾端快照同步 M1

狀態：**READY**（2026-08-10）

本規格只處理 `.DT1` 尾端已由 IDA 讀寫端證實的位移，以及 block 10
`word_70026` 的非破壞性 session／autosave 接線。它不宣稱已解出劇本生成時機、
勢力覆滅／繼任的完整原版等價。

## 1. 契約

1. 信用度是 offset `14610..14619` 的 10 × `u8`。
2. 外債是 offset `14620..14659` 的 10 × little-endian `u32`。
3. `word_70026` 是 offset `14662..14681` 的 10 × little-endian `u16`。
4. `14606..14609`、`14660..14661`、`14682` 是目前未映射的單 byte 欄位；任何
   通用寫回不得改動。
5. 解析／寫回以原始 bytes 為基底；沒有 runtime 快照時不可自行建立 block 10 名冊。

## 2. 實作邊界

- `internal/game.ParseDiplomacyLedger`／`WriteDiplomacyLedger` 只改帳本兩段。
- `internal/game.ParseMajorPowerLeaders`／`WriteMajorPowerLeaders` 只改 block 10。
- `internal/game.WriteFactionOfGeneralLeaders` 只改區塊 7 中目前勢力領袖的已證實
  反查格；殘留格與已覆滅舊格維持原始 bytes。
- `cmd/dsds.buildSession` 載入 block 10，`autosave` 以同一快照寫回；其他未解欄位
  維持原始值。
- 勢力表（24 槽）與 block 10（10 槽）不得互相替代。

## 3. 驗收

- `internal/game` 測試以 `SAVE(1)`／`SAVE(2)` 驗證信用度 100、外債 0、block 10
  首槽 58／166。
- writer 測試逐 byte 驗證未授權區域不變，並 round-trip 讀回相同值。
- `cmd/dsds` Xvfb session／autosave 測試通過。
- Docker `go test`、`git diff --check`、no-cgo 掃描與資產拒絕掃描通過；一次性
  容器清除，不能留下專案相關執行中容器。
