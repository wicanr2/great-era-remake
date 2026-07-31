#!/usr/bin/env python3
"""偵測字模檔的「定長槽位」寬度，把字模序列切回詞條。

反查字表後發現的結構（docs/formats/01-glyph-fonts.md）：字模檔不是「字表 + 索引」，
而是**已排好版的定長欄位陣列**——每 w 個字模 = 一個詞條，不足的格子填空白字模。

判定 w 不靠肉眼。對每個候選 w（n 的因數），檢查每個槽位的非空字模位置是否符合
四種排版之一，取**符合率最高、且平手時取最小**的 w：

    全滿  ████        靠左  ██▁▁        靠右  ▁▁██        分散  █▁█▁█

分散對齊是這批檔案真實存在的排版（「鎮␣威␣軍」「劉␣驥」），不是解析誤差。

用法：
    tools/py.sh tools/glyph_slots.py            全部檔案
    tools/py.sh tools/glyph_slots.py TN15.1     單一檔案，列出詞條
"""

from __future__ import annotations

import json
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
TABLE_DIR = ROOT / "docs" / "formats" / "glyph-tables"

FILES = (
    [f"TN15.{i}" for i in range(1, 40)]
    + [f"{i}.15" for i in range(1, 5)]
    + [f"FAN({i}).15" for i in range(1, 4)]
    + ["JAPAN1.15", "JAPAN2.15"]
    + ["MAN115", "MAN215", "MAN315"]
)


# 槽寬的判定依據分三級，文件裡要照實標，不可混為一談：
#   auto    空白排版模式推出來的（有空白可資判別，符合率 100%）
#   content 檔案零空白，空白模式推不動，改由詞條語意判定（證據較弱）
#   待反組譯確認的一律不在這裡寫死
KNOWN_WIDTH: dict[str, tuple[int, str]] = {
    # 零空白 → auto 無法判別，靠內容：全部是二字詞（調動/軍事/運補/徵稅…）
    "2.15": (2, "content"),
    # 幾乎零空白，靠內容：全部是四字訊息（您確定嗎/報告司令/按任意鍵…）
    "4.15": (4, "content"),
    # 單字字元表，不是定長詞條：全形數字 0-9 + 常用單字 + 游標符號 ◣◥
    "1.15": (1, "content"),
    # 番號模板，空白是「執行期填數字」的佔位（第␣集團軍 / 騎兵第␣軍 / 第␣␣戰區）。
    # 佔位在詞條中間，不屬於靠左/靠右/置中/分散任一種，auto 因此誤判成 w=2。
    "FAN(2).15": (5, "content"),
    "FAN(3).15": (5, "content"),
}


def divisors(n: int, lo: int = 2, hi: int = 12) -> list[int]:
    return [w for w in range(lo, hi + 1) if n % w == 0]


def classify(mask: list[bool]) -> str | None:
    """判斷一個槽位的非空分布屬於哪種排版；都不像則回 None。"""
    w = len(mask)
    k = sum(mask)
    if k == 0:
        return "empty"
    if k == w:
        return "full"
    pos = [i for i, m in enumerate(mask) if m]
    if pos == list(range(k)):
        return "left"
    if pos == list(range(w - k, w)):
        return "right"
    # 置中連續
    start = (w - k) // 2
    if pos == list(range(start, start + k)):
        return "center"
    # 分散對齊：k 個字等間隔攤在 w 格
    if k > 1:
        just = [round(i * (w - 1) / (k - 1)) for i in range(k)]
        if pos == just:
            return "justify"
    return None


def score(chars: list[str | None], w: int) -> tuple[float, dict[str, int]]:
    n = len(chars)
    kinds: dict[str, int] = {}
    good = 0
    for s in range(0, n, w):
        mask = [c is not None for c in chars[s : s + w]]
        kind = classify(mask)
        if kind == "empty":
            # 全空槽位是解析錯誤的訊號，直接扣分
            kinds["empty"] = kinds.get("empty", 0) + 1
            continue
        if kind:
            good += 1
            kinds[kind] = kinds.get(kind, 0) + 1
        else:
            kinds["odd"] = kinds.get("odd", 0) + 1
    return (good / (n / w), kinds)


def best_width(chars: list[str | None]) -> tuple[int, float, dict]:
    """挑最能解釋這份資料的槽寬。

    排序鍵（依序）：符合率高 → **排版種類少** → w 大。

    「排版種類少」這條是關鍵：小的 w 常常是真正槽寬的因數，會把詞條攔腰切開，
    切出來的碎片剛好散落成 left/right/justify 各一些，符合率一樣是 100%，
    但排版模式雜亂。真正的槽寬對齊方式是一致的（TN15 全部靠左）。
    w 大優先則排除「w=真值的倍數」這種反向誤判。
    """
    n = len(chars)
    cands = divisors(n)
    if not cands:
        return (n, *score(chars, n))  # 質數：整份當一個槽位，交給人判讀
    ranked = []
    for w in cands:
        sc, kinds = score(chars, w)
        real_kinds = {k: v for k, v in kinds.items() if k not in ("empty", "odd")}
        ranked.append((sc, -len(real_kinds), w, kinds))
    ranked.sort(reverse=True)
    sc, _, w, kinds = ranked[0]
    return (w, sc, kinds)


def load(name: str) -> list[str | None]:
    p = TABLE_DIR / f"{name}.json"
    return json.loads(p.read_text(encoding="utf-8"))["chars"]


def entries(chars: list[str | None], w: int) -> list[str]:
    out = []
    for s in range(0, len(chars), w):
        slot = chars[s : s + w]
        out.append("".join(c for c in slot if c is not None))
    return out


def resolve(name: str, chars: list[str | None]) -> tuple[int, float, dict, str]:
    """回傳 (槽寬, 符合率, 排版分布, 依據)。"""
    if name in KNOWN_WIDTH:
        w, basis = KNOWN_WIDTH[name]
        sc, kinds = score(chars, w) if w > 1 else (1.0, {"single": len(chars)})
        return w, sc, kinds, basis
    w, sc, kinds = best_width(chars)
    return w, sc, kinds, "auto"


def main() -> int:
    targets = sys.argv[1:] or FILES
    if len(targets) == 1 and targets[0] in FILES:
        name = targets[0]
        chars = load(name)
        w, sc, kinds, basis = resolve(name, chars)
        print(
            f"{name}: {len(chars)} 字模, 槽寬 w={w} ({basis}), "
            f"符合率 {sc:.1%}, {kinds}\n"
        )
        for i, e in enumerate(entries(chars, w)):
            print(f"  [{i:>3}] {e}")
        return 0

    print(f"{'檔案':<12} {'字模':>5} {'槽寬':>4} {'詞條':>5} {'符合率':>7} {'依據':<8} 排版分布")
    print("-" * 84)
    result = {}
    for name in targets:
        chars = load(name)
        w, sc, kinds, basis = resolve(name, chars)
        result[name] = {
            "glyphs": len(chars),
            "width": w,
            "score": sc,
            "basis": basis,
            "kinds": kinds,
            "entries": entries(chars, w),
        }
        dist = " ".join(f"{k}={v}" for k, v in sorted(kinds.items()))
        print(
            f"{name:<12} {len(chars):>5} {w:>4} {len(chars) // w:>5} "
            f"{sc:>6.1%} {basis:<8} {dist}"
        )
    (TABLE_DIR / "_slots.json").write_text(
        json.dumps(result, ensure_ascii=False, indent=1), encoding="utf-8"
    )
    n_auto = sum(1 for v in result.values() if v["basis"] == "auto")
    print(
        f"\n共 {len(result)} 檔："
        f"{n_auto} 檔由空白排版模式判定，{len(result) - n_auto} 檔由詞條語意判定"
        f"（較弱，待反組譯確認）"
    )
    return 0


if __name__ == "__main__":
    sys.exit(main())
