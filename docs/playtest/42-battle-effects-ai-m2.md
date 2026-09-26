# 戰鬥效果／撤退／AI chain M2 驗收

日期：2026-08-11

## 已驗證

- `internal/game`：正規副作用、協同支援分攤與死亡停止、騎兵衝鋒 pass／暫時格位、
  炮兵三次戰損／反擊抑制、特殊視覺分支、人物欄位同步與部署零值正規化。
- `cmd/dsds`：Xvfb 下戰鬥 UI、撤退目的地輸入、AI 守方回合、autosave 與語系 overlay。
- `internal/i18n`：英／日各 387 筆 `machine-draft` overlay 的來源／譯文 SHA-256、
  ID 順序、姓名一致、runtime fail-closed 套用。

## 證據邊界

- `18←19` 撤退候選是唯一正常原版樣本；其他省份用鄰接順序 fallback，不能稱 parity。
- 六個原版攻擊 handler 的完整選擇、彈藥／動畫／音效時機，及 `.DT2` 469-byte 戰後
  oracle 仍未閉合。
- `AutoResolveByChain` 的 remake handler 已接通且 `Unimplemented==0` 可作場次 gate；
  `sub_3A9F4`／`sub_3AABA` 細分、預約表生命週期、方位映射與正常玩家 AI oracle
  未證實。YouTube 未找到可重播本作 AI 的相關影片，未納入外部 parity 證據。

## 重現命令（Docker）

```text
tools/py.sh tools/gen_biography_overlays.py
tools/py.sh tools/locale_bio_gate.py --locale en --overlay translations/en/people-biography-overlay.json
tools/py.sh tools/locale_bio_gate.py --locale ja --overlay translations/ja/people-biography-overlay.json
tools/go.sh test -count=1 ./...
```

`tools/go.sh` 的 `cmd/dsds` 測試需在同一 Go image 內先啟動有界 Xvfb；不得把無
`DISPLAY` 的 GLFW 失敗誤記成產品缺陷。
