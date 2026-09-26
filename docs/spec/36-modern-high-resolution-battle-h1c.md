# SPEC-36：Modern 高解析戰鬥 HUD H1-c

> 狀態：**READY（remake renderer／輸入 slice）**　日期：2026-08-11

## 1. 範圍

本切片把既有 `BattleSim`／`battleState` 接到 1280×720 Modern HUD：左側 3/2 戰場與
部隊／游標，右側資源卡、五項戰鬥命令、三個大控制鍵與撤退數字鍵盤。它不重解原版
未閉合的第 6 鍵、逐樣本動畫、AI oracle 或 `.DT2` 未解欄位。

H3 視覺收斂後，右側面板把攻守摘要、三組資源、戰況訊息、五項命令及三個控制鍵拆成
獨立帶狀區；Modern 部隊以乾淨重繪的陣營 token 呈現，但保留原有 18 個 unit slots、
朝向與輸入語意。這避免舊高解析預覽裡資料列、log 與按鈕彼此碰撞，並不改任何戰鬥規則。

## 2. 垂直鏈

```text
BattleSim／battleState
  → ModernBattleSurfaceData（只讀快照）
  → Theme／UnitProvider 圖示與 3/2 六角幾何
  → ModernBattleLayout panel／command／control／retreat Rect
  → 既有 BattleCommand、BattleAttackTarget、BattleMove、Digit、Submit Action
  → 原有 updateBattle 規則與 DT2／MEM_WAR 寫回入口
```

`internal/ui/layout.ModernBattleLayout` 與 `cmd/dsds/highModernBattlefieldTargets`
共用同一組地圖／命令／控制／鍵盤矩形；高解析頁不再把 Surface 座標反算到 640×350。

## 3. 明確差異

- `ModernBattleSurfaceData` 只呈現已存在的 combatant、資源與 log；不在 renderer 內
  執行 AI、傷害、撤退候選推導或存檔寫回。
- 命令與控制按鈕是現代化外殼，輸入仍沿用既有鍵盤／滑鼠／觸控 Action；高解析畫面
  可玩不等於原版六種攻擊、部署、結算與 AI parity。
- 重新繪製的 token／地形紋理是 clean-room Modern 資產；不嵌入、放大或散布原版
  `.TPC`、`.RGB` 或字型資料。
- Modern 戰鬥 HUD 優先使用可再散布 `GEMF` atlas；缺少／損壞時才由玩家自備倚天字庫
  fallback，圖示來自純 Go Modern provider，不嵌入原版資料。來源／hash 見
  `docs/licenses/modern-font-atlas.md`。

## 4. 驗收

- layout 測試：地圖／面板不重疊，命令與控制達 48 px，六角位移維持 48／18；
- render 測試：1280×720 可畫出戰場／面板，缺 UnitProvider 時 fail-closed；
- cmd/dsds：高解析戰鬥命中中央轉回既有 Action，既有全 repo 回歸通過；
- `cmd/screenshot -theme modern -resolution high -battle -province 26`：以同一份
  `BattleSim` deployment snapshot 產生 `docs/images/high-modern-battle-26.png`
  （1280×720）。展示工具只讀 `BattleSim`／語系／資源資料，不回寫規則或存檔；本圖是
  remake renderer 證據，不是原版逐像素 oracle。
- 同一個展示工具以 `-theme modern -resolution high -biography -province 26` 由
  `PeopleDB`／`DrawModernBiographySurface` 產生
  `docs/images/high-modern-biography-058-01.png`（1280×720），人物自傳資料與翻頁
  狀態不另造副本。
- 原版 oracle、Android 實機 48 dp／背景恢復、音訊裝置與三平台封裝仍屬獨立 gate。
