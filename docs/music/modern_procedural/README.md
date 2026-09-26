# Modern procedural 音訊預覽

> 狀態：**技術預覽（非發行資產）**　日期：2026-08-11

`cmd/modern_audio` 已在既有 `dsds-go:1.25` Docker 映像中重生八條 Modern 情境
cue 與八類效果音，再以隔離的 `u5cht/video:latest`（FFmpeg 5.1.9）轉成 Vorbis
Ogg。輸出放在被 `.gitignore` 忽略的
`workplace/promo/modern_audio/procedural/`，方便本機聽審與推廣片剪輯；不會進入
`assets/music/modern/`，也不會被遊戲 manifest 當成已清權音源。

## 可重生命令

```sh
tools/go.sh run ./cmd/modern_audio \
  -out workplace/promo/modern_audio/procedural \
  -sample-rate 48000 -effects \
  -report workplace/promo/modern_audio/procedural/report.json
```

WAV 轉 Ogg 的命令在 `u5cht/video:latest` 內執行，使用 `libvorbis -q:a 5`；每個
輸出再以 `ffprobe` 檢查 codec、取樣率、聲道與時長，並以 FFmpeg 解碼至 null sink。

## 2026-08-11 技術結果

所有檔案均為 Vorbis、48,000 Hz、立體聲，且可完整解碼：

| 類別 | 檔案 | 時長（秒） |
|---|---|---:|
| cue | `scene.ogg` | 62.608688 |
| cue | `strategy.ogg`、`main-theme.ogg`、`battle-1.ogg`、`battle-2.ogg`、`battle-alt.ogg` | 51.428563 |
| cue | `wall.ogg` | 70.243896 |
| cue | `final.ogg` | 55.384625 |
| FX | `fx-confirm.ogg` | 0.130000 |
| FX | `fx-cancel.ogg` | 0.140000 |
| FX | `fx-key.ogg` | 0.055000 |
| FX | `fx-command.ogg` | 0.180000 |
| FX | `fx-battle-move.ogg` | 0.110000 |
| FX | `fx-battle-attack.ogg` | 0.280000 |
| FX | `fx-battle-hit.ogg` | 0.200000 |
| FX | `fx-battle-turn.ogg` | 0.240000 |

WAV 的峰值、RMS、mono 相容性、首尾 sample 與 clipping 結果保存在同目錄的
`report.json`。Ogg 轉檔的 SHA-256 也只保留在本機輸出，不把衍生音檔或雜湊誤當作
正式 provenance。

## 尚未解除的發行閘門

- 自然人編曲／演奏、音色／樣本來源與授權紀錄；
- 連續三圈 loop 的人耳聽審、mono／耳機／喇叭混音檢查；
- Windows、macOS、Linux 與 Android 真機音訊驗收；
- 正式 Ogg 的 manifest、loop sample、版本與 SHA-256。

因此目前仍以純 Go procedural composition 作為 `audio=modern` 的可玩 fallback；本
預覽只證明音檔可以由同一 renderer 重生、轉碼並被 Vorbis 解碼，不宣稱完成正式
音樂發行或原版旋律 parity。
