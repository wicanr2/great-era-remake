# SPEC-06：外交帳本與外援資源基準

狀態：**READY**（2026-08-09）

本規格定義 `WAR.EXE` 執行期已由 IDA 證實的外交帳本、其 `.DT1` 區塊 8／9
持久化，以及請求外援前的四種資源基準。它讓 Go 規則層可以保留信用度／外債的
狀態並以非破壞性方式讀寫原版存檔；玩家貸款／外援／償還畫面另由
`docs/spec/07-diplomacy-ui.md` 與 `docs/spec/12-diplomacy-ui-m1.md` 管理，停火玩家
呼叫端與數字頁另由 `docs/spec/13-ceasefire-ui-m1.md` 管理。本規格仍不宣稱 Android
外殼或完整外交流程已完成；信用度為零的貸款早期 gate 另由
`docs/spec/15-loan-credit-gate-m1.md` 管理。

## 1. 證據身分

- 輸入：`WAR.EXE`，SHA-256
  `11dbfcf24686ab7765f788b38514cefd2039d0f60b6bd517d89fb5a84c068015`。
- 工具：IDA Pro `9.4.0.260610`；下列位址均先標 **IDA 線性位址**，組語中的
  `[di-xxxxh]` 仍原樣保留；`ds:` 換算基準為 `DSEG_BASE = 0x641B0`。
- 勢力索引：`sub_215F0` 以省份司令反查 `word_70026`，使用 1..10 的勢力槽位。
  第 0 格只在 remake 中作無效哨兵，不能當成原版第 0 勢力。

## 2. 十槽外交帳本

`DiplomacyLedger` 以 1-based 索引保存十個勢力的執行期狀態：

| Go 欄位 | 原始運算元 | index 1 的位置 | 型別／初值 | 證據 |
|---|---|---|---|---|
| `Credit[1..10]` | `[di-4225h]` | `ds:BDDB + 1`；IDA `6FF8Ch` | 10 × `u8`／100 | `sub_391E1` 寫 64h；`sub_2164A` 扣；`sub_223ED` 加並夾 100 |
| `Debt[1..10]` | `[di-421Eh]`／`[di-421Ch]` | `ds:BDE2 + 4`；IDA `6FF96h` | 10 × little-endian `u32`／0 | `sub_391E1` 清零；`sub_2164A` `add/adc`；`sub_223ED` `sub/sbb` |

`internal/game.NewDiplomacyLedger` 重現上述初始化。`sub_595D4` 在
`595D4h..598C0h` 把 `byte_6FF96` 的 28 bytes 與 `byte_6FF8C` 的 10 bytes
寫入 `.DT1` 暫存記錄；`sub_59CBF` 在 `59CBFh..5A031h` 以相同大小讀回。
因此 `.DT1` offset `14620..14659` 是 10×little-endian `u32` 外債，
offset `14610..14619` 是 10×`u8` 信用度。兩段在原版搬運順序中是「外債後寫、
信用度先寫」，但在檔案中信用度位於外債之前；`ParseDiplomacyLedger`／
`WriteDiplomacyLedger` 只改這兩段，檔頭、區塊 10、以及夾在尾端的 7 個 runtime
byte 原樣保留。

## 3. 貸款規則層接合

既有 `AIWorld.RequestLoan` 保留純公式；`DiplomacyLedger.RequestLoan` 只在核准後
追加已證實的帳本副作用：

```text
units = amount ÷ 500
roll  = Random(10)
若 roll + units > 12：拒絕，信用度／外債／省份不變
否則：信用度 -= units；省份黃金 += amount；外債 += amount（u32）
```

信用度為零時的原版額外阻擋分支已由 `docs/re/34-loan-credit-gate.md` 證實並接入
`LoanResult.CreditBlocked`；正信用度不足以支付本次額度單位時，仍不另加未證實門檻，
保留既有核貸公式的行為邊界。

## 4. 償還規則層接合

`DiplomacyLedger.RepayDebt` 先拒絕非正數或超過 `Debt[slot]` 的額度，再呼叫
`AIWorld.RepayDebt` 完成已證實的省份／信用度規則；成功後同步扣除帳本外債：

```text
省份黃金 -= amount
Debt -= amount（u32）
credit += amount ÷ 500，最高 100
```

這個入口不允許超額償還，也不把未解的畫面失敗分支或存檔欄位自行補上。

## 5. 外援基準

`sub_21D1D` 在四次 `Random` 前，對目前省份的資源欄位分別計算：

```text
base[0] = 60000 - Gold
base[1] = 60000 - Food
base[2] = 60000 - Ammo
base[3] = 60000 - Fuel
```

Go 端以 `internal/game.AidResourceBases` 與 `AIWorld.AidResourceBases` 提供同一個
窄入口；已達上限或超過上限的欄位以 0 作為不再請求的基準。外援拒絕／核准分支、
特殊核准覆寫、援助國除數與黃金 6000 上限仍由 `RequestAid` 的既有規格／測試管理，
不在本帳本結構重複實作。

## 6. `.DT1` 持久化與驗證

`ParseDiplomacyLedger` 以 1-based slot 解碼兩段；`WriteDiplomacyLedger` 從原始
bytes 複製後只覆蓋已證實的 50 bytes。真實 `SAVE(1).DT1` 錨點為：

- offset `14610` 的信用度第 1 槽為 `64`（100）；
- offset `14620..14623` 的外債第 1 槽在兩份基準存檔均為 `00 00 00 00`；
- offset `14606..14609` 的 `01 01 01 00`／`05 00 01 01` 是四個未映射 runtime
  byte，不能誤讀成外債；區塊 10 則從 offset `14662` 開始。

`internal/game/diplomacy_ledger_test.go` 另驗證短檔失敗、讀回一致與未授權 offset
逐 byte 不變；`cmd/dsds` 的 `buildSession`／`autosave` 已接上同一份帳本。

## 7. 明確不在本規格內

- 停火（指令 10）的正常 DOSBox 完整等價、文字切換的全流程與完整輸入流程；信用 gate
  的規則／玩家窄接另見 SPEC-15；
  停火 UI 的已接窄切片見 SPEC-13，貸款／外援／償還見 SPEC-07／SPEC-12。
- `sub_2164A` 開頭的司令／分期顯示表；信用度零 gate 與正信用度下的核貸拒絕成本
  分別見 SPEC-15／SPEC-16；外援結果成本見 SPEC-17。
- Android 封裝、觸控命中區與外交畫面排版。

上述項目必須取得各自的 IDA 交叉參照、正常 DOSBox 玩家路徑或存檔 byte diff，
再另立規格，不得以本規格的單元測試代替原版等價證據。

## 8. 驗證入口

- `internal/game/diplomacy_ledger_test.go`：初始化、1-based 邊界、核准／拒絕貸款、
  償還／超額保護與四種外援基準。
- `docs/mechanics/50-diplomacy.md`：逐函式證據與尚未閉合的 runtime／存檔邊界。
- `tools/addr.py -4225h -421Eh`：保留原始負位移並顯示 `confirmed` 語意與出處。
