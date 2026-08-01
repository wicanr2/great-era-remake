#!/usr/bin/env python3
"""Dump 一支函式的組語，**每一行的位址自動附上已知語意**。

    ./tools/py.sh tools/dump_func.py sub_4E627
    ./tools/py.sh tools/dump_func.py sub_3E691 --range 1,80

## 為什麼是這個，而不是「在 IDA 裡命名」

`word_64944` 被推錯兩次。事後檢討時第一個想到的解法是「把已解出的全域變數
在 `.i64` 裡改名，讓 `.asm` 自帶語意」。**那個方案站不住**：

- 當初看不懂的不是 `word_64944`（它有符號），是 `sub_4E627` 裡的
  `mov ax, [di-6221h]`——**`di` 相對偏移沒有符號，IDA 無從命名**
- 命名只覆蓋有符號的全域變數，恰好蓋不到出錯的那一類
- 還要維護命名表、重產 5.7 MB 的 `.asm`

`tools/addr.py` 解決了「查得到」，但沒解決「**想到要查**」——
`word_64944` 那次的失敗模式正是「我以為自己已經知道答案了，沒想到要查」。
而紀律靠不住：`CLAUDE.md` §7.1 白紙黑字寫著先查手上已有的，
這個 session 還是違反了五次。

所以把查詢**併進 dump 流程**：只要用這支 dump 函式，位址語意就自動出現在
眼前，不需要任何額外動作，也不需要記得要做。

## 輸出長什麼樣

    mov     ax, [di-6221h]        ; ← 省份記錄 +20 = 司令（將領 ID）
    mov     word_64944, ax        ; ← 當前交戰省的司令
"""

import re
import subprocess
import sys
import pathlib

ROOT = pathlib.Path(__file__).resolve().parent.parent
ASM = ROOT / "workplace" / "ida" / "WAR.EXE.asm"

sys.path.insert(0, str(ROOT / "tools"))
import addr as A  # noqa: E402  （同目錄的位址反查表）

# 抓組語行裡的位址：[di+7A8Bh]、[di-6221h]、word_64944、byte_6FFC4、62A8h
HEX = re.compile(r"\[(?:[a-z]{2}[+\-])?([0-9A-F]+)h\]|\b(?:word|byte|dword)_([0-9A-F]+)\b")
NEG = re.compile(r"\[[a-z]{2}-([0-9A-F]+)h\]")


def semantics(line: str) -> list[str]:
    """回傳這一行裡所有位址的語意說明（去重、保序）。"""
    out, seen = [], set()

    cands = []
    for m in NEG.finditer(line):
        cands.append(-int(m.group(1), 16))
    for m in HEX.finditer(line):
        raw = m.group(1) or m.group(2)
        if raw is None:
            continue
        v = int(raw, 16)
        # 負偏移已經在上面收過了，這裡跳過同一個位置
        if any(abs(c) == v for c in cands if c < 0):
            continue
        cands.append(v)

    for v in cands:
        for r in A.lookup(v):
            if r["level"] == "未知":
                continue  # 還沒解的不刷版面
            # ⚠️ **推論等級一定要印出來**。使用者質疑「在 IDA 裡命名」的理由
            # 就是名字不帶等級、讀起來像事實；附加式註記若拿掉等級，
            # 就犯同一個錯。confirmed 以外的一律標出來。
            tag = "" if r["level"] == "confirmed" else f"  ⚠️[{r['level']}]"
            t = f"{r['kind']}：{r['desc']}{tag}"
            if t in seen:
                continue
            seen.add(t)
            out.append(t)
    return out


def dump(fn: str, lo: int | None, hi: int | None) -> int:
    if not ASM.exists():
        print(f"找不到 {ASM}。先跑 tools/ida.sh analyze WAR.EXE", file=sys.stderr)
        return 1
    # awk 取函式區段，比在 Python 裡讀 5.7 MB 快
    txt = subprocess.run(
        ["awk", f"/^{fn}[[:space:]]+proc/,/^{fn}[[:space:]]+endp/", str(ASM)],
        capture_output=True, text=True, check=True,
    ).stdout.splitlines()
    if not txt:
        print(f"{fn}: 在 {ASM.name} 裡找不到", file=sys.stderr)
        return 1

    lines = txt[lo - 1: hi] if lo else txt
    hits = 0
    for line in lines:
        sems = semantics(line)
        if sems:
            hits += 1
            print(f"{line}")
            for s in sems:
                print(f"{'':>40}; ← {s}")
        else:
            print(line)

    print(f"\n[dump_func] {fn}：{len(txt)} 行，其中 {hits} 行有已知語意的位址。",
          file=sys.stderr)
    docs_hits(lines)
    return 0


# 組語裡出現的具名符號：word_64944、byte_6FE7E、sub_3A4CE
SYM = re.compile(r"\b((?:word|byte|dword|sub|loc|unk)_[0-9A-F]+)\b")


def docs_hits(lines: list[str]) -> None:
    """把這段組語裡的符號拿去 grep `docs/`，**印出誰已經被寫過**。

    為什麼要這個（2026-08-02）：

    `addr.py` 的三張表是**手維護**的，只涵蓋我想到要登記的位址。
    `byte_6FE7E` 不在表裡，於是 dump 出來一片空白，我就照那片空白
    在 `docs/re/31` §49 寫下「語意未解（疑似戰鬥種類）」——
    而 `docs/re/16` 與 `docs/re/19` 早就寫明它是**月份**。

    這是同一類錯誤的第三次（前兩次是 `word_64944`、比率門檻的來源）。
    三次的共同點都不是「查不到」，是「**沒想到要查**」。
    紀律已經證明靠不住，所以把查詢做成 dump 的一部分：
    只要符號在 `docs/` 裡出現過，這裡就會叫出來，不必記得去查。

    跳過 `loc_*`（純跳轉標籤，噪音）與已經在 `addr.py` 裡的位址
    （那些上面已經逐行標過了）。
    """
    syms, order = set(), []
    for line in lines:
        for m in SYM.finditer(line):
            t = m.group(1)
            if t.startswith("loc_") or t in syms:
                continue
            syms.add(t)
            order.append(t)
    if not syms:
        return

    docs = ROOT / "docs"
    if not docs.is_dir():
        return
    args = ["grep", "-rn", "--include=*.md", "-F"]
    for s in order:
        args += ["-e", s]
    r = subprocess.run(args + ["."], cwd=str(docs), capture_output=True, text=True)

    found: dict[str, list[str]] = {}
    for ln in r.stdout.splitlines():
        for s in order:
            if s in ln:
                found.setdefault(s, []).append(ln.split(":", 1)[0].lstrip("./"))
    if not found:
        print("[dump_func] docs/ 裡沒有任何一個符號被提過——這段大概真的是新的。",
              file=sys.stderr)
        return

    print("\n[dump_func] ⚠️ 下列符號 **`docs/` 裡已經寫過**，"
          "下結論前先去讀，不要重推：", file=sys.stderr)
    for s in order:
        if s not in found:
            continue
        files = sorted(set(found[s]))
        tail = f"（另 {len(files) - 3} 份）" if len(files) > 3 else ""
        print(f"  {s:<14} {'  '.join(files[:3])}{tail}", file=sys.stderr)


def main() -> int:
    argv = sys.argv[1:]
    if not argv or argv[0] in ("-h", "--help"):
        print(__doc__)
        return 0
    fn = argv[0]
    lo = hi = None
    if "--range" in argv:
        rng = argv[argv.index("--range") + 1]
        lo, hi = (int(x) for x in rng.split(","))
    return dump(fn, lo, hi)


if __name__ == "__main__":
    raise SystemExit(main())
