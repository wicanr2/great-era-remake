#!/usr/bin/env python3
"""列出某支函式對「記錄欄位」的寫入，用來快速判斷它到底改了什麼。

`docs/re/22` 的教訓：欄位掃描的過濾器**自己會有洞**，而洞的症狀是
「零命中」——看起來像結論，其實是沒抓到。所以這支工具：

1. 寫入形式涵蓋 `mov` / `add` / `sub` / `or` / `and`、立即數與暫存器、
   有沒有 `byte ptr` / `word ptr`、有沒有 `es:` 前綴。
2. **內建正對照**（`--self-test`）：拿 `sub_3412B`（慰勞軍民，`docs/re/22`
   已解出它寫將領 `+1`／`+29`／`+30` 與省份 `+19`）驗過濾器抓不抓得到。
   抓不到就直接報錯，不讓你拿壞掉的過濾器去下「零命中」的結論。

記錄是哪一張表，靠函式裡的 stride 判斷：`mul 21h` = 將領（33 B）、
`mul 25h` = 省份（37 B）。兩個都有就兩個都列，**不猜**。

用法：
    tools/py.sh tools/field_writes.py --self-test
    tools/py.sh tools/field_writes.py sub_16047 sub_17236
"""

from __future__ import annotations

import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
ASM = ROOT / "workplace" / "ida" / "WAR.EXE.asm"

GeneralStride, ProvinceStride = 0x21, 0x25

# 寫入型指令：目的運算元是記憶體。`cmp` 與 `test` 不算（那是讀）。
WRITE = re.compile(
    r"^\s+(mov|add|sub|or|and|xor|inc|dec)\s+"
    r"(?:(?:byte|word|dword)\s+ptr\s+)?"
    r"(?:(es|ss|ds):)?\[(?:di|si|bx)([+-][0-9A-Fa-f]+h?)?\]\s*(?:,|$)"
)
STRIDE = re.compile(r"mov\s+(?:dx|cx),\s*([0-9A-Fa-f]+)h\s*$|mul\s+([0-9A-Fa-f]+)h")
MULIMM = re.compile(r"mov\s+ax,\s*([0-9A-Fa-f]+)h\s*;?.*\n\s+mul\s+")


def body(lines: list[str], name: str) -> list[str]:
    out, on = [], False
    for l in lines:
        if re.match(rf"^{name}\s+proc", l):
            on = True
        if on:
            out.append(l.rstrip())
            if re.match(rf"^{name}\s+endp", l):
                break
    return out


def num(s: str | None) -> int:
    if not s:
        return 0
    s = s.strip()
    neg = s.startswith("-")
    s = s.lstrip("+-")
    v = int(s[:-1], 16) if s.lower().endswith("h") else int(s, 16)
    return -v if neg else v


def strides(b: list[str]) -> set[int]:
    """函式裡出現過哪些記錄大小——決定它在走哪一張表。"""
    found = set()
    text = "\n".join(b)
    for m in re.finditer(r"mov\s+ax,\s*([0-9A-Fa-f]+)h[^\n]*\n\s+mul\s", text):
        found.add(int(m.group(1), 16))
    for m in re.finditer(r"mul\s+(?:dx|cx)", text):
        pass
    for m in re.finditer(r"mov\s+(?:dx|cx),\s*([0-9A-Fa-f]+)h[^\n]*\n(?:[^\n]*\n){0,2}?\s*mul\s", text):
        found.add(int(m.group(1), 16))
    return found


def writes(b: list[str]) -> list[tuple[int, str]]:
    out = []
    for l in b:
        m = WRITE.match(l)
        if not m:
            continue
        # bp 相對是區域變數，不是記錄欄位。
        if "[bp" in l:
            continue
        out.append((num(m.group(3)), l.strip()))
    return out


def describe(name: str, lines: list[str]) -> str:
    b = body(lines, name)
    if not b:
        return f"{name}: 找不到這支函式"
    st = strides(b)
    tables = []
    if GeneralStride in st:
        tables.append("將領(33B)")
    if ProvinceStride in st:
        tables.append("省份(37B)")
    w = writes(b)
    head = f"{name} ({len(b)} 行) 走訪 {'／'.join(tables) or '未偵測到記錄 stride'}"
    if not w:
        return head + "\n    (沒有欄位寫入)"
    seen = {}
    for off, src in w:
        seen.setdefault(off, []).append(src)
    lines_out = [head]
    for off in sorted(seen):
        lines_out.append(f"    +{off:<3} ({len(seen[off])} 處)  {seen[off][0]}")
    return "\n".join(lines_out)


def self_test(lines: list[str]) -> int:
    """正對照：`sub_3412B` 必須抓到將領 +1／+29／+30 與省份 +19。"""
    b = body(lines, "sub_3412B")
    offs = {off for off, _ in writes(b)}
    need = {0x01, 0x1D, 0x1E, 0x13}
    missing = need - offs
    print(f"正對照 sub_3412B（慰勞軍民）抓到的偏移：{sorted(offs)}")
    if missing:
        print(f"❌ 過濾器有洞——漏掉 {sorted(missing)}。**不要拿它去下零命中的結論**",
              file=sys.stderr)
        return 1
    print("✅ 過濾器抓得到已知寫入，可以信任零命中")
    return 0


def main() -> int:
    if not ASM.exists():
        print(f"找不到 {ASM}", file=sys.stderr)
        return 1
    lines = ASM.read_text(encoding="latin-1").splitlines()
    args = sys.argv[1:]
    if not args or args[0] == "--self-test":
        return self_test(lines)
    if self_test(lines) != 0:
        return 1
    print()
    for name in args:
        print(describe(name, lines))
        print()
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
