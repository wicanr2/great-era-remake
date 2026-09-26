# 工作清單與交接：2026-08-11

本檔已合併原 `HANDOFF.md` 的完整交接紀錄，現為本儲存庫唯一的工作清單與交接入口。
後續請直接更新本檔；`CONTEXT.md` 仍是狀態單一真相來源，深層證據則連回 `docs/`。

專案開工於 2026-08-01，兩天內 252 個 commit。`CONTEXT.md` 是狀態的單一真相來源，
這一份只回答「現在該從哪裡接下去」。

---

## 1. 現在在哪裡

2026-08-02 最新修正：remake 先前載入存檔時仍從 `MAN(1).DAT` 取將領初值，
自動存檔也只寫省份，會丟失經驗、兵力、體力、士氣與將領忠誠度。現在已改為
`ParseSaveGenerals` 載入存檔將領區，`WriteSave` 非破壞性寫回省份與將領；慰勞
指令的體力／士氣／忠誠度也已全部接上並同步 AI 戰力鏡像。

原版素材已全部解密，規則層跑得動；目前收尾集中在**Modern 音樂內容／授權**、英日
人審譯稿、跨平台發行與保留的原版 unknown evidence boundary，不再把 Modern UI 或
SDFA register parity 當成前置卡點。

| | 狀態 |
|---|---|
| M0 資產解密 | ✅ `.TPC`／`.15`／`.RGB`／`.GLB`／`.MUS`+`.TIM`／`.DAT`／`NWMAP`／`RAIL.TPC` 全解，逐像素 round-trip 過 |
| M1 文本還原鏈 | ✅ 四步全數完成。51 個字模檔 6,174 字全部反查回 Big5 |
| M2 執行檔反組譯 | 🔶 5 支裡 4 支進 IDA（`SDFA.EXE` 解不開，優先度低）|
| M3 規則規格 | 🔶 政略 15/15 指令定位、戰鬥公式全解、AI 決策鏈結構 confirmed |
| M4 Go 規則層 | 🔶 戰鬥能端到端跑完；AI 決策鏈 + 執行層都實作了 |
| M5 呈現層 | ✅ 640×350 政略／戰鬥與 Modern UI 邏輯外殼已接；寬版／向量字型仍是 polish |
| M6 多語系 | 🔶 154 個 UI／面板／敘事語意鍵、英／日 387 篇 machine-draft 自傳 overlay、17 張新聞來源圖像畫廊已接；正式人審譯稿仍待 |
| Modern 音訊 runtime | ✅ `audio=modern` manifest／SHA-256／授權欄位、純 Go Vorbis、loop reader、18 frame crossfade 與缺檔降級已接；現代 `.ogg` 內容仍待 |

實測數字：**419 個測試全綠**、100 份 `docs/*.md`、6/6 規格標 READY、
`tools/deny_scan.sh` 掃 373 檔**原版資產零命中**。

### 這兩天最實質的兩件事

**存檔格式從兩個區塊變成六個。** `.DT1` 的十個 `$basg` 區塊，現在解開的有：
省份表、**每省戰爭記錄表**、**勢力表（24 槽 × 59 B，含 24×24 關係矩陣）**、
將領表、**停火狀態表**、**勢力領袖表**、**將領→勢力反查表**。
十大勢力的領袖 ID 與北伐史實逐一對得上（張作霖／閰錫山／馮玉祥／吳佩孚／孫傳芳／
袁祖銘／劉湘／唐繼堯／蔣中正 + 一個空槽）。

**自傳研究稿已完成，但產品整合沒有完成**：名冊 417 種寫法、417 個槽位的骨架皆已
完成；其中 30 位 `unknown` 依法不立傳，排除後 **387/387 位有唯一成文**。
舊統計 `high`／`medium`／`low` 合計 389 已被後續稽核與 C2／C3 合併改變，不得再拿來
推算篇數。2026-08-02 以執行期來源 `people.json` 重量，實際仍只有 **326 篇**
非空 `bio_zh`；其餘 61 篇已存在 `bios-batch*.md`，缺的是 DESIGN-21 尚未定案的
可追溯寫回／語系產生流程，不能直接複製進 `people.json`。
`people.json` 自己的 authoritative 分類仍是 confirmed 230／partial 96／unknown 91，
不能拿後來的稽核批次數字直接覆寫這三欄。
先前把 389 寫成「成文」是交接統計口徑錯誤，現保留此勘誤以免再次誤報。

---

## 2. ⚠️ 正在跑的事——接手第一件事是確認它們

> **✅ 2026-08-02：D1／D2／D3 三批已全數交件驗收完畢，自傳程式在此暫停。**
> 暫停原因是 token 預算，不是遇到障礙。
> **名冊 417 個槽位的骨架全部建完（`--cross` 顯示未做 0 筆）**，
> 21 個批次的品質閘全過，裁決寫成規格 §27。這是乾淨的停點。

三批的交件都照這個順序驗收過，續做時沿用：

1. `./tools/py.sh tools/bio_gate.py <批次>`——四項要全過，
   **自己重跑，不採信 agent 的回報**（規格 §7.3）
2. `./tools/py.sh tools/bio_gate.py --cross`——查跨批次重複
3. 把 `G-*-n`（規格沒涵蓋到的情形）寫成規格的下一節

**派 agent 期間主迴圈不要 `git add -A`、不要 `git add .`。**
2026-08-01 就是這樣把一支 agent 的**未完成檔**掃進 commit 的
（`CLAUDE.md` §10 有完整事故紀錄）。要 commit 就逐一列路徑。

---

## 3. 下一步的建議順序

### 3.1 DOSBox 輸入已重驗，正常玩家撤退路徑已閉合

2026-08-02 已用受限 Docker 容器重跑固定序列，主選單可進十個進度的載入選單；
PointerRoot 焦點修法有效。`tools/dosbox_smoke.sh` 是固定 gate，證據見 `docs/playtest/15`。
正常新局已再次重播至「江西省 8 月 1 日／蔣中正攻擊孫傳芳」部署畫面；
玩家人數後必須另按 `2 → Return` 選孫傳芳，舊操作備忘漏了這一頁。
同一工作階段已用固定六方向路徑進入陝西省（18）的真正戰鬥選單，按 `3` 後輸入合法
撤退省份 `19` 並回車，正常回到政略地圖。戰鬥選單、撤退提示與地圖後的 `.DT2`／
`MEM_WAR.DAT` 雜湊均未改變；這是「尚未執行攻擊或回合結算的立即撤退」控制樣本，
不代表有戰損的戰鬥也不寫回。IDA `.i64` 已證實 `sub_3964E` 只由戰鬥主函式尾段
`sub_39B6E+6F4` 呼叫，而且只有結束旗標新立、`byte_64901 == 0` 的未分勝負分支
才會進入；勝負戰報會跳過。下一步是以正常駐軍／待命流程打滿 16 回合（2 月 15 回合）
後擷取它寫回的
`MEM_WAR.DAT`，不是先做主機程序記憶體傾印。詳見
`docs/playtest/16-dt2-live-battle-snapshot.md` §4 與 `docs/re/05-mem-war-record.md` §5。
相同設定下十次練兵也出現過未觸發江西戰鬥的反例，因此 `train:10` 只能縮短正常
玩家操作，不能作為固定戰鬥 gate；未驗證成功的像素追蹤器已移除。

### 3.2 自傳的後續（研究稿 417/417 完成；整合規格待完成）

剩下的不是「再寫幾篇」，而是產品整合與品質面：

- **`partial` 96/96 已完成機械稽核**（batch01–batch10）；來源鏈、隔離視圖與品質閘
  均已逐批重跑。#16 孫連仲、#21 商震的 C2／C3 重複已人工合併，`--cross` 已歸零
- **Tier A 的 87 筆是下限**，另有 124 筆的一手來源從未量測
- schema 遷移 `note` → `note_fact`／`note_source`（規格 §23.4）
- 制定「成文研究稿 → `people.json` → 發行語系檔」的逐句來源追蹤與可重生流程；
  在 DESIGN-21 解除 DRAFT 禁令前不得手工寫回

⚠️ **同時段只能有一支作業查 gpost**（規格 §26.4）。搶額度會讓別的批次拿到
假的零命中，而假的零命中會變成「這個人查不到」寫進資料——
`unknown` 那 91 筆就是這樣來的。

### 3.3 `.DT1` 剩下的四塊

`docs/formats/07` 有完整佈局。還沒解的是區塊 2 的 `+0/+2/+38`、
勢力表的 `+2`、區塊 8／9／10、檔尾 7 B。

⭐ **查法已經有了**：`tools/addr.py` 現在會把 IDA 線性位址換算成遊戲的 `ds:` 偏移，
`tools/ida_range_xref.idc` 可以對一段範圍逐 byte 查 xref 圖。
下一塊照 §4 那條路走就行。

### 3.4 呈現層與音訊

`internal/ui` 還很薄。音訊那邊 `.MUS`／`.TIM` 已全解、旋律可完整取出
（`tools/mus.py midi`），但 OPL 合成器選型還沒定案。

2026-08-02 政略面板已補上本省剩餘「指令數」：標籤沿用原版 `1.15`／`2.15`
字模，數值直接接 `CommandBudget.Remaining`，並有底部數值區的像素測試。

同日也把已全解的指令 12「商業活動」接進呈現層：`T` 進入進口／出口，
再選品項與輸入數量；畫面全用原版 `2.15` 詞條。Docker Xvfb 已端到端驗證
買進 300 糧食會扣 10 黃金並消耗一個指令數。

指令 3「運補」也已接通：IDA 補證目標只收同司令、未交戰的鄰省，`S` 依序輸入
目標省與黃金／糧食／彈藥／燃料。Docker Xvfb 已用湖北→河南
100／200／300／400 驗證四項扣減與指令數 2→1。下一個呈現層候選應從
規則層已完成的指令中挑，不要回頭重解這兩條。

接著已逐條重讀玩家側四支徵兵函式，修正規則層原先把步兵匯率套給所有兵種
的錯誤：步兵 10 人／金、騎兵 5 人／金、砲兵 1 人／金、裝甲兵 1 人／10 金。
`RecruitLimit` 現在也會取黃金可負擔量與同兵種部隊缺額的較小值。特別注意：
電腦側騎兵是 2 人／金，與玩家側不同，不能共用 `AIRecruitSoldiers` 的表。
指令 5 的徵兵分支現已接到 Ebiten：徵兵／整編第一層、四兵種、數量上限、成本確認
與 Y/N 都有獨立狀態。Docker Xvfb 已實跑湖北徵步兵 1,000：黃金 4,200→4,100、
兵力 97,500→98,500、指令數 2→1。重新整編也已依 `sub_25B2B`／`sub_26D69`／
`sub_27DCA`／`sub_28FF7` 接上剩餘兵力池、四兵種滿員上限與原始候選順序；
步兵／騎兵會加權重分配武裝程度，砲兵／裝甲兵保留原值。規則與畫面測試已通過，
Docker Xvfb 也已走到候選將領與兵力輸入畫面；目前數值區間距偏擠，尚不能宣稱
像素級完成，但操作路由與原版字模載入已驗證。

指令 13「練兵」也已由 `sub_1C916` 補成規則與玩家流程：同省、可行動將領的
士兵戰技增加 `帶兵能力 ÷ 10`，夾到 100；`R` 開確認畫面、`Y` 執行並消耗一個
指令數。同步修正 `buildWorld`，不再把原始 `+16` 非 1 的將領硬標為可行動。
Docker Xvfb 已完成確認、執行、結果回報與另存檔驗證。

指令 6 已恢復原版四項選單。查閱他省可輸入省編號看 13 項資料；查閱所屬各省
依原始省編號列同司令領地與兵力；查閱將領顯示本省可行動將領，Enter 看詳細，
左右鍵切換；原版四套字模含 `FAN(1).15` 都已接上，所以番號可顯示成
「討賊軍第25師」這類組合。士兵攻擊力由 `game.Strength` 即時計算，不另造欄位。
Docker Xvfb 已驗證四項選單、河南省詳細頁、吳佩孚所屬河南／湖北概況、湖北
15 人清單、吳佩孚詳細頁及切換陳家謨。第四項「查閱省名」也已依
`sub_2C12F` 接成 1–20／21–36 兩頁並實測空白鍵換頁。選定他省後的概況／將領
次選單也已補完；河南 25 人清單已實測自動切到第二頁 21–25。指令 6 的已知
操作路徑至此全部可達。

### 3.5 人物自傳已可在遊戲閱讀；P2c 增量資料已接入

2026-08-02 重新盤點，必須把「資料完成」與「玩家可用」分開：

- 人物自傳已有 417/417 槽位事實骨架；30 位 `unknown` 依法不立傳，其餘 387/387 位
  皆有唯一成文研究稿。既有 base `people.json` 的 326 篇 `bio_zh` 保持不動，
  `SPEC-11` 的 `people-authored.json` additive overlay 已把新增 61 篇接入，執行期
  共有 387 篇正文；這不是把研究稿整批覆寫回產品。`internal/ui/textlayout` 已完成
  28×13 格、半形半格、中文標點禁則與分頁，既有 326 篇與新增正文均由同一套排版路徑處理。
  `internal/i18n` 的人物接合、倚天三套字型載入、`render/biography.go` 與
  「查閱將領」B 鍵入口均已完成並實跑。完整、部分、未知三種資料狀態與兩頁翻頁
  都有 Xvfb 截圖證據（`docs/playtest/20-biography-ui.md`）。
- modern UI 已有 `internal/ui/theme`、`-theme retro|modern`、`F2` 與顯示設定頁，
  P0/P1/P2a/P2b 的地形／鐵路／部隊圖示切換已接通；資源／指令 icon、十勢力色、
  向量字型與寬版 HUD 仍未完成。原典／白話的 `wording.json`、失敗即關閉載入器與
  `-wording` 已建立。自傳及指令 1「調動」五階段完成 original／plain 雙路徑；O
  顯示設定與 XDG `prefs.json` 持久化已接，其餘十四項指令尚未接。
  這份設計也仍標 DRAFT；使用者已確認未來移植 Android，並將滑鼠與
  Android 單指點擊納入正式路線。`docs/design/40-pointer-touch-input.md` 規定兩者
  共用動作派送、640×350 座標反算、48dp 觸控目標與分期路線；M0 第一批已實作，
  地圖右上也已有可用滑鼠／觸控叫出選單的 90×48「指令」入口。

因此目前可以說「人物自傳功能已整合 387 篇可立傳正文」；但仍不能說完整傳記資料已
回寫 `people.json`，因為 P2c 是可回復的 additive overlay。也不能把目前 P2b 窄切片誤稱
為完整 modern UI；資源／指令圖示、寬版面與整體 modern 美術仍是後續工作。

2026-08-02 又修正指令 7 的呈現缺陷：`screenDevelop` 原先誤畫完整 15 項主選單，
現已分別顯示原典「墾地／建兵工廠／挖金礦」與白話「開墾土地／建造兵工廠／開採金礦」；
指令 13 練兵確認也接上白話。兩種模式實跑墾地後存檔逐位元組相同，證據見
`docs/playtest/28-wording-develop-train.md`。人工截圖曾抓到把另一個三字緩衝區索引誤套
`3.15` 而顯示「熱河省」，勘誤已保留在程式註解，勿再只看立即數、不追字模來源。

指令 11 秘密行動也已接上白話：選單顯示「派遣游擊隊／鼓動學生發起抗議」，目標頁
把「何省」改成「要在哪個省鼓動學生抗議？」。湖北→河南的固定種子實跑在兩種模式下
皆影響 25 位將領、花費 1,500 黃金、剩餘指令 1，存檔逐位元組相同；見
`docs/playtest/29-wording-covert.md`。游擊隊成本仍缺原版證據，維持明確拒絕，不可臆造。

指令 6 查閱流程的四項入口、省份輸入、河南資料／將領分支、所屬省份概況與省名對照
兩頁也已接上白話；省名字模、編號、兵力及詳細欄位保持原典。兩模式完整唯讀路徑的
輸出均與原始存檔逐位元組相同，十二張畫面人工檢查通過；見
`docs/playtest/30-wording-view.md`。

原版指令 15 的八項已另立 `docs/design/41-other-options.md`。目前只有離開、啟動時載入、
自動存檔及 remake 顯示設定具備底層能力；執行期原子載入也已於 O1 完成。
音效與音樂仍沒有播放後端。不得先畫可切換的 ON／OFF 假裝完成；後續應依
O2–O4 順序建訊息佇列與媒體後端。

O0 已落地：O 先顯示原版八項與第 9 項 remake 顯示設定；1–7 明示「尚未完成」且拒絕
變更，8 接既有存檔離開確認，9 接顯示設定。另修正 F10／第 8 項取消後固定回地圖的
缺陷，現在透過 `quitBack` 回到原畫面；雙模式十二張 GUI 與唯讀存檔證據見
`docs/playtest/31-other-options-o0.md`。

O1 已接通：`buildSession` 在純記憶體中完整解析省份、該期將領、勢力、領袖、
勢力反查與停火，並交叉比對三份勢力索引；全部成功才替換 app 快照。
啟動載入也改走同一條路徑，不再對損壞勢力／停火表只警告後繼續。儲存使用
同目錄暫存檔、`fsync`、原子 `rename`；其他選項 1／2 的 Y／N／ESC 與雙模式 GUI
已實跑，無操作輸出與原始存檔位元組相同。證據見
`docs/playtest/32-other-options-o1-save-load.md`。

O2 已接通。IDA Pro 9.4 證實其他選項第 6 項寫入 `byte_6FE85`，值域
1..10、開局預設 5；各結果分支執行 `DELAY(byte_6FE85 × 0x190)`，因此每級
是 400 ms。remake 現用先進先出畫面訊息佇列保留結果，停留期間像原版一樣擋住
新輸入；`message_time` 以裝置偏好寫入 XDG `prefs.json`，不污染 `.DT1`。
原典／白話 GUI、4 秒擋輸入與存檔位元組等價均已驗證；見
`docs/re/34-message-time.md`、`docs/playtest/33-other-options-o2-messages.md`。

滑鼠／Android 點擊已完成至 M2 的第一個戰場切片。`internal/ui/actions` 定義與裝置無關的動作，
`internal/ui/layout.Grid` 讓 renderer 與命中區共用同一份 placement；Ebiten adapter
同時收滑鼠與 Android Touch ID，放開才派送，拖曳、畫面改變與訊息等待期間均取消。
第一批已接政略選單 15、其他選項 1–9、確定／取消、顯示設定與返回。
十五項政略選單與多個非數字子選單已接共用動作；所有既有 Esc 子畫面已有
48×48 可見返回箭頭，省名／自傳有翻頁動作，將領清單可直接點人名。純滑鼠可由地圖穿過
「指令→徵兵→徵兵→兵種」抵達數量輸入頁。九張第一批 GUI、地圖入口前後畫面及
第二批四張多層路徑均已人工檢查，存檔與鍵盤路徑位元組相同；見
`docs/playtest/34-pointer-m0-options.md`。M1 第一批另加入 6×2、每鍵 64×48 的
共用觸控數字鍵盤，接到九類短目標／數量
畫面。純滑鼠徵步兵 1,000 的存檔與實體鍵盤基準逐位元組相同；見
`docs/playtest/35-pointer-m1-keypad.md`。第二批已讓將領、省份、自治、調動與整編
候選直接點選，調動多選另有條件式勾號送出；純滑鼠驗證見 `docs/playtest/36`。
第三批已完成產能四列直點、高亮與共用鍵盤輸入，鐵礦 20% 存檔與實體鍵盤基準
逐位元組相同；見 `docs/playtest/37`。不可將這誤報為 Android 實機完成：密集清單
48dp、可見戰鬥控制、觸控中斷與 Android 封裝仍待驗證或實作。

同日接上指令 11 的「鼓動學潮」玩家流程：快捷鍵 `V`、兩項子選單、目標省輸入、
固定 1,500 黃金花費、20% 判定與成功效果均已完成，並以 Docker Xvfb 實跑存檔。
游擊隊只缺原版成本公式；在證據補齊前不得填入暫定成本。

指令 1 的玩家流程已另立 `docs/re/33-player-transfer.md`：確認它不是 AI 的八種模式，
而是部分／全部調動，且會連動體力、四種物資、司令、省長與省份旗標。
`internal/game/playertransfer.go` 現已完成唯讀選取工作階段與確認後的原子交易，涵蓋
體力同步、物資上限退回、100 人容量競態、領袖搬遷、自治清除、省長重選與存檔重載。
DOSBox 亦證實 25 人／140,500 兵力完整搬入目標；並發現原版「欲何將留守」沒有提交
非全選集合的可達按鍵。五段玩家 UI 現已接通：模式、目標、選將、四物資與 Y/N 確認，
以 Enter 明確提交作為有標記的現代修正。Docker Xvfb 已實跑湖北 15 人全移河南，湖北
兵力／將領數歸零、指令數 2→1。文件仍為 DRAFT，只缺原版完成前後 `.DT1` byte 差分。

自傳方面，較便宜的 Terra 子代理已完成舊 `partial` 的 batch01–batch10，合計
96/96 人來源鏈與隔離視圖稽核；各批、全體與跨批次品質閘均通過，待稽核數為 0。
仍有 124 筆 Tier A 來源未量測；#16／#21 跨批次重複已完成人工合併與重驗，不能把
機械稽核完成誤寫成重新查證史料完成。

2026-08-02 接著完成自傳 P0/P1 基礎：`tools/gen_people.py` 從研究檔可重現產生
`translations/shared/roster-slots.json`、繁中人物資料與正規化帳本；Go 接合層驗證
三期 486 槽全覆蓋，排除「無省長」，田鎮南兩槽共用一筆，三組跨期異名保留各自
原版寫法但共享 canonical identity。遊戲啟動時已載入 `PeopleDB`，載入失敗會明示並
停用自傳入口。`textlayout` 已讓 326 篇全文通過 28×13 排版。`render/biography.go`
與 B 鍵入口現已接通；Go 倚天載入器另加入 `ASCFONT.15` 與從合法字庫碼位反建的
Unicode 索引，修正 WHATWG Big5 對「偽」選到 `FA66` 的重複碼問題。全部玩家可見
人物欄位掃描後只剩帳本中的 `榘／藁`，會畫可見缺字框並寫診斷訊息。

同日完成用語切換的第一個規則垂直切片：調動方式、目標、選將、四物資與確認頁在
`wording=plain` 下改走完整倚天字庫的語意文字 renderer；原典路徑仍使用原版場景字模，
未改任何像素契約。兩套模式各實跑湖北 15 人全移河南，存檔逐位元組相同且將領數
15/25→0/40。必要語意鍵現在載入時全量檢查，缺一項即拒絕整份 catalog。

用語偏好也已從命令列工具提升為玩家功能：政略選單按 O 進入 remake 顯示設定，1／2
切原典／白話並原子寫入 `prefs.json`；重啟沿用，`-wording` 只做單次優先覆寫。
壞 JSON 或未知模式會整份退回內建預設並明示警告。三次實跑只切設定、未下遊戲指令，
輸出 `.DT1` 全部與輸入 byte-identical。⚠️ 這不是原版指令 15 的完整實作；
顯示設定後續仍應納入完整「其他選項」子選單。勘誤：授權自治屬於指令 8
「政策」，玩家流程已接通，見 `docs/playtest/23-policy-autonomy.md`。
政策第二項「產能分配」也已完成：IDA 證實 `+34/+35/+36/+33`
分別為鐵礦／煤礦／石油／糧食，黃金為剩餘。Docker Xvfb 實跑將湖北鐵礦
25% 調為 20%，輸出存檔僅對應的第 964 byte 改變，見 `docs/playtest/24-policy-production.md`。

---

## 4. 未結的帳

| 事項 | 狀態 |
|---|---|
| **`#16 孫連仲`、`#21 商震` 跨批次重複** | ✅ 已以 C3 為權威記錄合併，保留 C2 的 10 項差額後移除重複；全體與跨批次品質閘通過 |
| **`02-status.md` 的三段分類整張作廢** | 往後判一個人查不查得到，只認 `facts-*.json` 的 `verified_by` 與 `04-confirmed-audit.md` |
| **DOSBox 按鍵回歸** | ✅ 已重驗並固化 smoke gate，見 §3.1、`docs/playtest/15` |
| **`SDFA.EXE` 沒解包** | IDA 自動解包失效（0 函式）。要動態 dump 或自寫解包器 |
| **設計決策待定** | 現代基準畫布、十大勢力配色、OPL 合成器選型 |
| **`.DT2`（39 × 469）** | 結構已解，部隊欄位還有兩個未定名 |

---

## 5. 接手須知

### 開工順序

```sh
cat CONTEXT.md            # 現況、術語、已被推翻的斷言、worklist
git status --short        # 既有改動屬於使用者或前一輪，不要 reset
./tools/go.sh test -count=1 ./...  # 目前 Docker 回歸為 474 個案例全綠
```

### 常用工具

| 想做什麼 | 用什麼 |
|---|---|
| 讀組語，位址翻語意 | `./tools/py.sh tools/dump_func.py sub_XXXXX` |
| 查一個位址是什麼 | `./tools/py.sh tools/addr.py -6221h`（會一併印 `ds:` 偏移）|
| 查誰讀寫某個全域 | `tools/ida.sh raw idat -A "-S/work/tools/ida_xref.idc <符號>" WAR.EXE.i64` |
| 查誰碰某**一段**記憶體 | `tools/ida_range_xref.idc <lo> <hi>`（逐 byte 查 xref 圖）|
| 自傳品質閘 | `./tools/py.sh tools/bio_gate.py <批次>` ／ `--all` ／ `--cross` |
| 發行前掃描 | `bash tools/deny_scan.sh` |

### 三條容易忘的紀律

1. **建置一律走 docker**（Go、IDA、DOSBox、Python 全部）。
   **禁止任何 `docker image/volume/system prune` 或 `rmi`**——這台機器同時放著
   多個客戶專案的 image，2026-07-27 有過刪掉別人 image 的事故。
2. **不要 grep `.asm`。** 它是攤平的文字，沒有交叉參考圖。而且在 16-bit 專案裡，
   **IDA 的線性位址根本不會出現在 `.asm` 文字中**（顯示的是 `segment:offset`），
   拿零命中下「沒人用它」的結論一定錯。
3. **推翻既有斷言之前，先找出當初支持它的證據。**
   `CONTEXT.md` §5 那份「已被推翻的斷言」清單要維護——
   它同時是 `CLAUDE.md` §7.1 那條方法論的驗收計數器。

---

## 6. 這兩天學到、值得帶走的四件事

### 「有值」不等於「有資料」

未初始化的記憶體滿滿都是值，而且看起來很像資料——勢力表殘留槽的領袖 ID
當 u8 讀是 103、12、142，**拿去查名冊還查得到人**。

**判準：拿第二份樣本比對。資料會一致，殘留不會。**
這一輪用它分辨了四個區塊，四次都成立，而且不需要先知道語意。

反面案例就在同一輪：我第一版的 `Active()` 用「領袖 ID 非零」判定，
274 個殘留值裡沒有一個是 0，於是十四個殘留槽全被判成「有勢力」——被自己寫的測試抓到。

### 同一個東西有兩個名字時，加索引沒有用，要加換算

`.DT1` 區塊 2 的語意猜錯五次，**而答案一直在 `internal/game/ceasefire.go` 的註解裡**。
`CLAUDE.md` §7 第 11.5 條要求「寫『語意未解』之前先 grep」——我照做了，零命中，
因為反組譯筆記叫它 `byte_6F532`（IDA 線性位址），程式碼註解叫它 `ds:0B382h`（遊戲偏移）。

已修成機制：`tools/addr.py` 的 `DSEG_BASE = 0x641B0`（取自 IDA 段表，不是推的）+
`lin_to_ds`／`ds_to_lin`。

### 下「不存在」的結論之前先做正對照

「grep `.asm` 找不到位址 → 沒有程式碼碰這塊記憶體」寫進文件又在同一輪撤回。
救回來的是一個問題：**「這個 grep 抓得到任何一個裸位址嗎？」** 答案是零，
才發現查法本身壞掉。結論剛好沒錯，但當時的理由是假的。

### 自動擋掉的代價可能高於放過

品質閘的禁用詞規則兩次逼承辦者**刪掉原件裡的官職全名**
（「剿匪總司令」「晉綏剿匪總預備軍司令官」）。第一次我補了構詞規則，
第二次還是漏——因為只要判定是二元的，就一定有第三個形狀。

修法是分級：**「該擋」那一級只留「怎麼寫都是錯」的詞，
凡是「在某些構詞裡是原件名稱」的詞一律只送複審。**

M2 依原版 `sub_50FF5` 的 32×24、奇數欄下移 12 像素公式加入共用格號反算。
戰鬥中點相鄰空格與數字鍵 1–6 共用 `BattleSim.Move`；點 Enter 原本鎖定的第一個
相鄰敵軍與鍵盤共用 `BattleSim.Engage`。主畫面的單省戰場保持只讀，不發明點格選省。
完整 Go 回歸已通過，GUI 與 Android 實機留待下週；見
`docs/playtest/38-pointer-m2-battlefield.md`。

modern icon 進度必須維持誠實標示：`docs/design/10-visual-modernization.md` 已有三套
圖示模式、資產清單與驗收規格，但 repo 尚無 `internal/ui/theme`、`-theme` 或
`assets/themes/modern` 實作，目前仍只有復古 renderer。白話方面，操作介面已涵蓋
十五項主選單及多數已接功能與 O0–O2；外交、O3／O4、劇情對話／敘事文本仍未完成，
不得把「介面白話大部分完成」寫成「全遊戲對話完成」。

人物自傳整合的前置工具已補上：`tools/gen_authored_bios.py --check` 會從所有成對的
`facts-*.json`／`bios-*.md` 建立記憶體索引，嚴格驗證 417 骨架、387 正文、30 unknown、
#274「無省長」佔位、姓名／信心度、格數、倚天可畫性、禁用詞與來源 SHA-256；並列出
精確 61 筆待整合 ID。`--output /tmp/...` 兩次輸出位元組一致。DESIGN-22 仍是 DRAFT，
所以工具刻意不會改 `docs/reference/people/people.json` 或發行語系檔；下一步是使用者裁決
是否採「獨立來源索引＋產生器合併」，不能把前置驗證誤報為 61 篇已進遊戲。

---

## 7. 2026-08-09 延續工作：`.DT1` 區塊 10 已確認角色

本輪沒有改動原版資料庫；以既有 `ida-pro-9.4-ver3` 容器、IDA Pro `9.4.0.260610`、
`WAR.EXE` SHA-256 `11dbfcf24686ab7765f788b38514cefd2039d0f60b6bd517d89fb5a84c068015`
匯出函式與 direct code xref。`word_70026` 是 10 個 u16 的十大勢力執行期領袖清單：
`sub_35005` 會移除清單成員，`sub_3512B` 清理勢力資料，兩者唯一 direct caller 都是
`sub_353C4`。研究細節與證據契約見 `docs/formats/07-dt1-layout.md` §4e.1、
`CONTEXT.md` §5.43。

Go 端新增 `MajorPowerLeaders`／`ParseMajorPowerLeaders`／`Contains` 與結構測試；
尚未把它替換成 session 的 `MajorPower` callback，因為現有兩份 `.DT1` 快照的區塊 10
值仍可能是場景初始化前／殘留資料，下一步要取得「初始化後正常存檔」對照。

另修正 `tools/ida.sh` 的 EXIT trap：查詢失敗時不再因區域 `query_dir` 離開作用域而報
`unbound variable` 並遮住原始錯誤；新增 `tools/ida_func_xref.idc`，固定保留原名、
IDA 線性位址與 direct caller，供後續查詢沿用。

## 8. 2026-08-09 延續：`.DT2` 寫回契約已閉合

`internal/game/battlestate.go` 新增 `WriteBattleStates`。它要求輸入正好是
`39×469` bytes，從 `ParseBattleStates` 保留的每筆 `Raw` 複製，只覆蓋目前已切出的
欄位；未知欄位與 `+68`／`+268` 檔案緩衝（其 live runtime 同步時機未明）不會因重建結構而被清零。三份現有
`MEM_WAR.DAT`／`SAVE(1).DT2`／`SAVE(2).DT2` 均有 round-trip、單一 u16 差分與錯誤
長度測試，使用 `tools/go.sh test -count=1 ./internal/game`（Docker）通過。

這是 remake 的安全存檔 API，不是原版「回合上限平局後 `sub_3964E` 寫回」的 oracle
證據；下一輪仍以正常玩家完成 16 回合（2 月 15 回合）取得至少兩份寫回後快照為優先，
在那之前不要替兩個 100×u16 區指派新的欄位語意。

## 9. 2026-08-09 戰場 AI 加權尋路切片

`sub_4FCCC` 的 196×196 矩陣與 `sub_5778B`／`sub_5770F` 的 Dijkstra 形狀已由 IDA
Pro 9.4.0.260610（`WAR.EXE` SHA-256
`11dbfcf24686ab7765f788b38514cefd2039d0f60b6bd517d89fb5a84c068015`）核對。矩陣權重
是鐵路 1、河海 30、沙漠 100、一般地形沿 `byte_9E2`、佔用 80、不可達 0xFF。

Go 端新增 `internal/game/battlepath.go` 與 `battlepath_test.go`，`RouteNextCell`
替換 `AutoResolveByChain` 及互動守方 AI 原本的曼哈頓貪心。測試已在 `dsds-go:1.25`
Docker 通過。這只完成矩陣讀端的可重現切片；`sub_567B9` 的目標候選排序、特殊模式
與正常玩家 `+12` 序列仍待 oracle 對照，不能宣稱戰鬥 AI 全等價。

## 10. 2026-08-09 玩家開戰單位識別修正

`cmd/dsds/battle.go` 原本先用 `GeneralsOf` 篩省，再以篩選後的 index 重新產生
`GeneralID`，並把 `Faction` 誤設為省編號。這會在省份將領不是連續槽位時指向錯誤的
`MAN` 記錄，也會讓已確認的執行期 `+14` 效忠勢力欄位失真。現在 `combatants` 直接
掃完整將領表，保留原始 1-based 槽位 ID，並使用來源／目標省的司令填入
`CombatUnit.Faction` 與 `StrengthInput.Faction`；守方領袖也直接沿用同一筆 `+20`。

新增 `cmd/dsds/battle_test.go` 的非連續槽位 `[2,4]` 回歸測試。這只修正 remake 內部
資料一致性，尚未宣稱玩家派將、原版出兵確認鍵或 `.DT2` 未知區塊已對齊；驗證仍須在
Docker 的 Xvfb／Go 工具鏈內完成。

## 11. 2026-08-09 玩家開戰扣款時機修正

`cmd/dsds/main.go` 的 `A` 開戰入口已改成「先成功建立戰鬥，再扣一個指令數」。
原本的順序會在 `startBattle` 失敗時仍減少額度；現在沒有可攻打鄰省、沒有守軍、
部署失敗或其他資料錯誤時，訊息列會顯示原因且額度不變。成功後保留防禦性
`Spend` 檢查，若未來改成非同步而扣款失敗會撤回暫建的戰鬥狀態。

依據是既有實機「查閱後指令數未減」的保守契約（`docs/playtest/02` §282），
不是對原版未知確認鍵的推測；玩家派將流程與 `.DT2` 真正寫回仍未閉合。

## 12. 2026-08-09 `.DT1` 戰爭記錄 +0/+2 已閉合

`sub_3964E` 的 IDA Pro 9.4 匯出確認每省 `ds:0B346h + 省編號×60` 記錄會把
`word_64942` 寫入 `+0`、`word_64944` 寫入 `+2`；兩者的領袖語意已在
`docs/re/31` §37 confirmed。`internal/game.ParseWarRecords` 現在提供 1-based
39 筆 `WarRecord`，以 `SideALeader`／`SideBLeader` 保留中性側別，並保存整列
`Raw`；`WriteWarRecords` 以原始檔案列為基底，只覆蓋這兩個 u16，不碰尚未解的
其餘 56 bytes，即使呼叫端的 `Raw` 被污染也不會清除原始殘留。
`SaveBlocks`、`docs/formats/07` §3 與 writer 測試已同步；整個區塊仍不可整片重建。

## 13. 2026-08-09 `.DT2` +68/+268 runtime 元素定位已閉合

以 IDA Pro `9.4.0.260610`、`WAR.EXE` SHA-256
`11dbfcf24686ab7765f788b38514cefd2039d0f60b6bd517d89fb5a84c068015` 匯出
`sub_4F6FF`、`sub_545B0`、`sub_5446D`。前者將兩側省份參戰資料的非零值分別壓入
runtime `0x6742`／`0x680A`；後兩者以每格值 × `0x21` + `0x7A7D` 索引 runtime
33-byte 將領記錄。`sub_4F468`／`sub_3964E` 則負責把 `MEM_WAR.DAT`／`SAVE(N).DT2`
的 `+68`／`+268` 載入／寫回 `word_6A8F4`／`word_6A9BC` 全域緩衝。故 runtime
清單元素是 1-based 將領 ID、0 是空槽；檔案緩衝與 live 清單的同步時機仍未由正常
戰鬥快照證實，不能把這項定位語意誤報成完整寫回規則。

Go 端新增 `BattleState.AttackerUnitIDs()`／`DefenderUnitIDs()`，只濾 0、不做
劇本人數值域清洗，並保留原始 `UnitsA`／`UnitsB` 與 `Raw` 的 byte-for-byte 寫回
契約。測試以 `0xBEEF` 殘料樣本守住「不可擅自清洗」規則。下一步仍是正常玩家打滿
16 回合（2 月 15 回合）的平局快照，以確認實際寫入與清除時機；不要因這次靜態閉合
就宣稱 `.DT2` 全欄位或玩家戰鬥流程完成。

## 14. 2026-08-09 江西部署提交稽核（未取得 `.DT2` 寫回）

以既有 DOSBox 正常新局時間線追加第 11 次練兵，進入江西（25）「蔣中正 攻擊
孫傳芳」部署畫面；對起始格送 `0 → Y` 後等待 20 秒，畫面回到政略地圖，未取得
可核對的戰鬥選單。這次輸入／畫面結果是 confirmed，部署是否合法與回到政略的控制流
是 unknown，不可當成戰鬥結算。

`SAVE(1).DT2`、`SAVE(2).DT2`、`MEM_WAR.DAT` 仍是既有基準雜湊，沒有 `sub_3964E`
寫回後樣本。下一輪優先找江西部署合法格或改用既有陝西格 195 路徑，完成至少一場
有回合結算的正常戰鬥，再回頭比對兩個 100×u16 區。

## 15. 2026-08-09 停火狀態寫回補強

反組譯已確認談判成功分支 `sub_20E05` 對 `ds:BCA5h` 的逐省 byte 做 `inc`。
`AIWorld.NegotiateCeasefire` 現在保留這個 byte 遞增與溢位行為；長期值是否是
剩餘月數仍標未知。`WriteCeasefireStates` 從原始 `.DT1` 副本只寫區塊 5 的 39
bytes，測試驗證只改指定省的 offset，其他未解 bytes 不動。

這只補上已證實的停火規則與存檔 writer；外交玩家操作仍未接入，正常 16 回合平局
後的 `.DT2` 寫回 oracle 仍是下一個戰鬥證據缺口。

## 16. 2026-08-09 戰爭記錄局部 writer

`WriteWarRecords` 已接上：每列先以 `WarRecord.Raw` 為基底，只寫 IDA 已證實的
`+0`／`+2` 兩個 u16；round-trip 與局部差分測試通過。`+4..+59` 仍是未知欄位，
所以區塊 2 仍不可整片重建，這個 writer 也不會清掉原版殘留。

## 17. 2026-08-09 攻擊輸入重播邊界

陝西省（18）戰例已可用正常玩家按鍵穩定走到部署確認與戰鬥選單。按選單 `2` 後，
IDA `sub_4BF27` 的輸入契約是再讀 `"123456"`；補送 `1` 可離開等待輸入並回到
戰鬥面板。這輪沒有取得可驗證的戰損或回合結算，三份 `.DT2`／`MEM_WAR.DAT`
雜湊前後相同，因此只記錄輸入邊界，不升格為攻擊公式或寫回證據。下一輪若要追
正常戰損，先在 `sub_4BF27` 對照 `sub_4A40D` 的目標清單與目前單位鄰位，再取
攻擊前後快照；不要盲送六個方向鍵。

## 18. 2026-08-09 戰鬥五項主選單已接入，攻擊／撤退未知仍 fail-closed

`cmd/dsds/battle.go` 已把原版已確認的 1 移動、2 攻擊、3 撤退、4 駐軍、5 查閱
接成 `battleMode` 狀態機；只有移動子層讀六角方向 1..6。攻擊子層只記錄已確認的
`sub_4BF27` 第二層 1..6 輸入，不把它誤當目標或直接呼叫 `Engage`；駐軍、查閱與
完整撤退輸入沒有證據時留在可 ESC 回退的提示狀態。Enter／相鄰點擊的直接近身攻擊
仍是 remake 差異捷徑。

`internal/ui/render/battlepanel.go` 新增五項 2.15 詞條選單，`ShowBattleMenu=false`
時不改既有戰鬥面板基準；新增測試確認各詞條有畫出。限定套件與完整 Go 測試均須在
`dsds-go:1.25` + Xvfb Docker 內執行。下一步仍是用正常 DOSBox 路徑閉合攻擊目標／
戰損與撤退省份輸入，取得 `sub_3964E` 寫回後 `.DT2` 差分；不要因面板與狀態機接好
就宣稱玩家戰鬥或 `.DT2` 全流程完成。

## 19. 2026-08-09 江西互動部署的目前格與拒絕樣本

`sub_42566` 進入玩家部署時會從目前單位執行期 `+5` 取游標；江西本輪起點是格 183。
來源省 27 在 `WARPOS.DAT` 的候選格為 `112,126,141,142,154,155,168,169,182,183`。
從 183 以方向 4、5、5,5 分別測到 182、169、155，均可看到游標移動，但 `0 → Y`
沒有取得確認或戰鬥選單，三份 `.DT2`／`MEM_WAR.DAT` 也沒有變更。這是「目前格
起點 confirmed、候選格提交被拒 confirmed、完整門檻 unknown」的正常玩家證據，詳見
`docs/playtest/16-dt2-live-battle-snapshot.md` §4g；不要再把陝西的固定格 0 路徑
套到江西，也不要用這輪替 `.DT2` 戰損寫回背書。

## 20. 2026-08-09 駐軍分支的確認鍵已閉合，效果保持未知

IDA `sub_4D3F6` → `sub_4C33B` 已確認主選單 `4` 之後還要再按 `0`／`Y`／`y`；確認
才會設 `byte_6AA8A = 1` 並呼叫 `sub_55C29(arg_0)`。`sub_55C29` 對目前將領記錄
`0x7A7D + 0x21×arg_0` 的 `+1D`、`+07`、`+1E` 做分段遞增，但欄位語意、旗標的
讀取／清除端與駐軍後回合效果尚未有正常玩家／`.DT2` oracle。這只補輸入契約與
原始寫入證據；`cmd/dsds` 暫不把 `4` 接成狀態變更，避免猜測。

## 21. 2026-08-09 remake 戰損兵力同步接回將領／存檔鏈

`internal/game.BattleSim.SyncForcesToGenerals` 以戰鬥副本的已確認
`Combatant.Strength.Force`，依 1-based `GeneralID` 更新 `[]General.Force`；nil、重複
或越界 ID 會失敗即關閉，且不碰 `General.Raw`、位置、指令、補給、`.DT2` 參戰清單
與其他未知欄位。`cmd/dsds/battle.go` 在攻擊與守方 AI 後設 `forceDirty`，離開戰鬥
前同步；`autosave()` 也走相同入口，避免 F10 存檔遺失已確認的戰損兵力。

這是 remake 內部資料鏈的窄修補，不是原版 `.DT2` 寫回等價宣告。原版 live runtime
清單何時落盤仍需不分勝負打滿回合的 DOSBox oracle；目前測試只確認 1-based 槽位、
錯誤拒絕與「只改 Force」的差分契約。下一輪仍應先取得該 oracle，再決定是否擴充
`.DT2` 的任何寫回欄位。

## 22. 2026-08-09 `sub_534FF` 無主省分支勘誤

本輪以 IDA Pro 9.4 重新讀取 `sub_534FF`，確認 `53588–535CC` 的零司令分支會
直接落到 `+32 & 40h`／`sub_5A881 < 100`，不會 break；非零異勢力才跳到
`5360A` 排除。外層從 1 遞增到 8，掃完整 8 個鄰省槽位。這推翻先前 `ReinforcementSources`
與文件所寫的「遇無主省停止掃描」，但保留歷史勘誤索引（`docs/re/31` §51、§60）。

Go 端已同步修改唯一入口 `internal/game/province.go`，新增
`TestReinforcementSourcesKeepsScanningPastUnownedNeighbour` 固定無主省後繼續與
第 8 格納入；`battlesupport.go` 仍只轉呼叫該入口，沒有第二份規則。證據輸入為
`WAR.EXE` SHA-256
`11dbfcf24686ab7765f788b38514cefd2039d0f60b6bd517d89fb5a84c068015`、IDA Pro
`9.4.0.260610`。

## 23. 2026-08-09 值 4 mode 1 司令候選接回

`sub_3BCED` 的 `0x3BDCF–0x3BE02` 已確認 `mode == 1` 且
`word_64944` 的執行期 `+5 != 0xFF` 時，會把當前交戰省司令追加到候選；值 4 的
`sub_3CA09` 在 `0x3CA9C` 傳入 `mode=1`。`internal/game.BattleSim` 新增
`AtCommander`，`cmd/dsds/startBattle` 從省份 `+20` 填入；`execStrikeForce` 排除
中心格，並在兩圈掃描後追加該司令，不去猜 `sub_55632(mode=0)` 的排序。

回歸：`TestStrikeForceModeOneAddsRemoteCommanderAndSkipsCenter`、
`TestStrikeForcePoolExcludesOwnUnits`。證據輸入為同一份 `WAR.EXE` SHA-256
`11dbfcf24686ab7765f788b38514cefd2039d0f60b6bd517d89fb5a84c068015`、IDA Pro
`9.4.0.260610`；完整 `sub_55CEC` 呼叫端仍未全解。

## 24. 2026-08-09 `sub_3D261` 城市後備部分接入

本輪在既有 IDA Pro 9.4 工具鏈重讀 `sub_3D261`、`sub_560D7`、`sub_3D57B` 與
`sub_3D411`。已確認：`sub_560D7(mode=1/2)` 依城市清單順序回第一個相鄰／兩格內
城市；值 13 的 `sub_3D57B` 只在兩格內守方城市的尋路失敗時，於 `0x3D7A0` 呼叫
`sub_3D261`。`sub_3D261` 先處理第一個相鄰城市，再掃同 `+14` 勢力的佔用城市，
且直接攻方分支不立 `+13 bit 7`、直接守方分支不清舊 `+10`，同勢力掃描失敗會
留下最後候選欄位。

Go 端新增 `internal/game/battlefallback.go` 的 `assignCityFallback`，並接到
`internal/game/battleexec.go` 值 13 命令 2／3；`sub_3D411` 的 `0x3D543` 後處理與
`byte[65BA]` 預約表仍未接。新增四個 `battleexec_test.go` 測試固定欄位契約與
callpoint 門檻。證據與未解邊界見 `docs/re/31-battle-ai-chain.md` §62、
`docs/mechanics/70-ai.md` §6p、`CONTEXT.md` §5.61。

## 25. 2026-08-09 `sub_3D411` 後處理窄接入

本輪在既有 IDA Pro 9.4 工具鏈重讀 `sub_3D411` 與 `sub_55CEC`。已確認：第二方
命令 4／5 先做六格敵鄰檢查；沒有敵鄰時依傳入城市清單順序找空且未被
`byte[65BAh+格]` 預約的城市，寫 `+12`／預約表；找不到時把命令改成 2，於
`0x3D543` 呼叫 `sub_3D261`。`sub_3D57B` 尾端 gate 的 bit 7／回合門檻已證實，
但 `byte_6B89E` 仍只知道由 `sub_39B6E` `arg_A` 寫入，語意未定。

Go 端新增 `internal/game/battlepost.go` 的 `execDefaultPost`，並在
`BattleChainGates.EnableDefaultPostStage` 明確開啟時接入 `AutoResolveByChain`；
預設不啟用，避免把尚未閉合的 gate 或預約表生命週期當成原版規則。三個新測試
固定敵鄰跳過、城市預約／順序、命令降級與 `sub_3D261` 後備欄位。完整 gate、
預約表清除時機、`sub_567B9` 排序與正常玩家 `.DT2` 差分仍是下一輪工作。

## 26. 2026-08-09 `sub_567B9` 候選排序資料流補證

本輪沒有把 `sub_567B9` 的排序硬接進 Go，而是在 IDA Pro 9.4 追完其可證實的
前置與輔助支線：六格候選會排除佔用／預約與超過 `sub_5619C` 成本上限的格；
`sub_56461` 依 `sub_503BB` 防禦值降冪、`sub_56548` 依目標格曼哈頓距離升冪，
`sub_566B6` 再按難度／特殊旗標選結果。`sub_581C0` 的特殊省集合、
`sub_562BF` 的候選值與 `sub_56729` 的條件式全矩陣尋路仍未取得正常玩家 oracle。

因此 `internal/game/battlepath.go` 仍只代表已驗證的矩陣讀端；不要把它宣稱成
完整 `sub_567B9`。證據與位址見 `docs/re/31-battle-ai-chain.md` §64、
`docs/mechanics/70-ai.md` §6r、`CONTEXT.md` §5.63。

## 27. 2026-08-09 `byte_6B89E` gate 的 `arg_A` 值已閉合

本輪用 IDA Pro 9.4.0.260610（`WAR.EXE` SHA-256
`11dbfcf24686ab7765f788b38514cefd2039d0f60b6bd517d89fb5a84c068015`）匯出
`sub_39B6E` direct code xref 與相關函式原始指令。`sub_39B6E+0x39CBB–0x39CBE`
把 `arg_A` 寫入 `byte_6B89E`；四個直接 callsite 的值已固定為：
`PROGRAM+0x110D3=0`、`sub_2D812+0x2D9F3=0`、`sub_3562B+0x357D4=1`、
`sub_368A8+0x36977=0`。這只確認原始 byte 的來源值，不替 `arg_A` 命名成
玩家／電腦模式。

Go 端 `BattleChainGates` 新增 `DefaultPostBit7`／`DefaultPostArgA`，
`DefaultPostStageOpen(turn)` 實作原版 gate；`EnableDefaultPostStage` 保留為
已掌握完整外部狀態的相容覆寫。新增 `TestDefaultPostStageOpenMatchesOriginalGate`，
驗證回合 4／5、`arg_A` 零／非零及覆寫。這輪沒有把 gate 自動推送到互動戰鬥，因為
`byte[65BAh+格]` 預約表生命週期、`arg_A` 高階語意、`sub_567B9` 排序與正常玩家
`.DT2` oracle 仍未閉合；證據與限制見 `docs/re/31` §65、`docs/mechanics/70-ai.md` §6s、
`CONTEXT.md` §5.64。

## 28. 2026-08-09 `sub_567B9` 候選排序原語已落地，勿誤接成完整 AI

本輪以同一份 `WAR.EXE`（SHA-256
`11dbfcf24686ab7765f788b38514cefd2039d0f60b6bd517d89fb5a84c068015`）及 IDA Pro
9.4.0.260610 匯出 `sub_567B9` 的四個直接輔助函式。`sub_58FD9`
（`0x58FE7–0x5900F`）只在地物碼 3 且 `sub_4FEF0` 無鐵路時回 1；`sub_56461`
（`0x56461–0x56545`）做 `sub_503BB` 防禦值降冪；`sub_56548`
（`0x56548–0x566B3`）做目標格矩形曼哈頓距離升冪，無鐵路河海後項不前移；
`sub_566B6`（`0x566B6–0x56726`）依特殊旗標／防禦值平手選兩個 stack 結果。
直接 caller 是 `sub_567B9` 的 `0x5698A`、`0x569A1`、`0x569A5`。

Go 端新增 `internal/game/battlecandidates.go` 與測試，僅落地上述純原語；未把
`[arg_0-30h]`／`[arg_0-36h]` 命名成高階排序，也未改 `RouteNextCell`。完整候選接合、
`sub_562BF`／`sub_56729`、預約表生命週期與正常玩家 `.DT2` 差分仍是下一步，現有
AI 仍採 fail-closed。

## 29. 2026-08-09 `sub_567B9` stack 方向勘誤與純候選切片

本輪用同一份 `WAR.EXE`（SHA-256
`11dbfcf24686ab7765f788b38514cefd2039d0f60b6bd517d89fb5a84c068015`）及 IDA Pro
`9.4.0.260610` 重新核對 `sub_503BB` 與 `sub_567B9` 的呼叫慣例。`sub_503BB` 的
physical `arg_0` 是最後壓入的格、`arg_2` 是先壓入的 runtime 單位 ID；三參數
caller 的 push 順序是 `(current, target, mode)`，故 `sub_567B9` 邏輯參數為
`(mode, target, current)`。`0x56844` 起的六鄰掃描使用 target，不是 current；舊
§64／§6r 的「目前格周圍六格」只保留作歷史，最新勘誤見 `docs/re/31` §67。

新增 `internal/game/battlecandidatepath.go`／`battlecandidatepath_test.go`：
`SelectOriginalBattleCandidate` 純粹落地 target 六鄰候選、佔用／注入預約過濾、
`sub_5619C` 的 12／13 成本門檻、兩份排序原語與最後相鄰回退。開啟
`EnableLastSteps` 且不在特殊省份時，`sub_562BF`／`sub_56729` 與預約生命週期未解，
結果會 `Complete=false`；沒有候選的已知分支則 `NoCell, Complete=true`。方法不改
`Occ`／`NextCell`，也沒有接入既有 AI 派工，`RouteNextCell` 保持矩陣讀端實作。

回歸：`./tools/go.sh test -count=1 ./internal/game` 已在 Docker 通過。下一輪優先
取得正常玩家 `+12` 序列與 `.DT2` 差分，找出真正的 `+12` 消費端，再決定是否能把
純選擇器接入某一個明確 callpoint；不要因這個切片通過測試就宣稱完整 AI 等價。

## 30. 2026-08-09 `+12` 消費端稽核：維持 fail-closed

本輪依 `docs/re/31-battle-ai-chain.md` §68，以 IDA Pro 9.4.0.260610（同一份
`WAR.EXE` SHA-256 `11dbfcf24686ab7765f788b38514cefd2039d0f60b6bd517d89fb5a84c068015`）
追查 `+12 = 0x7A89` 的讀寫與 direct xref。`sub_4A1C0+0x4A267` 寫 `+5` 的唯一
direct caller 是玩家移動輸入；`sub_50992` 讀 `+12` 只作畫面動畫；`sub_58449`
暫寫 `+5` 只作交戰動畫。`sub_3F0EF`、`sub_3C777`、`sub_3B492`、`sub_4732C`
的 `+12` 讀取目前只閉合到未指派檢查、重複目標過濾與清理，尚未找到正常 AI
把 `+12` 落成 `+5` 的完整 call chain。

這是可重現的「尚未找到」負證據，不是對間接函式指標的不存在宣告。後續需固定
電腦回合擷取 `+12`、`+5`、`word_62A8` 與 `.DT2` 差分，或沿新的間接 caller
追查；在此之前不要把 `RouteNextCell` 或 `SelectOriginalBattleCandidate`
接成實際 AI 移動，也不要宣稱完整流程等價。

## 31. 2026-08-09 自動守方部署改接 NWMAP `0x4000`

本輪用同一份 `WAR.EXE`（SHA-256
`11dbfcf24686ab7765f788b38514cefd2039d0f60b6bd517d89fb5a84c068015`）及 IDA Pro
`9.4.0.260610` 匯出 `sub_4166E`、`sub_41513`、`sub_4180D` 與
`sub_55CEC` xref。`sub_4180D+0x419A5` 的自動守方先以
`word[0x796h + 格×2] & 4000h` 由格 0→195 掃候選；第一輪加上
`sub_55CEC(mode=1)` 拒絕條件，失敗才放寬。稍後 `sub_4180D+0x41AF3` 才以
攻方 `+8=1` 呼叫 `sub_41513`，按 WARPOS 來源省 195→0 部署。

Go 端新增 `Map.DefenderDeployZone`／`DefenderDeployFlag`，並把
`cmd/dsds/startBattle` 原先以 `WARPOS==0` 腹地擺守軍的 remake 假說改為
NWMAP `0x4000` 候選。`internal/game/deploy_test.go` 逐省固定候選非空、旗標
命中及掃描順序；`0x1000`／`0x2000`／`0x8000` 與 `sub_55CEC` 完整門檻仍未知。

## 32. 2026-08-09 `.DT1` 戰爭記錄的兩個玩家派將 layout 已窄接解析

本輪以同一份 `WAR.EXE`（SHA-256
`11dbfcf24686ab7765f788b38514cefd2039d0f60b6bd517d89fb5a84c068015`）及 IDA Pro
`9.4.0.260610` 追完玩家派將支線：`sub_2DD1F+0x2DD57–0x2DD68` 將選中的
1-based 將領 ID 複製到 `[bp-1BAh]`；`sub_2D812` 第一分支把四種資源寫到
`+4/+6/+12/+16`、將領清單寫到 `+18`，第二分支則把資源寫到 `+8/+10/+14/+18`、
將領清單寫到 `+38`。這兩段不是 `ds:A358h` 參戰部隊表的同名偏移，且 `+18`
是兩個 layout 共用的 union，位址空間與分支必須分開。

Go `WarRecord` 新增 `Branch4`／`Branch8`，各自解析四種資源與中性將領槽位；
`ParseWarRecords` 不把兩個 union 分支同時視為有效，`WriteWarRecords` 仍只覆蓋
已證實的 `+0/+2`，保留 `+58..+59` 與同步時機未知的原始 bytes。新增測試固定
資源映射、union、首尾槽位、1-based 記錄與既有 byte-for-byte 保留契約。證據與
限制見 `docs/re/31` §70、`docs/formats/07` §3。

## 33. 2026-08-09 戰鬥滑鼠／Android 觸控控制外殼已接入

新增 READY 規格 `docs/spec/05-battle-pointer-controls.md`。右側五項已由 DOSBox／IDA
確認的戰鬥命令現在各有裝置無關動作 `battle.command.1..5`；renderer 與指標命中
共用 `internal/ui/layout.BattleCommandButton`，不把點擊當成另一套規則。

面板底部另畫三個 56×48 邏輯像素的 remake 大按鍵：攻擊沿用既有
`battle.attack`（Enter／相鄰第一個敵軍捷徑）、換部隊等同 Tab、結束回合等同 Space。
`cmd/dsds/updateBattle` 只讀這些動作並走原有鍵盤分支；沒有把原版尚未閉合的
攻擊六方式、駐軍效果、撤退省份或 `.DT2` 寫回時機臆測成已完成。

`internal/ui/actions`、`internal/ui/layout`、`cmd/dsds/pointer.go`、
`internal/ui/render/battlepanel.go` 均有單元／像素護欄；限定 Docker/Xvfb 測試已通過。
下一步是補多目標選取與 Android 實機／封裝驗證，並繼續取得正常 16 回合平局的
`.DT2` oracle。原版資料檔仍不得由這些 remake 控制按鍵直接改寫。

## 34. 2026-08-09 外交帳本與外援基準窄接規則層

本輪用同一份 `WAR.EXE`（SHA-256
`11dbfcf24686ab7765f788b38514cefd2039d0f60b6bd517d89fb5a84c068015`）及 IDA Pro
`9.4.0.260610` 核對 `sub_215F0`、`sub_2164A`、`sub_21D1D`、`sub_223ED`、
`sub_391E1`。`sub_215F0` 以 `word_70026` 反查得到 1..10 勢力槽位；
`[di-4225h]` 是信用度表（index 1：ds:BDDB+1／IDA `6FF8Ch`，初值 100），
`[di-421Eh]`／`[di-421Ch]` 是 10×u32 little-endian 外債表（index 1：
ds:BDE2+4／IDA `6FF96h`，初值 0）。貸款核准時 `add/adc` 累加外債，償還時
`sub/sbb` 減債；信用度每 500 黃金扣／回補 1，最高 100。`sub_21D1D` 的
四個外援基準也確認為 `60000 − 目前的 Gold/Food/Ammo/Fuel`。

Go 端新增 `internal/game/diplomacy_ledger.go` 與測試，`DiplomacyLedger` 只在
規則層同步核准貸款／償還的信用度、外債與省份黃金；`AidResourceBases` 提供
同一個外援基準。新增 READY 規格 `docs/spec/06-diplomacy-ledger.md`，並在
`tools/addr.py` 加入負位移索引，保留原始運算元、位址空間、推論等級與出處。

未完成且刻意未接：玩家外交 UI、信用度不足的原版 gate、`.DT2` 持久化與
Android 外殼。`.DT1` 區塊 8／9 的帳本持久化已由 IDA 讀寫端閉合，仍不可把規則層
單元測試當成完整外交玩家流程等價。

## 35. 2026-08-09 直接 `WAR.EXE` 開局勘誤與練兵垂直切片

本輪新增的最重要交接不是新規則，而是把「載入」與「重新開始」分開。直接執行
`WAR.EXE` 的主選單上，`Return` 不會選 `1.重新開始`；必須先按 `1` 再按
`Return`。舊時間線 `Return → 2 → Return` 實際會進 `2.載入遊戲`，所以此前落在
政略／調動子選單的畫面不能拿來證明新局設定。完整勘誤與可重播的新局序列在
`docs/playtest/13-new-game-settings.md` §7，`docs/playtest/16-dt2-live-battle-snapshot.md`
§4g 已保留舊證據並標出被推翻的操作斷言。

校正後的新局曾正常進入「山西省（16）／日本攻擊中國／放置閻錫山的軍隊」；
約 2 秒的「進入戰場中…」到約 4 秒的部署畫面有 AI／亂數分岔，尚未形成固定
部署 oracle，也沒有新增 `MEM_WAR.DAT` 寫回樣本。

可穩定重播的載入進度一垂直切片是：

```text
2 → Return → 1 → Return → 等電腦回合 → 1 → 3 → Return → y
```

玩家主選單的河南指令數由 4 降為 3，確認畫面「司令欲練兵嗎？（Y/N）」與
`tools/dosbox_runner.sh` 的四鍵契約。最新截圖雜湊：
`train-prompt.png = 68cf5f2e032a4a1477d946245df277054c63e5c4f224d8cc4a48886ae6282274`、
`train-after.png = 84719dbd87bb1f312f04a4854292ce1232b3efc49ec9a3b658def67c26af73f9`；同輪
`SAVE(1).DT2`／`MEM_WAR.DAT` 維持 `3fbe8f4fe12f4bbd9f363f8775c87fe4a27109f1c37f79207b31f6a54a4563c1`／
`994e2f71adbb014a042120490855d7b9dfecd55bc0569087b59f4adc059f86a9`。
這輪只有 oracle／文件更新，沒有宣稱平局、戰損、Android 或外交 UI 完成。

### 36. 2026-08-09 連續練兵的剩餘指令門檻

在同一個 `SAVE(1)` 載入路徑以 `tools/dosbox_runner.sh` 執行 `train:10`，每組
仍逐一送出 `1 → 3 → Return → y` 並等待 5 秒。最後畫面位於河南省政略主選單，
指令數由 4 降為 2；這表示只有兩次練兵被目前省份／剩餘指令條件接受，不能把送出
十組按鍵當成十次已執行。截圖 `after-train.png` SHA-256 為
`f017defb34403f49795df20afe037df517d1e343af708372c57df4103f38db7a`。

同輪 `SAVE(1).DT1`、`SAVE(1).DT2`、`MEM_WAR.DAT`、`SAVE(2).*`、`CONFIG.DAT`、
`PLACD.SAV` 都沒有相對載入副本的變化；`SAVE(1).DT2` 與 `MEM_WAR.DAT` 仍為
`3fbe8f4fe12f4bbd9f363f8775c87fe4a27109f1c37f79207b31f6a54a4563c1`、
`994e2f71adbb014a042120490855d7b9dfecd55bc0569087b59f4adc059f86a9`。這只增加
政略指令的正常玩家觀察量，不足以閉合戰鬥結算或存檔同步。

### 37. 2026-08-09 `.DT1` 外交帳本區塊 8／9 讀寫閉合

以 `WAR.EXE` SHA-256
`11dbfcf24686ab7765f788b38514cefd2039d0f60b6bd517d89fb5a84c068015`、IDA Pro
`9.4.0.260610` 匯出 `sub_595D4`（`595D4h..598C0h`）與 `sub_59CBF`
（`59CBFh..5A031h`）。`sub_595D4` 將 `byte_6FF8C` 的 0Ah bytes 與
`byte_6FF96` 的 28h bytes 放入 `.DT1` 暫存記錄後整筆寫出；`sub_59CBF` 在
載入時以相同大小讀回。這是直接的讀寫端證據，不是由快照數值猜測。

因此區塊 8（offset `14606..14645`）是 10×little-endian `u32` 外債，區塊 9
（offset `14646..14655`）是 10×`u8` 信用度。新增
`internal/game.ParseDiplomacyLedger`／`WriteDiplomacyLedger`，並把
`cmd/dsds` 的 `buildSession`／`applySession`／`autosave` 接上同一份帳本；只改
這 50 bytes，區塊 10、檔尾 7 B 與未知區域原樣保留。真實 `SAVE(1)`／`SAVE(2)`
錨點、round-trip 與未授權 offset 測試見 `docs/spec/06-diplomacy-ledger.md`、
`docs/formats/07-dt1-layout.md`。

這項進度只完成帳本存檔同步；玩家外交畫面、信用度不足 gate、區塊 10 同步時機、
Android 外殼與 `.DT2` 戰鬥寫回仍未完成。

### 38. 2026-08-10 人物自傳滑鼠／觸控與發行預覽交接

將領詳細頁現在畫出「人物自傳／人物生平」可見按鈕（144×48），滑鼠／Android 觸控
命中 `actions.OpenBiography` 後與 `B` 鍵共用 `openBiography`；`ESC`／左箭頭仍只返回。
命中區由 `internal/ui/layout.BiographyButton` 同時供 renderer 與 pointer layer 使用，
缺語系／字庫時不留下隱形 target。Docker/Xvfb 只跑相關套件並通過：
`cmd/dsds`、`internal/ui/layout`、`internal/ui/render`、`internal/ui/actions`。

推廣預覽已放入 `docs/promo/great-era-remake-teaser.mp4`：15 秒、98,816 bytes、
SHA-256 `36d42df8d464fe1d6a8cd65fbaaa043ee239eb8d0f7e5ed40e79c6f4c4c3b9c2`；是自製
文字卡／抽象背景，無原版素材、音樂，並明示「開發預覽」。`tools/package.sh` 已建立
三平台（Linux amd64／Windows amd64／macOS arm64）包裝入口，但目前 image 缺
MinGW／osxcross，且 modern tileset、外交／停火仍未完成，不能宣稱正式三平台發行。

下一輪若要達成正式包，先完成 `docs/design/10-visual-modernization.md` 的 READY
切片與可切換 tileset，再閉合外交／停火正常玩家流程，最後提供含三套 CGO 交叉工具鏈
的可重現 release image；完整測試仍可依使用者要求縮小，但發行前至少要跑目標平台建置、
資產拒絕掃描與 dirty-tree／Docker 清理檢查。

### 39. 2026-08-10 外交 UI M0／no-cgo 交接

原版 DOSBox 正常路徑已確認指令 9 三列外交選單、貸款 `0-5000`、信用度 100；送出
1000 後固定樣本指令數 4→3。Remake 現在把 `screenDiplomacy`、`screenLoanAmount`
接到 `DiplomacyLedger` 與 autosave 區塊 8／9；原典／白話文字、滑鼠／Android 觸控
命中區、數字鍵盤共用同一套 `actions`。外援／償債、拒絕成本與信用度 gate 仍未完成。
規格／雜湊／限制見 `docs/spec/07-diplomacy-ui.md`、`docs/playtest/35-diplomacy-ui-m0.md`。

使用者所說的「不使用 cgo」已分成兩層記錄：遊戲與規則層是純 Go，沒有 `import "C"`
或 `#cgo`；目前 Ebiten v2.8.8 的 Linux desktop GLFW 與 macOS desktop backend
仍需要 cgo。已驗證 Windows amd64 可 `CGO_ENABLED=0`，Darwin arm64 在 `CGO=0` 會
缺 `glfw.Window` 等符號，因此 `tools/package.sh` 採 Linux `CGO=1`、Windows `CGO=0`、
Darwin `CGO=1`，不把後端需求誤寫成專案使用 cgo。若未來必須三平台全純 Go，先建立
明確版本的 Ebiten backend 遷移規格與驗證，再改依賴；不要在腳本強行忽略 linker／GLFW。

### 40. 2026-08-10 modern 戰場主題 P0／P1 已接通

READY 切片為 `docs/spec/08-modern-theme-p1.md`。`internal/ui/theme` 提供不依賴
Ebiten 的窄 `Theme`；retro adapter 包住既有 `TileSet`，`render.DrawThemedBattlefield`
保留唯一六角幾何／地物索引／鐵路疊圖／像素 0 透明／邊界框。`cmd/dsds` 與
`cmd/screenshot` 支援 `-theme retro|modern`，遊戲中 `F2` 可即時切換，偏好只進
設定檔，不進 `.DT1`／`.DT2`。

`NewModern` 以純 Go deterministic provider 生成 22 張 32×24 地形與 21 張 32×24
鐵路；沒有嵌入原版像素或 PNG。已按 `RAIL.TPC` 校正 index 0 是有效縱線，像素 0
才是透明；modern rail 接線表保留原版 0..20 的直線、彎角、T 字與十字拓撲。

Docker 驗證已完成：

- `go test ./internal/ui/theme ./internal/ui/render`；`go test ./cmd/dsds ./cmd/screenshot`。
- 固定 `NWMAP.DAT`／`WARPOS.DAT` 的 modern screenshot 39/39 省；湖北 26
  retro `4aa9f2f3ec0e78144663e0f0d7194d356e48c1145c4f88a91a734be7c0c85298`，
  modern `24936f68e87d8202150dd715444ff11d250e92950ac1cb156fb3c7105ace025b`；
  modern 省 01／39 分別為
  `2c24153122a6b21536e3b4ab217fc62bddeef38dece0c938428c5abd7bbc23ab`／
  `d09b0aca8e2306452c165ae07c613219ff8ee9052404a2919da363282973e03f`。

這不是完整 modern UI。部隊／資源／指令 icon、向量字型、wide layout、可發現的
設定選單，以及「現代重繪 PNG／SVG」美術規格仍待另立 READY；目前 P1 刻意使用程式
生成資產以避免原版衍生物與 deny-list 白名單風險。後續若要做真正高解析美術，先補
來源／授權／索引審查，再替換 provider，不要直接把 PNG 放進 repo。

### 41. 2026-08-10 modern 部隊圖示 P2a 已接通

新增 READY 規格 `docs/spec/09-modern-unit-icons-p2.md`。`theme.UnitProvider` 以
`NEWICON.TPC` 相同的 0..17 索引提供部隊圖示；`render.BranchIcon` 仍是唯一兵種／
攻守／砲兵朝向對照，避免 theme package 複製規則。retro provider 直接包玩家自備
原版圖示，modern provider 以純 Go 生成 18 張 32×17 圖，像素 0 透明，砲兵 1..6
朝向各有不同像素。

`cmd/dsds` 的 `F2` 現在同步切換地形／鐵路與部隊圖示 provider；戰鬥 live path、
`cmd/screenshot -units`、`-battle` 都走同一組主題資產。modern 只保留原版兩方的
jade／vermilion 色意義，尚未做十大勢力色、symbol 軸、資源／指令 icon 或完整設定
選單，不能宣稱 P2 全部或完整 modern UI。

Docker 驗證：`go test ./internal/ui/theme ./internal/ui/render ./cmd/screenshot` 通過；
`cmd/screenshot -game workplace/orig/game -province 26 -units` 的 retro／modern
輸出雜湊分別為 `6864274a1cadd25b022d9eefde53136f325d9a2d771eb90fcedaeaefe623c490`／
`2fa3e30ccccfd79efd4851f14169f535b3a776ff9f67145480ada2e0f182a7ac`。`cmd/dsds` 需
在 Xvfb 下驗證；本輪直接無 DISPLAY 執行只得到 Ebiten 預期的 GLFW 錯誤，未把它當成
測試通過。

同一份固定存檔以 `-battle -theme retro|modern` 跑正常 `BattleSim` 部署，攻守各
10 個單位均成功畫出；湖北 26 的 retro／modern 雜湊為
`4a5c63d17aac0ce0cd7a67c072bd4f864f158c7602544b4ad0f0ee77d0a5ebb4`／
`edaa8ff50b43dec74db92fd8c1bc4cc11a70601fadbd46f37c77b7b558869be5`。

### 42. 2026-08-10 modern 顯示設定 P2b 已接通

新增 READY 規格 `docs/spec/10-theme-preferences-p2b.md`。既有「其他選項 → 顯示設定」
現在保留用語 1／2，並新增並排的圖形主題 3／4；原典／白話標籤分別是「原版圖形／
經典原味」與「現代圖形／現代清晰版」。鍵盤、滑鼠與觸控都轉成同一組
`actions.Select1..Select4`，幾何由 `internal/ui/layout.DisplayWordingOption`／
`DisplayThemeOption` 同時提供 renderer 與 pointer layer，返回區不重疊。

`cmd/dsds.setThemePreference` 現在是 F2 與設定頁共用的 fail-closed 入口：先檢查目標
`Theme` 與 `UnitProvider`，再原子保存 XDG `prefs.json`，最後同步交換地形／鐵路／部隊
圖示 provider。`internal/prefs` 也會拒絕未知主題；主題偏好不進遊戲存檔、不改規則／
亂數／戰鬥狀態。新增三個雙模式 `settings.theme*` 語意鍵。

Docker 驗證：`internal/ui/layout`、`internal/ui/render`、`internal/i18n`、`internal/prefs`
目標測試通過；`cmd/dsds` 在 Xvfb 下通過。這只完成 P2b 可發現性窄切片，資源／指令
icon、十勢力色、向量字型、寬版面、戰鬥中獨立設定與三平台正式包仍未完成。

### 43. 2026-08-10 純 Go／Ebiten cgo 邊界固定檢查

使用者重申遊戲以 Go／Ebiten 重寫，不在專案原始碼使用 cgo。新增
`tools/check_no_cgo.sh`，只掃 `cmd/`／`internal/`，Docker 執行結果為沒有
`import "C"` 或 `#cgo`。Ebiten v2.8.8 module cache 不列入掃描；Linux／macOS
桌面 GLFW backend 的 cgo 是外部依賴限制，不是 remake 自己的 C 程式碼。

`tools/package.sh` 不改成三平台 `CGO_ENABLED=0`：目前 Linux／Darwin desktop
在該旗標下會因 Ebiten 的 `glfw.Window` 符號缺失而失敗，Windows amd64 才已驗證可用
`CGO_ENABLED=0`。若未來要求三平台都不帶 cgo，先升級／替換 backend 並建立 READY
遷移規格與目標平台驗證，再修改封裝矩陣。

### 44. 2026-08-10 人物自撰傳記 additive overlay P2c

使用者要求人物自傳串入遊戲；不覆寫既有 base `people.json` 的 326 篇，新增 READY
規格 `docs/spec/11-authored-biography-overlay-p2c.md`。`tools/gen_authored_bios.py
--overlay` 從研究批次只輸出執行期目前空白的 61 篇，產物為
`translations/zh-Hant/people-authored.json`（60,065 bytes，SHA-256
`6b4077833913f8759c8bdfa17818208a47e6b45d7c6eba2c4bdb11c1aee2518b`）；overlay 每筆
保留 facts／bios／正文 SHA-256 provenance，連跑兩次位元組相同。

`internal/i18n.LoadPeople` 先載入 base，再以 fail-closed 規則套用同 locale 的
`people-authored.json`：檔案不存在仍相容舊包；存在但 schema、語系、排序／ID、姓名、
正文、信心度、provenance 不符，或嘗試覆蓋既有正文，整份載入失敗。成功後只更新
`Biography`／`Confidence`，#274 與 30 位 `unknown` 仍無正文。Docker 已通過
`python3 -m unittest tools/test_gen_authored_bios.py`、`go test ./internal/i18n`、
`go test ./internal/ui/textlayout ./internal/ui/render`；完整 387 篇回寫 `people.json`
仍屬 DESIGN-22 DRAFT，沒有被本切片解除。

### 45. 2026-08-10 外交玩家 UI M1：外援／償還外債

新增 `docs/spec/12-diplomacy-ui-m1.md`（READY）。外交第 2 項已接到
`AIWorld.RequestAid`：援助基準、70% 判定、張作霖禁運與黃金夾值仍由規則層唯一實作，
核准才扣本月指令；第 1 期援助國呼叫端尚未閉合，代碼 99 是規格明示的 remake 政策，
不是原版國家對應結論。

外交第 3 項已新增 `screenRepayAmount`。貸款原有頁面不改，償還頁以同一組數字鍵盤、
pointer／touch targets 輸入金額；`DiplomacyLedger.RepayDebt` 在一次提交中驗證外債／
黃金、扣款、回補信用度與減少 32-bit 外債，成功才扣指令。新增語系鍵涵蓋兩種用語與
結果；存檔仍沿用 `.DT1` 區塊 8／9 autosave，未解 bytes 不動。

Docker 驗證已跑 `internal/game`、`internal/i18n`、`internal/ui/render`、
`internal/ui/actions`、`internal/ui/layout` 與 Xvfb 下 `cmd/dsds`。仍未完成：信用度不足
貸款 gate、拒絕成本、第一期援助國原版選擇、Android 封裝與完整外交／停火流程。

### 46. 2026-08-10 談判停火玩家 UI M1

新增 `docs/spec/13-ceasefire-ui-m1.md`（READY）。本輪依 `WAR.EXE` 的 IDA Pro 9.4
呼叫端證據接上指令 10：stage 1 的輸入上限是 36 省；司令必須在目前省，目標省必須
有司令且目前有戰事。無法使用、司令不在、目標無司令或沒有戰事時，不立原版
`byte_6FE81`，因此不消耗指令；同意與拒絕都立該 byte，回主迴圈各消耗一次指令，
同意另遞增 `ds:BCA5h` 的停火狀態 byte。原始證據仍保留 `WAR.EXE` SHA-256
`11dbfcf24686ab7765f788b38514cefd2039d0f60b6bd517d89fb5a84c068015`、IDA 線性位址與
`ds:` 位址分開記錄。

程式接合：`internal/game/ceasefire.go` 新增 stage／36 省／司令與戰事 gate；
`CombatUnit.Deployed` 對應執行期記錄 `+16` bit 2；`cmd/dsds` 新增
`screenCeasefireTarget`、鍵盤 `0`、原典 `DrawCeasefireTarget`、現代雙用語數字頁，
並讓滑鼠／觸控與 `actions.Select10` 共用同一命中區。新增語系鍵、規格測試、renderer
測試與 pointer 測試。

Docker 驗證：`go test ./internal/game ./internal/i18n ./internal/ui/render` 與 Xvfb 下
`go test ./cmd/dsds ./internal/ui/actions ./internal/ui/layout` 通過；本輪工作容器均為
一次性 `--rm`，無專案相關容器需清理。尚未完成信用 gate、第一期援助國原版選擇、正常
DOSBox 停火畫面重播、`.DT2` 戰鬥寫回、Android 封裝與正式三平台發行。

### 47. 2026-08-10 modern HUD 資源／指令圖示 P3

新增 `docs/spec/14-modern-hud-icons-p3.md`（READY）與
`docs/playtest/40-modern-hud-icons-p3.md`。`internal/ui/theme.HUDIconProvider` 提供
6 種資源與 15 項政略指令的 16×16 transparent-index 圖示；`Modern` 在建構時一次以
純 Go deterministic 幾何生成完整資產組，沒有新增 PNG／SVG 或原版像素。

`render.PanelData`、`DrawCommandPageWithIcons`、`DrawSemanticCommandPageWithIcons` 保留
原典／白話文字與數值，另疊 modern HUD 圖示；retro 仍使用原版字模 fallback。主程式與
`cmd/screenshot` 的主題切換同時交換 battlefield／unit／HUD provider，modern 缺 HUD
provider 時 fail-closed。資源圖示不參與命中區，滑鼠／觸控動作契約未變。

Docker 驗證：`go test ./internal/ui/theme ./internal/ui/render ./cmd/screenshot ./internal/ui/actions ./internal/ui/layout ./internal/i18n` 與 Xvfb 下
`go test ./cmd/dsds` 通過；renderer 對超出調色盤的 HUD 像素有 fail-closed 回歸測試；
省 26 `-menu` retro／modern 截圖 SHA-256 為
`0d5a36a776b0b61566e816d3488c33b010dffd5bc35c88c3c27949711fe8316a`／
`11d04a46e9a7e20061e3a93cbe7695b6ce09061d3670b8be26b35636588b7572`。本輪容器均使用
`--rm`，未留下專案容器。尚未完成十勢力色、向量字型、寬版 HUD、Android 版面與正式
三平台發行包。

### 48. 2026-08-10 貸款信用度零 gate M1

先讀既有 `workplace/ida/WAR.EXE.asm`，沒有以畫面文字猜測。`WAR.EXE` SHA-256
`11dbfcf24686ab7765f788b38514cefd2039d0f60b6bd517d89fb5a84c068015`、IDA Pro
9.4.0.260610、線性 `sub_2164A`（`2164Ah`）在貸款輸入與 `Random(10)` 前執行
`cmp byte ptr [di-4225h], 0`；信用度為零會顯示「無法貸款」並跳回主迴圈。原始
運算元 `[di-4225h]` 保留，證據索引見 `docs/re/34-loan-credit-gate.md`。

新增 `docs/spec/15-loan-credit-gate-m1.md`（READY）。`LoanResult.CreditBlocked` 是
唯一規則標記；`AIWorld.RequestLoan` 零信用度不消耗亂數、不改黃金，帳本不增加外債。
`cmd/dsds` 在外交第 1 項進入貸款額度頁前檢查 ledger slot，提交時再檢查規則結果；
`diplomacy.loan.unavailable` 有原典／現代白話兩套文字。正信用度下的額度核貸公式與
隨機拒絕仍不變，也沒有自行新增「信用度小於額度單位」門檻。

Docker 驗證：`go test ./internal/game ./internal/i18n` 與 Xvfb 下
`go test ./cmd/dsds` 通過；回歸測試確認信用度零時黃金、帳本與 Rand seed 均不變。

### 49. 2026-08-10 貸款隨機拒絕指令成本 M1

`sub_10193` 每次政略指令清 `byte_6FE81`；`sub_2296C` 依它決定是否返回主迴圈，
主迴圈 `loc_103DF` 在非零時扣命令。`sub_2164A` 的 `loc_21967` 在 `Random(10)` 前
立旗標，因此 `loc_21C42` 的隨機拒絕仍扣一個指令；信用度零 gate 在立旗標前返回，
不扣。完整證據見 `docs/re/35-loan-command-cost.md`。

新增 `docs/spec/16-loan-command-cost-m1.md`（READY）。`LoanResult.CommandCompleted` 與
`CreditBlocked` 分開；`cmd/dsds` 只在隨機拒絕提交 `CommandBudget.Spend`，並以
`diplomacy.loan.refused` 的白話文字明示已消耗一個指令，原典保持原版拒絕文字。

Docker 驗證：規則／語系／呈現／輸入測試、Xvfb `cmd/dsds`、Windows `CGO_ENABLED=0`
建置、no-cgo、deny scan、diff check 與容器清理均通過。尚未完成第一期援助國、`.DT2`
戰鬥寫回、Android 與正式三平台包。

### 50. 2026-08-10 外援拒絕／禁運指令成本 M1

`sub_21D1D` 在 `loc_21E9D` 第一次 `Random(10)` 前寫入 `byte_6FE81=1`；隨機拒絕
與張作霖民國 17 年 2–6 月禁運都保留旗標，故與核准一樣消耗一個政略指令。完整
輸入雜湊、IDA Pro 9.4 線性位址與 `sub_10193`／`sub_2296C` 呼叫端資料流見
`docs/re/36-aid-command-cost.md`。

新增 READY `docs/spec/17-aid-command-cost-m1.md`。`AidResult.CommandCompleted` 在
援助亂數前設為真；`cmd/dsds.executeDiplomacyAid` 對拒絕／禁運提交
`CommandBudget.Spend`，防禦性失敗回復省份與亂數種子。雙模式用語改為原典
「各國均拒絕提供援助」，白話明示「已消耗一個指令」。`internal/game/diplomacy_test.go`
新增固定種子回歸測試。

本輪 Docker 已通過規則／語系／呈現／輸入測試、Xvfb `cmd/dsds`、Windows
`CGO_ENABLED=0` 建置、no-cgo、deny scan 與 `git diff --check`；一次性容器均已自動移除，
沒有專案相關容器殘留。第一期援助國、`.DT2` 正常戰鬥寫回、Android 與正式三平台包仍未
完成。

### 51. 2026-08-10 外援核准／拒絕分支勘誤

以同一份 `WAR.EXE`（SHA-256
`11dbfcf24686ab7765f788b38514cefd2039d0f60b6bd517d89fb5a84c068015`）和 IDA Pro
`9.4.0.260610` 重追 `sub_21D1D` 的畫字索引與資源寫入端：`Random(10) <= 6`
會進 `loc_21EF5`，逐字畫出「各國均拒絕提供援助」；`>= 7` 才進四次
`sub_5A467` 的資源入帳。因此原先 5.83 的「70% 核准／禁運拒絕」是被推翻的舊結論，
目前規則是一般 70% 拒絕／30% 核准。

司令 166、民國 17 年、月份 2..6 的條件會從 `loc_21EC5` 直接跳到核准路徑，
目前以 `AidResult.SpecialOverride` 記錄；未把它命名為禁運或補歷史動機。規則、UI、
SPEC-17 與現況索引已同步，完整證據在 `docs/re/37-aid-branch-correction.md`。
下一輪仍先閉合第一期援助國代碼／勢力槽映射，再處理 `.DT1` block 10 同步時機。

### 52. 2026-08-10 第一期外援數值代碼 M1

`sub_21D1D` 的 `[di-5319h]` 是 `.DT1` 區塊 7 的 1-based 勢力反查值；
`22088h` 前的分支已逐項記在 `docs/re/38-aid-donor-code-mapping.md`。新增
`internal/game.AidDonorCode` 與 `RequestAidForFaction`：第一期保留 10／1／5／4／3／2／
default 的原版代碼分支，第二／三期在原版位置覆成 99；選代碼使用外援判定的同一顆
`Random(10)`，不改資源亂數順序。`cmd/dsds.executeDiplomacyAid` 從目前司令反查勢力槽，
新局無 `.DT1` 後半時傳 0，避免把缺資料冒充國家。

新增 READY `docs/spec/18-aid-donor-code-m1.md` 與逐項測試；代碼與國家名稱仍分離，
99／100／101／141／142／146 只當數值 ID。下一個窄切片是 `.DT1` 區塊 10 的 runtime
生成／存檔同步時機，並非再猜援助國名稱。

### 53. 2026-08-10 `.DT1` 尾端位移勘誤與 block10 快照同步

本輪先用 IDA Pro 9.4 重新讀 `sub_595D4`／`sub_59CBF`，沒有沿用舊 offset：

- `byte_6FF8C`（信用度）是 record `+3912h`，檔案 `14610..14619`；
- `byte_6FF96`（外債）是 record `+391Ch`，檔案 `14620..14659`；
- `word_70026` 是 record `+3946h`，檔案 `14662..14681`；
- `14606..14609`、`14660..14661`、`14682` 是七個獨立單 byte runtime 欄位，不能
  當成帳本或「連續檔尾」。

兩份原版樣本交叉驗證：信用度首槽 `64`（100）、外債首槽均為 `00 00 00 00`，
block10 首槽為 `SAVE(1)=58`、`SAVE(2)=166`。先前把 `14606` 解成外債、把
`14656` 解成區塊 10 的結論已推翻；舊紀錄保留，正式格式以 `docs/re/39-dt1-tail-sync.md`
與 `docs/formats/07-dt1-layout.md` 為準。

程式已改正 `SaveBlocks` 與帳本 parser／writer；新增 `WriteMajorPowerLeaders`，session
載入並保存 block10 runtime 快照，autosave 只寫回該 20 B，不自行用 24 槽勢力表造名冊。
新增 READY `docs/spec/19-dt1-runtime-tail-sync-m1.md`。Docker `internal/game`、Xvfb
`cmd/dsds`、no-cgo／Windows `CGO_ENABLED=0`（沿用既有邊界）、`git diff --check` 與
deny scan 要在交接收尾重跑；一次性容器須清理。尚未知的是 block10 的劇本生成／覆滅
同步時機，不可宣稱十大勢力完整動態等價。

### 54. 2026-08-10 音訊解碼 P0（純 Go）

本輪以 `docs/formats/06-mus-tim-audio.md` 的 READY 證據為入口，新增 READY
`docs/spec/20-audio-decode-p0.md`、`internal/audio/mus/mus.go` 與其測試。解析器只做
bytes → typed records：MUS 固定檔頭、running status、F8 filler、SysEx、FC 結束、絕對
tick；TIM 名稱與每筆 28 個 little-endian word 全量保留。它不依賴 Ebiten、音效裝置或
cgo，也沒有宣稱 OPL2 合成／播放已完成。

8 首 MUS 與 8 份 TIM golden、running-status／SysEx、畸形輸入及 TIM residue 測試已通過。
另新增 `docs/re/40-mus-tick-metadata-anomaly.md`：`MAINTHEM` 的 header `totalTick`
為 62880、事件累計 60800；`STRATEGY` 為 15540 對 15600。這兩筆原版 metadata 不一致
由 parser 以 `ActualTick`／`HeaderTickMatches` 明確暴露，不修正原始 bytes；舊有「8/8
totalTick 完全相等」敘述已在 `docs/formats/06` 追加勘誤。

驗證：Docker `go test -count=1 ./internal/audio/mus` 通過；無 cgo 邊界維持不變。下一輪
音訊工作應先寫純 Go OPL2／sample-tick 排程規格與 oracle，不能把本輪 P0 當成已能播音。

### 55. 2026-08-10 戰鬥立即撤退輸入外殼 M1（純 Go／Ebiten）

新增 READY `docs/spec/21-battle-retreat-ui-m1.md`，以唯一已閉合的 DOSBox 正常玩家樣本
陝西省 18 ← 河南省 19 為範圍：撤退候選 `19,26,14,17`。`cmd/dsds/battle.go` 新增
證據限定的候選注入、省份輸入上限（1..39）、刪除／送出、合法提交回政略地圖流程；
立即撤退不呼叫 `.DT2`／`MEM_WAR.DAT` 寫回。其他交戰組合不補鄰接假說，按撤退保持
主選單並標示候選尚未閉合。

`internal/ui/layout.BattleRetreatKeypadButton`、`cmd/dsds/pointer.go` 與
`internal/ui/render/battlepanel.go` 共用 3×4、56×48 的 48 像素命中區與同一套
`actions.Digit*`／`DeleteDigit`／`Submit`；右側面板以既有「撤退／何省」字模畫候選
與鍵盤。新增 cmd、layout、render 回歸測試，既有戰鬥標籤逐像素測試保持不變。

本輪已在 Docker 內格式化並通過 Xvfb 下完整 `go test -count=1 ./...`、
`CGO_ENABLED=0 GOOS=windows go build ./cmd/dsds`、`tools/check_no_cgo.sh`、
`tools/deny_scan.sh --all` 與 `git diff --check`；一次性容器已清除，其他專案容器未觸碰，
本輪檔案擁有權均為目前使用者。不要宣稱一般撤退候選、攻擊六方式、駐軍／查閱、正常
結算 `.DT2` 寫回、OPL2 播放、Android 或三平台發行完成；下一輪可先處理 `.DT2` 平局
oracle 或純 Go OPL2 排程，依證據入口選擇。

### 56. 2026-08-10 純 Go OPL2 風格離線 PCM P1a

使用者確認 remake 不需要 cgo，作者碼維持純 Go／Ebiten。新增 READY
`docs/spec/22-audio-opl2-p1a.md` 與 `internal/audio/opl2`：
`NormalizeEvents` 保留 MUS 已證實的 program、note、volume、pitch bend、tempo 事件；
`Chip`／`RenderSong` 以兩個 operator 的 deterministic 風格合成，輸出 bounded 16-bit
stereo PCM，並對錯誤 sample rate、負 `MaxFrames`、零 tick rate 與無音色 fail-closed。
真實 OPL2 的 register stream、`An` 映射、鼓組與 envelope 尚未證實，打擊聲道暫為靜音
fallback；這不能當成原版逐樣本等價。此層不 import Ebiten，下一步另寫 adapter 才能
接 `audio.Context`／`audio.Player`，並保留 `-audio=off`。

Docker 內已通過 `gofmt` 與 `CGO_ENABLED=0 go test ./internal/audio/opl2`，測試含
PCM digest 與 typed event 回歸。Linux desktop `CGO_ENABLED=0` 整體建置仍受 Ebiten
v2.8.8 上游 GLFW 型別限制；這不代表專案作者碼引入 cgo。不要宣稱 OPL2 播放、正常
結算 `.DT2` 寫回、一般撤退、Android 或三平台包完成。

### 57. 2026-08-10 Ebiten 音訊適配器 M0（開局曲目）

新增 READY `docs/spec/23-audio-ebiten-adapter-m0.md` 與 `internal/ui/audio`。主程式的
`-audio=off` 完全跳過 `audio.Context`；`-audio=retro` 從唯讀 `gameDir` 載入
`SCENE.MUS`／`SCENE.TIM`，以 `internal/audio/opl2` 產生 PCM，建立單一 Ebiten
`audio.Player` 播放開局預覽。重播會先關閉舊 player，off／未知模式／資產與 player 錯誤
都有明示處理；新增 `oto/v3.3.3` 鎖定依賴。這是 UI 適配器窄切片，不代表戰鬥／戰報／
劇情／結局曲目切換、F3／prefs 音訊軸、modern Ogg、原版 register parity、Android 或
三平台音訊發行完成。

Docker/Xvfb 已通過 `go test ./internal/ui/audio ./cmd/dsds`；後續完整回歸仍需包含
Windows `CGO_ENABLED=0`、no-cgo、deny scan、diff check 與容器清理。作者程式碼不得加入
`cgo`，也不要把 `SCENE.MUS/TIM` 複製進版控。

### 58. 2026-08-10 戰鬥收尾分類 M0（純 Go／Ebiten）

新增 READY `docs/spec/24-battle-settlement-classification-m0.md` 與
`internal/game/battlesettlement.go`。分類入口保留原版 `byte_64901` 的三種已證實
控制流：立即撤退（不寫回）、第一／第二方已分勝負（不進平局 `sub_3964E`），以及
回合上限且無勝方的平局（需要進入 `sub_3964E`）。一般月份上限 16、二月上限 15；
中途無勝方狀態與未知勝方值都 fail-closed。

`cmd/dsds` 的交戰全滅、補給見底、AI 必勝、回合上限與已證實立即撤退分支現在共用
`BattleSettlement` 分類旗標，仍不自行填寫 `.DT2`／`MEM_WAR.DAT` 未解欄位。純 Go 測試
覆蓋 15／16 回合與非法值。這輪沒有新增 `cgo`；Ebiten／純 Go 邊界不變。

`battleState` 在開戰時從 session 月份固定回合上限（二月 15、其餘月份 16），不在
戰鬥途中重新讀取日期；若舊測試 fixture 沒填月份，才退回一般月份 16。

Docker 收尾仍需重跑完整 `go test -count=1 ./...`、Windows `CGO_ENABLED=0` 建置、
`tools/check_no_cgo.sh`、`tools/deny_scan.sh --all`、`git diff --check`、`go mod verify`
並檢查專案容器清理狀態。尚未完成正常平局 `.DT2` 寫回 oracle、一般撤退／攻擊六方式、
駐軍／查閱、戰鬥曲目切換、Android 與正式三平台包。

### 59. 2026-08-10 戰鬥寫回／操作／自傳／音訊 M1

本輪依使用者要求暫停 DOSBox 取樣，先完成可運行的 remake 垂直切片：

- 新增 READY `docs/spec/25-battle-state-writeback-m1.md`。`BattleState.ApplyRemakeSnapshot`
  以 `Raw` 為基底寫入已證實的資源、roster、runtime ID 與來源省欄位，保留 SlotsA/B
  等未知 bytes；`cmd/dsds` 將結束戰鬥與 F10 autosave 寫到 `-save` 同目錄的 DT2／
  MEM_WAR 副本，立即撤退仍不寫回。這是 remake 快照策略，不是原版 sub_3964E 時機等價。
- 戰鬥操作已接：攻擊子選單 `1..6`／滑鼠目標共用相鄰目標排序與 `Engage`；駐軍會尋找
  第一個可達城市並更新命令；查閱顯示選中單位摘要。尚未解出的原版攻擊方式細節仍以
  remake 差異明列，不能當成原版逐分支完成。
- 人物自傳既有 417 人／486 槽位／387 篇正文與 30 筆 unknown 邊界保留；新增
  `cmd/dsds/biography_test.go` 驗證實際 roster → biography → 分頁入口，未以杜撰資料
  填補沒有可追溯來源的槽位。
- 新增 READY `docs/spec/26-audio-context-switch-m1.md`。`Track` 曲目庫與延遲 PCM
  產生接到政略／戰鬥畫面，缺檔 fallback、`-audio=off` gate 與單一 player 生命週期均保留；
  淡入淡出、register parity、modern Ogg、Android／三平台音訊發行仍不在本輪。

目前已在 Docker/Xvfb 通過完整 `go test -count=1 ./...`；Windows
`CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build ./cmd/dsds`、no-cgo、deny scan（550
檔案、原版資產零命中）、diff check 與 `go mod verify` 也通過。一次性容器使用 `--rm`；
容器狀態檢查只看到既有其他專案容器，未停止或刪除它們；本輪新增檔案擁有權均為目前使用者。

### 60. 2026-08-10 三條優先切片補強

本輪依使用者指示不做 DOSBox 取樣，完成下列可回歸的窄切片：

- `internal/game/battleranged.go` 依 `WAR.EXE`／IDA Pro 9.4 的
  `sub_58854`（`58854h`）、`sub_55632`（`55632h`）與 `sub_57B15`（`57B15h`）證據，
  產生六方向遠程候選，排除森林／高山，並執行三次戰損、只扣遠程目標。`cmd/dsds` 的
  鍵盤／滑鼠／觸控攻擊目標表現已共用，兵種 4 的非相鄰候選可直接操作；邊界折返、
  彈藥消耗、動畫／音效與六種原版攻擊方式仍是 unknown／remake 差異。
- `internal/game.WriteWarRecordBranch` 明確寫入 `.DT1` Branch4／Branch8 的已證實
  四資源與將領 ID 前綴，保留 `+0x12` union、未提供槽位與未知尾端；不把同步時機
  宣稱完成。`docs/formats/07-dt1-layout.md`、`docs/re/05-mem-war-record.md` 與
  `docs/re/09-ranged-attack.md` 已追加證據契約。
- `internal/ui/audio.Manager` 快取已渲染 PCM；`internal/audio/opl2/registers.go`
  提供 `ProgramRegisterWrites` 靜態 YM3812 bitfield adapter（strong inference，非
  SDFA parity）。

目標測試已在 Docker/Xvfb 通過：
`go test ./cmd/dsds ./internal/game ./internal/audio/... ./internal/ui/audio`。
收尾仍需跑完整 `go test -count=1 ./...`、`CGO_ENABLED=0` Windows build、
`tools/check_no_cgo.sh`、`tools/deny_scan.sh --all`、`git diff --check`、`go mod verify`，
並檢查專案相關容器均已清理。不要宣稱六攻擊分支全等價、原版 DT1/DT2 同步時機、
SDFA register parity、Android 或三平台包已完成。

### 61. 2026-08-10 DT1 統一存檔與 SDFA 直接播放決策

本輪依使用者明確指示完成兩條窄切片：

1. **DT1：** 新增 `DT1State`、`ParseDT1`、`WriteDT1`，並將 `cmd/dsds.autosave`
   改走單一入口。勢力表已證實的領袖／sentinel／關係欄位、24 槽勢力領袖表、戰爭
   記錄 `+0/+2`、停火、block10 runtime 清單、外交帳本與既有省份／將領 writer 都
   接上；區塊 7 非領袖殘留、戰爭記錄未知欄位、勢力表未知欄位與七個單 byte runtime
   欄位維持原始 bytes。未修改 `ParseDT1` snapshot 寫回 `SAVE(1).DT1` 已通過
   byte-for-byte 測試，mutation 測試也驗證未知範圍未被碰觸。
2. **SDFA／音訊：** 不再解包或追求 `SDFA.EXE` 寄存器等價。新增
   `opl2.RenderAdLib` 直接將 MUS/TIM 轉 bounded PCM；OPL2 rhythm channel 6..10
   由純 Go kick／snare／hat／tom／cymbal 合成，不再是靜音 fallback；UI 以既有
   `Ebiten audio.Player` 播放。完成條件是音樂與音效可聽，不是 register trace 或
   逐樣本 parity。`ProgramRegisterWrites` 僅保留 optional strong-inference adapter。

本輪已在 Docker 內通過 `internal/game`、`internal/audio/opl2`、`internal/ui/audio`。
完整 Xvfb 回歸、no-cgo、deny scan、Windows build、`go mod verify`、diff check 與容器
清理仍需收尾重跑；`SDFA.EXE` 只列獨立精度研究，不得重新成為播放阻塞條件。

### 62. 2026-08-10 戰鬥世界結算／人物自傳可驗收 M2

本輪先完成使用者指定的兩條主線，範圍以可玩的 remake 路徑為準：

- 新增 READY `docs/spec/27-battle-settlement-biography-m2.md` 與
  `docs/playtest/41-battle-biography-m2.md`。`startBattle` 在 `BattleSim` 成功後才
  設 `ProvinceFlagInBattle`；ESC／立即撤退／守方勝／回合平局會清旗。攻方勝由
  `game.ApplyBattleSettlement` 讓來源省司令接管目標省，移動存活攻方的 `Province`
  與 `world.Units`，地圖游標跟隨目標；`battleState.settled` 使重複結算冪等。
- `syncBattleForces` 同步戰場已證實的 `Force` 到 `a.generals` 與 `a.world.Strengths`，
  避免存檔側與政略 AI 讀到兩份不同兵力。未知 `.DT2` bytes、命令、格位、補給仍不猜。
- `PeopleDB.PersonAt` 的排除槽位不再畫自傳按鈕／pointer target；#274「無省長」只可
  返回。沒有正文但可接合的人物仍顯示一頁 unknown fallback。486 槽位驗收為 485 可
  接合／1 排除，387 篇有正文；新增實際語系接合、排除與 fallback 測試。
- 戰鬥攻擊子狀態在敵軍格上畫穩定 `1..6` 目標標號；修正原本 attack mode 會把
  `BattleAttackTarget` 誤濾掉的 pointer bug，三種輸入裝置共用同一份目標表。

Docker `internal/game` 與 Xvfb `cmd/dsds` 目標測試通過。交接收尾仍要跑完整
`go test -count=1 ./...`、Windows `CGO_ENABLED=0` build、no-cgo、deny scan、
`git diff --check`、`go mod verify` 與專案容器清理；完成宣告不可擴張為原版六種攻擊、
一般撤退候選、正常平局原始同步時機或全流程等價。人物自傳仍是 remake 新增的唯讀
文化功能，unknown 不以杜撰內容填補。

### 63. 2026-08-10 存檔反查格與英日語系接續

本輪在既有 DT1／DT2 非破壞性鏈上補完可由證據支持的最後一小段：

- `internal/game.WriteFactionOfGeneralLeaders` 只寫區塊 7 中目前非零勢力領袖的
  confirmed 反查格，`WriteDT1` 已接入；重複／超出 274 筆將領範圍 fail-closed，
  舊領袖格與其餘 265 個殘留維持原 bytes。`docs/formats/07`、`docs/re/39`、
  `docs/spec/19` 與 DT1 測試同步。
- `tools/gen_locale_packs.py` 可重生 `translations/en`／`translations/ja`，每套含
  132 個 wording 鍵、39 省名、417 人物。`cmd/dsds -locale` 遇非繁中語系會自動採
  semantic wording；英語詞條與地名可直接呈現，日語先採倚天可呈現的繁體漢字短句。
- 英／日人物保留歷史姓名與繁中自傳正文，寫入 `bio_language=zh-Hant`、
  `bio_status=source-fallback`，自傳頁顯示來源提示；因此「語系資料鏈／入口完成」
  不等於 417 篇完整英／日譯稿完成。日文假名字型、比例字型與可追溯人物譯稿仍待後續。

本輪 Docker 內已通過 `go test ./internal/game ./internal/i18n ./internal/ui/render`，
其中包含英日 `%d` 契約與倚天字型覆蓋測試；cmd/dsds 的 Xvfb 全量回歸、no-cgo、deny
scan、Windows 交叉建置與 diff check 要在收尾再跑。不要把未知 DT1／DT2 同步時機或
英／日人物翻譯宣稱成原版等價／完整國際化。

## 64. 2026-08-10 合併後剩餘工作盤點

本節是目前可執行的 backlog；狀態以 `CONTEXT.md`、目前程式與 `READY` 規格為準。
`WORKLIST.md` 已取代原 `HANDOFF.md`，歷史條目保留在本檔前文，新的進度只追加於此。

### A. 存檔與逆向（需要證據／原版 oracle）

- **DT1 剩餘欄位：**補區塊 2 戰爭記錄的未解欄位與兩種 layout 的有效分支／同步時機，
  勢力表 `+2`／尾端，以及檔尾 7 B；區塊 7 未確認殘留不可猜寫。
- **DT1 block10：**確認劇本生成、勢力覆滅與 runtime 名冊何時同步，不能把目前快照值當成
  固定名冊。
- **DT2／`MEM_WAR.DAT`：**取得正常攻擊、回合上限平局與戰後結算的固定 DOSBox 樣本，
  完成 469-byte 剩餘欄位及原版寫回時機；目前的 `ApplyRemakeSnapshot` 只代表 remake
  非破壞性副本策略。
- **執行檔交接：**`SDFA.EXE` 解包與寄存器等價列為可選精度研究，不是播放前置；其他模組
  間檔案交接仍需在 IDA Pro 9.4 中補證，所有新增語意維持位址／雜湊／推論等級索引。

### B. 戰鬥與規則（實作及 oracle 兩條線）

- 補原版六種攻擊方式的分支差異、邊界折返、彈藥／戰損副作用、動畫與音效時機；目前
  遠程兵種 4 是已證實的窄切片，不是六分支全等價。
- 補一般部署、完整撤退候選、駐軍後處理、戰後資源／命令／格位副作用與正常平局流程，
  並用正常玩家路徑驗證，不使用傳送、強制勝利或直接賦值。
- 補戰鬥 AI 未解行動、執行期單位欄位與 196×196 成本矩陣用途；新結論同步到
  `docs/mechanics/70-ai.md`。
- 逐指令重播政略回合、年度結算、讀取／存檔／退出／重載，建立原版／remake 固定狀態
  對照；測試綠燈本身不等於原版等價。

### C. 人物自傳與語系（產品內容）

- 為可追溯的 387 篇正文建立英文／日文逐句譯稿與來源 metadata；目前英／日 417 筆仍以
  `bio_status=source-fallback` 顯示繁中原文，不能稱完整國際化。
- 依 `docs/design/22-biography-integration.md` 決定並升級「研究稿 → additive overlay →
  發行語系檔」的正式流程；未解除 DRAFT 寫回禁令前，不直接覆寫 authoritative
  `people.json`。
- 補日文假名、比例字型、標點禁則與英文／日文長文分頁；目前日文是倚天可呈現的漢字短句
  相容層，非自然日文排版。
- 翻譯尚未覆蓋的新聞圖像、劇情／敘事文本與其來源索引；`132` 個 UI wording 鍵與 `39`
  個省名只代表目前語系資料鏈已接通。

### D. 呈現、輸入與音訊

- 完成 modern UI 的十勢力配色、向量／比例字型、寬版 HUD 與完整資產對齊；目前
  `P0/P1/P2a/P2b/P3` 是窄切片，設計總稿仍為 DRAFT。
- 將滑鼠／觸控由目前選單、清單與戰鬥 M0–M2 擴至全部玩家指令、長清單、訊息與設定頁，
  維持鍵盤／滑鼠／Android 單指共用 action dispatch。
- 純 Go OPL2／AdLib 已可播放開局與情境音訊；仍需補完整曲目／戰報／劇情切換、音效事件、
  音量／淡入淡出與裝置失敗路徑。SDFA register parity 仍是獨立研究，不得重新阻塞播放。
- `其他選項` 尚有原版 3–7 的實際功能、訊息／音效等後續選項；不可只畫 ON／OFF 假畫面。

### 65. 2026-08-10 block10／攻擊分派證據收斂與翻譯決策閘

本輪先完成不依賴新 oracle 或翻譯授權的安全切片：

- 新增 docs/re/41-block10-attack-dispatch.md，記錄 WAR.EXE SHA-256、IDA Pro 版本、
  IDA linear 位址與推論等級。證據已修正「六種攻擊」說法：sub_4D585 的 1..6 是
  六方向／目標輸入，sub_4BF27 之後才分派到正規、協同、五次正規、遠程及特殊
  handler；block10 目前只證實初始化／讀取／移除／接班確認後替換，劇本生成與覆滅
  同步時機仍 unknown。
- 新增 internal/game/battleattack.go 的 EngageRepeated，依 sub_53428 的五次
  sub_51D68 窄切片重用既有戰力／戰損函式；新增測試並明示 strong inference。
  沒有把未知的協同支援除法、特殊動畫／音效或 handler callpoint 猜成完成。
- docs/mechanics/30-combat.md 與 cmd/dsds/battle.go 已把玩家可見文字改成「選擇
  目標」，避免把 1..6 顯示成六種攻擊。
- 新增 DRAFT docs/spec/28-locale-biography-overlay.md 與
  tools/locale_bio_gate.py。gate 會從繁中 base 加上 61 篇 authored overlay 重建
  387 篇 expected source，驗證 locale、姓名、來源／譯文 SHA-256、review metadata
  與 machine-draft／human-reviewed；只有 --release 且全篇 human-reviewed 才可通過
  正式發行閘。

仍待使用者決策／外部證據的項目：

- DT1 block10 劇本生成／勢力覆滅／runtime 名冊同步時機，以及 DT2／MEM_WAR 469 B
  剩餘欄位，須在使用者解除「暫停 DOSBox 取樣」後，以固定正常玩家 oracle 取得。
- 六種命名攻擊並未被證據發現；原版 handler 的完整選擇條件、協同戰損副作用、
  一般撤退／部署／結算與 AI 執行期欄位仍不能用 remake 測試綠燈代替。
- 英／日 387 篇人物正文目前仍是 source-fallback。正式 locale overlay 需要先決定
  是否接受標記為 machine-draft 的可審稿初稿；在決定前不能把繁中正文或自動代換
  寫成 translated，也不能解除 DESIGN-22／SPEC-11 的權威資料禁令。

### E. 驗收、打包與發行

- 建立無 debug shortcut 的完整玩家路徑 gate：新局、移動、指令、戰鬥、結算、存檔、退出、
  重載與至少一個後期狀態；補固定狀態原版／remake 截圖與差異標記。
- 完成 Linux／Windows／macOS 三平台可重現包裝與 smoke test，確認純 Go／Ebiten backend
  的平台限制；Android 另建移植包與實機觸控驗收，不把桌面交叉建置代替 Android 完成。
- 每一輪收尾維持 Docker `--rm` 容器、`go test -count=1 ./...`、no-cgo、deny scan、
  `go mod verify`、`git diff --check` 與 dirty-tree 檢查；完成前同步修正 README、
  `CONTEXT.md`、本檔與 READY／DRAFT 狀態。

### 66. 2026-08-11 block10 證據收斂與英日 source-fallback 接合

- 新增 `docs/re/42-block10-init-and-new-game-flags.md`：以 IDA Pro 9.4、WAR.EXE
  SHA-256 `11dbfcf24686ab7765f788b38514cefd2039d0f60b6bd517d89fb5a84c068015` 補上
  `sub_38839` 的第一期新局選項→block10 寫入、`sub_391E1` 清零／預設、
  `sub_3562B`→`sub_353C4` 月結算接班鏈，以及 `sub_38DFE` 的電腦個性／AI 等級／
  難度位元寫入。並訂正 `[di-418Ch]` 與 `word_70026` 是同一個 block10 儲存區；
  記憶體接班更新已可標 confirmed，月結算後何時呼叫 `sub_595D4` 落盤仍 pending。
- `docs/re/41`、`docs/playtest/13-new-game-settings.md`、`docs/mechanics/70-ai.md`
  已同步上述勘誤與設定位元證據；不再把 block10 誤稱為第二份獨立 runtime 表，
  也不再把三組設定的 bit 對應留作假說。
- `tools/gen_locale_packs.py` 現在會先接合已驗證的繁中 `people-authored.json` 61 篇，
  英／日語系包因此各含 387 篇可追溯 source-fallback（417 人物、#274 仍排除），
  不再只複製 326 篇 base。新增 `internal/i18n` 測試固定此數量。
- `internal/i18n.PeopleDB` 新增可選 `people-biography-overlay.json` 載入器：整份
  387 筆先驗證 ID／姓名／來源與譯文 SHA-256／review metadata，再原子套用；缺檔仍
  維持 source-fallback，machine-draft 不會自動變成 human-reviewed。
- DOSBox 正常新局 oracle 兩輪只取得部署／非差分控制樣本；最近一次 `train:10` 在
  snapshot 前已回到政略畫面，`0` 被當成主選單輸入，沒有新增戰鬥寫回證據。不要把
  該輪當成 DT2 平局樣本；下一輪仍須抓住部署窗口並完成無勝方回合上限。

本節之後仍待：DT1 未解欄位與存檔落盤時點、DT2 469 B 正常平局／勝負差分、原版
handler 完整副作用與 AI 未解行動，以及英／日真正 387 篇逐句譯稿、日文假名字型／
敘事文本。翻譯父層決策尚未收到使用者答案，SPEC-28 維持 DRAFT。

### 67. 2026-08-11 DT2 欄位搬運證據與正式 overlay 文件接線

- 新增 `docs/re/43-dt2-record-write-fields.md`，以同一份 `WAR.EXE` SHA-256 與
  IDA Pro 9.4 逐項對齊 `sub_3964E`／`sub_4F468` 的 `0Ah`、`14h`、`0C8h` 搬運。
  `.DT2`／`MEM_WAR.DAT` 的 `+0..+7` 現標為第一方四項資源；第二方相鄰的
  `word_64934/38/3C/40` 是寫回分支直接更新省份記錄的 runtime 欄位，不再誤列為
  第二份 DT2 header。`+8/+18` 只確認 10-byte 形狀與 `0xFF` 樣本哨兵，元素語意
  仍 unknown；`+68/+268` 確認 100×u16 runtime ID buffer，但檔案／live 同步時機
  仍 unknown。
- `internal/game.BattleState`、`BattleResources` 與 `docs/re/05` 已同步上述中性
  第一方資源註記；`ApplyRemakeSnapshot` 不碰 `SlotsA/B`，避免把 byte 形狀升格成
  部署／部隊語意。現有 byte-for-byte writer 測試維持不變。
- `tools/gen_locale_packs.py` 與 `translations/README.md` 已把 387 篇
  source-fallback、`people-biography-overlay.json`、來源／譯文 SHA-256、review
  metadata 與 fail-closed gate 寫入可重生流程；重生後英／日 README 同步。這仍不是
  英／日逐篇譯稿，也不解除 SPEC-28 的 machine-draft／human-reviewed 父層裁決。

本節後仍待：DT1／DT2 正常玩家存檔 oracle 與未解欄位、六個原版 handler 的完整副作用、
撤退／部署／AI、英日 387 篇實際譯文、日文假名／比例字型與敘事文本。上述 pending
不能用 remake 單元測試或 source-fallback 進度冒充完成。

### 68. 2026-08-11 block10 明確存檔入口 xref

- 使用專案指定的 IDA Pro 9.4 image 執行 `tools/ida_func_xref.idc`，
  `sub_595D4` 的直接 code xref 只有 `sub_1B399@1B6BAh` 與
  `sub_5B1B6@5B5CEh`（輸出 `workplace/ida/user-output/funcxref-sub_595D4.txt`）。
- `sub_1B399` 是存檔選單在槽號／Y-N 確認後呼叫；`sub_5B1B6` 是 `h..q` 快捷存檔。
  兩者都把呼叫當下的 `word_70026` 寫到 `.DT1 +3946h`，所以「玩家明確存檔時
  block10 落盤」已 confirmed。
- 目前沒有直接 xref 顯示 `sub_3562B`／`sub_353C4` 必然自動落盤；函式指標與
  未匯出的間接呼叫仍是限制。月結算／接班後未手動存檔的 DT1 差分仍需正常 oracle，
  不可把 manual-save confirmed 擴大成自動同步完成。

### 69. 2026-08-11 戰鬥 helper 證據與 oracle 重播勘誤

- 新增 `docs/re/44-combat-handler-helper-boundary.md`：以同一份 `WAR.EXE` SHA-256、
  IDA Pro `9.4.0.260610` 與 IDA linear address 保存 `sub_4BF27`、`sub_4D585`、
  `sub_4B827`、`sub_53111`、`sub_530B4`、`sub_5301B`、`sub_51D68`、
  `sub_51EC0`／`sub_51F19` 的呼叫與寫入邊界。補上 `+7A9A`／`+7A9B` 只確認
  有扣值與夾位，尚不能命名為彈藥／燃料或協同除法；不把 1..6 輸入誤稱為六種
  命名攻擊。
- `docs/playtest/16-dt2-live-battle-snapshot.md` 追加 2026-08-11 DOSBox 正常
  玩家重播：部署畫面出現但 AI／亂數窗口不穩；負控制的 `0→Y` 與每 10 秒截圖均
  沒有新增 `sub_3964E` 寫回後樣本。`tools/dosbox_runner.sh` 新增時間線步驟
  兩端空白裁切，避免 `; snap:name` 在最後一步遮蔽已取得的證據。

本節後仍待：DT1／DT2 未解欄位與正常平局／勝負 oracle、六個 handler 的完整資源與
特殊副作用、撤退／部署／結算／AI 的原版時機，以及英／日 387 篇實際譯文與日文
字型／敘事文本。`SPEC-28` 仍等待使用者裁決 machine-draft 或 human-reviewed-only，
不得由代理自行選擇。

### 70. 2026-08-11 machine-draft／戰鬥副作用／AI chain 收尾

本輪使用者已接受 machine translation 初稿，並要求補上協同／特殊副作用、完整撤退部署
結算與 AI parity 接線；以下是目前唯一有效的狀態表，未完成的原版 oracle 仍不被綠色
測試覆蓋：

- **自傳與語系：完成 remake 預覽鏈。** `tools/gen_biography_overlays.py` 在 Docker
  內以 deterministic 詞表／日期／事實模板產生英／日各 387 筆 overlay；`locale_bio_gate`
  修正為正確檢查來源路徑與 SHA-256，兩包 gate 通過。`PeopleDB` 套用後標
  `machine-draft`，正式 release 仍需全篇 `human-reviewed`。日文假名、比例字型、
  長文排版與敘事／新聞翻譯列為待辦，不得把 machine-draft 寫成正式完成。
- **協同／特殊副作用：完成規則層接線。** `ResolveBattleAttack` 統一正規、協同、
  衝鋒、炮擊與特殊視覺分支；經驗、體力、士氣、低士氣兵損、支援分攤／死亡停止、
  衝鋒暫時格位事件與遠程三次戰損均有測試。六個原版攻擊 handler 的完整選擇、
  動畫／音效／彈藥時機仍 unknown。
- **撤退／部署／結算：完成 remake 正常路徑。** 攻方部署零值正規化、守方呼叫端落點、
  省份鄰接順序撤退 fallback、立即撤退 UI 清旗／切換目的地、勝負／平局投影與人物／
  世界統計同步已接通；18←19 是唯一原版撤退樣本，其他候選是明示 fallback。完整
  原版撤退時機、部署來源與 `.DT2` 469-byte 寫回仍 pending。
- **AI：完成 remake chain 垂直接線，未宣稱 oracle parity。** `AIWorld.ResolveAttack`
  與正常玩家守方 AI 都走 `AutoResolveByChain`／13 handler，資源比率、彈藥、後援與
  後段旗標由當下狀態填入；`BattleRunStats.Unimplemented==0` 是場次 gate。原版
  `sub_3A9F4`／`sub_3AABA` 細分時機、預約表生命週期、第一／第二方方位與正常 DT2
  oracle 尚未閉合；YouTube 未找到可重播的本作 AI 影片，無外部 parity 證據可升級。
- **已驗證：** Docker Xvfb `go test -count=1 ./...` 通過；其餘 no-cgo、deny scan、
  `go mod verify`、Windows 交叉建置、diff check、封裝與 Docker 清理需收尾執行。

#### 本輪剩餘工作（依 gate）

1. 在 Docker 內完成 no-cgo、資產拒絕掃描、`go mod verify`、Windows `CGO_ENABLED=0`
   建置、`git diff --check`，並檢查無殘留專案容器。
2. 若要宣稱原版 AI parity，需新的正常玩家 oracle 或 IDA 證據閉合上列時機／方位／
   預約表；不可用 YouTube 無關影片或 remake 自測代替。
3. 逐篇人審 machine-draft、日文假名／比例字型與長文 overlay，完成後才可跑
   `tools/locale_bio_gate.py --release`；在此之前保持預覽狀態。
4. DT1／DT2 未解欄位、一般撤退／部署原版時機、六 handler 完整資源與三平台封裝仍依
   原 worklist 待辦，SDFA register parity 不重新成為音訊播放前置。

### 71. 2026-08-11 DT1／DT2 與攻擊「完成」回報重核對

本輪重新查閱目前 `CONTEXT.md`、`docs/formats/07`、`docs/re/05`、`docs/re/41`、
`docs/re/43`、`docs/re/44` 與 IDA 匯出。結論是：先前的「完成」必須分層解讀。

- `.DT1` 十個 `$basg` 區塊的**位置與大小**已完成；檔頭未解 byte、區塊 2／3 的
  剩餘欄位與 layout 時機、區塊 7 殘留、block10 自動同步，以及七個獨立 byte 尚未全解。
- `.DT2` 的 469-byte **切分**已完成；`+8/+18` 元素語意、`+68/+268` 檔案／live
  同步與正常戰鬥後 oracle 尚未完成。現有 writer 是保留未知 bytes 的 remake writer。
- `sub_4D585` 的 `1..6` 是六方向／目標輸入；不是六種命名攻擊。`sub_4BF27` 的
  handler 控制流與部分 helper 已接到 remake，但完整選擇、副作用、資源／動畫／音效
  時機與原版正常玩家 parity 仍 pending。

此項重核對不回退已完成的 parser、writer 或 remake handler；它只把「結構定位／窄切片」
與「原版完整語意／parity」分開，避免同一項工作再次被重開或誤報。

### 72. 2026-08-11 第二層六鍵選單勘誤後的停止邊界

- 已由同一份 `WAR.EXE`／IDA Pro 9.4 直接確認：`sub_4D585` 的 `"0123456"` 是
  六方向／目標游標；`sub_4BF27` 的 `"123456"` 是另一個目標檢查後的第二層選單。
- `sub_4BF27` 的 `1..5` 分別 call `sub_4B827`、`sub_53111`、`sub_4B854`、
  `sub_4BA1E`、`sub_4BCBE`；`6` 僅在輸入層被接受，隨後落到清理路徑，沒有第六個
  handler call。這是控制流完成，不是六個攻擊名稱／副作用完成。
- `docs/re/41`、`docs/re/44`、`docs/mechanics/30-combat.md`、`docs/spec/29`、
  UI／CLI 註解與 `docs/playtest/16` 已同步；舊的「只看六方向」結論保留為歷史勘誤，
  後續不得重新把兩條輸入鏈合併。
- 因此本項不再反覆重開靜態核對。下一個可解鎖工作只剩：正常玩家 oracle（第 6 鍵／
  handler 名稱與副作用）、DT1／DT2 寫回差分、完整撤退／部署／AI 時機；在 oracle
  暫停期間維持 `unknown`，不以 remake 綠燈或無關影片補洞。

### 73. 2026-08-11 接受近似宣稱，停止 parity 迴圈

使用者已明確選擇 B：不再為 DT1／DT2、handler、第 6 鍵、撤退／部署與 AI 時機建立
原始 `WAR.EXE` 函式 oracle，也不跑完整 DOSBox 劇本。交付口徑改為：

- **DT1／DT2：** 以全檔差分遮罩、byte-for-byte round-trip、未解 sentinel 保留與
  分支輸入 fail-closed，宣稱「remake 非破壞性寫回近似契約已驗證」；不宣稱原版
  寫回時機或未解欄位全解。
- **handler／第 6 鍵：** 以 deterministic `Combatant` 矩陣驗證已知的經驗／體力／
  士氣／支援／死亡停止／炮擊／衝鋒效果；`battle.command.6` 明確拒絕，
  `battle.attack-target.6` 僅是第六個目標序號；不宣稱原版第 6 鍵語意或完整
  動畫／音效／資源副作用。
- **撤退／部署／結算：** 以固定地圖、佔位與省份表合成測試驗證十格部署、掃描／
  避讓、18←19 原版樣本與鄰接 fallback；其他候選、駐軍後處理與 DT2 落盤維持
  `unknown`，但不阻塞近似 remake 交付。
- **AI：** 以 `TraceDecisions`／`AutoResolveByChain` 的事件序列、回合上限、補給順序、
  決勝 gate、`DefaultPostStageOpen` 與 `Unimplemented==0` 驗證 deterministic
  近似時序；不宣稱 `sub_3A9F4`／`sub_3AABA` 逐回合 parity。

新增規格 `docs/spec/30-approximate-validation-m2.md` 與回歸測試
`internal/game/approximate_validation_test.go`；第 6 鍵 namespace 測試在
`internal/ui/actions/battle_test.go`。本節取代「必須取得原版 oracle 才能交付」的
前置條件，但保留所有原版 unknown 證據與歷史勘誤。除非使用者另行要求原版等價或
新證據改變玩家體驗，否則不重新開啟 emulator／長流程 DOSBox 分支。

### 74. 2026-08-11 近似驗證 gate 已通過

本輪在 `dsds-go:1.25` Docker（`internal/game`／`internal/ui/actions`）與 Xvfb
（`cmd/dsds`）完成快速回歸；`check_no_cgo.sh`、`deny_scan.sh --all`（586 檔案、
原版資產零命中）、`go mod verify`、`git diff --check` 均通過，`great-era` 相關
容器沒有殘留。這些是近似 remake 的驗收結果，不是原版 parity；原版 unknown 清單
仍依 `CONTEXT.md` §5.104–§5.106 保留。

### 75. 2026-08-11 Modern UI／語系敘事／音訊收尾

本輪已把使用者指定的三條玩家垂直鏈接上，完成邊界如下：

- **Modern UI：** 640×350 邏輯畫布上的 `UIStyle`、資訊卡、十勢力 remake 色帶、
  資源／指令 HUD、戰鬥面板控制鍵、地圖外框與主題切換已接；`retro` 仍保留原版
  字模／色彩。寬版 1280×720、比例／向量字型、symbol 軸與正式美術資產不是本輪
  gate，避免把安全的完整外殼重新開成無限視覺研究。
- **語系：** `wording.json` 現有 154 個必要鍵，panel／narrative caption 也走語系
  catalog；`en`／`ja` 的 387 篇人物自傳 overlay 接受 machine-draft，runtime 明示
  `bio_status` 與來源。正式人審、日文假名字型、長文比例排版仍是發行 gate。
- **敘事：** `SPEC-31` 與 `NarrativeCatalog` 接上 17 張 `NEWSDATA.DAT` 來源圖像，
  `N`／滑鼠／觸控共用入口，每頁四張；#0／#9 有可追溯 caption，其餘 15 張
  `source-image` fallback。未知字型不做猜測 OCR，也不改事件規則。
- **Retro 音序：** `SCENE`／`STRATEGY`／戰鬥／`WALL`／`FINAL` 情境已接，Manager 以
  bounded 18 frame（約 300 ms）crossfade 關閉舊 player；缺檔沿用目前曲目，
  `-audio=off` 不受影響。SDFA register parity 不列為播放前置。
- **Modern 音樂規劃：** `docs/design/32-modern-music-direction.md` 只建立原創 cue
  矩陣、3–5 音新動機、Ogg loop／stem／授權／QA 與分期；尚未產生音檔，不能宣稱
  `audio=modern` 已完成。

#### 75.1 剩餘工作（不重開已閉合路徑）

1. 逐篇人審英／日 machine-draft 自傳；完成來源、翻譯與 reviewer metadata 後才跑
   `tools/locale_bio_gate.py --release`。
2. 若產品要寬版 modern，先做可丟棄 prototype，再決定字型授權、1280×720／1280×700
   畫布與 Android fallback；目前 640×350 不阻塞。
3. 依 Modern 音樂 DRAFT 取得作曲／生成 prototype，做旋律相似性、授權、loop、mono、
   loudness 與 Android QA；音檔與 manifest 未進目前工作樹。
4. 三平台／Android 封裝與實機驗收；Linux／Xvfb 播放成功不能代替 macOS／Windows／
   Android 音訊與點擊驗收。
5. 原版 DT1／DT2 未解欄位、handler 副作用、第 6 鍵語意、撤退／部署時機與 AI parity
   仍是 `unknown` evidence boundary；使用者已選 B，除非明確改變目標，不重開 DOSBox
   長流程或 SDFA parity。

### 76. 2026-08-11 Modern Ogg runtime 接線完成（音樂內容未交付）

審稿代理確認 Modern UI 640×350 邏輯外殼的主題、semantic 面板、人物／省份名稱、敘事
畫廊與 fail-closed 字庫入口已有程式／測試證據；主線新增 `docs/spec/32-modern-ogg-runtime-m1.md`
並完成：

- `internal/ui/audio.ModeModern`、`ModernTrackSource`、`NewModernTracks`：使用 Ebiten
  內建純 Go `audio/vorbis`，初始 Ogg 取樣率建 context，其他 cue 延遲解碼。
- `LoadModernTracks`：schema 1、六個固定 cue、同目錄 `.ogg`、`author`／`license`、
  SHA-256 與每聲道 sample 的 loop metadata；scene 必要，optional cue 缺檔留下 warning。
- `cmd/dsds -audio modern -modern-audio <dir>`：manifest／scene／codec／loop 失敗時
  stderr 診斷並安全降級 `off`；不會把舊 player 先關掉。
- 測試：`internal/ui/audio` 新增 mode、manifest provenance、路徑／雜湊、壞 Ogg、loop
  reader 與既有 crossfade 測試；`cmd/dsds` 在 Docker＋Xvfb 通過。`go.mod` 鎖定
  `oggvorbis v1.0.5`、`vorbis v1.0.2`、`lz4/v4 v4.1.21`。

儲存庫刻意沒有 placeholder `.ogg`，也沒有把 `workplace/audio/midi` 或原版 MUS/TIM
轉成 Modern 成品。剩餘 Modern 音樂工作是取得可追溯作曲／生成 prototype、人工相似性／
授權／loop／mono／響度／Android QA，完成後才可列入發行包；這不改變 `audio=retro` 或
純 Go／無 cgo 的既有邊界。

### 77. 2026-08-11 Modern 音樂首支 cue 原創草稿已交付

- **路線已鎖定：** 技術參考＋全新作曲。原版 AdLib／OPL2 只作硬體／情境參考；不採
  原版旋律、和聲、節奏、取樣或其衍生改編，也不把新曲稱作原聲帶。
- **已完成（草稿層）：** `docs/music/modern_scene/README.md`、`score.json`、
  `modern_scene_draft.mid`、`provenance.json` 與 `tools/music/modern_scene_draft.py`。
  草稿為 92 BPM、4/4、96 PPQN、8 小節，五音新動機 `D4–G4–F♯4–A4–E4`，只含鋼琴／
  低弦導引；決定性重建 SHA-256 為
  `38d4e862d01befdc2ba14dbbe73117c42bcd03589df9574390947db267589099`。
- **尚未完成：** 人耳三圈與旋律相似性審稿、loop 波形／混音、mono／響度／裝置 QA、
  指定自然人作者與音源授權。未完成前 MIDI 不進 playable asset 或 `audio=modern` manifest。
- **下一步：** 先審核 `modern_scene`，通過後再以獨立 stem 混音成 48 kHz Ogg；接著各自
  作曲 `modern_strategy`、`modern_battle_a`、`modern_battle_b`、`modern_story`、
  `modern_final`，不得複製首支 cue 填滿矩陣。三平台／Android 聽感驗收仍屬發行 gate。

本節把「作曲草稿已交付」與「可聽／可發行 Modern 音樂尚未完成」分開，避免把 runtime
接線或 MIDI 可解析誤報成音樂完成。

### 78. 2026-08-11 代理音樂驗證與實際遊玩推廣片

- **代理驗證已完成：** `modern_scene` MIDI 在 Docker 內以 FluidSynth 2.3.1 渲染為
  48 kHz／立體聲 Ogg；Vorbis、三圈、mono、RMS、響度與 true peak 均有可重現結果。
  代理 QA 為 `-18.1 LUFS`、`-5.9 dBFS true peak`、首尾樣本差 0；不冒稱真人聽感。
- **實際遊玩片已完成（本機預覽）：**
  `workplace/promo/gameplay/great-era-remake-gameplay.mp4`，1280×700、30 fps、24 秒，
  H.264＋AAC，SHA-256 `b0de0ad0fa0a72d98d6f7128c1e47c33c93b4e0d71f15322aef0b9786c9f95ed`。
  片中依正常入口展示復古→Modern、政略指令、將領查閱與人物自傳翻頁，並混入新的
  `modern_scene` 預覽音樂。
- **輸出邊界：** 影片、Ogg、MIDI render 與 SoundFont 都只留在被忽略的
  `workplace/promo/`；不得因「影片已能播放」就把原版資料／字型／音源或未清權 Ogg
  放入 GitHub／發行包。`docs/promo/README.md` 與 `storyboard.md` 已記錄分鏡與限制。
- **本輪閘門：** Docker＋Xvfb `go test -count=1 ./internal/ui/audio ./cmd/dsds` 通過；
  deny scan（600 檔案）、no-cgo、`git diff --check` 與 Docker 清理狀態均通過。
- **後續：** 其餘五首 cue／modern FX、正式作者／授權 metadata、真人或外部聽審（若日後
  需要）、三平台／Android 音訊與公開影片素材 clearance 仍待辦；本輪不重開 parity／SDFA。

### 79. 2026-08-11 高解析度 Modern／Android 規劃（目前工作焦點）

使用者要求先規劃高解析度版本，Android 以現代化高解析 UI 為基準；三平台桌面打包
暫停，不能因既有 `tools/package.sh` 存在就宣稱已交付。

- **已完成的盤點：** 640×350 是 retro 與現有 Modern shell 的設計畫布；`scale=2`
  只造成 1280×700 視窗，沒有高解析排版。滑鼠／觸控共用 `Action`，但 Android 48 dp、
  密度、安全區、旋轉、背景恢復與實機音訊尚待驗證。
- **新的 DRAFT：** `docs/design/33-high-resolution-modern-android.md` 定義 Modern
  design surface、響應式面板／人物自傳／敘事、比例字型、共用命中區、Android
  `ebitenmobile bind` 技術閘門與 H0–H6 分期；`docs/design/10`、`40` 已加交叉連結。
- **待使用者決定：** A：1280×720（16:9 橫向，建議）或 B：1280×800（16:10 橫向，
  平板垂直空間較多）。決定前不改正式 renderer，也不把文件狀態升為 `READY`。
- **Android 限制：** 官方 Ebitengine 文件列 Android `cgo required`；專案作者程式仍
  必須通過 `tools/check_no_cgo.sh`。先做工具鏈／NDK feasibility smoke，不能以桌面
  測試或 source no-cgo 掃描代替 Android 包與實機驗收。
- **暫停項目：** Linux／Windows／macOS 打包、Android APK／AAB、Modern 字型清權、
  五首 Modern cue 與裝置音訊 QA，均維持 release gate；本輪未執行打包與外部發布。

### 80. 2026-08-11 A 畫布確認、F3 解析度切換與配樂豪情修正

使用者已選定 **A：1280×720、16:9、橫向**，並要求三平台可在原版／高解析畫布間切換，
以及讓配樂更有《大時代的故事》的豪情。這一節取代 §79 中「待使用者決定 A／B」的
狀態；§79 的歷史規劃仍保留，不重寫原紀錄。

- **已接 H0：** `internal/ui/resolution`、`Preferences.Resolution`、`-resolution`
  旗標、`F3` 快捷鍵、`其他選項 → 解析度切換`、跨語系設定鍵，以及高解析 Surface
  的安全區／輸入反算。`F2` 仍只切換圖形主題；F3 不會碰戰鬥、規則、亂數或存檔。
- **H0 邊界：** `high` 現在是 1280×720 的安全等比 frame 橋接，方便先驗證三平台
  窗口／Surface 與命中座標；它不是最終的 Modern 向量／重排版面。H1 仍需完成
  1280×720 metrics、比例字型與 map／status／command／battle／biography／narrative
  的高解析截圖 gate。
- **輸入／Android：** 桌面 F3 與設定頁共用 `setResolutionPreference`；Android 無
  實體鍵盤時走設定頁。48 dp、密度、安全區、背景恢復、`ebitenmobile` binding 的
  cgo／NDK 與實機驗收仍未完成。
- **配樂：** `modern_scene` 維持原創 draft／技術 preview，不把它冒充正式豪情主題。
  新增 `modern_campaign`／推廣 cue 的方向列入下一個音樂切片：全新動機、軍政材料、
  低密度前奏→銅管／低弦／軍鼓高潮、可循環 stem 與完整 provenance；不得取用原版
  `.MUS`／`.TIM`、國歌、軍歌或可辨識旋律。
- **驗證：** Docker＋Xvfb 的 `cmd/dsds`、`internal/ui/resolution`、`prefs`、`i18n`、
  `render` 定向測試已通過；三平台打包仍暫停，沒有發布 APK／AAB／桌面包。

### 81. 2026-08-11 開工：H1-a Modern 高解析地圖／資訊卡

使用者要求開始已確認的 1280×720 A 方案。本輪採窄切片，不把所有頁面一次改成半套
高解析，保留 H0 fallback 作為穩定交付邊界。

- **已完成 metrics：** `internal/ui/layout/modern.go` 與 READY 規格
  `docs/spec/33-modern-high-resolution-layout-h1.md` 固定 1280×720、24 px 安全區、
  320 px 資訊卡、右側地圖卡、136×56 地圖入口；14×14 六角格以 3/2 顯示倍率保留
  48×36 格與 18 px 奇數欄位移。renderer／pointer 共用 `layout.Rect`。
- **已完成純 image renderer：** `internal/ui/render/modern_high.go` 使用既有 modern
  `Theme`／`PanelData`，不嵌入原版圖像或字型；有玩家自備字庫時顯示語系文字，無字庫
  仍可驗證地圖／卡片幾何。
- **已接玩家／預覽路徑：** Modern + high + `screenMap` 直接使用 H1-a renderer，
  其「指令／新聞」點擊座標也直接使用 1280×720 metrics；其他頁面維持 H0 的
  Surface→640×350 fallback。`cmd/screenshot -theme modern -resolution high` 已產生
  湖北 #26 的 1280×720 預覽，暫存 SHA-256 為
  `c6904a5f022486739e8351d835f7131ca7806faaa1369a66fa589f36978d42a9`。
- **尚待：** command／battle／biography／narrative 四類頁面的 H1 metrics 與截圖、
  比例字型清權、Android 48 dp／背景恢復／音訊實機 QA，以及三平台正式封裝；本輪沒有
  執行 `tools/package.sh`。
- **閘門：** Docker＋Xvfb 定向測試 `cmd/dsds`、`cmd/screenshot`、`internal/ui/layout`
  與 `internal/ui/render` 通過；高解析 Modern 啟動 smoke 正常 timeout；後續仍須保留
  deny scan、no-cgo、`git diff --check` 與容器清理紀錄。

### 82. 2026-08-11 H1-b 文件頁、觸控命中與 Modern procedural audio

- **已完成：** `ModernPageLayout`、`ModernCommandLayout`、`ModernOptionLayout` 與
  READY 規格 `docs/spec/34-modern-high-resolution-pages-h1b.md`；Modern + high 的
  政略 15 卡、顯示設定、解析度、其他選項、人物自傳、敘事圖庫已接 1280×720 renderer。
  renderer／pointer 共用 `Rect`，設定與自傳／敘事導覽維持既有 Action ID；戰鬥仍 H0。
- **文字／語系：** 高解析自傳沿用既有分頁狀態，以兩倍 bitmap 字模與留白提升可讀性；
  三套 wording 新增 `command.title`／`other.title`，英日繁中 key／字庫覆蓋測試通過。
- **已完成可播放 runtime：** `internal/ui/audio/procedural.go` 在 `audio=modern` 缺少
  manifest／scene 時提供八個 deterministic 原創 48 kHz stereo loop；通過 Ogg 仍優先。
  `Manager.PlayEffect` 提供 8 類短效果音與 12 player 上限，`cmd/dsds` 對鍵盤／滑鼠／
  觸控 Action 共用接線。這不等於正式作曲、作者／授權、人耳或 Android 音訊 QA。
- **新增規格：** `docs/spec/35-modern-procedural-audio-sfx-m1.md`；SPEC-32 的缺檔行為已
  改為 procedural fallback，不再把可播放 Modern fallback 降成 off。
- **驗證：** Docker Go wrapper 下 `internal/ui/audio`、`internal/ui/layout`、
  `internal/ui/render` 與隔離 Xvfb `cmd/dsds`、`internal/i18n` 通過。下一步重跑全 repo
  deny/no-cgo、`git diff --check`、產出 H1-b 五頁預覽；H1-c 戰鬥、正式 Ogg／授權、
  Android 48 dp／背景恢復／真機音訊與三平台包仍待辦。

### 83. 2026-08-11 H1-b gate 收束

- **全 repo 回歸：** Docker＋Xvfb `go test -count=1 ./...` 通過；deny scan 掃描 618
  個檔案且原版資產零命中，no-cgo／`git diff --check` 通過，專案容器已清理。
- **預覽：** `/tmp/modern-preview/menu-26.png` 為 1280×720 Modern 政略卡預覽，SHA-256
  `926917611f61bb5ca10afc48632bb3e37c497a9bed756a1ece5e17afb10c4615`；暫存輸出不進 repo。
- **狀態：** H1-a 地圖／資訊卡與 H1-b 指令／設定／自傳／敘事是 remake 已接玩家路徑；
  H1-c 戰鬥高解析、可再散布比例字型、Android 48 dp／背景恢復／真機音訊、正式 Ogg／
  作者授權、推廣片 clearance 與三平台包仍是未完成 gate。原版 oracle 未知不因本輪
  renderer／音訊測試而改寫。

### 84. 2026-08-11 H1-c Modern 高解析戰鬥 HUD

- **已完成：** `internal/ui/layout.ModernBattleLayout` 與 READY 規格
  `docs/spec/36-modern-high-resolution-battle-h1c.md`；左側 3/2 戰場、右側資源卡／
  五項命令／三控制鍵、撤退 3×4 鍵盤皆有共用矩形。
- **已接玩家路徑：** `cmd/dsds` Modern + high + `screenBattle` 使用
  `DrawModernBattleSurface`；戰場單位／游標／攻擊目標序號／資源／AI 摘要／log 只讀
  現有 battle state，命中直接派送既有 Action，不改 `BattleSim` 或 writer。
- **邊界：** 這是 remake HUD 完成，不是原版第 6 鍵、動畫、部署／撤退完整 oracle、
  AI parity 或 DT2 未解欄位完成；比例字型／Android 真機／音訊裝置／正式 Ogg／授權／
  三平台封裝仍待辦。
- **驗證：** layout／render／cmd/dsds 受影響測試與先前全 repo Docker＋Xvfb 回歸通過；
  修改語系的日文字元覆蓋測試也已修正並通過。

### 85. 2026-08-11 H1 全 repo gate（目前狀態）

- H1-a 地圖／資訊卡、H1-b 政略／設定／自傳／敘事、H1-c 戰鬥 HUD 已接上
  Modern 1280×720 玩家路徑；共用 layout 與 pointer，未改規則、Action ID 或存檔 writer。
- Docker＋Xvfb、`dsds-go:1.25`、既有 cache、`--network none` 下
  `go test -count=1 ./...` 通過；deny scan 掃描 620 檔案、原版資產零命中，no-cgo 與
  `git diff --check` 通過；`great-era` 容器無殘留。
- H1 暫存政略預覽 `/tmp/modern-preview/menu-26.png` SHA-256 為
  `926917611f61bb5ca10afc48632bb3e37c497a9bed756a1ece5e17afb10c4615`，不進 GitHub。
- **下一個工作焦點：** 比例字型／英日長文與 machine-draft overlay、Android
  `ebitenmobile`／NDK feasibility 及 48 dp／背景／音訊實機 gate、正式 Modern Ogg
  provenance／授權與人耳 QA。三平台封裝依使用者要求暫停；原版 oracle 與 parity 不因
  renderer 測試通過而升格。

### 86. 2026-08-11 H2 高解析共用流程與 Modern 豪情 cue（目前狀態）

- **已完成 remake 路徑：** Modern + high 的政略／查閱／外交／徵募／調度／交易／補給／
  確認／數字輸入頁面已經由 `modernFlowData`、`DrawModernFlowSurface` 與
  `ModernFlowLayout` 接到 1280×720；選項卡與 3×4 keypad 共用既有 Action，滑鼠與觸控
  不另造規則。`common.back`／`common.previous`／`common.next` 已加入三套 wording。
- **已完成可播放 runtime：** 純 Go procedural Modern cue 為六音原創動機的 24 小節
  進行（低音／脈衝／銅管／留白／高潮），八 track 與八類 FX 皆可 deterministic 產生；
  manifest Ogg 仍優先，缺檔時才使用 fallback。`docs/music/modern_campaign/` 的 MIDI／
  Ogg 仍是不可發行草稿，未宣稱正式作者、音源清權或人耳 QA。
- **回歸 gate：** Docker＋Xvfb `go test -count=1 ./...` 通過；deny scan 623 檔案、
  no-cgo、`git diff --check` 通過；專案容器無殘留。
- **仍待辦：** 英／日 387 篇人物機翻 overlay 的逐篇 provenance、日文字型與比例長文排版、
  Android `ebitenmobile`／NDK feasibility、48 dp／安全區／背景恢復／真機音訊、正式 Ogg
  混音與授權、人耳／裝置 QA、三平台封裝與推廣片。原版六攻擊／AI／撤退部署／DT1／DT2
  oracle 不因 Modern UI 或 runtime 音樂完成而改寫。

### 87. 2026-08-11 H2 將領詳情與人物自傳入口勘誤

- **修正：** `screenViewGeneral` 不再顯示泛用空白頁；Modern 1280×720 直接列出既有
  typed 將領資料與攻擊力摘要，三套 wording 新增 `view.general.*`。只讀省名／省況／
  將領詳情卡不送假的 Selection，返回仍由共用頁尾處理。
- **人物自傳：** 若 `currentBiography` 有資料，頁尾高解析按鈕送既有 `OpenBiography`，
  進入 H1-b 人物自傳 renderer；未知傳記仍保留既有 fallback，不把 machine-draft 當正式
  史料。
- **驗證：** 受影響套件與全 repo Docker＋Xvfb 回歸通過（`TEST_STATUS=0`）；deny scan
  623 檔案、no-cgo、`git diff --check` 通過，專案容器清理完成。

### 88. 2026-08-11 Android viewport M0 與 Modern 音訊離線預覽

- **已完成純 Go 行動幾何：** `internal/ui/mobile/viewport.go`／`viewport_test.go` 與
  `docs/spec/38-modern-mobile-viewport-m0.md` 固定 1280×720 design surface 的安全區、
  letterbox、density 反算及 48 dp invisible touch target；平台 binding 尚未宣稱完成。
- **已完成音訊 preview 入口：** `cmd/modern_audio` 可從同一份 runtime composition 產生
  `main-theme`／其餘 cue 與八類 FX 的 WAV 技術檔；`docs/music/modern_campaign/README.md`
  與 `provenance.json` 已記錄命令、格式與主題 hash。正式 Ogg、作者／音源授權、人耳／
  裝置 QA 仍保持獨立 gate。
- **已完成情境選曲與測試：** 地圖／政略主頁、敘事、人物／將領、戰鬥、離開確認各自
  選用對應 Modern cue；新增效果音 Action mapping 測試。Docker＋Xvfb 全 repo 測試與
  `go build` 通過，deny scan 629 檔案、no-cgo、diff check 通過，容器清理完成。
- **仍待辦：** Android 工具鏈（目前 image 缺 `ebitenmobile`／`gomobile`／`javac`／`adb`）、
  真機 lifecycle／音訊／48 dp 量測、可再散布比例字型、英／日人物 machine-draft overlay、
  正式 Ogg／授權與三平台封裝。這些 gate 未完成前不宣稱完整發行；原版 oracle／parity 不
  因 Modern runtime 增強而改寫。

### 89. 2026-08-11 高解析度 Modern 視覺驗收圖

- **新增成果圖：** Docker 內以 `cmd/screenshot -theme modern -resolution high -province 26`
  與 `-menu` 產生 1280×720 Modern 地圖／資訊卡與政略指令卡，檔案為
  `docs/images/high-modern-province-26.png`（SHA-256
  `c6904a5f022486739e8351d835f7131ca7806faaa1369a66fa589f36978d42a9`）及
  `docs/images/high-modern-menu-26.png`（SHA-256
  `36f30516656f884199d82303273b919f196a98a23a4ab34df6560853c8e6e386`）。
- **保護既有圖：** `docs/images/province-26.png`／`menu-26.png` 維持原復古成果，未以
  Modern 圖覆寫；展示圖位置符合 deny-list 白名單。
- **狀態：** Modern 高解析度畫面已可視化驗收；比例／向量字型、英日長文與 Android
  binding／實機 gate、正式 Ogg／授權／人耳 QA、三平台封裝仍未完成。

### 90. 2026-08-11 Modern Ogg runtime 有界抽測

- Docker＋Xvfb 直接執行預覽 manifest 時先遇到缺少 ALSA `default` 裝置；分類為環境／
  播放裝置問題，不把它寫成 Ogg 解碼缺陷。
- 在同一容器的暫時 ALSA null 裝置重跑 `-theme modern -resolution high`，程式持續到
  8 秒有界 timeout（`RUNTIME_EXIT=124`），無 manifest／Vorbis／procedural fallback 錯誤；
  這確認 preview Ogg runtime 接線，仍不代表人耳、Android 真機或正式授權完成。
- `workplace/promo/modern_audio` 只供技術預覽；正式 Ogg／provenance／混音與裝置 gate
  仍保持未完成，三平台封裝依使用者要求暫停。

### 91. 2026-08-11 GitHub README 狀態同步

- README 已改正 Modern 高解析 H1-a／H1-b／H1-c／H2、1280×720、`F3`／顯示設定解析度
  切換、滑鼠／觸控 Action 與 procedural audio fallback 的舊說法。
- README 已連結兩張高解析展示圖與 Android／Modern 規劃；比例字型、英日長文、Android
  實機、正式 Ogg／授權／人耳 QA 與三平台包仍明示未完成。

### 92. 2026-08-11 Modern procedural fallback 正常啟動抽測

- 在 Docker＋Xvfb、正確 `DISPLAY` 與隔離 ALSA null 裝置下，使用不存在的 Modern Ogg
  目錄啟動高解析遊戲；程式明確輸出 fallback 診斷並持續運作至有界 timeout，沒有
  GLFW／音訊 panic。
- 這閉合「缺 Ogg 時仍能玩」的 runtime 證據；正式 Ogg 清權、Android 真機與人耳 QA
  仍不宣稱完成。

### 93. 2026-08-11 高解析 Modern／音訊全套回歸收束

- Docker＋Xvfb 全 repo `go test -count=1 ./...`、三個命令建置、deny scan、no-cgo 與
  `git diff --check` 全部通過；掃描 631 個檔案、原版資產零命中，專案容器無殘留。
- 新增的 8 條 Modern cue／8 類 FX 邊界與 deterministic 測試，以及高解析指令／設定／
  自傳／敘事 renderer 測試已納入正式閘門。
- **仍待辦（不可誤報完成）：** 可再散布比例字型與英／日 387 篇長文 overlay；正式
  Ogg 混音、作者／音源 provenance、人耳／裝置 QA；Android `ebitenmobile`／NDK、
  lifecycle、48 dp／安全區／背景恢復／真機音訊；三平台封裝與推廣片。三平台依使用者
  先前指示仍暫停，原版 oracle／parity 也不因 renderer 或 procedural runtime 綠燈而改寫。

### 94. 2026-08-11 Modern cue／SFX 可重生 QA 報告

- `internal/ui/audio.AnalyzeStereoPCM` 與 `cmd/modern_audio -report` 已提供逐檔 JSON QA：
  sample rate、時長、峰值／RMS、mono RMS、首尾差、clipping、非靜音與 SHA-256。
- Docker 48 kHz 全量重生 8 cue＋8 FX（16 WAV）成功；`main-theme` 的最新 hash、峰值與
  RMS 已同步 `docs/music/modern_campaign/provenance.json`。輸出仍只在 `/tmp` 技術預覽，
  不進 GitHub／正式 manifest。
- **仍待辦：** 真人／盲聽、正式 Ogg 混音、作者與音源授權、Android 音訊裝置 QA；
  高解析比例字型／英日長文與 Android binding gate 仍未關閉。

### 95. 2026-08-11 Modern GEMF 字型 atlas 與長文 coverage

- **已完成技術字型路徑：** `tools/gen_modern_font.py` 在 Docker 的
  `mm-font-tools:dev` 以 Noto Sans CJK TC（SIL OFL 1.1）產生 `internal/assets/data/modern-font.bin`；
  2,391 glyph／95,660 bytes，來源／atlas hash／license 與重生命令見
  `docs/licenses/modern-font-atlas.md`。不散布 Noto 原檔或原版倚天字型。
- **已接 renderer：** `internal/assets.ParseModernFont`／`EmbeddedModernFont` 純 Go
  fail-closed parser；Modern 高解析文字優先用 GEMF 比例字距，缺少或損壞才退回玩家
  自備 Eten。復古 renderer 不讀 atlas。
- **已接 coverage gate：** `internal/assets/modernfont_test.go` 走實際 `PeopleDB`，
  英／日 387 篇 machine-draft 自傳正文 glyph coverage 綠燈；parser 邊界與基本字元
  也有測試。這不等於逐篇人審、日文禁則／長文截圖、正式 OFL notice 或裝置 QA。
- **仍待辦（不可誤報完成）：** 387 篇英／日翻譯逐篇 human review、長文版面與日文
  禁則／截圖；正式 Ogg／作者與音源授權／盲聽、人耳／Android 音訊 QA；Android binding、
  lifecycle／48 dp／安全區／背景恢復；三平台封裝與推廣片。三平台依使用者先前指示
  仍暫停，原版 oracle／parity 不因 GEMF 或 renderer 綠燈而改寫。

### 96. 2026-08-11 Android lifecycle M0 純 Go 契約

- `internal/ui/mobile.LifecycleGate` 與 `docs/spec/39-modern-mobile-lifecycle-m0.md` 已
  固定 active／background／destroyed、輸入拒絕、pointer／touch reset 與音訊一次性
  pause／resume／close effect；重複事件去重測試通過。
- 這只是 `ebitenmobile` adapter 的平台無關前置，不宣稱 Android binding、真機旋轉／
  背景恢復／音訊／48 dp 或 APK／AAB 已完成。

### 97. 2026-08-11 GEMF 後高解析展示圖與最後驗證

- Docker 內重產並更新 `docs/images/high-modern-province-26.png`（1280×720、SHA-256
  `725623c7f592da93376b5290b250a47629710cf49695528db91a0e89d02ea9a3`）與
  `docs/images/high-modern-menu-26.png`（1280×720、SHA-256
  `cb905520043fd1e1d0f7145cab41c9b01fa2f12a2f09c245aad0ec4f6d5b66d1`）；復古展示圖
  未覆寫，白名單位置合規。
- 新增 GEMF／lifecycle 後的全 repo Docker `go test -count=1 ./...`、三命令 build、
  no-cgo、deny scan（641 個檔案）與 `git diff --check` 通過；容器已清理。
- **仍待辦：** 逐篇英／日人審、日文禁則／長文截圖、正式 OFL notice／商標檢查；正式
  Ogg／作者與音源授權／盲聽與裝置 QA；Android binding／SDK／NDK／真機 lifecycle／
  48 dp／音訊；三平台封裝與推廣片。原版 oracle／parity 不因 GEMF、lifecycle 或截圖
  綠燈而改寫。

### 98. 2026-08-11 GEMF 資訊卡接線後最終回歸

- H1 高解析資訊卡／按鈕已改用 GEMF 優先、Eten fallback；最新兩張 1280×720 展示圖與
  hash 已在 §97／`CONTEXT.md` 5.130 登記，復古圖未覆寫。
- 最終 Docker＋Xvfb 全 repo 測試、三命令 build、no-cgo、deny scan（641 檔）與
  `git diff --check` 通過，容器清理完成。
- **收尾 gate：** 逐篇英／日 human review、日文禁則／長文截圖、OFL notice／商標檢查；
  正式 Ogg／授權／盲聽與 Android 音訊；`ebitenmobile` binding／SDK／NDK／真機 lifecycle、
  48 dp／安全區；三平台包與推廣片。原版 oracle／parity 狀態維持未知或近似，不因
  Modern renderer 綠燈改寫。

### 99. 2026-08-11 高解析人物自傳比例排版、Modern Ogg 預覽與 Android smoke 勘誤

- **人物自傳技術完成：** `LayoutProportional` 以 GEMF pixel advance 換行；高解析自傳
  不再依賴 Eten 才能啟動，Eten 只作 fallback。英／日最長 PeopleDB 正文的首／末頁、
  多頁與缺字 gate，以及比例 advance／標點／非法 metric 測試均通過。GEMF 2,391 glyph、
  95,660 bytes，SHA-256 `d0d6916622090df79e657eaba43255dcc0cbe5dcf329faa71da372c60e63717c`。
- **音訊技術預覽：** Docker 重生 8 cue＋8 FX WAV／JSON QA，再用 `u5cht/video:latest`
  的 FFmpeg 5.1.9 `libvorbis -q:a 5` 轉被忽略 Ogg；全數 48 kHz stereo Vorbis，
  `ffprobe` 與完整解碼通過。檔名／時長／重生命令在
  `docs/music/modern_procedural/README.md`；不把預覽當正式 Ogg、授權或人耳驗收。
- **Android smoke 勘誤：** `rich2-go-android:20260809` 有 `ebitenmobile`、SDK 35、
  NDK 27.2.12479018、`javac`、`adb`，但 `gomobile`／`ndk-build` 缺；取得工具模組後，
  目前 `./cmd/dsds` 因 `main` package 被 bind 拒絕。下一步是非 `main` `mobileapp` 邊界，
  本輪不產 AAR／APK、不把官方 cgo／NDK 需求混入作者 no-cgo 程式。
- **最終回歸：** Docker＋Xvfb 全 repo test、三命令 build、no-cgo、deny scan（644 檔、
  原版資產零命中、展示圖 4 張合規）、diff check 全綠；專案容器無殘留，展示圖 hash 維持
  `725623c7…`／`cb905520…`。
- **仍待辦：** 英／日逐篇 human review、日文禁則／OFL notice／商標、正式 Ogg provenance／
  盲聽／混音／裝置 QA；`mobileapp`／AAR／APK、Android 48 dp／安全區／背景恢復／真機音訊；
  三平台包與推廣片 clearance。原版 oracle／六攻擊／AI／撤退部署／DT1／DT2 parity 不變。

### 100. 2026-08-11 H1-c 高解析戰鬥展示入口

- **已完成：** `cmd/screenshot -theme modern -resolution high -battle` 改用共用
  `buildLiveBattle`／`game.NewBattleSim` 佈署快照，Modern 分支只讀整理面板／語系／資源，
  不執行攻擊、AI 或存檔寫回；復古 screenshot 路徑保持原狀。
- **成果圖：** `docs/images/high-modern-battle-26.png`（1280×720，SHA-256
  `19a3acb0ba0202a02320799fc92fa5c3dae717880fd5d9abef7e195d59d2fa80`）已加入 README、
  高解析 Android 設計與 H1-c 規格；目視確認戰場、攻守單位、資源卡、五項命令與三個
  Modern 控制鍵均可見。
- **驗證：** Docker＋Xvfb 全 repo test、三命令 build、high battle screenshot、no-cgo、
  deny scan（646 檔、原版資產零命中、展示圖 5 張合規）與 diff check 通過；容器清理完成。
- **仍待辦：** Android `mobileapp`／AAR／APK 與實機 48 dp／背景／音訊；正式 Ogg／作者
  provenance／人耳／授權；英／日逐篇人審；三平台包／推廣片。原版 oracle／AI／撤退部署／
  DT1／DT2 parity 狀態維持未知或近似。

### 101. 2026-08-11 高解析 Modern 正常玩家啟動抽測

- Docker＋Xvfb 以正常 `cmd/dsds`（不是展示工具）啟動 `-theme modern -resolution high
  -audio off`，使用專案既有 Go cache，持續 12 秒後由有界 timeout 結束（`game_rc=124`）；
  沒有 panic／GLFW 啟動錯誤。
- 執行中擷取的暫存根畫面由 `identify` 確認為 1280×720；暫存檔未進工作樹與 PNG 白名單。
  Xauthority 與既有 `NEWSDATA.DAT` 尺寸訊息分類為環境／資產 fallback，不是高解析 renderer
  缺陷。
- 這只閉合高解析 Modern 正常入口的持續執行證據；音訊沿用 §92／§93 的 procedural／Ogg
  技術結果。Android binding／實機、正式 Ogg／授權／人耳、英／日人審與三平台包／推廣片
  仍是未完成 gate，原版 oracle／parity 不變。

### 102. 2026-08-11 高解析人物自傳展示入口

- `cmd/screenshot` 新增 `-biography`（限定 Modern／high），由現有 `PeopleDB`／省份司令
  取得人物資料，再呼叫 `DrawModernBiographySurface`；wording、來源提示與 fallback 都
  走語系 catalog，不另造資料或規則路徑。
- Docker 產生 `docs/images/high-modern-biography-058-01.png`（1280×720，SHA-256
  `87769559d02f6f819aceaeb5b0180f95ba65f8deafa9490c80e15f4e7112dc02`），目視確認人物名、
  生平正文、可靠度與上一／下一頁控制均可讀；#058 正文為單頁，長文多頁仍由測試驗證。
- README、Modern／Android 高解析規劃與 H1-c 驗收文件已加入命令／成果圖。尚未關閉的
  gate 維持：正式 Ogg／作者授權／人耳 QA、英／日逐篇人審、Android binding／真機與三
  平台包；原版 oracle／parity 不變。

### 103. 2026-08-11 高解析 Modern 推廣預覽整合

- 在影音 Docker 以既有正常玩家錄影保留復古→Modern→人物流程，再串接四張高解析成果圖，
  以 `modern_campaign` 豪情 cue 產生 ignored
  `workplace/promo/gameplay/great-era-remake-high-preview.mp4`。
- 影片 1280×720／30 fps／36 秒／AAC 48 kHz stereo，SHA-256
  `7236c91b039caf8a952c76fd7896d563efe1947799aeeecc26dc1206c99de0c8`；`docs/promo/README.md`
  與 `storyboard.md` 已記錄前 24 秒實際遊玩、後 12 秒成果頁串接的界線。
- 這閉合本機「可看到高解析畫面並聽到 Modern 豪情 cue」預覽，不等於正式 Ogg／授權／
  人耳或 Android 音訊驗收；三平台發行與公開影片 clearance 仍待。

### 104. 2026-08-11 高解析／Modern／音效目標完成稽核

- 直接以 `cmd/dsds -theme modern -resolution high -audio modern` 在 Docker＋Xvfb＋
  ALSA null 裝置啟動 12 秒；正常入口明確走原創純 Go Modern 音樂 fallback，沒有 panic
  或初始化錯誤。這不是 `cmd/screenshot` 展示捷徑。
- 同一工具鏈的全 repo `go test -count=1 ./...`、三命令 build、8 cue／8 FX＋JSON QA、
  no-cgo、deny scan（646 檔、6 張展示圖合規）與 diff check 全綠，專案容器清理完成。
- **完成判定：** 高解析 Modern UI、主題／解析度切換、人物自傳與敘事流程、音樂與 8 類
  SFX 已達 remake 可玩完成；正式 Ogg 清權／人耳／Android／三平台是獨立發行 gate，
  不再把它們誤列為本目標的未完成核心功能，也不重開 DOSBox／SDFA parity 迴圈。

### 105. 2026-08-11 不含原版遊戲資料的 Linux release 預覽包

- `tools/package.sh` 新增目標清單參數；已用 Docker 執行
  `tools/package.sh 0.1.0-release "$PWD/dist" linux-amd64`，產生
  `dist/great-era-remake-linux-amd64-0.1.0-release.tar.gz`（約 4.1 MiB，SHA-256
  `892c7aca1e499caa7872c4e88bc4a63375b2aa5d0fdc0dad10d90d6039b6cb20`）。
- 包內含 `dsds`、`translations/`、Modern atlas 的 `OFL-1.1.txt`／來源說明、release／
  推廣說明與 `RELEASE-MANIFEST.txt`；沒有 `workplace`、原版副檔名、原版音樂或任何
  衍生 `.MUS`／`.TIM`／`.DAT`／Ogg／WAV。manifest 明確標記
  `contains_original_game_data=no`，並記錄每檔 SHA-256。
- Docker smoke：tar 清單拒絕規則通過，解包後 Linux `dsds` 可執行，Xvfb 下 `dsds -h`
  回傳 0；正常遊玩仍需玩家自備合法 `-game`／`-eten`。Windows／macOS／Android 與
  公開發行 clearance 仍不宣稱完成。

### 106. 2026-08-11 目前狀態與下一個 gate

- **已完成（本輪 scope）：** 高解析 Modern 已採用「策略儀表板」而非放大地圖；六張
  真實資料卡（本回合指令、兵力、將領、黃金、糧食、忠誠度）、指令卡、人物自傳與戰鬥
  路徑皆可由正常滑鼠操作抵達。指令標題與實際規則代號的映射已有回歸測試，避免畫面
  誤把 `查閱` 顯示為「秘密行動」。新版成果圖與 60 秒連續實機推廣片均已產出。
- **已完成（封包結構）：** `0.1.0-release` 的 Windows ZIP、macOS app ZIP、Linux
  AppImage 已重建；ZIP 入口／CRC 與 AppImage 解包／素材排除已通過。精確檔名與雜湊見
  `docs/release/README.md`。
- **下一個發行 gate：** 不重開 Modern 介面設計；待有對應平台時，依序執行 Windows
  正常玩家 smoke、macOS Apple Silicon smoke 與簽署／公證、不同 Linux 發行版的
  AppImage smoke。這些都是發行驗收，不能由 Docker 交叉建置代替。
- **仍為獨立 gate：** 正式 Modern Ogg 的作者／授權／人耳與裝置 QA、英日人物稿人審、
  Android `mobileapp`／AAR／APK 與真機 lifecycle／48 dp／音訊驗證。原版六種攻擊、AI、
  撤退部署、DT1／DT2 未解時機維持既有「近似／未知」界線，不因本輪 UI 與封包工作重開
  parity 研究。

### 107. 2026-08-11 Modern UI 重新規劃前沿

- **已撤回的驗收：** H3 策略儀表板的規則、命中區與滑鼠／觸控接線可保留，但其現有
  視覺與排版不再是最終 Modern UI；使用者判定它太乾燥、仍近似放大舊版。
- **現在唯一的設計前沿：** 先由使用者選擇 Modern 的根層資訊架構（持續可見的指揮室、
  省份作戰板或沉浸式地圖），再把頁面、戰鬥 HUD、人物自傳、Android 響應式斷點與資產
  系統拆成 READY 規格。確認前只可做可丟棄 prototype，不得在既有 H3 上任意堆卡片。
- **推廣片：** 現存 60 秒影片只作連續玩家路徑基準；新 UI 的候選片必須重錄完整 remake
  遊玩，不能用靜態成果圖／mockup 代替。所有候選包、雜湊與推廣影片現集中在被忽略的
  `dist-all/`，發行包與 `dist-all/promo/` 保持分離。

### 108. 2026-08-12 已確認 H4「全域指揮室」與人物肖像素材政策

- **已決：** 採用 A「全域指揮室」，排除 B「省份作戰板」與 C「沉浸式地圖」作為根層
  資訊架構。下一份 H4 規格需拆出持續導覽、中央地圖、脈絡檢視器、底部指令／事件、戰鬥
  HUD、人物檔案與 Android 響應式斷點；保留 H3 的 Action／命中區／資料接線，不在它上面
  疊卡片修飾。
- **已決：** 使用者選定 A+B：可驗證公有領域／CC0 與具完整署名資訊（attribution）的
  CC BY／CC BY-SA 離線素材；排除外部連結方案。下一步以 `docs/spec/40` 建立語系無關
  的登錄、雜湊與發行 notices gate，先以少量高信心人物驗證全鏈；沒有登錄資料時仍顯示
  中性檔案卡後備畫面（fallback），不下載未驗證圖片，也不用 AI 臉孔補洞。

### 109. 2026-08-12 H4 軍事衛星作戰室 M1

- **已接：** `SPEC-41` 的主地圖、導覽／狀態軌、軍情檢視器、操作／事件列、衛星 palette、
  人物檔案閱讀欄與現有 `OpenCommands`／`OpenNarrative` 同義命中區，以及 `SPEC-42` 的
  5×3 政略指令矩陣。H3 六卡畫面仍只作歷史回歸，不得重新拿來當 H4 視覺驗收。
- **下一個 UI 切片：** 依相同外殼逐頁檢視政略流程、設定、敘事與戰鬥資訊層的密度；只在有
  正常玩家路徑、對應 action、固定截圖與觸控矩形驗證後，才可把它們列入新推廣片。Android
  實機、Modern Ogg 與三平台重新打包保持獨立待辦。

### 110. 2026-08-12 H4 交付候選包與實機推廣片

- **已完成本項交付：** `dist-all/2026-08-12-h4-m1/` 集中保存 Windows amd64 ZIP、macOS
  arm64 app ZIP、Linux x86_64 AppImage、各檔雜湊與 `promo/`。ZIP 的全檔 CRC／入口、
  AppImage 解包／禁止素材掃描、全部雜湊，以及 H4 推廣片的 H.264／AAC 串流都已在受限
  Docker 重新檢查。
- **影片已完成：** 66.176 秒連續錄影可見復古→Modern→H4、5×3 指令矩陣、人物自傳與肖像、
  河南戰鬥、滑鼠攻擊與結束回合；沒有靜態畫面剪接。`tools/capture_h4_promo.sh` 已固定
  X11 聚焦後送鍵流程，避免指定視窗事件漏送造成黑畫面。
- **不要重開成 parity loop：** 本次封包／影片不關閉 Windows／macOS／Linux 真機 smoke、
  macOS 簽署／公證、正式音樂權利／人耳 QA、Android、英日人物稿人審或原版 mechanics
  parity。後續應依這些獨立 gate 前進，不得把已完成的 H4 正常玩家展示退回成舊版 UI 或
  無限 DOSBox 研究。

### 111. 2026-08-12 Modern A2 明亮土黃 M1

- **已決並已接：** A2 取代深藍黑衛星色組作 Modern 主視覺；採明亮土黃紙面、朱紅重點、
  深褐文字。省名、短標題、司令姓名與可容納的短按鈕標籤已放大，長字串仍受同一文字安全
  矩形限制。
- **已接：** `MENU1..5.TPC`／`WARMENU.TPC` 由玩家原版目錄在執行期解碼並重著色為按鈕／
  面板裝飾，不進 repo 或 release；缺檔時使用程式化雙框折角。地圖、導覽、政略矩陣、流程、
  keypad、一般導覽與戰鬥控制已共用裝飾 primitive。
- **已接：** 區域司令官固定 108×148 肖像槽，與人物自傳共用 `PortraitCatalog`。省 27 正常
  資料抽樣顯示蔣中正已驗證照片；無照片人物維持相同檔案卡版面，不使用 AI 或原版頭像。
- **下一個 UI gate：** 逐頁檢視英／日長詞、人物檔案、敘事、設定與戰鬥的 A2 密度，修正仍
  偏小或留白過多的欄位；完成後再更新版本化成果圖、重錄連續遊玩推廣片並重建 `dist-all/`
  三平台候選包。舊 H4 包與影片是歷史候選，不得冒稱 A2 發行成果。
- **mentor 複核已補：** 便宜模型完成受限實作與只讀審稿後，主代理以三語系四類正常資料截圖
  找出英文指令的 16 px 退化與無提示斷字，再要求 follow-up。指令／選項主要標籤現固定 2 倍
  字級，空白詞組最多兩行，超長詞用省略號；戰鬥／人物檔案／敘事標題已放大，司令照片保持
  crop 長寬比。pointer／touch 矩形、action、規則、存檔與翻譯均未改。
- **剩餘 gate 重分類：** A2 桌面高解析 M1 的頁面裝飾、主標題、指令長字與司令照片比例已閉合；
  Android viewport／真機、版本化成果圖、連續推廣片及三平台重包仍是後續發行 gate，不應再以
  舊 H4 深色畫面或舊包冒稱完成。

### 112. 2026-08-12 A2 高解析字形與圖片

- **已決／已接：** Modern 採 C 字體分工：主標題宋體；按鈕、內文、數值與導覽黑體。
  兩份 Noto CJK 來源已離線重生為 32×32、8-bit alpha 的 GEMF v2，renderer 使用純 Go
  coverage 合成；GEMF v1 parser 保留相容，`retro` 不讀新 atlas。
- **已接：** Modern 指令／資源圖示以目的尺寸、4 倍 supersampling 重繪，不再放大 16×16
  圖示；人物照片依放大／縮小分流使用雙線性／面積取樣。原版花框刻意保持像素銳利，且仍
  只存在玩家自備資料的執行期遮罩。
- **已接：** Modern 地形、21 種鐵路接線與 18 個部隊／砲兵方向另有目的尺寸的 4 倍
  supersampling 路徑；高解析 renderer 優先使用它，舊索引圖只保留原解析度相容。
- **驗證：** 雙字體重生逐 byte 一致；Docker＋Xvfb 全 repo test、no-cgo、deny scan
  （668 檔、原版資產零命中、6 張展示圖合規）通過；繁中政略頁與人物自傳正常資料截圖已
  在 `/tmp` 人工檢視。
- **後續發行 gate：** 版本化成果圖、A2 連續玩家推廣片與三平台重包仍待；Android 真機、
  正式音樂權利／人耳與英日人審維持獨立 gate，不重開已閉合的高解析字形與圖片路徑。

### 113. 2026-08-12 無頭像決定與 Modern 視覺第二輪

- **已決／已接：** Modern 全面無頭像；地圖與人物自傳不顯示照片、剪影或缺圖卡，正常玩家
  路徑不再載入 `PortraitCatalog`。吳佩孚的舊無圖狀態明確訂正為登錄缺口，不是沒有歷史照片。
- **已接：** 人物自傳使用完整文件欄；軍情面板保留司令／省長姓名與六張摘要卡，不留肖像空洞。
- **已由後續決定取代：** 第一版設計稿因可見性與代表性不足退回；使用者其後選定第二版 A，
  正式實作與驗證結果見 §114。此項不再是待確認前沿。
- **肖像決定再被覆寫：** 使用者最新要求 Modern 必須有人物頭像；受控肖像恢復與交付結果見
  §115。保留本節只為追溯決定歷史，不可再拿「全面無頭像」阻擋目前玩家路徑。

### 114. 2026-08-12 民國測繪風 M2 已進正式實作

- **已確認：** 使用者選 A，第二版地圖與指令頁同時採用；`SPEC-44` 已 READY。
- **已完成：** 15 個既有指令各自一個純 Go 專屬圖示，高低解析語意一致且不再五種循環；
  Modern 高山、丘陵、農地、水系、森林、沙漠、城市、關口、長城、橋、鐵路已按測繪語彙
  重繪。長城十種變體與地物 22 的 unknown fallback 均有測試，22／21／18 索引契約不變。
- **驗證：** 正常資料產生 39 省 1280×720 地圖，人工抽查多省及省 26 的地圖／指令／戰鬥；
  Docker＋Xvfb 全 repo test、no-cgo、deny scan（670 檔、原版資產零命中）與 diff check 通過。
- **不在本切片：** 新玩法、命令更名、規則／存檔／命中區改動、推廣片重錄與三平台重包。

### 115. 2026-08-12 A2 M2 受控肖像、三平台包與推廣片

- **已接：** Modern 地圖與人物自傳重新顯示 `PortraitCatalog` 通過 gate 的照片；目前蔣中正
  有公有領域照片，其他人物維持明示未登錄的檔案卡。地圖摘要卡、人物正文與肖像欄固定不重疊。
- **正常畫面修正：** 首次影片抽幀發現玩家原版厚花框會壓住自傳頁 metadata 與正文左緣；
  已加大文字安全內距並以相同原版 runtime ornament 截圖重驗，不把較乾淨的程式化 fallback
  截圖冒充實機完成證據。
- **交付：** 新版 Windows amd64 ZIP、macOS arm64 app ZIP、Linux x86_64 AppImage 與
  61 秒同一視窗連續遊玩影片集中於 `dist-all/2026-08-12-a2-m2-portraits/`；影片涵蓋復古→
  Modern、民國測繪地圖、司令頭像、查閱、人物自傳大圖與戰鬥。正式公開音樂權利、人耳 QA
  與三平台真機 smoke 仍是獨立 gate。
