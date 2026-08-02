# 原典／現代白話：發展與練兵

日期：2026-08-02  
狀態：**confirmed（指令 7、13 的呈現與規則等價）**

## 修正內容

`screenDevelop` 原本錯畫完整 15 項政略主選單，雖然輸入狀態機只接受 1–3。
現在原典路徑顯示「墾地／建兵工廠／挖金礦」，白話路徑顯示
「開墾土地／建造兵工廠／開採金礦」。練兵確認也新增白話提示
「確定要訓練本省部隊嗎？」及可見的 Y／N 操作。

原典詞條以 IDA `sub_24D59`、`docs/re/27-screen-map.md` 與
`docs/formats/glyph-tables/_slots.json` 交叉確認。過程曾把 IDA 的另一個三字緩衝區
立即數 #10 誤套到 `3.15`，GUI 因而顯示「熱河省」；人工截圖抓到後改用槽表確認的
`3.15` 0-based #52「挖金礦」。此勘誤保留，避免日後再次把不同字模的同索引混用。

## GUI 路徑

兩種用語模式均使用第一期 `SAVE(1).DT1`、湖北、固定種子 1：

```text
Enter → 7 → 截取發展選單 → 1 墾地／開墾土地 → F10 → Y
Enter → R → 截取練兵確認 → Esc → F10 → Y
```

四張畫面均經人工檢查，沒有缺字、裁切或殘留上一幀：

- `workplace/screenshots/develop-wording/original-develop.png`
- `workplace/screenshots/develop-wording/plain-develop.png`
- `workplace/screenshots/develop-wording/original-train.png`
- `workplace/screenshots/develop-wording/plain-train.png`

## 狀態與存檔

- 兩次發展均回報：地價 31（+9）、體力 45、剩餘指令 1。
- 原典與白話的發展輸出逐位元組相同，SHA-256 皆為
  `0fd1f2d00eb2bf963cd3f88827574c75dd99594c402e839bb0bb54e00b3a3efa`。
- 原始輸入 SHA-256 為
  `9ad9359b436ce2450ff3460df4efede478d871dfe234f60299759e781959ad0d`，
  證明不是未執行指令的假等價。
- 練兵截圖路徑按 Esc 取消，兩份輸出皆與原始輸入相同，證明取消沒有改狀態。

## 環境

`dsds-go:1.25` 建置；`fd2-dosbox-screenshot-local:latest` 提供 Xvfb、xdotool 與
ImageMagick。每次執行均為無網路、有限資源、目前 UID/GID 的一次性容器。
