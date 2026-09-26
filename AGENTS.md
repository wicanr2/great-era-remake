# AGENTS.md —《大時代的故事》remake 作業規範

本檔是 Codex、Claude 與其他自動化代理在本儲存庫工作的共同入口，由舊 `CLAUDE.md`
與舊 `AGENTS.md` 合併而成（2026-09-26）。面向使用者與撰寫文件，預設一律使用
繁體中文；程式識別字、命令、API、工具、產品與檔名保留原文，並以完整中文句型包覆。

`CONTEXT.md` 是全專案的狀態單一入口（現況、文件索引、術語表、已被推翻的斷言、
worklist）。對話被壓縮或新 session 接手時先讀它。本檔只放「不論做什麼都要遵守」
的目標、原則與硬規則；考證細節留在 `docs/`，此處只保留結論與指向。

## 1. 每輪開工順序

1. 先讀 `CONTEXT.md` 的現況、已被推翻的斷言與 worklist。
2. 再讀本檔的硬規則，以及與任務直接相關的 `docs/spec/`、`docs/re/`、
   `docs/mechanics/` 或 `docs/playtest/`；交接時另讀 `WORKLIST.md`。
3. 需求涉及老遊戲、中文化、remake、逆向、開發環境、打包、配樂、字型或實機試玩時，
   先讀 `~/.codex/knowledge-base/knowledge-router.md` 的「復古遊戲路由」，再只載入一份
   任務入口與必要 reference。找不到時要明說，不能假裝已讀。
4. 執行 `git status --short`。工作樹中的既有變更屬於使用者或前一輪工作，不得 reset、覆蓋或丟棄。
5. 先查既有結論與實作，再開始新的逆向或編碼；搜尋落空不等於不存在。
   第一動作是 grep `docs/`、程式碼、攻略、譯名表，不是讀組合語言。

狀態來源有衝突時，以可重現的原版實測、IDA 證據、目前程式與實際測試結果為準，並訂正文件。
`CONTEXT.md` 其中部分早期狀態列會被同檔後文推翻，不能只讀表格單列。

## 2. 專案目的、定位與邊界

目標是完整逆向漢堂國際資訊 1992 年 DOS 遊戲《大時代的故事》，以 Go／Ebiten 乾淨重寫
跨平台引擎，還原繁體中文母本，並提供英文與日文語系。定位是文化資產保存。

- 核心規則一律對齊原版。現代化外殼（高解析呈現、視窗縮放、現代操作、存檔相容）
  允許，但每一項改動都要在文件裡標記為 remake 差異，不得默默改動遊戲規則。
- 使用者已確認本 remake 以 Go／Ebiten 重寫，作者程式碼不得引入 `cgo`、`import "C"`
  或 `#cgo`；新增畫面、輸入與音訊接線都沿用 Ebiten／純 Go 邊界（閘門見
  `tools/check_no_cgo.sh`）。
- 不自行發明、平衡或「順手修正」玩法。疑似原版 bug 先照錄、標證據與推論等級，
  是否修改由使用者決定。
- 不散布原版執行檔、資料、美術、音樂、其旋律衍生物或倚天字型。公開產出只有
  引擎程式碼與翻譯文本；原版資料一律 gitignore。
- 唯一圖像白名單是 `docs/images/*.png` 的成果展示截圖（渲染結果，不是資產本身；
  使用者 2026-08-01 明示）；其他位置的 PNG 也會被拒絕清單擋下。
- 原始素材在 `workplace/orig/`，一律唯讀。存檔實驗使用副本或明確輸出目錄；
  測試存檔一律寫到 `/tmp` 或明確測試輸出目錄，不覆蓋原版存檔。

兩條開發原則：完整性優先於投報（卡關就換方法，不寫「暫緩／低投報」當結論）；
規格驅動開發（SDD），只有標記 `READY` 的規格才能實作。

## 3. 遊戲背景與外部資料索引

漢堂國際資訊，1992-10-15（臺灣），MS-DOS 回合制策略（戰棋），背景為民國初年
國民革命軍北伐與對日抗戰（北伐時期、抗戰前期、抗日後期）。程式為林明輝、沈友瑋，
音樂為吳宗凱、遊戲工場。漢堂第二部原創作品，精神續作為 2009 年鄭立《民國無雙》。
（來源：維基百科；出處細節與玩法輪廓見舊 `CLAUDE.md` §1.5，現收攏於
`docs/reference/` 與 `docs/mechanics/01-vocabulary.md`。）

| 資料 | 網址 |
|---|---|
| 維基百科：大時代的故事 | https://zh.wikipedia.org/zh-tw/大時代的故事 |
| 巴哈姆特 攻略百科（遊戲簡介）| https://wiki2.gamer.com.tw/wiki.php?n=19132 |
| 巴哈姆特 遊戲介紹（a91000800）| https://home.gamer.com.tw/artwork.php?sn=3916730 |
| 巴哈姆特 論漢堂 | https://home.gamer.com.tw/artwork.php?sn=1724320 |
| 民國無雙 哈啦板（大時代心得串）| https://forum.gamer.com.tw/C.php?bsn=23236 |
| 遊戲基地 討論串 | https://www.gamebase.com.tw/forum/1207/thread/12070002 |
| 中文 DOS 遊戲資料庫 | https://cdosgame.simagame.me/ |

巴哈姆特頁面對自動抓取回 403，要人工開瀏覽器看。抓回來的攻略整理進
`docs/reference/`，標明來源與抓取日期，不要散在對話裡。社群資料屬 oracle
最低位階，只能當搜尋線索；社群數值常混入修改版，不可直接當結論。

已確認的執行結構（細節見 `docs/re/02`、`docs/re/03`）：`play.bat` 依序跑
`sdfa → grt → war → sr → grte`；`WAR.EXE` 是遊戲本體（Turbo Pascal，
政略、戰鬥、存檔全在內），`GRT`／`SR`／`GRTE` 是開場、過場、結局播放器，
`SDFA` 尚未解包。顯示模式為開場 320×200 256 色與主遊戲 640×350 16 色（BGI）並用。

## 4. Docker-only 硬規則

分析、批次搜尋、轉檔、建置、測試、抓圖、執行程式、IDA、Wine、Xvfb、DOSBox、SDL、
音訊及 GUI 自動化全部只能在 Docker 容器內執行。主機只用於必要的 `docker`、`git`、
工作樹狀態檢查與儲存庫檔案編輯；不得在主機直接執行 Python、專案程式或測試。

- 一次性工作使用 `docker run --rm`，設定合理的 `--memory`、`--cpus`、`--pids-limit`
  與外層逾時；預設 `--network none`。
- 原始資料與研究輸入唯讀掛載；只有工作樹或明確輸出目錄可寫。
- 可寫容器必須使用 `-u "$(id -u):$(id -g)"`，寫前檢查目標擁有權，寫後抽查。
- 不得把主機 Python、虛擬環境或未鎖版 library 掛入工具映像掩蓋版本問題。
  Python 一律走容器內環境，不污染系統。
- 優先沿用既有映像：Go 使用 `dsds-go:1.25`；IDA 使用 `ida-pro-9.4-ver2`。
- 禁止全域 `docker image/volume/system/container prune`、`docker builder prune` 與 `docker rmi`。
- 每批工作後檢查專案相關容器；只清理由本輪建立且已無用途的容器。不得碰其他專案容器。

既有包裝器：`tools/go.sh`、`tools/py.sh`、`tools/ida.sh`、`tools/dosbox.sh`。
若包裝器缺少必要的資源限制或隔離設定，應修正可重現工具鏈，不可退回主機執行。

## 5. 逆向證據契約

Oracle 優先序：固定狀態的 DOSBox 原版實測 > IDA Pro 9.4 反組譯 >
官方說明書／當年資料 > 社群資料。社群資料與修改版只能當搜尋線索。
衝突時記入 `CONTEXT.md`「已被推翻的斷言」區，寫清誰推翻誰、憑什麼；
推翻舊結論時保留原證據索引，不可抹去歷史。

- 反組譯一律先用 IDA Pro 9.4（`.i64` 的交叉參照、函式邊界、資料流），Ghidra 只作交叉驗證。
  Image 來源 `/home/anr2/ida_94_official/dist`，image 名稱 `ida-pro-9.4-ver2`，16-bit DOS loader。
  IDAPython 在此 image 跑不起來，稽核腳本寫 IDC；headless 的 `print`／`Message()` 不進
  stdout，腳本一律 `fopen` 寫檔。
- 不以攤平 `.asm` 取代 `.i64` 關係圖。讀函式用 `tools/dump_func.py`（自動把位址翻成語意，
  附出處與推論等級），查位址用 `tools/addr.py`，查直接交叉參照用 IDC 工具
 （`tools/ida_xref.idc`）。`dump_func.py` 會把範圍內每個 `word_*`／`byte_*`／`sub_*`
  拿去 grep `docs/**/*.md`，已寫過的一律列出。
- IDA 線性位址與遊戲 `ds:` 偏移是兩套名稱；使用 `tools/addr.py` 的
  `DSEG_BASE = 0x641B0` 換算並清楚標示位址空間。每份筆記標「輸入檔 + SHA-256 +
  IDA 位址」；同時引用 Ghidra 位址時明講是哪一套。每個結論標明在哪個執行檔上驗的，
  不跨檔案外推。
- `add di, 常數` 等算術不會建立 xref；直接 xref 也抓不到取址後的間接讀寫。
  「讀多寫少」或「零命中」時，先驗查詢工具的正對照，再追取址端與間接存取。
  讀寫判定用 `XrefType()`（`dr_W`／`dr_R`／`dr_O`），不要比對助憶碼字串。
- 保留原函式名、全域名、位址、結構偏移與原始運算元。註記要非破壞性：
  原始位址留著，語意是附加的且帶推論等級；不以推測性改名覆蓋唯一定位資訊。
- 每筆語意標 `confirmed`／強證據／假說／未知，附輸入檔名、SHA-256、工具版本、
  位址空間與出處。未達 `confirmed` 的項目必須有醒目警示；名稱本身不是證據。
  任何 `confirmed` 結論要有位址、byte range、資料 diff、原版截圖或可重現實驗。
- 比對結構化存檔時至少使用兩份樣本；一致資料與未初始化殘留必須分開判定。
- 防拷／磁片檢查會擋住 oracle，提早排除，不要等到 M4 之後。

前人教訓（ condensed，完整案例見舊 `CLAUDE.md` §7 與各 `docs/re/`）：動手前先查
函式索引（`docs/re/00-function-index.md`）與位址索引（`tools/addr.py` 三張表，
解出新欄位就回填）；掃常數不要掃結果（老軟體數字常是公式或 Real 編碼算出來的）；
Turbo Pascal 的 `for` 入口跳不經過 `inc`；不要用 `grep -v` 過濾組語（會濾掉索引計算）；
寫下一個值之前先 grep 它；省／將領編號對照表一律從資料取，不要憑印象編；
一條規則只留一份實作，新增規則函式前先 grep；訂正比新結論需要更硬的證據；
從零建立的路徑會藏 bug，開新遊戲是一等驗收路徑；先量熵再斷言壓縮／加密；
猜三次編碼沒中就停手，從繪製分派器反追。

## 6. 文本還原與多語系

- 文本儲存機制已解：不是索引，是定長槽位的字模序列（每 `w` 個字模為一詞條，
  不足補全零字模；`w` 依檔案固定）。51 檔 6,174 字模已全量反查（倚天命中 4,799、
  空白填充 1,374、例外 1）。索引來源是執行檔裡的立即數（1-based）。
  驗收標準是逐像素 round-trip，不是肉眼看起來對。細節見 `docs/formats/01-glyph-text.md`。
- 敘事文本（事件、新聞、劇情）是畫成圖的（如 `NEWSDATA.DAT` 的 17 條新聞橫幅），
  不在上述 51 檔詞表內。
- 繁中是還原母本（§5 還原出的原文，不是重寫）。原版字模索引機制到 remake 就結束；
  重寫版用真正的字串表 + 完整字型，不再沿用每場景字模子集（原版機制仍完整記錄在
  `docs/formats/`）。所有玩家可見文字放語系資料，不寫死在 Go。
- 英文版需重算排版（640×350 全形版面換比例字後字寬、行高、對話框全變）；
  排版層必須先抽離。譯名表 `translations/glossary.md` 是唯一真相。
- 日文版可沿點陣字路線（先確認 `JAPAN1.15` 內容）。英／日人物自傳 overlay 目前為
  `machine-draft`，不得當成 `human-reviewed`；正式人審、假名／比例字型仍是 release gate。

## 7. 遊戲機制文件化

每解出一條遊戲機制，當場按屬性歸檔到 `docs/mechanics/`（與程式碼同等的交付物）：

| 檔案 | 屬性 |
|---|---|
| `00-index.md` | 索引與狀態總表 |
| `10-political.md` | 政略：15 個指令各自做什麼、命令數上限 |
| `20-military.md` | 軍事：出兵流程、部隊編成、兵種 |
| `30-combat.md` | 戰鬥：六角格移動、機動力、地形修正、戰損公式 |
| `40-economy.md` | 經濟：六種資源的生產與消耗、人口、開發 |
| `50-diplomacy.md` | 外交：結盟、談判停火、秘密行動 |
| `60-personnel.md` | 人事：將領忠誠度、任免、能力值怎麼用 |
| `70-ai.md` | 電腦 AI 的判斷邏輯（優先度最高） |
| `80-victory.md` | 勝負判定、事件、時期推進 |

電腦 AI（選指令優先序、出兵目標、兵力配置、派將、撤退／求和門檻、任何帶 `Random`
的判斷、難度參數）一律當場記進 `70-ai.md`，不要事後回頭補。每條機制標推論等級
並寫明在哪個執行檔的哪個位址驗的，或哪次實機測試看到的。未知就寫未知，
附「下一步該從哪裡挖」；不准為了讓文件看起來完整而編機制。

## 8. 程式與文件紀律

- 分層維持單向：`internal/ui` → `internal/game` → `internal/assets`；
  排版邏輯（`internal/ui/textlayout`）不得依賴 Ebiten，否則無頭環境無法測試。
- 一條規則只保留一份實作；新增規則函式前先搜尋現有程式與文件。
- 原版使用 `0xFF` 等哨兵時，不得讓 Go 零值悄悄取代；在唯一入口正規化並測試。
- 存檔寫回採「從原始 bytes 改寫已解欄位」，未解 bytes 不動；要求 byte-for-byte round-trip。
- 每解出一條機制，當場更新 `docs/mechanics/`；AI 相關一律同步到 `70-ai.md`。
- 完成宣告至少包含：相關測試、原版／reference 對照、資產拒絕掃描、dirty-tree 檢查
  與 Docker 清理狀態。測試綠只是必要條件，不是完成；宣稱 `confirmed-remake`／
  `approximate` 還要有固定狀態、差分遮罩或事件 trace 等獨立證據，並明寫尚未取得的
  原版 oracle。只有宣稱 `original parity` 時才要求固定狀態的原版 oracle。
  不得用 debug 傳送、強制勝利、傳送座標或直接賦予資源代替正常玩家路徑。
  截圖驗收要帶固定亂數種子；`ESC` 只取消／退回，`F10` 才離開（Y／N 確認並自動存檔，
  存檔失敗就不離開）。
- 若有代理並行寫檔，主迴圈不得使用 `git add -A` 或 `git add .`；提交時列出明確路徑，
  真的要全加就先 `git status --short` 確認沒有 agent 正在寫的檔。
- 除非使用者明確要求，不自行 commit 或 push。
- 派 subagent 時把邊界寫進 prompt：只能清理自己建立的 container；明列不准改的目錄
 （其他 repo、`~/.claude/`、`workplace/orig/`）；明列不准做的收尾動作
 （commit、push、重編、清理，沒寫的等於允許）。收到「我順便做了 X」一律當事故處理，
  先查影響範圍再談結果。
- 做完一項：更新 markdown → 清掉被推翻的斷言（刪掉並記進推翻清單，不是加註解）
  → 跑測試 → 留視覺／實跑證據 → 更新 `CONTEXT.md` 現況。

里程碑：M0 資產解密 → M1 文本還原鏈 → M2 執行檔反組譯 → M3 規則規格
（含 `docs/mechanics/`）→ M4 Go 規則層（純邏輯，不認識畫面）→ M5 呈現層
→ M6 多語系 → M7 打包發行。先出 spec（`docs/spec/`，標 `DRAFT`／`READY`），
M0／M1 沒完成之前不碰 M4 以後。

目錄：`docs/re/` 反組譯筆記、`docs/formats/` 檔案格式、`docs/spec/` 實作規格、
`docs/mechanics/` 遊戲機制、`docs/playtest/` 實跑驗收、`docs/reference/` 外部資料、
`translations/` 譯名表與語系、`internal/assets/` 純解碼層、`internal/game/` 規則層、
`internal/ui/` 呈現層、`cmd/` 可執行程式、`tools/` 容器包裝與稽核腳本、
`workplace/orig/` 原版素材（gitignore，唯讀）、`workplace/ida/` 解包與 IDA database（gitignore）。

## 9. 進度基線（歷史快照，現況以 `CONTEXT.md` 為準）

2026-08-11 經本工作樹校準：M0 完成；M1 完成；M5 remake 核心完成；
M2／M3／M4／M6 進行中；M7 Linux 預覽包完成。當時 `go test -count=1 ./...` 通過，
`tools/deny_scan.sh --all` 646 檔零命中（展示圖 6 張合規）。

後續勘誤與範圍決策（DT1 區塊位移、SDFA 改純 Go 直接播放、不以 register parity
阻塞玩家路徑、machine-draft 自傳、SPEC-30 近似驗證口徑、DOSBox 取樣暫停）以
`CONTEXT.md` §5.85 起、`WORKLIST.md` 53 起與 `docs/spec/30-approximate-validation-m2.md`
為準，不在此重複貼上。本檔的快照日期是歷史基準，不是狀態來源。

## 10. 目前優先順序

1. 維持戰鬥世界結算／兵力鏡像／攻擊目標標號，以及遠程戰鬥／`.DT1` 分支 writer／
   音訊快取的 Docker 回歸；以 SPEC-30 差分遮罩與 deterministic trace 作近似 remake gate。
2. 以非破壞性 writer、未解 sentinel 與 fail-closed 分支維持 `.DT1`／`.DT2` 安全；
   區塊 2 剩餘欄位、勢力表 `+2`／尾端、block10 自動同步與 `.DT2` live 時機列為
   `unknown` 研究項，不阻塞近似交付。
3. 維持第二層六鍵選單 1..5 handler 的已知副作用近似、`battle.command.6` 拒絕、
   撤退／部署 fallback 與 AI deterministic trace；不把第 6 鍵語意、完整資源／動畫／
   音效時機或逐回合 AI 提升成 parity。
4. 維持純 Go OPL2／AdLib 直接播放與 `-audio=off`；正式 Ogg provenance／人耳／裝置 QA
   另列 release gate；SDFA register parity 不列為 remake 播放前置條件。
5. DOSBox oracle 取樣依使用者指示暫停；後續只處理逐篇人審、正式 Ogg provenance／盲聽、
   Android 實機與跨平台發行，不倒灌成已完成 remake 核心的循環研究。

`SDFA.EXE` 可作獨立精度研究，但不得重新變成 remake 播放的前置條件。

## 11. 常用驗證

下列命令代表專案慣例，但實際執行仍須遵守本檔的 Docker 資源限制與網路隔離：

```sh
tools/go.sh test ./...
tools/py.sh tools/dump_func.py sub_XXXXX
tools/py.sh tools/addr.py -6221h
tools/ida.sh raw idat -A "-S/work/tools/ida_xref.idc <symbol>" WAR.EXE.i64
tools/deny_scan.sh --all
```

## 附錄：舊章節對照

舊 `CLAUDE.md` → 新 `AGENTS.md`：§1 目的 → §2；§1.5 背景 → §3；
§2 原則 → §2 末；§3 已確認事實 → §3 末＋各 `docs/`（不再複貼考證表）；
§4 oracle → §5；§4.1 IDA → §5；§5 文本鏈 → §6；§6 多語系 → §6；
§6.5 機制文件化 → §7；§7 教訓 → §5 末；§8 里程碑 → §8 末；
§9 硬規則 → §8；§10 工作紀律 → §1＋§8；§11 目錄 → §8 末。
舊 `AGENTS.md` §6 長篇快照收攏為 §9 基線，現況以 `CONTEXT.md` 為準。
