# PLAYTEST-35：外交選單／貸款 M0

日期：2026-08-10  
狀態：**PASS（原版觀察 + remake 目標接線；不代表外交全流程完成）**

## 原版重播

工具：`fd2-dosbox-screenshot-local`、DOSBox 0.74-3、Xvfb `:99`、US keymap。

固定路徑：

```text
2 → Return → 1 → Return → 等待政略畫面
9 → Return → 1 → Return → 1000 → Return
```

中間畫面依序確認外交三列、貸款額度 0–5000 與信用度 100；送出 1000 後回到政略畫面，
固定樣本的指令數由 4 降到 3。未以這一輪畫面推論信用度 gate、拒絕成本或外債文字版面。

暫存證據（不進版控）雜湊：

- 選單：`c955d9af9e47a952ef709f5b9b79118e3970648d9eeea4124f62219a31479b11`
- 貸款輸入：`12879d2622da25e740495600caa28841bd37d671886f54e26c62fb9c311f9e83`
- 貸款送出：`ecba1f40b0bd10e365d0439f91b814e47bbf235d09ec7faa430d02b0c38e9401`

## Remake 驗證

`cmd/dsds` 已將指令 9、貸款額輸入與 `DiplomacyLedger` 串起來；pointer layer 會在外交選單
公告三個 `select.*` 命中區，在貸款額頁公告共用 12 鍵數字鍵盤。鍵盤與滑鼠／觸控共用
`actions.Select1..3`、`actions.Digit*`、`actions.Submit`，不另造裝置分支。

Docker/Xvfb 目標測試（非完整測試）：

```text
go test ./cmd/dsds ./internal/ui/render ./internal/ui/actions ./internal/ui/layout ./internal/i18n
```

結果：全部通過。外援與償債目前只回報尚待證據，不會偷偷修改遊戲狀態。
