# SPEC-54：正式新局、固定玩家身份與單檔保存

狀態：CONFORMED，限定本規格的原版新局與保存。依使用者已採用的 [開局畫面](../images/opening-approved-20261016.png)、[SPEC-52](52-expansion-campaign-foundation.md)及 [SPEC-53](53-scenario-faction-catalog.md)實作。範圍為三個原版劇本的全部有領地勢力。統一後抗日模式保留畫面與限制說明，入侵尚未接妥時不得開始該模式。

工作入口為 [#21](https://github.com/wicanr2/great-era-remake/issues/21)及 [#22](https://github.com/wicanr2/great-era-remake/issues/22)。本規格不授權變更戰損、AI 強度或原版勝負條件。

## 新局初始化證據

輸入為解包 `WAR.EXE`，SHA-256 `11dbfcf24686ab7765f788b38514cefd2039d0f60b6bd517d89fb5a84c068015`。工具為 IDA 9.4，映像 `ida-pro-9.4-idapython:py312-v1`，ID `4ac62de83339c215bab10e455cee3a22d9c6efed9fd0d8ed0f068327b83a06ab`。下列均為 IDA 線性位址，原名、指令、bytes、xref 及工具探針保存在本機 `workplace/hoi4-ui-prototype/expansion-foundation-20261016/ida-newgame/export.json`，由該目錄 README 索引。正式資料庫及原始 EXE 未修改。

| 結論 | 等級與定位 |
|---|---|
| 選劇本後載入該期 MAN／TOWN，清停火、戰爭暫存，再初始化外交與勢力 | confirmed，`sub_391E1`，`0x393AD–0x39478` |
| 三劇本的日期、省數與期別 | confirmed，`sub_38C01`，`0x38DAD–0x38DF3`；沿用 `Scenario` |
| 依有效省編號掃描，收錄首次出現的非零司令，建立 24 槽領袖清單 | confirmed，`sub_32F03`，`0x32F32–0x32FA7`；`sub_32EC0` 的 `0x32EDB–0x32EFA` 掃描 24 槽、比對傳入司令並回傳重複標記。這是靜態資料流證據，不宣稱原版同狀態 parity |
| 寫領袖反查值及勢力表領袖，非自我關係值 100、自我 0，第二矩陣 0 | confirmed，`sub_32FB2`，`0x33066–0x33139`；包括對未使用槽的 100，不以附帶存檔的消滅後矩陣重建 |
| 槽位 +2 初始值 | confirmed，`sub_32FB2`，`0x32FC0–0x33010`；第 1 槽 12，第 2 槽第一期 9、其餘期 3，後續為 11、14、1、10、4、7、5、2 |
| 每省戰爭暫存 +4 至 +59 清零 | confirmed，`sub_5ADA6`；+0／+2 不在此函式清除範圍，新局的無戰鬥零值屬重製版狀態，不推稱 DOS 未知 bytes |
| 外交 credit 100、debt 0 | confirmed，`sub_391E1`，`0x39458–0x39475`；沿用 `NewDiplomacyLedger` |
| 十槽清單由玩家選擇填入，與 24 槽領袖表不同 | confirmed，`sub_38839`、`sub_391E1`；單人重製新局僅記玩家，不改現行戰力 callback |

新局不得讀取範例 SAVE、DT2 或 MEM_WAR 當初始模板。複製原始 typed TOWN／MAN，未啟用省記錄仍原樣保存，規則迴圈只處理該期有效省。領袖所在省為初始選中省。原版資料的 Raw bytes 保留。新局沒有 DOS 原始 DT1，不捏造未知區塊或允許匯出偽 DT1。

## 玩家路徑

零參數進入採用畫面的原生開局選單，使用現有常駐工具列。模式、劇本、勢力均可用鍵盤及滑鼠選擇。所有三劇本原版勢力可開始，包括兩個抗戰劇本的日軍；姓名未核對的映射保持既有限制。美術、介面、音樂不決定戰役模式。

開始前完整建立快照並驗證。失敗不替換舊狀態。明示 `-province` 或舊 DT1 `-save` 保留既有載入入口，避免破壞既有操作與驗收工具；新局與繼續進度採 `.gesave`。開始新局不預先覆蓋既有進度，F10 確認或儲存命令才保存。

新戰役固定玩家領袖。查看敵省仍可查閱及設定，不能開戰、徵稅或進入修改型指令；AI 以玩家領袖排除玩家省，不能因觀看敵省而交換身份。跨年採當前劇本的期別。

日本姓名的全域 ID 映射仍未確認。日軍勢力使用語系名稱；缺少名表對應的將領在原版字模介面顯示編號，不能因越界中斷整個畫面。開局標籤使用現有內嵌 Unicode 字型，仍逐語系驗證字模覆蓋。

## 保存 v1

本節保存原版開局子集的 v1 契約與驗收範圍。現行容器寫入及舊版本遷移改依 [SPEC-63](63-post-unification-invasion.md)，統一使用 v4；原版開局、固定玩家與 Base 資料規則不因此改變。

一個 ZIP 容器包含 `manifest.json`、`state.json` 及可選 `legacy.dt1`。格式識別為 `great-era-save`、版本 1。每一項有 SHA-256，拒絕重複項目、未知名稱、未知版本、巢狀路徑、解壓總量超過 8 MiB、毀損、缺欄及不合法索引。讀取只在記憶體解碼，不解壓到玩家檔案系統。

世界狀態包含完整 typed DT1State、省／將領 Raw、玩家領袖、期別、目前省、日期、原始來源雜湊、LCG 32-bit 狀態、每省已具體化的剩餘指令數、世界 Units／Strengths、戰力選項純量及 EnableExtra。原始 DT1 匯入時原樣另存，後續世界修改寫入 state，不覆寫 legacy。新局無 legacy。偏好仍由既有偏好檔處理。

戰鬥另保存兩方完整 Combatant、來源與戰場省、司令、回合與上限、選中單位、指令層、資源、結算分類及已套用標記、AI 記錄與全期單位。還原佔用表及 ID 索引，不再次部署、重置機動力或重跑結算。一般臨時輸入對話恢復到穩定地圖或戰鬥畫面；已扣款的世界與戰鬥狀態不得重做。整編已把兵力移入私有兵力池時，須額外保存兵種、候選、兵力池、武裝總量及 Changed 標記，恢復到整編選將畫面，配完才依原有 finish 路徑扣指令；不能丟棄兵力池後返回地圖。匯入 DT2／MEM 的原始副本也收在 state，後續在容器內保存，不另寫旁檔。

寫入以現有 `writeFileAtomic` 同目錄暫存、同步與替換。替換前失敗保留舊檔，不宣稱斷電持久性。載入先完整驗證與建立交叉引用，再一次套用；失敗保持現行遊戲。擴充模式與未來 schema 須由後續 READY 規格另行啟用，v1 不接受尚未實作的入侵狀態。

## 驗證與檔案入口

實作入口：`internal/game/newcampaign.go`、`internal/game/campaignsave.go`、`cmd/dsds/campaign.go`、`cmd/dsds/opening.go`、`internal/ui/render/opening.go`。各檔的針對測試保存在同目錄；私有正常操作收據掛入原型 README。

以固定 seed 1 核對全部 14 個勢力的原始資料、新局、身份、合法命令、AI、存讀與 RNG／預算；另抽測戰鬥中途保存與結算不重刷、毀損保存不替換、DT1 匯入原 bytes 保留。正常 GUI 至少從開局以鍵盤／滑鼠各選一個勢力，操作、存檔、關閉、重啟。原生畫面在相同選擇狀態與採用圖對照。最後執行相關 Go 套件、資產拒絕掃描、變更範圍與 Docker 清理核對。單元測試不等於原版 parity。

## 驗證結果

原生介面與正常玩家鏈已接入，[零參數開局截圖](../images/opening-native-20261016.png)由實際執行檔擷取。三劇本的 14 個勢力通過資料、合法指令、AI、保存及重啟的 session 測試；Linux 真視窗抽測蔣中正、趙恆惕及兩期日軍，包含鍵鼠、原版／HD 美術與兩種介面。戰鬥與整編均經正常入口，中途 F10 保存、重啟並繼續；整編完成確認兵力守恆與指令扣除。

正式 Go 套件回歸見本機 `native-all-tests-r3.log`。GUI 最終收據為 `native-r4/native-playtest.json`，包含事件與逐圖雜湊；畫面採用相同模式／劇本／領袖，核對構圖及主要配色取樣。首幀合成與訊息閘門導致的早期腳本失敗另保留，不記成產品缺陷。真正零參數從不同工作目錄啟動、預設 HD 音訊非靜音、使用者目錄保存及繼續進度見 `native-zero-r3/receipt.json`；技術非靜音不代替人耳確認，也不代表配樂 #19 已整合。

以上收據均由 `workplace/hoi4-ui-prototype/expansion-foundation-20261016/README.md` 索引。原始資料、既有配樂與已發布產物、素材掃描、純 Go 閘門、文件引用及擁有權由 `native-final-check.json` 核對。本次沒有原版 oracle，不宣稱原版完整 parity、Windows／macOS 真機、入侵及裝備完成；程式尚未提交／推送或發布。
