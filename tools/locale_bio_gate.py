#!/usr/bin/env python3
"""驗證英／日人物自傳 overlay 的來源、數量、雜湊與審稿狀態。

這個 gate 不翻譯正文，也不寫回 people.json。它只讓正式流程可重現：
研究母稿 → locale overlay → 預覽／審稿 → 發行。machine-draft 可以通過預覽
模式；加上 --release 時必須是 human-reviewed，避免把模型初稿冒充正式譯稿。
"""

from __future__ import annotations

import argparse
import hashlib
import json
import sys
from pathlib import Path


EXPECTED_TRANSLATED = 387
ALLOWED_LANGUAGES = {"en", "ja"}
ALLOWED_STATUSES = {"machine-draft", "human-reviewed"}


def read_json(path: Path):
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except OSError as exc:
        raise SystemExit(f"讀取 {path} 失敗：{exc}") from exc
    except json.JSONDecodeError as exc:
        raise SystemExit(f"解析 {path} 失敗：{exc}") from exc


def rows(data, path: Path) -> list[dict]:
    value = data if isinstance(data, list) else data.get("people")
    if not isinstance(value, list) or not all(isinstance(row, dict) for row in value):
        raise SystemExit(f"{path} 的 people 必須是物件陣列")
    return value


def sha256_text(value: str) -> str:
    return hashlib.sha256(value.encode("utf-8")).hexdigest()


def sha256_file(path: Path) -> str:
    digest = hashlib.sha256()
    try:
        with path.open("rb") as stream:
            for chunk in iter(lambda: stream.read(1024 * 1024), b""):
                digest.update(chunk)
    except OSError as exc:
        raise SystemExit(f"讀取 generated_from 來源 {path} 失敗：{exc}") from exc
    return digest.hexdigest()


def source_file(root: Path, value: str) -> Path:
    """把 overlay 的相對來源解析到 repo 內，拒絕絕對路徑與跳出 root。"""
    candidate = Path(value)
    if candidate.is_absolute():
        raise SystemExit(f"generated_from 不得使用絕對路徑：{value}")
    path = (root / candidate).resolve()
    try:
        path.relative_to(root.resolve())
    except ValueError as exc:
        raise SystemExit(f"generated_from 路徑跳出專案：{value}") from exc
    if not path.is_file():
        raise SystemExit(f"generated_from 找不到檔案：{value}")
    return path


def expected_sources(root: Path) -> dict[int, tuple[str, str]]:
    base_path = root / "translations" / "zh-Hant" / "people.json"
    overlay_path = root / "translations" / "zh-Hant" / "people-authored.json"
    base = rows(read_json(base_path), base_path)
    authored = rows(read_json(overlay_path), overlay_path)
    result: dict[int, tuple[str, str]] = {}
    for person in base:
        person_id = person.get("id")
        bio = person.get("bio_zh", "")
        if isinstance(person_id, int) and isinstance(bio, str) and bio.strip():
            result[person_id] = (person.get("name_ingame", ""), bio)
    for person in authored:
        person_id = person.get("id")
        bio = person.get("bio_zh", "")
        if not isinstance(person_id, int) or not isinstance(bio, str) or not bio.strip():
            raise SystemExit("繁中 authored overlay 含無效或空白正文")
        if person_id in result:
            raise SystemExit(f"繁中 authored overlay 重複既有正文 id={person_id}")
        result[person_id] = (person.get("name_ingame", ""), bio)
    if len(result) != EXPECTED_TRANSLATED:
        raise SystemExit(
            f"繁中研究正文基數應為 {EXPECTED_TRANSLATED}，實際 {len(result)}"
        )
    return result


def validate(root: Path, locale: str, overlay_path: Path, release: bool) -> None:
    if locale not in ALLOWED_LANGUAGES:
        raise SystemExit(f"不支援的 locale：{locale}")
    data = read_json(overlay_path)
    if not isinstance(data, dict):
        raise SystemExit("overlay 頂層必須是物件")
    if data.get("schema_version") != "1":
        raise SystemExit("schema_version 必須是 1")
    if data.get("language") != locale:
        raise SystemExit("overlay language 與 --locale 不一致")
    if data.get("overlay_scope") != "translated-biographies":
        raise SystemExit("overlay_scope 必須是 translated-biographies")
    status = data.get("translation_status")
    if status not in ALLOWED_STATUSES:
        raise SystemExit(f"translation_status 無效：{status!r}")
    generated = data.get("generated_from")
    if not isinstance(generated, list) or not generated:
        raise SystemExit("generated_from 必須是非空陣列")
    for source in generated:
        if not isinstance(source, dict):
            raise SystemExit("generated_from 含非物件項目")
        for key in ("facts", "bios"):
            if not isinstance(source.get(key), str) or not source[key].strip():
                raise SystemExit(f"generated_from 缺少有效 {key} 路徑")
        for key in ("facts_sha256", "bios_sha256"):
            if not isinstance(source.get(key), str) or len(source[key]) != 64:
                raise SystemExit(f"generated_from 缺少有效 {key}")
        facts_path = source_file(root, source["facts"])
        bios_path = source_file(root, source["bios"])
        if sha256_file(facts_path) != source["facts_sha256"]:
            raise SystemExit(f"generated_from facts SHA-256 不一致：{source['facts']}")
        if sha256_file(bios_path) != source["bios_sha256"]:
            raise SystemExit(f"generated_from bios SHA-256 不一致：{source['bios']}")

    expected = expected_sources(root)
    people = data.get("people")
    if not isinstance(people, list) or len(people) != EXPECTED_TRANSLATED:
        raise SystemExit(f"翻譯人物應有 {EXPECTED_TRANSLATED} 筆")
    previous = 0
    seen: set[int] = set()
    for row in people:
        person_id = row.get("id")
        if not isinstance(person_id, int) or person_id <= previous or person_id in seen:
            raise SystemExit(f"人物 ID 未遞增或重複：{person_id!r}")
        previous, seen = person_id, seen | {person_id}
        source = expected.get(person_id)
        if source is None:
            raise SystemExit(f"overlay 指向非 387 篇成文人物 id={person_id}")
        name, source_bio = source
        if row.get("name_ingame") != name:
            raise SystemExit(f"id={person_id} name_ingame 與繁中母稿不一致")
        bio = row.get("bio")
        if not isinstance(bio, str) or not bio.strip():
            raise SystemExit(f"id={person_id} 翻譯正文為空")
        if row.get("bio_language") != locale:
            raise SystemExit(f"id={person_id} bio_language 不一致")
        if row.get("bio_status") != status:
            raise SystemExit(f"id={person_id} bio_status 不一致")
        if row.get("source_bio_sha256") != sha256_text(source_bio):
            raise SystemExit(f"id={person_id} source_bio_sha256 不一致")
        if row.get("translated_bio_sha256") != sha256_text(bio):
            raise SystemExit(f"id={person_id} translated_bio_sha256 不一致")
        review = row.get("review")
        if not isinstance(review, dict):
            raise SystemExit(f"id={person_id} 缺少 review metadata")
        review_status = review.get("status")
        if status == "machine-draft" and review_status not in {"unreviewed", "in-review"}:
            raise SystemExit(f"id={person_id} machine-draft review status 無效")
        if status == "human-reviewed":
            if review_status != "human-reviewed":
                raise SystemExit(f"id={person_id} 尚未標成 human-reviewed")
            if not review.get("reviewer") or not review.get("reviewed_at"):
                raise SystemExit(f"id={person_id} 缺少 reviewer／reviewed_at")
        if release and status != "human-reviewed":
            raise SystemExit("正式發行 gate 只接受 human-reviewed")
    print(f"locale bio gate 通過：{locale} {status} {len(people)} 篇")


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--root", type=Path, default=Path(__file__).resolve().parents[1])
    parser.add_argument("--locale", required=True, choices=sorted(ALLOWED_LANGUAGES))
    parser.add_argument("--overlay", type=Path, required=True)
    parser.add_argument("--release", action="store_true")
    args = parser.parse_args()
    validate(args.root, args.locale, args.overlay, args.release)
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except SystemExit:
        raise
    except Exception as exc:  # pragma: no cover - 最外層 fail-closed 防護
        print(f"locale bio gate 未預期失敗：{exc}", file=sys.stderr)
        raise SystemExit(1) from exc
