"""把 51 個 `.15` 字模檔還原出來的詞條，生成 remake 用的**繁中語系表**。

`CLAUDE.md` §6：原版那套「每個場景帶一份字模子集」的機制到 remake 就結束，
重寫版用真正的字串表 + 完整字型。這支就是那個轉換：

    docs/formats/glyph-tables/*.json   ← 反查結果（chars 陣列，null = 空白格）
        ↓
    translations/zh-Hant/glyphtext.json  ← 語系表（每個檔一組詞條）

每個檔的槽寬 `w` 固定（`docs/formats/01-glyph-text.md` §3），
所以 `chars` 每 `w` 個一組就是一個詞條。輸出兩種形式：

    text   去掉空白格 → 給 remake 用（「西藏」）
    raw    保留空白格（全形空格）→ 給還原原版排版用（「西　藏」）

⚠️ 反查的可信度：`tools/dup_glyph.py` 證實倚天字庫**沒有重複字形**，
所以 byte-exact 反查是單射的（`docs/formats/01-glyph-text.md` §5b）。
51 檔 6,174 個字模裡只有 **1 個** unmatched（`1.15` 的逗號，已知例外），
其餘 1,374 個空白全是排版填充。

用法：tools/py.sh tools/gen_locale.py
"""

from __future__ import annotations

import json
from pathlib import Path

TABLES = Path("docs/formats/glyph-tables")
OUT = Path("translations/zh-Hant/glyphtext.json")

# `1.15` 第 12 格是全檔唯一的 unmatched：倚天「，」左移 3px 下移 2 列的版本
# （`docs/formats/01-glyph-text.md` §5）。人工判讀，不讓它變成 null。
KNOWN_UNMATCHED = {("1.15", 12): "，"}

IDEOGRAPHIC_SPACE = "　"


def main() -> int:
    summary = json.loads((TABLES / "_summary.json").read_text("utf-8"))
    files = {}
    total_entries = 0
    for e in summary:
        name = e["file"]
        path = TABLES / f"{name}.json"
        if not path.exists():
            continue
        data = json.loads(path.read_text("utf-8"))
        chars = data["chars"]
        # 槽寬：`_summary.json` 的 text 是把 chars 逐格接起來的，
        # 所以 len(text) == glyphs，而詞條數 = glyphs / w。
        # w 直接從 `01-glyph-text.md` 的表推不出來，改用檔案自己的欄位。
        w = data.get("slot_width") or e.get("slot_width") or slot_width(name)
        # ⚠️ 槽寬猜錯不會報錯，只會讓最後一組殘缺、整份詞條全部錯位。
        # 所以**驗整除**——第一版把 FAN/JAPAN 猜成 3/4（正確是 5），
        # 235 ÷ 3 不整除卻照樣輸出了 78 條半的東西。
        if len(chars) % w:
            raise SystemExit(
                f"{name}: {len(chars)} 個字模不被槽寬 {w} 整除——槽寬錯了")
        entries = []
        for k in range(0, len(chars), w):
            slot = chars[k : k + w]
            fixed = [
                KNOWN_UNMATCHED.get((name, k + i), c) for i, c in enumerate(slot)
            ]
            raw = "".join(c if c else IDEOGRAPHIC_SPACE for c in fixed)
            entries.append({"text": "".join(c for c in fixed if c), "raw": raw})
        files[name] = {"slot_width": w, "entries": entries}
        total_entries += len(entries)

    OUT.parent.mkdir(parents=True, exist_ok=True)
    OUT.write_text(
        json.dumps(
            {
                "language": "zh-Hant",
                "source": "51 個 .15 字模檔，經倚天字庫 byte-exact 反查",
                "note": "text 去掉排版空白格，raw 保留（全形空格）",
                "files": files,
            },
            ensure_ascii=False,
            indent=1,
        ),
        "utf-8",
    )
    print(f"寫入 {OUT}：{len(files)} 檔、{total_entries} 條詞條")
    for probe in ("3.15", "2.15"):
        if probe in files:
            head = [x["text"] for x in files[probe]["entries"][:5]]
            print(f"  {probe} 前五條: {head}")
    return 0


def slot_width(name: str) -> int:
    """槽寬。**照 `docs/formats/01-glyph-text.md` §4 的表**，不猜。

    那張表分兩級證據：46 檔是 `auto`（由空白的排版模式推出，符合率 100%），
    5 檔是 `content`（檔案沒有空白，改由詞條語意判定，證據較弱）。
    兩級都照抄，因為那是目前唯一的來源。
    """
    if name.startswith("TN15"):
        return 5
    if name.startswith("MAN"):
        return 3
    if name.startswith("FAN"):
        return 5
    if name.startswith("JAPAN"):
        return 5
    fixed = {"1.15": 1, "2.15": 2, "3.15": 3, "4.15": 4}
    if name in fixed:
        return fixed[name]
    raise SystemExit(f"{name}: `01-glyph-text.md` §4 沒有這個檔的槽寬")


if __name__ == "__main__":
    raise SystemExit(main())
