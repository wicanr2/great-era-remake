#!/usr/bin/env python3
"""量全部原版檔案的熵與結構特徵。

`CLAUDE.md` §7 第 8 條：任何「這東西被壓縮／加密了」的說法，先量熵再說。

熵的判讀（bit/byte，理論上限 8.0）：
    < 4.0   結構化資料、稀疏點陣、文字表
    4.0–6.5 一般二進位（程式碼、混合資料）
    6.5–7.5 壓縮過，或高熵的點陣圖
    > 7.5   壓縮或加密，幾乎沒有冗餘

用法：tools/py.sh tools/entropy.py
"""

from __future__ import annotations

import math
import sys
from collections import Counter
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
GAME = ROOT / "workplace" / "orig" / "game"


def entropy(data: bytes) -> float:
    if not data:
        return 0.0
    counts = Counter(data)
    n = len(data)
    return -sum((c / n) * math.log2(c / n) for c in counts.values())


def printable_ratio(data: bytes) -> float:
    """可列印 ASCII 佔比。純文字檔會接近 1.0。"""
    if not data:
        return 0.0
    ok = sum(1 for b in data if 32 <= b < 127 or b in (9, 10, 13))
    return ok / len(data)


def zero_ratio(data: bytes) -> float:
    return data.count(0) / len(data) if data else 0.0


def guess(name: str, size: int, ent: float, pr: float, zr: float) -> str:
    tags = []
    if pr > 0.95:
        tags.append("純文字")
    if ent > 7.5:
        tags.append("壓縮/加密")
    elif ent > 6.5:
        tags.append("高熵(壓縮或點陣圖)")
    elif ent < 4.0:
        tags.append("低熵(結構化/稀疏)")
    if size % 30 == 0 and size > 0:
        tags.append(f"÷30={size // 30}")
    if zr > 0.4:
        tags.append(f"零佔{zr:.0%}")
    return " ".join(tags) or "-"


def main() -> int:
    files = sorted(p for p in GAME.iterdir() if p.is_file())
    rows = []
    for p in files:
        data = p.read_bytes()
        ent = entropy(data)
        pr = printable_ratio(data)
        zr = zero_ratio(data)
        rows.append((p.name, len(data), ent, pr, zr, guess(p.name, len(data), ent, pr, zr)))

    rows.sort(key=lambda r: -r[2])
    print(f"{'檔案':<16}{'大小':>10}{'熵':>7}{'可列印':>8}{'零':>7}  判讀")
    print("-" * 86)
    for name, size, ent, pr, zr, tag in rows:
        print(f"{name:<16}{size:>10}{ent:>7.2f}{pr:>8.0%}{zr:>7.0%}  {tag}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
