# RE-43：`.DT2`／`MEM_WAR.DAT` 欄位搬運與未解邊界

> 狀態：**confirmed（欄位搬運）／unknown（live 同步時機）**  
> 日期：2026-08-11  
> 輸入：`WAR.EXE`（375,568 B）  
> SHA-256：`11dbfcf24686ab7765f788b38514cefd2039d0f60b6bd517d89fb5a84c068015`  
> 工具：IDA Pro `9.4.0.260610`；位址空間：IDA linear address

本筆證據把 `sub_3964E`（寫入）與 `sub_4F468`（讀取）的 `bitcopy` 大小及順序
逐項對齊。它回答的是「469-byte record 哪些 byte 被搬到哪些 runtime buffer」；
不把 runtime buffer 何時與戰鬥清單同步，誤當成已解出的存檔 oracle。

## 1. 469 bytes 的對稱搬運

`sub_3964E` 在 `397F8–39810` 將四個 u16 寫入區域變數，接著在
`39814–39887` 依序以 `0Ah`、`0Ah`、`14h`、`14h`、`0C8h`、`0C8h` 搬運六段，
並在 `3988C–3988F` 複製尾端一個 byte。它最後於 `39938–39956` 以
`Reset(..., 1D5h)`、`Seek(省號−1)`、`Write` 寫入 `MEM_WAR.DAT` 的該筆記錄。

`sub_4F468` 在 `4F522–4F542` 讀取同一筆 `1D5h` 記錄，於 `4F555–4F5EC`
以完全相反的順序將資料放回相同的全域變數。表格中的「已證實」只表示這個
檔案偏移與 runtime buffer 的對應，不代表每個元素都已具有玩法名稱。

| record 偏移 | 大小 | 寫入端（`sub_3964E`） | 讀取端（`sub_4F468`） | 語意等級 |
|---:|---:|---|---|---|
| `+0` | 2 | `word_64932` | `word_64932` | confirmed：第一方資源 1（黃金） |
| `+2` | 2 | `word_64936` | `word_64936` | confirmed：第一方資源 2（糧食） |
| `+4` | 2 | `word_6493A` | `word_6493A` | confirmed：第一方資源 3（彈藥） |
| `+6` | 2 | `word_6493E` | `word_6493E` | confirmed：第一方資源 4（燃料） |
| `+8` | 10 | `byte_6AA8C` | `byte_6AA8C` | shape confirmed；元素語意 unknown |
| `+18` | 10 | `byte_6AA96` | `byte_6AA96` | shape confirmed；元素語意 unknown |
| `+28` | 20 | `word_64902` | `word_64902` | confirmed：一方 10 個參戰將領 ID |
| `+48` | 20 | `word_64916` | `word_64916` | confirmed：另一方 10 個參戰將領 ID |
| `+68` | 200 | `word_6A8F4` | `word_6A8F4` | confirmed：100×u16 runtime ID buffer；檔案／live 同步時機 unknown |
| `+268` | 200 | `word_6A9BC` | `word_6A9BC` | confirmed：100×u16 runtime ID buffer；檔案／live 同步時機 unknown |
| `+468` | 1 | `byte_6AB65` | `byte_6AB65` | confirmed：攻方來源省欄位（角色方向仍以既有證據為準） |

長度驗算：`8 + 10 + 10 + 20 + 20 + 200 + 200 + 1 = 469`，沒有剩餘 byte。

## 2. 不要把第二方資源誤塞回 DT2 header

在 `sub_3964E` 的 `39784–397D0`，`word_64934`、`word_64938`、`word_6493C`、
`word_64940` 被依省號乘 `25h` 後，直接寫入省份 runtime 記錄的不同偏移；
它們沒有被放入上表的 `var_256` 前 8 bytes。這四個變數是第一方四個資源旁邊的
另一方執行期欄位，`sub_546D1` 的結算資料流以成對欄位相加回省份資源，
但「第二方何時再次成為可載入的 DT2 header」目前沒有證據。

因此 Go parser 將 `Header` 定義為 `[4]uint16` 的**第一方**資源，並保持
`BattleState` 的 A/B 中性命名；它不宣稱 469-byte record 內有第二份資源。

## 3. `+8`／`+18`：只有形狀與哨兵已知

兩區的 `@$basg` 長度都是 `0Ah`，在目前 `MEM_WAR.DAT` 樣本可觀察到 `0xFF`
空槽以及少量 0..190 值。但已匯出的直接讀／寫端尚未把單一 byte 接到一個
可重播的部隊、部署或旗標規則。故目前只准使用：

1. 原始 byte-preserving 讀寫；
2. `0xFF` 空槽計數，作為樣本描述而不是玩法判定；
3. 待新 oracle 或更多寫入端證據後再建立 typed 欄位。

禁止把它們命名成「部隊旗標」或把非 `0xFF` 值轉成將領 ID；這會把形狀證據
越級成語意。

### 3.1 2026-08-11 直接 xref 負結果（strong evidence，非全域讀取否定）

本輪重新檢查 `WAR.EXE` 的 IDA 匯出與函式邊界：`byte_6AA8C`／`byte_6AA96` 的
符號命中仍只落在 `sub_3964E` 與 `sub_4F468` 的整段 `0Ah` `bitcopy`，沒有一個
已匯出的直接讀取端把單一元素接到部署、部隊或旗標規則。這支持「元素語意 unknown」
的現況，但不能排除經指標／暫存器／相對基址完成的間接讀寫；若未來出現該取址端，
須另附原始運算元與位址空間，不得把本次負結果升格為「完全沒有讀者」。

## 4. `+68`／`+268` 與 live runtime 的界線

`sub_4F6FF`（`4F75D–4F784`、`4F816–4F83D`）另外從每省 60-byte 戰爭記錄
壓縮非零值到 live runtime `0x6742`／`0x680A` 的 100 槽清單，並在後續更新
機動欄位。這證實兩個 200-byte 區的**元素形狀與 runtime ID 定位**，但不證實：

- 讀檔後何時把 `word_6A8F4`／`word_6A9BC` 複製到 `0x6742`／`0x680A`；
- 每一回合、部署、撤退或戰後結算哪一個事件先寫 live、哪一個事件再落盤；
- `sub_3964E` 的寫回分支是否會在所有勝負路徑執行。

這些仍標為 `unknown`，直到取得至少兩份同一正常玩家路徑的前／後
`.DT2`／`MEM_WAR.DAT` 快照，並能以 byte diff 對應到 `sub_3964E` 的門檻。

## 5. 對 remake 的實作邊界

- `internal/game.ParseBattleState`／`WriteBattleStates` 已能以原始 bytes 為基底
  安全讀寫上述已解欄位；未解區域不因 Go 零值而被清除。
- `ApplyRemakeSnapshot` 的 `+68`／`+268` 是 remake 快照投影，不是原版同步時機
  的證明；`+8`／`+18` 刻意保持不動。
- 原版存檔 oracle 尚未完成，不阻塞 remake 內部 writer 測試，但阻塞「原版
  `.DT2` 全流程等價」的宣告。

## 6. 可回查證據

原始函式匯出（均保留原始函式名、IDA 位址與運算元）：

- `workplace/ida/user-output/function-sub_3964E.txt`
- `workplace/ida/user-output/function-sub_4F468.txt`
- `workplace/ida/user-output/function-sub_4F6FF.txt`

以上檔案的輸入均為本文件標頭的 `WAR.EXE` SHA-256；本文件沒有把攤平 `.asm`
當作唯一證據，也沒有改寫 IDA database 的原始名稱。
