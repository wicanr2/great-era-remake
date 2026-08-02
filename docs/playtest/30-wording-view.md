# 原典／現代白話：查閱流程

日期：2026-08-02  
狀態：**confirmed（指令 6 的唯讀路徑）**

## 範圍

白話化四項入口與容易誤解的省份選擇文字：

| 原典 | 現代白話 |
|---|---|
| 查閱他省 | 查看其他省份 |
| 查閱所屬各省 | 查看所屬省份 |
| 查閱將領 | 查看本省將領 |
| 查閱省名 | 查看省份編號 |
| 查閱他省（輸入頁） | 輸入要查看的省份編號 |
| 概況／查閱將領 | 查看省份資料／查看該省將領 |
| 所屬各省概況 | 所屬省份概況 |
| 查閱省名 | 省份編號對照表 |

省名、將領名、編號、兵力及詳細頁的黃金／糧食／忠誠度等名詞沿用原版資料，
不為了增加改動量而重打；兩種用語模式共用相同候選、分頁與返回狀態機。

## GUI 路徑

第一期 `SAVE(1).DT1`、湖北、固定種子 1：

```text
Enter → 6 → 四項選單
→ 1 → 輸入 19 河南 → 河南的資料／將領選擇 → Esc → Esc
→ 2 所屬省份概況 → Esc
→ 4 省份編號對照表 → Space 第二頁 → Esc
→ F10 → Y
```

原典與白話各截取六張圖：

- `{original,plain}-menu.png`
- `{original,plain}-province-select.png`
- `{original,plain}-province-choice.png`
- `{original,plain}-owned.png`
- `{original,plain}-names-1.png`
- `{original,plain}-names-2.png`

位置：`workplace/screenshots/view-wording/`。十二張均經人工檢查，無缺字、裁切、
錯頁或上一幀殘留。

## 唯讀證據

原始、原典輸出與白話輸出的 SHA-256 均為：

`9ad9359b436ce2450ff3460df4efede478d871dfe234f60299759e781959ad0d`

三檔逐位元組相同，證明查閱與用語切換都沒有修改遊戲狀態。

## 環境

`dsds-go:1.25` 建置；`fd2-dosbox-screenshot-local:latest` 執行 Xvfb、xdotool 與
ImageMagick。兩次皆使用無網路、有限資源、目前 UID/GID 的一次性容器。
