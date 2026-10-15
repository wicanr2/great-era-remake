# 工作紀錄

本檔保存主要分支整併後的工作歷程。現況見 [CONTEXT.md](CONTEXT.md)，後續工作由 [GitHub 總 Issue](https://github.com/wicanr2/great-era-remake/issues/1)追蹤；整併前的研究紀錄與日期化文件保留在原位置及 Git 歷史。

## 2026-10-15 整併正式版至 master

使用者要求合併分支並整理 GitHub 專案。遠端預設分支為 `master`，原提交 `ad1c42e298a4cd76d80ddf56dda138f48a90d1af`。以獨立工作樹將 `release/v.1.0.1-20261014` 快轉整併至 `85ed56095117134fd3fdbf64efd7a5411eaeabf7`；其中已包含 `exp/dosgolem-isr` 及前版發行的全部提交，無衝突，也沒有未合併 PR。

`dosgolem/` 隨既有提交 `d612b18` 保留研究工具源碼快照，來源為 dosgolem 專案提交 `d9c0c27`。本輪沒有擴充模擬器或重開原版研究。

README 整理為正式下載、現行畫面、操作、驗證限制及文件導航。CONTEXT 的開頭改為目前狀態，原有長篇研究內容明示為歷史。發行文件補上正式版與驗證附件連結，SPEC-48 追加已發布版本的 CONFORMED 收據；舊 tag 內容保持原狀。

新增兩張既有正常玩家路徑截圖至 `docs/images/` 展示白名單，來源是 v.1.0.1-20261014 的 Linux 指揮介面與正常出兵戰術畫面，種子皆為 1；原尺寸 1280×720，沒有加工或重新生成素材。來源、收據與 SHA-256 見 [SPEC-48](docs/spec/48-hd-release-integration.md#主要分支整併與展示收據)。

驗證比對正式版全部 1,405 份檔案；1,074 份遊戲、工具、研究工具快照、資料及授權輸入逐檔不變。只修改五份既有文件，新增本工作紀錄及兩張展示圖。52 個相對連結、截圖尺寸／來源、純 Go 閘門及 1,408 檔拒絕掃描通過，11 張展示圖位置合規。遊戲執行期未變，沿用同來源正式版已通過的 19 套件回歸收據，不重跑相同測試。

[Issue #9](https://github.com/wicanr2/great-era-remake/issues/9)保存本輪驗證摘要、推送後的 GitHub 核對與容器收尾；[總 Issue #1](https://github.com/wicanr2/great-era-remake/issues/1)仍為後續工作入口。原工作樹的四筆未提交變更，以及兩版正式 tag、Release、封包與發行分支保留。本輪 Docker 使用非 root、無網路、有資源限制的一次性容器，檔案擁有權檢查零異常。
