"""Turbo Pascal 48-bit Real 的解碼。

格式（`docs/re/10` §3）：

    byte0      exponent，0 表示整個數是 0
    bytes1-5   mantissa，**低位在前**，最高位是隱含的 1（bit 39 是符號）

    值 = ±(mantissa | 1<<39) / 2^39 × 2^(exponent - 0x81)

編譯器也會把常數放進暫存器對 `AX:BX:DX` / `CX:SI:DI` 傳遞，
排列是 `AL`=exponent、`AH`=mantissa[0]、`BL`=[1]、`BH`=[2]、`DL`=[3]、`DH`=[4]。

自我驗證（`--self-test`）用五個已經在別處驗過的常數當正對照：
10（`docs/re/25`）、0.8（學潮，`covert.go`）、1.5、2.5、3（`docs/re/30`）。
"""

from __future__ import annotations

import sys


def decode(b: bytes) -> float:
    """6 bytes → float。"""
    if len(b) != 6:
        raise ValueError(f"Real 是 6 bytes，收到 {len(b)}")
    exp = b[0]
    if exp == 0:
        return 0.0
    m = 0
    for i, x in enumerate(b[1:6]):
        m |= x << (8 * i)
    sign = -1.0 if (m >> 39) & 1 else 1.0
    m |= 1 << 39
    m &= (1 << 40) - 1
    return sign * m / (1 << 39) * 2.0 ** (exp - 0x81)


def from_regs(ax: int, bx: int, dx: int) -> float:
    """`AX:BX:DX` 三個暫存器 → float。"""
    return decode(bytes([ax & 0xFF, ax >> 8, bx & 0xFF, bx >> 8, dx & 0xFF, dx >> 8]))


CASES = [
    ("10", bytes([0x84, 0, 0, 0, 0x00, 0x20]), 10.0),
    ("0.8", bytes([0x80, 0xCD, 0xCC, 0xCC, 0xCC, 0x4C]), 0.8),
    ("1.5", bytes([0x81, 0, 0, 0, 0, 0x40]), 1.5),
    ("2.5", bytes([0x82, 0, 0, 0, 0, 0x20]), 2.5),
    ("3", bytes([0x82, 0, 0, 0, 0, 0x40]), 3.0),
]


def self_test() -> int:
    bad = 0
    for name, b, want in CASES:
        got = decode(b)
        ok = abs(got - want) < 1e-9
        print(f"  {'OK ' if ok else 'FAIL'} {name:>5} {b.hex(' ')} → {got}")
        bad += not ok
    print("正對照全過" if not bad else f"{bad} 個不符")
    return 1 if bad else 0


if __name__ == "__main__":
    if "--self-test" in sys.argv:
        raise SystemExit(self_test())
    args = [int(a, 16) for a in sys.argv[1:]]
    if len(args) == 3:
        print(from_regs(*args))
    elif len(args) == 6:
        print(decode(bytes(args)))
    else:
        print(__doc__)
        print("用法: tools/py.sh tools/tp_real.py --self-test")
        print("      tools/py.sh tools/tp_real.py 5280 1EB8 2B85     # AX BX DX")
        print("      tools/py.sh tools/tp_real.py 80 52 B8 1E 85 2B  # 六個 byte")
