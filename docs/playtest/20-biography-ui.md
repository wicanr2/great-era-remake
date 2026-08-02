# 重製版人物自傳端到端驗證

日期：2026-08-02  
狀態：**confirmed（第一階段玩家 UI）**

## 操作路徑

使用第一期 `SAVE(1)`，從湖北（26）執行：

```text
Enter → 6 查閱 → 3 查閱將領 → 上下選人 → B 開啟自傳
左右：上一人／下一人
Space／PgDn：下一頁；PgUp：上一頁
ESC／B：回到進入前的將領清單或詳細頁
```

`B` 是 remake 新增的唯讀入口；Enter 原有的「查看將領數值」不變，F10 仍是唯一離開鍵。

## 實際結果

- **完整資料**：湖北清單第 1 位吳佩孚，七個核心欄位齊全，434 字正文分成 2 頁；
  Space 可由 `1/2` 到 `2/2`。
- **部分資料**：第 3 位楊源濬，三個核心欄位，頁尾顯示「可靠度：部分資料」，
  不直接暴露資料層的 `partial` 代碼，正文仍正常排版。
- **查無資料**：第 6 位陳德麟，頁面明示「查無可靠傳記記載」，來源 0 筆、可靠度未知；
  沒有空白框，也沒有編造代替內容。
- 數字與拉丁字母由 `ASCFONT.15` 的 8×15 字模繪製，半格 advance 10 px。
- 初次實跑發現「偽」被 WHATWG Big5 encoder 選成 `FA66`；倚天實際碼為 `B0B0`。
  載入器改為從倚天合法碼位反建 Unicode 索引後，「偽」正常顯示。
- 全部玩家可見人物欄位掃描只剩正規化帳本已知的 `榘／藁`；兩字會畫可見缺字框並
  寫入 stderr，不會安靜留白。

## 環境與產物

- 建置：`dsds-go:1.25`，受限 Docker 容器。
- 操作：既有 `fd2-dosbox-screenshot-local` 的 Xvfb／xdotool／ImageMagick 工具鏈；
  `--network none`、有限記憶體／CPU／程序數，容器使用目前使用者 UID/GID。
- 驗證畫面：`workplace/screenshots/biography-ui/01-confirmed-page1.png`、
  `02-confirmed-page2.png`、`03-partial.png`、`04-unknown.png`（不進版控）。
- 清單中繼畫面保留游標證據：`03-partial-list.png`、`04-unknown-list.png`。
- 一次性容器均使用 `--rm`；完成後須確認無專案相關容器殘留。

## 性質

人物自傳是原版沒有的 remake 新功能。它只讀人物語系資料，不寫回規則層、回合狀態或
存檔；版面座標、B 鍵入口與半形 10 px 字距均登記於 `docs/design/20-biography-system.md`。
