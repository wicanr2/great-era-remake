# SPEC-53：劇本可選勢力目錄資料

狀態：CONFORMED，僅限唯讀資料 API。實作前已標 READY，資料與規則套件測試通過。完整開局、玩家控制權、原版敵軍姓名映射與保存接線仍依 [SPEC-52](52-expansion-campaign-foundation.md)及 [Issue #21](https://github.com/wicanr2/great-era-remake/issues/21)，不因這份規格宣稱新局已可玩。

## 證據與範圍

使用者要求三劇本全部有效勢力可選。輸入格式、三劇本檔名與 SHA-256、10／2／2 勢力盤點及名表限制見 SPEC-52 的唯一證據表；本規格不重新複製來源。既有 `ParseTownFile`、`ParseGenerals`、`ProvinceTable.Factions`、`General.Province` 與 `General.Force` 提供所需語意，不需新增原版 RE 假說。

本 API 是 `confirmed-data` 的目錄產生器，不更動遊戲規則、外交、原始 bytes、勢力槽順序或正式啟動入口。沒有原版選單 oracle，不宣稱原版曾允許選取所有勢力。

## API 契約

`ScenarioFactionChoices(stage uint8, town *ProvinceTable, generals []General) ([]ScenarioFaction, error)`：

- 期別由唯一 `ScenarioByStage` 取得，使用已解析的 TOWN 與該期有效 MAN 記錄。
- 拒絕未知期別、nil 省表、帶存檔 Date 的省表，以及不等於該期有效筆數的將領切片，避免把進度快照或第 192 筆後殘料當新局。
- 第一期只納入 1–36 省，後兩期 1–39 省；無主省與該期範圍外記錄不產生勢力。
- 沿用 `ProvinceTable.Factions` 分組，不新增另一套勢力判定規則。每個有領地司令 ID 必須在有效將領範圍內，且領袖所屬省須在其轄區。沒有可選勢力時回錯誤。
- 按領袖 ID 遞增輸出，結果不依賴 map 遍歷順序。這是資料 API 的穩定順序，UI 順序仍由開局原型確認。
- 每項包含領袖 ID、領袖所在初始省、全部轄區、轄區所屬有效將領 ID 及 `uint32` 兵力總和。將領以其 Province 對應的省份司令歸屬，不用有殘留的 `FactionOfGeneral` 查非領袖。
- 保留沒有接入姓名的有效部隊，尤其第二、三期 107–191。姓名與勢力顯示標籤交給語系及後續姓名映射，不在規則層猜名字。
- API 不寫輸入、不返回可修改輸入的切片。結果不能當完整世界或保存快照。

## 驗證

實際三劇本資料核對 10／2／2 勢力、有效記錄邊界、各領袖初始省及後兩期敵方 85 筆部隊。北伐第 141 位趙恆惕必須存在，不能承襲附帶存檔的九勢力結果。

合成資料只驗證輸入邊界與純資料行為：錯誤期別、筆數、存檔省表、越界領袖與初始省、無領地、範圍外記錄、可重複排序、結果不別名，以及兵力總和超過 u16。這些測試不替代正常開局、合法指令、AI、存讀檔或原版 oracle。

## 實作與驗證收據

入口為 [`scenariochoices.go`](../../internal/game/scenariochoices.go)，測試為 [`scenariochoices_test.go`](../../internal/game/scenariochoices_test.go)。2026-10-16 於既有 `logh3-hd-preview:20261011-r1` 容器執行 `go test -count=1 ./internal/game` 通過；本機 `game-tests.log` 保存結果，`scenario-audit.json` 以獨立分組盤點核對新 API 的三劇本勢力、領地、將領及兵力。

盤點收據保存 Go 版本、工具來源雜湊與原始檔雜湊。確認範圍是相同原始資料的目錄結果，不是原版玩家路徑或可通關聲明。其餘舊 Go 檔只訂正姓名與存檔槽註解，去除註解後語法樹相同；未修改既有規則函式。
