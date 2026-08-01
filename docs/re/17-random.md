# 亂數：Turbo Pascal 6.0 的 LCG，序列完全可重現

輸入檔：`WAR.EXE`（375,568 B，未打包）
位址：檔案位移（不是 IDA linear address——這一段在 runtime 裡，IDA 沒展開）
日期：2026-08-01

`docs/playtest/12` 證實徵稅的收入帶亂數（同省同操作結果差三倍），
結論是「不能寫成確定公式」。這一輪把那個門檻拆掉。

---

## 1. 先確認編譯器版本

```
0x59BC2: "Copyright (c) 1983,90 Borland"
```

**1990 → Turbo Pascal 6.0。** `docs/re/01` 已經確認 `WAR.EXE` 是
Turbo Pascal 編譯的，這裡再拿到版本。

---

## 2. LCG 的完整反組譯

`WAR.EXE` 偏移 `0x5AF80` 附近（IDA 標成 `@Random$q4Word` 的呼叫目標）：

```asm
A1 2A 07        mov  ax, [072Ah]            ; RandSeed 低 16
8B 1E 2C 07     mov  bx, [072Ch]            ; RandSeed 高 16
8B C8           mov  cx, ax
2E F7 26 26 16  mul  word ptr cs:[1626h]    ; × 乘數的低 word
D1 E1 ×3        shl  cx, 1                  ; cx <<= 3
02 E9           add  ch, cl                 ; cx × 0x101
03 D1           add  dx, cx
03 D3           add  dx, bx                 ; seed_hi × 1
D1 E3 ×2        shl  bx, 1                  ; bx <<= 2
03 D3           add  dx, bx                 ; seed_hi × 4
02 F3           add  dh, bl                 ; seed_hi × 0x400
B1 05           mov  cl, 5
D3 E3           shl  bx, cl                 ; bx <<= 5
02 F3           add  dh, bl                 ; seed_hi × 0x8000
05 01 00        add  ax, 1
83 D2 00        adc  dx, 0                  ; +1（32-bit）
A3 2A 07        mov  [072Ah], ax
89 16 2C 07     mov  [072Ch], dx
C3              ret
```

### 移位序列逐位對上 `0x08088405`

這是**用 16-bit 指令手工展開的 32-bit 乘法**，兩半分開驗：

| 半邊 | 指令 | 等於乘上 |
|---|---|---|
| `seed_lo × K_hi` | `cx <<= 3` 然後 `ch += cl` | `× 8 × 0x101` = **`0x0808`** |
| `seed_hi × K_lo` | `+bx` `+bx<<2` `+bl<<8` `+(bx<<5)<<8` | `1 + 4 + 0x400 + 0x8000` = **`0x8405`** |

合起來 **`K = 0x08088405 = 134,775,813`** ——正是 Turbo Pascal 的標準乘數。

> 常數本身在 binary 裡搜不到（`05 84 08 08` 零命中），
> 因為它**被編譯成移位與加法**，從來沒有以立即數的形式出現。
> 這是 `CLAUDE.md` §7 第 2 條「掃常數不要掃結果」的反面案例：
> **有時連常數都掃不到，要從算式的形狀反推。**

### 遞推式

```
RandSeed := RandSeed × 134775813 + 1        （32-bit 環繞）
```

---

## 3. `RandSeed` 在 `ds:072Ah`；`Randomize` 用系統時間

緊接在 LCG 後面：

```asm
B4 2C           mov  ah, 2Ch                ; DOS「取系統時間」
CD 21           int  21h                    ; → CH:CL = 時:分, DH:DL = 秒:百分秒
89 0E 2A 07     mov  [072Ah], cx
89 16 2C 07     mov  [072Ch], dx
CB              retf
```

**`Randomize` 把系統時間直接寫進 `RandSeed`。**
`PROGRAM` 只呼叫它一次（檔案行 1416），之後 55 個 `Random` 全部
從那個種子往下推。

這解釋了 `docs/playtest/12` 為什麼同省同操作結果不同——
**每次開遊戲的種子都不一樣**。

---

## 4. `Random(range)` 取的是高位，不是餘數

Turbo Pascal 的 `Random(n)`：

```
Random(n) = (RandSeed_unsigned × n) >> 32
```

**不是 `seed mod n`。** 兩者的分佈不同，寫錯會與原版分岔。

無參數的 `Random: Real` 走 8087 模擬（`int 35h`／`39h`／`3Dh`），
值是 `seed / 2^32`。

---

## 5. 這解掉了什麼

`CLAUDE.md` §9 有一條「截圖驗收要帶固定亂數種子」。
`docs/playtest/12` 把它從建議變成硬需求：

> 沒有可重現的亂數，徵稅的行為就無法與原版逐次對照。

現在**序列完全可重現**：

- 規則層有 `internal/game/random.go`，同種子同序列
- 要與原版逐次對照，可以在 DOSBox 裡讀 `ds:072Ah` 比對
- 測試與截圖驗收一律用固定種子

### 對徵稅（與其他帶亂數的規則）的意義

`docs/playtest/12` 的結論「不能寫成確定公式」**仍然成立**，
但「無法定式」不再成立——**只要讀出 `sub_2C6C6` 裡 `Random` 的
呼叫點與參數，公式就完整了**。實機再量幾次只能量到分佈，
現在有了確定性的產生器，可以反過來**用序列驗證公式**。

---

## 6. 還沒做的

- **動態驗證**：在 DOSBox 裡 dump `ds:072Ah` 的實際序列，
  與 `internal/game/random.go` 對照。靜態推導很硬，但沒有實測過。
- `cs:[1626h]`（乘數低 word 的存放處）沒有直接讀出來——
  是從移位序列反推的。動態 dump 可以一併確認。
- **55 個 `Random` 呼叫點的用途**：25 支函式，
  `sub_19D6A`／`sub_3964E` 各 6 次最多。逐一對出來才知道
  哪些規則帶隨機。
