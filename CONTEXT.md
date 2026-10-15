# CONTEXT — 專案脈絡與文件索引

## 目前狀態（2026-10-15）

| 項目 | 現況與入口 |
|---|---|
| 主要分支 | GitHub 僅保留 `master`；已整併正式版、前版發行與研究分支全部提交，其餘三個遠端分支依使用者要求移除；整理見 [Issue #9](https://github.com/wicanr2/great-era-remake/issues/9)，清理見 [Issue #10](https://github.com/wicanr2/great-era-remake/issues/10) |
| 正式版 | [v.1.0.3-20261015](https://github.com/wicanr2/great-era-remake/releases/tag/v.1.0.3-20261015)，來源提交 `60bc1521ae0de7135ed07ef3bd276a4b0ea75c10`；五份公開附件雜湊核對相符，舊 tag、Release 與附件保留 |
| 呈現與音訊 | 大地色指揮介面、俯視六角戰場與原版鐵路，原版／HD 美術及介面可獨立切換；八曲沿用[已採用混音](assets/music/modern/acceptance.json) |
| 本版修正 | 常駐五欄工具列加入 HD／原版／靜音、自動／八首選曲及偏好保存；完整版零參數啟動自行定位資料，存檔寫使用者目錄；見 [Issue #17](https://github.com/wicanr2/great-era-remake/issues/17)及 [SPEC-51](docs/spec/51-live-music-startup.md) |
| 人物 | 416 筆檔案、392 篇中文正文；24 位待考依使用者決定停止研究，#3 以 not planned 關閉；英日各 392 篇保留 machine-draft |
| 驗證 | 19 個 Go 套件、seed 1 的 40 張正常操作實錄、25 條 Linux 音訊路徑；Linux／Wine 完整包零參數啟動及公開包匯入、切換、存檔重啟通過，四份存檔一致並經 Go parser 與原位元組寫回核對；[公開收據](https://github.com/wicanr2/great-era-remake/releases/download/v.1.0.3-20261015/RELEASE-VERIFICATION.json) |
| 交付與權利 | 三平台本機完整版、三份公開引擎及驗收集中於 `dist-all/v.1.0.3-20261015/`；公開附件含八曲原創配樂；完整版、原版資料及倚天字庫不公開上傳。比較影片保留於前版目錄，見交付頁的影片入口 |
| 正式工具列 | [SPEC-50](docs/spec/50-permanent-toolbar.md)及 [SPEC-51](docs/spec/51-live-music-startup.md)達 CONFORMED，已收入本版；美術／介面／音樂獨立切換、現有 AI 說明、輸入偏好、右鍵返回及滾輪支援完成。本機預覽 `workplace/hoi4-ui-prototype/play.sh` 已接同版正式 Linux 執行檔；Windows／macOS 真機驗收另由 #7 追蹤 |
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
| 常駐工具列與滑鼠 | [SPEC-50](docs/spec/50-permanent-toolbar.md)，含已採用原型與正常操作驗收範圍 |
| 即時音樂與雙擊啟動 | [SPEC-51](docs/spec/51-live-music-startup.md)，已驗證並正式發行；本機收據入口為 `workplace/hoi4-ui-prototype/audio-default-20261015/README.md` |
| 規則及忠實度限制 | [機制索引](docs/mechanics/00-index.md)、[SPEC-30](docs/spec/30-approximate-validation-m2.md) |
| 人物撰寫與整合 | [撰寫規格](docs/design/21-biography-authoring-spec.md)、[SPEC-11](docs/spec/11-authored-biography-overlay-p2c.md) |
| 原版證據定位 | [函式索引](docs/re/00-function-index.md)、[玩法紀錄](docs/playtest/43-remake-original-gameplay-diff.md) |

過去的現況表、已停止工作清單及被後續方案取代的設計文件不再放在目前工作樹。其完整內容由 Git 提交歷史保存，入口集中於 WORKLOG；原始研究輸入、獨有證據與已發布交付物不因去重刪除。文件去重與引用核對見 [Issue #11](https://github.com/wicanr2/great-era-remake/issues/11)。

README 以北伐與抗戰故事推廣文開場，介紹玩法、HD 特色、遊戲中即時切換及正式啟動器的配樂預設，保留八張 DOS／原版模式／HD 畫面；地圖與開發頁已收錄於正式版。文案見 [Issue #15](https://github.com/wicanr2/great-era-remake/issues/15)。最新影片為本機 `promo/great-era-remake-promo-comparison-v.1.0.2-20261015.mp4`，由該版本交付頁提供；舊版原片、影片與發行附件保留。來源、展示範圍及驗證見 [SPEC-48](docs/spec/48-hd-release-integration.md#v102-20261015-呈現修正重新發行)及 [Issue #14](https://github.com/wicanr2/great-era-remake/issues/14)。
