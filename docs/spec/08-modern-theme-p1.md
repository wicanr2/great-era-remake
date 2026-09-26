# 現代戰場主題 P1：地形／鐵路可切換

> **狀態：READY**
> 日期：2026-08-10
> 範圍：`internal/ui/theme`、戰場 renderer、`cmd/dsds`／`cmd/screenshot`。
> 本規格只改呈現層，不改規則、亂數、存檔或輸入動作。

## 1. 目的與差異邊界

P0 要先讓復古圖形與主題介面共存；P1 再提供一套可散布、可重現的現代戰場
地形／鐵路圖塊。它的用途是降低玩家辨識地形與路網的門檻，不是替原版素材做
重新打包，也不是宣稱完整 modern UI 已完成。

本切片的主題差異明確限定為：

- `retro` 使用玩家自備的 `NEWTERR.TPC`、`RAIL.TPC` 與原版調色盤。
- `modern` 以純 Go deterministic provider 生成 22 張 32×24 地形圖與 21 張
  32×24 鐵路圖；程式碼與常數是新作，不讀取或嵌入原版像素。
- 兩套主題共用同一個地物索引、鐵路索引、六角座標、邊界框與透明像素契約。
  因此切換只會改顏色／紋理／筆畫，不會改移動成本、可通行性、路網連接或
  存檔 bytes。
- P1 尚不涵蓋部隊圖示、資源圖示、向量字型、寬版面或完整設定選單；這些仍
  是 `docs/design/10-visual-modernization.md` 的後續 DRAFT。

## 2. 已證實的輸入契約

原版資料與既有 READY 規格證實：

| 項目 | 契約 |
|---|---|
| `NWMAP.DAT` | 地物編號 1..22，減 1 對應 `NEWTERR.TPC` 0..21 |
| `NEWTERR.TPC` | 22 張 32×24 BGI 圖塊；地形索引與規則層 `TileKind` 一致 |
| `RAIL.TPC` | 21 張 32×24 chunky 圖塊；圖塊索引 0 是有效縱線，不是空圖 |
| 鐵路透明 | 圖塊內像素索引 0 透明；非零像素疊在地形上 |
| 戰場幾何 | 14×14 六角格，奇數欄向下 12 px；renderer 不得由主題改寫 |

證據入口是 `docs/spec/04-battlefield-tiles.md`、`docs/formats/05-tpc-tilesets.md`
與 `internal/assets/nwmap.go`／`rail.go`。本輪對照輸入雜湊如下（只作證據索引，
原版檔案不進版控）：

```text
NEWTERR.TPC  5b4bdec9211667776e8ce7c232e8afe39cc6f7e88cc16748bb5b775fa1926205
RAIL.TPC     bd2d24f89c8f9895824ae5354376731efef069975c9b8d0b92edb5d0e1930232
NWMAP.DAT    452863bd7f0cf2e52f25054bd9f7ab7913895a24b98f790c9570a2b164059275
WARPOS.DAT   1e4058838676a2e1390f2595992111155d5528769e50c16a850d7005cbbe1846
```

## 3. 實作契約

`internal/ui/theme.Theme` 是不依賴 Ebiten／renderer 的窄介面。它只提供 `Tile` 與
`Rail`，每張 `Bitmap` 同時帶自己的 16 色調色盤，避免主題切換時誤套另一套顏色。

`render.DrawThemedBattlefield` 負責唯一一份六角幾何與疊圖順序：

1. 依 `TileKind.TileIndex()` 取地形並完整覆蓋 32×24 格。
2. 若 `Tile.HasRail()`，依原始 0..20 鐵路索引疊圖；像素 0 跳過。
3. 最後繪製邊界省份框線。

`retroThemeAdapter` 保留既有 `DrawTiledBattlefield` API，因此舊測試與外部工具的
復古路徑不需搬移。`cmd/dsds` 的 `-theme retro|modern` 優先於偏好檔；遊戲中
`F2` 只在兩個已載入的 provider 之間切換，寫入偏好但不寫入遊戲存檔。

現代鐵路依原版 21 張圖的邊緣接線遮罩生成：0 縱線、1/2 橫線、3/4 左下彎、
5/6 右下彎、7/8 左上彎、9/10 右上彎、11–18 四種 T 字、19/20 十字。成對索引
保留為不同索引，即使 P1 的新筆畫暫時相同；不能把索引合併成規則層的新編號。

## 4. 驗收

以下是 P1 必須維持的可重現護欄：

- `go test ./internal/ui/theme ./internal/ui/render`：22 個地形與 21 個鐵路索引
  全部存在、尺寸正確、越界拒絕，且鐵路 0 保留縱線接法。
- `go test ./cmd/dsds ./cmd/screenshot`：命令列主題解析與應用程式接線可編譯。
- `cmd/screenshot -theme retro` 與未指定主題輸出同一復古路徑；`-theme modern`
  能用同一份 39 省地圖資料輸出全部省份，不含原版檔案到輸出目錄。
- 固定省份（目前湖北 26）截圖必須看得到長江、城市／山地差異、鐵路路網與
  邊界框；現代截圖要標明是 P1 程式化新作，不得冒充原版畫面。
- `tools/deny_scan.sh --all` 維持原版資產零命中；PNG 白名單仍只允許
  `docs/images/`，P1 不新增原版衍生 PNG。

若任一 provider 缺圖、尺寸錯誤或調色盤不足，renderer 必須報錯（fail-closed），
不能回退成另一張索引圖而掩蓋資料不完整。

