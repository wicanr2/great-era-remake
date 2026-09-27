# Modern Ogg 執行期接線 M1

狀態：**READY／runtime slice 已實作；音樂內容仍未交付**  
日期：2026-08-11

本規格只處理 `audio=modern` 的安全載入與播放邊界，不決定作曲風格，也不把
MIDI 草稿或原版 `.MUS`／`.TIM` 轉成發行素材。創作 brief、3–5 音新動機、授權與
混音 QA 仍以 [`docs/design/32-modern-music-direction.md`](../design/32-modern-music-direction.md)
為準。

## 1. 已接通的垂直鏈

```text
manifest.json + Ogg bytes
        ↓ schema／檔名／author／license／SHA-256 驗證
ModernTrackSource（loop sample metadata）
        ↓ Ebiten 內建 vorbis.DecodeWithSampleRate（純 Go）
audio.Context → audio.Player → 18 frame bounded crossfade
        ↓ 畫面情境
scene／strategy／battle_a／battle_b／story／final
```

- `internal/ui/audio.ParseMode` 接受 `off`、`retro`、`modern`。
- `internal/ui/audio.LoadModernTracks` 只讀 manifest 明列的同目錄 `.ogg` 檔；不接受
  路徑穿越、未知 cue、缺 `author`／`license` 或錯 SHA-256。
- `modern_scene` 是必要曲目；其他 cue 可缺，載入結果留下 warning，切換時沿用目前
  曲目。壞 codec／壞循環點也只會讓該次切換失敗，不關掉舊 player。
- `internal/ui/audio.NewModernTracks` 以初始 Ogg 的取樣率建立 Ebiten context；其他
  曲目切換時用同一取樣率解碼，避免把 Ogg bytes 誤當 PCM。
- 解碼後的 stereo signed-16 stream 由 loop reader 依 manifest 的每聲道 sample
  `loop_start`／`loop_length` 循環；兩者皆零表示整首循環。
- `cmd/dsds -audio modern -modern-audio <dir>` 找不到 manifest、scene 或有效 Ogg 時
  印出 stderr 診斷並安全改用 SPEC-35 的原創純 Go cue；不阻塞遊戲與存檔。

## 2. Manifest 契約

```json
{
  "schema": 1,
  "version": "modern-music-2026-08-11",
  "tracks": {
    "modern_scene": {
      "file": "modern_scene.ogg",
      "loop_start": 0,
      "loop_length": 0,
      "author": "待填入實際作者",
      "license": "待填入實際授權",
      "sha256": "64 位十六進位雜湊"
    }
  }
}
```

允許的 cue ID 是 `modern_scene`、`modern_strategy`、`modern_battle_a`、
`modern_battle_b`、`modern_story`、`modern_final`；每個 `file` 必須是 manifest 同目錄
的 `.ogg` basename。`author`、`license` 與實際檔案 SHA-256 是必要 provenance，不能
用空字串或「generated」代替。

## 3. 交付狀態（2026-09-27 更新）

- `assets/music/modern/` 已有六首出貨 `.ogg`＋`manifest.json`
 （`tools/make_modern_ogg.sh` 可重現：程序作曲 WAV → libvorbis q4／44.1kHz；
  manifest 具 author／license／SHA-256，版本 `modern-music-tech-preview-*`）。
  曲目為原創程序作曲的技術預覽，**不是**正式作曲署名、盲聽或人耳混音完成。
- 新主題動機、MIDI sketch、分軌／音源 provenance、loop／響度／mono 人耳驗收尚未完成。
- Windows／macOS／Android 真機音訊驗收與發行包仍是 release gate；Linux 測試通過不等於
  實機驗收。
- `SDFA.EXE` register parity 不屬於此接線的前置條件；`audio=retro` 仍維持純 Go OPL2
  路徑，兩者可獨立選擇。

## 4. 最小測試

- `internal/ui/audio`：模式解析、manifest provenance／路徑／雜湊、壞 Ogg fail-closed、
  loop reader 重複範圍、procedural cue／FX deterministic 與既有 retro crossfade。
- `cmd/dsds`：Xvfb 下既有 GUI／Modern UI 測試；`-audio=modern` 缺 manifest 時改用
  原創純 Go loop，不宣稱正式作曲／授權已完成。
