# CONTEXT — 專案脈絡與文件索引

> **這份是全專案的單一入口。** 對話被壓縮、或換一個新 session 接手時，先讀這份，
> 就能重建完整全局，再依索引跳到需要的文件。硬規則與工作紀律在
> [`CLAUDE.md`](./CLAUDE.md)。
>
> 最後更新：2026-08-01（M0 資產解密完成、Go 解碼層上線、M2 反組譯起步）

---

## 1. 這個專案在做什麼

把漢堂國際資訊《大時代的故事》(1992, DOS) 完整逆向，在 Go / Ebiten 上乾淨重寫引擎，
還原原版繁中文本並另做英文與日文語系。定位是文化資產保存。
目標、原則、里程碑見 `CLAUDE.md`；原作背景與外部資料索引見 `CLAUDE.md` §1.5。

---

## 2. 現況一覽

| 領域 | 狀態 |
|---|---|
| 原版素材 | ✅ 已搬進 `workplace/orig/`（gitignore）：`game/` 148 檔 + js-dos zip |
| 發行資料 | ✅ 漢堂國際資訊，1992-10-15，回合制戰棋（`CLAUDE.md` §1.5，維基百科）|
| 首輪偵查 | ✅ 執行檔鏈、packer、BGI 繪圖、資產檔頭（`CLAUDE.md` §3）|
| **`.TPC` 圖像** | ✅ **confirmed**：BGI `getimage`，逐列交錯 plane、plane 0 為最高位。**逐像素 round-trip 通過**；`+20` 差額 = `palt` |
| `.15` 字模格式 | ✅ **已驗證**：16×15、30 B/字，51 檔全部整除 30 |
| **文本儲存機制** | ✅ **已解**：不是索引，是**定長槽位的字模序列**。見 `docs/formats/01-glyph-text.md` |
| **詞表還原** | ✅ **已完成**：51 檔 6,174 字模，倚天命中 4,799、空白 1,374、例外 1（位移的逗號）|
| **敘事文本** | ✅ **是畫成圖的**：`NEWSDATA.DAT` = 17 條新聞橫幅（215×16 BGI 圖）。不是字模序列 |
| `.RGB` 調色盤 | ✅ 純文字，16 行 RGB **百分比**（不是 DAC）。見 `docs/formats/04-tpc-bgi-image.md` |
| **`.GLB`／`.GTB` 圖庫** | ✅ **confirmed**：PKWARE DCL implode + 13 B 目錄，45 筆全解。**逐像素 round-trip 通過**。見 `docs/formats/02-glb-image-library.md` |
| **執行檔反組譯（M2）** | 🔶 4/5 進 IDA（`SDFA` 解不開）。**戰鬥模組已定位**：`sub_39B6E` ← `PROGRAM+537`，主迴圈在 `+18F`–`+6F4`。見 `docs/re/04` |
| **`PLACD.SAV`** | ✅ **是 ZIP**：整包遊戲的 1993 年備份，145/147 檔與 `game/` byte-identical。見 `docs/formats/03-placd-archive.md` |
| `.MUS`／`.TIM` | ⬜ 未開始；音源未確認 |
| **`.DAT` 資料表** | 🔶 **39 省是核心維度**：`TERNAME`／`WARPOS` 每省 196 B、`MEM_WAR` 每省 469 B。`MAN(N).DAT` 33 B/人。欄位語意未解 |
| 存檔 `.DT1`/`.DT2` | 🔶 `.DT1` **整塊讀進記憶體**（recsize=14683），不是省份陣列；湖北的 6 個資源在 929 起連續 u16。`.DT2` = 39×469，語意未解 |
| **DOSBox oracle** | ✅ **跑得起來**，可走到「選擇歷史背景」。見 `docs/playtest/01-dosbox-probe.md` |
| **顯示模式** | ✅ **兩種並用**：開場／標題 320×200 256 色（mode 13h）、主遊戲 640×350 16 色（BGI）|
| 防拷／磁片檢查 | ✅ **有密碼但過得去**：載入存檔後問說明書第 36 頁，三選一，選 `1` 通過。磁片提示只是 I/O 失敗的後備路徑 |
| **Go 解碼層** | ✅ `internal/assets`：DCL、BGI、`.RGB`、`.15`、`.GLB`、**省份地圖**。測試全綠，**輸出對實機逐像素相同** |
| **政略規則（M3）** | 🔶 指令 1=調動、2=軍事/攻打、6=查閱（其餘 12 個未定）、命令數上限、13 個省份欄位。見 `docs/playtest/02` |
| **地圖／地形（M3）** | ✅ **SPEC-01 READY**：39 省鄰接表（對稱性 + 實機雙重驗證）、每省 14×14 戰場格子。地形編號 0–22 語意未解 |
| **將領（M3）** | ✅ **SPEC-02 READY**：33 B/人、三個能力值、兵力 u16@17、**省兵力 = 將領加總**（實機 97500 零誤差）|
| **Go 規則層（M4）** | 🔶 `internal/game`：地圖層（SPEC-01）+ 將領層（SPEC-02，274 人 byte-for-byte round-trip）|
| Ebiten 呈現層（M5）| ⬜ 未開始 |
| 多語系 | ⬜ 未開始。字形來源已定案：倚天點陣字（使用者指示 2026-08-01）|

---

## 3. 文件索引

| 路徑 | 內容 | 目前 |
|---|---|---|
| `docs/formats/01-glyph-text.md` | **`.15` 字模與文本儲存機制**（定長槽位、51 檔用途、例外記帳）| ✅ |
| `docs/formats/02-glb-image-library.md` | **`.GLB`／`.GTB` 256 色圖庫**（DCL implode、調色盤、透明色、round-trip）| ✅ |
| `docs/formats/03-placd-archive.md` | **`PLACD.SAV` = 1993 年整包 ZIP 備份**（真實年代、逐檔比對）| ✅ |
| `docs/playtest/01-dosbox-probe.md` | DOSBox 探路（兩種顯示模式、磁片檢查範圍）| ✅ |
| `docs/formats/glyph-tables/` | 還原出的字表與詞條（`_slots.json` 為完整詞條）| ✅ |
| `docs/re/01-war-exe-turbo-pascal.md` | **`WAR.EXE` 是 Turbo Pascal**；字模槽寬、圖集分割、雙磁碟路徑、磁片提示 | ✅ |
| `docs/re/02-grt-exe-borland-c.md` | **五支執行檔盤點**：編譯器、角色、PKLITE 自動解包、DCL 字串佐證 | ✅ |
| `docs/re/03-war-exe-is-the-game.md` | **`WAR.EXE` 是遊戲本體**；39 省資料表結構、存檔 `.DT2` = 39×469 | ✅ |
| `docs/re/04-battle-module.md` | **戰鬥模組定位**：`sub_39B6E` 主函式、呼叫圖、32 個函式清單 | ✅ |
| `docs/playtest/02-load-save-strategy-phase.md` | **政略畫面實測**：13 個省份欄位、15 個指令、命令數上限、密碼可過 | ✅ |
| `docs/spec/01-map-and-terrain.md` | **地圖規格**：39 省鄰接表（READY）、14×14 戰場格子（READY）、地形編號（DRAFT）| ✅ |
| `docs/spec/02-generals.md` | **將領規格**：33 B 記錄、三個能力值、兵力 u16、省兵力 = 將領加總（READY）| ✅ |
| `docs/playtest/` | 實跑驗收紀錄與截圖 | 空 |
| `docs/reference/` | 外部中文資料整理（來源 + 抓取日期）| 空 |
| `translations/glossary.md` | 譯名表（繁中母本 → en / ja）| 骨架 |

---

## 4. 術語表

| 詞 | 定義 |
|---|---|
| 定長槽位 | 原版存文字的方式：每 `w` 個字模為一個詞條，不足補全零字模。`w` 依檔案固定 |
| 排版填充 | 槽位放不滿時的空白。對齊方式依檔案而異：靠左（地名）、靠右（日將名）、分散（中將名）|
| 數字佔位 | `FAN(2)/(3).15` 詞條**中間**的空白，執行期填入 `1.15` 的數字字模（「第␣集團軍」）|
| 場景字模檔 | 51 個 `.15` 族檔案。每個場景／類別各帶一份自己的詞表 |
| 字模索引 | ⚠️ **已推翻的說法**，見 §5。原本以為文字存索引，實際字模序列本身就是文本 |
| BGI 緩衝 | Borland `getimage` 的記憶體格式，`.TPC` 用的就是它（4-plane、16 色 planar）|
| DCL implode | PKWARE Data Compression Library 的壓縮法，`.GLB` 每一筆各自是一段。header `00 06` |
| 透明色索引 | `.GLB` 圖像的索引 0 不寫入，露出底圖。書法標題就是這樣疊上去的 |
| DAC 6-bit | VGA 調色盤是 6-bit（0–63）。轉 8-bit 要用 `(v<<2)\|(v>>4)`，不是 `v*255/63` |
| 命令數 | 政略階段每個省份可下的指令上限（社群資料，待驗）|
| oracle | 判斷「原版怎麼做」的證據來源，優先序見 `CLAUDE.md` §4 |

---

## 5. 已被推翻的斷言

### 5.1 「文字以字模索引儲存」→ 實為定長槽位的字模序列

- **原結論**（`CLAUDE.md` §3.5、§5）：文字以「第 N 個字模」的索引儲存，
  畫面上的中文是查表畫出來的。
- **當初憑什麼**：五個執行檔、`GRT.GLB`、所有 `.DAT` 用 Big5 解碼命中數為零，
  加上字模檔是「每場景一份子集」，推論出中間有一層索引。**推論方向對，但多推了一層。**
- **新證據**：51 檔全量反查後，字模序列直接讀出來就是通順的詞條
  （`蔣中正 何應欽 馮玉祥`、`調動 軍事 運補 徵稅`），且空白的分布完全符合定長欄位的
  排版填充。**沒有索引那一層**——程式從 `k × w × 30` 直接取 bytes 送 BGI 即可。
  Big5 零命中的原因不是「存了索引」，而是**原始碼裡本來就沒有碼點**。
- **受影響**：`CLAUDE.md` §3.5 假說、§5 還原鏈步驟 2（「索引來源」對這 51 檔不適用）。
  已在 `docs/formats/01-glyph-text.md` §6 記錄。
- **仍成立的部分**：`1.15` 的數字字模被 `FAN(2)/(3).15` 的佔位引用，
  這是目前唯一觀察到的索引式取字。敘事文本若在 `NEWSDATA.DAT`，仍可能是索引流。

### 5.2 「46 個字模檔」→ 實為 51 個

- **原結論**（`CLAUDE.md` §3.4）：46 個 `.15` 族檔案，清單列出
  `TN15.1`–`39`、`FAN(1)`、`FAN(2)`、`JAPAN1`、`MAN115/215/315`。
- **新證據**：完整列舉後是 **51 個**，原清單漏了 `FAN(3).15`、`JAPAN2.15`，
  且沒列 `1.15`–`4.15` 這四個。
- **為什麼會漏**：`TN15.9` 的副檔名是 `9` 不是 `15`，用 `ls *.15` 或
  `grep '\.15$'` 篩會篩掉大半——**過濾器自己有洞**。
  這正是 `rulebook/83`「同類檔先列舉全再做 pipeline」的踩坑點。
- **受影響**：`CLAUDE.md` §3.4 已更正。

### 5.5 「`GRT` 是政略主程式、`WAR` 是戰鬥模組」→ 反了，`WAR` 是遊戲本體

- **原結論**（`CLAUDE.md` §3.7）：`GRT` 政略主程式、`WAR` 戰鬥模組
  （375 KB 最大，且戰棋的戰鬥系統最重）。
- **當初憑什麼**：`play.bat` 的順序 `grt → war`，加上「戰鬥系統最重」的直覺。
  **兩者都只是類比，沒有證據。**
- **新證據**：比對每支引用的資料檔字串——`WAR.EXE` 引用 80 個以上
  （全部字模、全部 `.RGB`、`nwmap`／`tername`／`warpos`／`mem_war`／`NewsData`、
  以及 `save(1).dt2` 到 `save(10).dt2`），另外三支各自只引用兩三個檔。
  `GRT`／`GRTE` 的字串表全是資源載入與記憶體管理，沒有任何政略或戰鬥的東西。
- **受影響**：M3 規則規格要挖的東西全部集中在 `WAR.EXE` 一支。
  `CLAUDE.md` §3.7 已更正，詳見 `docs/re/03-war-exe-is-the-game.md`。

### 5.3 「繪圖是 BGI，與 mode 13h 完全是兩套」→ 兩種模式並用

- **原結論**（`CLAUDE.md` §3.2）：遊戲用 BGI 繪圖，640×350 或 640×480 16 色 planar，
  「與大富翁2 的 mode 13h chunky 完全是兩套」。
- **當初憑什麼**：`EGAVGA.BGI` 存在，且 `.TPC` 經驗算確實是 BGI `getimage` 的
  4-plane 16 色格式，`w=640` 也對得上。**這些證據本身沒錯，錯在外推。**
- **新證據**：DOSBox 實測用 `xdotool getwindowgeometry` 量到——標題階段視窗是
  **320×200**、單張畫面 217 個獨立色（mode 13h 256 色）；進主選單才切成 **640×350**、
  5–16 色（BGI）。兩種模式都在用。
- **為什麼會錯**：從「`.TPC` 是 BGI 格式」推到「整個遊戲都是 BGI」，
  正是 `CLAUDE.md` §7 第 9 條的「不要跨檔案外推」。`.TPC` 只證明了 `.TPC` 自己。
- **受影響**：`CLAUDE.md` §3.2 已改寫；remake 呈現層要支援兩種模式。
- **順帶的線索**：開場的 256 色圖不可能是 `.TPC`，存在別處；
  `GRT.GLB` 熵 7.96 bit/byte 一直沒解釋，壓縮過的 256 色圖是合理候選。

### 5.4 「`WAR.EXE` 是 Turbo C 編譯的」→ 是 Turbo Pascal

- **原結論**（`CLAUDE.md` §3.1）：`WAR.EXE` 未打包，Borland Turbo C runtime。
- **當初憑什麼**：檔案裡有 `Copyright (c) 1983,90 Borland` 字串。
- **新證據**：IDA 自動辨識出的 RTL 全是 Turbo Pascal——`BlockRead`、`Concat`、
  `Copy`、`Insert`、`ASSIGNCRT`、`IOResult`，以及 **Graph unit**（`CLOSEGRAPH`、
  `BAR`、`CIRCLE`、`DRAWPOLY`）。字串也是 Pascal 長度前綴形式
  （`db 33,'Please remove Disk MARK-A from A:'`）。
- **為什麼會錯**：Borland 的版權字串在 Turbo C 與 Turbo Pascal 都有，
  光憑它分不出來。
- **受影響**：**後續每一份反組譯筆記**——呼叫慣例是 Pascal（參數由左至右壓棧、
  被呼叫者清棧），字串是長度前綴而非 NUL 結尾。用 C 的直覺讀參數順序會全部反過來。
  `CLAUDE.md` §3.1、§4.1 已更正。

---

## 6. Worklist（狀態的單一真相來源）

### 6.0 一句話現況

**M0 資產解密完成**：`.15` 字模、`.TPC`／`.RGB`、`.GLB`／`.GTB`、`PLACD.SAV`、
`NEWSDATA.DAT` 全部解開，其中 `.TPC` 與 `.GLB` 通過逐像素 round-trip。
**Go 解碼層 `internal/assets` 上線**，輸出對實機逐像素相同。
DOSBox oracle 可走到「選擇歷史背景」。

**敘事文本的答案是「畫成圖」**：`NEWSDATA.DAT` 是 17 條新聞橫幅的點陣圖
（「調動軍隊至」「爆發大規模示威遊行」），不是字模序列。所以 remake 的翻譯
工作要處理的是「把圖轉回文字」，而不是改字串表。

**M2 已起步**：`WAR.EXE` 是 Turbo Pascal（不是 Turbo C），字模槽寬與圖集分割
都已從 `Reset` 的 recsize 直接讀出來，把先前 5 個弱證據升級為 confirmed。

### 6.1 下一步

**M2 是現在的關鍵路徑**——規則層與 Ebiten 呈現層必須等 M2/M3 出規格才能動手
（`CLAUDE.md` §2：spec 齊了才實作）。

| # | 項目 | 說明 |
|---|---|---|
| 1 | **戰鬥主迴圈** | `sub_39B6E` 的 `+18F`–`+6F4`。六角格、機動力、地形修正、戰損都在裡面。**M3 最大的一塊** |
| 1a | 地形語意 | 從 `sub_4EFBE`（讀 `tername.dat`）寫入的緩衝區往下追，看 23 種編號怎麼被用 |
| 1b | **`.DT1` 記憶體佈局** | recsize = 14,683 = 整檔一次讀，**不是定長記錄陣列**。要反追資料被搬到哪些全域變數，別再用「檔案÷省數」猜（已錯兩次，見 `docs/playtest/02` §6）|
| 1b | 其餘 13 個政略指令 | 已確定 1=調動、6=查閱。**一次 timeline 只測一個**，別靠 ESC 回主選單（ESC 語意不明，上次因此誤判，見 `docs/playtest/02` §5.4）|
| 1d | 地形編號語意 | 0–22 共 23 種，語意未知。打進戰鬥畫面對照，或從 `WAR.EXE` 讀 `TERNAME.DAT` 後的分支反追 |
| 1e | `NWMAP.DAT` | 15,360 B，**不是**鄰接表（鄰接在 `WARPOS.DAT`）。內容像 u16 對且含 `0x1001`/`0x4001` 旗標，用途待查 |
| 1c | `.DT2` 省狀態 469 B | `MEM_WAR` 與 `SAVE(1)` 只差 2 省，是最好的 diff 素材 |
| 2 | 模組間交接 | `GRT`→`WAR`→`SR`→`GRTE` 靠哪些檔案傳狀態。`Config.dat`（2 B，四支都讀）是最可能的媒介 |
| 2b | `NWMAP.DAT` recsize | 15,360 / 392 = 39.18 不整除，recsize 可能抓錯。待查 |
| 2b | `SDFA.EXE` 解包 | IDA 自動解包對這支失效（0 函式）。要動態 dump 或自寫解包器。優先度低 |
| 3 | `MAN(N).DAT` 欄位 | 33 B/人，`+0/+1/+2` 已知是能力值。從 `WAR.EXE` 讀取該檔的程式碼反追其餘欄位 |
| 4 | `byte_6FE88` 語意 | 值 1/4 才載字模，推測是幕別或階段（`docs/re/01` §2）|
| 5 | `sub_10AB5` | MARK-A → MARK-C 換片流程，確認是否有主動防拷 |
| 6 | 戰鬥規則 | 六角格、機動力、地形防禦（社群說法，全部待驗）|
| 7 | 存檔格式 | `SAVE(1)` 1993-03-23、`SAVE(2)` 1993-01-10，兩份可做 diff；`MEM_WAR.DAT` 與 `.DT2` 同為 18,291 B |
| 8 | `NEWSDATA` 字型 | 17 張新聞圖用倚天反查全部落空，字形較粗、字高 14，來源待查 |
| 9 | `RAIL.TPC` | 不是 BGI，是 chunky 資料（值 0/7/8），格式待查 |
| 10 | 開場動畫轉場 | `.GLB` 的圖與截圖索引對不上（Jaccard ≈ 0.01），轉場機制未確認 |
| 11 | 找其他 `.GLB` | 只有 `GRT`／`GRT2` 兩組，`WAR`／`GRTE` 的 256 色圖裝在哪未查 |
| 12 | `NEWICON`／`NEWTERR` | recsize 是整檔，切張在記憶體裡做。296×18、408×22 整除但未經反組譯確認 |
| 13 | 整理外部資料 | 巴哈攻略百科等頁面人工抓下來放 `docs/reference/`，標來源與日期 |
| 14 | `JAPAN1` vs `JAPAN2` | 各 85 人，兩檔差異未比對 |

### 6.2 已建好的工具

| 工具 | 用途 |
|---|---|
| `tools/py.sh` | Python 執行包裝（docker + uv venv，venv 持久化在 `workplace/.venv-py`）|
| `tools/go.sh` | Go 執行包裝（docker，module cache 在 `workplace/.gocache`）|
| `tools/ida.sh` | IDA Pro 9.4 headless（`analyze` / `script` / `raw`）|
| `tools/dosbox.sh` | DOSBox oracle。沿用現成 image，原版唯讀，寫入導到 `workplace/dosbox/drive_c` |
| `tools/dosbox_runner.sh` | 容器內的 Xvfb + DOSBox + timeline 送鍵截圖（`wait`／`key`／`type`／`shot`）|
| `tools/deny_scan.sh` | 原版資產 deny-list 掃描。commit 前與發行前都要跑 |
| `tools/eten.py` | 倚天字庫讀取 + Big5 分區索引（含 oracle 自我驗證）|
| `tools/glyph_lookup.py` | 字模反查 Big5（`oracle` / `dump` / `match`）|
| `tools/glyph_slots.py` | 槽寬偵測與詞條切分 |
| `tools/entropy.py` | 全檔熵與結構特徵掃描（判「壓縮或加密」前先跑這支）|
| `tools/dcl.py` | PKWARE DCL implode 解壓（`blast.c` 移植，過官方測試向量）|
| `tools/glb.py` | `.GLB` 圖庫解包（`list` / `extract`），輸出 PNG |
| `tools/bgi.py` | BGI `getimage` 解碼（`.TPC`／`NEWSDATA.DAT`）與 `.RGB` 調色盤 |
| `cmd/assetdump` | Go 版解碼器的 CLI：把原版素材匯出成 PNG（139 張），端到端驗證用 |
