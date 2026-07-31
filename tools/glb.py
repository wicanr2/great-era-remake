#!/usr/bin/env python3
"""`.GLB` 圖庫解包（配 `.GTB` 目錄）。

格式（`docs/formats/02-glb-image-library.md`）：

    GTB = 目錄，每筆 13 bytes
        +0  u8   flag       0 = 調色盤(768 B), 1 = 圖像
        +1  u16  raw_size   解壓後大小（圖像 = w × h，每像素 1 byte）
        +3  u16  comp_size  壓縮後大小
        +5  u32  offset     GLB 內位移
        +9  u16  width
        +11 u16  height

    GLB = 每筆各自是一段獨立的 PKWARE DCL implode 資料（header `00 06`）

調色盤沿用「最近的前一筆調色盤」——GLB 內調色盤與圖像交錯，一個調色盤
管到下一個調色盤出現為止。

用法：
    tools/py.sh tools/glb.py list GRT
    tools/py.sh tools/glb.py extract GRT --out workplace/glb-out
"""

from __future__ import annotations

import argparse
import struct
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
from dcl import explode  # noqa: E402

ROOT = Path(__file__).resolve().parent.parent
GAME = ROOT / "workplace" / "orig" / "game"

ENTRY = 13


def read_dir(stem: str) -> list[dict]:
    gtb = (GAME / f"{stem}.GTB").read_bytes()
    out = []
    for k in range(len(gtb) // ENTRY):
        e = gtb[k * ENTRY : (k + 1) * ENTRY]
        flag = e[0]
        raw, comp = struct.unpack_from("<HH", e, 1)
        off = struct.unpack_from("<I", e, 5)[0]
        w, h = struct.unpack_from("<HH", e, 9)
        out.append(
            {
                "index": k,
                "flag": flag,
                "kind": "palette" if w == 0 else "image",
                "raw": raw,
                "comp": comp,
                "offset": off,
                "w": w,
                "h": h,
            }
        )
    return out


def load(stem: str) -> tuple[list[dict], bytes]:
    return read_dir(stem), (GAME / f"{stem}.GLB").read_bytes()


def decode_entry(glb: bytes, ent: dict) -> bytes:
    return explode(glb[ent["offset"] : ent["offset"] + ent["comp"]], ent["raw"])


def cmd_list(args) -> int:
    entries, glb = load(args.stem)
    print(f"{args.stem}.GTB: {len(entries)} 筆 / {args.stem}.GLB: {len(glb)} B\n")
    print(f"{'#':>3} {'類型':<8}{'w':>5}{'h':>5}{'壓縮':>8}{'原始':>8}{'位移':>9}  壓縮率")
    for e in entries:
        ratio = e["comp"] / e["raw"] if e["raw"] else 0
        print(
            f"{e['index']:>3} {e['kind']:<8}{e['w']:>5}{e['h']:>5}"
            f"{e['comp']:>8}{e['raw']:>8}{e['offset']:>9}  {ratio:>6.1%}"
        )
    return 0


def cmd_extract(args) -> int:
    from PIL import Image

    entries, glb = load(args.stem)
    out = Path(args.out) / args.stem
    out.mkdir(parents=True, exist_ok=True)

    pal: list[int] | None = None
    pal_from = -1
    n = 0
    for e in entries:
        data = decode_entry(glb, e)
        if e["kind"] == "palette":
            # VGA DAC 是 6-bit（0-63）→ 8-bit 用位元複製擴展 (v<<2)|(v>>4)，
            # 不是 v*255//63。兩者差 1（16→65 vs 64、51→207 vs 206），
            # 逐像素比對原版截圖時這 1 的差會讓整張圖判定為不同。
            pal = [(v << 2) | (v >> 4) for v in data]
            pal_from = e["index"]
            (out / f"{e['index']:02d}-palette.pal").write_bytes(data)
            continue
        if pal is None:
            print(f"  #{e['index']} 沒有可用的調色盤，跳過")
            continue
        img = Image.frombytes("P", (e["w"], e["h"]), data)
        img.putpalette(pal)
        name = f"{e['index']:02d}-{e['w']}x{e['h']}-pal{pal_from:02d}.png"
        img.convert("RGB").save(out / name)
        n += 1
    print(f"{args.stem}: 輸出 {n} 張圖 → {out}")
    return 0


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__)
    sub = ap.add_subparsers(dest="cmd", required=True)
    p = sub.add_parser("list")
    p.add_argument("stem", help="GRT 或 GRT2")
    p = sub.add_parser("extract")
    p.add_argument("stem")
    p.add_argument("--out", default="workplace/glb-out")
    args = ap.parse_args()
    return cmd_list(args) if args.cmd == "list" else cmd_extract(args)


if __name__ == "__main__":
    sys.exit(main())
