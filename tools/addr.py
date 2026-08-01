#!/usr/bin/env python3
"""位址反查：把組語裡的 `[di-6221h]` 這種東西翻回「省份記錄 +20 司令」。

    ./tools/py.sh tools/addr.py -6221h
    ./tools/py.sh tools/addr.py 7A8Bh 62A8h 625Ah 7A92h
    ./tools/py.sh tools/addr.py --bases          列出所有已知基址

## 為什麼要這個（2026-08-01）

`word_64944` 被推錯兩次，而**答案四份文件裡都有**：

    docs/re/10 L202   mov [di-6221h], cx  ; 省份[目標].+20 = 我方勢力
    docs/re/20 L22    省份記錄基址 -6235h、將領記錄基址 +7A7Dh
    docs/spec/03 L59  +20 u16 = 司令
    internal/game/combat.go   +14 = 效忠的勢力領袖

答案是完整的，缺的是**反向索引**。專案有函式索引（`sub_XXXXX` → 結論，
`docs/re/00-function-index.md`），卻沒有位址索引——而讀組語時手上拿到的
正是位址，最需要的那份索引剛好缺席。

這支工具補的就是那一塊：**輸入位址，輸出「哪個記錄的第幾個欄位、語意是什麼」**。

## 維護

解出新的基址或欄位就加進下面兩張表。表是手寫的——**故意的**，
因為從 markdown 自動抽會把「假說」跟「confirmed」混在一起，
而這張表要能直接拿來下判斷，混了就沒有價值。每一列都標推論等級。
"""

import sys

# ── 記錄基址（16-bit，可為負，會被當成 0x10000 的補數）─────────────────
#
# 每一列：基址 → (名稱, 記錄大小, 出處, 推論等級)
BASES = [
    (0x7A7D, "執行期單位記錄", 33, "docs/re/05 / docs/spec/02", "confirmed"),
    (-0x6235, "省份記錄", 37, "docs/re/20 §基址", "confirmed"),
]

# ── 已解出的欄位（記錄名 → {偏移: (語意, 出處, 推論等級)}）──────────────
FIELDS = {
    "執行期單位記錄": {
        0: ("能力值", "docs/spec/02", "confirmed"),
        3: ("經驗（劇本檔一律 0，只有存檔有值）", "docs/spec/02 §7 / playtest/08", "confirmed"),
        4: ("所屬省", "docs/spec/02", "confirmed"),
        5: ("戰場格編號 0..195（0xFF = 不在場上）", "docs/re/07 §2", "confirmed"),
        6: ("機動力上限", "docs/re/07 §7", "confirmed"),
        7: ("機動力剩餘", "docs/re/07 §7", "confirmed"),
        8: ("攻方 1／守方 0", "docs/re/07 §8", "confirmed"),
        9: ("命令類型 1–6", "docs/re/31 §15/§36", "confirmed"),
        10: ("目標單位 ID", "docs/re/31 §23", "confirmed"),
        12: ("往目標走的下一格（0xFF = 走不到）", "docs/re/31 §12", "confirmed"),
        13: ("旗標位元組；bit 7 = 本回合已指派目標（bit 5 由值 19 設，語意未解）", "docs/re/31 §31/§40", "confirmed"),
        14: ("效忠的勢力領袖（將領 ID）", "docs/re/31 §37 / combat.go", "confirmed"),
        16: ("bit 0 在場、bit 2 已部署", "docs/re/05", "confirmed"),
        17: ("兵力（u16）；sub_3A4CE 把一方 10 格加總當補給分母", "docs/spec/02 §3 / re/31 §48", "confirmed"),
        19: ("士兵戰技（實機「查閱將領」定名）", "docs/playtest/08 §2", "confirmed"),
        20: ("武裝程度（實機「查閱將領」定名）", "docs/playtest/08 §2", "confirmed"),
        21: ("兵種 1 步兵／6 騎兵／4 砲兵／5 裝甲", "docs/spec/02 §4", "confirmed"),
        29: ("體力", "docs/spec/02", "confirmed"),
        29: ("體力（0..100）", "docs/playtest/08 §4", "confirmed"),
        30: ("士氣；>40 時每回合 −20%", "docs/playtest/08 §4 / mechanics/30-combat", "confirmed"),
        31: ("砲兵射擊朝向 1..6", "docs/formats/05 §3", "強證據"),
    },
    "省份記錄": {
        20: ("司令（將領 ID，1-based；0 = 無主）", "docs/spec/03 §2", "confirmed"),
        22: ("鄰省表", "docs/spec/03", "confirmed"),
        22: ("鄰省列表（8×u8）", "docs/spec/03 §2/§4", "confirmed"),
        32: ("旗標；bit 2 = 本回合已處理過；bit 6 讓鄰省被跳過（語意未解）", "docs/re/14 §1 / re/31 §47", "confirmed"),
    },
}

# ── 獨立的全域表與變數（不屬於任何記錄）────────────────────────────────
GLOBALS = {
    0x62A8: ("格佔用表 word[196]：格 → 站在上面的單位 ID", "docs/re/07 §已記", "confirmed"),
    0x625A: ("六角位移表：cell + word[625Ah + (cell mod 2)×12 + dir×2]", "docs/re/07 §6", "confirmed"),
    0x750: ("第一方單位陣列（10 × word，index 1..10）", "docs/re/05", "confirmed"),
    0x764: ("第二方單位陣列（10 × word，index 1..10）", "docs/re/05", "confirmed"),
    0x64902: ("第一方單位陣列的第 1 格（「首位單位」）", "docs/re/31 §20", "confirmed"),
    0x64916: ("第二方單位陣列的第 1 格", "docs/re/31 §20", "confirmed"),
    0x64900: ("戰鬥回合數（1-based，開戰 1、回合末 +1；上限 16）", "docs/re/31 §49", "confirmed"),
    0x64901: ("戰鬥勝方：1 = 第一方、2 = 第二方、0 = 未分", "docs/re/31 §49", "confirmed"),
    0x6492A: ("第一方兵力總和（32-bit，與 6492C 成對；sub_4DD04 算好快取）", "docs/re/31 §49", "confirmed"),
    0x6492E: ("第二方兵力總和（32-bit，與 64930 成對）", "docs/re/31 §49", "confirmed"),
    0x6A40A: ("戰鬥結束旗標", "docs/re/31 §49", "confirmed"),
    0x64932: ("第一方帶進戰場的黃金（結算寫回省份 +0）", "docs/re/05 §2 / re/31 §48", "confirmed"),
    0x64934: ("第二方帶進戰場的黃金", "docs/re/05 §2 / re/31 §48", "confirmed"),
    0x64936: ("第一方帶進戰場的糧食（結算寫回省份 +2）", "docs/re/05 §2 / re/31 §48", "confirmed"),
    0x64938: ("第二方帶進戰場的糧食", "docs/re/05 §2 / re/31 §48", "confirmed"),
    0x6493A: ("第一方帶進戰場的彈藥（結算寫回省份 +4）；分支 B 值 2 要它為 0", "docs/re/05 §2 / re/31 §43/§48", "confirmed"),
    0x6493C: ("第二方帶進戰場的彈藥", "docs/re/05 §2 / re/31 §48", "confirmed"),
    0x6493E: ("第一方帶進戰場的燃料（結算寫回省份 +8）", "docs/re/05 §2 / re/31 §48", "confirmed"),
    0x64940: ("第二方帶進戰場的燃料", "docs/re/05 §2 / re/31 §48", "confirmed"),
    0x64942: ("首位單位效忠的勢力領袖（= 將領[word_64902].+14）", "docs/re/31 §37", "confirmed"),
    0x64944: ("當前交戰省的司令（= 省份[byte_6FFC4].+20）", "docs/re/31 §37", "confirmed"),
    0x6FFC4: ("當前交戰的省份編號", "docs/re/31 §37", "confirmed"),
    0x6FE7D: ("遊戲內年（民國紀年）", "docs/re/16 §3b", "confirmed"),
    0x6FE7E: ("遊戲內月；戰鬥回合上限在 2 月是 15、其他 16", "docs/re/16 §3b / re/31 §49", "confirmed"),
    0x6FE88: ("時期編號 1/2/3，也是 MAN(N).DAT／TOWN(N).DAT 的 N", "docs/re/31 §50", "confirmed"),
    0x7001E: ("該期可用的省份數（北伐 36、抗戰 39）", "docs/re/11 / re/31 §50", "confirmed"),
    0x6BC4A: ("該期的將領數（274／191／191）", "docs/spec/02 §7", "confirmed"),
    0x6AA84: ("戰鬥 AI 分支 B 的決策值（& 1Fh）；bit 6/7 另有用途", "docs/re/31 §3", "confirmed"),
    0x6AA85: ("戰鬥 AI 分支 A 的決策值（& 1Fh）；bit 6/7 見 §18", "docs/re/31 §3/§18", "confirmed"),
    0x6FFCA: ("難度／階段旗標位元組", "docs/mechanics/70-ai", "confirmed"),
}


def norm(v: int) -> int:
    """把 16-bit 負偏移正規化成 0..0xFFFF。"""
    return v & 0xFFFF


def parse(tok: str) -> int:
    t = tok.strip().lower().rstrip("h")
    neg = t.startswith("-")
    if neg:
        t = t[1:]
    t = t.removeprefix("0x")
    v = int(t, 16)
    return -v if neg else v


def lookup(addr: int) -> list[dict]:
    """回傳結構化的查詢結果，**含推論等級**。

    ⚠️ 等級一定要跟著語意一起傳出去。使用者 2026-08-01 質疑
    「在 IDA 裡命名」的理由正是：**名字不帶推論等級，讀起來像事實**，
    取錯了就把錯誤烙進之後每一次 dump。附加式的註記若也拿掉等級，
    就犯同一個錯。
    """
    out = []
    a = norm(addr)

    for base, (desc, src, level) in GLOBALS.items():
        if norm(base) == a:
            out.append({"kind": "全域", "desc": desc, "src": src, "level": level})

    for base, name, size, src, level in BASES:
        delta = (a - norm(base)) & 0xFFFF
        if delta < size:
            fields = FIELDS.get(name, {})
            if delta in fields:
                d, fsrc, flevel = fields[delta]
                out.append({"kind": f"{name} +{delta}", "desc": d,
                            "src": fsrc, "level": flevel})
            else:
                out.append({"kind": f"{name} +{delta}", "desc": "這個欄位還沒解",
                            "src": src, "level": "未知"})
    return out


def explain(addr: int) -> list[str]:
    """人讀的版本（`addr.py` 直接執行時用）。"""
    out = []
    for r in lookup(addr):
        mark = "⭐" if r["level"] == "confirmed" else "・"
        out.append(f"  {mark} {r['kind']}：{r['desc']}")
        out.append(f"     出處 {r['src']}｜**{r['level']}**")
    return out


def fmt(v: int) -> str:
    return f"-{abs(v):X}h" if v < 0 else f"{v:X}h"


def main() -> int:
    # ⚠️ 不能用 argparse 收位址：組語裡的負偏移長得像 `-6221h`，
    # argparse 會當成未知選項直接報錯。自己解析 argv 最省事。
    argv = sys.argv[1:]
    want_bases = "--bases" in argv
    if "-h" in argv or "--help" in argv:
        print(__doc__)
        return 0
    addrs = [a for a in argv if a != "--bases"]

    class A:
        pass

    args = A()
    args.addrs = addrs
    args.bases = want_bases

    if args.bases or not args.addrs:
        print("## 已知記錄基址\n")
        for base, name, size, src, level in BASES:
            print(f"  {fmt(base):>10}  {name}（{size} B/筆）  {src}｜{level}")
            n = len(FIELDS.get(name, {}))
            print(f"              已解 {n} 個欄位")
        print("\n## 已知全域表與變數\n")
        for base, (desc, src, level) in sorted(GLOBALS.items()):
            print(f"  {fmt(base):>10}  {desc}")
            print(f"              {src}｜{level}")
        if not args.addrs:
            print("\n用法：./tools/py.sh tools/addr.py -6221h 7A8Bh")
            return 0
        print()

    for tok in args.addrs:
        try:
            a = parse(tok)
        except ValueError:
            print(f"{tok}: 解析不了，格式如 -6221h / 7A8Bh / 0x62A8", file=sys.stderr)
            continue
        print(f"{tok}  →  {fmt(norm(a))}")
        lines = explain(a)
        if lines:
            print("\n".join(lines))
        else:
            print("  ✗ 查無。可能是還沒解出的基址，或不屬於任何已知記錄。")
            print("    解出來之後請加進本檔的 BASES／FIELDS／GLOBALS 三張表。")
        print()
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
