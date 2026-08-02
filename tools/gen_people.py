"""產生人物自傳的執行期語系資料與期別／槽位接合表。

輸入是研究資料，只讀：
  docs/reference/people/people.json
  docs/reference/people/00-roster.md

輸出是可發行資料：
  translations/shared/roster-slots.json
  translations/zh-Hant/people.json
  translations/zh-Hant/people-unmapped.md

用法：tools/py.sh tools/gen_people.py
"""

from __future__ import annotations

import json
import re
from collections import Counter
from pathlib import Path

PEOPLE = Path("docs/reference/people/people.json")
ROSTER = Path("docs/reference/people/00-roster.md")
SLOTS_OUT = Path("translations/shared/roster-slots.json")
PEOPLE_OUT = Path("translations/zh-Hant/people.json")
UNMAPPED_OUT = Path("translations/zh-Hant/people-unmapped.md")

PERIODS = {"北伐時期": 1, "抗戰前期": 2, "抗日後期": 3}
EXPECTED = {1: 274, 2: 106, 3: 106}
ROW = re.compile(r"^\|\s*(\d+)\s*\|\s*(.*?)\s*\|\s*(.*?)\s*\|$")
APPEARANCE = re.compile(r"(北伐時期|抗戰前期|抗日後期)#(\d+)")

# DESIGN-20 §5.4 已裁決的乾淨正規化。只作用於發行語系檔。
NORMALIZE = {
    "鲁": "魯", "献": "獻", "専": "專", "継": "繼", "荣": "榮", "钟": "鐘",
    "衞": "衛", "啓": "啟", "羣": "群", "〇": "零",
}
# 沒有使用者裁決的字保留原樣，不能安靜改成形近字。
UNRESOLVED = {"榘", "藁"}


def normalize(value, changed: Counter, unresolved: Counter):
    if isinstance(value, str):
        out = []
        for ch in value:
            if ch in NORMALIZE:
                changed[(ch, NORMALIZE[ch])] += 1
                out.append(NORMALIZE[ch])
            else:
                if ch in UNRESOLVED:
                    unresolved[ch] += 1
                out.append(ch)
        return "".join(out)
    if isinstance(value, list):
        return [normalize(v, changed, unresolved) for v in value]
    if isinstance(value, dict):
        return {k: normalize(v, changed, unresolved) for k, v in value.items()}
    return value


def main() -> int:
    people = json.loads(PEOPLE.read_text("utf-8"))
    by_id = {p["id"]: p for p in people}
    if sorted(by_id) != list(range(1, 418)):
        raise SystemExit("people.json 的 id 必須連續為 1..417")

    slots = {1: {}, 2: {}, 3: {}}
    rows = 0
    for line in ROSTER.read_text("utf-8").splitlines():
        m = ROW.match(line)
        if not m:
            continue
        person_id = int(m.group(1))
        if person_id not in by_id:
            continue
        rows += 1
        name, appearances = m.group(2).strip(), m.group(3)
        if name != by_id[person_id]["name_ingame"]:
            raise SystemExit(
                f"名冊 #{person_id} 是 {name!r}，people.json 是 "
                f"{by_id[person_id]['name_ingame']!r}")
        found = APPEARANCE.findall(appearances)
        if not found:
            raise SystemExit(f"名冊 #{person_id} 沒有期別槽位")
        for label, raw_slot in found:
            period, slot = PERIODS[label], int(raw_slot)
            if slot in slots[period]:
                raise SystemExit(
                    f"期別 {period} 槽 {slot} 同時指向 {slots[period][slot]} 與 {person_id}")
            slots[period][slot] = person_id
    if rows != 417:
        raise SystemExit(f"名冊解析到 {rows} 人，預期 417")
    for period, count in EXPECTED.items():
        actual = sorted(slots[period])
        if actual != list(range(1, count + 1)):
            missing = sorted(set(range(1, count + 1)) - set(actual))
            raise SystemExit(f"期別 {period} 槽位不完整：{len(actual)}/{count}，缺 {missing}")

    slot_doc = {
        "periods": {
            str(period): [
                {"slot": slot, "person": slots[period][slot]}
                for slot in range(1, EXPECTED[period] + 1)
            ]
            for period in (1, 2, 3)
        },
        "excluded": [
            {"period": 1, "slot": 274, "person": 274,
             "reason": "placeholder-no-governor"}
        ],
        # Identity 關係是資料，不寫死在 Go 查詢邏輯裡。
        "same_person": [[4, 290], [50, 289], [99, 399]],
    }

    changed, unresolved = Counter(), Counter()
    localized = [normalize(p, changed, unresolved) for p in people]
    people_doc = {
        "language": "zh-Hant",
        "source": "docs/reference/people/people.json（事實重新撰寫；非執行期直接讀研究檔）",
        "people": localized,
    }

    SLOTS_OUT.parent.mkdir(parents=True, exist_ok=True)
    PEOPLE_OUT.parent.mkdir(parents=True, exist_ok=True)
    SLOTS_OUT.write_text(json.dumps(slot_doc, ensure_ascii=False, indent=2) + "\n", "utf-8")
    PEOPLE_OUT.write_text(json.dumps(people_doc, ensure_ascii=False, indent=2) + "\n", "utf-8")

    ledger = [
        "# 人物語系資料字元正規化帳本", "",
        "> 由 `tools/gen_people.py` 產生；研究資料不修改。", "",
        "## 已正規化", "", "| 原字 | 輸出字 | 次數 |", "|---|---|---:|",
    ]
    for (old, new), count in sorted(changed.items()):
        ledger.append(f"| {old} | {new} | {count} |")
    ledger += ["", "## 尚未裁決（原樣保留）", "", "| 字 | 次數 |", "|---|---:|"]
    for char, count in sorted(unresolved.items()):
        ledger.append(f"| {char} | {count} |")
    ledger += ["", "`榘`（韓復榘）與 `藁`（藁城）沒有經裁決的 Big5 替代；不得自行改成形近字。", ""]
    UNMAPPED_OUT.write_text("\n".join(ledger), "utf-8")

    print(f"寫入 {SLOTS_OUT}：486 槽，排除 1")
    print(f"寫入 {PEOPLE_OUT}：{len(localized)} 筆，bio_zh {sum(bool(p.get('bio_zh')) for p in localized)} 篇")
    print(f"正規化 {sum(changed.values())} 字；未裁決 {sum(unresolved.values())} 字")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
