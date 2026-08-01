"""量「不同 Big5 碼點、相同 30 bytes 字形」在倚天字庫裡有多少。

**為什麼要量**：`docs/formats/01-glyph-text.md` 的字模反查是 byte-exact
比對，命中就當作那個碼點。可是只要字庫裡有兩個碼點共用同一個字形，
反查就永遠取到排前面的那一個——後面那個碼點會**系統性地查不到**，
而且症狀是「查得到、但是錯字」，比查不到還難發現。

`docs/reference/people/02-status.md` 列出 22 筆「遊戲寫法 vs 通行寫法」
的差異（閰/閻、勛/勳、爛/斕…），其中一部分有可能就是這個原因造成的，
不是遊戲真的用了異體字。這支把風險量出來。

用法：

    tools/py.sh tools/dup_glyph.py
"""

from __future__ import annotations

import collections
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

import eten  # noqa: E402

FONT_DIR = Path("workplace/eten")


def big5_table() -> dict[int, bytes]:
    """線性序號 → Big5 兩 bytes（只建一次，反查用）。"""
    out: dict[int, bytes] = {}
    for hi in range(0xA1, 0xFA):
        for lo in list(range(0x40, 0x7F)) + list(range(0xA1, 0xFF)):
            out.setdefault(eten.raw(hi, lo), bytes([hi, lo]))
    return out


def main() -> int:
    std = (FONT_DIR / "STDFONT.15").read_bytes()
    n = len(std) // eten.STRIDE
    print(f"STDFONT.15: {len(std)} B → {n} 字")

    buckets: dict[bytes, list[int]] = collections.defaultdict(list)
    for i in range(n):
        buckets[std[i * eten.STRIDE : (i + 1) * eten.STRIDE]].append(i)

    zero = bytes(eten.STRIDE)
    dups = {g: v for g, v in buckets.items() if len(v) > 1 and g != zero}
    masked = sum(len(v) - 1 for v in dups.values())
    print(f"非空白的重複字形組數: {len(dups)}")
    print(f"被遮蔽的碼點數: {masked}（{masked / n * 100:.2f}%）")
    if zero in buckets:
        print(f"（全零字形另佔 {len(buckets[zero])} 個碼點，那是空白不是誤判來源）")

    if not dups:
        print("→ 反查不會取錯字：每個非空白字形都對應唯一碼點。")
        return 0

    tbl = big5_table()
    for g, idxs in list(dups.items())[:20]:
        codes = []
        for i in idxs:
            b = tbl.get(i)
            if not b:
                continue
            try:
                codes.append(f"{b.hex()}={b.decode('big5')}")
            except Exception:
                codes.append(b.hex())
        print("  重複:", ", ".join(codes))
    if len(dups) > 20:
        print(f"  …（共 {len(dups)} 組，只列前 20）")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
