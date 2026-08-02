#!/usr/bin/env python3
"""由版控研究批次建立可追溯的人物自傳索引。

預設只驗證並印出摘要，不寫入產品資料：

    python tools/gen_authored_bios.py --check
    python tools/gen_authored_bios.py --output /tmp/authored-bios.json

DESIGN-22 仍是 DRAFT，因此本工具刻意沒有直接改寫 people.json 或 translations 的
模式。輸出只包含自傳、信心度與來源雜湊，供裁決後的合併器使用。
"""

from __future__ import annotations

import argparse
import hashlib
import json
import pathlib
import re
import sys

ROOT = pathlib.Path(__file__).resolve().parent.parent
PEOPLE_DIR = ROOT / "docs" / "reference" / "people"
PEOPLE_PATH = PEOPLE_DIR / "people.json"

sys.path.insert(0, str(ROOT / "tools"))
import bio_gate  # noqa: E402

HEADER = re.compile(r"^## #(\d+)\s+(.+?)\s+—\s+`(high|medium|low|unknown)`(?:.*)?$")


class InputError(ValueError):
    """研究批次不符合失敗即關閉契約。"""


def sha256(path: pathlib.Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def load_people() -> dict[int, dict]:
    raw = json.loads(PEOPLE_PATH.read_text("utf-8"))
    rows = raw["people"] if isinstance(raw, dict) else raw
    by_id = {row["id"]: row for row in rows}
    if len(rows) != 417 or sorted(by_id) != list(range(1, 418)):
        raise InputError("people.json 必須恰有連續的 1..417，且 id 不得重複")
    return by_id


def parse_sections(path: pathlib.Path) -> dict[int, dict]:
    """解析每個人物標題及緊接標題的正文段落。

    `unknown` 仍有標題與稽核說明，但不得有正文；項目符號、引用及表格都視為
    稽核資料，不會被誤收成遊戲文字。
    """
    lines = path.read_text("utf-8").splitlines()
    sections: dict[int, dict] = {}
    i = 0
    while i < len(lines):
        if not lines[i].startswith("## #"):
            i += 1
            continue
        header_line = lines[i]
        match = HEADER.fullmatch(header_line)
        if match is None:
            raise InputError(f"{path.name}:{i + 1} 無法解析人物標題：{lines[i]!r}")
        person_id, name, confidence = int(match.group(1)), match.group(2), match.group(3)
        if person_id in sections:
            raise InputError(f"{path.name} 的 #{person_id} 標題重複")
        i += 1
        block: list[str] = []
        while i < len(lines) and not lines[i].startswith("## "):
            block.append(lines[i])
            i += 1
        body = ""
        if confidence == "unknown":
            audit = header_line + "\n" + "\n".join(block)
            if not any(mark in audit for mark in ("不立傳", "未寫小傳", "不寫小傳", "bio_zh` 留空")):
                raise InputError(f"{path.name}:{i + 1} 的 unknown 小節沒有明示不立傳")
        else:
            for line in block:
                stripped = line.strip()
                if not stripped or stripped.startswith(("-", ">", "|", "*")):
                    continue
                body = stripped
                break
        sections[person_id] = {
            "name_ingame": name,
            "confidence": confidence,
            "bio_zh": body,
        }
    return sections


def build_document() -> tuple[dict, dict]:
    canonical = load_people()
    facts_files = sorted(PEOPLE_DIR.glob("facts-*.json"))
    facts_by_id: dict[int, tuple[pathlib.Path, pathlib.Path, dict]] = {}
    authored: list[dict] = []

    for facts_path in facts_files:
        batch = facts_path.stem.removeprefix("facts-")
        bios_path = PEOPLE_DIR / f"bios-{batch}.md"
        if not bios_path.exists():
            raise InputError(f"{facts_path.name} 沒有成對的 {bios_path.name}")
        facts_doc = json.loads(facts_path.read_text("utf-8"))
        sections = parse_sections(bios_path)
        facts_ids = {row["id"] for row in facts_doc.get("people", [])}
        extra = sorted(set(sections) - facts_ids)
        if extra:
            raise InputError(f"{bios_path.name} 有不屬於 facts 的人物：{extra}")

        for row in facts_doc.get("people", []):
            person_id = row["id"]
            if person_id in facts_by_id:
                old = facts_by_id[person_id][0]
                raise InputError(f"#{person_id} 同時屬於 {old.name} 與 {facts_path.name}")
            if person_id not in canonical:
                raise InputError(f"{facts_path.name} 的 #{person_id} 不在 people.json")
            name = row.get("name_ingame")
            if name != canonical[person_id]["name_ingame"]:
                raise InputError(
                    f"{facts_path.name} 的 #{person_id} 姓名 {name!r} 與 people.json 不同"
                )
            confidence = row.get("confidence")
            if confidence not in {"high", "medium", "low", "unknown"}:
                raise InputError(f"{facts_path.name} 的 #{person_id} confidence 非法：{confidence!r}")
            section = sections.get(person_id)
            if person_id == 274:
                if (
                    name != "無省長"
                    or confidence != "unknown"
                    or row.get("identity", {}).get("type") != "placeholder"
                    or section is not None
                ):
                    raise InputError("#274 必須是沒有成文標題的『無省長』placeholder")
                facts_by_id[person_id] = (facts_path, bios_path, row)
                continue
            if section is None:
                raise InputError(f"{bios_path.name} 缺 #{person_id} 標題")
            heading_name = section["name_ingame"]
            # 少數標題在原版名後用全形括號補「通行寫法」等導覽註記；原版名本身
            # 仍須是完整前綴，不能用 startswith 放過另一個較長姓名。
            name_ok = heading_name == name or (
                heading_name.startswith(name + "（") and heading_name.endswith("）")
            )
            if not name_ok or section["confidence"] != confidence:
                raise InputError(
                    f"{bios_path.name} 的 #{person_id} 標題 {heading_name!r}/"
                    f"{section['confidence']!r} 與 facts {name!r}/{confidence!r} 不同"
                )
            body = section["bio_zh"]
            if confidence == "unknown":
                if body or (row.get("bio_zh") or "").strip():
                    raise InputError(f"#{person_id} 是 unknown，卻有正文")
            else:
                if not body:
                    raise InputError(f"#{person_id} 不是 unknown，卻沒有正文")
                if bio_gate.cells(body) > bio_gate.CELL_LIMIT:
                    raise InputError(f"#{person_id} 正文超過 {bio_gate.CELL_LIMIT} 格")
                if bad := bio_gate.unpaintable(body):
                    raise InputError(f"#{person_id} 正文含倚天畫不出的字：{bad}")
                stop, _review = bio_gate.banned_hits(body)
                if stop:
                    raise InputError(f"#{person_id} 正文含禁止詞：{stop}")
                authored.append({
                    "id": person_id,
                    "name_ingame": name,
                    "bio_zh": body,
                    "confidence": confidence,
                    "provenance": {
                        "facts": facts_path.relative_to(ROOT).as_posix(),
                        "facts_sha256": sha256(facts_path),
                        "bios": bios_path.relative_to(ROOT).as_posix(),
                        "bios_sha256": sha256(bios_path),
                        "bio_sha256": hashlib.sha256(body.encode("utf-8")).hexdigest(),
                    },
                })
            facts_by_id[person_id] = (facts_path, bios_path, row)

    if sorted(facts_by_id) != list(range(1, 418)):
        missing = sorted(set(range(1, 418)) - set(facts_by_id))
        raise InputError(f"研究骨架未完整覆蓋 1..417，缺 {missing}")
    authored.sort(key=lambda row: row["id"])
    if len(authored) != 387:
        raise InputError(f"自撰正文應為 387 篇，實際 {len(authored)}")

    old_nonempty = {i for i, row in canonical.items() if (row.get("bio_zh") or "").strip()}
    authored_ids = {row["id"] for row in authored}
    newly_available = sorted(authored_ids - old_nonempty)
    unknown = sorted(set(canonical) - authored_ids)
    if len(old_nonempty) != 326 or len(newly_available) != 61 or len(unknown) != 30:
        raise InputError(
            f"基準漂移：舊正文 {len(old_nonempty)}、新增 {len(newly_available)}、unknown {len(unknown)}"
        )

    document = {
        "schema_version": "1",
        "generated_from": "versioned research batches",
        "people": authored,
    }
    summary = {
        "facts": len(facts_by_id),
        "authored": len(authored),
        "unknown": len(unknown),
        "runtime_existing": len(old_nonempty),
        "newly_available": newly_available,
    }
    return document, summary


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--check", action="store_true", help="只驗證（預設行為）")
    parser.add_argument("--output", type=pathlib.Path, help="明確寫出索引；不得指向產品 people.json")
    args = parser.parse_args()
    try:
        document, summary = build_document()
    except (InputError, json.JSONDecodeError) as error:
        print(f"⛔ {error}", file=sys.stderr)
        return 1

    if args.output:
        forbidden = {
            PEOPLE_PATH.resolve(),
            (ROOT / "translations" / "zh-Hant" / "people.json").resolve(),
        }
        if args.output.resolve() in forbidden:
            print("⛔ DESIGN-22 尚未 READY，不得直接覆寫產品人物資料", file=sys.stderr)
            return 1
        args.output.parent.mkdir(parents=True, exist_ok=True)
        args.output.write_text(json.dumps(document, ensure_ascii=False, indent=2) + "\n", "utf-8")
        print(f"寫出 {args.output}")
    print(
        f"✅ 骨架 {summary['facts']}、自撰 {summary['authored']}、unknown {summary['unknown']}、"
        f"執行期既有 {summary['runtime_existing']}、待整合 {len(summary['newly_available'])}"
    )
    print("待整合 id：" + ",".join(map(str, summary["newly_available"])))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
