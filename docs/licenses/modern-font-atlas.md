# 字模來源與完整授權

GEMF 字模來源為 Noto Sans／Serif CJK TC，face index 3。完整作者與商標見 [第三方告知](THIRD-PARTY-NOTICES.md)，授權全文見 [OFL 1.1](OFL-1.1.txt)，套件原始告知見 [Noto 來源](Noto-CJK-copyright.txt)。

| 字模 | 字數 | 大小 | SHA-256 |
|---|---:|---:|---|
| sans | 2563 | 2645036 | `21cb07d073c30d3033c295e0f483fe1357fe7875e6010c7cc78dc868c2d42e68` |
| serif | 2563 | 2645036 | `ebeeb13282a2f760003dcf15ebbdd84decd7ddcfe363c7a3aeaf0a093a9233f5` |

原始來源：NotoSansCJK-Regular.ttc SHA-256 b76b0433203017ca80401b2ee0dd69350349871c4b19d504c34dbdd80541690a；NotoSerifCJK-Regular.ttc SHA-256 a04178ec485dffdff7cc0c0c20e1fce9202d7e2160d805e8e44a4c8841c58481。

重生工具為 `tools/gen_modern_font.py`，須在含 Pillow 與 Noto CJK 的 Docker 工具鏈中執行。執行期以純 Go 解碼內嵌字模。倚天字庫由玩家自行提供，不包含在公開引擎包。

[常駐工具列](../spec/50-permanent-toolbar.md)新增 `便内変断鼠＋` 六個字模，兩包共增加 12,384 bytes。來源字型、face index 及既有 2,554 字的位元組保持不變；本機核對收據位於 `workplace/hoi4-ui-prototype/toolbar-20261015/font-coverage.json`。

[SPEC-51](../spec/51-live-music-startup.md)補入日文音樂選單的 `々択楽` 三個字模，兩包共增加 6,192 bytes；既有 2,560 字模位元組不變。核對收據為 `workplace/hoi4-ui-prototype/audio-default-20261015/font-coverage.json`。
