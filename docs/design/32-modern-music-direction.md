# Modern 音樂版本規劃

> 狀態：**DRAFT（Modern procedural composition／FX runtime 已接；正式 Ogg 與人審仍待完成）**
> 日期：2026-08-11
> 範圍：`audio=modern` 的原創配樂、音效層、情境切換與發行 QA。
> 不取代 `audio=retro` 的純 Go OPL2／AdLib 路徑，也不把新音樂稱為 1992 年原聲帶。

## 1. 設計目標

Modern 音樂要讓新玩家更容易進入民國初年的軍政、都市與人物敘事，同時保留歷史
氣味。它是**新的 remake 差異**，不是原版旋律的重配、放大或「AI 修復」。原版
`SCENE`、`STRATEGY`、`BATTLE1`、`BATTLE2`、`WALL`、`FINAL` 與 `MAINTHEM` 的
AdLib／OPL2 事件只作時長、情境與硬體語彙的研究參照；不得逐音引用、改寫主旋律或
取樣原版音源。

建議的聲音身份是「1920–40 年代城市樂隊與軍樂的語法，經現代遊戲混音重建」：
五聲／自然小調的片段可以作色彩，銅管、木管、鋼琴、低音弦與節制的打擊建立時代感，
但不使用真實國歌、軍歌、政治宣傳曲或以音樂替勢力貼道德標籤。音樂表達場景壓力，
不替玩家判定哪個歷史人物是正邪。

## 2. 建議的原創聲音識別

### 2.1 新主題動機

- 寫一個全新、可逆向辨識的 3–5 音動機；先以四度／二度的短呼吸建立「行動尚未
  結束」感，避免與原版或已知軍歌相似。
- 動機只作新曲之間的共同語彙，不在遊戲規則中表示勝負、勢力或隱藏資訊。
- 先製作無鼓、鋼琴／低弦的 MIDI sketch，再做銅管、木管與打擊配器；每一版都保留
  `motif-id`、作者、日期與相似性自檢紀錄。

### 2.2 情境 cue 矩陣

| Cue | 遊戲畫面 | 速度／性格 | 核心編制 | 循環建議 |
|---|---|---|---|---|
| `modern_scene` | 開局／一般敘事 | 88–96 BPM，稀疏、帶廣播室內感 | 鋼琴、低弦、單簧管、極淡的磁帶噪聲 | 48–64 秒，8 小節無縫循環 |
| `modern_strategy` | 政略地圖／選單 | 72–84 BPM，規整但不亢奮 | 馬林巴或鋼琴脈衝、低音提琴、短銅管回答 | 64–80 秒，8 小節邊界 |
| `modern_battle_a` | 戰鬥開端／一般交鋒 | 104–116 BPM，行進與緊張 | 低銅管、短弦樂 ostinato、軍鼓／大鼓、木管反拍 | 48–64 秒，8 小節邊界 |
| `modern_battle_b` | 戰鬥後段／壓力升高 | 120–132 BPM，密度增加但不改規則 | `battle_a` 的新變奏、額外低音與金屬打擊 | 48–64 秒，8 小節邊界 |
| `modern_story` | 人物自傳／新聞畫廊 | 60–72 BPM，留白、可讀性優先 | 鋼琴、二胡或單簧管單線、低弦長音 | 40–56 秒，淡入後可停在循環點 |
| `modern_final` | 結算／結局 | 68–80 BPM，克制的收束 | 全編制但保留空間，不寫成勝利頌歌 | 32–48 秒，明確尾奏＋可選循環 |

`battle_a`／`battle_b` 的差異只能是配器與密度，不能讀取或洩漏 AI 內部狀態。若
戰鬥勝負尚未判定，不提前播放 `modern_final`；畫面情境由目前已接的 `Track` 映射
決定，缺檔時沿用目前曲目並保留 stderr 診斷。

### 2.3 分軌與音效

第一版採兩個可獨立關閉的 stem：`music` 與 `fx`。不做依兵力／血量自動改配器的
動態分軌，避免新增玩法回饋與測試維度。純 Go OPL2 已能播放原版風格音效；modern
FX 可先以原創短音（按鍵、確認、撤退、頁面翻轉、地圖游標）補足現代 UI，不引用
原版取樣，並由 `-audio=off` 一次關閉。

## 3. 技術交付形狀

```text
assets/music/modern/        （日後新增，現在不提交未清權音檔）
  modern_scene.ogg
  modern_strategy.ogg
  modern_battle_a.ogg
  modern_battle_b.ogg
  modern_story.ogg
  modern_final.ogg
  manifest.json              （版本、作者、來源、loop sample、授權）

internal/ui/audio/           Manager 已有 bounded crossfade；Modern Ogg 已沿用同一入口
cmd/dsds                    -audio=modern|retro|off -modern-audio <dir>
```

格式建議 Ogg Vorbis、48 kHz、立體聲、約 160 kbps；執行期以 manifest 記錄並驗證
`author`／`license`／SHA-256，loop metadata 使用每聲道 sample。若要以 Vorbis comment 記錄
`LOOPSTART`／`LOOPLENGTH`（sample）、`CUE_ID`、`MOTIF_ID`、`MIX_VERSION`。執行期
只接受 manifest 所列的檔案與數值；缺檔、錯 loop 或不支援 codec 時 fail-closed，退回
`audio=off` 或目前可用曲目，不把壞檔變成啟動崩潰。

`docs/spec/32-modern-ogg-runtime-m1.md` 定義目前已實作的 loader、純 Go Vorbis
decode 與循環 reader；`docs/spec/35-modern-procedural-audio-sfx-m1.md` 另定義沒有
Ogg 時的原創純 Go cue／效果音 fallback。現行 fallback 已把六音動機、低音進行、
脈衝、銅管進場與 13–14 小節真空編成 24 小節情境 loop；它是可直接播放的 runtime
composition 與播放管線完成證據，不代表正式 Ogg
作曲、作者／授權或人耳混音 QA 已完成；儲存庫仍不放 placeholder `.ogg`。

情境選曲也已接到正常玩家路徑：地圖／政略主頁使用 `main-theme` 豪情 cue，
人物／將領頁使用 `wall`，新聞／史事頁使用 `scene`，政略子流程使用 `strategy`，
戰鬥使用三條 battle cue，離開確認使用 `final`；若沒有 manifest Ogg，八條 cue 都由
同一份純 Go composition lazy 產生。這是 runtime 接線，不是人耳驗收或正式清權宣告。

現有音訊 manager 的 18 frame（約 300 ms）交叉淡入淡出可作第一版 cue 切換；這是
remake 聽感選擇，不是原版 AdLib 時序。切換曲目從新 cue 的零點開始，暫不承諾跨 cue
同小節接續；若日後要做到小節對齊，先在 manifest 增加 BPM／拍號／loop bar，再加
離線測試，不在播放執行緒猜測。

## 4. 製作提示與工作流

### 4.1 編曲 brief（給作曲／生成工具）

> 原創 1920–40 年代中國城市軍政題材遊戲配樂；中速、克制、帶鋼琴與低弦的室內
> 樂隊，局部使用五聲語彙與西式銅管／軍鼓的節奏語法。寫一個全新 3–5 音短動機，
> 不引用任何現有歌曲、國歌、軍歌或《大時代的故事》旋律。保留 8 小節乾淨循環點，
> 避免過度電影式英雄和聲、民族刻板配器與對勢力的道德暗示；讓玩家能在閱讀 UI
> 與人物文字時聽清中頻。交付 MIDI、分軌 WAV、最終 Ogg、loop sample、BPM／拍號、
> 音色／樣本來源與授權紀錄。

這段 brief 是方向，不是已產生的音樂；任何 AI 生成草稿都必須人工聽辨、做旋律相似性
檢查並保留生成平台、模型、提示、日期與輸出雜湊。沒有完整 provenance 的檔案不能進
發行包。

### 4.2 首支 cue 的編曲師交付

`docs/music/modern_scene/` 現在保存首支 cue 的可重生草稿包：

- `README.md`：技術參考、全新六音動機、24 小節配器 brief、排除項與人耳閘門。
- `score.json`：固定 96 PPQN、92 BPM、8 小節／3,072 tick 的音符與低弦導引資料。
- `modern_scene_draft.mid`：只含鋼琴／低弦的原創 MIDI sketch，不是 Ogg 或 playable asset。
- `provenance.json`：來源排除、產生器、雜湊、授權與尚未完成的 QA 狀態。

這份交付只證明作曲方向與可重生草稿已落檔；它沒有把原版旋律轉寫成新音色，也沒有
解除正式作者／音源授權、相似性聽審、loop 混音或裝置驗收閘門。

### 4.3 代理驗證與推廣片預覽

使用者已授權由代理執行技術驗證；本輪以 `FluidSynth 2.3.1` 在 Docker 內把 MIDI 草稿
渲染成 48 kHz／立體聲 Ogg，並以 FFmpeg 檢查 Vorbis 解碼、三圈重播、mono 路徑、整合
響度與 true peak。結果為 `-18.1 LUFS`、`-5.9 dBFS true peak`；完整音訊的首尾樣本
差為 0，三圈串接可解碼。這是**代理技術驗證**，不是人類主觀聽感或法律 clearance。

舊的 `modern_scene` 實機預覽集中在被忽略的 `dist-all/promo/gameplay/` 作歷史紀錄；
目前的連續錄製基準片為
`dist-all/promo/gameplay-modern-strategy-v2/great-era-remake-modern-strategy-playthrough.mp4`。
它從復古地圖經 `F2`／`F3` 到 Modern 高解析策略儀表板，再以滑鼠進入政略卡、查閱、
將領詳情、人物自傳與翻頁，最後由正常政略指令進入高解析戰鬥並操作攻擊選擇與結束回合。
畫面沒有接入靜態成果 PNG 或 mockup。

該基準影片以本機 `modern_campaign_loop.ogg` 另行混音，藉由低密度前奏、留白與銅管／
低弦高潮示範豪情方向；它不等於遊戲已打進正式 Ogg，也不把 `draft-not-for-release` 草稿
稱為正式 soundtrack。使用者已退回現有 Modern 視覺驗收，故新版 UI 完成後必須以完整
正常玩家流程重新錄製，不得把本片當作最終宣傳成品。影片／音檔均不進版本庫或 release，
公開前仍需作者與音源權利、盲聽相似性、人耳混音與各平台播放驗收。完整規格、雜湊與實機畫面範圍見
[`docs/promo/README.md`](../promo/README.md)。

### 4.4 每首曲目的交付包

1. `*.mid`：可編輯母稿，標記 cue、motif、loop bar。
2. `*_stems.zip`：music／fx 或編制 stem，內含 `README` 與音源授權。
3. `*.ogg`：遊戲播放版與 loop metadata。
4. `manifest.json` entry：作者、演出／生成工具、來源、授權、版本、SHA-256。
5. QA report：loop、峰值／整體響度、mono 相容、耳機／喇叭／Android 裝置抽聽。

## 5. 驗收閘門

- **文化／授權**：無原版 `.MUS`／`.TIM`／PCM、無未授權既有旋律、無真實政治歌曲；
  `tools/deny_scan.sh --all` 通過，manifest 可回溯每個音色／樣本來源。
- **音樂**：每首至少連續循環三次無爆音／空拍；8 小節邊界在波形與人耳都平順；
  modern cue 不改變規則或 AI 可見資訊。
- **混音**：立體聲不過寬，轉 mono 不消失主旋律；UI 語音／文字閱讀區保留 headroom；
  音量與 `audio=retro`／`audio=off` 的設定一致。
- **執行**：Docker 無頭可載入 manifest、缺檔可降級、18 frame crossfade 在 bounded
  frame 內關閉舊 player；Windows／macOS／Linux 的音訊初始化錯誤不阻塞 `-audio=off`。
- **發行**：現代音檔屬新增資產，與玩家自備原版素材分開；三平台包與 Android
  實機聲音驗收另開 release gate，不能以 Linux 播放成功代替。

## 6. 分期與目前狀態

| 階段 | 產出 | 狀態 |
|---|---|---|
| M0 | 原版 MUS／TIM 純 Go 解碼與 OPL2 風格直接播放 | 已完成（`audio=retro`） |
| M1 | 情境 cue、Wall／Final、bounded 300 ms crossfade | 已完成 remake 串接 |
| M1a | `audio=modern` manifest／Ogg loader、循環 metadata、safe fallback | 已完成 runtime 接線；缺 Ogg 時改用原創純 Go cue |
| M2a | `modern_scene` 新五音動機、8 小節 MIDI sketch、brief 與 provenance | 已交付草稿；代理技術驗證完成，真人聽感仍未宣稱 |
| M2c | `modern_scene` 預覽 Ogg 與實際遊玩推廣片 | 已產生本機預覽；不屬發行資產 |
| M2b | 其餘 `strategy`／`battle_a`／`battle_b`／`story`／`final` MIDI sketch | runtime 已由同一原創動機產生 24 小節 cue；正式 MIDI／Ogg 編曲仍待作曲 |
| M3 | 正式混音、loop、Ogg／manifest、modern FX | procedural 24 小節 cue／8 類效果音已可播放；正式外部音檔與人耳／授權 QA 待完成 |
| M4 | 多平台／Android 音訊 QA 與發行包 | 待打包里程碑 |

在 M2 取得可聽 prototype 之前，不把「Modern 音樂內容」寫成已完成，也不為了填入暫時
音檔而改動 `internal/game` 規則層；目前 `audio=modern` 缺 manifest／cue 時會安全使用
可重生 procedural composition，只有 runtime 失敗才回報並保留既有曲目。`SDFA.EXE` register parity 仍是獨立研究，不是 Modern 音樂或 remake
播放前置條件。

## 7. `modern_campaign` 豪情方向（2026-08-11）

使用者指出先前推廣片配樂缺少《大時代的故事》的豪情感；`modern_scene` 的內省／閱讀
職責不變，另新增 `docs/music/modern_campaign/` 作為政略／推廣片的原創 cue 草稿包。
它以「電報紙張、木製指揮桌、遠方鐵路」為聲音材料，採全新動機
`D4–G4–F4–B♭4–A4–D5`（`+5,-3,+6,-1,+5`），112 BPM、24 小節，明確安排
「低密度前奏 → 集結 → 行動 → 兩小節真空 → 銅管／低弦／軍鼓高潮 → 未解決回望」。

本輪在 Docker 內以 FluidSynth／FluidR3_GM 渲染本機 loop 預覽，並做 0.5 秒邊界
crossfade；48 kHz 立體聲 Vorbis、約 51.43 秒，`-15.8 LUFS`、`-1.7 dBFS true peak`。
輸出只在被忽略的 `workplace/promo/modern_audio/modern_campaign_loop.ogg`，SHA-256
與 stem／授權待辦見 `docs/music/modern_campaign/provenance.json`。這是可聽技術預覽，
不等於正式作者、音源清權、盲聽相似性或 Android 聲音驗收完成。
