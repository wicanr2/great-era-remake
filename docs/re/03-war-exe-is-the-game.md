# RE-03 `WAR.EXE` 是遊戲本體，另外三支只是播放器

> 依據：四支執行檔的 IDA database（`docs/re/01`、`docs/re/02` 記了 SHA-256）
> 方法：比對每支引用的資料檔字串
> 日期：2026-08-01

---

## 1. 結論

`CLAUDE.md` §3.7 的假說是「`GRT` 政略主程式、`WAR` 戰鬥模組（375 KB 最大，
且戰棋的戰鬥系統最重）」。**反了。**

**`WAR.EXE` 是整個遊戲**——政略、戰鬥、存檔、資料表全在裡面。
另外三支各自只讀兩三個檔，是開場／過場／結局的**播放器**。

判定方法很單純：看每支引用哪些資料檔。這比讀組合語言便宜得多，
而且結論是硬的——程式不會引用它不讀的檔案。

---

## 2. 證據：每支引用的資料檔

### `WAR.EXE`（Turbo Pascal，375 KB）

```
字模   1.15 2.15 3.15 4.15  fan(1)(2)(3).15  japan1.15 japan2.15
       man115. man215. man315.  tn15.
圖     choose1/2/31/32/41/42.tpc  map1/2.tpc  menu1-5.tpc  warmenu.tpc
       book1/2/3.tpc  frame1/2.tpc  news.tpc  Report.TPC  warrep.tpc
       hmeg1.tpc  mark.tpc  w.tpc  rail.tpc  newicon.tpc  NewTerr.tpc
       AC.TPC  head1/2.tpc  scene22.tpc
調色盤 bridge.rgb  choose.rgb  cloud.rgb  germeny.rgb  head1.rgb
       kill.rgb  port.rgb  speech.rgb            ← 八個 .RGB 全在這
資料   nwmap.dat  tername.dat  warpos.dat  mem_war.dat  NewsData.dat
       man(N).dat  town(N).dat  config.dat
存檔   save(1).dt2 … save(10).dt2                ← 十個存檔位
```

### 另外三支

| 執行檔 | 引用的檔案 | 角色 |
|---|---|---|
| `GRT.EXE` | `config.dat`、`grt.GTB`、`GLB`、`GTB` | 開場圖庫播放器 |
| `SR.EXE` | `config.dat`、`scene141/142/271/272.tpc`、`.mus`、`.tim` | 劇情過場播放器 |
| `GRTE.EXE` | `config.dat`、`grt2.GTB`、`GLB`、`GTB` | 結局圖庫播放器 |

三支合計引用不到 10 個檔案，而 `WAR.EXE` 引用 80 個以上。
`GRT`／`GRTE` 的字串表也全是資源載入與記憶體管理（`docs/re/02` §4），
沒有任何政略或戰鬥相關的東西。

---

## 3. 順帶確認的三件事

### 十個存檔位

`WAR.EXE` 引用 `save(1).dt2` 到 `save(10).dt2`，正好對應 `CHOOSE2.TPC` 那張
「1.進度一 … 10.進度十／欲載入哪個檔案（1-10）」的畫面
（`docs/formats/04-tpc-bgi-image.md` §3 解出來的那張）。

但 `workplace/orig/game/` 裡只有 `SAVE(1)` 與 `SAVE(2)` 各一組 `.DT1`/`.DT2`——
玩家只存過兩格。

### `.DT1` 與 `.DT2` 的分工

`WAR.EXE` 引用 `save(N).dt2` 是**列舉十個**，而 `.dt1` 只出現一次（`).dt1`，
是 `save(%d).dt1` 這種組出來的字串）。兩者用途不同，**分工未確認**。

### `a:mem_war.dat` 的 `a:` 前綴

其他檔案的第二條路徑都是 `b:`（雙磁碟機，見 `docs/re/01` §2），
只有 `mem_war.dat` 出現 `a:` 前綴。`MEM_WAR.DAT`(18,291 B) 與
`SAVE(N).DT2`(18,291 B) **大小相同**，推測是戰鬥中的記憶體映像，
但 `a:` 為什麼特殊**未確認**。

---

## 3.5 資料表結構：39 個省是核心維度

Turbo Pascal 的 `Reset(f, recsize)` 直接說出每個資料表的記錄大小
（同 `docs/re/01` §3 解字模槽寬的手法）：

| 檔案 | recsize | 檔案大小 | 筆數 | 判讀 |
|---|---:|---:|---:|---|
| `TERNAME.DAT` | 196 | 7,644 | **39** | 每省 196 B |
| `WARPOS.DAT` | 196 | 7,644 | **39** | 每省 196 B |
| `MEM_WAR.DAT` | 469 | 18,291 | **39** | 每省 469 B |
| `SAVE(N).DT2` | (469) | 18,291 | **39** | 同上結構 |
| `NEWSDATA.DAT` | 1,752 | 29,784 | 17 | 新聞橫幅（`docs/formats/04`）|
| `MAN(N).DAT` | 9,042 | 9,042 | 1 | 整檔一次讀（274 人 × 33 B）|
| `TOWN(N).DAT` | 1,443 | 1,443 | 1 | 整檔一次讀 |
| `CONFIG.DAT` | 1 | 2 | 2 | 逐 byte |
| `NWMAP.DAT` | 392? | 15,360 | 39.18 | **不整除，recsize 可能抓錯**，待查 |

**39 就是省份數**——`TN15.1`–`TN15.39` 的數量，也是 `3.15` 裡的 39 個省名
（`docs/formats/01-glyph-text.md` §4）。三份資料表都以省為單位，
這印證了 worklist 上「`TN15.N` 對應第 N 省」那條待驗項。

`NWMAP.DAT` 是唯一對不上的：15,360 / 392 = 39.18。可能是我從
`mov ax, N` 抓到了錯的那一個（15,360 / 384 = 40 剛好整除，但 384 沒有出現）。
**待查，先不寫進格式文件。**

### 存檔 `.DT2` = 39 個省的狀態

`MEM_WAR.DAT` 與 `SAVE(N).DT2` 同為 18,291 B = 39 × 469，餘 0。
拿手上兩份存檔比對：

| 比對 | 結果 |
|---|---|
| `SAVE(1)` vs `SAVE(2)` | 8/39 個省不同，變動集中在 71/469 個 byte 位置 |
| `MEM_WAR` vs `SAVE(1)` | **只有 2/39 個省不同** |

`MEM_WAR.DAT` 與 `SAVE(1)` 的時間戳都是 1993-03-23（`docs/formats/03-placd-archive.md`），
所以 `MEM_WAR` 是那次遊戲的戰鬥暫存，與 `SAVE(1)` 幾乎同步。
**這是存檔格式最好的 diff 素材**：同一局、相鄰時間點、只有兩個省不同。

省 0 的前 64 bytes（兩份存檔一致的部分）：

```
ad 07 4b 35 cc 0a df 55          ← 8 B，兩份存檔相同
ff × 20                          ← 空槽標記（未駐紮的部隊？）
42 00 a0 00 b4 00 6e 00 ba 00    ← 5 個 u16：66 / 160 / 180 / 110 / 186
00 × 12
3b 00                            ← 59
```

**欄位語意全部未解**，上面只是位元組佈局的觀察。要從 `WAR.EXE` 讀寫這些
偏移量的程式碼反追才能定名（`rulebook/62` 靜態反追溯源）。

---

## 4. 對 M3 / M4 的意義

規則規格（M3）要挖的東西全部集中在 `WAR.EXE` 一支：

- 政略指令與省份命令數
- 戰鬥的六角格、機動力、地形
- `MAN(N).DAT` 的 33 bytes 將領屬性
- `NWMAP.DAT`(15,360 B)、`TERNAME.DAT`(7,644 B)、`WARPOS.DAT`(7,644 B) 的語意
- 存檔格式（`.DT1`/`.DT2`）
- 亂數

好消息是它是 Turbo Pascal，`Reset(f, recsize)` 會直接說出每個資料表的記錄大小
（這招在 `docs/re/01` §3 已經解出字模槽寬與圖集分割）。

壞消息是它 375 KB，而且畫面文字全是字模圖，C 字串那種「靠訊息定位功能」的
捷徑用不上。

---

## 5. 未解

- `play.bat` 的鏈是 `sdfa → grt → war → sr → grte`，但 `WAR.EXE` 自己也引用
  `scene22.tpc`。**模組間怎麼交接、誰決定跳到哪一支，未確認。**
  `Config.dat`（2 bytes，四支都讀）是最可能的交接媒介。
- `SR.EXE` 只播 `scene141/142/271/272`，但 `SCENE*.TPC` 有 12 個檔。
  其餘 8 個由誰播未確認。
- `.DT1` 與 `.DT2` 的分工。
- `a:mem_war.dat` 的 `a:` 前綴為何與眾不同。
