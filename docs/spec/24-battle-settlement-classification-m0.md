# SPEC-24：戰鬥收尾分類 M0

狀態：**READY**  
日期：2026-08-10

## 1. 目的

把已由 `WAR.EXE`／正常 DOSBox 路徑證實的戰鬥結束原因整理成單一規則型別，供
Ebiten 畫面、戰報與後續 `.DT2` 寫回切片共用。這一層只分類已知狀態，不猜戰後省份
易主、將領移防或未解欄位的 bytes。

## 2. 證據邊界

| 收尾 | 證據等級 | 寫回判定 | 出處 |
|---|---|---|---|
| 立即撤退 | confirmed | 不呼叫 `sub_3964E` | `docs/playtest/16-dt2-live-battle-snapshot.md` §4b、`docs/spec/21-battle-retreat-ui-m1.md` |
| 已分出勝負（`byte_64901` 為 1 或 2） | confirmed | 不因 `sub_3964E` 寫回 `.DT2` | `docs/re/05-mem-war-record.md` §6、`docs/mechanics/30-combat.md` §X |
| 回合上限且 `byte_64901 == 0` | confirmed | 需要進入 `sub_3964E` 寫回路徑 | 同上；非二月上限 16、二月上限 15 |

「需要寫回」只表示原版控制流程會進入該函式；實際 469-byte `.DT2` 欄位同步時機與
部分欄位仍未解，不能由本規格直接產生檔案差分。

## 3. 規則型別與契約

- `game.ClassifyBattleSettlement(turn, turnCap, winner)` 是唯一的已結束戰鬥分類入口。
- `winner` 只能是 `BattleSideNone`、`BattleSideFirst` 或 `BattleSideSecond`。
- `winner` 為第一／第二方時，分類為 `BattleSettlementDecisive`，並標記不需 `.DT2`
  寫回；這包含全滅、補給見底與已證實的必勝結算，呼叫端另保存具體原因。
- `winner == BattleSideNone` 且 `turn >= turnCap` 時，分類為
  `BattleSettlementTurnLimitDraw`，標記需要 `.DT2` 寫回。
- `winner == BattleSideNone` 且尚未達上限不是收尾；函式回傳錯誤，避免把中途畫面當成
  平局。
- `turnCap <= 0`、未知勝方或其他非法值 fail-closed。
- `BattleSettlementImmediateRetreat()` 另建立即撤退結果，勝方為 none 且不需寫回；
  它不接受回合數，也不改變世界狀態。

## 4. 不在本切片

- 推導一般撤退候選、省份佔領與將領移防。
- 產生或猜測 `.DT2`／`MEM_WAR.DAT` 的未解 bytes。
- 將 UI 結束畫面自動改成地圖，或替玩家決定戰報確認流程。
- 戰鬥第二層六鍵選單的完整 handler／第 6 鍵語意、駐軍、查閱與完整原版 AI。

## 5. 驗收

1. 純 Go 單元測試覆蓋三種已證實分類、上限 15／16、未結束與非法值。
2. `cmd/dsds` 的全滅、補給見底、回合上限與立即撤退分支都透過同一分類入口留下
   `BattleSettlement`，不直接散落寫回旗標。
3. Docker 內 `go test -count=1 ./...`、Windows `CGO_ENABLED=0` 建置、
   `tools/check_no_cgo.sh`、`tools/deny_scan.sh --all`、`git diff --check` 通過。
4. 完成宣告只涵蓋分類與可追溯旗標；不得宣稱正常平局 `.DT2` 寫回已完成。

## 6. 2026-08-10 勘誤與後續切片

本規格原本把「原版 `.DT2` 寫回時機尚未閉合」與「remake 副本 writer」放在同一個
缺口中。現在兩者分開：原版 oracle／完整欄位仍未閉合，但使用者已授權先做可玩的
remake 持久化，故新增 READY `docs/spec/25-battle-state-writeback-m1.md`。該 writer
只從 `Raw` 改寫已證實欄位，且對所有已結束戰鬥保存快照；立即撤退的已證實不寫回
行為不變。這項實作不推翻本規格對原版控制流的分類，只補上明列差異的 remake 輸出層。
