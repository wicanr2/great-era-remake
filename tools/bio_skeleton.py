"""產出**成文階段可以看的骨架**（`docs/design/21-biography-authoring-spec.md` §23）。

    ./tools/py.sh tools/bio_skeleton.py batchC4 > /tmp/skel.json
    ./tools/py.sh tools/bio_skeleton.py batchC4 --audit    # 只報會被拿掉什麼

## 為什麼要這個

規格 §0 的著作權隔離建立在一句話上：

> **骨架是欄位，欄位沒有表達；從欄位長出來的句子是新表達。**

⛔ **這句話對兩個欄位不成立**（batchC4 交件時指出，`G-C4-2`）：

- `note` 是**自由文字**，實測 2,589 個裡有 59% 超過 40 字，
  而且同時裝著三種東西：schema 沒欄位可放的事實、裁決紀錄、
  **以及來源的原話片段**。
- `conflicts[].statements[].says` 規格 §2.1 就**明訂允許是句子**
  （要原樣記錄來源的說法）。

也就是說：成文階段「只看骨架」時，**仍然看得到原文的表達**。
承辦者只能靠自律迴避——**而 §0 的整個設計就是為了不靠自律**。

## 這支做什麼

輸出一份**拿掉表達性欄位**的骨架，讓成文階段**物理上看不到**原話：

| 欄位 | 處置 |
|---|---|
| `note_fact` | **保留**——新 schema 的「事實補充」欄（§23） |
| `note_source` | 拿掉——新 schema 的「來源原話／裁決紀錄」欄 |
| `note`（舊） | **拿掉**，並在輸出裡留一筆計數，讓寫的人知道有東西被扣住 |
| `conflicts[].statements[].says` | 拿掉，只留 `src` 與 `conf` |

⚠️ **拿掉不等於刪除**：原檔一個字都不動，這支只產出**視圖**。
`note` 裡那些「沒有欄位可放的事實」要靠 §23 的 schema 遷移救回來，
不是靠成文階段偷看。
"""
from __future__ import annotations

import json
import pathlib
import sys

ROOT = pathlib.Path(__file__).resolve().parent.parent
PEOPLE = ROOT / "docs" / "reference" / "people"

# 可能帶著來源表達的欄位，成文階段看不到。
STRIP = ("note", "note_source", "says")
# 明確保留的自由文字欄位（新 schema）。
KEEP = ("note_fact",)


def scrub(obj, stats: dict[str, int]):
    """遞迴拿掉表達性欄位，回傳新結構。原物件不動。"""
    if isinstance(obj, dict):
        out = {}
        for k, v in obj.items():
            if k in KEEP:
                out[k] = v
                continue
            if k in STRIP:
                if isinstance(v, str) and v.strip():
                    stats[k] = stats.get(k, 0) + 1
                continue
            out[k] = scrub(v, stats)
        return out
    if isinstance(obj, list):
        return [scrub(x, stats) for x in obj]
    return obj


def main() -> int:
    argv = sys.argv[1:]
    if not argv or argv[0] in ("-h", "--help"):
        print(__doc__)
        return 0
    batch = argv[0]
    audit = "--audit" in argv

    path = PEOPLE / f"facts-{batch}.json"
    if not path.exists():
        print(f"找不到 {path}", file=sys.stderr)
        return 1
    data = json.loads(path.read_text(encoding="utf-8"))

    stats: dict[str, int] = {}
    view = scrub(data, stats)
    view["$writing_view"] = (
        "成文階段專用視圖：已拿掉可能帶原文表達的欄位"
        f"（{stats or '無'}）。原檔未變動。詳見 docs/design/21 §23。"
    )

    if audit:
        print(f"[{batch}] 會被拿掉的欄位：{stats or '✔ 無'}")
        n = sum(stats.values())
        print(f"[{batch}] 合計 {n} 處" + ("——成文階段看不到這些。" if n else ""))
        return 0

    json.dump(view, sys.stdout, ensure_ascii=False, indent=2)
    print()
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
