# 戰鬥世界結算／人物自傳 M2 驗收紀錄

日期：2026-08-10  
環境：`dsds-go:1.25`；一次性 Docker；Ebiten 測試以 Xvfb `:99` 執行

## 已驗證

- `go test -count=1 ./internal/game` 通過：攻方勝接管目標省並移防存活攻方、守方勝／
  回合平局只清 `ProvinceFlagInBattle`、立即撤退不投影、無效勝方失敗即關閉。
- `DISPLAY=:99 go test -count=1 ./cmd/dsds` 通過：戰鬥結算接線、存檔／AI 兵力鏡像、
  相鄰／遠程目標指標、鍵盤／滑鼠／觸控共用 action、自傳入口與分頁。
- 人物槽位驗收：第一至三期 486 槽位中，485 個可接合，#274「無省長」明確排除；
  沒有正文的可接合人物仍顯示一頁 unknown fallback。
- 攻擊模式會在敵軍格上顯示穩定的 `1..6` 目標標號；標號順序與
  `actions.BattleAttackTarget(n)` 相同，不依賴滑鼠座標猜測。

## 尚未以原版 oracle 證實

- 原版六種攻擊方式的完整分支、彈藥／動畫／音效副作用。
- 所有一般交戰組合的撤退候選與原版戰後 `.DT2`／`MEM_WAR.DAT` 同步時機。
- Android 封裝、三平台正式包與完整正常玩家長流程；本紀錄只代表程式與規則層切片。

## 可重跑命令

```sh
tools/go.sh test -count=1 ./internal/game
# 在 Docker/Xvfb :99 內：
DISPLAY=:99 go test -count=1 ./cmd/dsds
```

所有分析、建置與測試均在 Docker 內完成；一次性容器使用 `--rm`，本輪未停止其他專案
容器。原版素材拒絕清單、no-cgo、Windows 交叉建置與全量回歸於本輪收尾一併重跑。
