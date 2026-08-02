# 重製版指令 1「調動行動」端到端驗證

日期：2026-08-02  
狀態：**confirmed（規則＋玩家 UI）**

## 操作路徑

使用第一期 `SAVE(1)`，從湖北（26）執行：

```text
Enter → 1 調動行動 → 2 全部調動 → 19 河南 → Enter 確認選取
→ 黃金／糧食／彈藥／燃料均為 0 → Y
```

選將畫面列出湖北 15 位將領，全部帶選取方塊；方向鍵移動游標、Space 切換選取，
Enter 明確提交目前集合。Enter 是刻意的現代操作修正：DOS 原版的「欲何將留守」
控制流無法提交非全選集合，證據見 `docs/playtest/18-player-transfer-selection.md`。

## 實際結果

- 程式回報：`調動 15 位將領至河南省；黃金 0、糧食 0、彈藥 0、燃料 0（剩 1）`。
- 湖北兵力由 97,500 變成 0，將領數由 15 變成 0。
- 湖北本月指令數由 2 變成 1。
- 確認前顯示原版「您確定嗎」字模，N／Escape 可退回物資輸入，不會先改狀態。
- 規則層另有存檔寫回再重載測試，確認將領省份、體力、司令與省長可以持久化。

## 環境與產物

- 建置：`dsds-go:1.25`，受限 Docker 容器。
- 操作：既有 `fd2-dosbox-screenshot-local` 的 Xvfb／xdotool／ImageMagick 工具鏈；
  `--network none`、有限記憶體／CPU／程序數，容器使用目前使用者 UID/GID。
- 畫面：`workplace/screenshots/transfer-ui/01-selection.png`、`02-confirm.png`、
  `03-result.png`（驗證素材，不進版控）。
- 一次性容器使用 `--rm`；完成後無殘留執行中或停止容器。

