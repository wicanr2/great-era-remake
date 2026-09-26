# 規格 07：外交玩家介面 M0

狀態：**READY（M0 窄切片）**  
日期：2026-08-10

本規格只覆蓋已由 DOSBox 正常玩家操作看見、且規則層已有證據的「外交 → 向外國貸款」路徑。
它不是指令 9 三個子項的完整宣告；外援與償還外債仍維持 fail-closed 提示。

> 後續的 `docs/spec/12-diplomacy-ui-m1.md` 已接上外援／償還外債的 remake 玩家路徑；
> 本檔保留 M0 的原版觀察與貸款契約，不再單獨代表目前整個 UI 狀態。

## 1. 原版觀察契約

以 `fd2-dosbox-screenshot-local`、DOSBox 0.74-3、Xvfb `:99`、US keymap 重播：

```text
2 → Return → 1 → Return → 等待進入政略畫面
9 → Return → 1 → Return
```

看到的外交選單依序是：

1. 向外國貸款
2. 請求外援
3. 償還外債

選第 1 項後看到：

```text
司令，欲貸款多少？（0-5000）
報告司令，目前信用度 100
```

輸入 `1000 → Return` 後回到政略畫面；同一固定存檔路徑中，指令數由 4 降為 3。
這只證明一筆核准的小額貸款會完成玩家指令，不推論拒絕時的指令成本或信用度不足 gate。

證據畫面只留在工作區暫存，不進版控：

| 畫面 | SHA-256 |
|---|---|
| `dip2-correct-menu.png` | `c955d9af9e47a952ef709f5b9b79118e3970648d9eeea4124f62219a31479b11` |
| `dip2-loan-prompt.png` | `12879d2622da25e740495600caa28841bd37d671886f54e26c62fb9c311f9e83` |
| `dip2-loan-submit.png` | `ecba1f40b0bd10e365d0439f91b814e47bbf235d09ec7faa430d02b0c38e9401` |

## 2. remake 接線

- `cmd/dsds` 的指令 9 進入 `screenDiplomacy`；三列使用同一份 wording catalog。
- 「向外國貸款」進入 `screenLoanAmount`，接受數字鍵、滑鼠與 Android 觸控數字鍵盤，輸入上限
  為原版畫面明示的 5000；0 會停留並提示，不執行規則。
- 貸款由 `game.DiplomacyLedger.RequestLoan` 執行；信用度零 gate 與隨機拒絕的指令
  成本不在本 M0 頁面規格內，現由 SPEC-15／SPEC-16 的原版資料流與 UI 接線管理。
  本段保留 M0 的貸款入口描述，不得再把「拒絕不扣指令」當成目前規則。
- `.DT1` 帳本仍只由既有 `WriteDiplomacyLedger` 寫回已證實的區塊 8／9；未知 bytes 不動。
- 原典／現代白話只替換顯示文字，不替換規則。現有鍵為：
  `diplomacy.loan`、`diplomacy.aid`、`diplomacy.repay`、
  `diplomacy.loan.prompt`、`diplomacy.loan.credit`。
- 命中區由 `pointerTargets` 與共用 `numericKeypadTargets` 公告；`actions` 是鍵盤、滑鼠、觸控
  的唯一入口。

## 3. 明確未完成

- `請求外援`：正常玩家入口已在 SPEC-12 接通；援助結果的指令成本由 SPEC-17
  管理，第一期援助國對應仍未閉合。
- `償還外債`：規則層已有窄接入口，玩家額度輸入與原版畫面流程尚未閉合。
- 三平台實機輸入與 Android 封裝；信用度零 gate、貸款拒絕成本另見 SPEC-15／SPEC-16。

完成 M0 不得寫成「外交完成」或「原版全流程等價」。
