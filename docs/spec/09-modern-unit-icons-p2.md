# 現代部隊圖示 P2a：原版索引可切換

> **狀態：READY**
> 日期：2026-08-10
> 範圍：`internal/ui/theme`、部隊 icon renderer、`cmd/dsds`／`cmd/screenshot`。
> 本規格只改圖示呈現，不改兵種規則、戰鬥數值、攻守判定或存檔。

## 1. 目的與切片邊界

P1 已讓地形／鐵路在 retro 與 modern 間切換；本切片把同一個主題 asset group
延伸到戰鬥中的 18 張部隊圖示。modern 圖示以純 Go 生成，保留原版圖示的索引、
尺寸、透明像素、攻守色意義與砲兵六朝向；這是可散布的程式化新作，不是
`NEWICON.TPC` 的轉存或重新包裝。

本切片刻意不宣稱：

- 十大勢力十色、symbol icon 軸或新增狀態資訊；這些會增加資訊量，需另立 spec。
- 向量／SVG／PNG 美術流程、寬版 HUD、資源／指令 icon 或完整設定選單。
- 砲兵朝向的規則意義以外的新旋轉數學；provider 只按既有 0-based 圖示索引取圖。

## 2. 已證實的索引契約

`internal/ui/render.BranchIcon` 是唯一的兵種到圖示索引對照：

| index | 原版含義 |
|---:|---|
| 0／1 | 步兵，綠／紅 |
| 2／3 | 裝甲兵，綠／紅 |
| 4／5 | 騎兵，綠／紅 |
| 6..11 | 砲兵，綠，朝向 1..6 |
| 12..17 | 砲兵，紅，朝向 1..6 |

尺寸是 32×17，像素索引 0 透明。兵種值 1／4／5／6 與砲兵朝向 `+31` 的
規則證據見 `docs/formats/05-tpc-tilesets.md`、`internal/game/general.go`、
`internal/game/combat.go`、`docs/re/09`；本切片不在 theme package 重複一份
兵種判斷。

## 3. 實作契約

`theme.UnitProvider.Unit(index int)` 接受上述 0..17 索引並回傳含調色盤的
`theme.Bitmap`。`render.DrawThemedUnitIcon`／`DrawThemedUnitAtCell` 共用透明、
尺寸與六角座標繪製；舊 `DrawUnitIcon`／`DrawUnitAtCell` 只是 retro 相容 wrapper。

retro provider 直接包住玩家自備 `NEWICON.TPC` 的解碼結果與 EGA 調色盤；modern
provider 預載完整 18 張新圖。`cmd/dsds` 的 `F2` 會同步交換地形／鐵路與部隊圖示
provider；任一組缺圖或尺寸不符就 fail-closed，不以另一套圖示偷偷補位。

modern 圖示的綠／紅只表示原版兩方語意，顏色由 modern palette 的 jade／vermilion
延伸；砲兵 1..6 以六向量繪製不同砲管。所有圖示仍是 32×17，故 P2 不改六角格
幾何，也不增加可見資訊量。

## 4. 驗收

- `go test ./internal/ui/theme ./internal/ui/render`：18 張 modern 圖示完整、尺寸／
  palette／越界檢查通過；砲兵相鄰朝向像素不同；retro provider 與原 wrapper
  逐像素相同。
- `go test ./cmd/dsds ./cmd/screenshot` 在 Xvfb 下通過；`cmd/screenshot -units`
  的 retro／modern 都能輸出湖北 26。
- `cmd/screenshot -battle -theme modern` 的部署單位使用 modern provider，
  砲兵朝向仍由 `BranchIcon` 的 6..17 索引決定；規則層與存檔不變。
- `tools/deny_scan.sh --all` 維持原版資產零命中；P2 不新增 PNG／SVG，避免繞過
  `docs/images/` 唯一白名單。

