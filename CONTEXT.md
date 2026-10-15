# CONTEXT — 專案脈絡與文件索引

## 目前狀態（2026-10-15）

| 項目 | 現況與入口 |
|---|---|
| 主要分支 | GitHub 僅保留 `master`；已整併正式版、前版發行與研究分支全部提交，其餘三個遠端分支依使用者要求移除；整理見 [Issue #9](https://github.com/wicanr2/great-era-remake/issues/9)，清理見 [Issue #10](https://github.com/wicanr2/great-era-remake/issues/10) |
| 正式版 | [v.1.0.1-20261014](https://github.com/wicanr2/great-era-remake/releases/tag/v.1.0.1-20261014)，來源提交 `85ed56095117134fd3fdbf64efd7a5411eaeabf7`；已發布 tag、Release 與附件保留 |
| 呈現與音訊 | 大地色指揮介面、俯視六角戰場與原版鐵路，原版／HD 美術及介面可獨立切換；八曲沿用[已採用混音](assets/music/modern/acceptance.json) |
| 主線新增 | 全國地圖改用連續地表配色，消除固定經緯度造成的矩形明暗斷層；開發頁加入墾地、兵工廠及金礦插圖，原版切換與既有指令保留；詳見 [Issue #13](https://github.com/wicanr2/great-era-remake/issues/13)，尚未收錄於上述正式版 |
| 進行中的發行 | 使用者要求重新打包，`v.1.0.2-20261015` 收錄主線呈現修正；三平台完整版、公開包及重新錄製的比較片由 [Issue #14](https://github.com/wicanr2/great-era-remake/issues/14)追蹤，驗證完成後發布 |
| 人物 | 416 筆檔案、392 篇中文正文；24 位待考依使用者決定停止研究，#3 以 not planned 關閉；英日各 392 篇保留 machine-draft |
| 驗證 | 19 個 Go 套件；Linux 容器及 Windows／Wine 的完整包、公開包正常操作、存讀檔與重啟通過，四份存檔一致並經 Go parser 核對；[公開收據](https://github.com/wicanr2/great-era-remake/releases/download/v.1.0.1-20261014/RELEASE-VERIFICATION.json) |
| 交付與權利 | 公開三平台引擎 ZIP、八曲原創配樂與驗證附件；本機六包及 92.5 秒 DOS／重製原版／HD 重新實錄比較片集中在 `dist-all/v.1.0.1-20261014/`，完整版、原版資料、倚天字庫及影片不公開上傳 |
| 後續 | [Issue #6](https://github.com/wicanr2/great-era-remake/issues/6)續追 Android；[Issue #7](https://github.com/wicanr2/great-era-remake/issues/7)續追 Windows／macOS 真機、簽署、公證及 Linux FUSE；原版未知行為維持既有證據限制 |

現行工作由 [GitHub 總 Issue #1](https://github.com/wicanr2/great-era-remake/issues/1)與子 Issue 追蹤。文件分工：[README](README.md)為穩定入口，[WORKLOG](WORKLOG.md)保存整理後的工作歷程；本檔保存現況及有效文件索引，不重新啟用舊待辦。

發行契約及現行展示收據見[整合規格](docs/spec/48-hd-release-integration.md)，平台使用方式見[發行文件](docs/release/README.md)。主線新呈現使用原創小型圖集；已發布程式、封包、tag 與 Release 附件保留。

未提交變更審查已補存研究工具的主鍵盤數字具名鍵功能，沿用同一字元轉換與掃描碼表，DOS 服務、probe 與 oracle 三套件回歸通過。兩段舊筆記保存於[玩法差異文件](docs/playtest/43-remake-original-gameplay-diff.md#未提交研究紀錄補存)，等級為 `historical-unreviewed`，不重開原版研究或據此修改規則。過期本機首頁及重複摘要已退役，必要的工具功能與歷史原文已在主線保存。

## 資訊與文件入口

唯一開發主線為專案根目錄的 `master`。現況只在本檔維護，後續工作只在 GitHub Issues 維護；README 只提供下載、操作與導航，AGENTS 只保存作業規範。

| 資訊 | 權威位置 |
|---|---|
| 現況與有效決策 | 本檔 |
| 待辦與外部驗收 | [GitHub 總 Issue](https://github.com/wicanr2/great-era-remake/issues/1)與子 Issue |
| 工作歷程及已退役文件索引 | [WORKLOG.md](WORKLOG.md) |
| 作業規範 | [AGENTS.md](AGENTS.md)；CLAUDE 入口直接匯入同一份 |
| 玩家執行方式 | [公開引擎](docs/release/README.md)、[本機完整版](docs/release/FULL-LOCAL.md) |
| 現行 HD 契約 | [SPEC-48](docs/spec/48-hd-release-integration.md) |
| 規則及忠實度限制 | [機制索引](docs/mechanics/00-index.md)、[SPEC-30](docs/spec/30-approximate-validation-m2.md) |
| 人物撰寫與整合 | [撰寫規格](docs/design/21-biography-authoring-spec.md)、[SPEC-11](docs/spec/11-authored-biography-overlay-p2c.md) |
| 原版證據定位 | [函式索引](docs/re/00-function-index.md)、[玩法紀錄](docs/playtest/43-remake-original-gameplay-diff.md) |

過去的現況表、已停止工作清單及被後續方案取代的設計文件不再放在目前工作樹。其完整內容由 Git 提交歷史保存，入口集中於 WORKLOG；原始研究輸入、獨有證據與已發布交付物不因去重刪除。文件去重與引用核對見 [Issue #11](https://github.com/wicanr2/great-era-remake/issues/11)。

README 現為故事背景、玩法及 HD 特色介紹，使用八張 DOS／原版模式／HD 畫面，地圖與開發頁已更新為主線預覽。最新影片仍為本機 `promo/great-era-remake-promo-comparison-v.1.0.1-20261014-20261015.mp4`，由該版本交付頁提供；原片與既有發行附件保留。來源、展示範圍及驗證見 [SPEC-48](docs/spec/48-hd-release-integration.md#readme-遊戲介紹與重新拍攝的展示)、[Issue #12](https://github.com/wicanr2/great-era-remake/issues/12)及 [Issue #13](https://github.com/wicanr2/great-era-remake/issues/13)。
