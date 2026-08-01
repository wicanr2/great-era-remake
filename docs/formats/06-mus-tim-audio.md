# `.MUS` / `.TIM` — AdLib（OPL2）音樂格式

> 狀態：**READY**（解碼層）
> 工具：`tools/mus.py`（`tim` / `header` / `events` / `midi` / `notes` 五個子命令）
> 日期：2026-08-01
> 音源判定：**AdLib / OPL2（YM3812），埠 0x388–0x389** — confirmed

---

## 0. 一頁摘要

| 問題 | 答案 | 推論等級 |
|---|---|---|
| 音源是什麼 | AdLib / OPL2（YM3812），單晶片 9 音；打擊模式下 6 旋律 + 5 打擊 | confirmed |
| 有沒有 MIDI／Sound Blaster 數位音效 | 沒有。五個執行檔都沒有 MPU-401（0x330）或 SB DSP（0x220）埠存取 | confirmed |
| `.TIM` 是什麼 | **AdLib Timbre Bank**（官方副檔名就是 `.snd` / `.tim`），每首曲子自帶的音色庫 | confirmed |
| `.MUS` 是什麼 | **AdLib MUS**，MIDI 風格的事件串 | confirmed |
| 誰在播 | 常駐 TSR，掛在 **INT 66h**；`SDFA.EXE` 是那支 TSR | 強證據 |
| 有沒有音效（BGM 以外）| 有，但只有 `GRT.EXE` / `GRTE.EXE` 的 **PC 喇叭掃頻音**，且與 OPL 無關 | confirmed |
| 旋律能不能取出來 | 能。8 首全部可解析成音符清單，`tools/mus.py midi` 匯出 SMF | confirmed |

熵掃描先做過（`CLAUDE.md` §7 第 8 條）：`.MUS` 4.07–4.76 bit/byte、`.TIM` 2.57–2.67，
**沒有壓縮也沒有加密**，是結構化資料。

---

## 1. 素材清單與雜湊

`workplace/orig/game/` 底下 8 組配對（唯讀；工作副本在 `workplace/audio/`）：

| 曲名 | `.MUS` | `.TIM` | SHA-256（前 16）`.MUS` |
|---|---:|---:|---|
| `MAINTHEM` | 31,227 | 1,176 | `bc6b8173cd20e91a` |
| `BATTLE1` | 13,343 | 1,176 | `db481cd5ec195178` |
| `SCENE` | 12,990 | 1,436 | `9550b3e804f616cc` |
| `FINAL` | 11,525 | 916 | `1d2268c30dc892b7` |
| `BATTLE2` | 11,420 | 1,436 | `0b7f4588adc71f17` |
| `BT02` | 7,132 | 591 | `03514f4cf1996dbd` |
| `WALL` | 5,697 | 591 | `1212cb97d695f3f3` |
| `STRATEGY` | 4,567 | 721 | `8e41204ff23c4703` |

`SCENE.TIM` 與 `BATTLE2.TIM` **位元組完全相同**（`ee1e991c143940c3…`）——同一套 22 個音色。

反組譯輸入檔（IDA Pro 9.4，`tools/ida.sh`）：

| 檔案 | SHA-256（前 16）|
|---|---|
| `WAR.EXE` | `11dbfcf24686ab77` |
| `GRT.EXE` | `dfbd230a968de461` |
| `GRTE.EXE` | `df344af67f20bca4` |
| `SR.EXE` | `991d67544539eda7` |
| `SDFA.EXE` | `79eb0570ba13ad97` |

---

## 2. 播放架構：INT 66h 上的常駐驅動程式

**遊戲本身沒有音樂驅動程式。** 五個執行檔全部只是把整個 `.TIM` / `.MUS` 檔讀進記憶體，
然後透過 **INT 66h** 把緩衝區指標交給一支常駐服務。

### 2.1 API（confirmed）

```
AX = 0,  DS:SI → .MUS 緩衝區    載入樂曲
AX = 1,  DS:SI → .TIM 緩衝區    載入音色庫
AX = 2                          開始播放
AX = 3                          停止播放
```

三支執行檔各自獨立驗證，**不是跨檔外推**（`CLAUDE.md` §7 第 9 條）：

| 執行檔 | 位址 | 呼叫方式 |
|---|---|---|
| `WAR.EXE` | `sub_5CFBB`（載入）、`sub_5D2B9`（AX=2）、`sub_5D2D8`（AX=3 + `FreeMem`）| Turbo Pascal `Intr(66h, Regs)`，`Registers` record：AX 在 +0、SI 在 +0Ah、DS 在 +0Eh |
| `SR.EXE` | `sub_10706`、`sub_10A04`、`sub_10A25` | 同上（與 `WAR` 是同一份原始碼，偏移量 +F8／+20D 完全一致）|
| `GRT.EXE` | `sub_13836`（AX=1 → AX=0 → AX=2）、`sub_1382C`（AX=3）| Borland C++，直接 `mov ax,N` / `int 66h` |
| `GRTE.EXE` | 同 `GRT`（`int 66h` × 4）| 同上 |

`GRT.EXE` 的順序寫得最直白，是這條 API 的最硬證據：

```asm
lds  si, dword_1829E      ; .TIM
mov  ax, 1
int  66h
lds  si, dword_1829A      ; .MUS
mov  ax, 0
int  66h
mov  ax, 2
int  66h                  ; 開始播放
```

### 2.2 那支 TSR 是 `SDFA.EXE`（強證據）

`play.bat`：

```
sdfa > nul
grt
war
sr
grte
```

`SDFA.EXE` 排在最前面、輸出丟棄（典型的 TSR 安裝橫幅抑制），10,243 B、PKLITE 1990-91 打包。
後面四支全部呼叫 INT 66h 卻都不含驅動程式碼——所以驅動只可能來自 `sdfa`。

**尚未解包**（見 §7 未決項目 U1）。這代表下列細節目前只能從資料側推：暫存器對應細節、
是否自動循環、`0xF8` 的實際處理。

> 註：Ad Lib 官方的 `SOUND.COM` 驅動掛的是 **INT 65h**，且不吃檔案格式（只收事件）。
> 這裡是 INT 66h、而且直接吞整個檔案，所以 `SDFA` 是另一支（自製或改編）的驅動，
> 不能直接套 Ad Lib 官方驅動的文件。

### 2.3 音效卡偵測在 `GRT.EXE` 裡（confirmed）

`GRT.EXE` `sub_144C3` 是教科書等級的 AdLib 偵測程序，經 `sub_136E3`（寫 OPL 暫存器）：

```
reg 0x04 ← 0x60      重設計時器旗標
reg 0x04 ← 0x80      重設 IRQ
讀狀態埠
reg 0x02 ← 0xFF      Timer 1 預設值
reg 0x04 ← 0x21      啟動 Timer 1
延遲（讀狀態埠 0xC8 次）
讀狀態埠 → 檢查高位元 → 有 AdLib 傳回 1
```

`sub_136E3` 就是標準的 AdLib 寫入慣例：`out 388h, index` → 讀 388h **4 次**當短延遲 →
`out 389h, data` → 讀 388h **23 次（0x17）**當長延遲。這兩個延遲次數是 AdLib 官方文件寫死的值。

**這一步把音源鎖死在 OPL2。** `WAR.EXE` / `SR.EXE` 沒有任何 `388h` 埠存取
（`WAR.EXE` 全檔只有 4 個 `out`，全是 `ds:63h` 的文字模式游標），完全靠 INT 66h。

`CONFIG.DAT`（2 byte，內容 `00 01`）被四支執行檔讀寫，推測存偵測結果／音樂開關（**假說**）。
`WAR.EXE` 裡每次切歌前都有一個 `jz` 閘門，代表音樂是可關閉的（**強證據**）。

---

## 3. `.TIM` = AdLib Timbre Bank

### 3.1 結構（confirmed）

```
+0x00  u8   majorVersion      = 1
+0x01  u8   minorVersion      = 0
+0x02  u16  nrTimbre          音色數 N
+0x04  u16  offsetDef         參數區起點 = 6 + 9N
+0x06  char[9] × N            音色名（8 字元 + NUL）
offsetDef  Instrument × N     每筆 56 byte
```

每筆 Instrument = **28 個 UINT16LE**：

```
[0..12]   modulator OPLREGS：ksl, multiple, feedback, attack, sustain, eg,
                             decay, release, level, am, vib, ksr, connection
[13..25]  carrier   OPLREGS：同樣 13 個欄位
[26]      modWaveSel
[27]      carWaveSel
```

**`carrier.feedback`（w15）與 `carrier.connection`（w25）在 OPL 上不存在**
（feedback / connection 是聲道級參數，只存在調變器那一格），檔案裡那兩格是未初始化殘值。

### 3.2 驗證

| 檢查 | 結果 |
|---|---|
| 檔案大小 `= 6 + 9N + 56N` | 8/8 精確符合 |
| `offsetDef == 6 + 9N` | 8/8 精確符合 |
| 所有欄位落在 OPL2 合法值域（KSL≤3、multiple≤15、feedback≤7、level≤63、am/vib/eg/ksr/con≤1、wave≤3）| **除 `hihat1` 外 100% 通過**（跳過已知未用的兩格）|
| 跨檔同名音色參數是否一致 | 40 個不重複音色，**0 個衝突** |

`hihat1` 的 carrier 半邊是殘值（w13–w24 出現 47872、60206 這種值）。原因不是解錯——
**OPL2 打擊模式下 Hi-hat 只用聲道 7 的調變器，carrier 那半邊根本不會被寫進晶片**，
所以編曲工具沒去初始化它。7 個含 `hihat1` 的檔案殘值完全相同，佐證它是同一份母帶。

### 3.3 音色庫

8 首曲子合計用到 **40 個不重複音色**，全部來自同一套母庫（跨檔零衝突）：

```
bass2      bassoon1   bdrum1     bells      brass1     clarinet   cymbal1    elguit1
elguit2    elpiano1   flute1     flute2     frhorn1    harp1      hihat1     marimba1
marimba2   oboe1      piano1     piano2     sax1       sitar1     snare1     syn1
syn2       syn3       synsnr1    tincan1    tom1       tromb1     tromb2     trumpet1
trumpet2   trumpet3   trumpet4   tuba1      vibra1     vibra2     violin1    violin2
```

這批命名（`tincan1`、`synsnr1`、`elpiano1`、`sitar1`…）是 **Ad Lib Visual Composer 隨附
標準音色庫**的命名風格。**強證據，不是 confirmed**——手上沒有原版 `STANDARD.BNK` 可逐位元組比對。
比對方法已寫在 §7 未決項目 U3。

---

## 4. `.MUS` = AdLib MUS

### 4.1 檔頭（0x46 byte，confirmed）

```
+0x00  u8    majorVersion   = 1
+0x01  u8    minorVersion   = 0
+0x02  i32   tuneId         本作全部為 0
+0x06  char[30] tuneName    本作全部為空
+0x24  u8    tickBeat       每拍 tick 數，本作全部 = 240
+0x25  u8    beatMeasure    每小節拍數，本作全部 = 4
+0x26  i32   totalTick      全曲長度（tick）
+0x2A  i32   dataSize       事件資料長度 = 檔案大小 − 0x46
+0x2E  i32   nrCommand      事件總數
+0x32  u8[8] 保留（全零）
+0x3A  u8    soundMode      0 = 旋律模式（9 聲部）／1 = 打擊模式（6 + 5）
+0x3B  u8    pitchBRange    弯音範圍（半音），本作全部 = 1
+0x3C  u16   basicTempo     速度（BPM）
+0x3E  u8[8] 保留（全零）
+0x46  事件串
```

八首的實測值（`tools/py.sh tools/mus.py header`）：

| 曲名 | totalTick | dataSize | nrCommand | soundMode | tempo | tick/秒 | 秒數 |
|---|---:|---:|---:|:---:|---:|---:|---:|
| `MAINTHEM` | 62,880 | 31,157 | 7,967 | 1 | 115 | 460 | 136.7 |
| `BATTLE1` | 31,680 | 13,273 | 3,413 | 1 | 110 | 440 | 72.0 |
| `BATTLE2` | 38,400 | 11,350 | 2,941 | 1 | 85 | 340 | 112.9 |
| `SCENE` | 26,880 | 12,920 | 3,262 | **0** | 94 | 376 | 71.5 |
| `FINAL` | 17,760 | 11,455 | 2,918 | 1 | 85 | 340 | 52.2 |
| `BT02` | 19,200 | 7,062 | 1,799 | 1 | 120 | 480 | 40.0 |
| `WALL` | 19,200 | 5,627 | 1,456 | **0** | 76 | 304 | 63.2 |
| `STRATEGY` | 15,540 | 4,497 | 1,213 | 1 | 40 | 160 | 97.1 |

tick 速率 = `tempo / 60 × tickBeat`；秒數 = `totalTick / tick 速率`。
0x46 檔頭裡除了上表列出的欄位以外**全部為零**（8/8 驗證過），沒有隱藏欄位。

### 4.2 事件串（confirmed）

```
[timing byte] [status (可省略＝running status)] [data…]
```

| status | 事件 | data | 語意 |
|---|---|---:|---|
| `8n` | Note Off | 2 | 本作**完全沒用到**（一律用 `9n` + 力度 0）|
| `9n` | Note On | 2 | 音高、力度；力度 0 = 放音 |
| `An` | **音量變更** | **1** | 聲道音量。注意：**不是**標準 MIDI 的 poly aftertouch，只有 1 個 data byte |
| `Bn` | 控制器 | 1 | 本作沒用到 |
| `Cn` | 音色切換 | 1 | **值 = `.TIM` 音色陣列的索引** |
| `Dn` | 通道壓力 | 1 | 本作沒用到 |
| `En` | 弯音 | 2 | LSB、MSB，中心 `00 40`（= 0x2000）；範圍由 `pitchBRange` 決定 |
| `F0 7F 00 <整數> <分數/128> F7` | **速度倍率** | — | `01 00` = 1.00 = 原速 |
| `F8` | 填充位元組 | 0 | 見下方衝突記錄 |
| `FC` | 樂曲結束 | 0 | 8/8 都以它收尾 |

**支援 running status。**

### 4.3 聲部 → OPL2 對應（confirmed）

| 聲部 | 旋律模式（soundMode=0）| 打擊模式（soundMode=1）|
|---|---|---|
| 0–5 | 一般樂器 | 一般樂器 |
| 6 | 一般樂器 | 大鼓 Bass drum |
| 7 | 一般樂器 | 小鼓 Snare |
| 8 | 一般樂器 | 中鼓 Tom |
| 9 | 不用 | 銅鈸 Top cymbal |
| 10 | 不用 | 腳鈸 Hi-hat |
| 11–15 | 不用 | 不用 |

**每個聲部同時只能發一個音**（OPL2 一聲道一音）。

資料側完全吻合：6 首 `soundMode=1` 的曲子裡，聲部 6/7/8/9/10 拿到的音色索引指向
`bdrum1` / `snare1` / `tom1` / `cymbal1` / `hihat1`；2 首 `soundMode=0` 的（`SCENE`、`WALL`）
最高只用到聲部 8，而且把 `bdrum1`、`synsnr1` 當旋律樂器擺在聲部 1／2 上。8/8 相符。

### 4.4 解析驗證

`tools/py.sh tools/mus.py events`：

| 檢查 | 結果 |
|---|---|
| 解析長度 = `dataSize` | **8/8 精確**（全部剛好停在 EOF）|
| 事件數 = `nrCommand` | **8/8 精確** |
| 總 tick = `totalTick` | 6/8 精確（例外見 U2）|
| 所有 `Cn` 音色索引 < `nrTimbre` | **8/8** |
| 未知 status | **0** |

事件數 8/8 逐一相符是最強的整體性驗證：只要 `An` 的 data 長度、running status、
`F8` 的處理任一條猜錯，7,967 個事件的檔案不可能剛好對上。

### 4.5 已量到的表情控制

- **速度倍率**（`F0 7F 00 …`）真的被拿來當音樂手法。實際落點如下
  （`tools/mus.py events` 可重跑；百分比是佔 `totalTick` 的位置）：

  | 曲名 | 位置 | 曲線 |
  |---|---|---|
  | `FINAL` | 8%–15% | 1.000 → 0.945 → 0.898 → 0.797 → 0.695 → 0.594 → 0.500 → **0.445**，然後在 16% 直接跳回 1.000 |
  | `FINAL` | 21%–27% | 第二次較淺的回落 0.984 → 0.977 → 0.969 → 0.953 → 1.000 |
  | `MAINTHEM` | 0%–1% | 0.945 → 0.969 → 0.977 → 0.984 → 1.000（**漸快進入**）|
  | `MAINTHEM` | 93%–96% | 0.984 → 0.977 → 0.969 → 0.977 → 0.984 → 1.000（尾段輕微彈性速度）|
  | `WALL` | 78%–80% | 1.000 → **0.500** → 1.000（一小段半速）|

  **注意 `FINAL` 的大幅度漸慢在曲子前段，不在結尾。** 這首以全速直接收在
  t=17760 的一組放音事件上。第一版文件把它寫成「結尾漸慢收束」是誤讀，已訂正。
- **聲部音量**（`An`）：`STRATEGY` 有 0x55→0x56→0x57→0x58 的連續遞增（漸強）。
  `MAINTHEM` 307 次、`BATTLE2` 222 次，是主要的表情手段。
- **弯音**（`En`）：只有 `BT02` 真的用（16 次 `51 06`，配合 `pitchBRange=1`
  約等於下彎 0.9 個半音，是滑音裝飾）。其餘 7 首只在開頭把每個聲部設回中心值。

---

## 5. 樂曲索引與觸發點

`.MUS` 檔名由一個整數索引選出（`WAR.EXE` 的 `word_70014`、`SR.EXE` 的 `word_1CF5E`）：

| 索引 | 檔名 | 磁片 | 誰設定它 |
|---:|---|---:|---|
| 1 | `battle1` | 3 | **無人設定** |
| 2 | `battle2` | 3 | **無人設定** |
| 3 | `scene` | 2 | `WAR.EXE`：程式啟動、`sub_1BDC6`、`sub_368A8`、`sub_3752F`、`sub_54E3B`（戰鬥九步結算收尾）|
| 4 | `bt02` | 3 | `WAR.EXE`：`sub_540B2`（戰鬥九步結算第 1 步，同一支函式載入 `warrep.tpc` 戰報圖）|
| 5 | `strategy` | 2 | `WAR.EXE`：`sub_3752F` 內兩處 |
| 6 | `wall` | 2 | `SR.EXE`：唯一的設定點 |

`GRT.EXE` 硬寫 `mainthem`，`GRTE.EXE` 硬寫 `final`（字串常數，非索引）。

磁片編號來自 `sub_5CFBB` / `sub_10706` 開頭那段 `var_84 = 2 or 3`，會餵給「請插入磁片」提示——
與 `CLAUDE.md` §3.6 的防拷／磁片檢查是同一條線。

### `BATTLE1.MUS` / `BATTLE2.MUS` 在這個版本裡取不到（強證據）

`WAR.EXE` 全檔對 `word_70014` 只有 13 處參照（9 次寫入、3 次讀取、1 個定義），
寫入值只有 3、4、5；`SR.EXE` 只寫 6。**沒有任何地方寫 1 或 2。**

兩首檔案最大的戰鬥曲反而播不到，是很反直覺的結論，所以把證據攤開：這是對兩個
執行檔做窮舉式交叉參照得到的，不是 grep 落空。`GRT`／`GRTE` 用硬寫字串、不走索引，
`SDFA` 是驅動不選曲。**下一步**：DOSBox 實跑一場戰鬥，錄 OPL 暫存器串流比對到底哪一首響
（見 U4）——實測若推翻這條，照 `CLAUDE.md` §10 記進 `CONTEXT.md` 的推翻清單。

> 順帶提醒 `rulebook/83`：檔名不等於用途。`BT02` 聽起來像 "battle 02"，但它在程式碼裡
> 是**戰報結算畫面**的音樂；`SCENE` 反而是政略主畫面在用。曲目到情境的對應在 §5 只到
> 「哪支函式設定它」為止是 confirmed，**「那支函式對應玩家看到的哪個畫面」仍是假說**，
> 要靠實跑確認。

---

## 6. BGM 以外的音效

**只有 PC 喇叭掃頻音，而且只在 `GRT.EXE` / `GRTE.EXE`（開場與結局動畫）。**

- `GRT.EXE` `sub_13E6C`：用 8253 通道 2（埠 0x43 寫 0xB6、埠 0x42 寫除數）+ 埠 0x61 開關閘門，
  除數由 `1193180 / freq` 算出，上限 `freq ≤ 0x1234`（4660 Hz）。參數是「起頻、迄頻、步進、
  每步延遲圈數」，**是掃頻不是單音**。延遲用空迴圈，隨 CPU 速度漂。
- 呼叫點 4 處，全部在 `sub_1473D`（`GRT` 的動畫主流程）裡，步進參數 1／5／0x0A 各有不同。
- `GRTE.EXE` 有同一段程式碼。
- **`WAR.EXE` 與 `SR.EXE` 完全沒有音效**：沒有 `Sound`／`NoSound`、沒有 0x42／0x43／0x61 埠存取、
  INT 66h 也只有 0–3 四個功能碼（沒有播音效的介面）。

所以遊戲本體（政略 + 戰鬥）**全程只有背景音樂，一個音效都沒有**。這對 remake 是設計決定點，
不是缺漏（見 `docs/design/30-audio-direction.md` §6）。

---

## 7. 未決項目（誠實標記）

| 編號 | 卡在哪 | 已試過什麼 | 下一步的具體方法 |
|---|---|---|---|
| **U1** | `SDFA.EXE` 沒解包，驅動內部行為（暫存器對應細節、是否自動循環、`F8` 實際處理、`An` 音量如何映射到 OPL total level）只能推 | IDA 9.4 自動解包對 `GRT`／`GRTE` 有效、對 `SDFA` 無效；PyPI 上沒有 PKLITE 解包套件（查過 6 個候選名稱全 404）| ① 照 `SDFA.EXE.asm` 裡那段 PKLITE 1990-91 解壓 stub 的位元組重寫解壓器；② DOSBox 跑 `sdfa` 後從 INT 66h 向量取得常駐段，用 FreeDOS `DEBUG` 的 `n`/`w` 把該段寫成檔再進 IDA；③ 用 DOSBox 的 raw OPL 擷取，從暫存器串流反推 |
| **U2** | `totalTick` 兩首對不上：`MAINTHEM` 檔頭 62,880／解析 60,800（差 2,080）；`STRATEGY` 檔頭 15,540／解析 15,600（差 60）| 事件數 8/8 精確相符，代表事件文法沒錯，只有時間累計對不上 | `STRATEGY` 的 60 剛好是結尾那個 `C3 03` 前面的最後一個 timing byte，若 `totalTick` 定義為「最後一個發聲事件的 tick」就成立；`MAINTHEM` 的 2,080 不是 240 的倍數，還沒有解釋。要靠 U1 的驅動碼定案 |
| **U3** | 40 個音色名疑似 Ad Lib `STANDARD.BNK`，但沒逐位元組比對 | 已比對格式（`.TIM` 確為 AdLib Timbre Bank）與命名風格 | 取一份 Ad Lib Visual Composer 的 `STANDARD.BNK`，把 BNK 的 UINT8 欄位加寬成 UINT16LE 後與 `.TIM` 逐筆比對。相符 → 「音色 100% 出自 Ad Lib 標準庫」可升到 confirmed（與 `docs/formats/01` 字模 100% 出自倚天是同一種結論）|
| **U4** | 曲目 ↔ 畫面的對應只有「哪支函式設定索引」是 confirmed | 靜態窮舉交叉參照 | DOSBox 固定存檔實跑：進政略、進戰鬥、看戰報、開場、結局各錄一段 OPL 串流，比對 `.MUS` 的音符序列。同時驗證 `BATTLE1`／`BATTLE2` 是否真的取不到 |
| **U5** | `F8` 的語意與公開規格衝突。VGMPF 的 AdLib MUS 規格說 timing 位置的 `0xF8` 是「溢位，代表 240 tick」；本作資料**不支援**這個解釋 | 兩種解法都實作比對過：把 `F8` 當 +240 溢位會讓 6 首的 `totalTick` 全部對不上；當成「不佔時間的獨立填充位元組」則 6/8 精確、且 `nrCommand` 8/8 精確 | 以本作資料為準（`CLAUDE.md` §4 oracle 序：執行檔／資料 > 外部文件）。要定案仍需 U1 |
| **U6** | 是否自動循環播放未知 | — | U1 或 U4 |

**沒有任何一項是「暫緩／低投報」。** 上面每一列都有下一步方法。

---

## 8. 工具

```sh
tools/py.sh tools/mus.py tim        # 8 個 .TIM 的音色庫 + 欄位值域驗證
tools/py.sh tools/mus.py header     # 8 個 .MUS 的檔頭與換算秒數
tools/py.sh tools/mus.py events     # 事件文法解析 + 對檔頭計數器的驗證
tools/py.sh tools/mus.py midi       # 匯出 SMF 到 workplace/audio/midi/
tools/py.sh tools/mus.py notes BT02 # 單曲音符清單 + 音級分佈
```

匯出的 `.mid` 是**分析用中介檔**，落在 gitignore 的 `workplace/` 底下。
匯出時做了兩處 remake 差異（不是原版語意）：`An` 音量事件轉成 MIDI CC7；不寫
Program Change（`.TIM` 音色名對不到 GM 音色，硬對會誤導）。

> 建議（不在本輪授權範圍，留給主線）：`.gitignore` 的資產 deny-list 目前沒有 `*.mid`。
> `workplace/` 已整個排除所以現在沒風險，但 §9 的發行前掃描應該把 `*.mid` / `*.ogg` /
> `*.wav` 一起列進去。

---

## 9. Remake 的落地介面

規則層要的最小抽象：

```
type Song struct {
    TicksPerBeat  int      // 240
    BeatsPerBar   int      // 4
    TotalTicks    int
    Tempo         int      // BPM
    PercussionMode bool    // soundMode
    BendRange     int      // 半音
    Bank          []Timbre // 來自 .TIM
    Events        []Event
}
```

- 呈現層要能切「復古 / 現代」兩條路（見 `docs/design/30-audio-direction.md` §5）。
  復古路徑吃的就是上面這個結構，直接餵軟體 OPL2 合成器。
- **不要在規則層碰音訊**。`internal/game` 只發「請播第 N 首 / 停」這種訊號。
- 原版 `.MUS` / `.TIM` 不進版控、不打包（`CLAUDE.md` §9）。

---

## 參考資料

- [AdLib Timbre Bank Format — ModdingWiki](https://moddingwiki.shikadi.net/wiki/AdLib_Timbre_Bank_Format)（`.snd` / `.tim`，UINT16LE 欄位）
- [AdLib Instrument Format — ModdingWiki](https://moddingwiki.shikadi.net/wiki/AdLib_Instrument_Format)（OPLREGS 13 欄位順序）
- [AdLib Instrument Bank Format — ModdingWiki](https://moddingwiki.shikadi.net/wiki/AdLib_Instrument_Bank_Format)（BNK 的 char[9] 名稱欄位）
- [MUS (AdLib) — VGMPF](https://www.vgmpf.com/Wiki/index.php/MUS_(AdLib))（檔頭欄位、事件表、聲部對應）
- [AdLib Sound Driver — VGMPF](https://vgmpf.com/Wiki/index.php/AdLib_Sound_Driver)（官方 `SOUND.COM` 走 INT 65h，與本作不同）
- [Programming the AdLib/Sound Blaster FM Music Chips](https://bochs.sourceforge.io/techspec/adlib_sb.txt)（暫存器與延遲慣例）

抓取日期：2026-08-01。
