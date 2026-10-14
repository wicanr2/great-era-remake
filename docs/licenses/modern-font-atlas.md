# 字模來源與完整授權

GEMF 字模來源為 Noto Sans／Serif CJK TC，face index 3。完整作者與商標見 [第三方告知](THIRD-PARTY-NOTICES.md)，授權全文見 [OFL 1.1](OFL-1.1.txt)，套件原始告知見 [Noto 來源](Noto-CJK-copyright.txt)。

| 字模 | 字數 | 大小 | SHA-256 |
|---|---:|---:|---|
| sans | 2554 | 2635748 | `38689a918af30fa89e891ae1af236d667f0c60950e9a019d3699b7342bde3f99` |
| serif | 2554 | 2635748 | `0e2daea86b2db3c260193ba5a8756b01e6bd67c6f98b471f81c272ff460ddf94` |

原始來源：NotoSansCJK-Regular.ttc SHA-256 b76b0433203017ca80401b2ee0dd69350349871c4b19d504c34dbdd80541690a；NotoSerifCJK-Regular.ttc SHA-256 a04178ec485dffdff7cc0c0c20e1fce9202d7e2160d805e8e44a4c8841c58481。

重生工具為 `tools/gen_modern_font.py`，須在含 Pillow 與 Noto CJK 的 Docker 工具鏈中執行。執行期以純 Go 解碼內嵌字模。倚天字庫由玩家自行提供，不包含在公開引擎包。
