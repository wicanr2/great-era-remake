"""從 `MAN(N).DAT` + `MAN{N}15` 產出「遊戲內的所屬省與效忠領袖」對照表。

為什麼要這個（2026-08-02）：

`people.json` 裡 91 筆 `unknown` 的 `faction` 與 `highest_post` **全部空白**
——那是外部查證的結果，查不到就沒有。但**遊戲資料本身知道**：
將領記錄的 `+4` 是所屬省、`+14` 是效忠的勢力領袖，兩格都是 confirmed 欄位。

⚠️ 這是 `S-ROSTER`（規格 §11.6）：**遊戲作者的編排，不是史料**。
只能當 identity 佐證與查證起點，**不得寫進正文**當史實陳述。

    ./tools/py.sh tools/ingame_affiliation.py            # 全部 274 筆
    ./tools/py.sh tools/ingame_affiliation.py --unknown  # 只列 people.json 的 unknown
"""
from __future__ import annotations
import json, pathlib, sys, collections
sys.path.insert(0, "tools")
import eten

ROOT = pathlib.Path(__file__).resolve().parent.parent
REC, NAME_SLOT = 33, 3


def load(stage: int):
    F = eten.EtenFonts(ROOT / "workplace" / "eten")
    man = (ROOT / f"workplace/orig/game/MAN({stage}).DAT").read_bytes()
    names = (ROOT / f"workplace/orig/game/MAN{stage}15").read_bytes()
    rev = F.reverse

    def name(idx: int) -> str:
        if idx <= 0:
            return "—"
        out = ""
        for k in range(NAME_SLOT):
            g = names[((idx - 1) * NAME_SLOT + k) * 30: ((idx - 1) * NAME_SLOT + k + 1) * 30]
            hit = rev.get(g)
            if hit:
                out += hit[0]
        # 兩字名用全形空白補到三格，還原成人類讀的寫法
        return out.replace("　", "").strip() or f"#{idx}"

    return man, name


def rows(stage: int, count: int):
    man, name = load(stage)
    out = []
    for i in range(1, count + 1):
        r = man[(i - 1) * REC: i * REC]
        out.append({
            "id": i,
            "name": name(i),
            "province": r[4],
            "leader_id": int.from_bytes(r[14:16], "little"),
        })
    for o in out:
        o["leader"] = name(o["leader_id"])
    return out


def main() -> int:
    only_unknown = "--unknown" in sys.argv
    data = rows(1, 274)
    if only_unknown:
        raw = json.loads((ROOT / "docs/reference/people/people.json").read_text(encoding="utf-8"))
        ppl = raw["people"] if isinstance(raw, dict) else raw
        unk = {p["id"] for p in ppl if p.get("confidence") == "unknown"}
        data = [d for d in data if d["id"] in unk]
    c = collections.Counter(d["leader"] for d in data)
    print(f"# 遊戲內的效忠領袖分佈（{len(data)} 人）\n")
    print("| 效忠領袖 | 人數 |")
    print("|---|---:|")
    for k, n in c.most_common():
        print(f"| {k} | {n} |")
    print("\n| id | 遊戲寫法 | 所屬省 | 效忠領袖 |")
    print("|---:|---|---:|---|")
    for d in data:
        print(f"| {d['id']} | {d['name']} | {d['province']} | {d['leader']} |")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
