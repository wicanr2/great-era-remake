# PLAYTEST-37 指標／觸控 M1：產能分配

日期：2026-08-03  
狀態：**PASS**

## 操作契約

- 尚未選欄位時，鐵礦、煤礦、石油、糧食四列直接派送 `select.1`..`select.4`；
  黃金是四項扣除後的剩餘值，不能直接選。
- 選定欄位後，以框線標示目前項目，停用列命中並顯示共用 6×2 數字鍵盤；
  數字、刪除、送出沿用 M1 的裝置無關動作。
- 返回第一下取消目前欄位與輸入值；未選欄位時再返回才結束產能頁。一次政策工作
  階段無論調幾項仍只扣一次指令，規則沒有因觸控外殼而改變。
- 鍵盤下移至 `y=250`，不遮糧食列；選定後的完整問句與輸入值放在獨立第二行。
  GUI 初版兩次因問句與數字黏連而未通過人工檢查，最終版才保存為證據。

## Docker GUI 實跑

建置使用 `dsds-go:1.25`，實跑使用一次性
`fd2-dosbox-screenshot-local:latest` 的 Xvfb／xdotool，全程不用實體鍵盤：

1. 地圖點「指令」→政策→產能分配。
2. 直接點鐵礦列；鐵礦列出現框線，底部顯示數字鍵盤。
3. 點 `2`、`0`，上方獨立輸入列顯示 20；點勾號送出。
4. 結果回報「產能已調整為 20%（剩 1）」；點返回後由其他選項另存。

合格證據位於 `workplace/pointer-m1-production/`：

- `01-production-list.png`
- `08-iron-selected-final.png`
- `09-production-20-final.png`
- `04-production-result.png`
- `05-production-saved.png`

## 存檔等價

純滑鼠輸出 `production-iron-20.DT1` 的 SHA-256：

```text
d34c17c57387d60ec0bfe9f4bff5acdc24109ca6eeea700d340fb71c4ab8b996
```

與 `docs/playtest/24-policy-production.md` 的實體鍵盤基準完全相同；該基準已證明
只修改湖北省記錄 `+34` 的鐵礦產能 byte（25→20），其餘 14,682 bytes 不變。

## 未完成

- Android 實機密度、安全區、背景恢復及系統返回鍵。
- 密集清單 48dp 的實機驗證或 modern UI 放大列表。
- 六角格與戰鬥觸控。
