# SPEC-27：戰鬥世界結算與人物自傳可驗收 M2

狀態：**READY**  
日期：2026-08-10

## 1. 目的

把目前已能執行的戰鬥與人物自傳，收斂成一條可驗收的 remake 玩家路徑：

1. 戰鬥結束後，已證實的世界欄位不留在戰鬥副本裡，省份司令、交戰旗與存活攻方
   將領的所屬省要一致。
2. 戰損同時回到存檔側將領表與政略 AI 使用的戰力鏡像。
3. 人物自傳入口只對真正接合的名冊槽位出現；有槽位但沒有可靠正文時顯示明確
   fallback，不用杜撰文章；「無省長」佔位槽不提供入口。
4. 戰鬥攻擊子狀態的 1..6 目標順序，要讓鍵盤、滑鼠與觸控看得到同一份目標。

本規格是 remake 完整度切片，不宣稱原版所有戰後副作用、第二層六鍵攻擊選單分支或所有人物
史料已解出。

## 2. 證據與邊界

- 省份 `ProvinceFlagInBattle` 的生命週期來自 `sub_15925`／`sub_54DAC`；勝負投影
  沿用 `internal/game.Province.Capture` 的已證實旗標與司令寫回。
- `Combatant.Strength.Force` 與 `General.Force` 都是已證實的兵力欄位；
  `SyncForcesToGenerals` 只改這一欄，不猜 `.DT2` 未解區域。
- 攻方勝的移防只處理存活、有效 `GeneralID` 的攻方單位；守方勝與回合上限平局
  只清除交戰旗，立即撤退不做世界投影。
- 原版第二層六鍵攻擊選單的完整 handler／第 6 鍵、彈藥消耗、動畫與音效差異仍是 unknown／remake 差異；
  目前只接已閉合的近身 `Engage` 與兵種 4 遠程 `EngageRanged`。
- `people.json` 與 additive overlay 是唯讀文化資料。沒有可追溯來源的槽位保持
  `unknown` fallback；不可用模板批量填字。

## 3. 實作契約

### 3.1 戰鬥世界層

- 開戰成功建立 `BattleSim` 後才設目標省 `ProvinceFlagInBattle`。
- ESC 放棄未結算戰鬥、已證實立即撤退與所有完成／平局結算都要清除該旗。
- `ApplyBattleSettlement` 是唯一世界投影入口；以 `battleState.settled` 保證
  F10／重複 autosave 不會再次套用。
- 攻方勝：目標省由來源省 `Commander` 接管；所有存活攻方將領的 `Province`
  （含 runtime `world.Units`）移到目標省；地圖游標跟隨目標省。
- 守方勝／回合上限平局：不改司令與將領省份，只清除交戰旗。
- `syncBattleForces` 同步所有參戰單位的 `Force` 到 `a.generals` 與
  `a.world.Strengths`，索引越界即失敗。

### 3.2 人物自傳與輸入

- `PeopleDB.PersonAt` 回傳 false 的排除槽位不可畫自傳入口，也不可由 pointer target
  進入；目前 #274「無省長」是明確排除。
- 有槽位但 `Biography == ""` 的人物仍保留入口，頁面顯示語系提供的 unknown
  fallback（目前一頁）。
- 第一／第二／第三期共 486 槽位，驗收結果為 485 個可接合、1 個明確排除；
  387 篇已有正文，其餘不以猜測補寫。
- 戰鬥攻擊模式的相鄰與已閉合遠程候選共用 `battleAttackTargets`；畫面在敵軍格上
  顯示穩定的 1..6 標號，`actions.BattleAttackTarget(n)` 供三種輸入裝置共用。

## 4. 驗收

1. `go test -count=1 ./internal/game`：世界結算、旗標、移防與未知資料邊界。
2. Docker/Xvfb `go test -count=1 ./cmd/dsds`：戰鬥攻擊目標、兵力鏡像、自傳入口、
   unknown fallback、排除槽與翻頁。
3. `go test -count=1 ./...`、Windows `CGO_ENABLED=0` 建置、
   `tools/check_no_cgo.sh`、`tools/deny_scan.sh --all`、`git diff --check`。
4. 完成宣告必須列明：這是 remake 世界投影與資訊入口；未解原版欄位仍由原始 bytes
   保留，不能寫成「原版戰鬥全等價」或「417 篇史料全部已撰寫」。
