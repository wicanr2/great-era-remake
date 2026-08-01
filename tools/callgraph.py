#!/usr/bin/env python3
"""從 IDA 匯出的 .asm 建呼叫圖，查某支函式的遞移閉包能不能到達目標。

    tools/py.sh scratchpad/callgraph.py WAR.EXE.asm sub_1ACCC sub_2DEFC sub_39B6E

用途：回答「這條路徑會不會走到 X」這種問題時，**讀骨架不夠**——
分支可能藏在任何一層。呼叫圖是完備的（在靜態呼叫的範圍內）。

⚠️ 限制：只認 `call sub_XXXX` 這種直接呼叫。**間接呼叫（函式指標、
跳表）看不到**，所以「零命中」是「沒有靜態路徑」，不等於執行期不可能。

⛔ **更大的限制：資料流看不到。** 「A 不呼叫 B」推不出「A 不會導致 B」——
A 可能只是填了一張表，由第三方讀它去跑 B。`docs/re/29` 就是這樣栽的：
電腦回合子樹對戰鬥主函式零命中，卻被拿來推「電腦不會攻打」，
而實際上它填的是與玩家派將同一張參戰部隊表。
**要問「會不會導致」，得另外掃共用資料結構的讀寫方**。
"""
import re
import sys


def build(path):
    funcs, cur = {}, None
    for line in open(path, encoding="latin-1"):
        m = re.match(r"^(sub_\w+)\s+proc", line)
        if m:
            cur = m.group(1)
            funcs[cur] = set()
        if cur:
            c = re.search(r"call\s+(sub_\w+)", line)
            if c:
                funcs[cur].add(c.group(1))
        if re.match(r"^\w+\s+endp", line):
            cur = None
    return funcs


def closure(funcs, root):
    seen, stack = set(), [root]
    while stack:
        f = stack.pop()
        if f in seen:
            continue
        seen.add(f)
        stack.extend(funcs.get(f, ()))
    return seen


def main():
    if len(sys.argv) < 4:
        print(__doc__)
        return 1
    asm, root, targets = sys.argv[1], sys.argv[2], sys.argv[3:]
    funcs = build(asm)
    if root not in funcs:
        print(f"找不到 {root}")
        return 1
    reach = closure(funcs, root)
    print(f"{root} 的子樹：{len(reach)} 支函式")
    for t in targets:
        mark = "命中" if t in reach else "**零命中**"
        callers = sorted(f for f, cs in funcs.items() if t in cs)
        print(f"  {t:12s} {mark}   直接呼叫者：{callers}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
