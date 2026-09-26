# 規格 13：談判停火玩家介面 M1

狀態：**READY（M1 窄切片）**  
日期：2026-08-10

本規格只閉合政略指令 10「談判停火」的玩家輸入、判定與指令成本；不宣稱
外交系統、停火長期語意或戰爭記錄的所有欄位都已解出。

## 1. 原版證據契約

證據輸入：`workplace/ida/WAR.EXE`／`workplace/ida/WAR.EXE.i64`。`WAR.EXE`
SHA-256：
`11dbfcf24686ab7765f788b38514cefd2039d0f60b6bd517d89fb5a84c068015`。
工具：IDA Pro 9.4；以下位址均是 IDA linear address，不是遊戲 `ds:` 偏移。

| 原版位置 | 證據 | 等級 |
|---|---|---|
| `sub_10193+1F7` | 政略選單第 10 項呼叫 `sub_211D5` | confirmed |
| `sub_211D5` | `cmp byte_6FE88, 1`；不等於 1 時畫「無法使用」並返回 | confirmed |
| `sub_211D5` | 第 1 期檢查目前省司令的駐地 `+7A81` 是否等於目前省；不符畫「司令不在本省」 | confirmed |
| `sub_211D5` | 輸入省編號的上限使用 `$`（36）；目標司令欄位為 0 時不進判定 | confirmed |
| `sub_211D5` | 目標省 `+32 bit 6（40h）` 未設時畫「並無戰事」並返回 | confirmed |
| `sub_20CF0` | 掃已部署將領，依 `sub_5A0B9` 累加攻守戰力；`Random(10)` 的 3／8 兩檔門檻 | confirmed |
| `sub_20E05` | 同意與拒絕兩條結果都設 `byte_6FE81=1`；同意另對 `ds:BCA5h` 做 byte `inc` | confirmed |
| `sub_103DF` | 只有 `byte_6FE81` 非 0 才 `dec` 指令數；因此無法使用、司令不在本省、無戰事不扣 | confirmed |

`sub_21168` 只負責依 `ds:BCA5h` 與 `ds:B346h` 取對手，戰爭記錄其餘欄位仍是
未知；M1 不把它們命名成攻方／守方，也不把停火 byte 命名成「剩餘月數」。

第 1 期可用省數 36、第二／三期 39 的劇本表已由
`docs/re/31-battle-ai-chain.md` §50 確認；因 `sub_211D5` 在第二／三期先走
「無法使用」，本規格的玩家輸入上限固定取第 1 期的 36。

## 2. Remake 接線契約

- `internal/game` 提供停火前置檢查：第 1 期、司令駐在目前省、目標省編號
  `1..36`、目標有司令、目標 `InBattle()`。檢查失敗不消耗亂數、不改狀態。
- `AIWorld.NegotiateCeasefire` 維持既有 confirmed 戰力／亂數公式；同意時只
  增加 `CeasefireState[target]`，由既有 `.DT1` writer 寫回該 39-byte 表。
- `cmd/dsds` 新增 `screenCeasefireTarget`。原典畫面沿用 `1.15`／`2.15` 的
  「司令欲在何省談判停火？」字模；現代白話使用 wording catalog。兩者共用
  `numericKeypadTargets`、`actions.Digit*`、`actions.DeleteDigit`、
  `actions.Submit` 與返回按鈕。
- 合法目標進入判定後，無論同意或拒絕都只扣一次目前省指令；扣款失敗時回復
  停火 byte 與亂數 seed。無戰事等前置拒絕回到政略選單且不扣指令。
- 指令 10 的滑鼠／觸控入口使用既有 `actions.Select10`；鍵盤另保留 `0` 作
  remake 快捷鍵，不改原版規則或存檔格式。

## 3. 明確不在 M1

- `ds:B346h` 60-byte 戰爭記錄除已證實的欄位外仍不寫回。
- 對手姓名、原版訊息的完整動態排版與音效不作新的規則假設；結果文字只保留
  已證實的「同意／拒絕／在停火」語意。
- Android／三平台封裝與實機觸控驗收另立發行規格；本切片只保證輸入動作走
  共用 adapter。

## 4. 驗收

1. `go test ./internal/game` 覆蓋第 1 期 gate、司令駐地、36 上限、無戰事不改
   狀態、同意／拒絕的指令成本與停火 byte。
2. `go test ./cmd/dsds ./internal/ui/render ./internal/ui/actions ./internal/ui/layout`
   在 Xvfb 下通過；pointer 測試確認第 10 項與數字鍵盤命中。
3. `tools/check_no_cgo.sh`、`tools/deny_scan.sh --all`、`git diff --check` 通過；
   Docker 工作批次後沒有留下本輪建立的容器。

