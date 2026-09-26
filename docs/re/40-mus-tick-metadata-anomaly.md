# RE-40：`.MUS` header `totalTick` 與事件累計值的兩筆差異

日期：2026-08-10  
工具：`tools/mus.py`（既有解碼器）與 `internal/audio/mus`（純 Go P0）  
位址空間：不適用（資料檔 parser）

## 輸入與可重現結果

輸入來自唯讀 `workplace/orig/game/`：

| 檔案 | SHA-256 | header `totalTick` | 事件累計 tick | 差異 |
|---|---|---:|---:|---:|
| `MAINTHEM.MUS` | `bc6b8173cd20e91aa79588c12265386c0267f48cc63b262b366128548baf0346` | 62880 | 60800 | -2080 |
| `STRATEGY.MUS` | `8e41204ff23c4703cb51505820d38791539d35dcb90e4881bddb8cdc8be192ca` | 15540 | 15600 | +60 |

其餘六首的 header／事件累計值一致。八首的事件數仍全部符合 header `nrCommand`，
事件串均在 `FC` 後於檔尾結束；因此差異只落在 metadata 的 tick 總長，不是截斷或
status 解碼失敗。

既有 Python parser 對 `F8` 採「無 delta、無 tick」處理；本次用相同規則重跑後，兩首
仍呈現上述差異。這足以推翻「八首 `totalTick` 必然與事件累計值一致」的舊敘述，
但尚不足以推論 `F8` 的播放時序或 `SDFA.EXE` 的隱藏處理。

## 實作決定

`internal/audio/mus.ParseSong`：

- 保留 header `TotalTick` 原值；
- 新增 `ActualTick` 與 `HeaderTickMatches`，讓呼叫端看得到差異；
- 對 `dataSize`、`nrCommand`、事件結束與 malformed status 仍 fail-closed；
- 不在解碼層補 tick、刪事件或猜測 F8 的播放意義。

這是資料保存與後續 OPL2 排程的邊界，不是音樂播放已完成的證據。`SDFA.EXE` 的時序、
循環與 tick 修正仍列為 unknown，待後續 P1 以原版暫存器／正常播放 oracle 裁決。
