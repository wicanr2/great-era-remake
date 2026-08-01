"""把戰鬥 AI 決策鏈（`sub_3AB99`）的每一步掃出來。

每一步都是同一個形狀：呼叫 `sub_3A320(mode, Real)` 做戰力比，
成立就設 `byte_6AA85`（分支 A）或 `byte_6AA84`（分支 B）並立「已決定」旗標。
所以只要抽出每支函式的

  * `sub_3A320` 的 mode 與 Real 常數
  * 設進 `byte_6AA85` / `byte_6AA84` 的值
  * 呼叫到的其他判斷函式

就能一次拿到全貌，不必逐支讀 100 多行組語。

用法：tools/py.sh tools/battle_ai_chain.py workplace/ida/WAR.EXE.asm
"""

from __future__ import annotations

import re
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from tp_real import from_regs  # noqa: E402

STEPS_A = ["sub_3A817", "sub_3A885", "sub_3A8C8", "sub_3A8F7", "sub_3A94E",
           "sub_3A988", "sub_3A9C9"]
STEPS_B = ["sub_3A9F4", "sub_3AA51", "sub_3AABA", "sub_3AAF4"]


def body(src: list[str], name: str) -> list[str]:
    out, on = [], False
    for line in src:
        if re.match(rf"^{name}\s+proc", line):
            on = True
        if on:
            out.append(line.rstrip())
            if re.match(rf"^{name}\s+endp", line):
                break
    return out


def scan(src: list[str], name: str) -> dict:
    b = body(src, name)
    info = {"lines": len(b), "reals": [], "sets": [], "calls": []}
    ax = bx = dx = mode = None
    for line in b:
        if m := re.search(r"mov\s+ax,\s+([0-9A-F]+)h", line):
            ax = int(m.group(1), 16)
        elif m := re.search(r"mov\s+bx,\s+([0-9A-F]+)h", line):
            bx = int(m.group(1), 16)
        elif m := re.search(r"mov\s+dx,\s+([0-9A-F]+)h", line):
            dx = int(m.group(1), 16)
        elif m := re.search(r"mov\s+al,\s+(\d+)\s*$", line):
            mode = int(m.group(1))
        if "call    sub_3A320" in line:
            if ax is not None and bx is not None and dx is not None:
                info["reals"].append((mode, round(from_regs(ax, bx, dx), 6)))
            ax = bx = dx = None
        # ⚠️ IDA 用 `12h` 表示 0x12。第一版的正則把 `h` 當可選又沒記下來，
        # 於是 `12h` 被讀成十進位 12（正確是 18）。基數要看有沒有 `h`。
        if m := re.search(r"mov\s+(byte_6AA8[45]),\s+([0-9A-F]+)(h?)", line):
            v, suffix = m.group(2), m.group(3)
            info["sets"].append((m.group(1), int(v, 16 if suffix else 10)))
        if m := re.search(r"call\s+(sub_\w+)", line):
            if m.group(1) != "sub_3A320":
                info["calls"].append(m.group(1))
    return info


def main() -> int:
    src = Path(sys.argv[1]).read_text("latin-1").splitlines()
    for label, steps in (("分支 A（arg_0 == 0）", STEPS_A),
                         ("分支 B（arg_0 != 0）", STEPS_B)):
        print(f"\n=== {label}")
        for n in steps:
            i = scan(src, n)
            print(f"  {n}  {i['lines']:>4} 行")
            if i["reals"]:
                print(f"      戰力比 sub_3A320: {i['reals']}")
            if i["sets"]:
                print(f"      決策值: {i['sets']}")
            others = sorted(set(i["calls"]) - {"sub_5A0B9"})
            if others:
                print(f"      其他呼叫: {', '.join(others)}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
