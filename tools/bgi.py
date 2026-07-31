"""Borland BGI `getimage` 緩衝區的解碼（`.TPC` 與 `NEWSDATA.DAT` 用的格式）。

    +0  u16  width  - 1
    +2  u16  height - 1
    +4  像素資料

像素值 = **plane0<<3 | plane1<<2 | plane2<<1 | plane3**。每個 plane 內是 MSB-first。

[雷] plane 0 是**最高**位元。用一般的 `plane3<<3|…|plane0` 解出來圖形完全正確、
色數也對，只有顏色索引被 4-bit 反轉（2↔4、3↔12、7↔14），肉眼看是「圖對但配色錯」。
這個錯只有拿實機畫面對照才抓得到——實測 `CHOOSE1.TPC` 的索引 2 對應
`CHOOSE.RGB` 的第 4 行，正是 `0b0010 → 0b0100`。

**[雷] plane 是逐列交錯，不是四個完整 plane 依序。** 佈局是

    列0: plane0 plane1 plane2 plane3   ← 每段 rowBytes = ceil(w/8)
    列1: plane0 plane1 plane2 plane3
    ...

大小公式兩種佈局算出來一樣（`4 + ceil(w/8) × h × 4`），所以**大小驗算過了
不代表佈局對**。猜成「四個完整 plane 依序」解出來是雜訊：色數會變成 15 且
分布平均，正確解法的色數少、且有大面積單色（`CHOOSE2.TPC` 實測 5 色，
最大宗佔 79%，與 DOSBox 截圖量到的色數一致）。
"""

from __future__ import annotations

import struct

# BGI/EGA 預設 16 色（RGB）
EGA16 = [
    (0, 0, 0), (0, 0, 170), (0, 170, 0), (0, 170, 170),
    (170, 0, 0), (170, 0, 170), (170, 85, 0), (170, 170, 170),
    (85, 85, 85), (85, 85, 255), (85, 255, 85), (85, 255, 255),
    (255, 85, 85), (255, 85, 255), (255, 255, 85), (255, 255, 255),
]


def header(data: bytes, off: int = 0) -> tuple[int, int, int]:
    """回傳 (width, height, 這張圖的總 bytes)。"""
    w1, h1 = struct.unpack_from("<HH", data, off)
    w, h = w1 + 1, h1 + 1
    return w, h, 4 + ((w + 7) // 8) * h * 4


def decode(data: bytes, off: int = 0) -> tuple[int, int, bytearray]:
    """解出 (w, h, 每像素 1 byte 的索引資料)。"""
    w, h, _ = header(data, off)
    row_bytes = (w + 7) // 8
    base = off + 4
    px = bytearray(w * h)
    for y in range(h):
        ry = y * w
        for p in range(4):
            rb = base + (y * 4 + p) * row_bytes   # 逐列交錯，見模組說明
            bit = 1 << (3 - p)   # plane 0 是最高位，見模組說明
            for x in range(w):
                if data[rb + (x >> 3)] & (0x80 >> (x & 7)):
                    px[ry + x] |= bit
    return w, h, px


def load_rgb_palette(path) -> list[tuple[int, int, int]]:
    """讀 `.RGB` 調色盤：純文字，16 行，每行一組 RGB **百分比**（0-100）。

    不是 6-bit DAC 值——實測八個 `.RGB` 檔的最大值都是 100。轉 8-bit 的鏈是

        百分比 → DAC 6-bit = round(pct × 63 / 100) → 8-bit = (dac<<2)|(dac>>4)

    中間一定要經過 6-bit DAC 這一步。直接 `pct × 255 / 100` 會差 1-2
    （60% → 153 vs 正確的 154），逐像素比對就過不了。
    """
    from pathlib import Path

    nums = [int(t) for t in Path(path).read_text().split()]
    out = []
    for i in range(0, len(nums) - 2, 3):
        dac = [round(v * 63 / 100) for v in nums[i : i + 3]]
        out.append(tuple((d << 2) | (d >> 4) for d in dac))
    return out


def to_image(w: int, h: int, px: bytearray, palette=None):
    """轉成 PIL Image。palette 是 16 個 (r,g,b)，預設用 EGA16。"""
    from PIL import Image

    pal = palette or EGA16
    img = Image.frombytes("P", (w, h), bytes(px))
    flat = []
    for i in range(256):
        flat.extend(pal[i] if i < len(pal) else (0, 0, 0))
    img.putpalette(flat)
    return img
