# 工作紀錄

本檔保存主要分支整併後的工作歷程。現況見 [CONTEXT.md](CONTEXT.md)，後續工作由 [GitHub 總 Issue](https://github.com/wicanr2/great-era-remake/issues/1)追蹤；整併前的研究紀錄與日期化文件保留在原位置及 Git 歷史。

## 2026-10-15 整併正式版至 master

使用者要求合併分支並整理 GitHub 專案。遠端預設分支為 `master`，原提交 `ad1c42e298a4cd76d80ddf56dda138f48a90d1af`。以獨立工作樹將 `release/v.1.0.1-20261014` 快轉整併至 `85ed56095117134fd3fdbf64efd7a5411eaeabf7`；其中已包含 `exp/dosgolem-isr` 及前版發行的全部提交，無衝突，也沒有未合併 PR。

`dosgolem/` 隨既有提交 `d612b18` 保留研究工具源碼快照，來源為 dosgolem 專案提交 `d9c0c27`。本輪沒有擴充模擬器或重開原版研究。

README 整理為正式下載、現行畫面、操作、驗證限制及文件導航。CONTEXT 的開頭改為目前狀態，原有長篇研究內容明示為歷史。發行文件補上正式版與驗證附件連結，SPEC-48 追加已發布版本的 CONFORMED 收據；舊 tag 內容保持原狀。

新增兩張既有正常玩家路徑截圖至 `docs/images/` 展示白名單，來源是 v.1.0.1-20261014 的 Linux 指揮介面與正常出兵戰術畫面，種子皆為 1；原尺寸 1280×720，沒有加工或重新生成素材。來源、收據與 SHA-256 見 [SPEC-48](docs/spec/48-hd-release-integration.md#主要分支整併與展示收據)。

驗證比對正式版全部 1,405 份檔案；1,074 份遊戲、工具、研究工具快照、資料及授權輸入逐檔不變。只修改五份既有文件，新增本工作紀錄及兩張展示圖。52 個相對連結、截圖尺寸／來源、純 Go 閘門及 1,408 檔拒絕掃描通過，11 張展示圖位置合規。遊戲執行期未變，沿用同來源正式版已通過的 19 套件回歸收據，不重跑相同測試。

[Issue #9](https://github.com/wicanr2/great-era-remake/issues/9)保存本輪驗證摘要、推送後的 GitHub 核對與容器收尾；[總 Issue #1](https://github.com/wicanr2/great-era-remake/issues/1)仍為後續工作入口。原工作樹的四筆未提交變更，以及兩版正式 tag、Release、封包與發行分支保留。本輪 Docker 使用非 root、無網路、有資源限制的一次性容器，檔案擁有權檢查零異常。

## 2026-10-15 未提交變更審查與遠端分支清理

使用者要求確認未提交內容是否需要保存到 GitHub，並明確要求遠端只保留 `master`。先核對三個待刪分支的全部提交均在 `master`，再以原子推送刪除遠端 `exp/dosgolem-isr`、`release/v.1.0.0-20261014`、`release/v.1.0.1-20261014`。已發布 tag、Release、附件及本機工作樹保留，不改寫歷史。

| 原未提交檔案 | 處理 |
|---|---|
| `README.md` | 舊版發行說明已由主線正式版首頁取代，不覆蓋新版 |
| `CONTEXT.md` | 發行及本機路徑由現況與 Issue 保存；5.180 的歷史摘要已涵蓋於補存的 playtest 第 8 節 |
| `docs/playtest/43-remake-original-gameplay-diff.md` | 保存原有第 8、9 節，附來源文件與當前輸入雜湊，標為歷史待覆核，不宣稱本輪重跑或恢復原版研究 |
| `dosgolem/internal/dos/scancode.go` | 保存主鍵盤數字名稱功能，改為共用既有 `KeyForRune`，避免複製第二份掃描碼表 |

既有數字鍵補丁在主線確實缺漏。先補 READY 契約及實際 BIOS 佇列測試，基線明確拒絕 `0`；修正後十個數字、Return、KP2 的字組、順序、查詢不取走、讀取計數與未知名稱拒絕通過。Docker 執行 `go test -count=1 ./internal/dos ./cmd/probe ./oracle` 三套件通過。初次測試將既有 int 計數誤寫為 uint64，比對型別修正後以相同容器重跑，失敗紀錄保留為測試腳本問題。

只改獨立研究工具、測試及文件，不修改遊戲引擎或重建正式封包。原工作樹四份差異保留；不把私人交付路徑、過期首頁與未覆核假說直接當成現行公開聲明。公開完成證據、素材掃描、推送與容器收尾集中 [Issue #10](https://github.com/wicanr2/great-era-remake/issues/10)。

提交前檢查 342 份遊戲程式、資料、配樂及授權逐檔不變，原工作樹差異逐位元組相同。38 個文件相對連結、Go 格式、純 Go 與 1,408 檔素材拒絕掃描通過；兩個已發布 Release 與全部附件的識別、大小及雜湊與整併前一致。GitHub 分支 API 已確認僅剩 `master`，本機原稿及工作樹不刪除。


## 2026-10-15 文件去重與唯一入口

依使用者要求移除過期工作清單、失效人物狀態表、被後續介面取代的方案及舊設計草稿。CONTEXT 僅留現況與有效索引，README、AGENTS、語系與發行文件不再各自維護正文數量及待辦。CLAUDE 入口直接匯入 AGENTS。歷史內容由下列既有 Git 提交保存，沒有再複製到另一份文件。

## 已退役文件與歷史入口

| 內容 | Git 歷史 |
|---|---|
| 舊 CONTEXT 完整脈絡 | [歷史版本](https://github.com/wicanr2/great-era-remake/blob/fae384579b8f63e974f5c0c6f6b701f800ce6b36/CONTEXT.md) |
| WORKLIST.md | [歷史版本](https://github.com/wicanr2/great-era-remake/blob/fae384579b8f63e974f5c0c6f6b701f800ce6b36/WORKLIST.md) |
| docs/reference/people/02-status.md | [歷史版本](https://github.com/wicanr2/great-era-remake/blob/fae384579b8f63e974f5c0c6f6b701f800ce6b36/docs/reference/people/02-status.md) |
| docs/design/10-visual-modernization.md | [歷史版本](https://github.com/wicanr2/great-era-remake/blob/fae384579b8f63e974f5c0c6f6b701f800ce6b36/docs/design/10-visual-modernization.md) |
| docs/design/20-biography-system.md | [歷史版本](https://github.com/wicanr2/great-era-remake/blob/fae384579b8f63e974f5c0c6f6b701f800ce6b36/docs/design/20-biography-system.md) |
| docs/design/22-biography-integration.md | [歷史版本](https://github.com/wicanr2/great-era-remake/blob/fae384579b8f63e974f5c0c6f6b701f800ce6b36/docs/design/22-biography-integration.md) |
| docs/design/30-audio-direction.md | [歷史版本](https://github.com/wicanr2/great-era-remake/blob/fae384579b8f63e974f5c0c6f6b701f800ce6b36/docs/design/30-audio-direction.md) |
| docs/design/32-modern-music-direction.md | [歷史版本](https://github.com/wicanr2/great-era-remake/blob/fae384579b8f63e974f5c0c6f6b701f800ce6b36/docs/design/32-modern-music-direction.md) |
| docs/spec/41-modern-satellite-command-center-h4-m1.md | [歷史版本](https://github.com/wicanr2/great-era-remake/blob/fae384579b8f63e974f5c0c6f6b701f800ce6b36/docs/spec/41-modern-satellite-command-center-h4-m1.md) |
| docs/spec/42-modern-command-matrix-h4-m1.md | [歷史版本](https://github.com/wicanr2/great-era-remake/blob/fae384579b8f63e974f5c0c6f6b701f800ce6b36/docs/spec/42-modern-command-matrix-h4-m1.md) |
| docs/spec/43-modern-bright-ochre-a2-m1.md | [歷史版本](https://github.com/wicanr2/great-era-remake/blob/fae384579b8f63e974f5c0c6f6b701f800ce6b36/docs/spec/43-modern-bright-ochre-a2-m1.md) |
| docs/spec/44-modern-cartographic-theme-m2.md | [歷史版本](https://github.com/wicanr2/great-era-remake/blob/fae384579b8f63e974f5c0c6f6b701f800ce6b36/docs/spec/44-modern-cartographic-theme-m2.md) |
| docs/spec/45-republic-wargame-theme-m1.md | [歷史版本](https://github.com/wicanr2/great-era-remake/blob/fae384579b8f63e974f5c0c6f6b701f800ce6b36/docs/spec/45-republic-wargame-theme-m1.md) |
| docs/spec/47-original-homage-theme-m1.md | [歷史版本](https://github.com/wicanr2/great-era-remake/blob/fae384579b8f63e974f5c0c6f6b701f800ce6b36/docs/spec/47-original-homage-theme-m1.md) |
| docs/spec/48-ink-wash-theme-m1.md | [歷史版本](https://github.com/wicanr2/great-era-remake/blob/fae384579b8f63e974f5c0c6f6b701f800ce6b36/docs/spec/48-ink-wash-theme-m1.md) |
| docs/spec/49-photo-overprint-theme-m1.md | [歷史版本](https://github.com/wicanr2/great-era-remake/blob/fae384579b8f63e974f5c0c6f6b701f800ce6b36/docs/spec/49-photo-overprint-theme-m1.md) |
| docs/spec/46-biography-full-coverage-r1.md | [歷史版本](https://github.com/wicanr2/great-era-remake/blob/fae384579b8f63e974f5c0c6f6b701f800ce6b36/docs/spec/46-biography-full-coverage-r1.md) |

獨有原始研究、事實表、音訊及圖集來源、正常操作收據和已發布交付物保留。舊工作樹的必要修正已在 `fae3845` 保存；本輪將根工作樹統一至 master，臨時主線副本移除。刪除清單、引用、雜湊與驗證由 [Issue #11](https://github.com/wicanr2/great-era-remake/issues/11)追蹤。

本輪移除 16 份已失效文件；完整內容均可由上述既有提交還原。另刪除 18 份本機 Issue 草稿，其中 16 份逐字核對遠端原文，兩份逐字核對既有 Issue 全文快照；六份獨有歷史草稿保留。原預覽首頁的獨有歷程移入既有 WORKLOG，首頁只保留入口。

Docker 驗證 87 個文件相對引用零失效，1,126 份受保護程式、資料、授權及史料輸入逐檔雜湊一致；人物正文與 provenance 檢查、純 Go 閘門及兩份 Python 工具語法樹等價檢查通過。Android 的有效畫布、輸入、平台與驗收契約保留。素材拒絕掃描及推送後核對由 Issue #11 記錄。遊戲執行期未改，不重建或覆寫已發布封包。

根目錄原有四筆未提交內容的必要修正已在 `fae3845` 保存；完整原稿另存本機 Git stash，再將根目錄切回 master。兩份正式版來源工作樹只作凍結驗證輸入，不維護現況。

根工作樹已切回 `master` 並推送 `8c47034`，臨時主線工作樹已移除。四份原稿在本機 stash `0f5db74c8f426caf2fe43c3821ce7d4f47557766` 逐檔核對原雜湊，沒有丟失內容。GitHub 分支查驗僅剩 `master`；1,392 檔素材拒絕掃描通過，11 張展示截圖位置合規。總 Issue 改為後續工作的索引，不再複製現況表及歷史清單。

## 2026-10-15 README 遊戲介紹、對照截圖與推廣片重拍

依使用者要求重寫首頁，介紹原作背景、三個歷史時期、軍政經營、六角戰場及 HD 介面、美術與音樂。README 移除工程紀錄、工具台帳與驗收流水帳，只保留玩家需要的下載、匯入、操作及授權入口。三個劇本日期由既有 Scenarios 核對，不新增原作情節。

由已發布 Linux 公開引擎入口、固定種子 1 與正常鍵鼠操作重新擷取六張原版／HD 畫面；DOSBox-X 從唯讀原始資料的副本擷取一張輔助展示。七張 PNG 替換十一張舊展示圖，圖片來源、雜湊及 Git 歷史回查見 SPEC-48，早期規格引用同步修正。人物介面切換會關閉指揮室視窗，原版自傳改由正常查閱選單與 B 鍵補錄；初次未顯示自傳的片段不交付。

重新錄製 DOS、軍政、戰術與人物閱讀原片，剪為 92.5 秒比較片。原版／HD 的地圖、自傳與戰場並列，另展示正常縮放、拖曳、徵稅、部隊移動及撤退。配樂取自已採用的原創 Ogg。舊封包、影片與已發布 manifest 逐檔不變，新片與獨立影音 manifest 放在同版本 promo，更新本機交付頁。

初剪包含未翻頁的單頁自傳，產生重複畫面，已移除此段。第二次校驗誤把即時 DOS 截圖當作編碼時間的同一狀態；核對原片後改為按精確剪輯時間解碼比對。兩項均為剪輯及驗證腳本問題，保留嘗試，不列為遊戲缺陷。最終完整解碼、音訊、黑幀、八秒以上凍結、字幕邊界與十七章來源比對通過；抽幀逐章視覺檢查通過。沒有宣稱新的人耳試聽或實機驗收。

GitHub 官方 Markdown 渲染及隔離 Chromium 的桌面／手機檢查載入全部七張圖片，無水平溢出，對照入口可定位。來源程式、規則、音訊、人物資料與原始遊戲輸入未改，不重跑相同 Go 回歸或重建封包。推送、拒絕掃描、工作樹與容器收尾見 [Issue #12](https://github.com/wicanr2/great-era-remake/issues/12)。

## 2026-10-15 地圖明暗斷層與開發插圖

使用者指出全國地圖有局部色差，開發頁只有金屬底紋。追查確認地表依固定經緯度硬切配色，方向光已停用，仍會留下矩形邊界。先補 READY 契約及會失敗的回歸，再以連續混色修正，保留省界、勢力色、地形投影及控制。

內建圖像生成工具製作墾地、兵工廠與金礦三幅原創場景。純 Go 工具縮成 384×480 圖集，新增解碼量 737,280 bytes；嵌入 HD 開發卡片並保留原選項、語系及 Action。索引無效或解碼失敗時保留原文字卡。精確白名單、原圖雜湊、提示詞全文及權利由既有生成素材紀錄維護。

19 個正式 Go 套件通過。固定種子 1 的正常鍵鼠操作確認地圖縮放、拖曳、美術及介面切換；三個開發指令與已發布 Linux 引擎的回報和存檔一致。兵工廠樣本正常拒絕黃金不足，沒有修改規則或注入資源。README 更新地圖並補開發頁截圖，標為主線預覽，舊圖的歷史收據改連固定提交。本機啟動入口改用根目錄主線預覽，避免啟動凍結舊工作樹。

驗證管線曾因新增套件不在離線快取、私人草稿被 `go test ./...` 納入，以及回報文字／黃金不足被腳本誤判而停止。縮製工具改用標準函式庫，正式套件以 `./cmd/... ./internal/... ./tools/...` 重跑，GUI 對照按既有固定狀態核對成功與拒絕結果。原始失敗保留於本機，不列為遊戲缺陷。865 份受保護程式與資料、153 份原始輸入雜湊不變；已發布封包、影片及 Release 不改。具體收據與收尾見 [Issue #13](https://github.com/wicanr2/great-era-remake/issues/13)。

## 2026-10-15 v.1.0.2-20261015 重新發行

依使用者要求，從 `master` 的乾淨提交 `87934a41e8447b02080103b6536eb132edf4fd9f` 固定來源，收錄地圖配色修正與開發 HD 插圖。新建三平台完整版，公開包核對同一份來源後重用本次三個二進位；版本、架構、CRC、逐檔雜湊、八曲音訊及第三方告知均通過。音訊只更新版本紀錄，保留八份已採用混音位元組。

19 個 Go 測試套件通過。Linux 與 Wine 各驗完整包及公開包，包含新開發頁、美術及介面切換、人物閱讀、徵稅、正常存讀檔與重啟。四份存檔相同，Go parser 核對省份數值並逐位元組寫回。macOS 限交叉建置與靜態檢查，真機、簽署、公證及 Linux FUSE 仍依既有 Issue 追蹤。

由本版 Linux 引擎包重新錄製重製原版及 HD 畫面，剪為 93.534 秒比較片，加入新版地圖與三幅開發插圖。DOS 段沿用當日原程式輔助原片，保留原收據；配樂沿用已採用原創 Ogg。17 章完整解碼、抽幀來源、字幕邊界、非靜音、黑幀及凍結驗證通過，並核對章節總覽及開發插圖全幅圖。沒有宣稱新的實機或人耳驗收。

包裝器先因執行位元不足、分離工作樹的唯讀掛載點不存在而停止，分別改以 Bash 啟動及補齊使用者擁有的空掛載點後，用相同工具鏈重跑。GitHub 草稿尚未由 tag 端點提供，改從授權列表查找，發布後 tag 端點正常。均為環境或工具管線問題，未改動遊戲規則。

建立全新不可變 tag 與 Release，先上傳草稿核對五份附件的遠端雜湊，再公開並設為最新。三平台完整版、公開包、影片及驗證集中於 `dist-all/v.1.0.2-20261015/`，本機首頁指向新版；242 份前版交付及 154 份原始輸入雜湊不變，後者包含 153 份素材及一份來源 ZIP。具體附件、來源與收尾見 [Issue #14](https://github.com/wicanr2/great-era-remake/issues/14)。

交付頁六個封包入口、影片中繼資料及 16:9 顯示比例通過隔離 Chromium 核對，整頁視覺檢查通過。最後拒絕掃描 1,393 檔零命中、104 個文件相對引用有效；997 份執行期程式及語系不變，新增輸出均由目前使用者擁有。既有 2,124 個 root 擁有路徑未改動，本輪 Docker 工作完成後清理，GitHub 仍只保留 master 分支。

## 2026-10-15 README 故事推廣文與切換說明

依使用者要求在 README 開頭加入故事推廣段落，從北伐與抗戰的三個劇本切入，介紹省份經營、軍需、將領調動與六角戰場。背景及年代依既有 `internal/game/scenario.go` 與玩法介紹，不新增虛構人物、劇情或未實作功能。

補明 F2 美術與 F4 介面可在遊戲中獨立即時切換，不需重啟，也不改動進度。配樂維持獨立設定；重新核對正式啟動器的 HD 配樂預設、八首 Ogg 與既有音訊測試紀錄，沿用使用者已採用的混音，不宣稱新的裝置或人耳驗收。只改文件，不重跑遊戲回歸或重建已發布封包；引用、素材掃描、Git 與 Docker 收尾由 [Issue #15](https://github.com/wicanr2/great-era-remake/issues/15)記錄。

## 2026-10-15 常駐工具列操作原型

依使用者提供的位置，唯讀核對《三國演義》重製版 `internal/ui/windowbar.go`、工具列實錄及操作文件。使用者選擇先保留現有 AI，只做工具列與滑鼠，並明確要求常駐、不收合。可操作原型沿用大地與鋼鐵配色，將美術、介面、現有 AI 與操作四欄放在完整遊戲畫面上方。背景取自既有實際截圖，頁面明示原型，不當作可玩版或正式驗收。

原型及 Chromium 檢查由既有 `workplace/hoi4-ui-prototype/README.md` 索引。桌面與較窄視窗檢查四種美術／介面組合、選單取消、滑鼠模式切回及無水平溢出，視覺檢查完成。初次使用圖片解碼 promise，無頭瀏覽器在等待完成前匯出 DOM；改以圖片載入事件後，用相同工具鏈重跑通過，屬驗證腳本問題。

現有滑鼠、數字鍵盤、美術切換及防拖曳誤觸的針對性 Go 回歸通過。程式查證顯示右鍵返回尚未統一接入，地圖與人物頁部分滾輪路徑直接讀取游標；後續接入常駐工具列時須一起處理畫面位移及選單輸入攔截。版面確認前未改正式遊戲、語系、偏好格式、AI、音樂或已發布產物。工作及後續收尾由 [Issue #16](https://github.com/wicanr2/great-era-remake/issues/16)追蹤。

原型曾以隔離容器瀏覽器在使用者桌面開啟，確認視窗顯示後關閉；本輪容器已清理。三份文件入口引用、原型工具語法及素材拒絕掃描通過，掃描 1,393 個受版控檔案、八張合規展示圖。新增原型均屬目前使用者擁有。根工作樹僅有 CONTEXT 與 WORKLOG 的現況紀錄尚未提交；原型留於既有本機工作區，沒有提交、推送或重新發行。

## 2026-10-15 常駐工具列接回可玩版

使用者確認原型後，保存定稿畫面並先建立 READY 的 SPEC-50。工具列以四欄常駐於完整遊戲畫面上方，原版與高解析皆保留原遊戲區；美術與介面沿用既有 setter，AI 只有現況說明。輸入新增 `keyboard_only` 偏好，缺欄沿用鍵盤加滑鼠；寫入失敗不部分套用，純鍵盤仍可用工具列切回。

統一滑鼠與觸控的工具列位移，地圖拖曳與滾輪也走同一座標入口。工具列攔截整次點擊與焦點，外部點擊只關閉選單，不觸發下方徵稅等指令。右鍵沿既有 Action 返回或取消，人物與新聞可用滾輪翻頁；F6 提供鍵盤工具列操作，Esc 與 F10 沿既有返回、確認及自動存檔契約。

新標籤先揭露字模缺字，沿用相同 Noto 來源與重生工具補齊六字，兩包共增加 12,384 bytes，全部舊字模位元組相同。工具列始終使用內嵌 Unicode 黑體，語系檢查依實際 renderer 分工核對，舊標籤的倚天要求保留。英文 Interface 實寬 50 px，改為依字寬配置，三語系完整單行測試通過。

最終 19 個 Go 套件通過。固定 seed 1 的 40 張正常入口實錄涵蓋四種切換、防穿透、純鍵盤切回、原版解析度、右鍵與滾輪、開發、徵稅、戰術移動、滑鼠撤退數字輸入、存讀檔與重啟。工具列操作後與前版存檔逐位元組相同；三份存檔經 Go parser 與原位元組 writer 驗證。徵稅結果與前版相同，英日工具列及保存結果亦通過。來源、雜湊及範圍集中於 SPEC-50 與本機收據。

第二個 GUI 視窗初次在完成繪圖前被擷取為黑幀，改以正確 PID 與可見首幀判定就緒後，用同一程式與容器重跑。macOS 第一次誤用無 cgo 的一般工具映像，依既有 package.sh 改用 hr-osxcross 與 oa64-clang 後成功；作者程式未引入 cgo。收據自我連結再因 Docker 雙重掛載路徑別名誤判，修正後通過。這些是驗證工具設定問題，未改遊戲規則。

最後素材拒絕掃描涵蓋 1,402 檔及十張合規展示圖，純 Go、文件引用與擁有權檢查通過。325 份受保護輸入不變，兩個字模包為明確呈現增量；原始資料、八曲音訊、規則與 AI 實作保持不變。既有 2,124 個 root 擁有路徑未改動，新增輸出皆為目前使用者擁有。本機 play.sh 已切至新版，預設預覽靜音，可加 `-audio modern` 使用已採用配樂。

本次只完成本機可玩預覽，工作樹變更保留待提交，沒有 commit、push、tag 或新版 Release。Windows amd64 與 macOS arm64 只驗交叉建置及格式，真機與簽署仍由既有 Issue 追蹤。完成與容器清理由 [Issue #16](https://github.com/wicanr2/great-era-remake/issues/16)記錄。

## 2026-10-15：即時音樂與零參數啟動

依使用者要求擴充常駐工具列為五欄，增加 HD／原版／靜音，以及自動選曲與八首指定曲目。偏好與遊戲存檔分開，準備或設定寫入失敗保留目前音訊；所有模式共用 Ebiten 音訊裝置，Ogg 循環點按裝置取樣率換算。原創八首 Ogg 位元組沿用已試聽版本，沒有重新作曲或混音。

執行檔從自身位置尋找 Windows、macOS Resources 或 Linux AppDir 資源，存檔沿用平台使用者目錄。明示參數優先，啟動器不強制覆蓋音樂偏好。19 個正式 Go 套件通過，40 張 seed 1 正常操作實錄與存讀檔重啟通過；Linux Pulse 隔離錄音驗證 HD、原版、八首指定曲目、靜音恢復、戰術切換及重啟偏好。收據在 workplace/hoi4-ui-prototype/audio-default-20261015/，工作由 Issue #17 追蹤。首輪偏好測試需補上新預設；音訊錄音等待改依實際三秒樣本，避免合成或 Pulse 恢復串流的環境時序誤判。

字型補入日文音樂選單三字，既有字模位元組一致；版號將使用 v.1.0.3-20261015，既有發行物保留。正式封包、遠端附件、權利掃描與容器清理的最終結果另追加。

封包首輪已完成三平台及 Linux 零參數操作。Wine 拒絕在非目前使用者擁有的 /tmp 直接建立設定目錄，原因為驗證腳本漏先建立 WINEPREFIX。已補上使用者擁有的專用目錄，屬驗證環境修正，未改玩家程式。正式輸出將從含此修正的乾淨提交重建，首輪未發布產物保留於本機工作區。
