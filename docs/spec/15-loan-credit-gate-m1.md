# 規格 15：貸款信用度零 gate M1

狀態：**READY**  
日期：2026-08-10

本規格把 `WAR.EXE` 貸款流程中已由 IDA 閉合的「信用度為零」早期 gate 接到 Go 規則、
外交帳本與玩家 UI。它不新增未證實的「信用度不足以支付本次額度」門檻。

## 1. 已證實行為

`WAR.EXE` SHA-256 `11dbfcf24686ab7765f788b38514cefd2039d0f60b6bd517d89fb5a84c068015`，
IDA Pro 9.4.0.260610、線性位址 `sub_2164A` 起點 `2164Ah`：

```text
信用度表 [di-4225h] == 0
    → 顯示「無法貸款」，返回主迴圈
    → 不進額度輸入、不擲貸款 Random、不改狀態
信用度 != 0
    → 才進入既有 amount/500 與 Random(10) 核貸公式
```

`[di-4225h]` 的 1-based 槽位與 `.DT1` 區塊 9 已由 SPEC-06 確認；本規格只補上呼叫前
的零值 gate。

## 2. Go／Ebiten 接線

- `internal/game.LoanResult.CreditBlocked` 是唯一規則結果標記。
- `AIWorld.RequestLoan` 在 `rng.Int` 前檢查 `credit == 0`，不修改省份。
- `DiplomacyLedger.RequestLoan` 原子保留零值 gate，不增加外債或扣信用度。
- `cmd/dsds` 在外交第 1 項進入貸款頁前檢查 ledger slot；提交時再次檢查規則結果，
  滑鼠／觸控與鍵盤共用同一個外交選單 action。
- 新語意鍵 `diplomacy.loan.unavailable` 與 `diplomacy.loan.refused` 同時提供原典與
  現代白話；隨機拒絕仍顯示「各國均拒絕提供貸款」，保留兩種失敗的可辨識性。

## 3. 驗收

1. 信用度零不消耗 `Rand`、黃金、外債、信用度或指令。
2. 正信用度的既有小額必過、邊界機率與大額必拒測試保持通過。
3. `wording.json` 兩種模式均有 gate 用語，缺鍵時 `LoadWording` fail-closed。
4. Docker 內執行規則／語系／`cmd/dsds` Xvfb 測試、no-cgo、deny scan、diff check。

## 4. 非目標

- 隨機拒絕的指令成本另由 `docs/spec/16-loan-command-cost-m1.md` 管理；本規格只管
  信用度零的早期 gate。
- `ds:ACE7h` 顯示準備表的完整語意。
- Android 封裝、三平台正式包與完整外交原版等價。
