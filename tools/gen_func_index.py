#!/usr/bin/env python3
"""從 docs/ 抽出所有被提到的 `sub_XXXXX`，產生一份查詢用的索引。

為什麼要這個：`CLAUDE.md` §7.1 說「動手挖之前先查手上已有的」，但這個專案的
反組譯筆記已經超過三十份，光靠記憶會一再重讀已經解過的函式。這個 session
就發生過五次——`sub_54DAC`、`sub_3A320`、`sub_55632`、`sub_567B9`、`sub_510E0`
全都是文件裡已有結論、卻又整支重讀了一遍。

用法（一律走 docker，`CLAUDE.md` §9）：

    ./tools/py.sh tools/gen_func_index.py

輸出 `docs/re/00-function-index.md`。**動手讀任何 sub_ 之前先 grep 它。**
"""

import pathlib
import re
import sys

ROOT = pathlib.Path(__file__).resolve().parent.parent
OUT = ROOT / "docs" / "re" / "00-function-index.md"

FUNC = re.compile(r"sub_[0-9A-F]{4,6}")
# 「這一行像不像一句結論」——帶等號、冒號或粗體的行優先當摘要。
VERDICT = re.compile(r"[=＝]|是\*\*|：|\*\*.+\*\*")


def summarise(line: str, name: str) -> str:
    """把一行 markdown 壓成一句可讀的摘要。"""
    s = line.strip().lstrip("#|-> ").strip()
    s = re.sub(r"\s+", " ", s)
    if len(s) > 110:
        s = s[:107] + "…"
    return s


def main() -> int:
    hits: dict[str, list[tuple[str, int, str]]] = {}
    docs = sorted((ROOT / "docs").rglob("*.md"))
    for path in docs:
        if path == OUT:
            continue
        rel = path.relative_to(ROOT).as_posix()
        for n, line in enumerate(path.read_text(encoding="utf-8").splitlines(), 1):
            for name in set(FUNC.findall(line)):
                hits.setdefault(name, []).append((rel, n, line))

    if not hits:
        print("沒有在 docs/ 找到任何 sub_ 提及——路徑對嗎？", file=sys.stderr)
        return 1

    rows = []
    for name in sorted(hits):
        refs = hits[name]
        # 摘要取第一個「像結論」的行；都不像就取第一行。
        best = next((r for r in refs if VERDICT.search(r[2])), refs[0])
        where = sorted({r[0] for r in refs})
        rows.append((name, summarise(best[2], name), best[0], len(refs), where))

    lines = [
        "# 已解函式索引（自動產生，勿手改）",
        "",
        "> `./tools/py.sh tools/gen_func_index.py` 重新產生。",
        "",
        "**動手讀任何 `sub_` 之前先在這裡 grep 它。**",
        "`CLAUDE.md` §7.1「動手挖之前先查手上已有的」——這份索引就是為了讓那一條真的做得到。",
        "",
        f"目前 **{len(rows)}** 個函式在文件裡有紀錄。",
        "",
        "| 函式 | 目前的結論（摘自文件）| 出處 | 提及次數 |",
        "|---|---|---|---:|",
    ]
    for name, summary, src, count, where in rows:
        summary = summary.replace("|", "\\|")
        extra = f"（另 {len(where) - 1} 份）" if len(where) > 1 else ""
        lines.append(f"| `{name}` | {summary} | `{src}`{extra} | {count} |")
    lines.append("")
    OUT.write_text("\n".join(lines), encoding="utf-8")
    print(f"寫出 {OUT.relative_to(ROOT)}：{len(rows)} 個函式")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
