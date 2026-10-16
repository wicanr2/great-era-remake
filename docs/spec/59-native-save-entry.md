# SPEC-59：原生新局的儲存入口

狀態：CONFORMED，限原生新局的儲存入口修正。工作屬於 [Issue #31](https://github.com/wicanr2/great-era-remake/issues/31)的操作補查，由 [#30](https://github.com/wicanr2/great-era-remake/issues/30)追蹤整體 Modern 重規劃。本項不依賴新構圖或整編取消方案。

## 原因與證據

修正前 `screenOtherOptions` 的儲存條件為 `savePath != "" && origSave != nil`，拒絕沒有匯入 DOS 存檔的原生戰役。既有 `autosave` 會先判斷 campaign 並呼叫 saveCampaign；指揮介面底部儲存已接受 campaign 或 origSave。兩個入口條件不一致，與實際 writer 能力不符。本次已共用 canSaveSession，writer 不變。等級為 confirmed-remake；修正前正常玩家重現及來源雜湊保存在本機 `modern-controls-20261016/save-entry-20261016/`。

## 最小契約

- 可儲存入口需有非空 savePath，且有 campaign 或 origSave；不以副檔名或畫面美術推測保存能力。其他選項與指揮介面共用此判定。
- 左鍵選擇「儲存」只開啟既有確認頁，不立刻寫入；返回、ESC、右鍵沿用 Cancel，不執行 writer、不扣資源／指令、不推進 RNG。
- 確認仍呼叫唯一 autosave。原生戰役分派 v1／v2，DT1 使用原 bytes 寫回；不更改保存格式、內容、版本或路徑。
- 路徑是目錄等寫入失敗仍由既有 writer 報錯，保留遊戲與舊保存；入口判定不能假稱磁碟可寫。沒有保存來源的舊入口仍拒絕。
- 原版介面與指揮介面均可開啟同一確認，不依賴主題；不修改原始素材、配樂、戰鬥或 AI。

垂直鏈：既有原生／DT1 狀態 → 共用保存入口判定 → screenSaveConfirm → 取消或 autosave → 原有 writer／parser → 正常續玩。此項是已授權的操作缺陷修正，不增加 DOS 規則或新的原版 oracle。

## 驗收

先以修正前執行檔從正常新局進入其他選項重現拒絕。修正後以 seed 1 正常開局、滑鼠確認、返回／取消、儲存、離開與續玩抽驗；保存後遊戲狀態、指令及 RNG 不變。另抽驗原版介面儲存及現有 DT1 原始 bytes 寫回測試，確認失敗輸出不破壞進度。沿用已存在的套件回歸、素材、差異及 Docker 清理閘門，證據只涵蓋實際完成的平台。

本機 `save-entry-20261016/` 收據：

- `before/receipt.json` 與 `before/fresh.log`：正常新局請求儲存，日誌為「目前沒有可寫回的存檔基底」，訊息結束後仍在其他選項。
- `after/receipt.json`：6 項正常流程、10 張視窗畫面；徵兵等舊返回收據保留，這輪補驗先前未實跑的儲存確認。滑鼠返回不寫入新檔、取消不改既有保存；確認產生有效 .gesave，原版介面保存、退出與正常續玩後 bytes 一致，世界／預算／RNG 不變。
- `checks-report.json`：19 個 Go 套件與隔離建置通過；包含既有 DT1 未知 bytes 保留及保存失敗回歸，不新增新的保存格式。
- `launcher-check.json`：本機 play-current.sh 指向已驗二進位與獨立 current.gesave，play.sh 預設仍為正式版；容器參數攔截只證明啟動接線，不能冒稱桌面音訊或其他平台已驗。

只有 Linux／Xvfb 正常玩家驗證；新 Modern 構圖、整編取消與戰術安全返回仍未接入。沒有原版 oracle、Windows／macOS 真機、提交、推送或發行聲明。
