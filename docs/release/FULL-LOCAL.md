# 三平台本機完整版

正式版號及驗收範圍統一見 [CONTEXT.md](../../CONTEXT.md)。本機封包與推廣片集中於專案根目錄 `dist-all/<版本>/`，素材權利不因版號改變。

本包含大地色俯視六角戰場、Modern 滑鼠介面、三語全文閱讀、原版遊戲資料、完整倚天字庫、八首原曲 HD 重編與三首軍歌 Ogg。
資料與字庫來自本機持有來源，尚無公開散布授權。此包僅供本機使用，不可上傳或公開分享。
「完整版」表示啟動所需資源隨包提供，待考人物仍維持明示狀態，數量見專案現況。

| 平台 | 啟動方式 | 存檔位置 |
|---|---|---|
| Windows amd64 | 解壓縮後執行 `GreatEraRemake/dsds.exe 或 launch.cmd` | `%LOCALAPPDATA%\GreatEraRemake\Saves` |
| macOS arm64 | 解壓縮後開啟 `GreatEraRemake.app` | `~/Library/Application Support/GreatEraRemake/Saves` |
| Linux x86_64 | 賦予 AppImage 執行權限後開啟；無 FUSE 時先解包，再執行 `squashfs-root/AppRun` | `$XDG_DATA_HOME/GreatEraRemake/Saves`，未設定時使用 `~/.local/share/GreatEraRemake/Saves` |

完整版的執行檔會自行尋找包內資料、字庫、語系與配樂。即使從其他工作目錄直接執行，也不需另設 `-game` 或 `-eten`。
原版 EXE、備份壓縮包與玩家進度不放入封包。`data/game/SAVE(1)`、`SAVE(2)` 是唯讀初始資料。
玩家進度與偏好寫使用者目錄，遊戲不改寫包內資料。

首次啟動預設採 Modern 介面、HD 美術、滑鼠及 HD 配樂。常駐工具列可即時切換美術、介面、HD／原版／靜音及自動或十一首指定曲目，並保存偏好。`F6` 可用鍵盤操作工具列，AI 沿用現況。`F2` 切換美術，`F4` 切換介面，`F7` 開啟整備模式的軍備局。
滑鼠滾輪縮放、拖曳地圖；人物全文可捲動。固定按鈕支援返回，徵兵可拖曳滑桿，調動可點選省份與將領。`ESC` 返回；`F10` 後按 `Y` 儲存離開。
八首原曲依完整 MUS 重編；《還我河山》《國民革命歌》《歌八百壯士》依歷史曲譜編配，以 48 kHz 雙聲道 Ogg 播放。軍歌切到原版音樂時播放當前情境曲，回 HD 後恢復軍歌。
曲目來源、全曲技術驗收與先前試聽方向採用分別記錄在包內 `assets/music/modern/provenance.json`、`audio-qa.json`、`acceptance.json`；不宣稱使用者已聽過本次所有全曲，其他平台真機仍待驗收。

開局可選三劇本全部 14 個勢力。中國勢力另可選軍備整備試玩，研究、生產、庫存、換裝與裝甲改編保存於單一 `campaign.gesave`；完整統一後入侵仍在開發。開局畫面可繼續存檔，並保留 DT1 匯入。戰術可暫停回地圖、保存及續戰；整編取消會還原部隊與裝備。

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
DSDS_PACKAGE_RELEASE_CHANNEL=stable bash tools/package.sh v.1.1.0-20261017 full-local
```

輸出固定為專案根目錄的 `dist-all/<版本>/full-local/`。腳本拒絕覆寫既有封包，所有工作在受限 Docker 中執行。
推廣影片另存 `promo/`，不放進遊戲封包。工作追蹤見 [Issue #7](https://github.com/wicanr2/great-era-remake/issues/7)。

零參數啟動與音樂回歸工具為 [tools/release/smoke.py](../../tools/release/smoke.py)，只在有界 Docker 中執行；依檔頭使用方式傳入版本、平台及封包種類。
