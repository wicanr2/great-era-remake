# 反組譯證據：貸款信用度為零的早期 gate

日期：2026-08-10  
狀態：**confirmed**（僅限 `WAR.EXE` 的 `sub_2164A`）

## 證據身分

- 輸入：`WAR.EXE`，SHA-256
  `11dbfcf24686ab7765f788b38514cefd2039d0f60b6bd517d89fb5a84c068015`。
- 工具：IDA Pro 9.4.0.260610；下列是 **IDA 線性位址**，不是檔案偏移。
- 函式：`sub_2164A`（IDA 線性起點 `2164Ah`）。原始運算元與符號保留，沒有改名覆蓋。
- 欄位：`[di-4225h]` 是已確認的外交信用度表；`tools/addr.py -4225h` 對應
  `ds:BDDBh`、1-based 十勢力槽位。

## 指令序列

`sub_2164A` 在建立貸款輸入畫面前先執行：

```asm
mov     al, ss:[di-6]
xor     ah, ah
mov     di, ax
cmp     byte ptr [di-4225h], 0
jnz     short loc_21718
; 顯示「無法貸款」的文字與提示音
jmp     loc_21D11
```

因此已證實的 gate 是 **信用度等於 0**，不是「信用度小於本次額度單位」；只有非零
才會進入輸入額度、`Random(10)` 與 `roll + amount/500 <= 12` 的核貸分支。信用度為零
時不擲貸款亂數、不改省份黃金、不改外債、不扣指令。

## Remake 接線

- `internal/game.LoanResult.CreditBlocked` 保留這個早期分支，並在 `RequestLoan` 擲骰前
  回傳；因此亂數種子保持不變。
- `cmd/dsds` 在外交第 1 項進入額度頁前先檢查同一個 ledger slot；若執行期狀態在頁面
  期間變成零，提交時仍由 `CreditBlocked` 再擋一次。
- `diplomacy.loan.unavailable` 供原典／現代白話兩種用語使用；不把信用 gate 與
  「各國均拒絕提供貸款」的隨機失敗混成同一條訊息。

## 驗收界線

- 規則測試驗證信用度零時：`CreditBlocked=true`、`Approved=false`、黃金／帳本不變、
  亂數種子不變。
- 不宣稱已解出拒絕貸款的指令成本；目前仍採既有 fail-closed remake 行為。
- 不把 `ds:ACE7h` 的司令／分期顯示表誤當成信用 gate；該段只影響 `var_6` 顯示準備，
  與核貸判定分開。
