# 字模來源與完整授權

GEMF 字模來源為 Noto Sans／Serif CJK TC，face index 3。完整作者與商標見 [第三方告知](THIRD-PARTY-NOTICES.md)，授權全文見 [OFL 1.1](OFL-1.1.txt)，套件原始告知見 [Noto 來源](Noto-CJK-copyright.txt)。

| 字模 | 字數 | 大小 | SHA-256 |
|---|---:|---:|---|
| sans | 2625 | 2709020 | `44632e260518b76ef00f1f0bef20577bbea5c401db2b1038597eb74557574be4` |
| serif | 2625 | 2709020 | `389a35de18b350e2bb346a4a0f598e2c14976d30b4d35fed215f01538c599094` |

原始來源：NotoSansCJK-Regular.ttc SHA-256 b76b0433203017ca80401b2ee0dd69350349871c4b19d504c34dbdd80541690a；NotoSerifCJK-Regular.ttc SHA-256 a04178ec485dffdff7cc0c0c20e1fce9202d7e2160d805e8e44a4c8841c58481。

重生工具為 `tools/gen_modern_font.py`，須在含 Pillow 與 Noto CJK 的 Docker 工具鏈中執行。執行期以純 Go 解碼內嵌字模。倚天字庫由玩家自行提供，不包含在公開引擎包。

[常駐工具列](../spec/50-permanent-toolbar.md)新增 `便内変断鼠＋` 六個字模，兩包共增加 12,384 bytes。來源字型、face index 及既有 2,554 字的位元組保持不變；本機核對收據位於 `workplace/hoi4-ui-prototype/toolbar-20261015/font-coverage.json`。

[SPEC-51](../spec/51-live-music-startup.md)補入日文音樂選單的 `々択楽` 三個字模，兩包共增加 6,192 bytes；既有 2,560 字模位元組不變。核對收據為 `workplace/hoi4-ui-prototype/audio-default-20261015/font-coverage.json`。

[SPEC-54](../spec/54-native-campaign-start.md)的三語開局用語補入 `你劇効単妥実玩装覧閲` 十個字模，兩包共增加 20,640 bytes。全部既有 2,563 字模 bytes 與兩份來源字型雜湊均逐筆核對不變；收據為 `workplace/hoi4-ui-prototype/expansion-foundation-20261016/native-font-coverage.json`。

[SPEC-55](../spec/55-armament-research.md)補入軍備用語 30 字，兩包共增加 61,920 bytes；全部既有 2,573 字模 bytes 保留。文案移除「妥」後仍保留該字元，重生時使用 `tools/gen_modern_font.py --retain-atlas <既有 GEMF v2>` 合併舊字元集合，再逐筆核對來源字型與既有字模。驗證入口為 `workplace/hoi4-ui-prototype/research-native-20261016/font_check.py`，收據為同目錄 `native-font-coverage.json`。

裝甲改編與 [SPEC-56](../spec/56-battle-information.md)再補入 `←〔〕両体却尺御数気薬` 十一字，兩包增加 22,704 bytes；既有 2,603 字模 bytes 保留，來源字型不變。最新收據沿用上述軍備工作區。

[SPEC-58](../spec/58-modern-interaction.md)的三語 Modern 表單補入八字，兩包增加 16,512 bytes。全部既有 2,614 字模逐筆核對不變，來源與 face index 保留；工具與收據為 `workplace/hoi4-ui-prototype/modern-native-20261016/font.py`、`font-report.json`。

[SPEC-60](../spec/60-original-hd-music.md)的音樂選單補入三字，兩包增加 6,192 bytes。全部既有 2,622 字模 bytes 保留；沿用同一重生工具及字型，收據為 `workplace/hoi4-ui-prototype/release-v.1.1.0-20261017/font-report.json`。
