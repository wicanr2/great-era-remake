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
