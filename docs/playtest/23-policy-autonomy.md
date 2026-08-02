# 指令 8「政策／授權自治」玩家流程驗證

> 日期：2026-08-02  
> 狀態：**confirmed（自治子流程）**；產能分配後續已於 `24-policy-production.md` 完成。

## 1. IDA 證據與勘誤

- 輸入：`WAR.EXE`
- SHA-256：`11dbfcf24686ab7765f788b38514cefd2039d0f60b6bd517d89fb5a84c068015`
- 工具：IDA Pro 9.4，image `ida-pro-9.4-ver3`
- 位址空間：IDA linear address
- 函式：`sub_23FF6` (`23FF6h..241C5h`)、`sub_22E25` (`22E25h..2347Dh`)
- 非破壞性匯出：`tools/ida_function_export.idc`

`sub_23FF6` 在 `24055h` 與 `24067h` 畫出 `1.`、`2.`，並在
`2419Eh..241B2h` 分派到：

1. `sub_22E25` 「授權自治」
2. `sub_23511` 「產能分配」

因此自治屬於指令 8「政策」，**不屬於指令 15「其他選項」**。

`sub_22E25` 的候選條件：

- `22E52h..22E69h`：當前省司令本人必須就在當前省。
- `22EF5h..22F36h`：該司令至少控制兩省。
- `2307Dh..230EAh`：目標省司令相同，排除領袖所在省，
  並要求 `ds:BCA5h + 省編號 == 0`。
- `23393h..23416h`：省份 `+32` 的 `0x20` 是 toggle，其他位元保留。

`ds:BCA5h` 是 `.DT1` 的 39-byte 停火狀態表，存檔區塊從省 1 開始。
remake 現在會載入這張表，不再假設全部為零。

## 2. 規則與介面

- `AIWorld.AutonomyTargets` 實作上述四個候選條件。
- `TogglePlayerAutonomy` 拒絕不在候選中的省份。
- 一次政策指令內可切換多省；第一次真正切換時只扣一個指令。
- `CommandBudget.Spend` 現在套用 `sub_10193`：自治省執行一項後剩餘數直接歸零。
- 原典／現代白話共用同一候選與 toggle，只更換顯示文字。

產能分配當時尚未開放；後續 IDA 座標、讀取與寫回證據已完成對應，
並通過真實存檔驗證，見 `24-policy-production.md`。

## 3. 真實存檔與 GUI 端到端

輸入 `SAVE(1).DT1`，起始省湖北（26）。規則層先機械驗出唯一合法候選是
河南（19）。兩次 Docker Xvfb 實跑使用相同時間線：

```text
Enter → 8 政策 → 1 授權自治 → 19 Enter → ESC → F10 → Y
```

一次 `wording=plain`，一次 `wording=original`。兩份輸出存檔逐位元組相同：

```text
SHA-256 bfa2fbd2bfeaf2d564c4a01adda98cca80f9c3727a400bc7c0612c049f426296
```

與輸入相比只有一個 byte 不同：

```text
cmp -l: 703   4  44      # 八進位：0x04 → 0x24
```

第 703 byte 是河南省記錄 `+32`。差分只加上 `0x20`，原有「本回合已處理」
`0x04` 保留。畫面上湖北指令數 `2 → 1`，stderr 也回報「河南省：自治（剩 1）」。

## 4. 驗證產物

- `workplace/screenshots/autonomy-ui/01-policy-plain.png`
- `02-autonomy-normal.png`、`03-henan-autonomous.png`
- `04-policy-original.png`、`05-autonomy-original-normal.png`、`06-autonomy-original-on.png`
- `workplace/autonomy-e2e/autonomy-on.DT1`
- `workplace/autonomy-e2e/autonomy-original.DT1`

產物皆為 ignored，UID/GID `1000:1000`。容器使用 `--rm`、`--network none`、
記憶體／CPU／PID 上限與明確使用者；Xvfb 有 trap，未留下背景程序。
