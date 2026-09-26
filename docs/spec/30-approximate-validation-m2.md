# SPEC-30：免 DOSBox 的近似驗證閘門 M2

狀態：**READY**（2026-08-11）

## 1. 範圍決策

使用者已選擇 B 方案：本 remake 接受「近似原版」的可玩性與資料安全宣稱，
不建立執行原始 `WAR.EXE` 的函式級 oracle，也不為了逐回合 parity 長時間跑
完整 DOSBox 劇本。

這個決定把兩種結果嚴格分開：

| 宣稱 | 可以說什麼 | 不可以說什麼 |
|---|---|---|
| `confirmed-static` | IDA 已確認的位址、分支、欄位大小與呼叫順序 | 未確認的欄位名稱或完整玩家時機 |
| `confirmed-remake` | Go writer／規則／UI 在固定輸入下自洽且可重播 | 原版 byte-for-byte 或逐回合等價 |
| `approximate` | remake 依已知證據接上合理的副作用、撤退／部署與 AI 時序 | 原版完整 handler、副作用或 AI parity |
| `unknown` | 尚未有證據的原版語意與落盤時機 | 用測試綠燈、影片或零值補洞 |

後續文件與發行說明可使用「近似原版的 remake 行為」；`original parity`、
「完整還原」與「原版等價」仍保留給未來真正取得 oracle 的研究分支，
不列為本 remake 的交付前置條件。

## 2. 四層免 DOSBox 驗證

### 2.1 DT1／DT2 差分遮罩

以原始檔副本（或全哨兵合成檔）作為 writer 基底，執行後只允許已解欄位的
offset 改變：

1. `ParseDT1`／`WriteDT1`：省份、將領、已證實勢力／停火／帳本、明確選定的
   `WarRecordBranch4`／`WarRecordBranch8` 與區塊 7 領袖格。
2. `ParseBattleStates`／`WriteBattleStates`：`+0..+7`、`+28/+48`、
   `+68/+268` 與 `+468`；`+8/+18` 及其他未解 bytes 必須保持原值。
3. 每份測試同時檢查：未修改快照 byte-for-byte、單欄位 diff 不越界、錯誤長度／
   不明確分支 fail-closed。

這能驗證「非破壞性寫回契約」與「差分遮罩」，不能宣稱原版 writer 在同一時機
寫入相同資料。相關基礎測試見 `internal/game/dt1save_test.go`、
`internal/game/ceasefire_test.go` 與 `internal/game/battlestate_test.go`。

### 2.2 handler／第 6 鍵

以固定 `Combatant` 狀態矩陣執行 Go 的唯一攻擊分派入口，涵蓋：

- 正規、協同、衝鋒、炮擊與特殊視覺分支；
- 經驗、體力、士氣、低士氣兵損、支援分攤與死亡停止；
- 兵種、相鄰／非相鄰、彈藥與邊界值。

第二層原版選單的靜態事實仍單獨保存：`1..5` 有 handler call，`6` 沒有第六個
handler call。remake 中的 `battle.attack-target.1..6` 是**目標序號**，不是原版
六個命名攻擊；沒有第六個目標時輸入層拒絕，不能把它誤報成原版第 6 鍵功能。

因此可宣稱「已知副作用的近似規則已接通」，不可宣稱五個原版 handler 的完整
選擇條件、`+7A9A/+7A9B` 真實名稱、動畫／音效／彈藥時機或第 6 鍵玩家語意已解。

### 2.3 撤退／部署／結算

不跑完整劇本，直接以固定地圖、佔位表與省份表做正常入口的合成測試：

- 攻方部署：`DeployZone` 十格、掃描順序、敵鄰避讓與滿位 fail-closed；
- 守方部署：`NWMAP` `0x4000` 候選格與穩定掃描順序；
- 撤退：已證實 `18←19` 樣本固定保留，其他省份使用鄰接順序 fallback，並在
  結果標示 `province-neighbour-fallback`；
- 結算：立即撤退只清除 remake 戰鬥旗與切換目的地，勝負／平局走同一個
  `ApplyBattleSettlement`，未知存檔 bytes 不碰。

這可宣稱「remake 撤退／部署／結算路徑可重播且近似」，不宣稱所有原版候選排序、
部署來源、駐軍後處理或 `.DT2` 落盤時機。

### 2.4 AI 時序

以固定 seed／固定 `BattleChainGates` 執行 `TraceDecisions`／`AutoResolveByChain`，
把每回合收斂為可比較的事件序列：

```text
turn → branch → action → movement → engagement → settlement
```

必要閘門：`Unimplemented == 0`、回合上限、補給先於回合遞增、決勝門檻、
`DefaultPostStageOpen` 邊界、觀測 trace 不得改變結果。這些是 remake 的
deterministic contract；不是原版 `sub_3A9F4`／`sub_3AABA` 的逐回合 oracle。

## 3. 最小驗收命令

所有命令仍在 Docker 內執行；不需要 DOSBox、IDA runtime 或網路：

```sh
tools/go.sh test -count=1 ./internal/game ./internal/ui/actions
# cmd/dsds 與完整套件測試需沿用 dsds-go:1.25 並在容器內啟動 Xvfb :99
tools/go.sh test -run 'Approximate|BattleState|WriteDT1|WriteWarRecord|Deploy|TraceDecisions|AutoResolve' ./internal/game ./internal/ui/actions
bash tools/check_no_cgo.sh
bash tools/deny_scan.sh --all
```

綠燈只代表 `confirmed-remake`／`approximate` 閘門通過；交接時仍需列出原版
`unknown` 清單，不得將這些命令描述為 parity 測試。

## 4. 明確排除與重新開啟條件

本規格排除：

- 原始 `WAR.EXE` 的局部 emulation／函式級 oracle；
- 完整 DOSBox 回合劇本、正常勝負後 `.DT2`／`MEM_WAR.DAT` 原版差分；
- 五個 handler 的逐分支動畫／音效／資源寄存器等價；
- 所有撤退候選、部署時機與 AI 預約表生命週期的原版 parity。

只有使用者另行要求原版等價，或出現足以改變玩家體驗的矛盾證據，才重新開啟
oracle 分支；否則維持本規格的近似邊界，避免 remake 再次陷入無限逆向迴圈。
