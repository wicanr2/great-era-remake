# SPEC-20：`.MUS`／`.TIM` 純 Go 解碼層 P0

狀態：**READY**  
日期：2026-08-10

## 1. 目的

把已由 `docs/formats/06-mus-tim-audio.md` 證實的 AdLib MUS 事件串與 TIM 音色庫
移植到 `internal/audio/mus`。本切片只負責**無頭、無裝置、可重現的解碼**；不在此
猜測 `SDFA.EXE` 的播放時序，也不宣稱 OPL2 合成或遊戲內播放已完成。

## 2. 輸入契約

### 2.1 MUS

- 檔頭固定 0x46 bytes；little-endian 欄位依格式文件的 `0x24..0x3D` 定義。
- 事件支援 running status、`8n..En` 通道事件、`F0...F7` SysEx、`F8` 即時填充與
  `FC` 結束事件。
- `F8` 沒有 delta、資料或 tick 增量；其餘事件先讀一個 u8 delta。
- `An` 是一個資料 byte 的自訂音量事件，不套用標準 MIDI 的 two-byte aftertouch。
- parser 必須 fail-closed：截斷事件、無 running status、未知 status、未終止 SysEx、
  header 的 `dataSize`／`nrCommand` 不一致都回錯。`totalTick` 要同時計算並暴露；
  原版 `MAINTHEM`／`STRATEGY` 已知與事件累計值不一致，這兩筆 metadata anomaly 不得
  被靜默修正，也不能因此丟掉整首可解析的事件串。

### 2.2 TIM

- 檔頭為 `u8 major`、`u8 minor`、`u16 count`、`u16 parameterOffset`。
- 名稱區是 `count` 筆、每筆 9 bytes 的 ASCII NUL 結尾欄位。
- 每筆音色是 28 個 little-endian u16（56 bytes）：modulator 13、carrier 13、
  `modWave`、`carWave`。
- parser 驗證 offset、總長度與資料邊界；carrier 的 feedback／connection 即使是
  原版已知殘值也必須保留，不可自行清零。

## 3. API 與分層

- `ParseSong([]byte) (Song, error)` 與 `ParseTimbreBank([]byte) (TimbreBank, error)`。
- `internal/audio/mus` 不依賴 Ebiten、`internal/ui` 或原版檔案路徑；只接受 bytes。
- 事件資料與名稱回傳副本，呼叫端修改不得改變 parser 內部狀態。
- 不寫入原始素材，不產生 MIDI／PCM，也不把 `.MUS`／`.TIM` 放入發行資產。

## 4. 驗收

1. 8 首原版 `.MUS` 的 `dataSize`、事件數、header `totalTick`、實際累計 tick、sound
   mode、tempo 與 parser 事件終點均符合 golden table；兩筆 tick mismatch 要明示保留。
2. 8 個 `.TIM` 的 offset、count、大小與 28-word 音色資料均可解析；跨檔共用音色的
   原始欄位不被正規化或覆寫。
3. 不帶原版素材的 malformed-input 測試仍必須執行，覆蓋所有 fail-closed 分支。
4. Docker 內 `go test ./internal/audio/mus`、no-cgo、`git diff --check` 通過。

## 5. 非本切片範圍

- OPL2 register stream、`SDFA.EXE` 播放時序與音量映射。
- PCM 合成、Ebiten audio player、曲目情境切換、循環點與 `-audio=off`。
- 原版／remake 音訊逐樣本比對；這些項目留給後續 READY P1／P2 規格。
