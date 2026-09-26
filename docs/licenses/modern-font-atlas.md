# Modern 字型 atlas 授權與重生紀錄

> 狀態：**技術資產已產生；正式發行仍需逐平台 license／商標／商店檢查**  
> 日期：2026-08-11

Modern 高解析度 renderer 使用 `internal/assets/data/modern-font-sans.bin` 與
`internal/assets/data/modern-font-serif.bin`。兩者分別由可再散布的 Noto Sans／Serif
CJK Traditional Chinese rasterize 成專案自己的 8-bit alpha Unicode atlas；儲存庫
不包含 Noto 原始字型檔，也不包含原版倚天字型。

## 來源與命令

- 黑體來源：`NotoSansCJK-Regular.ttc`，font face index `3`（Noto Sans CJK TC）；
  SHA-256 `b76b0433203017ca80401b2ee0dd69350349871c4b19d504c34dbdd80541690a`。
- 宋體來源：`NotoSerifCJK-Regular.ttc`，font face index `3`（Noto Serif CJK TC）；
  SHA-256 `93069d8e9e45d515cc421c971a79e6a5777704b348e36a9ef86578bf58adef77`。
- 來源授權：SIL Open Font License 1.1；Debian 套件 metadata 標為 `SIL-1.1`，
  上游來源為 <https://github.com/notofonts/noto-cjk>。發行包仍須隨包提供完整
  OFL notice／copyright，不能只依賴本段摘要。
- 重生工具：`tools/gen_modern_font.py`，只讀 `translations/**/*.json` 與固定
  Modern UI 種子字串，不讀 `workplace/` 或任何原版素材。
- Docker 工具映像：`mm-font-tools:dev`；輸出由目前使用者 UID/GID 產生。
- 命令：

  ```sh
  python3 tools/gen_modern_font.py \
    --root /work \
    --font /usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc \
    --font-index 3 \
    --family sans \
    --output /work/internal/assets/data/modern-font-sans.bin

  python3 tools/gen_modern_font.py \
    --root /work \
    --font /usr/share/fonts/opentype/noto/NotoSerifCJK-Regular.ttc \
    --font-index 3 \
    --family serif \
    --output /work/internal/assets/data/modern-font-serif.bin
  ```

## 輸出契約

- GEMF v2，2,392 個 Unicode glyph，32×32 8-bit alpha cell、比例字距；兩份輸出皆為
  2,468,564 bytes。
- 黑體 atlas SHA-256：`ac8a9246d4c165fe21186023f2fa6629ebc143a9e864324745802de0668503c3`。
- 宋體 atlas SHA-256：`8bffb8a35914ca72d2d6c2534656a1588147781c690f23ee1a1a00dd1dbaa6af`。
- 使用者確認的 C 分工是：畫面主標題使用宋體；按鈕、內文與數值使用黑體。純 Go
  renderer 在不透明畫布上合成 alpha coverage；不在執行期載入 FreeType，也不引入 `cgo`。
- `ParseModernFont` 同時保留 GEMF v1 相容解析；`EmbeddedModernFontFamily` 會驗證 magic、
  版本、cell、長度、code point 排序與字距。舊 `modern-font.bin`（95,660 bytes，SHA-256
  `d0d6916622090df79e657eaba43255dcc0cbe5dcf329faa71da372c60e63717c`）只作 v2 缺失／損壞時
  的 16×16、1-bit 可讀性 fallback，不是高解析驗收資產。復古 renderer 不讀這些 atlas。
- 英／日人物 387 篇 machine-draft overlay 的實際 `PeopleDB` coverage 已有
  `internal/assets/modernfont_test.go`；缺字不會靜默吞掉，而由 renderer 的
  `Missing`／fail-closed 路徑回報。

## 尚未關閉的 gate

atlas 的技術覆蓋不等於正式內容完成：387 篇英／日 machine-draft 仍需逐篇人審、
日文禁則與長文截圖／裝置 QA；正式發行時必須補入 OFL notice、作者／商標聲明與
各平台包的 license 檢查。這些 gate 不改變原版 oracle／parity 狀態。
