# 新 HD 三平台引擎與配樂包

[正式下載](https://github.com/wicanr2/great-era-remake/releases/latest)提供三平台引擎包。現行發行版為 [v.1.0.1-20261014](https://github.com/wicanr2/great-era-remake/releases/tag/v.1.0.1-20261014)，各包雜湊及驗收範圍見[公開驗證附件](https://github.com/wicanr2/great-era-remake/releases/download/v.1.0.1-20261014/RELEASE-VERIFICATION.json)。主要分支為 `master`，發行來源固定在同名 tag。

公開引擎包使用本文件。本機資源完整包另見 [FULL-LOCAL.md](FULL-LOCAL.md)，
兩種封包的素材範圍、啟動預設及散布狀態分開記錄。

本包使用已確認的指揮介面、HD 小型材質與三語全文閱讀。預設為新介面與 HD，
`F2` 切換美術、`F4` 切換操作介面。`ESC` 返回；`F10` 後按 `Y` 儲存並離開。
實機、簽署與公證由 [Issue #7](https://github.com/wicanr2/great-era-remake/issues/7)追蹤。
八曲混音已採用，試聽範圍見包內 `assets/music/modern/acceptance.json`。

封包包含引擎、三語資料、已核對照片、RRSAL-1.0 與字型／素材告知。
玩家需自備合法原版遊戲資料與倚天字庫，以 `-game`、`-eten` 指定目錄。
原版資料、倚天字庫及玩家存檔不放進引擎包。包內包含八首專案原創 Ogg 配樂及來源／混音紀錄。

| 平台 | 入口 | 預設存檔目錄 |
|---|---|---|
| Windows amd64 | `GreatEraRemake/launch.cmd` | `%LOCALAPPDATA%\GreatEraRemake\Saves` |
| macOS arm64 | `GreatEraRemake.app/Contents/MacOS/launch` | `~/Library/Application Support/GreatEraRemake/Saves` |
| Linux x86_64 | 解壓縮 ZIP 後的 `AppDir/AppRun` | `$XDG_DATA_HOME/GreatEraRemake/Saves`，未設定時使用 `~/.local/share/GreatEraRemake/Saves` |

從包內入口指定自備資料，例如 Linux：

```sh
./AppDir/AppRun \
  -game /絕對路徑/遊戲資料 -eten /絕對路徑/eten
```

可加 `-locale translations/en` 或 `-locale translations/ja` 切換語系。
可用 `-save` 指定另一個可寫存檔位置；偏好設定沿用系統使用者設定目錄。
Go 1.26 的系統需求為 Windows 10 或 Windows Server 2016 以上、macOS 12 以上，見
[Go 官方最低需求](https://go.dev/wiki/MinimumRequirements)。macOS 封包為 Apple Silicon，封包與 Mach-O 均宣告 12.0。
Linux 需要 glibc 2.34 以上、桌面 X11、OpenGL 與 ALSA 函式庫；
glibc 需求由執行檔的實際符號版本核對。公開引擎 ZIP 不包含 AppImage runtime，也不需要 FUSE。

啟動器預設以 `-audio modern` 播放八首原創 Ogg 混音；`-audio off` 關閉。
地圖及替代戰鬥已有 Ogg，程序生成只作缺檔後備。使用者已試聽確認採用；其他平台真機仍待驗收。
圖形、美術與音訊選項彼此獨立。

`-version` 顯示完整版號。每包的 `RELEASE-MANIFEST.json` 保存版本、平台及逐檔雜湊；
發行輸出根目錄的 `SHA256SUMS.json` 保存封包的大小、雜湊、來源與建置工具。

在隔離專案根目錄重建：

```sh
DSDS_PACKAGE_RELEASE_CHANNEL=stable bash tools/package.sh v.1.0.1-20261014 patch
```

輸出固定在 `dist-all/<版本>/patch/`，驗證記錄放 `smoke/`。腳本拒絕覆蓋既有輸出，
全程使用受限、無網路、目前使用者身分的 Docker。版本識別不建立 tag 或公開 Release。
Linux 容器與 Wine 的啟動結果分別記錄；它們不能代替 Windows／macOS 真機聆聽或簽署。
## 重建來源

`tools/package.sh <版本> full-local` 在Docker內執行一次Go回歸及三平台編譯。接著以`DSDS_PACKAGE_REUSE_FULL=1 tools/package.sh <版本> patch`產生公開包，`tools/release/reuse_build.py`核對相同來源提交、工具映像、逐檔雜湊與三份二進位後重用。兩種包各自驗證內容、架構與啟動器。
