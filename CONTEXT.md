# CONTEXT — 專案脈絡與文件索引

## 目前狀態（2026-10-19）

| 項目 | 現況與入口 |
|---|---|
| 正式版 | [v.1.2.0-20261019](https://github.com/wicanr2/great-era-remake/releases/tag/v.1.2.0-20261019) 已發布並設為最新版本，tag 指向 `120b60170a101a08ea165c1305cc8db01d04fe87`。三平台公開 ZIP 與兩份校驗／驗收附件共五份，回下載 SHA-256 相符；舊 tag、Release 與附件保留。[發布工作區](workplace/hoi4-ui-prototype/release-v.1.2.0-20261019/README.md)保存收據，由 [#37](https://github.com/wicanr2/great-era-remake/issues/37)追蹤 |
| 交付與雙片 | 六份封包及原版主題 90 秒／HD 68 秒兩片集中於 `dist-all/v.1.2.0-20261019/`。[觀看與封包](dist-all/v.1.2.0-20261019/index.html)、[公開驗收](dist-all/v.1.2.0-20261019/RELEASE-VERIFICATION.json)、[本機候選收據](dist-all/v.1.2.0-20261019/DELIVERY-VERIFICATION.json)。公開僅引擎三包與校驗；完整包、原版資料、字庫、原曲 HD、軍歌及影片留本機。候選來源快照保留，正式公開包從乾淨提交重建，三平台程式與完整包逐位元組相同 |
| 完整抗日戰役 | [SPEC-63](docs/spec/63-post-unification-invasion.md)限定已驗本機範圍達 CONFORMED，已整合 master 並收錄正式版：統一判定、宣戰插圖、十九位月制名冊／自傳／照片、6＋2＋2、真實守城與撤退、弱點追擊、有限補給、v4 與勝敗。1937／1941 正常宣戰、真實反攻及兩支新影片有實錄；1926 統一與長期勝敗邊界明示受控，不宣稱自然長期通關或原版 oracle。[功能收據](workplace/hoi4-ui-prototype/invasion-20261018/README.md) |
| 原版鍵盤回饋 | [SPEC-66](docs/spec/66-original-keyboard-feedback.md)限定本機達 CONFORMED，已發行：2／3 與數字鍵盤、Backspace／Delete、情境式 0／Esc、省地圖返回全國、完整直式部隊資料及翻頁。正常 GUI 五組／38 張及新原版 90 秒影片通過；唯讀與取消不改世界／RNG／指令／軍備帳。[收據](workplace/hoi4-ui-prototype/keyboard-feedback-20261018/README.md) |
| HD 亮底 | [SPEC-64](docs/spec/64-light-hd-interface.md)限定本機達 CONFORMED，已發行：砂鐵亮底、磨損金屬、原版／Modern 深色／亮底切換。正常 GUI 與地圖、照片保色核對通過；深色仍為預設 |
| 裝甲與軍備 | [SPEC-55](docs/spec/55-armament-research.md)、[SPEC-65](docs/spec/65-equipment-armor.md)已整合發行。研究、生產、庫存、換裝及裝甲改編接回正常流程，快速／戰術共用裝甲、穿甲與損失帳。減傷上限 50%，戰車砲穿甲 30／50／70；這些是已採用的重製設計值，不冒稱原版資料 |
| 人物與照片 | 保留原版 416 筆、392 篇中文正文及停止考證的 24 位，另加入十九位日軍自傳與照片，共 411 篇正文、三十位可靠歷史肖像。英日仍標示 machine-draft。原版 JAPAN 全域姓名對應未知，不猜補；來源與權利見 [肖像紀錄](docs/licenses/portraits-NOTICE.md) |
| 呈現與音訊 | 正式版預設 Modern 深色、HD 美術、滑鼠與 HD 音樂。常駐六欄可分別切換美術、介面、操作、音樂及軍備＋抗日，既有 AI 不變。公開包八首原創曲，完整包十一首原曲 HD／軍歌；最後兩首各增加 40 秒，技術循環及 Pulse 輸出通過，延長段人耳確認仍待 #33 |
| 封包驗證 | 十九 Go 套件、純 Go、素材拒絕掃描、三平台架構與逐檔清單通過。完整／公開包的 Linux 與 Wine 四條正常玩家流程已驗，存檔 SHA-256 一致；缺軍歌的公開包安全保留既有選曲。macOS 僅交叉建置與靜態檢查，不擴張為真機或原版對拍 |
| 新推廣片 | 兩片獨立錄製正常 1937 新局、seed 1，展示自然宣戰、首波六支與真實反攻第二回合。原版另示範鍵盤修正與返回；HD 另示範軍備、植田謙吉歷史照片、滑鼠返回及亮底。共 36 幕目視、完整影音、字幕、黑幀與非靜音通過；原版靜態表單的全畫面凍結誤判以實際 230→23→0 證據核對。原片由 [交付工作區](workplace/hoi4-ui-prototype/delivery-20261019/README.md)索引 |
| 主要分支 | GitHub 僅保留 master；這次來源已提交並推送，既有分支清理決策不變。根目錄舊推廣文字含過期私人分享連結，保留於本機，未提交 |
| 後續 | 長期自然通關與平衡玩家回饋由 #20／#27 追蹤；延長音樂聽審留 #33；Windows／macOS 真機、簽署、公證及 Linux FUSE 留 #7，Android 留 #6。已驗收功能不因歷史規格重新啟動 |

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
| 原版鍵盤回饋契約 | [SPEC-66](docs/spec/66-original-keyboard-feedback.md)與 [Issue #35](https://github.com/wicanr2/great-era-remake/issues/35)：數字鍵、輸入更正、返回全國地圖與完整部隊資料 |
| 裝甲與穿甲 | [SPEC-65](docs/spec/65-equipment-armor.md)：2026-10-18 採用 50% 減傷上限、型錄係數及戰車砲補齊提案；實作與驗收由 #25 追蹤 |
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

README 以北伐與抗戰故事開場，介紹玩法、Modern 與 HD、十四勢力開局、軍備整備及即時切換。[新版操作](README.md#新版操作modern)集中說明滑鼠流程、返回與取消、戰術暫停續戰、存讀檔及快捷鍵。正式版與開發預覽實錄、故事與操作只由 README 維護；影片與下載入口為本機 `dist-all/v.1.2.0-20261019/index.html`。歷史原片與舊交付保留，來源及驗收見 [SPEC-48](docs/spec/48-hd-release-integration.md)與 #32。
