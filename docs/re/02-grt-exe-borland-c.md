# RE-02 `GRT.EXE` — Borland C++，IDA 自動解開 PKLITE

> 輸入檔：`GRT.EXE`（34,708 B）
> SHA-256：`dfbd230a968de4613a79f74041faa2f022c11e929f70e32c131e9bfa0435cd49`
> 工具：IDA Pro 9.4 headless，16-bit DOS loader
> 日期：2026-08-01

---

## 1. 五個執行檔用不同編譯器

| 執行檔 | 大小 | 編譯器 | 函式數 | 角色（從字串判斷） |
|---|---:|---|---:|---|
| `SDFA.EXE` | 10 KB | **未解** | — | PKLITE 打包，**IDA 解不開**。`play.bat` 第一個跑且丟棄輸出 |
| `GRT.EXE` | 34 KB | **Borland C++ 1991** | 178 | GF 資源載入（`grt.GTB`）、DCL 解壓、`mainthem` 音樂 |
| `WAR.EXE` | 375 KB | **Turbo Pascal** | — | 戰鬥（最大的一支，`docs/re/01`）|
| `SR.EXE` | 36 KB | **Turbo Pascal** | 118 | **劇情過場**：`scene141/142/271/272.tpc` + `battle1`/`battle2`/`bt02`/`strategy`/`wall` 音樂 |
| `GRTE.EXE` | 57 KB | **Borland C++ 1991** | 191 | **結局**：與 `GRT.EXE` 幾乎同一套程式，但用 `grt2.GTB` 與 `final` 音樂 |

編譯器的判定依據：Borland C++ 看 `db 'Copyright 1991 Borland Intl.',0`（NUL 結尾字串）；
Turbo Pascal 看 IDA 認出的 RTL（`ASSIGNCRT`、`IOResult`、`BlockRead`）與 Graph unit。

`CLAUDE.md` §3.7 把五支的角色列為**假說**（`GRT` 政略、`WAR` 戰鬥、`GRTE` 結局或事件、
`SDFA` 環境偵測、`SR` 未知）。現在 `SR` 與 `GRTE` 有了字串證據，
但**`GRT` 是不是政略主程式仍未證實**（見 §5）。

**這是 `CLAUDE.md` §7 第 9 條「不要跨檔案外推」的教科書案例**：`WAR.EXE` 是
Turbo Pascal 不代表其他四個也是。反組譯每一支之前都要先確認編譯器，
否則字串形式（長度前綴 vs NUL 結尾）與呼叫慣例都會讀錯。

---

## 2. PKLITE：IDA 自己解開了，不必另外寫解包器

`CLAUDE.md` §4.1 規劃「PKLITE 打包的三個要先解包，否則只會看到解壓 stub」。
實測 **IDA 9.4 的 DOS loader 會自動解 PKLITE**，直接 `tools/ida.sh analyze GRT.EXE`
就得到解壓後的程式：**178 個函式、9 個 code 段、完整的 C 字串表**。
不需要 UNP、depklite 或自己寫解包器。

### 判斷「有沒有被壓縮」不能只看熵

`GRT.EXE` 的熵是 **6.57**，`GRTE.EXE` 是 6.41——比真正被壓的 `SDFA.EXE`（7.60）低很多，
看起來像沒壓縮。**但它們確實被壓了。** PKLITE 的壓縮率遠不如 deflate，
壓完的機器碼熵停在 6.4–6.6 是正常的。

可靠的判準是**可見字串的數量**：34 KB 的程式只有 44 條可見字串、其中 2 條像檔名，
明顯不合理（解開後有 30+ 條有意義的字串）。

`CLAUDE.md` §7 第 8 條說「任何『這東西被壓縮／加密了』的說法，先量熵再說」——
這裡要補一句反向的：**熵低也不能反推「沒被壓縮」**，弱壓縮器的輸出熵不會很高。

---

## 3. 字串表證實了 `.GLB` 的格式判斷

解出來的字串裡有三條直接相關：

```
ERROR!!! in Decompress..err_code : %d
CMP_INVALID_DICTSIZE..compressed data stream are in error.
CMP_INVALID_MODE..compressed data stream are in error.
```

`CMP_INVALID_DICTSIZE` 與 `CMP_INVALID_MODE` 是 **PKWARE Data Compression Library
的標準錯誤訊息**——遊戲直接連結了 DCL。這是對 `docs/formats/02-glb-image-library.md`
的**獨立佐證**：那份文件是從資料反推出 DCL implode 的，這裡是從程式碼確認。

而且對得上我們解析的 header：`DICTSIZE` 對應 header 第 2 個 byte（4/5/6），
`MODE` 對應第 1 個 byte（literal 編碼與否）。兩個欄位剛好各有一條錯誤訊息。

### `GF` = Graphic File，就是 `.GTB` + `.GLB`

```
ERROR open GF table !!!      ← .GTB（目錄）
ERROR open GF data !!!       ← .GLB（資料）
GLB / GTB / grt.GTB
Not found RRN !              ← RRN = Relative Record Number，GTB 的記錄編號
```

原作者把這組檔案叫 **GF（Graphic File）**，table + data 兩份。
我們從資料反推的「目錄 + 圖庫」結構與原作者的命名完全一致。
`RRN` 這個詞也印證 GTB 是定長記錄的目錄（13 bytes/筆）。

---

## 4. 完整字串表

```
By Semy                      ← 作者署名
COMPAQ                       ← 硬體偵測
SDFA                         ← 引用了 SDFA（play.bat 第一個跑的那支）
mainthem                     ← MAINTHEM.MUS
config.dat / grt.GTB
GLB / GTB
rb / wb                      ← fopen 模式

BGM load (Mus file) error!!!
BGM load (Tim file) error!!!
ERROR open GF table !!!
ERROR open GF data !!!
Not found RRN !
DATA file seek error!
OPEN FILE ERROR !!!
can not open file!!!
Input/Output buffer storage short!
Requestted Buffer storage short!            ← 原版拼字錯誤（Requestted）
NO!!! no spaces to alloc a WorkBuff !!
Unable to allocate BIG WorkBuffer.. wait a moment ..
memory avaliable is %lu                     ← 原版拼字錯誤（avaliable）
length =%u  bufsize =%u
ERROR!!! in Decompress..err_code : %d
CMP_INVALID_DICTSIZE..compressed data stream are in error.
CMP_INVALID_MODE..compressed data stream are in error.
Sorry this GAME must have at least EGA video card.
print scanf : floating point formats not linked
```

### 觀察

- **字串表裡沒有任何政略指令、省份、將領相關的文字**。合理——那些都是字模圖，
  不經過 C 字串。C 字串只剩錯誤訊息與檔名。
- `By Semy` 是目前唯一的作者線索。`CLAUDE.md` §1.5 記程式設計是林明輝、沈友瑋，
  **但無法確認 Semy 對應誰**，不要臆測。
- 兩個拼字錯誤（`Requestted`、`avaliable`）是原版的，remake 若要保留原始錯誤訊息
  就照抄，不要「修正」。
- `Sorry this GAME must have at least EGA video card.` 說明最低需求是 EGA，
  與 `EGAVGA.BGI` 一致。

---

## 5. `GRT.EXE` 的角色

34 KB、178 個函式，字串全是資源載入與記憶體管理。看起來是
**資源載入與圖形管線模組**，不是政略規則的所在地。

規則在哪裡還沒確定——`play.bat` 的鏈是 `sdfa → grt → war → sr → grte`，
政略邏輯可能在 `SR.EXE`（36 KB，來歷不明）或以 overlay 形式載入。
**不要假設 `GRT` 就是政略主程式**（`CLAUDE.md` §3.7 把它列為假說，尚未驗證）。

---

## 6. `SR.EXE` 與 `GRTE.EXE` 的字串

### `SR.EXE`（Turbo Pascal，劇情過場）

```
scene / scene141.tpc / scene142.tpc / scene271.tpc / scene272.tpc
battle1 / battle2 / bt02 / strategy / wall          ← .MUS 音樂
.mus / .tim / .BGI / .CHR / EGAVGA
Config.dat / b:Config.dat / b:
Please Insert Disk Mark-
BGI Error: Graphics not initialized (use InitGraph)
```

`SCENE*.TPC` 那 12 個檔（447×270 與 640×190）就是劇情圖，由這一支播放。
`.CHR` 是 BGI 的向量字型副檔名——**但遊戲的中文是點陣字模**，
這個 `.CHR` 應該只用在英數，未確認。

### `GRTE.EXE`（Borland C++，結局）

字串與 `GRT.EXE` 幾乎逐條相同（同一套資源載入程式碼），差別只有兩處：

| | `GRT.EXE` | `GRTE.EXE` |
|---|---|---|
| 圖庫目錄 | `grt.GTB` | **`grt2.GTB`** |
| 音樂 | `mainthem` | **`final`** |

`FINAL.MUS` + `GRT2.GLB`（裡面是國民革命軍軍官配青天白日旗那批插圖）
→ 這一支負責結局。與 `play.bat` 把它排在最後一致。

---

## 7. 未解

- **`SDFA.EXE` 解不開**。檔頭有 `PKLITE Copr. 1990-91` magic，但 IDA 9.4 只認出
  3 個段、0 個函式、0 條字串——**自動解包對這一支失效**（對 `GRT`/`GRTE` 有效）。
  可能是 PKLITE 版本或 `-e` 之類的選項差異。要另外想辦法（動態 dump 或自寫解包器）。
  優先度不高：10 KB，`play.bat` 第一個跑且丟棄輸出，推測是環境偵測。
- **政略規則在哪一支還沒確定**。`GRT.EXE` 的字串全是資源載入與記憶體管理，
  34 KB / 178 個函式也偏小。可能在 `SR.EXE`、以 overlay 載入、或散在多支。
  **不要因為 `play.bat` 的順序就假設 `GRT` 是政略主程式。**
- `By Semy` 對應哪位程式設計師（`CLAUDE.md` §1.5 記林明輝、沈友瑋，無法確認）。
- `SR.EXE` 的 `.CHR` 用在哪裡。
