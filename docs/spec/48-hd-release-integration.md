# SPEC-48：大地鋼鐵介面與八曲配樂整合

狀態：CONFORMED。實作前的原型已獲使用者確認。使用者於 2026-10-14 授權從獨立發行分支提交、推送並發布 `v.1.0.0-20261014`。
工作追蹤：[介面整合 #8](https://github.com/wicanr2/great-era-remake/issues/8)、[配樂 #5](https://github.com/wicanr2/great-era-remake/issues/5)、[發行 #7](https://github.com/wicanr2/great-era-remake/issues/7)。

## 已確認的呈現與操作

| 項目 | 契約與目前實作 |
|---|---|
| 預設介面 | 參考《鋼鐵雄心 IV》與《大戰略》的頂部資源、中央地圖、側欄操作及底部指令。採土褐、砂岩、鐵黑、暗紅與舊紙。 |
| 獨立切換 | `F2` 切換原版／HD 美術，`F4` 切換原版／指揮介面。保留四種組合與偏好儲存。 |
| 國土地圖 | 高程呈現、縮放、拖曳及傾斜保留，移除強烈方向光陰影。地圖並非現代地理資料的逐點重製。 |
| 戰術畫面 | 垂直俯視六角格、部隊選取、即時部隊及地形資料、既有命令與取消操作。兵種保留原版辨識，採磨損鋼鐵質感。 |
| 鐵路 | 保留 21 種形狀、上／下半格變體、車站、有效索引 0 與 -1 哨兵。繪圖、鐵路端點與點擊共用六角幾何，不新增連接或修改移動成本。 |
| 人物全文 | 416 筆人物、387 篇正文與 29 筆待考。全文、搜尋、分頁及來源保留；待考不補造故事。英日譯稿保留 machine-draft 標示。 |
| 素材大小 | 地表材質 512×512、介面圖集 512×256、兵種圖集 384×192。遊戲不載入大型生成原圖。 |

程式入口：[`earth_battle.go`](../../internal/ui/render/earth_battle.go)、[`earth_frame.go`](../../internal/ui/render/earth_frame.go)、[`earth_hex.go`](../../internal/ui/layout/earth_hex.go)、[`earth.go`](../../internal/ui/hdart/earth.go)。
規則層與存檔 writer 沿用既有實作，新增呈現不宣稱原版逐像素或完整戰鬥 parity。

## 八曲原創配樂

沿用專案原創程序作曲，以純 Go renderer 產生 PCM，再以固定增益製作 44.1 kHz、雙聲道 Ogg Vorbis。八曲響度約 -18 LUFS，保存原始 PCM、增益、來源、編碼參數及 SHA-256。
不使用原版旋律、外部取樣或 SoundFont。六曲 manifest 相容路徑及缺檔後備保留，增加主題與替代戰鬥 cue。

使用者於 2026-10-13 明確回覆「已試聽正常，採用新版混音」。本次保留相同八份 Ogg 位元組，只更新發行版本與告知文件。這份確認涵蓋使用者當時的設備，不推定各平台真機音訊均已驗收。

來源與驗收入口：[`provenance.json`](../../assets/music/modern/provenance.json)、[`acceptance.json`](../../assets/music/modern/acceptance.json)、[`manifest.json`](../../assets/music/modern/manifest.json)。
製作入口：[`finalize_ogg.py`](../../tools/music/finalize_ogg.py)，原始製作腳本另保存在同目錄的 `mastering_source_20261013.py`。

## 發行與權利邊界

公開附件為 Windows amd64、macOS arm64、Linux x86_64 引擎 ZIP，含原創配樂、已核對來源的 HD 圖集及完整第三方告知。公開 Linux ZIP 不含 AppImage runtime。
玩家須自備合法原版資料與倚天字庫，依[發行使用說明](../release/README.md)指定路徑。
完整本機包、原版資料、倚天字庫及含原版畫面的影片均留在本機 `dist-all/v.1.0.0-20261014/`，不附到 GitHub Release。

三張生成圖集的唯一白名單、來源與固定雜湊見 [`generated-hd-assets.json`](../licenses/generated-hd-assets.json)，由 [`deny_scan.sh`](../../tools/deny_scan.sh)檢查。
現代字模來源、完整 OFL 與原始著作權告知見[字模授權](../licenses/modern-font-atlas.md)。專案授權為 [RRSAL-1.0](../../LICENSE)。

## 驗證範圍

- 三平台從同一份乾淨來源重建，封包驗證版本、架構、CRC、逐檔雜湊及授權告知。
- 以舊版版本字串重建三平台執行檔，均與 2026-10-13 已驗收產物逐位元組相同，證明本次移轉來源與驗收來源一致。
- Go 回歸與純 Go 閘門通過。鐵路測試核對原版解碼端點、相鄰格與點擊幾何。
- 新版 Linux 與 Wine 封包由正常入口驗證四種介面／美術組合、自傳、徵稅、存讀檔與重啟。公開 Linux ZIP 另驗自備資料的啟動路徑。
- 既有固定種子 1 的戰術操作證據涵蓋換部隊、移動、攻擊選單與取消、原版切換及撤退。攻擊選單驗收不代表戰損已獨立對拍。
- macOS 真機、Windows 真機、簽署與公證仍待驗證。交叉建置、靜態檢查及 Wine 均保留各自證據範圍。

封包、存檔、音訊與來源的本機收據位於版本目錄的 `SHA256SUMS.json`、`verification.json` 與 `smoke/`；公開發行結果與其後續驗證集中於 [Issue #7](https://github.com/wicanr2/great-era-remake/issues/7)。

## v.1.0.1-20261014 人物資料發行封存

狀態：CONFORMED。使用者於2026年10月14日確認本版號，要求正式Release、三平台完整版與推廣影片，停止剩餘24位考證。發行及公開附件已完成，收據見本文件最後的主要分支整併記錄；已發布 tag 內的原規格保持不動。

- 輸入為前版發行來源，以及已驗收392篇正文的18份增量。來源限定短傳、未審譯稿與待考標示保留，不補身份或生平。
- 原有玩法、介面、美術與八曲混音沿用。原始資料唯讀；公開包與本機完整包維持既定權利界線。
- 三平台同一批編譯產物分裝公開與本機包。Go回歸執行一次，另一種包核對相同來源與二進位雜湊後重用，逐包獨立驗證內容與啟動器。
- 驗收包括來源拒絕掃描、純Go、語系來源、封包架構及逐檔雜湊、Linux／Wine正常操作與存讀檔、代表新短傳閱讀、影音完整解碼及抽幀。macOS限交叉建置與靜態檢查。
- 影片保留DOS原版與重製原版／HD比較，新重製段由本版封包錄製；沿用DOS輔助實錄時明記原錄製來源，不宣稱同態或原版規則對拍。
- 已發布tag、Release與附件不移動或覆寫。新版本通過上述檢查後建立獨立提交、tag與Release，交付根目錄為`dist-all/v.1.0.1-20261014/`。

## 主要分支整併與展示收據

2026-10-15 依使用者要求，將已發布來源及其提交歷史整併至 `master`。正式版仍為 [v.1.0.1-20261014](https://github.com/wicanr2/great-era-remake/releases/tag/v.1.0.1-20261014)，來源 `85ed56095117134fd3fdbf64efd7a5411eaeabf7`。三平台公開包與遠端附件、392 篇正文、八曲配樂、19 個 Go 套件及四條正常操作驗收見[公開收據](https://github.com/wicanr2/great-era-remake/releases/download/v.1.0.1-20261014/RELEASE-VERIFICATION.json)。

展示截圖沿用 v.1.0.1-20261014 的實際 Linux 遊戲視窗，固定種子 1，原尺寸 1280×720，沒有加工。指揮畫面來自正常封包啟動；戰術畫面來自正常出兵流程，未注入座標、勝利或資源。範圍是介面與美術展示，不能由圖片宣稱原版戰損或規則等價。

| 展示圖片 | 原始本機收據路徑 | PNG SHA-256 |
|---|---|---|
| `docs/images/command-v.1.0.1-20261014.png` | `dist-all/v.1.0.1-20261014/smoke/linux/01-command-hd.png` | `26d1e69e3eee5b9e13c946a7ce221f4ffbc2ef7cfa4c3275467248a7b23b032c` |
| `docs/images/battle-v.1.0.1-20261014.png` | `dist-all/v.1.0.1-20261014/smoke/promo-capture/15-battle.png` | `12108c6b23a0c1a4e25f5bfde687a7a55d42bfc7fafaa0e101057b2360127f7b` |

指揮畫面的 `smoke/linux/gui-report.json` SHA-256 為 `cf847193bee4d3e0e7b0b850d2b3d4bb899589668afed1f86f588d3749b3df88`；影片擷取收據 `promo/capture.json` SHA-256 為 `185e29f49cf1c27742afeb8223da1dc7946c07d6bb8fb6308e4dd39e2376f827`。兩份收據由本機正式版交付保存，圖片符合既有 `docs/images/*.png` 公開展示白名單。

主要分支整理沒有修改遊戲程式、語系資料、圖集、音訊或授權。後續真機及簽署仍由 [Issue #7](https://github.com/wicanr2/great-era-remake/issues/7)追蹤；分支與文件整理見 [Issue #9](https://github.com/wicanr2/great-era-remake/issues/9)。
