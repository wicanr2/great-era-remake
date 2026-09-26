#!/usr/bin/env python3
"""產生 Modern 高解析度用的可再散布 GEMF 字型 atlas。

這個工具只把語系與 UI 文字 rasterize 成專案自己的 atlas；不會複製原版
倚天字型，也不把 Noto 字型檔放進儲存庫。輸入字型應是可再散布的 Noto
Sans／Serif CJK Traditional Chinese（SIL OFL），輸出由純 Go parser 載入。
"""

from __future__ import annotations

import argparse
import json
import struct
from pathlib import Path

from PIL import Image, ImageDraw, ImageFont


MAGIC = b"GEMF"
VERSION = 2
CELL_W = 32
CELL_H = 32


def collect_strings(value: object, out: set[str]) -> None:
    if isinstance(value, str):
        out.add(value)
    elif isinstance(value, dict):
        for item in value.values():
            collect_strings(item, out)
    elif isinstance(value, list):
        for item in value:
            collect_strings(item, out)


def collect_runes(root: Path) -> list[int]:
    strings: set[str] = set()
    translations = root / "translations"
    for path in sorted(translations.rglob("*.json")):
        try:
            collect_strings(json.loads(path.read_text(encoding="utf-8")), strings)
        except (OSError, json.JSONDecodeError) as exc:
            raise SystemExit(f"讀取語系 JSON 失敗：{path}: {exc}") from exc

    # Modern renderer 的固定導航與狀態短句不一定都在語系 JSON；以明確種子
    # 固定進 atlas，避免只因畫面還沒打開某一頁就漏掉字。
    strings.add(
        "返回上一頁下一頁確認取消輸入刪除資料來源可靠度人物自傳人物生平｜／　"
        "政略指令顯示設定原典用語現代白話復古圖形Modern圖形新聞史事"
        "攻擊防守撤退部署結算戰鬥省份日期軍隊兵力糧食金錢人口"
        "民國大時代的故事"
    )
    chars = {ord(ch) for text in strings for ch in text if ch not in "\r\n\t"}
    # ASCII 空白與可見字元是 Modern 狀態列、數字輸入與 debug 診斷的固定邊界。
    chars.update(range(0x20, 0x7F))
    return sorted(chars)


def pack_alpha(image: Image.Image) -> bytes:
    return image.tobytes()


def render_glyph(font: ImageFont.FreeTypeFont, char: str) -> tuple[int, bytes]:
    image = Image.new("L", (CELL_W, CELL_H), 0)
    draw = ImageDraw.Draw(image)
    bbox = draw.textbbox((0, 0), char, font=font)
    if bbox is None:
        return 0, bytes(CELL_W * CELL_H)
    width = max(0, bbox[2] - bbox[0])
    height = max(0, bbox[3] - bbox[1])
    # Noto 32 px CJK glyph 應落在 32×32 cell；組合符號或標點若回報異常
    # bounding box，便縮小到 cell 內，不允許越界。
    scale = min(1.0, CELL_W / max(1, width), CELL_H / max(1, height))
    if scale < 1.0:
        size = max(8, int(round(CELL_H * scale)))
        resized = ImageFont.truetype(font.path, size, index=getattr(font, "index", 0))
        bbox = draw.textbbox((0, 0), char, font=resized)
        font = resized
    draw = ImageDraw.Draw(image)
    bbox = draw.textbbox((0, 0), char, font=font)
    # Modern renderer 以左上角定位字模，因此墨跡靠上；advance 仍保留
    # 比例字距，使 Latin 與 CJK 可在同一行對齊。
    draw.text((-bbox[0], -bbox[1]), char, font=font, fill=255)
    advance = int(round(ImageDraw.Draw(Image.new("L", (1, 1))).textlength(char, font=font)))
    if char == " ":
        advance = max(4, advance)
    else:
        advance = max(1, min(CELL_W, advance))
    return advance, pack_alpha(image)


def write_atlas(root: Path, font_path: Path, font_index: int, output: Path, family: str) -> None:
    chars = collect_runes(root)
    try:
        font = ImageFont.truetype(str(font_path), CELL_H, index=font_index)
    except OSError as exc:
        raise SystemExit(f"載入 Modern 字型失敗：{font_path}: {exc}") from exc

    output.parent.mkdir(parents=True, exist_ok=True)
    with output.open("wb") as stream:
        # magic, version, cell_w, cell_h, row_bytes(alpha bytes), reserved, glyph_count
        stream.write(struct.pack("<4sHHHHHHI", MAGIC, VERSION, CELL_W, CELL_H,
                                 CELL_W, 0, 0, len(chars)))
        for codepoint in chars:
            char = chr(codepoint)
            advance, bitmap = render_glyph(font, char)
            stream.write(struct.pack("<IB3x", codepoint, advance))
            stream.write(bitmap)
    print(f"Modern {family} GEMF v2：{len(chars)} glyphs，輸出 {output}，來源字型索引 {font_index}")


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--root", type=Path, default=Path(__file__).resolve().parents[1])
    parser.add_argument("--font", type=Path, required=True)
    parser.add_argument("--font-index", type=int, default=3)
    parser.add_argument("--family", choices=("sans", "serif"), default="sans")
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    write_atlas(args.root, args.font, args.font_index, args.output, args.family)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
