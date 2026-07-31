"""倚天中文系統 (ETEN 3.53) 16x15 點陣字庫的讀取與 Big5 分區索引。

格式與索引公式的出處：knowledge-base/retro-cht/eten-bitmap-font。
公式不在這裡重推——這支只負責實作它，並提供 oracle 自我驗證。

  STDFONT.15  漢字區   16x15  30 B/字  13094 字
  SPCFONT.15  全形符號 16x15  30 B/字    408 字

點陣佈局：每列 (16+7)/8 = 2 bytes，MSB-first，由上而下共 15 列。
"""

from __future__ import annotations

from pathlib import Path

GLYPH_W = 16
GLYPH_H = 15
ROW_BYTES = (GLYPH_W + 7) // 8   # 2
STRIDE = ROW_BYTES * GLYPH_H     # 30

N_STDFONT = 13094
N_SPCFONT = 408


def raw(hi: int, lo: int) -> int:
    """Big5 (hi, lo) → 線性序號。lo 的 0x7F–0xA0 是空洞，要跳過。"""
    return (hi - 0xA1) * 157 + ((lo - 0x40) if lo < 0x7F else (lo - 0x62))


LAST_SPC = raw(0xA3, 0xBF)      # 符號區尾 = 407
BASE_A440 = raw(0xA4, 0x40)     # 漢字區起點
LAST_COMMON = raw(0xC6, 0x7E)   # 常用字尾
BASE_C940 = raw(0xC9, 0x40)     # 次常用起點
N_COMMON = 5401


def big5_to_font(hi: int, lo: int) -> tuple[str, int] | None:
    """Big5 (hi, lo) → (字庫名, 字庫內索引)。不在任何字庫則回 None。"""
    r = raw(hi, lo)
    if r < 0:
        return None
    if r <= LAST_SPC:
        return ("SPCFONT", r)
    if r < BASE_A440:
        return None                      # A3C0–A3FE 造字/控制區，字庫沒有
    if r <= LAST_COMMON:
        return ("STDFONT", r - BASE_A440)
    if r < BASE_C940:
        return None                      # C6A1–C8FE 造字區
    idx = N_COMMON + (r - BASE_C940)
    if idx >= N_STDFONT:
        return None
    return ("STDFONT", idx)


def iter_big5() -> "list[tuple[int, int]]":
    """所有合法的 Big5 雙位元組組合。"""
    out = []
    for hi in range(0xA1, 0xFA):
        for lo in list(range(0x40, 0x7F)) + list(range(0xA1, 0xFF)):
            out.append((hi, lo))
    return out


class EtenFonts:
    """載入倚天字庫，提供 glyph ↔ Big5 雙向查詢。"""

    def __init__(self, eten_dir: Path):
        self.dir = Path(eten_dir)
        self.std = (self.dir / "STDFONT.15").read_bytes()
        self.spc = (self.dir / "SPCFONT.15").read_bytes()
        self._check_sizes()
        self._rev: dict[bytes, list[str]] | None = None

    def _check_sizes(self) -> None:
        if len(self.std) != N_STDFONT * STRIDE:
            raise ValueError(
                f"STDFONT.15 大小 {len(self.std)} != {N_STDFONT * STRIDE}"
            )
        if len(self.spc) != N_SPCFONT * STRIDE:
            raise ValueError(
                f"SPCFONT.15 大小 {len(self.spc)} != {N_SPCFONT * STRIDE}"
            )

    def glyph_at(self, font: str, idx: int) -> bytes:
        buf = self.std if font == "STDFONT" else self.spc
        return buf[idx * STRIDE : (idx + 1) * STRIDE]

    def glyph_of_big5(self, hi: int, lo: int) -> bytes | None:
        hit = big5_to_font(hi, lo)
        if hit is None:
            return None
        return self.glyph_at(*hit)

    @property
    def reverse(self) -> dict[bytes, list[str]]:
        """30 B 字模 → 對應的 Big5 字元清單（同形字可能多個）。"""
        if self._rev is None:
            rev: dict[bytes, list[str]] = {}
            for hi, lo in iter_big5():
                g = self.glyph_of_big5(hi, lo)
                if g is None or len(g) != STRIDE:
                    continue
                try:
                    ch = bytes([hi, lo]).decode("big5")
                except UnicodeDecodeError:
                    continue
                rev.setdefault(g, []).append(ch)
            self._rev = rev
        return self._rev


def render(glyph: bytes, on: str = "██", off: str = "  ") -> list[str]:
    """把 30 B 字模畫成 15 列 ASCII art。"""
    rows = []
    for y in range(GLYPH_H):
        bits = []
        for bx in range(ROW_BYTES):
            b = glyph[y * ROW_BYTES + bx]
            for k in range(7, -1, -1):
                bits.append(on if (b >> k) & 1 else off)
        rows.append("".join(bits))
    return rows


def is_blank(glyph: bytes) -> bool:
    return not any(glyph)
