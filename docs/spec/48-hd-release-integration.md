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

生成圖集的唯一白名單、來源與固定雜湊見 [`generated-hd-assets.json`](../licenses/generated-hd-assets.json)，由 [`deny_scan.sh`](../../tools/deny_scan.sh)檢查。
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
| [command-v.1.0.1-20261014.png](https://github.com/wicanr2/great-era-remake/blob/6d58705096c9ada9dd7f525647401655e96d4e33/docs/images/command-v.1.0.1-20261014.png) | `dist-all/v.1.0.1-20261014/smoke/linux/01-command-hd.png` | `26d1e69e3eee5b9e13c946a7ce221f4ffbc2ef7cfa4c3275467248a7b23b032c` |
| [battle-v.1.0.1-20261014.png](https://github.com/wicanr2/great-era-remake/blob/6d58705096c9ada9dd7f525647401655e96d4e33/docs/images/battle-v.1.0.1-20261014.png) | `dist-all/v.1.0.1-20261014/smoke/promo-capture/15-battle.png` | `12108c6b23a0c1a4e25f5bfde687a7a55d42bfc7fafaa0e101057b2360127f7b` |

指揮畫面的 `smoke/linux/gui-report.json` SHA-256 為 `cf847193bee4d3e0e7b0b850d2b3d4bb899589668afed1f86f588d3749b3df88`；影片擷取收據 `promo/capture.json` SHA-256 為 `185e29f49cf1c27742afeb8223da1dc7946c07d6bb8fb6308e4dd39e2376f827`。兩份收據由本機正式版交付保存，圖片符合既有 `docs/images/*.png` 公開展示白名單。

主要分支整理沒有修改遊戲程式、語系資料、圖集、音訊或授權。後續真機及簽署仍由 [Issue #7](https://github.com/wicanr2/great-era-remake/issues/7)追蹤；分支與文件整理見 [Issue #9](https://github.com/wicanr2/great-era-remake/issues/9)。

## README 遊戲介紹與重新拍攝的展示

2026-10-15 依使用者要求將首頁改為故事背景、玩法與 HD 特色，移除工程紀錄及舊展示圖。工作見 [Issue #12](https://github.com/wicanr2/great-era-remake/issues/12)。以下圖片來自重新擷取的實際視窗。重製版由正式公開 Linux 引擎 ZIP 啟動，種子 1；DOS 由唯讀原版資料副本在 DOSBox-X 執行，只作輔助視覺展示。

人物原版模式從正常查閱選單進入自傳，HD 模式保留同一人物。介面切換會關閉指揮室的閱讀視窗，因此以重新開啟後的真實頁面取代初次示意截圖，不改遊戲程式。原版與 HD 的字距及欄寬不同，頁數也可能不同。

| 現行展示圖 | 來源與範圍 | PNG SHA-256 |
|---|---|---|
| [dos-original-20261015.png](../../docs/images/dos-original-20261015.png) | DOS 原版，DOSBox-X 輔助畫面 | `2f32bb66d2365f902b057b3ab5991f9d1e93485f975aaf11b10334f033c48d0f` |
| [command-original-20261015.png](../../docs/images/command-original-20261015.png) | 重製原版軍政介面 | `ebcdb854c86306fa3fe0a4d61119e01014974eab877e6f135275a5d9986ee60e` |
| [command-hd-20261015.png](https://github.com/wicanr2/great-era-remake/blob/6732585445babd9f9a006aa43d9d4639d3822880/docs/images/command-hd-20261015.png) | 重製 HD 指揮介面 | `26d1e69e3eee5b9e13c946a7ce221f4ffbc2ef7cfa4c3275467248a7b23b032c` |
| [biography-original-20261015.png](../../docs/images/biography-original-20261015.png) | 重製原版人物自傳 | `44f63d0dc185b78715274200fc593a1bc9ed8d1890962e2a1765af8584efe91d` |
| [biography-hd-20261015.png](../../docs/images/biography-hd-20261015.png) | 重製 HD 人物自傳 | `9bb30b6c27b97a27dff502f20fd2d21cea482e38db7de417f6ffecc1b5a1590b` |
| [battle-original-20261015.png](../../docs/images/battle-original-20261015.png) | 重製原版戰場 | `884a71119e8ac50e118e7277503c77f5cfec9a5b2ff5262645002d18a301a553` |
| [battle-hd-20261015.png](../../docs/images/battle-hd-20261015.png) | 重製 HD 六角戰場 | `12108c6b23a0c1a4e25f5bfde687a7a55d42bfc7fafaa0e101057b2360127f7b` |

重製畫面為 1280×720 原始 PNG，未加工；DOS 圖為 640×350，只裁掉模擬器的視窗邊框。DOS 與重製版使用不同遊戲進度，不宣稱同狀態或規則對拍。11 張舊圖由 Git 歷史保存，早期 renderer 及前次展示收據的連結已改到固定歷史版本。原片、正常輸入、版本及來源雜湊保存在本機 `workplace/hoi4-ui-prototype/readme-promo-20261015/`，驗證結果由 Issue #12 收錄。

首頁三個時期的起點對照 `internal/game/scenario.go` 的既有 `Scenarios`：1926 年 7 月、1937 年 7 月及 1941 年 12 月。沒有新增劇情、推導人物生平或改動規則。

本次推廣片為 92.5 秒、1280×720、30 fps，DOS 與重製版全部片段於本輪重新錄製，不沿用舊影片片段。涵蓋原版／HD 介面、地圖縮放拖曳、自傳、開發、徵稅、實際六角格移動、攻擊選單及撤退。配樂使用已採用的原創 `modern_main_theme.ogg`，未使用 DOS 原音。

成片為本機 `dist-all/v.1.0.1-20261014/promo/great-era-remake-promo-comparison-v.1.0.1-20261014-20261015.mp4`，SHA-256 `317e7469db0b7bdd5289189b1a9cddd034bf84d3598e1ba50f1be61c50925d2d`。影音、來源、字幕及抽幀驗證保存在同目錄的 `reshoot-20261015-verification.json`、`reshoot-20261015-captures.json`、`reshoot-20261015-rights.json` 與獨立 `reshoot-20261015-SHA256SUMS.json`。新版只更新本機交付頁的影片入口，不覆寫既有封包、影片、正式 manifest、tag 或 Release 附件。

## 地圖色塊修正與開發插圖

狀態：CONFORMED，限主線呈現與下述測試範圍。2026-10-15 先以 READY 契約實作；使用者指出全國地圖色差及局部陰影，並要求開發頁加入 HD 插圖。沿用已確認的大地色、鋼鐵質感、1280×720 與小型素材限制；工作由 [Issue #13](https://github.com/wicanr2/great-era-remake/issues/13)追蹤，已發布正式版維持原狀。

- 已證實的 remake 原因：`atlasmap.prepareSurface` 以經度 101、104、124 及緯度 32、40 的硬條件切換地表底色；因此產生不依省界的長方形明暗斷層。該路徑的方向光陰影已停用，這次不再以關閉光源掩蓋調色問題。
- 地圖改為連續混合地表色與紋理，限制底色亮度落差。勢力配色、省界、河流、高程、投影、縮放、拖曳及命中邊界沿用。地理呈現仍屬既有近似，不增加規則或真實地理精度聲明。
- 開發頁只替三個現有選項補上墾地、兵工廠、金礦的原創場景插圖。選項順序、編號、語系、矩形、滑鼠及數字鍵 Action 不改；插圖不作規則輸入，也不畫入文字或虛構資源數值。其他流程頁維持既有呈現，原版模式維持原樣。
- 內建圖像生成工具的三份原圖與提示詞先保存在本機 `workplace/hoi4-ui-prototype/map-development-20261015/`。檢查內容後縮製為 384×480 圖集，包含三張 384×160 插圖；執行期只解碼一次，新增 RGBA 上限 737,280 bytes，全部靜態 HD 圖集總量不超過 2.5 MiB。生成原圖不進遊戲包。
- 新圖集由唯一權利清單登錄固定 SHA-256、工具、提示詞與來源雜湊；拒絕掃描僅增列這一個精確路徑。圖片無法解碼或索引無效時，退回既有文字卡片。
- 可重現縮製入口為 [`tools/bake_development_art.go`](../../tools/bake_development_art.go)，在既有 Docker 工具鏈以 `GO111MODULE=off go run tools/bake_development_art.go <三份原圖目錄> <輸出.png>` 執行；來源檔名為 `reclaim-source.png`、`arsenal-source.png`、`goldmine-source.png`。只做固定 4×4 預乘透明度取樣縮圖，不作創意改圖；完整提示詞及原圖雜湊見上述權利清單。
- 驗收先以舊程式重現三條經緯度斷層，再做回歸；正常啟動抽測地圖縮放／拖曳、三個開發卡片及原版切換，以固定種子 1 留圖。測試包含素材尺寸、快取、記憶體、指令對應與缺圖退回。規則、原始輸入及已發布封包不改，不重開原版對拍。

實作驗證：19 個正式 Go 測試套件通過。舊版三條斷層的單通道差為 41.50、57.63、43.16；新版各通道均在回歸容許值 6 以內。PNG 為 383,434 bytes，解碼為 737,280 bytes；全部靜態 HD 圖集合計 2,605,056 bytes，低於 2.5 MiB，重複讀取共用同一圖集。公開縮製工具可重生相同 SHA-256 `f95fd3b3e0616d5d3c5063eae32d609073c92aaafaa46211f379bd9dbc9598ac`。

固定種子 1 從正常入口操作，確認縮放、拖曳、原版美術及介面切換，三個開發卡片完整顯示。墾地使用數字鍵，兵工廠及金礦使用滑鼠；與已發布 Linux 引擎的三組回報及 `.DT1` 逐位元組相同。既有湖北樣本的兵工廠因黃金不足正常拒絕，沒有注入資源製造成功。865 份規則、解碼、語系、配樂及研究程式與 153 份原始遊戲／字庫輸入雜湊不變。不據此宣稱 DOS 規則對拍或平台真機驗收。

目前展示：[修正後地圖](../images/command-hd-20261015-r2.png)、[HD 開發頁](../images/development-hd-20261015.png)，SHA-256 分別為 `9ed15adb908a46da1f62062ab746fc9ef66097129ea5137685094ccc2ca4496d`、`995c5ffeddf1cd648def18c5ca527f37ac22e17af68d0ecac4f65950c1e56e73`。上節原發行畫面的來源收據與舊圖雜湊保留，地圖圖片改連固定 Git 歷史。

本機可玩預覽、比較頁與收據集中於 `workplace/hoi4-ui-prototype/map-development-20261015/`，入口由該父目錄的 README 索引。預覽 binary SHA-256 為 `fb80e4e398ff59b794bad574338e99955364adc32ba90e818aa99fce6fce4658`；`capture.json` 保存正常輸入及截圖，`actions.json` 保存六次正式版／主線指令操作與存檔，`integrity.json` 保存保護雜湊與素材重製結果。

## v.1.0.2-20261015 呈現修正重新發行

狀態：CONFORMED。使用者要求重新打包完整版、Release 與推廣影片，工作由 [Issue #14](https://github.com/wicanr2/great-era-remake/issues/14)追蹤。正式 [Release](https://github.com/wicanr2/great-era-remake/releases/tag/v.1.0.2-20261015)已發布，來源為 `87934a41e8447b02080103b6536eb132edf4fd9f`，固定來源保留於本機工作目錄，不另設發行分支。

- 從乾淨 `master` 提交建立固定來源，收錄上述地圖配色修正與三幅開發插圖。保留 392 篇中文正文、24 位待考、英日 machine-draft 及八曲已採用 Ogg；只更新版本與相關告知，音訊位元組不變。
- 版號為 `v.1.0.2-20261015`。三平台同批二進位分裝本機完整版與公開引擎包，集中於 `dist-all/v.1.0.2-20261015/`。公開仍為 Windows amd64、macOS arm64、Linux x86_64 ZIP；原版資料、倚天字庫、AppImage runtime 與含原版畫面的影片留本機。既有 tag、Release、封包與影片不改。
- 打包入口為 [`tools/package.sh`](../../tools/package.sh)，正式 Go 套件範圍為 `./cmd/... ./internal/... ./tools/...`，避免本機私人草稿被當成正式套件。完整包先重建三平台，公開包核對相同來源後重用本次二進位，不重用舊正式版程式。
- 逐包核對版本、平台架構、CRC、內容雜湊、來源、配樂與第三方告知；Linux／Wine 經正常入口驗證介面、美術、開發插圖、人物、存讀檔與重啟。macOS 限交叉建置及靜態檢查，不擴張真機驗收聲明。
- 推廣片保留 DOS 原版、重製原版與 HD 比較。重製畫面從本版正式封包重新錄製，明確展示地圖及開發插圖；沿用 DOS 輔助原片時保留原錄製來源。配樂沿用使用者已採用的原創混音，依既有決定不改用 DOS 原曲。
- 影片驗證完整解碼、視訊及音訊、時長、非靜音、黑幀、長時間凍結、字幕邊界與抽幀來源。完成後建立新 tag 與 Release，核對遠端附件雜湊，再更新唯一現況與本機交付首頁。

本版六個封包已從固定來源重建，三平台公開與完整包使用相同的本次二進位，均包含新開發圖集。19 個正式 Go 套件、純 Go 閘門、版本、架構、CRC、逐檔雜湊、八曲音訊與完整告知通過。Linux／Wine 各驗完整包及公開包，包含四種介面／美術組合、新開發頁、人物分頁、徵稅、正常存讀檔及重啟。四份 `.DT1` 的 SHA-256 均為 `d275dfadaec6fc6b524e8e1ea512d66e909f6c3b338c55c6c455a16b86b97166`，Go parser 核對省份變化並逐位元組寫回相同。

推廣片為 93.534 秒、1280×720、30 fps、H.264 與雙聲道 AAC，檔名 `great-era-remake-promo-comparison-v.1.0.2-20261015.mp4`，SHA-256 `c4120355e0c986d2308be0caf9db1278f520f7733797495fe71deaa44cee47d4`。17 章完整解碼、字幕邊界、非靜音、黑幀及八秒以上凍結檢查通過；章節中心抽幀及開發插圖全幅圖已視覺核對。平均音量 -20.5 dB，峰值 -6.3 dB，不宣稱新的實機或人耳驗收。

重製片段由本版公開 Linux 封包重新實錄；DOS 片段沿用 2026-10-15 原程式的 DOSBox-X 輔助原片，原收據及雜湊保留，不冒稱本輪重新錄製或規則對拍。素材、字幕、技術及視覺紀錄集中於版本目錄 `promo/`，原片及工具位於 `workplace/hoi4-ui-prototype/release-v.1.0.2-20261015/`。公開僅五份引擎／驗證附件，發布前後的遠端 SHA-256 均與本機核准內容相符；242 份前版交付及 154 份原始輸入雜湊不變，後者包含 153 份素材及一份來源 ZIP。
