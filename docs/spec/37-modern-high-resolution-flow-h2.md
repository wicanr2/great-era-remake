# SPEC-37：Modern 高解析政略流程頁（H2）

狀態：**READY／已實作**　日期：2026-08-11

## 目的

`screenDevelop`、調動／商業／運補、徵兵／整編、秘密行動、外交／停火、查閱（含將領
詳情與人物自傳入口）、政策、自治／產能、數字輸入與確認頁不能在高解析模式退回 640×350。它們共用
`ModernFlowLayout` 與 `DrawModernFlowSurface`，維持原有規則、`Action` 與資料來源。

## 契約

- Modern 高解析固定使用 1280×720、24 px safe inset；選項卡至少 48×48 設計像素。
- 清單數量依欄數自適應（2／3／5／6／8／11 欄），但只改排版，不改
  `Selection(n)` 的順序與數值。
- 數字輸入固定 3×4 鍵盤：1–9、刪除、0、確認；映射既有 `Digit`／`DeleteDigit`／
  `Submit`，返回鍵映射 `Back`。
- 確認頁使用兩張卡，左為 `Confirm`、右為 `Cancel`；畫面不直接執行規則。
- 將領詳情的數值卡是只讀資料，不產生假的 `Selection`；若既有人物資料可接合，頁尾
  自傳按鈕送 `OpenBiography`，進入 H1-b 的分頁 renderer。
- 所有標題、提示、選項與導覽由 wording catalog 提供；省名／將領名沿用既有 locale／
  biography provider。缺 wording、錯誤畫布或錯誤 style 時 fail-closed。
- pointer／touch 與 keyboard 進入同一個 `Action`；黑邊與非設計座標仍由既有
  Surface adapter 拒絕。

## 完成證據

- `internal/ui/layout/modern.go`：`ModernFlowLayout` 與 48 px／列表／鍵盤幾何測試。
- `internal/ui/render/modern_pages.go`：`DrawModernFlowSurface` 純 image renderer；
  `internal/ui/render/modern_high_test.go` 覆蓋選項、數字輸入與錯誤畫布。
- `cmd/dsds/modern_flow.go`：把既有 screen state 轉成純資料模型；不新增規則 helper。
- `cmd/dsds/main.go`／`pointer.go`：Modern + high 的非專用頁直接使用高解析 surface，
  pointer／touch 共用 flow layout。
- `cmd/dsds/modern_flow_test.go`：選項卡、鍵盤與設計座標命中測試。

## 邊界

這個規格完成 remake 呈現／輸入接線，不證明原版 oracle、AI、撤退部署或未解存檔欄位
等價；640×350 retro 路徑與遊戲規則沒有改動。高解析專用頁（地圖、政略主選單、
人物自傳、敘事、戰鬥）仍由 H1 renderer 負責，流程頁只共用 Modern 外殼與語意資料。
