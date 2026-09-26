# SPEC-18：第一期外援數值代碼 M1

狀態：**READY**  
日期：2026-08-10  
依據：`docs/re/38-aid-donor-code-mapping.md`

## 目標

把 `sub_21D1D` 已證實的第一期援助代碼分支接到規則層與正常玩家入口，保留原版
勢力反查值、同一顆 `Random(10)` 與第二／三期的 99 覆寫；不把未證實的數值 ID
翻譯成國家名稱。

## 規則契約

1. `AidDonorCode(stage, factionCode, roll)` 的 `factionCode` 是 1-based，對應
   `.DT1` 區塊 7；未提供資料使用 0，走原版 default 分支。
2. stage 2／3 一律回傳 99；stage 1 依 `docs/re/38` 的五組分支回傳原版數值。
3. `RequestAidForFaction` 在第一次援助判定取得 roll 後選代碼，不能另擲一顆骰或
   改變四項資源亂數順序。
4. 99／142、100、其餘代碼的除數仍由 `AidDivisor` 唯一實作；本規格不命名國家。

## 玩家入口契約

- `executeDiplomacyAid` 從已載入的勢力表把目前司令映射成 1-based 槽位；新局沒有
  `.DT1` 後半時傳 0，不因缺資料把援助國硬編成某一國。
- 原典／現代白話只影響訊息文字，不影響代碼、除數、亂數或存檔。

## 非目標

- 不處理 `.DT1` 區塊 10 的生成／寫回時機。
- 不把 99、100、101、141、142、146 映射成美國、英國、俄國、法國或日本。
- 不改援助拒絕／特殊核准分支、指令成本或四項資源公式；那些分別見
  `docs/re/37-aid-branch-correction.md` 與 SPEC-17。

## 驗收

- 單元測試逐項覆蓋勢力代碼 10、1、5、4、3／2、default，以及 stage 2／3 覆寫。
- 固定種子下 `RequestAidForFaction` 與 `AidDonorCode` 使用相同 roll，核准時
  `Donor`／`Divisor` 對應正確。
- Docker 內通過規則／語系／UI 定向測試、no-cgo source scan、deny scan、
  `git diff --check`，且不留下專案相關容器。

