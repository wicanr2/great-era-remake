# SPEC-35：Modern 純 Go 音樂 fallback 與效果音 M1

> 狀態：**READY（runtime slice）**　日期：2026-08-11

## 1. 目的

讓 `-audio modern` 在沒有尚未清權的 Ogg 時仍有可驗證、可聽的原創音訊；玩家不會
因 manifest 尚未提供而得到靜音。這是 remake 外殼的技術 fallback，不宣稱原版
`SDFA.EXE`、MUS／TIM、旋律或寄存器 parity。

## 2. 實作契約

- `internal/ui/audio/procedural.go` 以 deterministic 純 Go 合成八個情境 cue；每首以
  新作六音動機（D–G–F–B♭–A–D）為基底，依敘事／政略／戰鬥／結局使用不同速度、低音
  進行、脈衝密度與銅管進場，形成 24 小節、48 kHz stereo signed-16 loop。政略／推廣
  弧線保留 13–14 小節真空，再進 15–20 小節豪情高潮；不同情境以 bounded crossfade
  切換。
- 通過 manifest 的 Modern Ogg 優先；缺少 optional cue 時才 lazy 產生 procedural cue。
- `Manager.PlayEffect` 播放有界短效果音（確認、取消、按鍵、指令、移動、攻擊、命中、
  回合），最多保留 12 個 player；`Stop`／`Close` 逐一釋放。
- `cmd/dsds` 只在 `audio=modern` 將裝置無關 `Action` 映射到效果音；效果音以升降音程、
  掃頻與 deterministic 撞擊瞬態區分確認／取消／按鍵／指令／移動／攻擊／命中／回合；
  音效建立失敗不阻塞規則、存檔或輸入。

## 3. 發行與證據邊界

Procedural PCM 是可重生的原創 runtime composition，不是原版音樂轉檔；它讓遊戲在
沒有外部音檔時仍有完整的情境音樂／效果音路徑，但不取代日後可再散布 Ogg 的作者／
授權與混音資料。
若未來加入 Ogg，仍須遵守 SPEC-32 的 manifest、SHA-256、loop、授權與三平台／Android
裝置 QA。`SDFA` register parity 維持獨立研究，不是本規格前置條件。

## 4. 測試

`internal/ui/audio` 驗證 cue deterministic、24 小節長度、真空／高潮能量差異、不同
情境有差異、無 clipping、效果音短且 distinct，以及 player 上限與關閉；`cmd/dsds` 在
Docker＋Xvfb 驗證畫面與輸入接線。正式 Ogg 的人耳混音、授權與 Android 真機聽感仍未
宣稱完成。

`cmd/modern_audio` 可透過 `tools/go.sh` 在 Docker 內把同一份 renderer 重生為 48 kHz
stereo signed-16 WAV，並以 `-report <path>` 產生 JSON 技術 QA。`report.json` 逐檔記錄
frame／時長、峰值／RMS、mono RMS、首尾 sample 差、clipping、非靜音與 SHA-256；同一
份 `internal/ui/audio.AnalyzeStereoPCM` 也由單元測試覆蓋。它不會自動寫入 Modern
manifest，也不會把技術預覽升格為正式 Ogg、真人聽感或授權核准。

### 2026-08-11 Ogg 技術預覽（不等於發行）

本輪以 `dsds-go:1.25` 重生八條 cue／八類 FX，再在隔離的 `u5cht/video:latest`
（FFmpeg 5.1.9）使用 `libvorbis -q:a 5` 轉碼。`ffprobe` 確認全部輸出為 48,000 Hz
立體聲 Vorbis；每個檔案另以 FFmpeg 解碼至 null sink，未發現解碼錯誤。完整檔名與
時長表見 [`docs/music/modern_procedural/README.md`](../music/modern_procedural/README.md)，
輸出只存在被忽略的 `workplace/promo/modern_audio/procedural/`。

這項結果只關閉「renderer → WAV → Vorbis 可重生／可解碼」的工程風險，不關閉自然人
作曲與音源授權、三圈 loop 人耳聽審、跨平台／Android 音訊或正式 manifest；遊戲的
正常 `audio=modern` 路徑仍以純 Go procedural fallback 為基準。
