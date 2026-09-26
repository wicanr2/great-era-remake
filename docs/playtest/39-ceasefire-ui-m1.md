# PLAYTEST-39：談判停火玩家 UI M1

日期：2026-08-10  
狀態：**PASS（remake 規則／呼叫端窄切片；不代表原版停火完整路徑已驗收）**

## 證據邊界

本切片使用 `WAR.EXE` SHA-256
`11dbfcf24686ab7765f788b38514cefd2039d0f60b6bd517d89fb5a84c068015`，以 IDA Pro
`9.4.0.260610` 讀取線性位址。`sub_10193+1F7` 呼叫 `sub_211D5`；stage 1 的玩家
省份輸入上限是 36。`sub_211D5` 另確認目前司令必須在目前省，目標省需有司令且
省份旗標 `+32` bit 6（`40h`）表示交戰；否則原版直接返回。成功後
`sub_20E05` 的同意／拒絕都設定 `byte_6FE81`，同意再對 `ds:BCA5h` 目標省 byte
執行 `inc`；`sub_103DF` 依該 byte 扣一次指令。

這份文件只把上述呼叫端與成本語意接入 remake，不把尚未重播的正常 DOSBox 停火畫面、
信用 gate 或戰鬥寫回宣稱為已證實。

## Remake 驗收

- `internal/game.CeasefireRequester`：stage／36 省、世界／目前省與司令所在地 gate。
- `internal/game.ValidateCeasefireTarget`：目標範圍、目標司令與交戰 gate。
- `CombatUnit.Deployed`：對應執行期 `+16` bit 2，讓存檔部署單位能進入停火戰力累加。
- `cmd/dsds`：指令 10 可由鍵盤 `0`、滑鼠或觸控進入 `screenCeasefireTarget`；原典
  使用 `DrawCeasefireTarget`，現代模式使用共用數字鍵盤與雙用語訊息。
- 未通過 gate 不消耗指令；同意與拒絕各消耗一次，只有同意遞增停火狀態。這個成本
  契約由固定 seed 的規則測試與 UI 狀態測試護欄。

## Docker 驗證

以下均在一次性、無網路 Docker 容器執行：

```text
go test ./internal/game ./internal/i18n ./internal/ui/render
DISPLAY=:99 go test ./cmd/dsds ./internal/ui/actions ./internal/ui/layout
```

兩組均通過；`ceasefire_test.go` 另驗證司令／交戰 gate 與 `Deployed` 戰力計算，
`pointer_test.go` 驗證 `actions.Select10` 與停火數字頁共用命中契約。工作容器使用
`--rm`，本輪沒有需要留下的專案容器。

## 尚未完成

- 正常 DOSBox 從可重播存檔進入停火、逐步輸入省份並取得原版畫面对照。
- 信用度不足貸款 gate、拒絕貸款的指令成本、第一期援助國原版選擇。
- `.DT2` 戰鬥正常回合寫回與 Android 封裝／實機觸控驗收。
