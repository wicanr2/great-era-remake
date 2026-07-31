#!/usr/bin/env python3
"""把《大時代的故事》的場景字模檔反查回 Big5，產出每個場景的字表。

還原鏈的第一步（CLAUDE.md §5 步驟 1）。用法：

    tools/py.sh tools/glyph_lookup.py oracle
    tools/py.sh tools/glyph_lookup.py dump TN15.1 --count 8
    tools/py.sh tools/glyph_lookup.py match

`oracle` 必須先過：idx=0 是「一」、「中」(A4A4)、「猴」(B555) 要能認得出來。
這關沒過就整批偏移，症狀是「有字但都不對」。
"""

from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
from eten import (  # noqa: E402
    STRIDE,
    EtenFonts,
    big5_to_font,
    is_blank,
    render,
)

ROOT = Path(__file__).resolve().parent.parent
ETEN_DIR = ROOT / "workplace" / "eten"
GAME_DIR = ROOT / "workplace" / "orig" / "game"
OUT_DIR = ROOT / "docs" / "formats" / "glyph-tables"

# 51 個 .15 族字模檔。命名不統一是原版的事實，不是筆誤：
#   TN15.1–39  場景字模          1.15–4.15  用途待查
#   FAN(1..3)  同上              JAPAN1/2   疑似日文，待驗
#   MAN1/2/315 無副檔名，人物相關（與 MAN(1..3).DAT 成對？）
GLYPH_FILES = (
    [f"TN15.{i}" for i in range(1, 40)]
    + [f"{i}.15" for i in range(1, 5)]
    + [f"FAN({i}).15" for i in range(1, 4)]
    + ["JAPAN1.15", "JAPAN2.15"]
    + ["MAN115", "MAN215", "MAN315"]
)


def cmd_oracle(fonts: EtenFonts) -> int:
    """驗證 Big5 分區索引公式沒有整批偏移。"""
    checks = [
        ("STDFONT idx=0 應為「一」", ("STDFONT", 0), "一"),
        ("中 A4A4", None, "中"),
        ("猴 B555", None, "猴"),
        ("一 A440", None, "一"),
        ("龍 C073", None, "龍"),
        ("，A141 (符號區)", None, "，"),
    ]
    ok = True
    for label, direct, expect in checks:
        if direct is not None:
            glyph = fonts.glyph_at(*direct)
        else:
            hi, lo = expect.encode("big5")
            glyph = fonts.glyph_of_big5(hi, lo)
        if glyph is None or is_blank(glyph):
            print(f"[FAIL] {label}: 取不到字模或全空")
            ok = False
            continue
        # 反查表要能從字模走回同一個字
        back = fonts.reverse.get(glyph, [])
        mark = "OK" if expect in back else "MISMATCH"
        if expect not in back:
            ok = False
        print(f"\n[{mark}] {label} → 反查得到 {back[:5]}")
        for row in render(glyph):
            print("    " + row)

    print(f"\n反查表大小：{len(fonts.reverse)} 個相異字模")
    print(f"oracle: {'全部通過' if ok else '有失敗，索引公式或字庫有問題'}")
    return 0 if ok else 1


def cmd_dump(fonts: EtenFonts, args) -> int:
    path = GAME_DIR / args.file
    data = path.read_bytes()
    if len(data) % STRIDE:
        print(f"[警告] {args.file} 大小 {len(data)} 不是 {STRIDE} 的倍數")
    n = len(data) // STRIDE
    print(f"{args.file}: {len(data)} B = {n} 個字模")
    end = min(n, args.start + args.count)
    for i in range(args.start, end):
        g = data[i * STRIDE : (i + 1) * STRIDE]
        hit = fonts.reverse.get(g, [])
        tag = f"→ {hit[0]}" if hit else "→ (反查不到)"
        print(f"\n#{i}  {tag}  {g.hex()}")
        for row in render(g):
            print("    " + row)
    return 0


def hamming(a: bytes, b: bytes) -> int:
    return int.from_bytes(a, "big").__xor__(int.from_bytes(b, "big")).bit_count()


def nearest(fonts: EtenFonts, glyph: bytes, top: int = 5) -> list[tuple[int, str]]:
    """反查落空時，用逐 bit hamming distance 找最像的字（多半是改過筆劃的自製字）。"""
    scored = []
    for g, chars in fonts.reverse.items():
        scored.append((hamming(glyph, g), chars[0]))
    scored.sort(key=lambda t: t[0])
    return scored[:top]


def cmd_match(fonts: EtenFonts, args) -> int:
    OUT_DIR.mkdir(parents=True, exist_ok=True)
    summary = []
    all_exceptions = []

    for name in GLYPH_FILES:
        path = GAME_DIR / name
        if not path.exists():
            summary.append({"file": name, "error": "檔案不存在"})
            continue
        data = path.read_bytes()
        rem = len(data) % STRIDE
        n = len(data) // STRIDE
        chars: list[str | None] = []
        exceptions = []
        blanks = 0
        for i in range(n):
            g = data[i * STRIDE : (i + 1) * STRIDE]
            if is_blank(g):
                # 空白是排版填充，不是缺字。只記位置，不存 bytes（全零、無資訊，
                # 且字模 bytes 是倚天字型的衍生物，能少存就少存）。
                blanks += 1
                chars.append(None)
                exceptions.append({"idx": i, "kind": "blank"})
                continue
            hit = fonts.reverse.get(g)
            if hit:
                chars.append(hit[0])
            else:
                chars.append(None)
                exceptions.append(
                    {
                        "idx": i,
                        "kind": "unmatched",
                        "hex": g.hex(),
                        "nearest": [
                            {"dist": d, "char": c} for d, c in nearest(fonts, g)
                        ],
                    }
                )
        hits = sum(1 for c in chars if c is not None)
        summary.append(
            {
                "file": name,
                "bytes": len(data),
                "glyphs": n,
                "remainder": rem,
                "matched": hits,
                "blank": blanks,
                "unmatched": n - hits - blanks,
                "rate": round(hits / n * 100, 2) if n else 0.0,
                "text": "".join(c if c else "�" for c in chars),
            }
        )
        for e in exceptions:
            e["file"] = name
        all_exceptions.extend(exceptions)

        (OUT_DIR / f"{name}.json").write_text(
            json.dumps(
                {"file": name, "glyphs": n, "chars": chars},
                ensure_ascii=False,
                indent=1,
            ),
            encoding="utf-8",
        )

    (OUT_DIR / "_summary.json").write_text(
        json.dumps(summary, ensure_ascii=False, indent=1), encoding="utf-8"
    )
    (OUT_DIR / "_exceptions.json").write_text(
        json.dumps(all_exceptions, ensure_ascii=False, indent=1), encoding="utf-8"
    )

    tot_g = sum(s.get("glyphs", 0) for s in summary)
    tot_m = sum(s.get("matched", 0) for s in summary)
    tot_b = sum(s.get("blank", 0) for s in summary)
    tot_u = sum(s.get("unmatched", 0) for s in summary)
    print(f"{'檔案':<12} {'字模':>5} {'命中':>5} {'空白':>4} {'未中':>4} {'命中率':>7}")
    print("-" * 46)
    for s in summary:
        if "error" in s:
            print(f"{s['file']:<12} {s['error']}")
            continue
        print(
            f"{s['file']:<12} {s['glyphs']:>5} {s['matched']:>5} "
            f"{s['blank']:>4} {s['unmatched']:>4} {s['rate']:>6.2f}%"
        )
    print("-" * 46)
    print(
        f"{'合計':<12} {tot_g:>5} {tot_m:>5} {tot_b:>4} {tot_u:>4} "
        f"{tot_m / tot_g * 100:>6.2f}%"
    )
    print(f"\n字表 → {OUT_DIR}")
    return 0


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__)
    sub = ap.add_subparsers(dest="cmd", required=True)
    sub.add_parser("oracle", help="驗證 Big5 分區索引公式")
    d = sub.add_parser("dump", help="把某個字模檔畫成 ASCII art")
    d.add_argument("file")
    d.add_argument("--start", type=int, default=0)
    d.add_argument("--count", type=int, default=5)
    sub.add_parser("match", help="全量反查，產出字表與例外清單")
    args = ap.parse_args()

    fonts = EtenFonts(ETEN_DIR)
    if args.cmd == "oracle":
        return cmd_oracle(fonts)
    if args.cmd == "dump":
        return cmd_dump(fonts, args)
    return cmd_match(fonts, args)


if __name__ == "__main__":
    sys.exit(main())
