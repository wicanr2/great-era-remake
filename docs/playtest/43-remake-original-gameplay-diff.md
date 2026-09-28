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
  已開車 logo→title（`\r`＋`2\r2\r`，80M 步，title 218 色確認），title 輸入迴圈在
  `6BAF:xxxx`（AH=0B 回 AL=FF 實測）。`*` 切換迴圈與確認鍵仍待由 title 繼續開車。
- 密碼 quiz：每局重抽（頁／行／字＋三選一）；`2＋Return` 兩次通過，
  誤答僅重出 quiz（未見鎖死）。答案需實體說明書，remake 無此關。

## 5. 本輪未覆蓋

- 原版戰鬥畫面同態對照（卡在派將確認鍵未知，需 IDA 找 input routine 或繼續試鍵）、
  整回合操作序列比對、音效／動畫時機。
