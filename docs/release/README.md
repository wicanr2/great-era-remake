# 發行狀態與三平台包裝

`tools/package.sh` 是不含遊戲資料的三平台發行候選包唯一入口。所有編譯、封裝、
AppImage 解包檢查與雜湊都在固定 Docker 工具鏈內完成，輸出預設在被 `.gitignore`
忽略的 `dist-all/`；主機只負責啟動受限容器與保留產物。推廣影片同樣集中在
`dist-all/` 的版次子目錄下，但絕不被加入任何 release archive。

```sh
# Windows amd64 ZIP、macOS arm64 app bundle ZIP、Linux x86_64 AppImage
tools/package.sh 0.1.0-release dist-all windows-amd64,darwin-arm64,linux-appimage

# 可選的 legacy Linux tar.gz，供 CI 或除錯使用，不是三平台主發行格式
tools/package.sh 0.1.0-release dist-all linux-amd64
```

第三個參數是逗號分隔的目標清單；未指定時的預設就是
`windows-amd64,darwin-arm64,linux-appimage`。同一版本若已有同名產物，腳本會
失敗即關閉（fail-closed）拒絕覆蓋；只有明確設定 `DSDS_RELEASE_OVERWRITE=1` 才可
重建。主要產物為：

| 平台 | 檔案 | 封裝內入口 | 本輪技術驗證邊界 |
|---|---|---|---|
| Windows x86_64 | `great-era-remake-windows-amd64-<版本>.zip` | `GreatEraRemake/dsds.exe` | Docker 交叉建置；Windows 實機 smoke 待補 |
| macOS Apple Silicon | `great-era-remake-macos-arm64-<版本>.zip` | `GreatEraRemake.app` | Mach-O arm64 檢查；macOS 實機、簽署與公證待補 |
| Linux x86_64 | `great-era-remake-linux-x86_64-<版本>.AppImage` | AppImage `AppRun` | AppImage 解包、payload 拒絕掃描；不同發行版實機 smoke 待補 |

每個包都附 `RELEASE-MANIFEST.txt`，記錄版本、目標、來源樹是否 dirty、commit、
封裝內每檔 SHA-256，以及下列不可省略的素材邊界：

```text
contains_original_game_data=no
contains_original_audio_or_font=no
requires_user_supplied_legal_game_data=yes
```

封裝前及 Linux AppImage 解包後都會拒絕原版副檔名、`workplace`／`game` 目錄與
`.MUS`／`.TIM`／`.DAT`／`.OGG`／`.WAV` 等素材。因此發行包只有 clean-room 引擎、
語系、Modern `GEMF` 字型 atlas、OFL notice、操作說明與本專案自行繪製的應用程式圖示；
不包含原版遊戲資料、原版美術、原版音樂、原版執行檔或倚天字型。玩家必須自行準備合法
取得的 `-game` 資料目錄與 `-eten` 字庫，這是刻意的版權邊界，不是缺檔錯誤。

遊戲與規則層沒有 cgo 原始碼；`tools/check_no_cgo.sh` 會檢查 `cmd/` 與 `internal/`
沒有 `import "C"`／`#cgo`。目前鎖定的 Ebiten v2.8.8 桌面 GLFW backend 仍使 Linux
與 macOS 的最終連結需要工具鏈 cgo：Linux 使用 `gcc`，macOS arm64 使用
`oa64-clang`；Windows amd64 目前以 `CGO_ENABLED=0` 交叉建置。這些是第三方桌面
backend 邊界，並不代表 remake 自行引入 C 程式碼。

這些是可供持有合法資料者測試的發行候選包，不等同已完成的公開正式發行：Windows／
macOS／Linux 的正常玩家路徑實機 smoke、macOS 簽署與公證、正式 Ogg 的 provenance
與人耳 QA、Android binding／真機，以及英日人物稿的人審皆保留為各自的 gate。推廣
影片工作流見 [`docs/promo/`](../promo/README.md)；影片與本機音訊預覽不打進 release。

## 2026-08-11 `0.1.0-release` 候選包

| 平台 | 產物 | SHA-256 | 已完成的封包驗證 |
|---|---|---|---|
| Windows x86_64 | `great-era-remake-windows-amd64-0.1.0-release.zip` | `78dfe45eff53efc1e10a3d6e41106a00ea3b6c0c663fe01d87a02375a8caefd2` | ZIP CRC 與 `GreatEraRemake/dsds.exe` 入口 |
| macOS Apple Silicon | `great-era-remake-macos-arm64-0.1.0-release.zip` | `be70f73383ae7f2cb1970371684b08137692da13b5224dd6e82c50a5199bc69d` | ZIP CRC 與 app bundle 入口；建置時 Mach-O arm64 檢查 |
| Linux x86_64 | `great-era-remake-linux-x86_64-0.1.0-release.AppImage` | `e92b3dc3c4ee1f1d9d3a24415d7c55f78de8e305c20c9b950aed8f385cf3450d` | AppImage 解包、`AppRun`／`dsds` 存在與原版素材副檔名拒絕 |

這是已重建的本機候選產物，不是公開正式版本。仍須補 Windows 與 macOS 的正常玩家路徑
實機 smoke、macOS 簽署／公證，以及不同 Linux 發行版上的 AppImage smoke；正式音訊、
英日人物稿與 Android gate 也維持獨立，不因封包結構通過而宣稱完成。

## 2026-08-12 `0.1.0-h4-m1-20260812` H4 候選包

本輪版本化交付根目錄為 `dist-all/2026-08-12-h4-m1/`；其中三個封包與
`SHA256SUMS-0.1.0-h4-m1-20260812.txt` 集中放置，連續實機影片則在其 `promo/` 子目錄。
這避免覆寫較早的候選包或歷史影片。

| 平台 | 產物 | SHA-256 | 已完成的封包驗證 |
|---|---|---|---|
| Windows x86_64 | `great-era-remake-windows-amd64-0.1.0-h4-m1-20260812.zip` | 同目錄 `SHA256SUMS-0.1.0-h4-m1-20260812.txt` | ZIP 全檔 CRC、入口 `GreatEraRemake/dsds.exe`、素材拒絕檢查 |
| macOS Apple Silicon | `great-era-remake-macos-arm64-0.1.0-h4-m1-20260812.zip` | 同上 | ZIP 全檔 CRC、app bundle 入口；交叉建置時 Mach-O arm64 檢查 |
| Linux x86_64 | `great-era-remake-linux-x86_64-0.1.0-h4-m1-20260812.AppImage` | 同上 | AppImage 解包、`AppRun`／`dsds` 存在與原版素材副檔名拒絕 |

本輪另以獨立 Docker 驗證重新讀取兩個 ZIP 的所有檔案與 CRC、檢查 Windows／macOS
入口、解出 AppImage 並掃描禁止素材類型，以及核對三個檔案雜湊；所有檢查通過。這只證明
交叉建置與封包結構，**不是** Windows／macOS 真機 smoke、簽署、公證或公開正式發行。
影片見 [`docs/promo/README.md`](../promo/README.md)，不含在上列任何 archive 中。

## 2026-08-12 `0.1.0-a2-m2-portrait-20260812` A2 M2 肖像版候選包

交付根目錄為 `dist-all/2026-08-12-a2-m2-portraits/`，包含 Windows amd64 ZIP、macOS
arm64 app ZIP、Linux x86_64 AppImage 與 `SHA256SUMS-0.1.0-a2-m2-portrait-20260812.txt`。
三包均以同一 dirty 工作樹重建，包含 Modern 民國測繪 UI、15 項專屬指令圖示、受控肖像
登錄、蔣中正公有領域 JPEG 與 `portraits-NOTICE.md`；不含原版資料、音訊或倚天字型。

| 平台 | SHA-256 |
|---|---|
| Windows x86_64 ZIP | `e36079703c9aad82cbfb3cf6e130d2b8d568826a6e8b468a70356167878228b3` |
| macOS arm64 app ZIP | `1c11662a7987afb46d1ca2cf8fa446ffc81d9629513018b105778c38a2923c0b` |
| Linux x86_64 AppImage | `3a6fa7a47537ca8dcca21da4d0fde0d9ebdc1f5f2fdc5deec56ee6b2a35ccfd5` |

ZIP 入口／CRC、Mach-O arm64、AppImage 解包與禁止素材掃描均由 `tools/package.sh` 的
失敗即關閉流程驗證。這仍是交叉建置候選，不等於 Windows／macOS／Linux 真機 smoke、
macOS 簽署／公證或公開正式發行。
