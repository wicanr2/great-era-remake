# 外援結果都消耗指令：`sub_21D1D` → `byte_6FE81`

狀態：**confirmed（原版指令完成旗標資料流）**  
日期：2026-08-10

## 輸入與工具

- 輸入：`workplace/ida/WAR.EXE.asm`
- `WAR.EXE` SHA-256：`11dbfcf24686ab7765f788b38514cefd2039d0f60b6bd517d89fb5a84c068015`
- 工具：IDA Pro `9.4.0.260610` 匯出的線性組語
- 位址空間：下列 `21xxxh` 是 IDA linear address；`byte_6FE81` 是 IDA
  符號，不能與遊戲 `ds:` 偏移混寫

## 呼叫端資料流

`sub_10193` 是政略指令主迴圈。每次進入指令前清除 `byte_6FE81`；返回後在
`loc_103DF` 依旗標是否為零決定是否遞減本月指令數。`sub_2296C` 呼叫
`sub_21D1D` 後在 `loc_22B5C` 讀同一旗標，零值留在外交選單，非零值返回主迴圈。

外援函式 `sub_21D1D` 在 `loc_21E9D` 的順序是：

```asm
call @Random             ; Random(10)，結果進入 var_14
mov  byte_6FE80, 1
mov  byte_6FE81, 1      ; 先立「指令已完成」旗標
```

後面才判斷 `var_14 <= 6` 的一般拒絕分支，以及張作霖／民國 17 年 2–6 月跳過
拒絕文字的特殊核准分支。兩條分支都沒有清除 `byte_6FE81`，因此都會回到主迴圈
並扣一個指令；一般核准也沿用同一旗標。

這與貸款信用度零 gate 不同：信用度零在 `sub_2164A` 的額度輸入與亂數前就返回，
根本不會立 `byte_6FE81`（見 `docs/re/34-loan-credit-gate.md`）。外援沒有同類的
前置拒絕；只要已通過省份查找並進入 `sub_21D1D`，第一次援助亂數前就算完成指令。

## Remake 邊界

`internal/game.AidResult.CommandCompleted` 對應上述旗標，並在第一次 `Random(10)`
前設為 `true`。`cmd/dsds.executeDiplomacyAid` 對核准、隨機拒絕與特殊核准都提交
`CommandBudget.Spend`；防禦性提交失敗會回復省份快照與亂數種子。規則結果與畫面用語
分離，不能用「未核准」推導「未消耗指令」。

本證據只閉合**指令成本**；第一期援助數值代碼另見
`docs/re/38-aid-donor-code-mapping.md`，國家名稱仍未證實。
