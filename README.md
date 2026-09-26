# 大時代的故事 remake

漢堂國際資訊《大時代的故事》(1992, MS-DOS) 的逆向、引擎重寫與多語系專案。
在 Go / Ebiten 上重寫一套跨平台引擎，還原原版繁體中文文本，另做英文與日文語系。

1992 年 10 月，漢堂出了這款以民國初年為背景的回合制戰棋。它扭轉了前作《隋唐群雄傳》
的評價，也是《天外聖劍錄》《炎龍騎士團》之前的那一部。三十四年過去，原始碼沒有留下來，
發行商也已經不在，剩下的只有一份 148 個檔案的遊戲目錄。

這個專案要把它讀懂，然後在現代平台上重新蓋一遍。

## 讓歷史說今天的人聽得懂的話

民國初年的白話文運動面對的，不只是「文言還是白話」的文體選擇，而是知識能否走出少數人
熟悉的書寫慣例，進入教育、報刊與日常生活。它留給今天的一項重要啟示是：保存文化，
不等於把知識鎖在舊有的表達形式裡；傳播技術與閱讀習慣改變時，語言也要繼續發展，
更多人才有機會理解、討論，再把文化傳下去。

一百年後，媒介從紙張走到電腦與手機，遊戲介面也形成了新的閱讀語法。若玩家必須先猜懂
「司令，欲調動何將？」才能開始理解遊戲，那麼被保存下來的可能只是一件可觀看、卻不再
容易進入的舊物。尤其對 2000 年後出生、第一次接觸民國初年人物與軍政制度的玩家而言，
介面的陌生不應成為認識歷史的資格考試。

因此重製版不把現代化理解成「把舊東西換掉」，而是讓兩種入口並存：

- **圖形可切換**：原版點陣圖、現代重繪與軍事符號並存，不覆蓋原始風格。
- **語言可切換**：原典文句完整保留；現代白話可把「何將」寫成「哪位將領」、
  把「欲如何調動？」寫成「要怎麼調動？」。
- **規則不切換**：無論使用哪一套圖形或文字，候選項、數值、結果與存檔都完全相同。

兩套呈現不是「舊版被新版取代」，而是通往同一份內容的兩個入口：現代圖形先幫玩家看懂
地形、部隊與狀態，現代白話先幫玩家理解要做的事；玩家可以隨時切回原版點陣圖與原典
文句，觀察 1992 年的美術、字模與電腦中文介面。切換也不必綁在一起——可以用現代圖形
讀原典，也可以保留復古畫面而改用現代白話。

這不是替歷史人物改寫發言，也不是用今天的價值觀補寫史料。現代白話只處理操作介面的
理解門檻；人物、事件、地名與制度仍以可追溯的史料呈現，推論與未知也必須如實標示。
原典永遠保留、差異可以回查、玩法不因呈現方式而改變。

因此，這個重製專案想保存的不是一包只能被模擬器打開的舊檔案，而是一段仍能被遊玩、
被閱讀、被提問，也能讓下一個世代主動走近的文化與歷史。遊戲在這裡不只是歷史題材的
包裝，而是一種讓文化重新進入日常經驗的媒介。

完整的切換架構與差異契約見
[`docs/design/10-visual-modernization.md`](./docs/design/10-visual-modernization.md)。
高解析度 Modern／Android 的 1280×720 設計畫布、safe-area、觸控 48 dp 與目前 gate 見
[`docs/design/33-high-resolution-modern-android.md`](./docs/design/33-high-resolution-modern-android.md)。

## 目前進度

| 里程碑 | 狀態 |
|---|---|
| M0 資產解密 | **完成**——`.15` 字模、`.TPC`／`.RGB`、`.GLB`／`.GTB`、`RAIL.TPC`、`PLACD.SAV`、`NEWSDATA.DAT` |
| M1 文本還原鏈 | **詞表層完成**；敘事文本已找到（是畫成圖的，不是字模序列）|
| M2 執行檔反組譯 | 進行中——五支已盤點，戰鬥模組已定位 |
| M3 規則規格 | 4 份 READY（地圖、將領、省份、戰場圖塊）+ `docs/mechanics/` 機制文件 |
| M4 規則層 | 進行中——地圖、將領、省份、戰場、戰損鏡像與世界層結算已接；原版全流程仍未宣告 |
| M5 呈現層 | **政略／戰鬥 + Modern UI 垂直鏈**——15 項指令選單、戰場圖塊、參戰單位、攻擊目標標號、人物自傳頁、modern 資訊卡／戰鬥外殼／十勢力色帶、新聞畫廊；外交已接選單／貸款 M0／信用度零 gate M1／貸款與外援結果成本 M1／外援與償債 M1／停火 M1，第一期援助代碼已接、國家名稱仍是假說 |
| M6–M7 多語系／發行 | `translations/en`／`translations/ja` 已交付 UI／面板／敘事語意鍵與 39 省名；英／日 387 篇人物自傳 machine-draft overlay 會明示來源；Windows／macOS／Linux 候選包已產生，但正式人審譯稿、Android 與三平台實機驗收仍待補 |
| 現代圖示／Modern UI | **Modern + high 的 H1-a 地圖／資訊卡、H1-b 指令／設定／自傳／敘事、H1-c 戰鬥 HUD 與 H2 一般流程已接通**；設計畫布固定 1280×720，`retro` 仍為 640×350；`-theme`／`F2` 切換圖形、`-resolution`／`F3` 切換解析度，滑鼠／觸控共用同一 Action |
| Modern 音訊 | **runtime 接線與純 Go fallback 已接通**——`audio=modern` 優先讀取有 provenance 的 manifest／Ogg，缺檔安全改用原創 procedural cue；八條 cue／八類 FX 有離線 WAV 技術預覽，正式 Ogg、清權、人耳／真機 QA 仍未完成 |
| 原典／現代白話 | **操作介面與敘事入口已接通**；外交貸款 M0、信用度零 gate、貸款／外援結果成本、外援／償債 M1、停火 M1 已加入；`NEWSDATA.DAT` 17 張來源圖像可由 `N`／滑鼠／觸控開啟，#0／#9 有 caption，其餘 source-image fallback |
| 滑鼠／Android 觸控 | M0–M1 涵蓋選單、長清單與數字鍵盤；M2 已接戰鬥相鄰格移動／攻擊，Android 封裝與實機驗證待續 |

DOSBox 實測已通，可載入存檔進政略階段。詳細現況見 [`CONTEXT.md`](./CONTEXT.md)。

```sh
tools/go.sh run ./cmd/dsds -game workplace/orig/game      # 需要顯示器
tools/go.sh run ./cmd/dsds -game workplace/orig/game -theme modern # 現代地形／鐵路（P1）
tools/go.sh run ./cmd/dsds -game workplace/orig/game -resolution original # 原版 640×350 邏輯畫布（桌面預設 2 倍視窗）
tools/go.sh run ./cmd/dsds -game workplace/orig/game -theme modern -resolution high # 1280×720 Modern 高解析畫布
tools/go.sh run ./cmd/dsds -game workplace/orig/game -audio retro # 情境純 Go OPL2 風格 BGM（需自備 MUS/TIM；SCENE 必要）
tools/go.sh run ./cmd/dsds -game workplace/orig/game -audio modern -modern-audio assets/music/modern # Ogg 優先；缺檔使用純 Go Modern fallback
tools/go.sh run ./cmd/modern_audio -out /tmp/great-era-modern-audio -report /tmp/great-era-modern-audio/report.json # 重生 cue／SFX 與技術 QA
tools/go.sh run ./cmd/dsds -game workplace/orig/game -audio off   # 無音訊裝置／CI 路徑
tools/go.sh run ./cmd/dsds -wording plain                 # 現代白話操作介面與已證實敘事 caption
tools/py.sh tools/gen_locale_packs.py                     # 重產英／日語系包（含人物來源標記）
tools/go.sh run ./cmd/dsds -locale translations/en       # 英文 UI／地名
tools/go.sh run ./cmd/dsds -locale translations/ja       # 日文漢字優先 UI／地名
tools/go.sh run ./cmd/screenshot -province 26             # 無頭產出政略畫面 PNG
tools/go.sh run ./cmd/screenshot -province 26 -theme modern -eten workplace/eten # Modern 資訊卡／地形對照
tools/go.sh run ./cmd/screenshot -province 26 -theme modern -resolution high -eten workplace/eten # H1-a 1280×720 預覽
tools/go.sh run ./cmd/screenshot -province 26 -theme modern -resolution high -battle -eten workplace/eten # H1-c 1280×720 戰鬥預覽
tools/go.sh run ./cmd/screenshot -province 26 -menu        # 換成 15 項指令選單
tools/go.sh run ./cmd/screenshot -province 19 -units      # 加上參戰單位圖示
tools/go.sh test ./...                                     # 逐像素驗證
```

## 不含原版遊戲資料的三平台 release 候選包

要建立只含 clean-room 引擎、Modern atlas、語系、授權／操作說明的 Windows、macOS
與 Linux 發行候選包：

```sh
tools/package.sh 0.1.0-release dist-all windows-amd64,darwin-arm64,linux-appimage
```

輸出是 Windows amd64 ZIP、macOS arm64 `.app` ZIP 與 Linux x86_64 AppImage，另附全域
`SHA256SUMS-<版本>.txt`；每個包均有 `RELEASE-MANIFEST.txt`。`workplace/orig`、原版
資料／音樂／美術、倚天字型與任何 `.MUS`／`.TIM`／`.DAT`／`.OGG`／`.WAV` 衍生物都不會
進包。玩家仍須自行準備合法的 `-game` 資料目錄與 `-eten` 字庫；這是「引擎與 Modern
外殼」包，不是重新散布原版遊戲。Windows／macOS 實機 smoke、macOS 簽署、公證與公開
發行仍分列為後續 gate；完整格式、驗證與限制見
[`docs/release/README.md`](./docs/release/README.md)。

2026-08-11 已重建 `0.1.0-release` 的三個候選包，並驗證 Windows／macOS ZIP 的入口與
CRC、Linux AppImage 的解包 payload 與素材拒絕規則。它們可供持有合法遊戲資料的測試者
下載試跑；尚未取代 Windows／macOS 真機 smoke、macOS 簽署／公證或跨發行版 Linux
驗收。實際檔名、雜湊與 gate 見 [`docs/release/README.md`](./docs/release/README.md)。

政略地圖右上可按 `N`，或用滑鼠／觸控點擊「新聞／史事」開啟 `NEWSDATA.DAT` 畫廊；
17 張來源圖像每頁四張，`Space`／`PageUp`／`PageDown` 翻頁，`Esc`／`N` 返回。
目前 #0／#9 有可追溯 caption，其餘 15 張保留原圖並顯示 source-image fallback，
不把未知點陣字猜成敘事文本。完整契約見
[`docs/spec/31-narrative-gallery-m1.md`](./docs/spec/31-narrative-gallery-m1.md)。

在「查閱將領」清單或詳細資料頁按 `B` 可開啟人物自傳；詳細頁也有可用滑鼠／觸控點擊的
「人物自傳／人物生平」按鈕。左右鍵切換人物，`Space`／`PgUp`／`PgDn` 翻頁，
`ESC` 或 `B` 返回。這是重製版新增的唯讀文化保存功能，
不改變遊戲狀態；目前 486 個期別槽位中 485 個可接合、1 個「無省長」明確排除；387 位
有自撰生平（其中新增 61 篇由 `translations/zh-Hant/people-authored.json` 的可重生
overlay 接入），沒有可靠正文的人物會明示「查無可靠傳記記載」。
英／日語系的人物姓名仍保留遊戲歷史寫法；目前沒有可追溯的完整英／日譯稿，頁面會顯示
「繁中原文」來源提示，不把自動代換冒充傳記翻譯。

戰鬥畫面目前五項操作都可進入：`1` 移動、`2` 攻擊後按 `1..6` 選相鄰目標、`3` 撤退
（僅已閉合的 18 ← 19 樣本）、`4` 駐軍（選中單位前往第一個可達城市）、`5` 查閱；
滑鼠／觸控可點戰鬥命令、相鄰格與攻擊目標，攻擊子狀態會在敵軍格顯示穩定的 `1..6` 標號。六種攻擊選項目前共用已確認的 `Engage`
近身公式，這是明列的 remake 差異；戰鬥結束會把已解欄位非破壞性寫入
`-save` 同目錄的 `.DT2` 與 `MEM_WAR.DAT` 副本，立即撤退不寫回。完整邊界見
[`docs/spec/25-battle-state-writeback-m1.md`](./docs/spec/25-battle-state-writeback-m1.md) 與
[`docs/spec/27-battle-settlement-biography-m2.md`](./docs/spec/27-battle-settlement-biography-m2.md)。

`-audio retro` 會先播放 `SCENE`，進入戰鬥時依序嘗試 `BATTLE1`／`BATTLE2`／`BT02`，
其他畫面嘗試 `STRATEGY`；其餘 `MAINTHEM`、`WALL`、`FINAL` 配對可一併提供，缺檔時
沿用目前曲目。情境切換使用 bounded 18 frame（約 300 ms）交叉淡入淡出；這是 remake
聽感選擇，不宣稱原版 OPL2 register parity。`audio=modern` 的純 Go Ogg loader、manifest
雜湊／授權、循環 reader 與 procedural fallback 已接；`cmd/modern_audio -report` 會對八條
cue／八類效果音輸出峰值、RMS、mono、首尾差與 SHA-256；儲存庫不放入未清權音檔，缺 manifest
時使用同一份原創純 Go cue。執行期契約見 [`docs/spec/32-modern-ogg-runtime-m1.md`](./docs/spec/32-modern-ogg-runtime-m1.md)，
音樂內容方向見 [`docs/design/32-modern-music-direction.md`](./docs/design/32-modern-music-direction.md)。
目前另有 `modern_campaign` 豪情 cue 的本機預覽草稿；它只留在 ignored
`workplace/promo/modern_audio/`，尚未清權，不會冒充正式發行音樂。

在政略指令選單按 `O` 可開啟 remake「顯示設定」，於遊戲中切換原典用語／現代白話，
以及原版圖形／現代圖形；鍵盤可按 `1`／`2` 選用語、`3`／`4` 選圖形，滑鼠／觸控
命中同一組選項。`F2` 是主題切換快捷鍵；`F3` 在原版 640×350 與高解析 1280×720
畫布間切換，兩者都不改變規則或存檔。也可以從指令 15 → 顯示設定 → 解析度切換進入，
因此沒有實體鍵盤的 Android 路徑仍可使用同一個偏好設定。
選擇會寫入 `$XDG_CONFIG_HOME/dsds/prefs.json`（未設定 XDG 時使用平台設定目錄），
不寫入遊戲存檔；`-wording` 可在單次啟動時優先覆寫偏好。
十五項政略主選單、調動、運補、徵兵／重新整編、查閱、發展、政策／自治／產能、
商業、練兵、秘密行動、人物自傳、外交選單／貸款 M0／信用度零 gate／貸款與外援結果成本／外援與償債 M1／停火 M1 及其他選項 O0–O2
已接上雙用語；媒體選項 O3／O4 與未解新聞模板仍保留 source-image fallback。這裡的「現代白話」
已涵蓋可玩的操作／面板與 #0／#9 敘事 caption，但不能宣稱全遊戲人物／劇情已有正式人審譯稿。

在政略指令選單按 `8` 進入原版「政策」：第一項「授權自治」已可操作，
能在同一項指令內切換多個合法省份；第二項「產能分配」也已完成，可調整鐵礦、煤礦、
石油與糧食百分比，黃金為剩餘產能；四個存檔欄位與
五種資源對應已完成 IDA 證據與真實存檔閉環。

## 目前畫得出什麼

載入民國 15 年 7 月的原版存檔，湖北省。左邊是政略主畫面與 15 項指令選單，
右邊是同一個省的六角格地圖：

![政略主畫面與指令選單](docs/images/menu-26.png)

![湖北省的六角格地圖與部隊圖示](docs/images/province-26.png)

Modern 高解析度對照（1280×720）：

高解析人物自傳成果：`docs/images/high-modern-biography-058-01.png`

![Modern 高解析度湖北地圖與資訊卡](docs/images/high-modern-province-26.png)

![Modern 高解析度政略指令卡](docs/images/high-modern-menu-26.png)

![Modern 高解析度戰鬥 HUD](docs/images/high-modern-battle-26.png)

![Modern 高解析人物自傳](docs/images/high-modern-biography-058-01.png)

> 以上成果圖都是 `cmd/screenshot` 在無頭環境合成的，沒有經過 DOSBox。
> 畫面上的中文字模與地形、部隊圖示都來自 1992 年的原版檔案——
> 這是文化資產保存的成果展示，本專案不散布原版的執行檔或資料檔本身
> （見[授權與素材](#授權與素材)）。

> 2026-08-03 已用目前 renderer 重產。前兩張刻意展示復古圖形路徑；後兩張是
> `cmd/screenshot -theme modern -resolution high` 產生的 Modern 1280×720 展示證據。
> Modern 高解析度地圖／指令／戰鬥玩家路徑與可再散布 `GEMF` 比例字型 atlas 已接通；人物自傳改用
> GEMF 像素字距換行，英／日 387 篇 machine-draft overlay 的 glyph coverage 與最長
> 自傳首／末頁已由測試鎖定，逐篇人審／長文截圖、Android
> 實機與正式音檔仍是發行 gate。字型來源與 hash 見
> [`docs/licenses/modern-font-atlas.md`](docs/licenses/modern-font-atlas.md)。

人物自傳首頁／末頁也可由同一份 `PeopleDB` 重生高解析成果圖：

```sh
tools/go.sh run ./cmd/screenshot -game workplace/orig/game -province 26 \
  -theme modern -resolution high -biography -out /tmp/high-biography
```

這個展示入口只讀既有人物資料並呼叫 `DrawModernBiographySurface`，不另造人物文字或
規則路徑；`docs/images/high-modern-biography-058-01.png` 是湖北 #26 司令的首頁成果。

政略畫面的 13 個欄位**全部與 DOSBox 實機截圖一致**：

| 欄位 | 湖北（民國 15 年 7 月存檔）|
|---|---|
| 司令／省長 | 吳佩孚／吳佩孚 |
| 黃金／糧食／彈藥 | 4200／18050／8787 |
| 燃料／煤礦／鐵礦 | 12048／13000／14031 |
| 地價／人口 | 22／1825 萬 |
| 城市／兵工廠 | 5／3 |
| **兵力／將領數** | **97500／15** |
| 人民忠誠度 | 79 |

兵力與將領數不是存檔裡的欄位——是把 `MAN(1).DAT` 裡所屬省為 26 的
15 位將領兵力加總出來的，所以會超過 u16。

右側戰場用原版的 `NEWTERR.TPC` 圖塊畫，鐵路疊 `RAIL.TPC`：
湖北看得到長江橫貫、六條鐵路連成路網、跨江處是鐵橋。

## 已經解開的：文字是怎麼存的

這款遊戲整個目錄用 Big5 解碼，中文命中數是零——但畫面上滿滿都是中文。

原因不是加密，是它根本不存碼點。文字直接以**字模序列**存在，每 `w` 個 16×15 點陣字模
組成一個詞條，不足的格子填全零字模。程式從 `k × w × 30` 取 bytes 直接送 BGI `putimage`，
中間沒有任何索引表。

51 個字模檔、6,174 個字模已全量反查回 Big5：

| 內容 | 數量 |
|---|---|
| 39 個省份的地名表（城市、山川、鐵路、關隘）| `TN15.1`–`TN15.39` |
| 中國將領姓名（分三個時期）| 274 + 106 + 106 人 |
| 日本將領姓名 | 170 人 |
| UI 詞彙（二字／三字／四字分檔）| 231 + 57 + 51 條 |
| 部隊番號（含「第␣集團軍」這類數字佔位模板）| 75 條 |

字形 100% 出自倚天中文系統的 16×15 字庫，唯一的例外是一個逗號——經二維位移窮舉確認，
它是倚天「，」左移 3 px、下移 2 列的版本。

完整規格見 [`docs/formats/01-glyph-text.md`](./docs/formats/01-glyph-text.md)。

## 已經解開的：一張戰場是怎麼組出來的

每個省有一張 14×14 的戰場，資料分散在三個檔案，各管一件事：

| 檔案 | 每格存什麼 |
|---|---|
| `WARPOS.DAT` | 這一格屬於哪個鄰省（0 = 本省腹地）——同時也是鄰接表的來源 |
| `TERNAME.DAT` | 地名索引，指向該省的 `TN15.<省>`（漢口、長江、大別山…）|
| `NWMAP.DAT` | 地物編號，決定畫哪一張 `NEWTERR.TPC` 圖塊 |

鐵路壓在同一個 u16 裡，用 `30 + 25×圖塊 + 地形` 編碼，拆出來疊上
`RAIL.TPC` 的鐵軌。驗證方式是把每省「地物 = 城市」的格數去對省份資料表的
城市數欄位——**39 個省全部吻合**。

湖北解出來是長江橫貫、五座城市、六條鐵路連成路網、跨江處是鐵橋；
河北是長城橫貫北緣配五座關隘。規格見
[`docs/spec/04-battlefield-tiles.md`](./docs/spec/04-battlefield-tiles.md)。

## 已經解開的：省份資料表

`TOWN(N).DAT` 是三個時期的初始省份狀態，`SAVE(N).DT1` 前段是同一個結構
（39 省 × 37 bytes，相位差 4 bytes）。六種資源、人口、城市數、地價、
兵工廠數、忠誠度、司令、省長、鄰省列表全部定名，對 DOSBox 實機截圖零誤差。

最有說服力的驗證不是數字對得上，是**司令欄位解出來就是 1926 年的割據圖**：
孫傳芳轄江蘇、安徽、浙江、福建、江西——史上的「五省聯軍」，一個不多一個不少；
張作霖轄東北九省加河北、山東；兩廣歸蔣中正。存檔裡省長會與司令分化成實際
駐守的人：遼寧張學良、山東張宗昌、江西盧香亭，湖南已易主為蔣中正／何應欽
——那正是民國 15 年 8 月北伐軍攻下長沙的時間點。

## 定位

文化資產保存。核心規則對齊 1992 原版，外殼（解析度、視窗、操作、存檔相容）現代化，
每一項改動都在文件裡標記為 remake 差異。不做玩法設計改動。

專案的工作紀律有兩條寫死的原則：**完整性優先於投報**（不以成本為由跳過任何素材），
以及 **spec 齊了才實作**（反組譯 → 規格 → 程式）。細節見 [`AGENTS.md`](./AGENTS.md)。

## 目錄

```
docs/formats/     檔案格式規格
docs/re/          反組譯筆記
docs/spec/        實作規格（標 DRAFT / READY，只有 READY 能動手）
docs/playtest/    實跑驗收紀錄
internal/assets/  純解碼層，不認識 Ebiten
internal/game/    規則層，不認識畫面
internal/ui/      Ebiten 呈現層
tools/            docker 包裝腳本與逆向工具
```

所有建置與工具一律走 docker，不裝進系統環境。

重製引擎與規則層本身是純 Go，`cmd/`／`internal/` 沒有 `import "C"` 或 `#cgo`。
目前 Ebiten v2.8.8 的 Linux／macOS 桌面 GLFW backend 仍有自己的 cgo 需求；Windows
amd64 已可用 `CGO_ENABLED=0`，完整目標矩陣與限制見 [`docs/release/README.md`](./docs/release/README.md)。
可用 `tools/check_no_cgo.sh` 重跑前述「本專案原始碼不引入 cgo」檢查；它刻意不掃
Ebiten module cache，避免把外部 backend 的實作誤報成遊戲程式碼。

## 授權與素材

程式碼、文件、逆向筆記、譯文與工具採 **RRSAL-1.0**（復古重製 source-available 授權條款，
SPDX `LicenseRef-RRSAL-1.0`）：非商業用途免費（含修改與再散布）；實況、影片、報導與平台分潤
明示允許，但要署名；商業用途需事先書面授權（wicanr2@gmail.com，歡迎來談）。這不是開放原始碼
授權，對外稱 source-available。條款全文、不涵蓋的原版素材與第三方元件見 [`LICENSE`](LICENSE)。

本專案**不散布原版執行檔、資料檔、美術或音樂，也不散布倚天字型**。
公開產出只有引擎程式碼、逆向筆記與翻譯文本，玩家需自備合法原版。
版控內容每次 commit 前都會跑 `tools/deny_scan.sh` 掃描原版資產。

**唯一的例外是 `docs/images/` 裡的展示截圖。** 那些畫面是由原版檔案
算出來的，因此含有原版的字模與美術。放它們的理由是：這個專案的成果
本來就是「能不能把原版畫回來」，沒有畫面就無從判斷進度。截圖不能拿來
還原任何一個原版檔案——它們是渲染結果，不是資產本身。

`deny_scan.sh` 對這條例外有專門的檢查：**PNG 只准出現在 `docs/images/`**，
別的位置一律報錯。這樣「例外」是白名單而不是漏洞。

## 致謝

漢堂國際資訊，以及當年做出這款遊戲的程式（林明輝、沈友瑋）與音樂（吳宗凱、游戲工場）。
還有倚天中文系統——三十四年後，讓這些字重新被讀出來的，還是同一套點陣字。
