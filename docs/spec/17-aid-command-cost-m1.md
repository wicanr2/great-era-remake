# SPEC-17：外援結果的指令成本 M1

狀態：**READY**  
日期：2026-08-10  
依據：`docs/re/36-aid-command-cost.md`

## 目標

把原版 `sub_21D1D` 的「進入外援亂數後即完成指令」接入 remake 的規則層與玩家
呼叫端，讓核准、隨機拒絕、日期／司令特殊核准三條結果都消耗一個本月指令。

## 規則契約

1. `AIWorld.RequestAid` 通過省份查找後，第一次 `Random(10)` 前將
   `AidResult.CommandCompleted` 設為 `true`。
2. `Roll <= AidRefusalMax` 且未命中特殊日期／司令分支時拒絕，不入帳資源，但保留
   `CommandCompleted`。
3. 司令 166、民國 17 年 2–6 月命中 `SpecialOverride`，跳過拒絕文字並照核准路徑
   入帳；仍保留 `CommandCompleted`。
4. 一般核准是 `Roll >= AidApprovalMin`；核准結果保留既有資源亂數、黃金 6000
   夾值與除數規則。
5. 省份查找等前置錯誤直接回傳 error，不產生可提交的結果；不得以這個錯誤路徑
   消耗指令。

## UI／提交契約

- `cmd/dsds.executeDiplomacyAid` 只要 `CommandCompleted` 為真，就在結果文字前提交
  `CommandBudget.Spend`。
- 提交失敗時回復省份與亂數種子，並回到政略畫面；不得留下半次資源或亂數副作用。
- 原典拒絕文字保留「各國均拒絕提供援助」；現代白話需明示「已消耗一個指令」。

## 非目標

- 不在本規格決定第一期援助國代碼與國家名稱對應；數值代碼另見 SPEC-18，
  國家名稱仍未證實。
- 不改援助資源公式、日期特殊核准條件、外交帳本格式或 `.DT1` 未解 bytes。

## 驗收

- 固定種子可觀察隨機拒絕與 `CommandCompleted=true`，且省份資源不變。
- 特殊日期／司令分支的固定種子有 `SpecialOverride=true`、核准入帳與
  `CommandCompleted=true`。
- `go test`、no-cgo source scan、deny scan 與 Xvfb `cmd/dsds` 測試在 Docker 內通過。
