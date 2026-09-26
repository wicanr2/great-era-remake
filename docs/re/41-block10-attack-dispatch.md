# RE-41：block10 讀取鏈與戰鬥攻擊分派證據

日期：2026-08-10  
輸入：WAR.EXE  
SHA-256：11dbfcf24686ab7765f788b38514cefd2039d0f60b6bd517d89fb5a84c068015  
工具：IDA Pro 9.4.0.260610，以下位址一律是 IDA linear address；不得與 ds: 檔案偏移混寫。

本紀錄只收斂能直接接到 remake 玩家路徑的最小證據。原始函式名、位址與未解欄位保留；
語意旁標示 confirmed、strong inference、hypothesis 或 unknown。這份紀錄不把一個
可執行的 Go handler 擴大解讀成原版全流程等價。

## 1. 「六種攻擊」的勘誤

### 1.1 1..6 是方向／目標輸入（confirmed）

sub_4D585（0x4D585..0x4D96E）建立 "0123456" 輸入串，接受 ASCII '1'..'6'，
以 sub_4D510 移動目前格／方向游標；0 或 Enter 離開。這是六方向目標選擇，
不是六個有名稱的攻擊模式。

sub_4A40D（0x4A40D..0x4A470）只畫攻擊目標提示並等待 Enter；真正的 1..6
讀取在後續 sub_4BF27／sub_4D585。因此 remake 的攻擊目標表可由鍵盤、滑鼠與
觸控共用，但不能把數字標籤寫成「六種攻擊」。

### 1.1a 2026-08-11：`sub_4BF27` 的第二層六鍵輸入勘誤（confirmed）

進一步逐指令核對 `function-sub_4BF27.txt` 後，前一節的「真正的 1..6 讀取在
`sub_4BF27`／`sub_4D585`」需要拆成兩條不同輸入鏈：

1. `sub_4D585` 在 `4D585h..4D96Eh` 建立 `"0123456"`，是六方向／目標游標；
2. `sub_4BF27` 在 `4C165h..4C18Ah` 先呼叫 `sub_4A40D`，再直接把
   `"123456"` 交給 `sub_5544F`。這是目標檢查後的第二層選單，並非再次呼叫
   `sub_4D585`。

`sub_4BF27` 在 `4C266h..4C2BAh` 的直接分支如下。位址均為 IDA linear address；
「第 6 鍵」只表示 `sub_5544F` 接受該字元，不能把清理落點升格成某個攻擊效果：

| 第二層按鍵 | 直接呼叫 | 指令證據 |
|---:|---|---|
| `1` | `sub_4B827` | `cmp al,31h` → `call 4B827h` |
| `2` | `sub_53111` | `cmp al,32h` → `call 53111h` |
| `3` | `sub_4B854` | `cmp al,33h` → `call 4B854h` |
| `4` | `sub_4BA1E` | `cmp al,34h` → `call 4BA1Eh` |
| `5` | `sub_4BCBE` | `cmp al,35h` → `call 4BCBEh` |
| `6` | 本函式沒有 handler call；落到 `loc_4C2CB` 清理 | `cmp al,35h` 不相等後直接清理；名稱／取消語意 unknown |

所以目前能確認的是「第二層六鍵選單存在、其中五鍵有直接 handler 分支」；尚不能
從這段控制流獨自得出六個攻擊名稱，也不能把第 6 鍵的無 handler 落點解釋成取消、
保留功能或原版 bug。舊段落把 `sub_4D585` 的方向輸入外推成整個攻擊選單，現由本節
勘誤收回；舊證據仍保留供追溯。

### 1.2 目標後的內部分派（confirmed／strong inference）

sub_4BF27（0x4BF27..0x4C2EA）在目標／狀態檢查後進入數個 handler：

| 函式 | 已觀察到的作用 | 等級與限制 |
|---|---|---|
| sub_4B827（0x4B827..0x4B854） | 呼叫 sub_530B4，設 byte_6AA8A=1 | confirmed：正規單次 wrapper；helper 副作用仍未知 |
| sub_53111（0x53111..0x53428） | 掃六鄰、收最多五個支援單位，分配主攻與支援戰損 | strong inference：協同攻擊鏈；+7A8B 等欄位的完整命名仍未知 |
| sub_4B854（0x4B854..0x4B9B3） | 目標兵種為 4／5 時走特殊畫面／動畫；其他路徑呼叫 sub_53428 | confirmed 的控制流；特殊效果與選取時機 unknown |
| sub_53428（0x53428..0x534FF） | 呼叫 sub_51D68 五次，套用雙方損失，收尾呼叫 sub_517BE／sub_52EEA | strong inference：五次正規戰損切片；helper 的全部 side effect 未閉合 |
| sub_4BA1E（0x4BA1E..0x4BCBE） | 兵種 4 的遠程路徑，經 sub_58854／sub_57B15 | confirmed：可移植的射程／三次目標戰損見 docs/re/09；彈藥、動畫、邊界仍未知 |
| sub_4BCBE（0x4BCBE..0x4BE33） | 兵種 6 對目標兵種 4／5 的特殊畫面，否則呼叫 sub_58613 | confirmed 的分支形狀；傷害、音效與寫回語意 unknown |

因此目前可交付的 Go 窄切片是：Engage（單次）、EngageRanged（兵種 4、射程內、
三次目標戰損）及 EngageRepeated（sub_53428 的五次呼叫）。EngageRepeated 位於
internal/game/battleattack.go，只重用既有戰力／戰損函式，不假定未解的畫面與音效
副作用。它的證據等級是 strong inference，不代表原版玩家選單已有五次模式按鍵。

### 1.3 仍未知的攻擊事項（unknown）

- 第二層選單六個按鍵的畫面名稱、第 6 鍵語意，以及每個 handler 對兵種／目標狀態的
  完整可達條件與順序。
- sub_53111 的 +7A8B／byte_6A40A 精確欄位語意、支援單位的戰損除法與失敗分支。
- sub_51EC0、sub_51F19、sub_5A3B2 及 sub_517BE／sub_52EEA 的資源、狀態、動畫、
  音效副作用。
- 邊界折返、攻方彈藥消耗、攻擊動畫／音效時機，以及 sub_4B854／sub_4BCBE 的
  特殊分支是否改寫兵力。

在取得正常 DOSBox 固定狀態樣本前，不能把這些 unknown 以「六攻擊」名稱寫入規則、
測試期望或玩家說明；第二層六鍵的存在已 confirmed，但五個 handler 的名稱、第 6 鍵
語意與完整副作用仍是 unknown。

## 2. block10 與 runtime 領袖名冊

### 2.1 已證實的讀取／更新鏈

> **2026-08-11 勘誤：** `[di-418Ch]` 在 1-based `i×2` 索引下就是
> `word_70026[1..10]` 的同一個 block10 儲存區；「runtime 名冊」是用途描述，
> 不是第二份獨立資料表。`docs/re/42-block10-init-and-new-game-flags.md` 保存
> 新局生成與接班寫入的位址證據。

sub_391E1（建立 DT1 snapshot 的初始化鏈）先清空 block10 的 10 個 u16
（word_70026），第一期的 `sub_38839` 會依新局選項填入同一個儲存區；以下函式
只讀取或更新該 block10 及其相鄰狀態：

| 函式 | 觀察 | 等級 |
|---|---|---|
| sub_30003（0x30003..0x30374） | 月／季結算讀 word_70026，以名冊值尋找將領並改寫 +7A7D／+7A8D 相關狀態 | confirmed：讀取端；生成時機未證實 |
| sub_2FCA9（0x2FCA9..0x30003） | 年／月處理讀 word_70026，命中時寫將領 +7A8E=20000 等狀態 | confirmed：讀取端；不是 block10 writer |
| sub_35005（0x35005..0x3512B） | 參照勢力領袖後，從 [di-418Ch]（即 block10）10 槽移除失效領袖並左移、尾槽歸零 | confirmed：移除鏈；觸發事件仍未知 |
| sub_3512B（0x3512B..0x3523B） | 領袖替換／勢力相關 bytes 清除，另清 reverse map | strong inference：覆滅／接班清理；欄位命名未知 |
| sub_34A4E（0x34A4E..0x34B0B） | 依勢力編號清其他勢力狀態，調整本勢力 +0Ah、清 +22h | strong inference：勢力清理；不是 block10 直接 writer |
| sub_353C4（0x353C4..0x3562B） | 領袖失效後顯示確認，玩家確認才在 block10 替換新領袖；設 byte_70056 | confirmed：確認門與替換寫入；事件來源未知 |
| sub_3562B（0x3562B..0x358FB） | 逐勢力呼叫戰鬥／結算鏈，保存 word_6FF8A 結果，之後使用 block10 槽值並呼叫 sub_353C4 | strong inference：劇本／名冊協調器；完整存檔時序未知 |

### 2.2 尚不能宣稱完成的時機

目前已證實「block10 會被初始化、第一期新局選項填入、讀取、移除與在接班確認後
替換」。仍沒有一條正常玩家 oracle 能同時證明：

1. 其他時期／載入分支的劇本資料何時第一次寫入 10 槽；
2. 勢力覆滅／領袖死亡與月結算入口的實際先後；
3. block10 更新後何時呼叫 `sub_595D4` 回寫 DT1；
4. 讀檔後十槽是否全部保留，或只在後續流程壓縮有效前綴。

2026-08-11 補充：IDA 直接 code xref 顯示 `sub_595D4` 目前只有
`sub_1B399@1B6BAh`（存檔選單確認後）與 `sub_5B1B6@5B5CEh`（`h..q` 快捷存檔）
兩個呼叫端。故「玩家明確存檔時 block10 快照落到 `.DT1 +3946h`」已 confirmed；
月結算／接班完成後是否另有自動或間接存檔仍 unknown，不能以這兩個呼叫端推定。
詳細 byte 搬運與限制見 `docs/re/42` §2.3。

因此 DT1State.WriteDT1 的 block10 目前仍是非破壞性 snapshot writer：只有呼叫端
提供已確認的 runtime 值才寫入，未解時保留 orig；不能把 block10 的存在誤報成
「劇本同步時機已完成」。

## 3. DT2／MEM_WAR.DAT oracle 邊界

sub_3964E（0x3964E..0x39B0B）是戰鬥狀態寫回鏈。已知控制流是：回合上限且
byte_64901==0 進入寫回；立即撤退與已有勝方的路徑不必然進入。469 bytes 中
runtime ID 與部分資源／roster 位置已確認，但 SlotsA/B、尾端與正常戰鬥後的
同步時機仍缺第二份固定樣本。

要把這一項升級成 confirmed，必須在隔離 DOSBox 內保存至少兩份固定狀態：

- 正常攻擊後仍有雙方存活的回合上限平局；
- 正常攻擊導致一方全滅／補給見底的勝負結算；
- 讀檔前後的 SAVE(N).DT2 與 MEM_WAR.DAT，並記錄輸入檔 SHA-256、DOSBox 版本、
  起始月份／回合／參戰名冊與完整 byte diff。

在使用者要求暫停 DOSBox 取樣的期間，以上屬 pending oracle，不是 remake writer 的失敗。
