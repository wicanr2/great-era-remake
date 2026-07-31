# RE-01 `WAR.EXE` — Turbo Pascal，不是 Turbo C

> 輸入檔：`WAR.EXE`（375,568 B）
> SHA-256：`11dbfcf24686ab7765f788b38514cefd2039d0f60b6bd517d89fb5a84c068015`
> 工具：IDA Pro 9.4 headless（`tools/ida.sh analyze WAR.EXE`），16-bit DOS loader
> 位址一律是 IDA database 的 linear address
> 日期：2026-08-01

---

## 1. 結論：編譯器是 Turbo Pascal

`CLAUDE.md` §3.1 記「`WAR.EXE` … Borland Turbo C runtime」——**這是誤判**。
那個 Borland 版權字串是 Turbo Pascal 的。

證據（IDA 自動辨識出的 RTL 函式簽名）：

| 類別 | 函式 |
|---|---|
| 字串 | `Concat(String&,String&)`、`Copy(String&,String&,Integer,Integer)`、`Insert(String&,String&,Word,Integer)`、`Delete(String&,Integer,Integer)` |
| 檔案 | `Assign(File&,String&)`、`Reset(File&,Word)`、`BlockRead(File&,Any&,Word,Word&)`、`FilePos`、`FileSize`、`IOResult` |
| CRT | `ASSIGNCRT(TEXT&)`、`CLRSCR`、`CLREOL`、`DELAY(WORD)` |
| **Graph unit** | `INITGRAPH`、`GETIMAGE`、`PUTIMAGE`、`SETRGBPALETTE`、`SETACTIVEPAGE`、`SETVISUALPAGE`、`BAR`、`CIRCLE`、`ELLIPSE`、`DRAWPOLY`、`OUTTEXTXY` |

字串也是 Pascal 形式（長度前綴）：

```
aPleaseRemoveDi db 33,'Please remove Disk MARK-A from A:'
```

**這件事會影響後續每一份反組譯筆記**：呼叫慣例是 Pascal（參數由左至右壓棧、
被呼叫者清棧），字串是長度前綴而非 NUL 結尾，函式名被 IDA demangle 成
`@Reset$qm4File4Word` 這種形式。用 C 的直覺讀參數順序會全部反過來。

`CLAUDE.md` §4.1 說「Turbo C 編譯的程式有一個額外優勢：BGI 呼叫的參數形狀固定，
可以拿來當函式識別的錨點」——**這個優勢仍然成立，而且更強**：Turbo Pascal 的
Graph unit 直接被 IDA 認出名字，不必自己辨識。

### 順帶印證的兩件事

- `SETRGBPALETTE(INTEGER,INTEGER,INTEGER,INTEGER)` 存在
  → 印證 `.RGB` 的百分比要經過 6-bit DAC（`docs/formats/04-tpc-bgi-image.md`）。
- `SETACTIVEPAGE` / `SETVISUALPAGE` 存在 → 遊戲用 page flipping 雙緩衝。
  remake 的呈現層要注意這一層語意。

---

## 2. 字模載入：`sub_5E4FC`

引用了 `japan1.15`（+166）、`fan(1).15`（+2AF）等檔名。開頭的流程：

```
byte_6FE88 == 1 或 4  →  載入；否則跳過
Assign(f, 'man115.')          ← Pascal 的無副檔名寫法，實際檔案是 MAN115
IOResult <> 0  →  Assign(f, 'b:man115.')    ← 雙磁碟機的第二條路徑
Reset(f, 5Ah)                 ← recsize = 90
IOResult <> 0  →  sub_5EB1B(2)              ← 跳「請插入磁片」提示
```

`byte_6FE88` 的值決定載不載，值 1 / 4 才載——推測是遊戲階段或幕別，**未確認**。

### `b:` 前綴 = 雙磁碟機支援

`CLAUDE.md` §3.1 觀察到「`b:` 前綴的雙磁碟路徑」，現在知道機制了：
每個檔案都先試當前目錄，`IOResult` 非 0 再試 `b:`，還是失敗才跳磁片提示。
**這對 DOSBox oracle 是好消息**——只要檔案都在同一個目錄，第一條路徑就會成功，
`b:` 那條與磁片提示都不會觸發。與 `docs/playtest/01-dosbox-probe.md` 的實測一致。

---

## 3. 槽寬：從 `Reset` 的 recsize 直接讀出來

Turbo Pascal 的 `Reset(f, recsize)` 指定 typed file 的記錄大小，程式之後用
`BlockRead` 一次讀一筆。**recsize 就是一個詞條的 bytes 數**，除以 30 就是槽寬。

| 檔名（Pascal 字串） | recsize | ÷30 = 槽寬 | 先前的判定依據 |
|---|---:|---:|---|
| `1.15` | 30 | **1** | content（弱）→ **confirmed** |
| `2.15` | 60 | **2** | content（弱）→ **confirmed** |
| `3.15` | 90 | **3** | auto ✓ |
| `4.15` | 120 | **4** | content（弱）→ **confirmed** |
| `japan1.15.` | 150 | **5** | auto ✓ |
| `man115.` | 90 | **3** | auto ✓ |

`docs/formats/01-glyph-text.md` 裡標成 `content`（證據較弱、待反組譯確認）的那幾筆，
**現在全部確認**。空白排版模式推出來的結果與反組譯完全一致，這也反過來說明
那套判準是可靠的。

---

## 4. 圖集分割規則：也是 recsize

`CLAUDE.md` §3.3 的「`AC.TPC` 差額巨大，是多張圖串接」——分割規則同樣藏在 recsize：

| 檔案 | recsize | 每張的組成 | 張數 | 驗算 |
|---|---:|---|---:|---|
| `AC.TPC` | 5,544 | 5,524（178×60）+ 20（palt） | **17** | 94,248 / 5,544 = 17 ✓ |
| `NewIcon.tpc` | 5,328 | — | 1 | 整檔一次讀 |
| `NewTerr.tpc` | 8,976 | — | 1 | 整檔一次讀 |

`AC.TPC` 的 17 張已由 Go 解碼層（`internal/assets.DecodeBGISet`）獨立解出，
張數吻合——兩條路徑互相驗證。

`NewIcon.tpc` / `NewTerr.tpc` 的 recsize 等於整個檔案大小，表示 `WAR.EXE`
一次把整份讀進記憶體，**在記憶體裡才切張**。切法要另外追（`NEWICON` 每張
276+20=296 → 18 張、`NEWTERR` 每張 388+20=408 → 22 張，數字整除但**未經反組譯確認**）。

---

## 5. 其他檔案的 recsize

| 檔名 | recsize | 意義 |
|---|---:|---|
| `NewsData.dat` | 1,752 | 與 `docs/formats/04-tpc-bgi-image.md` 算出的區塊大小一致 ✓ |
| `MAN(N).DAT` | 9,042 | 整檔一次讀（274 人 × 33 B） |
| `TOWN(N).DAT` | 1,443 | 整檔 |
| `SAVE(N).DT1` | 14,683 | 整檔 |
| `a:mem_war.dat` | 469 | **注意 `a:` 前綴**，與其他檔的 `b:` 不同 |
| `frame1.tpc` | 664 | 整檔 |

`Config.dat` 的 recsize 是 1（逐 byte 讀），檔案只有 2 bytes。

---

## 6. 磁片檢查

| 位址 | 內容 |
|---|---|
| `aPleaseRemoveDi` | `'Please remove Disk MARK-A from A:'`，XREF 在 `sub_10AB5` 附近 |
| `aAndInsertDiskM` | `'and insert Disk MARK-C in B:'`，XREF `sub_10AB5+8A` |
| `aPleaseInsertDi` | `'Please Insert Disk Mark-'`，XREF `sub_5EB1B+EE` |

`sub_5EB1B` 是「請插入磁片 N」的提示函式，由檔案開啟失敗的路徑呼叫
（見 §2）。也就是說**磁片提示是 I/O 失敗的後備路徑，不是主動的防拷檢查**。

這修正了 `CLAUDE.md` §3.6 的措辭（「有磁片／防拷檢查」）：目前看到的是
**雙磁片版的換片提示**，還沒有看到序號、密碼、或刻意的防拷驗證。
`sub_10AB5` 那組（MARK-A → MARK-C）還沒細讀，**先不下定論**。

---

## 7. 未解

- `byte_6FE88` 的語意（值 1/4 才載字模）。
- `sub_10AB5` 的完整流程（MARK-A → MARK-C 換片）。
- `NEWICON` / `NEWTERR` 在記憶體裡的切張方式。
- 主迴圈、模組間交接（`GRT` → `WAR` → `SR` → `GRTE`）、亂數、存檔讀寫。
- `MAN(N).DAT` 的 33 bytes 欄位語意（見 `docs/re/02-man-dat.md`，尚未建立）。
