# 遊戲內用語切換與偏好持久化驗證

日期：2026-08-02  
狀態：**confirmed（顯示設定／prefs.json）**

## 入口與邊界

政略指令選單按 `O` 開啟 remake 新增的「顯示設定」子頁：

```text
1 原典用語
2 現代白話
ESC 返回
```

`O` 目前是 Options 快捷鍵，未覆蓋既有按鍵。這個子頁不冒充原版第 15 項「其他選項」
的完整實作。勘誤：「授權自治」實際是第 8 項「政策」的第一項，
玩家流程已於 `docs/playtest/23-policy-autonomy.md` 接通。本顯示設定頁仍可在後續納入
完整的第 15 項子選單。

## 持久化契約

- 預設路徑：`$XDG_CONFIG_HOME/dsds/prefs.json`；未設 XDG 時由平台使用者設定目錄推導。
- `-prefs <path>` 可指定測試或可攜版路徑。
- 寫入權限 `0600`。
- 同目錄建立暫存檔，寫完 `fsync` 後以 `rename` 原子取代；不遺留半份 JSON。
- 檔案不存在：使用內建預設 `theme=retro`、`scale=2`、wording 空白＝original。
- 壞 JSON 或未知 wording：整份退回內建預設並在 stderr 明示，不部分套用。
- 優先序：明示 `-wording` > `prefs.json` > 內建 original。
- 命令列覆寫只影響該次啟動，不改寫偏好檔。

## 三次 Docker Xvfb 實跑

測試偏好路徑為 `workplace/prefs-e2e/prefs.json`。

### 1. 遊戲內切換

從沒有偏好檔開始：

```text
Enter → O → 畫面選中「1 原典用語」
→ 2 → 畫面立即改選「2 現代白話」
→ ESC → 1 調動
```

調動畫面立即顯示「要怎麼調動？」；stderr 明示切換與偏好路徑。產物內容：

```json
{
  "theme": "retro",
  "wording": "plain",
  "scale": 2
}
```

### 2. 重啟沿用

不給 `-wording`、沿用同一 `-prefs` 路徑重啟，進入調動後仍顯示現代白話。

### 3. 命令列優先

沿用同一偏好檔但加 `-wording original`，調動畫面切回原典字模；退出後偏好檔仍是
`"wording": "plain"`，證明覆寫沒有偷改持久設定。

## 遊戲狀態隔離

三次只操作顯示設定並由 F10/Y 自動存檔。三份 `.DT1` 均與原始輸入 `cmp` 逐位元組相同，
SHA-256 都是 `9ad9359b436ce2450ff3460df4efede478d871dfe234f60299759e781959ad0d`。
因此偏好不在遊戲存檔中，也沒有因畫面切換改到回合、資源、將領或指令數。

## 驗證產物

- `workplace/screenshots/prefs-ui/01-default-original.png`
- `02-switched-plain.png`
- `03-transfer-now-plain.png`
- `04-restart-pref-plain.png`
- `05-cli-override-original.png`
- `workplace/prefs-e2e/prefs.json` 與三份 `.DT1`

全部是 ignored 驗證產物；一次性容器使用 `--rm`、`--network none`、有限 CPU／記憶體／
程序數，並以目前 UID/GID 寫入。
