# [已解決] DOSBox 按鍵送不進去——先前能做到的操作現在重現不了

> **2026-08-01 解決。** 根因一（PointerRoot 焦點）成立且當時就修好了；
> **根因二（keymap 空）是誤診**——`xmodmap` 在那個 image 裡根本不存在，
> 「0 行」是命令缺失不是 keymap 空。修好焦點之後只差**重測**。
> 完整經過見 `docs/playtest/06-dosbox-restored.md`。
> 以下保留原文，記錄當時的診斷。

日期：2026-08-01
目標：進戰鬥畫面，取戰鬥中的 `SAVE(N).DT2`，用來解開 200 B 單位詳細資料的欄位
（`docs/re/05` §4 記的「最有效的下一步」）

---

## 1. 為什麼想打進戰鬥

`MEM_WAR.DAT` 的 200 B 單位區在**戰前是全 0**，靜態資料看不出欄位切法。
唯一可靠的取證方式是打進一場實際戰鬥、存檔、比對 `.DT2`——
那是 `CLAUDE.md` §4 oracle 優先序的第一位。

而且現在有一條捷徑：`byte_7003B` 是**觀戰開關**（`docs/mechanics/70-ai.md` §2），
打開之後主迴圈對每個省都會進戰鬥。它在設定選單裡，
而政略指令 15「其他選項」很可能就是設定入口。

`MEM_WAR.DAT` 也顯示河南正被張作霖系攻打（`docs/re/05` §2），
載入存檔就有現成的戰鬥。

---

## 2. 卡在哪：按鍵完全沒有作用

三種送鍵方式都試過，畫面全部停在主選單（「1. 重新開始 / 2. 載入遊戲」），
游標始終在第 1 項：

| 方式 | 實作 | 結果 |
|---|---|---|
| `type` | `xdotool type --delay 80`（XTEST，不帶 `--window`）| 無反應 |
| `key` | `xdotool key --window $w`（XSendEvent）| 無反應 |
| `key` 改版 | `xdotool windowactivate --sync` + `xdotool key --clearmodifiers`（XTEST）| 無反應 |

`Return` 也一樣沒作用——從標題到主選單那一步是自動跳的，不是按鍵造成的。

視窗確實存在且尺寸正確（`WINDOW=2097155 WIDTH=640 HEIGHT=350`），
所以不是找錯視窗。

### 這是回歸，不是從未成功

`docs/playtest/02` 的整份紀錄（政略畫面 13 個欄位、15 個指令、攻打子選單）
都是靠同一套工具做出來的，截圖還在 `workplace/dosbox/shots/`。
**同樣的路徑現在走不通。**

---

## 3. 診斷：兩個根因

在 image 裡直接跑 Xvfb + DOSBox 做了一次完整診斷：

```
=== 所有視窗 ===
  2097155  name=DOSBox 0.74-3 ... Program: GRT   WIDTH=640 HEIGHT=400
=== 焦點 ===
  getactivewindow: Your windowmanager claims not to support _NET_ACTIVE_WINDOW
  getwindowfocus:  XGetInputFocus returned the focused window of 1
=== keymap ===
  xmodmap -pk → 0 行
```

只有一個視窗，所以不是找錯視窗。問題有兩個：

### 根因一：焦點是 PointerRoot

`XGetInputFocus` 回傳 **1 = PointerRoot**。那表示「焦點跟著滑鼠走」，
而滑鼠預設停在 root window 上——**XTEST 送出的鍵盤事件全部進了 root**。

`windowactivate` 救不了：它需要 WM 支援 `_NET_ACTIVE_WINDOW`，
而 image 裡**沒有任何 WM**（`matchbox`／`twm`／`openbox`／`icewm` 全都沒有）。

已經照這個修了 `tools/dosbox_runner.sh`：送鍵前先 `windowraise` +
`mousemove --window`（PointerRoot 模式下等於把焦點交給它）。
**但畫面仍然沒反應**，所以焦點不是唯一的問題。

### 根因二：keymap 是空的

`xmodmap -pk` 輸出 **0 行**——X server 沒有鍵盤映射表。
沒有 keymap，`xdotool key 2` 就算「執行成功」也產不出有效的 keycode，
DOSBox 收到的是無法解讀的事件。

image 裡也沒有 `setxkbmap`。

### 候選解法（依成本排序）

1. **在 runner 裡補 keymap**：Xvfb 啟動時帶 `-keybd` 選項，
   或用 `xkbcomp` 灌一份標準映射。要確認 image 裡有哪些 xkb 工具。
2. **自建 DOSBox image**：加一個輕量 WM 與完整的 xkb 資料。
   最正規但要新建 image（不能改別人的）。
3. **繞過 X 層**：用 DOSBox 的 `-c` 參數或 mapper 檔綁定輸入。
4. **改資料而不是按鍵**：直接改存檔讓遊戲進入想要的狀態，
   例如設省份記錄 `+32` 的 bit 6（正在打仗）再載入。
   要先確認遊戲會不會在載入時重算那個旗標。

### 為什麼先前會成功

`docs/playtest/02` 的截圖是真的，所以這條路曾經通過。
最可能的解釋是那時的 Xvfb 帶了 keymap（啟動參數或 image 版本不同）。
**沒有保留當時的環境快照**，這是這次的教訓——
oracle 環境本身也該記錄，不然回歸了無從比對。

---

## 4. 記帳

按 `CLAUDE.md` §2：**卡關就換方法，記錄卡在哪、試過什麼**，
不寫「暫緩／低投報」當結論。

這一項沒有解決，但它擋住的是「戰鬥單位欄位」這一條線，
不擋靜態反組譯——`sub_54826` 那批函式還有幾支沒讀完，
那條路不需要 DOSBox。

工具的改動（`tools/dosbox_runner.sh` 改用 XTEST + `windowactivate`）
保留了，因為那本來就比 XSendEvent 正確，即使目前還不足以解決問題。
