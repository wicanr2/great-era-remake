# 三平台本機完整版

正式版號及驗收範圍統一見 [CONTEXT.md](../../CONTEXT.md)。本機封包與推廣片集中於專案根目錄 `dist-all/<版本>/`，素材權利不因版號改變。

本包含大地色俯視六角戰場、磨損鋼鐵 HD 介面、三語全文閱讀、原版遊戲資料、完整倚天字庫與八首 Ogg 配樂。
資料與字庫來自本機持有來源，尚無公開散布授權。此包僅供本機使用，不可上傳或公開分享。
「完整版」表示啟動所需資源隨包提供，待考人物仍維持明示狀態，數量見專案現況。

| 平台 | 啟動方式 | 存檔位置 |
|---|---|---|
| Windows amd64 | 解壓縮後執行 `GreatEraRemake/launch.cmd` | `%LOCALAPPDATA%\GreatEraRemake\Saves` |
| macOS arm64 | 解壓縮後開啟 `GreatEraRemake.app` | `~/Library/Application Support/GreatEraRemake/Saves` |
| Linux x86_64 | 賦予 AppImage 執行權限後開啟；無 FUSE 時先解包，再執行 `squashfs-root/AppRun` | `$XDG_DATA_HOME/GreatEraRemake/Saves`，未設定時使用 `~/.local/share/GreatEraRemake/Saves` |

入口會自動指定包內資料、字庫與配樂，不需另設 `-game` 或 `-eten`。
原版 EXE、備份壓縮包與玩家進度不放入封包。`data/game/SAVE(1)`、`SAVE(2)` 是唯讀初始資料。
玩家進度與偏好寫使用者目錄，遊戲不改寫包內資料。

預設採指揮介面、HD 美術與現有 modern 配樂。`F2` 切換美術，`F4` 切換介面，
滑鼠滾輪縮放，拖曳地圖。「人物檔案」可搜尋與翻頁。`ESC` 返回；`F10` 後按 `Y` 儲存離開。
配樂沿用使用者確認的專案原創程序作曲，八曲統一至約 -18 LUFS，以 44.1 kHz 雙聲道 Ogg 播放。
曲目來源與固定增益混音記錄在包內 `assets/music/modern/provenance.json`。使用者已確認目前設備試聽正常；其他平台真機仍待驗收。

可在啟動入口後加參數：

```text
-audio off
-locale translations/en
-locale translations/ja
-version
```

系統需求：Windows 10 以上、Apple Silicon 與 macOS 12 以上；Linux glibc 2.34 以上，
並有 X11、OpenGL、ALSA。Windows／macOS 真機、簽署與公證尚待驗收，
Linux 容器與 Wine 輔助測試分別記錄。AppImage runtime 的公開發行來源核對仍未完成。

`RELEASE-MANIFEST.json` 保存平台、版本及逐檔雜湊。版本根目錄的 `SHA256SUMS.json`
保存三包大小、SHA-256、建置來源與工具映像；`full-local/PRIVATE-INPUTS.json` 列出本機素材來源。
RRSAL-1.0 適用於有權授權的 remake 程式碼，不授予原版素材或倚天字庫的散布權。

從專案根目錄建置正式版：

```sh
DSDS_PACKAGE_RELEASE_CHANNEL=stable bash tools/package.sh v.1.0.2-20261015 full-local
```

輸出固定為專案根目錄的 `dist-all/<版本>/full-local/`。腳本拒絕覆寫既有封包，所有工作在受限 Docker 中執行。
推廣影片另存 `promo/`，不放進遊戲封包。工作追蹤見 [Issue #7](https://github.com/wicanr2/great-era-remake/issues/7)。
