# CONTEXT — 專案脈絡與文件索引

## 目前狀態（2026-10-17）

| 項目 | 現況與入口 |
|---|---|
| 本輪交付 | v.1.1.0-20261017 已正式發布，六份封包、76 秒本機影片及 README 十三張實錄已驗證完成。[Issue #32](https://github.com/wicanr2/great-era-remake/issues/32)保存發行工作；README 已補新版操作。工具、附件回下載及發布收據入口為 `workplace/hoi4-ui-prototype/release-v.1.1.0-20261017/README.md` |
| 主要分支 | GitHub 僅保留 `master`；已整併正式版、前版發行與研究分支全部提交，其餘三個遠端分支依使用者要求移除；整理見 [Issue #9](https://github.com/wicanr2/great-era-remake/issues/9)，清理見 [Issue #10](https://github.com/wicanr2/great-era-remake/issues/10) |
| 正式版 | [v.1.1.0-20261017](https://github.com/wicanr2/great-era-remake/releases/tag/v.1.1.0-20261017)，來源提交 `36157deb67d86f40f29093e673e3f60cb22cabc8`；五份公開附件回下載 SHA-256 全部相符，GitHub latest 已切換。舊 tag、Release 與附件保留；[公開驗收收據](https://github.com/wicanr2/great-era-remake/releases/download/v.1.1.0-20261017/RELEASE-VERIFICATION.json) |
| 呈現與音訊 | 主線預設 Modern、HD 美術、滑鼠及 HD 音樂，美術／介面／音樂可獨立切換。原版介面保留；本機完整版含八首原曲 HD 與三首軍歌，公開引擎包保留已採用八首原創 Ogg。接線見 [SPEC-60](docs/spec/60-original-hd-music.md) |
| v1.0.3 已發行修正 | 常駐五欄工具列加入 HD／原版／靜音、自動／八首選曲及偏好保存；完整版零參數啟動自行定位資料，存檔寫使用者目錄；見 [Issue #17](https://github.com/wicanr2/great-era-remake/issues/17)及 [SPEC-51](docs/spec/51-live-music-startup.md) |
| 人物 | 416 筆檔案、392 篇中文正文；24 位待考依使用者決定停止研究，#3 以 not planned 關閉；英日各 392 篇保留 machine-draft |
| 驗證 | 本輪十九個 Go 套件、純 Go 及資產拒絕掃描通過。Linux／Wine 完整包與公開包四條正常流程共 68 張實錄；保存重開 SHA-256 一致，三首軍歌實際輸出非靜音。十一曲另經 44.1／48 kHz 三次完整循環；[本機驗收](dist-all/v.1.1.0-20261017/RELEASE-VERIFICATION.json)。不擴張為 Windows／macOS 真機或原版規則對拍 |
| 交付與權利 | 六份封包、76 秒比較影片與收據集中於 `dist-all/v.1.1.0-20261017/`。公開僅三平台引擎及兩份校驗／驗收附件；本機完整版、原曲 HD、軍歌、原版資料、字庫及比較影片不公開。舊交付、tag 與附件保留 |
| 正式工具列 | [SPEC-50](docs/spec/50-permanent-toolbar.md)、[SPEC-51](docs/spec/51-live-music-startup.md)及 [SPEC-60](docs/spec/60-original-hd-music.md)已接入主線。模式／選曲保存、三首軍歌、缺檔保護及零參數預設完成；本機 [play-current.sh](workplace/hoi4-ui-prototype/play-current.sh)接同版執行檔與十一曲，保留使用者存檔。平台真機仍由 #7 追蹤 |
| 後續 | [Issue #6](https://github.com/wicanr2/great-era-remake/issues/6)續追 Android；[Issue #7](https://github.com/wicanr2/great-era-remake/issues/7)續追 Windows／macOS 真機、簽署、公證及 Linux FUSE；原版未知行為維持既有證據限制 |
| 推廣片重錄 | [Issue #18](https://github.com/wicanr2/great-era-remake/issues/18)已完成 76 秒本機影片，含新錄 DOS／重製原版／HD、Modern、軍備、雙方資訊、暫停續戰與配樂展示。影音、黑幀、凍結、字幕與逐章視覺核對通過。原片與工具由本輪工作區 README 索引，正式影片為 `dist-all/v.1.1.0-20261017/promo/great-era-remake-promo-comparison-v.1.1.0-20261017.mp4` |
| 新配樂方向 | [Issue #19](https://github.com/wicanr2/great-era-remake/issues/19)已完成八首完整原曲 HD 與《還我河山》《國民革命歌》《歌八百壯士》，僅本機使用。7,183 筆原曲有效旋律音符、歷史全譜、同步分軌、Ogg、循環與遊戲指定選曲已驗證。先前片段配器方向已採用，不宣稱全曲或各平台人耳確認。來源與收據見 `workplace/hoi4-ui-prototype/original-hd-music-20261015/README.md` |
| 統一後抗日擴充 | 唯一計畫為 [Issue #20](https://github.com/wicanr2/great-era-remake/issues/20)及 #21–#29。原版三劇本的 10／2／2 勢力已接入原生開局、固定玩家與 `.gesave` v1，保留明示期別的 DT1 匯入；原有 14 勢力回歸、正常戰鬥／整編保存、三語及零參數 HD 音訊證據保留。擴充現在開放明示的軍備整備試玩與 v2 保存，供中國勢力使用；當年在役名冊、東北入侵、完整裝甲／穿甲、增援與最終勝負仍未完成。變更已整合 master，本輪交付由 #32 追蹤；不把整備接入宣稱完整抗日戰役 |
| 陸軍裝備研發 | [Issue #28](https://github.com/wicanr2/great-era-remake/issues/28)及 [SPEC-55](docs/spec/55-armament-research.md)的軍備整備達 CONFORMED：實際資源研究、生產、庫存、逐批換裝、裝備火力及玩家步兵改編裝甲。改編保留將領／部隊 ID，退回步兵裝備，保存餘數預備人員；單一 v2 保存完整進度。19 套件、12 中國勢力保存與 seed 1 的正常 GUI 15 項／37 張實錄通過。F7 開啟軍備局，F6 保留工具列。完整入侵、穿甲及敵軍改編政策仍由 #24–#27 追蹤 |
| 雙方戰鬥詳情 | [Issue #29](https://github.com/wicanr2/great-era-remake/issues/29)及 [SPEC-56](docs/spec/56-battle-information.md)達 CONFORMED；「5 查閱」或滑鼠列出雙方兵種兵力、將領、戰技、武裝、體力、士氣、機動及來源省資源。擴充另列實際配裝、缺裝與火力係數。原版／HD 的滑鼠查閱、翻頁、返回及保存重啟已實跑，不消耗回合或 RNG。原版未知的厚度、穿甲、車砲實數不猜補；[HD 實錄](docs/images/battle-information-hd-20261016.png)、[原版介面實錄](docs/images/battle-information-original-20261016.png) |

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
| 擴充開局與保存證據 | [SPEC-52](docs/spec/52-expansion-campaign-foundation.md)保留完整擴充 DRAFT；[開局原型已採用](docs/images/opening-approved-20261016.png)，原版子集已依 SPEC-54 接入正式玩家路徑。工具、盤點、原型與實跑收據入口為 `workplace/hoi4-ui-prototype/expansion-foundation-20261016/README.md`，工作仍只由 #20–#28 追蹤。日軍可玩原版戰役，抗日擴充供中國勢力遊玩 |
| 研發操作與史料證據 | [SPEC-55](docs/spec/55-armament-research.md)保存採用流程、[原圖](docs/images/research-approved-20261016.png)及[原生畫面](docs/images/research-native-20261016.png)。本機原型為 `workplace/hoi4-ui-prototype/research-20261016/README.md`，正式接入的工具、正常實跑及收據為 `workplace/hoi4-ui-prototype/research-native-20261016/README.md`，最終為 `gui-r5/receipt.json`、`checks-report.json` 與 `final-check.json`。原型與史料收據保留，與正常遊戲驗收分開 |
| 雙方戰鬥資訊 | [SPEC-56](docs/spec/56-battle-information.md)及 [Issue #29](https://github.com/wicanr2/great-era-remake/issues/29)：雙方並列兵種、兵力、逐將與擴充裝備；原版未知的車輛、火砲實數與穿甲不猜補 |
| Modern 操作與返回 | 使用者已採用的 [SPEC-58](docs/spec/58-modern-interaction.md)達 CONFORMED：1280×800 常駐工具列、固定返回／地圖／確認、具名選省／選將及數量表單接回可玩版。整編取消還原兵力、武裝及裝備帳，保存重開仍可取消；戰鬥可暫停查閱地圖、保存及續戰，期間阻止新命令與換月。19 套件及 seed 1 正常 GUI 13 項／57 張通過；[地圖](docs/images/modern-map-native-20261016.png)、[徵兵](docs/images/modern-recruit-native-20261016.png)。本機收據入口為 `workplace/hoi4-ui-prototype/modern-native-20261016/README.md`，最新建置由 [play-current.sh](workplace/hoi4-ui-prototype/play-current.sh)啟動，原版／HD 美術、介面及音樂獨立。[#30](https://github.com/wicanr2/great-era-remake/issues/30)／[#31](https://github.com/wicanr2/great-era-remake/issues/31)已整合主線，本輪交付由 #32 追蹤，平台真機留 #7；舊證據、素材及發行保留，已整合 master；本輪發行由 #32 追蹤 |
| 原生儲存入口補查 | [SPEC-59](docs/spec/59-native-save-entry.md)達 CONFORMED：原生新局可從其他選項開啟儲存確認，兩個入口共用既有 autosave 的能力判定；19 套件與 seed 1 正常 GUI 6 項／10 張通過，取消不寫檔、原版保存及正常續玩 bytes 一致。收據在 `modern-controls-20261016/save-entry-20261016/`；[play-current.sh](workplace/hoi4-ui-prototype/play-current.sh)可試玩最新本機建置，獨立保存於 `player/current.gesave`。新 Modern 構圖、整編取消及戰術安全返回已依 SPEC-58 完成本機接入；本列保存 SPEC-59 的先前收據，已整合 master；本輪發行由 #32 追蹤 |
| 劇本可選勢力資料 API | [SPEC-53](docs/spec/53-scenario-faction-catalog.md)為 CONFORMED，遊戲規則套件測試通過；只處理唯讀目錄，不代表完整開局或控制權已接入 |
| 正式開局接入契約 | [SPEC-54](docs/spec/54-native-campaign-start.md)為 CONFORMED，限定原版三劇本的新局、固定玩家身份、世界／戰鬥／整編保存；[原生開局畫面](docs/images/opening-native-20261016.png)。最終收據在本機工作區 `native-r4/`、`native-zero-r3/` 與 `native-final-check.json`；統一後抗日入侵依後續規格啟用 |
| 規則及忠實度限制 | [機制索引](docs/mechanics/00-index.md)、[SPEC-30](docs/spec/30-approximate-validation-m2.md) |
| 人物撰寫與整合 | [撰寫規格](docs/design/21-biography-authoring-spec.md)、[SPEC-11](docs/spec/11-authored-biography-overlay-p2c.md) |
| 原版證據定位 | [函式索引](docs/re/00-function-index.md)、[玩法紀錄](docs/playtest/43-remake-original-gameplay-diff.md) |

過去的現況表、已停止工作清單及被後續方案取代的設計文件不再放在目前工作樹。其完整內容由 Git 提交歷史保存，入口集中於 WORKLOG；原始研究輸入、獨有證據與已發布交付物不因去重刪除。文件去重與引用核對見 [Issue #11](https://github.com/wicanr2/great-era-remake/issues/11)。

README 以北伐與抗戰故事開場，介紹玩法、Modern 與 HD、十四勢力開局、軍備整備及即時切換。[新版操作](README.md#新版操作modern)集中說明滑鼠流程、返回與取消、戰術暫停續戰、存讀檔及快捷鍵。十三張本版實錄與故事只由 README 維護；影片與下載入口為本機 `dist-all/v.1.1.0-20261017/index.html`。歷史原片與舊交付保留，來源及驗收見 [SPEC-48](docs/spec/48-hd-release-integration.md)與 #32。
