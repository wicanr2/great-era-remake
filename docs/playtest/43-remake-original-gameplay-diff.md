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
- 密碼 quiz：每局重抽（頁／行／字＋三選一）；`2＋Return` 兩次通過，
  誤答僅重出 quiz（未見鎖死）。答案需實體說明書，remake 無此關。

## 5. 本輪未覆蓋

- 原版戰鬥畫面同態對照（卡在派將確認鍵未知，需 IDA 找 input routine 或繼續試鍵）、
  整回合操作序列比對、音效／動畫時機。
