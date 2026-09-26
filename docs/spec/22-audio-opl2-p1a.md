# SPEC-22：純 Go OPL2 風格音源 P1a（離線 PCM）

狀態：**READY**  
日期：2026-08-10

## 1. 目的

在不使用 `cgo`、不依賴音效裝置、也不把原版音訊檔案放進發行包的前提下，建立
`.MUS`／`.TIM` 到 deterministic PCM 的第一條可測試路徑。這是 `docs/spec/20`
解碼層之後的 P1a；它先閉合「事件排程 → 2-operator FM 風格聲音 → 16-bit stereo
PCM」的純 Go 邊界，供後續 Ebiten `audio.Player` 適配器使用。

本切片**不宣稱**與原版 `SDFA.EXE` 的暫存器串流或逐樣本聲音等價。`An` 音量映射、
完整 OPL2 envelope、原版循環規則仍屬 remake approximation；本輪的交付判準是
「音樂與音效能播放」，不再以解包 `SDFA.EXE` 或追求寄存器等價作為阻塞條件。

## 2. 輸入與證據界線

- `mus.Song` 的絕對 tick、`mus.TimbreBank` 的 28-word 音色欄位是
  `docs/formats/06-mus-tim-audio.md` 與 SPEC-20 的 confirmed 輸入。
- 音色欄位到 OPL2 兩個 operator 的 bit packing 依公開 YM3812/AdLib 暫存器格式實作，
  在程式註解標為 strong inference；不得拿來取代原版暫存器錄製。
- MIDI note 到頻率、`An` 到線性增益、attack/release 的簡化是 remake approximation，
  必須留在 `internal/audio/opl2`，不得污染 `internal/game` 規則層。
- 輸入事件必須按 tick 單調處理；錯誤的音色索引、零 tick rate、超出合理 sample
  上限都 fail-closed。

## 3. API 與分層

- `RenderAdLib(musData, timData, RenderOptions) ([]byte, error)` 是 MUS/TIM 到 PCM 的
  直接播放入口；AdLib 與 OPL2 在此層視為同一個純 Go software path。
- `RenderSong(song, bank, RenderOptions) ([]byte, error)` 產生 little-endian
  signed 16-bit、雙聲道 PCM；不 import Ebiten。
- `Chip` 暴露可重現的旋律與 rhythm `NoteOn`／`DrumOn`／`Render` 邊界，供 Ebiten
  adapter 使用；`ProgramRegisterWrites` 只作可選的靜態研究工具，不是播放前置條件。
- 音訊層只能依賴 `internal/audio/mus` 與標準函式庫；不得讀取原版路徑、寫檔或
  啟動裝置。
- Render 預設以 `song.ActualTick` 為結束位置；呼叫端可用 `MaxFrames` 限制離線
  產量，避免 malformed metadata 造成無界配置。

## 4. 本切片行為

1. `Cn` 切換音色；`9n` note-on、`8n` 或 velocity 0 note-off；`An` 以明示的線性
   0..127 增益近似；`En` 以 pitch bend 範圍近似。
2. `F0 7F 00 <整數> <分數/128> F7` 只調整後續 tick-to-sample 的速度倍率；未知
   SysEx 保留為 no-op，不猜測其語意。
3. 每個旋律聲道最多一個 active voice，符合原版 OPL2 9 聲道的外觀限制；打擊模式
   的聲道 6..10 使用 deterministic 純 Go kick／snare／hat／tom／cymbal 合成，
   讓含鼓曲目實際有音效輸出。其 envelope 是 remake approximation，不宣稱原版鼓組等價。
4. 輸出固定雙聲道同值（mono folded to stereo），不引入未規格化的空間化玩法。

## 5. 驗收

- synthetic MUS／TIM 測試覆蓋：program、note-on/off、volume、pitch bend、tempo
  SysEx、malformed input、`MaxFrames` 上限與 deterministic PCM digest。
- 8 首真實素材只驗證「可解析、可在有限 frame 預覽輸出」；不把音色或逐樣本結果
  與原版相等列為本切片通過條件。
- Docker 內 `go test ./internal/audio/...`、`CGO_ENABLED=0` 的純 Go package build、
  `tools/check_no_cgo.sh` 與 `git diff --check` 通過。

## 6. 非本切片範圍

- Ebiten `audio.Context`／`audio.Player` 的初始化、`-audio=off`、曲目情境切換與
  交叉淡入淡出（已由 `docs/spec/26-audio-context-switch-m1.md` 接續完成，非本切片的
  驗收範圍）。
- `SDFA.EXE` 解包、DOSBox OPL register trace、原版 `An` 映射與逐樣本 parity；這些是
  可選的未來精度研究，不是直接播放的完成門檻。
- `audio=modern` 的重新編曲、Ogg 資產與跨平台音訊裝置驗收。

## 7. 2026-08-10 精度邊界與範圍決策

`internal/audio/opl2.ProgramRegisterWrites` 已加入純 Go 的靜態寄存器打包器：依
公開 YM3812 layout 將 TIM 的 KSL／multiple／feedback／envelope／wave bitfield
輸出為 9 聲道、兩 operator 的 `RegisterWrite{Address, Value}` 序列，並忽略已知
未初始化的 carrier feedback／connection 殘值。它的證據等級是 **strong inference**；
沒有把 SDFA 的寫入順序、AdLib port delay、F-number／Key-On、`An` 映射或循環時序
偷偷升格為 confirmed。這讓未來取得 register trace 時可以替換 adapter，卻不必改動
MUS/TIM parser、PCM 測試或遊戲規則層。**本輪決策是直接採用 `RenderAdLib` 與
純 Go OPL2 風格合成；不為了 SDFA 的寄存器等價再建立解包／取樣阻塞。** 若日後
真的需要硬體精度，另開獨立研究支線，不能回頭把它變成玩家可玩路徑的前置條件。
