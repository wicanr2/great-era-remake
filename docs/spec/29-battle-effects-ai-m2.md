# SPEC-29：戰鬥協同／特殊副作用與結算接線 M2

狀態：**READY**（2026-08-11）

本規格把目前已由 `WAR.EXE`／IDA Pro 9.4 閉合的戰鬥 helper 接到純 Go 規則層，並
明確分離「可玩的 remake 行為」與「原版仍缺正常玩家 oracle 的部分」。它不宣稱
第二層六鍵選單的五個 handler、動畫、音效、彈藥寄存器或 `.DT2` 未解 bytes 已等價。
目前直接證據顯示按鍵 `1..5` 才有 handler call，按鍵 `6` 在 `sub_4BF27` 落到清理
路徑；這不代表第 6 鍵的玩家語意已解出。

## 1. 證據契約

| 行為 | 證據等級 | 出處 |
|---|---|---|
| `sub_5301B` 每次雙方經驗 +5、體力 −1、士氣 −1，再呼叫 `sub_51D68` | confirmed | `docs/re/44-combat-handler-helper-boundary.md`、`WAR.EXE` `5301Bh` |
| 經驗達 100 時清零並提升 `+0`，`+0` 已達 100 時經驗保留 100 | confirmed | `WAR.EXE` `sub_5A3B2` |
| `sub_51F19` 士氣降至 10 以下會再扣 `Round(force/3)` | confirmed | `WAR.EXE` `sub_51F19` |
| `sub_52EEA` 任一參與者兵力歸零後立死亡旗，協同迴圈停止 | confirmed | `WAR.EXE` `sub_52EEA` |
| `sub_53111` 掃攻擊者六鄰、排除兵種 4、按目標勢力領袖分攤 | strong inference | `WAR.EXE` `53111h`，`docs/re/44` |
| `sub_58449` 衝鋒回合數、目標損失 ×2/3、兵力 ≤3 的停止條件 | confirmed | `WAR.EXE` `58449h` |
| `sub_58613` 衝鋒後雙方體力／士氣／經驗副作用 | confirmed | `WAR.EXE` `58613h` |
| `sub_57B15` 炮擊三次、攻方不吃反擊；攻方體力 −5、士氣按 `Round(morale/10)`、雙方經驗 +5 | confirmed | `WAR.EXE` `57B15h` |
| 兵種 4／5 目標的特殊 handler 只做畫面／音效，不扣兵力 | confirmed（控制流） | `WAR.EXE` `sub_4B854`、`sub_4BCBE` |

## 2. 規則層契約

`BattleSim.EngageWithEffects` 是單次正規攻擊的唯一實作；`Engage` 只回傳相容的
兩個損失整數。正規 helper 的副作用先套用，再計算 `sub_51D68` 的戰損，避免用
攻擊前兵力計算；炮擊／衝鋒則依各自已證實的「戰損後副作用」順序處理。

`EngageCoordinated`：

1. 從攻擊者格依方向 1..6 掃鄰格；只收存活、同目標勢力領袖、非兵種 4 的單位，
   排除主目標，最多六名。
2. 主攻的目標損失除以支援數（無支援則不除）；攻擊者損失不除。
3. 依同一順序讓每名支援單位攻擊主攻者，支援單位損失除以支援數。
4. 任一對戰參與者歸零時停止後續支援，回傳 `SkipSupportAfterDeath=true`。

`EngageCavalrySpecial`：兵種 6 對兵種 1／6、且目標地物不在 `12..21` 長城段時，
依 `floor(Ability/30)+floor(Morale/40)+1` 執行衝鋒 pass；每次目標損失為原始目標
損失的 `Round(2/3)`，攻擊者照原始損失；任一兵力 ≤3 即停止。所有 pass 後再套用
目標體力 −2、攻方體力 −5、目標士氣 −3、攻方士氣 −10、雙方經驗 +15。

`ResolveBattleAttack` 是 UI／AI 的分派入口：

- 炮兵且目標非相鄰 → `EngageRanged`；
- 目標兵種 4／5 的特殊畫面 → 零損失 `SpecialVisualOnly`；
- 騎兵符合衝鋒 gate → `EngageCavalrySpecial`；
- 有六鄰支援候選 → `EngageCoordinated`；
- 其餘 → `Engage`（呼叫端若明確選五次 handler，使用 `EngageRepeated`）。

目標兵種 4／5 的特殊分支優先於一般近戰，但不攔截兵種 4 的非相鄰炮擊；
這是依 `sub_4B854`／`sub_4BCBE` 的窄控制流所作的 remake 分派補完，不等於
原版玩家選單的六種攻擊已全部命名。

## 3. 撤退／部署／結算

- `NewBattleSim` 是唯一攻方部署入口：守方落點由呼叫端提供，攻方依 `DeployZone`
  與佔用表掃描；`NoCell`／待命命令在入口正規化。
- `RetreatCandidates` 提供來源省與戰場省的省份表鄰接順序作為 remake 泛化候選，
  並在結果中標 `Source=province-neighbour-fallback`；目前唯一原版 oracle
  `18←19` 的順序仍以固定樣本覆蓋。這個泛化清單是 remake 差異，不是原版 parity。
- `ApplyBattleSettlement` 對勝方、守方勝與回合上限平局清除戰鬥旗；立即撤退的
  `BattleSettlement` 本身不改世界，由 UI 的 `clearBattleFlag` 與候選目的地切換
  完成。這避免把撤退誤當成勝負或易主。
- `SyncBattleStatsToGenerals` 同步 Force、AbilityA、F19、F20、Experience、Stamina、F30，
  `Raw` 仍由 `General.Bytes` 保留未解欄位。

## 4. AI parity gate

本輪閉合的是戰鬥規則 helper、13 行動 handler 的正常執行入口、死亡停止條件與
決策追蹤統計；`BattleRunStats.Unimplemented` 必須為 0 才能把一場 remake 路徑視為
可重播。仍不得宣稱原版逐回合 parity，因為 `byte_64901` 方位對照、部分 gate、
`sub_567B9` 目標／路徑時機與一般 `.DT2` oracle 尚未閉合。

## 5. 驗收

- 純規則測試：副作用、支援分攤、死亡停止、衝鋒 pass／terrain gate、炮擊副作用、
  兵力／人物同步、泛化撤退清單。
- Docker `go test -count=1 ./...`、`tools/check_no_cgo.sh`、`tools/deny_scan.sh --all`、
  Windows `CGO_ENABLED=0` 建置與 `git diff --check`。
- 正常玩家樣本仍只可把 `18←19` 的撤退清單標為原版 oracle；其他清單與 AI 差異需
  在交付文件中保留 evidence level。
