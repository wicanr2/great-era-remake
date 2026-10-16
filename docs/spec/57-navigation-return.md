# SPEC-57：確認頁返回與調動提交

狀態：CONFORMED，限本規格的返回 Action 與調動提交修正。使用者於 2026-10-16 回報朋友進入其他畫面後不能返回，並明確允許操作與原版不同。唯一工作為 [Issue #31](https://github.com/wicanr2/great-era-remake/issues/31)，Modern 重規劃另由 [#30](https://github.com/wicanr2/great-era-remake/issues/30)及 SPEC-58 處理。

## 已確認原因與範圍

修正前 `navigationTargets` 的高解析流程分支一律送 actions.Back，但調動、徵兵、練兵、存讀檔及離開確認的 handler 只接受 actions.Cancel。畫面有返回按鈕卻無作用；右鍵已有 `pointerBackAction` 正確分派，本次已重用該入口，不複製第三份取消規則。等級為 confirmed-remake，來源為目前工作樹與正常玩家重現，來源雜湊及修正前執行檔記入本機 `modern-controls-20261016/baseline.json`。

調動將領清單的原版路徑已有 Submit，但高解析導覽沒有公告對應按鈕。當已選至少一位將領時，沿用空閒的 Next 導覽位置提供相同 Submit；未選人時不提供假提交。此修正不重新解釋 DOS 選單或修改部隊調動規則。

## 輸入與狀態

1. 高解析返回按鈕、右鍵與 ESC 對確認頁均代表取消；一般頁代表返回上一層。確認頁左鍵返回不能送出 Confirm，也不能執行原待確認動作。
2. 返回使用原有 handler，保留其草稿數量或選擇，不扣軍費、指令、裝備或 RNG，不寫檔或讀檔。
3. 調動提交與鍵盤 Enter 共用現有 Submit handler，進入物資輸入；不能直接移動部隊或扣指令。
4. 幾何仍由既有 ModernPageLayout 提供，繪圖與命中共用；不重畫已採用的軍備或戰術構圖。原版既有導覽及右鍵語意保留。
5. 整編剩餘兵力池禁止離開，是另一條原版規則，不能將本次 Action 修正冒稱已解決。取消整次並還原或保存待續的操作及保存契約在 SPEC-58 做原型後由使用者決定，不丟棄池中人員。

垂直鏈為既有狀態及規則 → navigationActions／pointerBackAction → 命中矩形 → 原有取消或提交 handler → 既有世界與保存。原始資料、存檔 schema、戰鬥、AI、已完成動作及音訊不改；無新原版 oracle。

## 驗收

至少從正常開局重現徵兵確認頁的失效返回，再以相同 seed 1 修正後重跑。抽驗六類確認頁、調動勾選／提交／返回、一般子選單、人物及兩種介面；取消前後核對世界、資源、指令及 RNG，保存重啟維持。Go 測試針對正確 Action 與選人提交條件，正常 GUI 證明實際命中。保存與原始 bytes 回歸、素材掃描、dirty-tree 及 Docker 清理另留收據。範圍只限本規格，不宣稱全部 Modern 重規劃或整編取消已完成。

本機收據入口為 `workplace/hoi4-ui-prototype/modern-controls-20261016/README.md`：

- `before-r2/receipt.json`：正常開局進入徵兵確認，左鍵返回前後畫面完全相同；ESC 可取消，完整保存不變。
- `after-r3/receipt.json`：8 項正常流程、25 張實際視窗畫面；徵兵、調動、練兵、讀檔、離開確認的滑鼠返回與原版取消通過，調動新 Submit 確實進入物資表單。所有取消後的完整 `.gesave` 與起點 byte-identical，含世界、資源、預算及 RNG。
- `checks-report.json`：六類確認 Action、未選人／已選人／取消勾選的提交條件，以及既有 19 套件回歸通過。儲存確認頁由 Action 測試涵蓋，未將它冒稱本輪正常新局 GUI 實跑。
- 人物與軍備沿用前輪 `research-native-20261016/gui-r5/receipt.json` 的正常返回證據，完整舊收據保留並由本輪基準核對。

驗證只限 Linux／Xvfb；沒有新原版 oracle、Windows／macOS 真機或正式發行聲明。本規格未涵蓋整編取消；後續採用及原生接線由 [SPEC-58](58-modern-interaction.md)記錄。
