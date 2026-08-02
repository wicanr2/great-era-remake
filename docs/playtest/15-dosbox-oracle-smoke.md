# 15 — DOSBox oracle smoke 固化

日期：2026-08-02  
結論等級：**confirmed（可重現實機證據）**

## 結論

`docs/playtest/03-input-regression.md` 記錄的按鍵回歸目前不存在。以現有
`fd2-dosbox-screenshot-local`、DOSBox 0.74-3、Xvfb `:99` 重跑固定序列：

```text
WAR.EXE → wait → Return → 2 → Return
```

畫面依序到達：

1. 主選單「重新開始／載入遊戲」；
2. `Return` 不換頁（遊戲本來就在等選項）；
3. 按 `2` 後進入載入轉場；
4. 按 `Return` 後顯示「進度一」到「進度十」的載入選單。

三個狀態的 SHA-256：

| 狀態 | SHA-256 |
|---|---|
| 主選單 | `ead5b9243736cd7d455d8f4810ea810b27c53eeb2b00b9b45f5cae139554875b` |
| 載入轉場 | `7d459121c1731da65d550d2a0252c66c7bc609b9efb2789aa1e47d2b7594d638` |
| 進度選單（游標滅） | `582963e04bc3f0686717775b47c07d1da2c7088e88a63ad56e727537c2096629` |
| 進度選單（游標亮） | `5ee11b2371482026358ae9cb64fbf7167a249152d20f263c0d32a2d892304b3d` |

截圖在 `workplace/dosbox/shots/`，屬原版美術衍生證據，不進版控。本輪已逐張人工檢視，
並確認最後一張確實是十個進度的載入選單，不只用「雜湊有變」代替畫面判讀。進度選單
有閃爍的底線游標，所以合法 oracle 有亮／滅兩個相位；第三次重跑正好抓到游標亮的版本。

## 工具鏈修正

`tools/dosbox.sh` 現在符合專案 Docker-only 規則：

- 原版 `/orig` 唯讀掛載，遊戲副本在容器內重建；
- `--network none`、`--memory 768m`、`--cpus 1`、`--pids-limit 256`；
- 以目前 UID/GID 執行，啟動前拒絕非目前使用者擁有的輸出目錄；
- root filesystem 唯讀，`/tmp` 使用 64 MiB tmpfs；
- 外層預設 240 秒 timeout，逾時後有界終止；
- 每輪清除舊的頂層 PNG，避免把舊截圖當成本輪結果。

`tools/dosbox_runner.sh` 也刪除了「Xvfb 沒有 keymap」的過時註解。`docs/playtest/06`
已證明那是因為 `xmodmap` 根本不存在而產生的假零值；真正修復仍是 PointerRoot 模式下的
`mousemove --window` + `windowfocus`。

## 固定 smoke gate

```sh
bash tools/dosbox_smoke.sh
```

腳本在受限 Docker 容器內跑上述時間線，再由另一個唯讀容器檢查四張截圖存在、三次狀態
關係符合預期，而且三個狀態逐張命中本頁已人工確認的 SHA-256；載入選單只接受已確認的
游標亮／滅兩種。這是原版畫面的有界 gate；若環境或擷取時機造成其他雜湊，必須重新人工
檢視並在本頁記錄證據，不能只更新常數讓測試轉綠。

## 非阻斷訊息

- `_XSERVTransmkdir: euid != 0`：Xvfb 以非 root 執行時不建立全域 socket 目錄；本輪
  `/tmp/.X11-unix/X99` 仍正常建立，後續截圖與輸入均成功。
- ALSA 找不到音效卡：此 oracle 設定 `nosound=true`，不影響鍵盤與畫面驗證。
