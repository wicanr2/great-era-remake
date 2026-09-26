# 規格 16：貸款隨機拒絕指令成本 M1

狀態：**READY**  
日期：2026-08-10

本規格把原版貸款的「提交額度後立完成旗標」接到 remake。信用度為零的早期 gate
仍由 SPEC-15 管理，兩條分支不可合併。

## 1. 已證實行為

- `sub_10193` 每次政略指令前清 `byte_6FE81`；非零返回時在 `loc_103DF` 扣命令。
- `sub_2296C` 呼叫 `sub_2164A` 後以 `byte_6FE81` 決定是否回主迴圈。
- `sub_2164A` 在 `loc_21967`、`Random(10)` 前立 `byte_6FE81=1`。
- 隨機拒絕的 `loc_21C42` 不清旗標，所以仍消耗一個指令。
- 信用度為 0 的 `loc_21718` 早期分支在立旗標前返回，不消耗指令。

完整反組譯證據與檔案雜湊見 `docs/re/35-loan-command-cost.md`。

## 2. Go／Ebiten 接線

- `LoanResult.CommandCompleted` 在正信用度、正額度送出後設定，即使 `Approved=false`。
- `cmd/dsds` 隨機拒絕時提交 `CommandBudget.Spend`；信用 gate 不提交。
- 若防禦性 `Spend` 失敗，恢復貸款前的 Rand seed 與畫面狀態，避免半提交。
- 原典／現代白話以 `diplomacy.loan.refused` 分開顯示；原典不加入規則說明，白話
  明確提示已消耗指令。

## 3. 驗收

1. 規則測試確認核准與隨機拒絕都 `CommandCompleted=true`。
2. 信用度零 gate 確認 `CommandCompleted=false`、Rand／黃金／帳本不變。
3. 語系檔含 `diplomacy.loan.unavailable` 與 `diplomacy.loan.refused` 兩種模式。
4. Docker 內執行規則／語系／`cmd/dsds` Xvfb、no-cgo、deny scan、diff check。

## 4. 非目標

- 其他外交失敗分支的指令成本。
- 第一期援助國對應、`.DT2` 正常戰鬥寫回、Android 與正式三平台包。
