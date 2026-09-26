# 反組譯證據：貸款隨機拒絕仍完成指令

日期：2026-08-10  
狀態：**confirmed**（僅限 `WAR.EXE` 政略主迴圈與 `sub_2296C` 外交呼叫端）

## 證據身分

- 輸入：`WAR.EXE`，SHA-256
  `11dbfcf24686ab7765f788b38514cefd2039d0f60b6bd517d89fb5a84c068015`。
- 工具：IDA Pro 9.4.0.260610；所有位址都是 **IDA 線性位址**。
- 主迴圈：`sub_10193`；外交選單：`sub_2296C`；貸款：`sub_2164A`。

## 旗標資料流

1. `sub_10193` 在每次政略指令 dispatch 前把 `byte_6FE81` 清為 0。
2. `sub_2296C` 呼叫 `sub_2164A` 後，在 `loc_22B5C` 檢查同一 byte；零值回到外交
   選單，非零才返回主迴圈。
3. `sub_10193` 的 `loc_103DF` 在 `byte_6FE81 != 0` 時對目前省份命令／狀態欄位
   執行 `dec`，也就是完成一次指令。
4. `sub_2164A` 的 `loc_21967` 在使用者提交正額度後、呼叫 `Random(10)` 前執行：

   ```asm
   mov     byte_6FE80, 1
   mov     byte_6FE81, 1
   mov     ax, 0Ah
   call    @Random$q4Word
   ```

   後續 `loc_21C42` 的「各國均拒絕提供貸款」分支沒有清除此旗標，因此隨機拒絕
   仍會被主迴圈扣一個指令。
5. 信用度零 gate（`sub_2164A` 開頭 `cmp byte ptr [di-4225h], 0`）在
   `loc_21967` 之前直接跳 `loc_21D11`，沒有立 `byte_6FE81`；因此無法貸款不扣指令。

## Remake 接線

- `LoanResult.CommandCompleted` 表示額度已送出、應由 UI 消耗指令；
  `CreditBlocked` 表示早期 gate，兩者不可混用。
- `cmd/dsds` 對 `CommandCompleted && !Approved` 的隨機拒絕呼叫
  `CommandBudget.Spend`；信用度零分支直接返回、不擲骰、不扣指令。
- `diplomacy.loan.refused` 的現代白話明示「已消耗一個指令」；原典保留原版
  「各國均拒絕提供貸款」文字。這是顯示模式差異，不改規則。

## 驗收界線

此證據只閉合「隨機拒絕的指令成本」；不推論其他外交失敗分支的成本，也不代表
正常 DOSBox 完整外交流程、Android 封裝或三平台發行已完成。
