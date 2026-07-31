"""PKWARE Data Compression Library (DCL) "implode" 解壓，純 Python。

《大時代的故事》的 `.GLB` 圖庫每一筆都是獨立的 DCL stream（header `00 06`：
literal 不編碼、字典 4096）。演算法本身是公開的，這支是 zlib contrib 的
`blast.c` 移植，沒有從遊戲反組譯任何東西。

位元序是 LSB-first，Huffman 碼是 bit-reversed（與 DEFLATE 不同，別套 inflate 的直覺）。
"""

from __future__ import annotations

# 各表是 run-length 編碼的碼長：每個 byte 高 4 bits = 重複次數-1，低 4 bits = 碼長
LITLEN = bytes(
    [
        11, 124, 8, 7, 28, 7, 188, 13, 76, 4, 10, 8, 12, 10, 12, 10, 8, 23, 8,
        9, 7, 6, 7, 8, 7, 6, 55, 8, 23, 24, 12, 11, 7, 9, 11, 12, 6, 7, 22, 5,
        7, 24, 6, 11, 9, 6, 7, 22, 7, 11, 38, 7, 9, 8, 25, 11, 8, 11, 9, 12,
        8, 12, 5, 38, 5, 38, 5, 11, 7, 5, 6, 21, 6, 10, 53, 8, 7, 24, 10, 27,
        44, 253, 253, 253, 252, 252, 252, 13, 12, 45, 12, 45, 12, 61, 12, 45,
        44, 173,
    ]
)
LENLEN = bytes([2, 35, 36, 53, 38, 23])
DISTLEN = bytes([2, 20, 53, 230, 247, 151, 248])

BASE = [3, 2, 4, 5, 6, 7, 8, 9, 10, 12, 16, 24, 40, 72, 136, 264]
EXTRA = [0, 0, 0, 0, 0, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8]

END_LENGTH = 519  # 結束碼


class _Bits:
    """LSB-first 位元讀取器。"""

    __slots__ = ("d", "i", "buf", "cnt")

    def __init__(self, data: bytes):
        self.d = data
        self.i = 0
        self.buf = 0
        self.cnt = 0

    def take(self, need: int) -> int:
        val = self.buf
        while self.cnt < need:
            if self.i >= len(self.d):
                raise ValueError("壓縮資料提前結束")
            val |= self.d[self.i] << self.cnt
            self.i += 1
            self.cnt += 8
        self.buf = val >> need
        self.cnt -= need
        return val & ((1 << need) - 1)


def _construct(rep: bytes) -> tuple[list[int], list[int]]:
    """把 run-length 碼長表展開成 (每長度的碼數, 依長度排序的符號表)。"""
    lengths: list[int] = []
    count = [0] * 16
    for b in rep:
        n = (b >> 4) + 1
        ln = b & 15
        count[ln] += n
        lengths.extend([ln] * n)

    offs = [0] * 16
    for ln in range(1, 15):
        offs[ln + 1] = offs[ln] + count[ln]
    symbol = [0] * len(lengths)
    for sym, ln in enumerate(lengths):
        if ln:
            symbol[offs[ln]] = sym
            offs[ln] += 1
    return count, symbol


def _decode(bs: _Bits, count: list[int], symbol: list[int]) -> int:
    code = first = index = 0
    for length in range(1, 14):
        # [雷] DCL 的 Huffman 碼是 bit-inverted：blast.c 裡是 `code |= (bitbuf & 1) ^ 1`。
        # 漏掉這個反轉，literal 仍解得出來（走的是 raw 8 bits），但一遇到
        # length/distance 對就會算出離譜的距離。
        code |= bs.take(1) ^ 1
        cnt = count[length]
        if code < first + cnt:
            return symbol[index + (code - first)]
        index += cnt
        first = (first + cnt) << 1
        code <<= 1
    raise ValueError("碼流損壞：找不到有效碼")


_LEN = _construct(LENLEN)
_DIST = _construct(DISTLEN)
_LIT = _construct(LITLEN)


def explode(data: bytes, expected: int | None = None) -> bytes:
    """解一段 DCL implode 資料。`expected` 給了就順便驗證輸出長度。"""
    if len(data) < 3:
        raise ValueError("資料太短")
    lit, dictbits = data[0], data[1]
    if lit > 1:
        raise ValueError(f"literal 模式非法: {lit}")
    if not 4 <= dictbits <= 6:
        raise ValueError(f"字典大小非法: {dictbits}")

    bs = _Bits(data[2:])
    out = bytearray()
    while True:
        if bs.take(1):  # 1 = 長度/距離對
            sym = _decode(bs, *_LEN)
            length = BASE[sym] + bs.take(EXTRA[sym])
            if length == END_LENGTH:
                break
            shift = 2 if length == 2 else dictbits
            dist = (_decode(bs, *_DIST) << shift) + bs.take(shift) + 1
            if dist > len(out):
                raise ValueError(f"距離 {dist} 超出已輸出的 {len(out)} bytes")
            start = len(out) - dist
            for k in range(length):
                out.append(out[start + k])
        else:  # 0 = 字面值
            out.append(bs.take(8) if lit == 0 else _decode(bs, *_LIT))

    result = bytes(out)
    if expected is not None and len(result) != expected:
        raise ValueError(f"解出 {len(result)} bytes，預期 {expected}")
    return result
