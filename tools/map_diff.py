#!/usr/bin/env python3
"""比對兩張政略畫面截圖的**地圖區**，量出各勢力的顏色面積變化。

用途：驗「電腦會不會出兵」這類問題（`docs/re/29`）。政略畫面的地圖用
顏色標勢力範圍，所以**面積變化 = 省份易主**，不必進去逐省查數值。

[雷] 裁切區必須避開左側面板與底部訊息列——兩張截圖如果停在不同的省，
面板的數字就不一樣，混進來會讓 diff 完全失去意義（第一次就是這樣栽的：
量到 9.86% 相異，其中最大宗是面板底色）。

[雷] **不可跨 run 比較。** 每次開新遊戲的起點與初始局面都不同
（`docs/playtest/13`）。只能比同一次 run 的前後截圖。

用法：
    tools/py.sh tools/map_diff.py shots/x00.png shots/x01.png
    tools/py.sh tools/map_diff.py shots/x0*.png        # 依序兩兩比
"""

from __future__ import annotations

import sys
from pathlib import Path

try:
    import numpy as np
    from PIL import Image
except ImportError:  # pragma: no cover
    print("需要 pillow 與 numpy", file=sys.stderr)
    raise SystemExit(1)

# 遊戲畫面在 1024×768 的 Xvfb root 左上角，邏輯解析度 640×350。
GAME_W, GAME_H = 640, 350

# 地圖區（在 640×350 的座標系）。左側面板到 x≈212，底部訊息列從 y≈266 起。
# 留一點邊避免框線抗鋸齒。
MAP_BOX = (218, 6, 630, 262)

# 已知的勢力色（從截圖取樣）。沒列到的顏色照樣統計，只是沒有名字。
KNOWN = {
    (255, 255, 255): "無主（白）",
    (255, 85, 85): "紅",
    (0, 0, 170): "深藍",
    (0, 170, 170): "青",
    (85, 255, 85): "亮綠",
    (0, 170, 0): "綠",
    (170, 170, 170): "亮灰",
    (85, 85, 85): "深灰",
    (255, 255, 85): "黃",
    (170, 0, 0): "暗紅",
    (255, 85, 255): "粉",
}


def map_area(path: Path) -> np.ndarray:
    im = Image.open(path).convert("RGB").crop((0, 0, GAME_W, GAME_H))
    return np.array(im.crop(MAP_BOX))


def hist(a: np.ndarray, floor: int = 200) -> dict[tuple[int, int, int], int]:
    px = a.reshape(-1, 3)
    uniq, cnt = np.unique(px, axis=0, return_counts=True)
    return {tuple(int(x) for x in u): int(c) for u, c in zip(uniq, cnt) if c >= floor}


def compare(p1: Path, p2: Path) -> None:
    a, b = map_area(p1), map_area(p2)
    if a.shape != b.shape:
        print(f"{p1.name} vs {p2.name}: 尺寸不同，跳過")
        return
    diff = (a != b).any(axis=2)
    ha, hb = hist(a), hist(b)
    print(f"\n=== {p1.name} → {p2.name}")
    print(f"地圖區相異 {diff.sum()} / {diff.size} 像素（{100 * diff.sum() / diff.size:.2f}%）")
    rows = []
    for k in set(ha) | set(hb):
        dv = hb.get(k, 0) - ha.get(k, 0)
        if abs(dv) >= 100:
            rows.append((dv, k, ha.get(k, 0), hb.get(k, 0)))
    if not rows:
        print("  沒有面積變化超過 100 px 的顏色——**局面沒動**")
        return
    for dv, k, before, after in sorted(rows, key=lambda r: -abs(r[0])):
        name = KNOWN.get(k, f"RGB{k}")
        print(f"  {name:<12} {before:6d} → {after:6d}  ({dv:+d})")


def main() -> int:
    args = [Path(a) for a in sys.argv[1:]]
    if len(args) < 2:
        print(__doc__, file=sys.stderr)
        return 1
    for p1, p2 in zip(args, args[1:]):
        compare(p1, p2)
    print()
    print("⚠️ 顏色面積變化只證明「有省份易主」，不證明是誰打誰——")
    print("   要歸因得看該省的司令欄位，或抓下一步的事件畫面。")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
