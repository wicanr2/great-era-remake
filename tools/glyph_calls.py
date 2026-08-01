#!/usr/bin/env python3
"""從 `WAR.EXE` 的反組譯掃出「畫字」呼叫點，把詞條索引還原成中文。

`docs/re/24` 解出四支繪字函式，差別只在一次畫幾個字模：

    sub_5D670  1 字   sub_5D695  2 字   sub_5D6BE  3 字   sub_5D710  5 字

每一支的最後一個參數是**詞條索引（1-based）**。索引配上已還原的詞表
（`docs/formats/glyph-tables/*.json`）就直接讀得出畫面上的中文。

只解得了立即數索引——`mov ax, N / push ax` 或 `push N`。索引來自變數的
（多半是資料表欄位，如將領姓名、省名）解不了，照實記進 `dynamic` 不假裝。

用法：
    tools/py.sh tools/glyph_calls.py              全部呼叫點，附候選中文
    tools/py.sh tools/glyph_calls.py sub_30003    只看某個函式裡的
"""

from __future__ import annotations

import json
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
ASM = ROOT / "workplace" / "ida" / "WAR.EXE.asm"
TABLE_DIR = ROOT / "docs" / "formats" / "glyph-tables"

# 繪字函式 → 一次畫幾個字模。字數就是詞條的槽寬，所以直接決定去查哪些詞表。
DRAW = {
    "sub_5D670": 1,
    "sub_5D695": 2,
    "sub_5D6BE": 3,
    "sub_5D710": 5,
}

# 槽寬 → 候選詞表。同一個槽寬有多個檔案，靠內容通不通順人工挑。
TABLES = {
    1: ["1.15"],
    2: ["2.15"],
    3: ["3.15", "MAN115", "MAN215", "MAN315"],
    4: ["4.15"],
    5: ["FAN(1).15", "FAN(2).15", "FAN(3).15", "JAPAN1.15", "JAPAN2.15"],
}

PUSH_IMM = re.compile(r"^\s+push\s+(?:small\s+)?([0-9A-F]+h?|\d+)\s*$", re.I)
MOV_AX = re.compile(r"^\s+mov\s+ax,\s*([0-9A-F]+h?|\d+)\s*(?:;.*)?$", re.I)
PUSH_AX = re.compile(r"^\s+push\s+ax\s*(?:;.*)?$", re.I)
CALL = re.compile(r"^\s+call\s+(?:near ptr |far ptr )?(sub_[0-9A-F]+)")
PROC = re.compile(r"^(\w+)\s+proc\b")


def num(s: str) -> int:
    s = s.strip()
    return int(s[:-1], 16) if s.lower().endswith("h") else int(s)


def load_tables() -> dict[str, list[str | None]]:
    out = {}
    for p in TABLE_DIR.glob("*.json"):
        if p.stem.startswith("_"):
            continue
        out[p.stem] = json.loads(p.read_text(encoding="utf-8"))["chars"]
    return out


def entry(chars: list[str | None], idx: int, width: int) -> str | None:
    """詞條索引是 1-based（`docs/re/24` §4 用民國 17 年那句驗過）。"""
    k = idx - 1
    if k < 0 or (k + 1) * width > len(chars):
        return None
    return "".join(c if c else "　" for c in chars[k * width : (k + 1) * width])


def main() -> int:
    if not ASM.exists():
        print(f"找不到 {ASM}——先跑 tools/ida.sh 產 .asm", file=sys.stderr)
        return 1
    want = sys.argv[1] if len(sys.argv) > 1 else None
    tables = load_tables()
    lines = ASM.read_text(encoding="latin-1").splitlines()

    cur = "?"
    hits, dynamic = [], 0
    for i, line in enumerate(lines):
        m = PROC.match(line)
        if m:
            cur = m.group(1)
        c = CALL.match(line)
        if not c or c.group(1) not in DRAW:
            continue
        if want and cur != want:
            continue
        width = DRAW[c.group(1)]

        # 最後一個參數（索引）是最靠近 call 的那個 push。往回找它的來源：
        # 直接 `push N`，或 `mov ax, N` 後接 `push ax`。中間夾 `push cs` 要略過。
        idx = None
        for j in range(i - 1, max(0, i - 6), -1):
            s = lines[j]
            if re.match(r"^\s+push\s+cs\s*$", s, re.I):
                continue
            p = PUSH_IMM.match(s)
            if p:
                idx = num(p.group(1))
                break
            if PUSH_AX.match(s):
                for k in range(j - 1, max(0, j - 3), -1):
                    mm = MOV_AX.match(lines[k])
                    if mm:
                        idx = num(mm.group(1))
                    break
            break
        if idx is None:
            dynamic += 1
            continue

        cands = []
        for t in TABLES[width]:
            if t not in tables:
                continue
            e = entry(tables[t], idx, width)
            if e:
                cands.append(f"{t}:{e}")
        hits.append((cur, c.group(1), width, idx, cands))

    for fn, draw, width, idx, cands in hits:
        txt = "  ".join(cands) if cands else "(索引超出所有 %d 字詞表)" % width
        print(f"{fn:<12} {draw} {width}字 #{idx:<4} {txt}")
    print(
        f"\n共 {len(hits)} 個立即數索引、{dynamic} 個索引來自變數（解不了，需個別追）",
        file=sys.stderr,
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
