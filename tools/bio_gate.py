"""人物小傳的品質閘（`docs/design/21-biography-authoring-spec.md` §6）。

    ./tools/py.sh tools/bio_gate.py batch09
    ./tools/py.sh tools/bio_gate.py --all
    ./tools/py.sh tools/bio_gate.py --cross    # 跨批次：重複 id 與整體進度

四項檢查：格數 ≤ 340、可畫性、`facts.json` 合法且 id 對得上、禁用詞。

## 為什麼從腳本片段升級成工具（2026-08-02）

規格 §6 原本只給一段程式碼片段，讓每個 agent 自己抄。後果有兩個：

1. **每批的實作略有不同**，主迴圈複核時要重寫一次。
2. ⭐ **它只掃正文，不掃會上畫面的欄位**——batch09 自己補跑，
   在 `highest_post`／`faction`／`aliases` 裡抓到 7 個不可畫字元。
   `DESIGN-20` §10 U5 那批簡體殘留**很可能就是這樣漏進來的**。

所以第 2 項改成掃**所有進顯示路徑的欄位**，不只正文。

## ⚠️ 12 字清單不是檢查本身

`DESIGN-20` §2.4 列了 12 個不可畫字，那是**已知的例子，不是窮舉**
——batch09 就撞到第 13 個（`勐`，滇西地名）。

**真正的檢查是 `big5_to_font()` 逐字問倚天字庫**（正對照），
清單只用來讓錯誤訊息更好讀。拿清單當過濾器就是
`~/diagnosis-notes/docs/02-query-returned-empty` 講的「自己的過濾器有洞」。
"""
from __future__ import annotations

import json
import pathlib
import re
import sys
import unicodedata

sys.path.insert(0, "tools")
import eten  # noqa: E402

ROOT = pathlib.Path(__file__).resolve().parent.parent
PEOPLE = ROOT / "docs" / "reference" / "people"

CELL_LIMIT = 340

# 已知的不可畫字，只用來讓訊息好讀。**檢查本身是 big5_to_font()。**
# ⚠️ 這串會長——每一批都可能撞到新的。`礮`／`礟` 是北洋公報的兵種異體字
# （U3 撞到 12 處），`勐` 是滇西地名（batch09）。
KNOWN_BAD = set("〇榘衞鲁啓藁献専継荣羣钟勐礮礟烟")

# §4.3 的禁用詞。機構／事件的官方名稱是誤報，見 ALLOW_IN_NAME。
BANNED = ["剿匪", "漢奸", "賣國", "竊據", "悍將", "名將", "梟雄",
          "東陵大盜", "投誠", "叛將", "偽軍", "匪首", "亂黨"]
# 這些機構／事件的官方名稱含禁用詞，必須原樣保留（規格 §4.3）。
ALLOW_IN_NAME = ["剿匪總司令部", "剿匪總部", "剿匪軍", "剿匪司令", "和平解放"]
# 帶歸屬詞又加引號的引述稱呼，§4.3 列為「限制使用」而非禁止
# ——閘不擋，但列出來讓人複審（要確認不當主語、且過兩來源門檻）。
ATTRIBUTION = ["稱", "號", "人稱", "綽號", "外號", "有「", "謂"]

# 會上遊戲畫面的欄位（`DESIGN-20` §5.7）。`aliases` 是查詢紀錄，不進顯示路徑。
DISPLAY_FIELDS = ("name_ingame", "name_common", "highest_post", "faction", "courtesy")


def cells(t: str) -> float:
    return sum(1.0 if unicodedata.east_asian_width(c) in "WF" else 0.5 for c in t)


def unpaintable(text: str) -> list[str]:
    """回傳這段文字裡畫不出來的字元。"""
    out = []
    for ch in sorted(set(text)):
        if ord(ch) < 128:
            continue
        try:
            b = ch.encode("big5")
        except UnicodeEncodeError:
            out.append(ch)
            continue
        if len(b) != 2 or eten.big5_to_font(b[0], b[1]) is None:
            out.append(ch)
    return out


def banned_hits(text: str) -> tuple[list[str], list[str]]:
    """回傳（該擋的, 該複審的）。

    §4.3 把禁用詞分成兩類，閘也要分兩類——一律當成違規會逼人刪掉
    「贛粵閩湘鄂剿匪軍」這種**原件名稱**，那與保存的目標相反。
    """
    stop, review = [], []
    for w in BANNED:
        for m in re.finditer(re.escape(w), text):
            ctx = text[max(0, m.start() - 8): m.end() + 8]
            if any(a in ctx for a in ALLOW_IN_NAME):
                continue  # 機構／事件的官方名稱，不是評價
            quoted = "「" + w in ctx or w + "」" in ctx
            if quoted and any(a in ctx for a in ATTRIBUTION):
                review.append(f"{w}（…{ctx}…）")
            else:
                stop.append(f"{w}（…{ctx}…）")
    return stop, review


def gate(batch: str) -> int:
    md_path = PEOPLE / f"bios-{batch}.md"
    fj_path = PEOPLE / f"facts-{batch}.json"
    if not md_path.exists() or not fj_path.exists():
        print(f"[{batch}] 找不到 {md_path.name} 或 {fj_path.name}", file=sys.stderr)
        return 1

    fj = json.loads(fj_path.read_text(encoding="utf-8"))
    raw = json.loads((PEOPLE / "people.json").read_text(encoding="utf-8"))
    ppl = {p["id"]: p for p in (raw["people"] if isinstance(raw, dict) else raw)}

    md = md_path.read_text(encoding="utf-8")
    bodies = [(int(m.group(1)), m.group(2).strip())
              for m in re.finditer(r"^## #(\d+) .*?$\n+(.+?)$", md, re.M)]

    bad = 0
    # 1. 格數
    over = [(i, cells(b)) for i, b in bodies if cells(b) > CELL_LIMIT]
    longest = max((cells(b) for _, b in bodies), default=0.0)
    print(f"[{batch}] 1 格數：{len(bodies)} 篇，最長 {longest}"
          + (f"  ⛔ 超標 {over}" if over else "  ✔"))
    bad += len(over)

    # 2. 可畫性——正文 + **所有進顯示路徑的欄位**
    all_text = "".join(b for _, b in bodies)
    field_bad: list[str] = []
    for p in fj["people"]:
        for f in DISPLAY_FIELDS:
            v = p.get(f)
            if isinstance(v, str) and (u := unpaintable(v)):
                field_bad.append(f"#{p['id']}.{f}={v!r} → {u}")
    body_bad = unpaintable(all_text)
    print(f"[{batch}] 2 可畫性：正文相異字 {len(set(all_text))}，"
          f"畫不出 {body_bad or '✔0'}；顯示欄位 {field_bad or '✔0'}")
    print(f"          已知問題字命中：{sorted(KNOWN_BAD & set(all_text)) or '✔0'}")
    bad += len(body_bad) + len(field_bad)

    # 3. JSON 與 id
    #
    # ⛔ 這一項原本寫「篇數 == id 數」，**與規格 §3.2「`unknown` 不寫小傳」
    # 直接衝突**（`G-U2-3`）。前九批母體是 `partial`、10 人 10 篇，
    # 所以從沒浮現；`unknown` 池是第一個必然撞到的母體。
    # 承辦的 agent 正確地選擇回報而不是改我的工具，這裡補上。
    ids = [p["id"] for p in fj["people"]]
    missing = [i for i in ids if i not in ppl]
    mism = [(p["id"], p["name_ingame"], ppl[p["id"]]["name_ingame"])
            for p in fj["people"]
            if p["id"] in ppl and p["name_ingame"] != ppl[p["id"]]["name_ingame"]]

    unknown_ids = {p["id"] for p in fj["people"] if p.get("confidence") == "unknown"}
    want_bodies = len(ids) - len(unknown_ids)
    written = [i for i, _ in bodies if i not in unknown_ids]
    count_ok = len(written) == want_bodies
    # `unknown` 的 `bio_zh` 必須是空的——查不到就是查不到，不准補話。
    fabricated = [p["id"] for p in fj["people"]
                  if p.get("confidence") == "unknown" and (p.get("bio_zh") or "").strip()]
    print(f"[{batch}] 3 id：{len(ids)} 筆（unknown {len(unknown_ids)} 筆不立傳），"
          f"不存在 {missing or '✔無'}，name_ingame 不符 {mism or '✔無'}，"
          f"立傳篇數 {len(written)}/{want_bodies} {'✔' if count_ok else '⛔'}"
          + (f"，⛔ unknown 卻有 bio_zh：{fabricated}" if fabricated else ""))
    bad += len(missing) + len(mism) + (0 if count_ok else 1) + len(fabricated)

    # 4. 禁用詞——分「該擋」與「該複審」兩級
    stop, review = [], []
    for i, b in bodies:
        s_, r_ = banned_hits(b)
        stop += [(i, h) for h in s_]
        review += [(i, h) for h in r_]
    print(f"[{batch}] 4 禁用詞：擋 {stop or '✔無'}")
    if review:
        print(f"          ⚠️ 待複審（帶歸屬詞的引述稱呼，§4.3 限制使用）：{review}")
    bad += len(stop)

    print(f"[{batch}] {'✅ 四項全過' if bad == 0 else f'⛔ {bad} 項不合格'}\n")
    return 0 if bad == 0 else 1


def cross_check() -> int:
    """跨批次檢查：**同一個 id 有沒有被兩批各做一次**，以及全體進度。

    ⛔ 2026-08-02 踩到：派工分群用 `faction` 的「含」字比對，
    `#11 朱培德` 同時含「滇」與「國民革命軍」，於是 C3 與 C4 各做了一次。
    警語擋不住這種事（我在派工單裡明明寫了「不在既有 facts-*.json 裡」），
    所以做成檢查。
    """
    raw = json.loads((PEOPLE / "people.json").read_text(encoding="utf-8"))
    ppl = {p["id"]: p for p in (raw["people"] if isinstance(raw, dict) else raw)}
    owner: dict[int, list[str]] = {}
    conf: dict[str, int] = {}
    for f in sorted(PEOPLE.glob("facts-*.json")):
        try:
            j = json.loads(f.read_text(encoding="utf-8"))
        except json.JSONDecodeError as e:
            print(f"⛔ {f.name} JSON 壞了：{e}")
            return 1
        b = f.stem.replace("facts-", "")
        for p in j.get("people", []):
            owner.setdefault(p["id"], []).append(b)
            conf[p.get("confidence", "?")] = conf.get(p.get("confidence", "?"), 0) + 1

    dups = {i: b for i, b in owner.items() if len(b) > 1}
    print(f"跨批次：已建骨架 {len(owner)} 筆 / 名冊 {len(ppl)} 筆")
    print(f"  自撰 confidence 分佈：{conf}")
    print(f"  ⛔ 重複的 id：{dups}" if dups else "  重複的 id：✔ 無")
    rest = sorted(set(ppl) - set(owner))
    print(f"  未做：{len(rest)} 筆" + (f"（前 10：{rest[:10]}）" if rest else ""))
    return 1 if dups else 0


def main() -> int:
    argv = sys.argv[1:]
    if not argv or argv[0] in ("-h", "--help"):
        print(__doc__)
        return 0
    if argv[0] == "--cross":
        return cross_check()
    if argv[0] == "--all":
        batches = sorted(p.stem.replace("bios-", "") for p in PEOPLE.glob("bios-*.md"))
    else:
        batches = argv
    return max(gate(b) for b in batches)


if __name__ == "__main__":
    raise SystemExit(main())
