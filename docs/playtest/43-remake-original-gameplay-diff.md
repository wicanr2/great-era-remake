# 43-remake 與原版玩法差異（retro 原味模式對照）

日期：2026-09-27。原版 oracle：DOSBox 固定操作序列（`tools/dosbox.sh`），
截圖 `workplace/dosbox/shots/{p1-menu,q1-era,r4-map,s1-attr,t1-intro,u1-map}.png`。
remake 對照：retro theme（`cmd/screenshot -theme retro`）。
狀態標記：`confirmed`（有截圖／碼證據）／`unknown`（待同態實測）。

## 1. 開局流程（confirmed：結構不同）

| 原版 | remake |
|---|---|
| 標題動畫（320×200）→ Return → 主選單「1.重新開始／2.載入遊戲」 | 無標題／主選單，啟動旗標直接進局 |
| 三幕歷史背景選擇（北伐／抗戰前期／抗戰後期） | `-province` 等旗標指定，無三幕選擇 UI（劇本由 `ScenarioByStage` 決定，有碼無互動屏） |
| 玩家人數（1–9）→ 選擇軍事領袖（0–9 群像） | 無 |
| 電腦個性（和平／好戰）＋AI 等級＋難易度（初學者／入門／進階） | 無（AI 難度走內定規則，見 `docs/mechanics/70-ai.md`） |
| 說明書密碼 quiz（第76頁第6行第2個字） | 無（程式內無「密碼」字串，grep 證） |

## 2. 政略主畫面（confirmed：結構不同）

- 原版：全國戰略地圖（39 省分色編號）為主畫面，左側為選中省面板
  （廣東省／蔣中正／黃金5000／糧食20005／彈藥20000／燃料14019／民忠44）。
- remake：省份六角 detail 為主畫面，無全國戰略地圖屏（程式內無對應 render，grep 證）。
- 省面板數值語意一致（兵力114600／將領17／民忠44 同值，`u1-map.png` vs
  `province-36.png`；黃金 5000 vs 3440 是不同回合狀態，非規則差）。

## 3. 規則層（SPEC-30 近似口徑，維持）

- 15 政略指令、戰鬥 chain、AI handler 已接；逐回合 parity、DT2 落盤時機、
  第二層六鍵第 6 鍵維持 `unknown`，不升格 parity（見 issue #1 E 組）。
- 戰鬥同態截圖待補：原版進真實戰鬥需完整戰役操作，本輪未達，列 `unknown`。

## 4. 進攻派將流程（confirmed，2026-09-27 補）

- 2 號軍事行動 → `欲攻打何省？（鄰省表）` → 數字鍵＋Return 送目標省。
- 目標接受後進 `欲派遣何將？`＋14 將領表（姓名＋兵力）；數字鍵＋Return 切換選取，
  已選標 `*`（實測 `10` 選中土府均）。確認鍵未知——已實測無效：
  Return／Space／0／Y／Tab／F1；Esc 為取消（退回下令選單，指令數不扣）。
  約 22 輪 DOSBox 盲試後暫收，需 IDA 找派將 input routine。
- GRT.EXE 靜態爬梳（2026-09-28，`confirmed` 顯示棧／`unknown` 鍵盤路徑，
  GRT.EXE.i64＋`tools/ida_callers.idc` code-xref，以上符號皆遠程函式）：
  `sub_10FDD`＝INT 10h 顯示包裝；`sub_11A25`←`sub_11A59`←`sub_11B23` 是
  螢幕字元讀寫＋數字輸入棧（ah=8 為讀螢幕字元，非鍵盤）；`sub_10D83`＝列表渲染，
  呼叫端僅 `sub_10E56`／`sub_11799`／`sub_117EF`／`sub_11937(×2)`；
  `sub_10E56` 唯一呼叫端是 `sub_119D4`（`seg000:11772` 靜態點）；
  `sub_11B23` 呼叫端僅上述三填充迴圈；`sub_10DA8` 純算術無 I/O。
  以上皆無鍵比較，`*` 切換迴圈不在此分支。
- 關鍵否定發現（`confirmed`）：全 GRT.EXE 無 `int 16h`、無 DOS 鍵盤功能
 （49 處 `int 21h` 僅向量存取 AH=25h/35h 與檔案 3Ch/3Dh/3Eh/40h）、無 60h port
  讀取。先前「鍵盤鏈」命名作廢，實為顯示棧。確認鍵維持 `unknown`。
- dosgolem 實測推翻「自掛鍵盤 ISR」假說（2026-09-28，分支 `exp/dosgolem-isr`，
  `dosgolem/` 快照 d9c0c27；GRT.EXE＋原版目錄唯讀掛載）：
  20M／30M／80M 步皆 `int 09h` 向量維持預設 `0080:0024`，IVT 0-400 唯一寫入是
  loader 在 #34 寫向量 0；那一次 AH=25h 未改變任何 IVT 項目。
  真正輸入路徑是 DOS stdin：`sub_11925`（全 EXE 唯一 `AH=0Bh`）查有無鍵，
  有鍵後走 `AH=3Fh` 讀（`Type` 通道；`PressKey`/IRQ1 對本作無效）。
  已開車 logo→title（`\r`＋`2\r2\r`，80M 步，title 218 色確認）；title 迴圈在
  `sub_1473D` 內（`0x14F8B`／`0x1508F` 輪詢＋動畫，鍵→`0x1509C`；OnCall 驗證
  poll 18→20 次。先前「6BAF:xxxx」是 CallTrace ES:BX 誤讀，作廢）。
- GRT title 後必退（2026-09-28 dosgolem＋靜態，`confirmed` 機制）：
  title 鍵由 `sub_1173D` 讀（全 EXE 唯一 `AH=07h`，另吃 `byte_1817A/B` 注入槽；
  先前「無 DOS 鍵盤功能」修正為「僅 AH=0B＋AH=07，無 INT16/60h/BDA」），讀完即棄，
  再判 `word_182A2`（音效旗，開機 `sub_144A6` AdLib 388h 偵測＋`sub_1346C`
  驗 SDFA 的 0x66 常駐簽章；dosgolem 無卡時為 0，有 `-adlib` 才過第一關，
  `sub_1346C` 仍失敗）。兩分支最終都 `retf` 回 start stub → `AH=4Ch`。
  SDFA 先駐＋`-adlib` 可把旗置 1（peek `lin:EBC2`＝01），但 title＋鍵仍退；
  GRT 本來就是 intro／title 模組，**派將流程在 WAR.EXE**（`docs/re/29` 輸入檔、
  `02` §5A.4 明指由 WAR 反追；2xxxx 位址系）。
- WAR.EXE 在 dosgolem 下可開車（SDFA 先駐＋`-adlib`，60M 步存活，14 檔：
  CONFIG／HEAD1.RGB／SCENE／1–4.15／EGAVGA.BGI／W.TPC／MARK.TPC／CHOOSE…；
  EGA mode 10h 640×350，先前 320×200 解讀作廢）。
- WAR 輸入鏈（2026-09-28，`confirmed`）：靜態 WAR.EXE 無 INT16 是因為 IDA 把
  CRT 段收合了——`@READKEY$qv`（0x6245D：AH=0 阻塞讀，AL=0 擴展鍵暫存
  `byte_7024D`）＋`@KEYPRESSED$qv`（0x6244B）＋讀後 `sub_62293` **清 buffer**
 （Ctrl-C 經 int 23h）。4.15M 次 INT16 皆經此。dosgolem `Type`/`-keys` 走
  Stdin 餵 INT16，`PressKey`/IRQ1 不需要。新工具 `tools/ida_range.idc`
  可反組譯收合區（`ida_callers.idc` 姐妹作）。
- 選單族（`confirmed`，靜態）：`sub_2B063`／`sub_2BA3A` 計數器選單——`4` 遞減、
  其餘有效鍵遞增、`ESC` 直達目標值；`sub_2AFF1`／`sub_2B9B9` 驗鍵（`3` 常被拒）。
  `sub_10AB5` 是換片提示（MARK-A/B/C，非 quiz）。
- 實測開車到選擇屏（紅框＋✓遊標，CHOOSE31/32→CHOOSE.RGB/41/42 逐级載入）：
  數字鍵絕對定位遊標（`1` 上／`2` 中），`CR`（0x0D）提交——`LF`（0x0A）無效，
  早先 Enter/Space/Esc 無反應多半是送錯位元組。**每次 READKEY 後清 buffer，
  必須一鍵一存檔分段餵**（batch 全被吃掉）。目前停在 s12（`war_s12.state`），
  遊標經 `1/2/4` 多次移動，待繼續。檢查點鏈全在 /tmp（不進版控）。
- 密碼 quiz：每局重抽（頁／行／字＋三選一）；`2＋Return` 兩次通過，
  誤答僅重出 quiz（未見鎖死）。答案需實體說明書，remake 無此關。

## 5. 本輪未覆蓋

- 原版戰鬥畫面同態對照（卡在派將確認鍵未知，需 IDA 找 input routine 或繼續試鍵）、
  整回合操作序列比對、音效／動畫時機。

## 6. 查閱省畫面等待迴圈（2026-09-29 dosgolem ISR；§7 有方法勘誤，本節部分結論已降級）

- 畫面＝`sub_2A941` 省份查閱（靜態 `confirmed`，欄位順序對 `docs/re/27`），
  內容安徽省（司令／省長蔣中正，兵力 50000，將領數 4，忠誠度 46）＋右側中國地圖；
  迴圈經 `sub_2B0F4`（XREF：`sub_2C351+2B0`）。執行期讀鍵鏈：
  `int16 AH=00 @43E3:01BE ← 386E:080A ← 4681:55xx` 輪詢（`58A7` 為模擬器 stub）。
- `2`–`7` 首按走慢路徑約 34–35K 道指令：
  `43E3:0272` 前後 → `5558:18F5`／`0EEE`（帶指標呼叫，像繪圖常式）→
  `5709:0000` → `74A9:0000`（音效模組；`74A9:0EE5` 同時是 timer-ISR 音樂入口，
  `-ip-log` 實測）。無 VRAM 變化、無開檔。同鍵重按轉快路徑拒收
  （`10×'5'`：首鍵 +35K，其餘 +192；`'1','2'` 重測：`'2'` 慢、`'1'` 快）。
  行為與 `mechanics/10` §4「`2`–`9` 立即生效、`1` 前綴」相容，但各指令均無
  畫面副作用（失敗前檢＋beep 回饋？面板未重繪？皆為假說，`unknown`）。
- 快路徑拒收（消耗但無效）：`0 1 8 9 CR ESC SPACE`、大小寫字母（已測
  `a–r s–z A–R`，大寫 `K S–Z` 未測）、`-+*/#=`、方向鍵、Home/End/PgUp/PgDn/
  Insert/Delete/Tab/Backspace、`F1`–`F12`、`CtrlC`（`AL=03` 送達亦無效）。
  `'0'` 有一次出現 272K 間隙（狀態相依，`unknown`）。可送按鍵空間已耗盡。
- 滑鼠零輪詢；IRQ1 全程死亡（未裝 int 09h，向量 `0080:0024`）；`-press` 不可用。
  方向鍵此前送不出去——本輪在 dosgolem 分支加了 `-sendkeys`（具名鍵走
  BDA／Keys 字組佇列，`int16` 優先讀；已驗 `Up/Down/Left/Right` 以
  `int16-AH00-bda AL=00` 送達）。分支 `exp/dosgolem-isr`，未提交。
- s22→本輪末僅 650px 差異（rows 278–291 消息行動態，日軍文本自動清除）；
  之後 50 億＋道指令零變化：＋2B 無鍵、`-adlib`＋100M、`-block-after` 皆無效。
- 靜態計數器選單驗鍵器（`sub_2AFF1`／`sub_2B9B9`，本輪 dump 確認）只收
  `4`／`6`／`CR`／`ESC`——執行期行為與兩者皆不合，實際 dispatch 函式未定位
  （`43E3` 段內另有 `warmenu.tpc` 字串，見 `.asm` `aWarmenuTpc`／`sub_4D97C`）。
- 下一步：`scancode.go` 已擴（`F1`–`F12`、`CtrlC`，本輪實測全拒收）；
  剩餘只能靜讀 `386E:080A` 上層 dispatch（需執行期→IDA 段換算，目前無錨點），
  或懷疑此迴圈根本不是等鍵（但 `-block-after` 未觸發、鍵盤輪詢持續中）。

## 7. 方法勘誤：調色盤分層＋AC 真相（2026-09-29 續輪，`confirmed`）

- **§6 的「畫面零變化」結論作廢（觀測盲區）**：只比了 VRAM 色號與 DAC，
  對 EGA 屬性控制器（AC）翻層失明。`writeEGA`（`-dump-at` PNG，走
  色號→AC→DAC）才是完整真相；`-dump-palette` 只看 DAC。此後一律以
  `-dump-at` PNG 為畫面依據，VRAM／DAC 只作輔助。
- **`'4'` 翻到戰場進入層（AC flip，`confirmed`）**：同 VRAM 色號＋戰場 AC
  ＝編號省份地圖＋日期（民國・秋）＋「進入戰場中」（`t4` 系列 PNG、
  `proof_battle.png` 交叉渲染為證；226／256 DAC entries 差異是 AC 映射後
  結果，不必逐項追）。`CR` 沒有翻回（`bf_c` 的省畫面是 DAC 渲染假象）。
- 執行期 dispatch 位元組與 `sub_4D585` 逐指令同構（`cmp '1'／'6'／'0'／CR`＋
  游標 `6AB64`／鍵盤 `6AB66`＋`sub_5544F` 驗鍵）；`sub_5544F` 吃掉不在呼叫端
  字串內的鍵、`CR` 恆通過；現行有效集含 `1`–`6` 不含 `0`（`'7'` 慢是誤讀，
  間隙屬於前一鍵；`'0'` 快拒、`'1'` 快、`'2'`–`'6'` 慢、`CR` 被接受（10 萬＋道 excursion：
  `74A9` 68K＋`53F9` 15K）。
- 按鍵盤＝`byte_6AB66`＝`ds:69B6`（`DSEG_BASE`＋執行期 `DI` 雙重確認）。
  讀者普查：`sub_42056`、`sub_4D585`（呼叫端僅 `42566`／`4BA1E`／`4C9CC`）、
  `sub_2B063`／`sub_2BA3A`、`sub_4B475`／`sub_4C33B`／`sub_4C3C4`／`sub_5375E`；
  `sub_42487` 是純繪圖。CRT：`sub_62293`（清 buffer＋`int 23h`）、
  `@READKEY 0x6245D`（`int 16h` 在 `+0xE`）。`sub_4C9CC` 有
  「`< 0 > : <Enter> :`＋CR 離開」迴圈。無 `.OVR`，TPOV 不適用。
- 戰場層試鍵（AC aware）：`CR`、`3`、`30`＋`CR`、`1` 皆無變化；卡點疑為無
  作戰設定（`ds:B346h` 空）致載入停滯（假說）。`playtest/16` 正常流程 2–4 秒
  進部署；此處零開檔停滯。
- `-dump-at` 在恰等於 `-steps` 時不觸發（邊界 quirk；`steps＝目標＋餘量` 規避）。
- 下一步：戰場層試 `2`／`5`／`6`／`7`／`8`／`9`／`ESC`／`SPACE`（AC aware），
  或讀 `sub_4C9CC` 部署等待條件，或 `-poke` 填 `ds:B346h` 診斷觸發（非正常
  玩家路徑，需標註）。
