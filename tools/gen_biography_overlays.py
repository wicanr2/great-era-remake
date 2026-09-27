#!/usr/bin/env python3
"""產生可追溯的英／日人物自傳 machine-draft overlay。

本工具不呼叫外部翻譯服務，也不把研究正文送出專案；它使用固定的片語詞表、
日期規則與人物事實欄位，產生可重跑的結構化初稿。未命中的歷史專名保留原文，
並在每篇末尾明示 machine-draft，避免把 deterministic 詞彙代換冒充人工譯稿。
"""

from __future__ import annotations

import argparse
import hashlib
import json
import re
from pathlib import Path


# 使用者決策（2026-09-27）：刪除逐篇人審 gate，英／日譯稿以 wiki 研究母稿
# 為主，狀態即為 wiki-sourced，可直接進正式發行鏈。
TRANSLATION_STATUS = "wiki-sourced"

# 由長到短套用；這些詞只描述常見制度／軍事用語，不覆蓋人物姓名。
EN_PHRASES = {
    "生卒年與籍貫不詳": "birth and birthplace are unknown",
    "生卒年不詳": "the birth and death years are unknown",
    "籍貫不詳": "the birthplace is unknown",
    "其餘經歷未查得記載": "no further record has been located",
    "此後行止未查得記載": "later movements are not recorded",
    "是否同一人待考": "whether this was the same person remains unconfirmed",
    "以此寫法未查得任何記載": "no record was found under this spelling",
    "另據": "another record states that",
    "據": "according to",
    "中文維基百科": "Chinese Wikipedia",
    "中華民國政府官職資料庫": "the Republic of China government office database",
    "國民革命軍": "National Revolutionary Army",
    "國民政府": "Nationalist government",
    "國民聯軍": "National United Army",
    "中國國民黨": "Kuomintang",
    "中國共產黨": "Chinese Communist Party",
    "中國人民解放軍": "People's Liberation Army",
    "中國戰區": "China Theater",
    "抗日戰爭": "War of Resistance against Japan",
    "抗日戰爭全面爆發": "the War of Resistance against Japan broke out nationwide",
    "抗戰期間": "during the War of Resistance",
    "北伐時期": "the Northern Expedition period",
    "北伐": "Northern Expedition",
    "中原大戰": "Central Plains War",
    "辛亥革命": "Xinhai Revolution",
    "第二次直奉戰爭": "Second Zhili–Fengtian War",
    "軍閥": "warlord",
    "軍隊": "army",
    "部隊": "unit",
    "所部": "his forces",
    "所率派系": "the faction he led",
    "軍長": "army commander",
    "總司令": "commander-in-chief",
    "司令長官": "commander",
    "司令": "commander",
    "參謀總長": "chief of the general staff",
    "參謀長": "chief of staff",
    "參謀": "staff officer",
    "軍事委員會": "Military Commission",
    "軍事委員會委員長": "chairman of the Military Commission",
    "行政院院長": "president of the Executive Yuan",
    "國民政府主席": "chairman of the Nationalist government",
    "總統": "president",
    "校長": "president of the school",
    "黃埔軍校": "Whampoa Military Academy",
    "軍校": "military academy",
    "陸軍": "Army",
    "步兵": "infantry",
    "騎兵": "cavalry",
    "炮兵": "artillery",
    "砲兵": "artillery",
    "裝甲": "armour",
    "師長": "division commander",
    "旅長": "brigade commander",
    "團長": "regimental commander",
    "營長": "battalion commander",
    "軍官": "officer",
    "少將": "major general",
    "中將": "lieutenant general",
    "上將": "general",
    "上校": "colonel",
    "中校": "lieutenant colonel",
    "少校": "major",
    "參與": "participated in",
    "參加": "took part in",
    "率部": "led his forces",
    "任": "served as",
    "歷任": "held the posts of",
    "曾任": "once served as",
    "出任": "became",
    "兼任": "also served as",
    "改編": "reorganized",
    "改編為": "was reorganized as",
    "駐守": "garrisoned",
    "駐紮": "was stationed in",
    "指揮": "commanded",
    "作戰": "combat operations",
    "作戰指揮": "operational command",
    "戰役": "campaign",
    "戰鬥": "battle",
    "會戰": "campaign",
    "攻占": "captured",
    "攻佔": "captured",
    "被俘": "was captured",
    "投降": "surrender",
    "起義": "uprising",
    "倒戈": "changed sides",
    "敗退": "retreated after defeat",
    "戰敗": "was defeated",
    "病逝": "died of illness",
    "身故": "died",
    "逝世": "died",
    "離開": "left",
    "成立": "was established",
    "建立": "established",
    "推動": "promoted",
    "領導": "led",
    "宣布": "announced",
    "提出": "proposed",
    "返回": "returned",
    "赴": "went to",
    "返國": "returned to China",
    "其後": "afterwards",
    "同年": "in the same year",
    "翌年": "the following year",
    "期間": "during this period",
    "長期": "for a long period",
    "逐步": "gradually",
    "一度": "for a time",
    "持續": "continued",
    "年任": "served in",
    "月任": "served in",
    "日任": "served on",
    "人稱": "was known as",
    "著有": "wrote",
    "出生": "was born",
    "生於": "was born in",
    "人。": ".",
    "。": ". ",
    "，": ", ",
    "、": ", ",
    "；": "; ",
    "：": ": ",
    "（": " (",
    "）": ") ",
    "「": '"',
    "」": '"',
}

JA_PHRASES = {
    "生卒年與籍貫不詳": "生没年・出身地不詳",
    "生卒年不詳": "生没年不詳",
    "籍貫不詳": "出身地不詳",
    "其餘經歷未查得記載": "その他の経歴は確認できない",
    "此後行止未查得記載": "その後の動向は記録されていない",
    "是否同一人待考": "同一人物かは未確認である",
    "以此寫法未查得任何記載": "この表記では記録を確認できない",
    "另據": "別資料によれば",
    "據": "によれば",
    "中文維基百科": "中国語版ウィキペディア",
    "中華民國政府官職資料庫": "中華民国政府官職データベース",
    "國民革命軍": "国民革命軍",
    "國民政府": "国民政府",
    "國民聯軍": "国民連軍",
    "中國國民黨": "中国国民党",
    "中國共產黨": "中国共産党",
    "中國人民解放軍": "中国人民解放軍",
    "中國戰區": "中国戦区",
    "抗日戰爭": "抗日戦争",
    "北伐時期": "北伐期",
    "北伐": "北伐",
    "中原大戰": "中原大戦",
    "辛亥革命": "辛亥革命",
    "軍閥": "軍閥",
    "部隊": "部隊",
    "所部": "その部隊",
    "軍長": "軍長",
    "總司令": "総司令",
    "司令長官": "司令長官",
    "參謀總長": "参謀総長",
    "參謀長": "参謀長",
    "軍事委員會": "軍事委員会",
    "行政院院長": "行政院長",
    "總統": "総統",
    "黃埔軍校": "黄埔軍校",
    "陸軍": "陸軍",
    "步兵": "歩兵",
    "騎兵": "騎兵",
    "炮兵": "砲兵",
    "砲兵": "砲兵",
    "裝甲": "装甲",
    "師長": "師長",
    "旅長": "旅長",
    "團長": "団長",
    "營長": "営長",
    "少將": "少将",
    "中將": "中将",
    "上將": "上将",
    "上校": "大佐",
    "中校": "中佐",
    "少校": "少佐",
    "參與": "参加した",
    "參加": "参加した",
    "率部": "部隊を率いた",
    "任": "に任じられた",
    "歷任": "歴任した",
    "曾任": "かつて務めた",
    "出任": "就任した",
    "兼任": "兼任した",
    "改編": "改編した",
    "改編為": "に改編された",
    "駐守": "守備した",
    "駐紮": "駐屯した",
    "指揮": "指揮した",
    "作戰": "作戦",
    "戰役": "戦役",
    "戰鬥": "戦闘",
    "會戰": "会戦",
    "攻占": "攻略した",
    "攻佔": "攻略した",
    "被俘": "捕虜となった",
    "投降": "降伏",
    "起義": "蜂起",
    "倒戈": "寝返った",
    "敗退": "敗退した",
    "戰敗": "敗北した",
    "病逝": "病没した",
    "身故": "死去した",
    "逝世": "逝去した",
    "離開": "離れた",
    "成立": "成立した",
    "建立": "築いた",
    "推動": "推進した",
    "領導": "率いた",
    "宣布": "宣言した",
    "提出": "提案した",
    "返回": "帰還した",
    "赴": "へ赴いた",
    "返國": "帰国した",
    "其後": "その後",
    "同年": "同年",
    "翌年": "翌年",
    "期間": "期間中",
    "長期": "長期にわたり",
    "逐步": "次第に",
    "一度": "一時",
    "持續": "継続した",
    "人稱": "と呼ばれた",
    "著有": "を著した",
    "出生": "出生した",
    "生於": "で生まれた",
    "。": "。",
    "，": "、",
    "、": "、",
    "；": "；",
    "：": "：",
    "（": "（",
    "）": "）",
    "「": "「",
    "」": "」",
}


def read_json(path: Path):
    return json.loads(path.read_text(encoding="utf-8"))


def sha256_file(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as stream:
        for chunk in iter(lambda: stream.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def sha256_text(value: str) -> str:
    return hashlib.sha256(value.encode("utf-8")).hexdigest()


def sources(root: Path) -> list[dict[str, str]]:
    result = []
    for facts in sorted((root / "docs/reference/people").glob("facts-*.json")):
        stem = facts.stem.removeprefix("facts-")
        bios = facts.with_name(f"bios-{stem}.md")
        if not bios.is_file():
            continue
        result.append({
            "facts": facts.relative_to(root).as_posix(),
            "facts_sha256": sha256_file(facts),
            "bios": bios.relative_to(root).as_posix(),
            "bios_sha256": sha256_file(bios),
        })
    if not result:
        raise SystemExit("找不到成對的 facts-*.json／bios-*.md")
    return result


def merged_people(root: Path) -> list[dict]:
    base = read_json(root / "translations/zh-Hant/people.json")["people"]
    authored = read_json(root / "translations/zh-Hant/people-authored.json")["people"]
    by_id = {row["id"]: dict(row) for row in base}
    for row in authored:
        if row["id"] in by_id and by_id[row["id"]].get("bio_zh"):
            raise SystemExit(f"自傳來源重複 id={row['id']}")
        by_id.setdefault(row["id"], {"id": row["id"], "name_ingame": row["name_ingame"]})["bio_zh"] = row["bio_zh"]
    people = [row for row in by_id.values() if row.get("bio_zh", "").strip() and row["id"] != 274]
    people.sort(key=lambda row: row["id"])
    if len(people) != 387:
        raise SystemExit(f"machine-draft 來源應有 387 篇，實際 {len(people)}")
    return people


def replace_phrases(text: str, phrases: dict[str, str]) -> str:
    for source, target in sorted(phrases.items(), key=lambda item: len(item[0]), reverse=True):
        text = text.replace(source, target)
    return text


def normalize_dates(text: str, language: str) -> str:
    def day(match: re.Match[str]) -> str:
        y, m, d = match.groups()
        if language == "en":
            return f"{y}-{int(m):02d}-{int(d):02d}"
        return f"{y}年{int(m)}月{int(d)}日"

    text = re.sub(r"(\d{4})年(\d{1,2})月(\d{1,2})日", day, text)
    if language == "en":
        text = re.sub(r"(\d{4})年(\d{1,2})月", lambda m: f"{m.group(1)}-{int(m.group(2)):02d}", text)
        text = re.sub(r"(\d{4})年", lambda m: f"in {m.group(1)}", text)
    return text


def lexical_translation(text: str, language: str) -> str:
    phrases = EN_PHRASES if language == "en" else JA_PHRASES
    text = normalize_dates(text, language)
    text = replace_phrases(text, phrases)
    text = re.sub(r"\s+", " ", text).strip()
    if language == "en":
        # 片語替換後常見「in 1926National Revolutionary Army」邊界；
        # 以字元類別補空格只改善可讀性，不捏造未命中的歷史詞義。
        text = re.sub(r"(?<=[A-Za-z])(?=[\u3400-\u9fff])", " ", text)
        text = re.sub(r"(?<=[\u3400-\u9fff])(?=[A-Za-z])", " ", text)
        text = re.sub(r"(?<=[A-Za-z])(?=\d)", " ", text)
        text = re.sub(r"(?<=\d)(?=[A-Za-z])", " ", text)
        text = re.sub(r"\s+([,.;:])", r"\1", text)
        text = text.replace("..", ".")
    return text


def fact_summary(person: dict, language: str) -> str:
    name = person.get("name_common") or person.get("name_ingame", "")
    birth = person.get("birth")
    death = person.get("death")
    birthplace = person.get("birthplace") or ("unknown" if language == "en" else "不詳")
    faction = person.get("faction") or ("unknown affiliation" if language == "en" else "所属不詳")
    post = person.get("highest_post") or ("no confirmed highest post" if language == "en" else "最高職不詳")
    periods = person.get("periods") or []
    if language == "en":
        life = []
        if birth:
            life.append(f"born in {birthplace} in {birth}")
        else:
            life.append(f"born in {birthplace}")
        if death:
            life.append(f"died in {death}")
        else:
            life.append("death year unknown")
        active = ", ".join(periods) if periods else "period unknown"
        return (
            f"{name} is recorded as a historical military and political figure: "
            f"{'; '.join(life)}; affiliation {faction}; highest recorded post {post}; "
            f"active period {active}."
        )
    life = []
    life.append(f"{birthplace}出身" + (f"（{birth}年生" if birth else "（生年不詳"))
    life[-1] += f"、{death}年没" if death else "、没年不詳"
    active = "・".join(periods) if periods else "時期不詳"
    return (
        f"{name}は歴史上の軍事・政治人物として記録される。"
        f"{life[0]}）。所属は{faction}、最高職は{post}、活動時期は{active}。"
    )


def draft_biography(person: dict, language: str) -> str:
    name = person.get("name_common") or person.get("name_ingame", "")
    source = person["bio_zh"].strip()
    translated = lexical_translation(source, language)
    if language == "en":
        return (
            f"Wiki-based biography draft. {fact_summary(person, language)} "
            f"Source narrative (lexical draft): {translated} "
            "[Proper names and unmatched historical terms remain in the source script; unreviewed wiki-derived draft.]"
        )
    return (
        f"ウィキ出典草稿。{fact_summary(person, language)} "
        f"原文語彙置換稿：{translated} "
        "［固有名詞と未対応の歴史用語は原文表記を残す。未確認草稿。］"
    )


def build_overlay(root: Path, language: str) -> dict:
    people = merged_people(root)
    rows = []
    for person in people:
        bio = draft_biography(person, language)
        rows.append({
            "id": person["id"],
            "name_ingame": person["name_ingame"],
            "bio": bio,
            "bio_language": language,
            "bio_status": TRANSLATION_STATUS,
            "source_bio_sha256": sha256_text(person["bio_zh"]),
            "translated_bio_sha256": sha256_text(bio),
            "review": {"status": "unreviewed", "reviewer": "", "reviewed_at": ""},
        })
    return {
        "schema_version": "1",
        "language": language,
        "overlay_scope": "translated-biographies",
        "translation_status": TRANSLATION_STATUS,
        "generated_from": sources(root),
        "people": rows,
    }


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--root", type=Path, default=Path(__file__).resolve().parents[1])
    parser.add_argument("--language", choices=("en", "ja"), action="append")
    args = parser.parse_args()
    root = args.root.resolve()
    for language in args.language or ["en", "ja"]:
        path = root / f"translations/{language}/people-biography-overlay.json"
        path.write_text(json.dumps(build_overlay(root, language), ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
        print(f"寫入 {path.relative_to(root)}：387 篇 {TRANSLATION_STATUS}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
