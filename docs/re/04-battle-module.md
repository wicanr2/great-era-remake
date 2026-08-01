# RE-04 `WAR.EXE` 戰鬥模組的位置與呼叫圖

> 輸入檔：`WAR.EXE`（375,568 B）
> SHA-256：`11dbfcf24686ab7765f788b38514cefd2039d0f60b6bd517d89fb5a84c068015`
> 位址是 IDA database 的 linear address
> 日期：2026-08-01

---

## 1. 怎麼找到的

`WAR.EXE` 的畫面文字全是字模圖，`GRT.EXE` 那種「靠錯誤訊息定位功能」的捷徑用不上
（`docs/re/02` §4）。改用**資料檔的引用點**當錨——戰鬥一定要讀地圖與戰鬥狀態，
所以從 `warpos.dat`／`tername.dat`／`mem_war.dat` 的 XREF 往上追。

這比讀組合語言便宜，而且錨點是硬的：程式不會引用它不讀的檔案。

---

## 2. 呼叫圖

```
PROGRAM+537
  └─ sub_39B6E                        ← 戰鬥主函式（主程式直接呼叫）
       ├─ +17B → sub_4F0E2            載入地圖
       │           ├─ +14  → sub_4ED9B
       │           │           └─ +12A → sub_4EC99   讀 warpos.dat
       │           └─ +170 → sub_4EFBE               讀 tername.dat
       ├─ +18F → sub_4F883
       │           └─ +A   → sub_4F468               讀 mem_war.dat
       └─ +6F4 → sub_3964E                           寫 mem_war.dat
```

生命週期完整：**載入地圖 → 載入戰鬥狀態 → （戰鬥主迴圈）→ 寫回戰鬥狀態**。

`sub_39B6E` 的 `+18F` 到 `+6F4` 之間就是戰鬥主迴圈所在，**這是下一步該讀的區段**。

---

## 3. 戰鬥模組的函式清單

`0x4E000`–`0x50000` 這一段有 32 個函式，位址連續，是同一個編譯單元：

```
sub_4E055  sub_4E239  sub_4E413  sub_4E579  sub_4E627  sub_4E65C
sub_4E7D0  sub_4EA34  sub_4EB20  sub_4EC99* sub_4ED9B  sub_4EECE
sub_4EF86  sub_4EFBE* sub_4F0B0  sub_4F0E2  sub_4F2AC  sub_4F468*
sub_4F5F3  sub_4F6FF  sub_4F883  sub_4F8E5  sub_4F952  sub_4FA01
sub_4FAFC  sub_4FC19  sub_4FC8B  sub_4FCCC  sub_4FE86  sub_4FEB7
sub_4FEF0  sub_4FF3B
```

`*` = 已確認用途（讀對應的資料檔）。其餘 29 個未分析。

這一段裡**沒有 Graph unit 的呼叫**，所以繪圖不在這裡——
它是純資料／邏輯層，戰場的繪製在別處。對 remake 是好消息：
規則與呈現在原版就是分開的，`internal/game` 與 `internal/ui` 的分層
（`CLAUDE.md` §11）與原版結構一致。

---

## 4. 雙磁碟路徑的差異

三個檔案的第二條路徑不一樣：

| 檔案 | 主路徑 | 備援路徑 |
|---|---|---|
| `warpos.dat` | `warpos.dat` | `b:warpos.dat` |
| `tername.dat` | `tername.dat` | `b:tername.dat` |
| `mem_war.dat` | `mem_war.dat` | **`a:mem_war.dat`** |

`mem_war.dat` 用 `a:` 而其他用 `b:`（`docs/re/01` §2 已注意到這點）。
`MEM_WAR.DAT` 是**戰鬥中會寫回**的檔案，而 `a:` 在雙磁碟機配置裡是
可寫的那一台——這解釋了差異，但**尚未從程式碼確認**。

`mem_war.dat` 有兩處引用（`sub_3964E` 與 `sub_4F468`），一處讀一處寫。

---

## 5. 對 M3 的意義

戰鬥系統是 M3 最大的一塊，也是目前完全未解的部分
（六角格移動、機動力、地形修正、戰損計算）。這份筆記把它從
「不知道在哪」推進到「知道從哪讀起」：

1. `sub_39B6E` 的 `+18F`–`+6F4` 是主迴圈。
2. 地形修正一定會讀 `tername.dat` 載入的陣列——
   從 `sub_4EFBE` 寫入的緩衝區往下追，就能找到 23 種地形編號怎麼被使用，
   這是 `docs/spec/01` §4 那個「地形語意未解」的解法。
3. 戰鬥狀態的結構 = `MEM_WAR.DAT` 的 469 B × 39，
   與 `SAVE(N).DT2` 同結構（`docs/re/03` §3.5）。

---

## 6. 未解

- `sub_39B6E` 主迴圈的內容（六角格、機動力、戰損）。
- 32 個戰鬥模組函式裡的 29 個。
- 出兵流程為什麼沒成立（`docs/playtest/02` §5A.4 的卡點）——
  「軍事行動」的前置條件在政略端，不在這個模組。
- 戰場繪製在哪個模組。
