# SPEC-10：顯示設定中的圖形主題（P2b）

> **狀態：READY**（2026-08-10）
> 
> 本規格只接通既有「其他選項 → 顯示設定」頁的 `retro`／`modern` 主題選擇。
> 它不是完整 modern UI，也不宣稱資源圖示、向量字型、寬版面或第三套 symbol
> 圖示已完成。

## 1. 目的與邊界

P1 的 `F2` 與命令列旗標已能切換戰場地形／鐵路，P2a 再把 18 張部隊圖示納入同一
個主題 asset group；玩家若不知道 `F2`，仍應能從既有可見的設定入口找到同一功能。
P2b 將主題加入顯示設定頁，並把偏好寫入獨立的 `prefs.json`。

本切片允許的變更：

- 用語仍由選項 1／2 選擇「原典用語／現代白話」。
- 圖形主題由選項 3／4 選擇「原版圖形／現代圖形」。兩者可組合，互不改變規則。
- 鍵盤、滑鼠與觸控都轉成同一個 `actions.Select1..Select4` 動作。
- 選取後立即重繪，並以原子方式交換地形、鐵路與部隊圖示 provider。
- `prefs.json` 只保存主題，不改寫遊戲存檔；壞主題值讀取／寫入皆 fail-closed。

刻意不在本切片處理：資源／指令圖示、十勢力顏色、向量字型、寬版面、symbol preset、
戰鬥中獨立設定頁與 Android 封裝。

## 2. 介面契約

顯示設定頁保留 640×350 邏輯畫布與原有用語兩列；下方新增兩個並排主題選項：

| 動作 | 鍵盤 | 顯示文字（原典／白話） | 效果 |
|---|---|---|---|
| `Select1` | `1` | 原典用語 | 只改語系軸 |
| `Select2` | `2` | 現代白話 | 只改語系軸 |
| `Select3` | `3` | 原版圖形／經典原味 | 交換 retro provider |
| `Select4` | `4` | 現代圖形／現代清晰版 | 交換 modern provider |
| `Back` | `ESC` | ESC 返回 | 回到其他選項，不保存遊戲狀態 |

`internal/ui/layout.DisplayWordingOption` 與 `DisplayThemeOption` 同時供 renderer、
pointer／touch 命中使用。命中區不得與返回按鈕重疊；拖曳放開或畫面已變更時不得派送
選項動作。

## 3. 資料與 fail-closed

- `internal/prefs.Preferences.Theme` 只接受 `retro` 或 `modern`；空值仍相容舊設定並由
  內建預設解成 `retro`。
- `setThemePreference` 先驗證目標 `Theme` 與 `UnitProvider`，再寫入暫存設定檔並一次
  更新 `battlefieldTheme`／`unitTheme`／`themeMode`；任一步失敗都不得留下半套主題。
- 主題切換不觸碰 `game` 規則、亂數、指令數、存檔 bytes 或戰鬥狀態。
- 語系資料缺少七個 `settings.*` 鍵時，整套語系載入拒絕，不退回另一套文字。

## 4. 驗收證據

實作完成的最低 gate：

1. `internal/ui/layout` 幾何測試確認用語兩列與主題並排選項的座標、命中尺寸一致。
2. `internal/ui/render` 像素測試確認用語與主題選取框都可畫出且無缺字。
3. `cmd/dsds` pointer 測試確認四個選項與返回鍵中央命中；Xvfb 測試確認套件可啟動。
4. `internal/prefs` 測試確認未知 `Theme` 不可讀寫，並保留原子設定檔契約。
5. Docker 內執行相關 Go 測試、`git diff --check` 與 `tools/deny_scan.sh --all`；掃描不得
   新增原版執行檔、原版資料、倚天字庫或不在 `docs/images/` 的 PNG。

## 5. 已知限制

這是可發現性與主題偏好切片，不是完整 modern UI。`docs/design/10-visual-modernization.md`
仍維持 DRAFT；P2c 之後才處理資源／指令 icon 與寬版面，三平台封裝仍受 Ebiten backend
工具鏈限制，不能由本規格宣稱完成。
