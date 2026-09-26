# 反組譯證據：第一期外援代碼選擇

> 輸入：`WAR.EXE`（未打包）  
> SHA-256：`11dbfcf24686ab7765f788b38514cefd2039d0f60b6bd517d89fb5a84c068015`  
> 工具：IDA Pro `9.4.0.260610`  
> 位址：IDA linear address；`[di-5319h]` 另標為 `.DT1` 區塊 7 的反查值

## 1. 呼叫點與資料流（confirmed）

`sub_21D1D` 在 `loc_21FCA` 以目前省份的司令 ID 找到 `[di-5319h]`，再以
`var_14`（同一支函式開頭的 `Random(10)` 結果或日期特殊覆寫前的值）選擇援助
代碼。`[di-5319h]` 是區塊 7 `FactionOfGeneral` 的 1-based 勢力代碼；
`internal/game.FactionOfGeneral.SlotOf` 與 `FactionTable.SlotOfLeader` 的索引對齊
已由 session 載入一致性檢查固定。

`22088h` 之後才依援助代碼選除數：`99`／`142` → 1，`100` → 3，其餘 → 5。
非第一期在同一處以 `byte_6FE88 != 1` 強制代碼 `99`；因此第二／三期不應把
第一期分支套用到玩家 UI。

## 2. 第一期間的數值映射（confirmed）

| `[di-5319h]` 勢力代碼 | `var_14`／原始 roll | 援助代碼 |
|---:|---|---:|
| `10` | 奇數 → 99；偶數 → 100 | 99／100 |
| `1` | 任意 | 142 |
| `5` | 任意 | 100 |
| `4` | 6 → 100；7 → 146；其他 → 99 | 100／146／99 |
| `3` 或 `2` | 任意 | 101 |
| 其他 | 6 → 99；7 → 100；8 → 101；9 → 141；其他保留原 roll | 99／100／101／141／0..5 |

代碼 `146`、`141` 的國家名稱未證實；它們只作原版數值 ID。`div 2` 後的
`xchg ax,dx` 是「比較原始 roll 除以 2 的餘數」，所以勢力代碼 10 的奇偶分支
不能簡化成除數或機率名稱。

## 3. Remake 接線

- `internal/game.AidDonorCode` 以 `stage`、1-based `factionCode` 與同一顆 roll
  回傳原版數值；stage 2／3 先回傳 99。
- `AIWorld.RequestAidForFaction` 在消耗援助判定 roll 後才選代碼，資源亂數順序
  與 `RequestAid` 保持一致；舊的固定 donor 入口仍保留給規則測試與相容呼叫端。
- `cmd/dsds.executeDiplomacyAid` 從目前司令反查勢力槽，無 `.DT1` 勢力表的新局傳
  0，走原版 default 分支；不把缺資料默認成美國或其他國家。

## 4. 邊界

本文件只閉合**數值代碼與除數的控制流**，不宣稱代碼等於某個國家，也不宣稱
第一期所有存檔／新局都已完成區塊 7／10 的同步。`.DT1` 區塊 10 的生成／寫回
時機仍見 `docs/formats/07-dt1-layout.md` 的未解清單。

