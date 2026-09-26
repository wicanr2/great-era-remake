# RE-39：`.DT1` 尾端區塊位移與 runtime 快照寫回勘誤

狀態：**confirmed（讀寫位移）／未知（區塊 10 生成時機）**  
日期：2026-08-10

## 證據身分

- 輸入執行檔：`WAR.EXE`，SHA-256
  `11dbfcf24686ab7765f788b38514cefd2039d0f60b6bd517d89fb5a84c068015`。
- 工具：IDA Pro `9.4.0.260610`；以下組語位址是 IDA linear address，
  `+xxxxh` 是暫存 record 內的相對位移，不與 `ds:` 偏移混用。
- 讀檔函式：`sub_59CBF`（`59CBFh..5A031h`）。
- 寫檔函式：`sub_595D4`（`595D4h..598C0h`）。
- `.DT1` 樣本：`SAVE(1).DT1` SHA-256
  `9ad9359b436ce2450ff3460df4efede478d871dfe234f60299759e781959ad0d`；
  `SAVE(2).DT1` SHA-256
  `1c46d3e5caaa468a9be03f521c3d1628f5c6a94e581db7ba8e7c9abae0ef18a7`。

## 1. 原始 `$basg`／單 byte 搬運

`sub_595D4` 與 `sub_59CBF` 都先把檔案當作 `395Bh`（14,683 B）的一筆 record。
區塊 7 在 record `+37FCh`、大小 `112h`，因此其尾端是 `+390Eh`。其後的原始
搬運契約如下：

| runtime 全域 | 原始 record offset | 大小 | 檔案 offset | 證據等級 |
|---|---:|---:|---:|---|
| `byte_6FE85..byte_7003C` | `+390Eh..+3911h` | 4 × 1 B | 14606..14609 | **confirmed**，逐 byte `mov` |
| `byte_6FF8C`（信用度） | `+3912h` | `0Ah` | 14610..14619 | **confirmed**，`$basg` |
| `byte_6FF96`（外債） | `+391Ch` | `28h` | 14620..14659 | **confirmed**，`$basg` |
| `byte_6BC4C` | `+3944h` | 1 B | 14660 | **confirmed**，逐 byte `mov` |
| `byte_6FE88` | `+3945h` | 1 B | 14661 | **confirmed**，逐 byte `mov` |
| `word_70026` | `+3946h` | `14h` | 14662..14681 | **confirmed**，`$basg` |
| `byte_6FFCA` | `+395Ah` | 1 B | 14682 | **confirmed**，逐 byte `mov` |

因此「區塊 8 → 9 → 10 首尾相接、末端另有 7 B」是錯誤的幾何描述。正確描述
是：區塊 9 位於區塊 8 之前；七個未映射欄位分散在區塊 7／8／10 之間。

## 2. 兩份樣本的交叉驗證

| 檔案 | 14606..14609（未映射） | 信用度第 1 槽 14610 | 外債第 1 槽 14620..14623 | block 10 第 1 槽 14662..14663 |
|---|---|---:|---|---|
| `SAVE(1).DT1` | `01 01 01 00` | `64`（100） | `00 00 00 00` | `3A 00`（58） |
| `SAVE(2).DT1` | `05 00 01 01` | `64`（100） | `00 00 00 00` | `A6 00`（166） |

舊測試把 14606 的四個 runtime byte 當成外債，並把 14656 的 `02 01` 當成
信用度／區塊 10；該結論已由上述兩份樣本與讀寫函式共同推翻。舊證據保留在
`CONTEXT.md`／`WORKLIST.md` 的歷史條目，新的位移以本文件與 `FMT-07` 為準。

## 3. remake 接線邊界

- `SaveBlocks` 改用 14610（信用度）、14620（外債）、14662（block 10）。
- `ParseDiplomacyLedger`／`WriteDiplomacyLedger` 從受版控的 `SaveBlocks` 取位移，
  只覆蓋已證實的 50 B；未映射七個 byte 原樣保留。
- `ParseMajorPowerLeaders`／`WriteMajorPowerLeaders` 逐欄讀寫 10 × `u16`；
  `cmd/dsds` 從 session 保存 runtime 快照，autosave 只寫回該快照。
- `word_70026` 的劇本初始化、覆滅／繼任更新與「何時擷取到存檔」仍未完全閉合；
  remake 不以 `FactionLeaders` 猜造 block 10，也不自行命名快照中的數值。
- 區塊 7 仍不是整塊可寫：`WriteFactionOfGeneralLeaders` 只依區塊 6 的非零領袖
  寫回已證實的 1-based 反查格，並保留已覆滅勢力舊格與 265 個殘留格。這個窄
  writer 已接入 `WriteDT1`，重複／越界領袖 ID 會 fail-closed。
