# PLAYTEST-40：modern HUD 資源／指令圖示 P3

日期：2026-08-10  
狀態：**PASS（P3-HUD 窄切片；Modern 完整 640×350 外殼另見 `docs/design/10`）**

## 固定輸入與差異界線

- 輸入資料：`workplace/orig/game` 的唯讀原版資料、`SAVE(1).DT1`、省 26、固定 screenshot
  路徑；原版資料不進輸出包。
- logical 畫布：640×350；主題：`retro`／`modern`；modern 圖示由純 Go provider
  deterministic 生成，沒有 PNG／SVG 或原版像素拷貝。
- P3 只新增 6 種資源與 15 種政略指令的 16×16 輔助圖示；文字、數值、索引、滑鼠／
  觸控命中區與規則狀態保持不變。retro 路徑維持原版字模，不畫 modern 圖示。

## 截圖證據

以 Docker 內 `cmd/screenshot` 產生同一省份／同一版面：

| 主題 | 入口 | SHA-256 | 用途 |
|---|---|---|---|
| retro | `-theme retro -menu` | `0d5a36a776b0b61566e816d3488c33b010dffd5bc35c88c3c27949711fe8316a` | 原典 fallback reference |
| modern | `-theme modern -menu` | `11d04a46e9a7e20061e3a93cbe7695b6ce09061d3670b8be26b35636588b7572` | modern HUD icon slice |

兩張圖均是近代化窄切片對照，不宣稱與原版逐像素相同；故意差異是 modern 地形／
鐵路／部隊／HUD 圖示。`docs/images/` 白名單只保留既有展示圖，這兩張暫留 `/tmp`，
不把含原版資料衍生的 PNG 放進儲存庫。

## Docker 驗證

```text
go test ./internal/ui/theme ./internal/ui/render ./cmd/screenshot ./internal/ui/actions ./internal/ui/layout ./internal/i18n
DISPLAY=:99 go test ./cmd/dsds
```

另驗證 renderer 對超出調色盤索引的 HUD 圖示採 fail-closed、
`tools/check_no_cgo.sh`、`tools/deny_scan.sh --all` 與 `git diff --check`；
所有工作容器使用 `--rm`，沒有留下專案容器。

## 後續 polish（不阻塞 640×350 Modern UI）

- 十勢力色／徽號、向量字型、寬版 HUD、完整 Android 48dp 版面。
- modern 專屬正式美術資產的來源／授權審查與三平台發行包。
