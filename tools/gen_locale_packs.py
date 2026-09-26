#!/usr/bin/env python3
"""產生可重現的英／日語系執行期資料。

這個工具只讀繁中母本，輸出 translations/en 與 translations/ja；不讀取原版
執行檔或研究來源。人物 base 仍保留繁中來源正文；若同目錄存在
people-biography-overlay.json，runtime 會另外以 machine-draft／human-reviewed
狀態套用翻譯正文。
"""

from __future__ import annotations

import argparse
import json
import re
from pathlib import Path


EN: dict[str, tuple[str, str]] = {
    "common.confirm": ("Commander, are you sure?", "Proceed?"),
    "common.cancel": ("Cancel", "Cancel"),
    "settings.title": ("Display settings", "Display settings"),
    "settings.wording": ("Wording", "Text style"),
    "settings.wording.original": ("Original register", "Original register"),
    "settings.wording.plain": ("Plain language", "Modern plain language"),
    "settings.theme": ("Visual theme", "Screen style"),
    "settings.theme.retro": ("Original graphics", "Classic look"),
    "settings.theme.modern": ("Modern graphics", "Clear modern look"),
    "other.save": ("Save game", "Save progress"),
    "other.load": ("Load game", "Load progress"),
    "other.save.confirm": ("Are you sure?", "Save the current progress?"),
    "other.load.confirm": ("Are you sure?", "Load progress? Unsaved changes will be lost."),
    "other.sound": ("Sound effects", "Sound effects"),
    "other.command_art": ("Command illustrations", "Command illustrations"),
    "other.music": ("Music", "Background music"),
    "other.message_time": ("Message time", "Message display time"),
    "other.message_time.prompt": ("Message time (1-10)", "How long should messages stay? (1-10)"),
    "other.spectate": ("Spectate", "Watch the computer battle"),
    "other.quit": ("Quit game", "Save and quit"),
    "other.display": ("Display settings", "Display settings"),
    "other.unavailable": ("Not available", "Not implemented"),
    "policy.title": ("Policy", "Policy"),
    "policy.autonomy": ("Authorize autonomy", "Set provincial autonomy"),
    "policy.production": ("Production allocation", "Adjust resource production"),
    "policy.production.unavailable": ("Allocate iron, coal, oil and food", "Move gold production to iron, coal, oil or food"),
    "autonomy.title": ("Authorize autonomy", "Set provincial autonomy"),
    "autonomy.normal": ("Normal", "Central control"),
    "autonomy.enabled": ("Autonomy", "Autonomy active"),
    "autonomy.prompt": ("Enter province number; ESC finishes", "Enter a province number; ESC finishes"),
    "production.title": ("Provincial production", "Resource production allocation"),
    "production.gold": ("Gold", "Gold remaining"),
    "production.iron": ("Iron", "Iron"),
    "production.coal": ("Coal", "Coal"),
    "production.oil": ("Oil", "Oil"),
    "production.food": ("Food", "Food"),
    "production.select": ("Commander, which production?", "Which production should change?"),
    "production.value": ("Commander, set it to what?", "What percentage should it be?"),
    "recruit.action": ("Recruit", "Recruit soldiers"),
    "recruit.reorganize": ("Reorganize", "Redistribute existing forces"),
    "recruit.infantry": ("Infantry", "Infantry"),
    "recruit.armour": ("Armour", "Armour"),
    "recruit.artillery": ("Artillery", "Artillery"),
    "recruit.cavalry": ("Cavalry", "Cavalry"),
    "recruit.amount": ("Commander, how many?", "How many to recruit?"),
    "recruit.limit": ("Recruitment limit", "Maximum recruitable"),
    "recruit.cost": ("Cost", "This action costs"),
    "recruit.gold": ("Gold", "Gold"),
    "recruit.confirm": ("Are you sure?", "Recruit these forces? (Y/N)"),
    "recruit.remaining": ("Remaining forces", "Unassigned forces"),
    "recruit.general": ("Commander", "General"),
    "recruit.force": ("Force", "Force"),
    "command.01": ("Movement", "Move forces"),
    "command.02": ("Military action", "Conduct a military action"),
    "command.03": ("Resupply", "Deliver supplies"),
    "command.04": ("Taxation", "Collect taxes"),
    "command.05": ("Recruitment", "Recruit soldiers"),
    "command.06": ("Inspect", "View information"),
    "command.07": ("Development", "Build and develop"),
    "command.08": ("Policy", "Set provincial policy"),
    "command.09": ("Diplomacy", "Handle diplomacy"),
    "command.10": ("Ceasefire talks", "Negotiate a ceasefire"),
    "command.11": ("Covert action", "Conduct a covert action"),
    "command.12": ("Commerce", "Buy and sell supplies"),
    "command.13": ("Training", "Train forces"),
    "command.14": ("Support troops and civilians", "Care for troops and civilians"),
    "command.15": ("Other options", "Other settings"),
    "diplomacy.loan": ("Foreign loan", "Apply for a foreign loan"),
    "diplomacy.aid": ("Request aid", "Request international aid"),
    "diplomacy.repay": ("Repay debt", "Repay foreign debt"),
    "diplomacy.loan.prompt": ("Commander, loan amount? (0-5000)", "How much gold to borrow? (0-5000)"),
    "diplomacy.loan.credit": ("Commander, current credit", "Current credit"),
    "diplomacy.loan.unavailable": ("Loan unavailable", "Credit is zero; no loan can be requested"),
    "diplomacy.loan.refused": ("All countries refused the loan", "The loan was refused (one command spent)"),
    "diplomacy.aid.result": ("Aid approved: gold +%d, food +%d, ammo +%d, fuel +%d (commands left %d)", "Aid approved: gold +%d, food +%d, ammo +%d, fuel +%d (commands left %d)"),
    "diplomacy.aid.refused": ("All countries refused aid", "International aid was refused (one command spent)"),
    "diplomacy.repay.prompt": ("Commander, repayment amount?", "How much foreign debt to repay?"),
    "diplomacy.repay.debt": ("Current debt %d, gold %d", "Current debt %d, gold %d"),
    "diplomacy.repay.none": ("No foreign debt", "There is no foreign debt to repay"),
    "diplomacy.repay.invalid": ("Repayment must be positive", "Enter an amount greater than zero"),
    "diplomacy.repay.result": ("Repaid debt %d (balance %d; credit %d; commands left %d)", "Repaid foreign debt %d (remaining %d; credit %d; commands left %d)"),
    "ceasefire.prompt": ("Commander, where to negotiate a ceasefire?", "Which province should negotiate a ceasefire?"),
    "ceasefire.range": ("Province number 1-%d", "Province number range 1-%d"),
    "ceasefire.unavailable": ("Unavailable", "Ceasefire talks are unavailable this period"),
    "ceasefire.no_commander": ("Commander is not in this province", "The commander is not here"),
    "ceasefire.invalid": ("Invalid ceasefire province", "Enter a valid ceasefire target"),
    "ceasefire.no_battle": ("No battle", "The target province is not at war"),
    "ceasefire.agreed": ("Ceasefire agreed in province %d (commands left %d)", "Ceasefire agreed in province %d (commands left %d)"),
    "ceasefire.refused": ("Ceasefire refused in province %d (commands left %d)", "The opponent refused a ceasefire in province %d (commands left %d)"),
    "develop.reclaim": ("Reclaim land", "Open new farmland"),
    "develop.arsenal": ("Build arsenal", "Build an arsenal"),
    "develop.mine": ("Mine gold", "Open a gold mine"),
    "train.confirm": ("Commander, train the forces?", "Train this province's forces?"),
    "covert.guerrilla": ("Guerrillas", "Send guerrillas"),
    "covert.student": ("Student agitation", "Agitate student protests"),
    "covert.student.target": ("Commander, which province to agitate?", "Where should students protest?"),
    "view.other": ("Inspect other provinces", "View another province"),
    "view.owned": ("Inspect owned provinces", "View provinces under control"),
    "view.generals": ("Inspect generals", "View generals here"),
    "view.province_names": ("Inspect province names", "View province numbers"),
    "view.select_prompt": ("Inspect another province", "Enter a province number to view"),
    "view.choice.overview": ("Overview", "View province data"),
    "view.choice.generals": ("Inspect generals", "View generals in this province"),
    "view.owned.title": ("Owned provinces", "Owned province overview"),
    "view.names.title": ("Province names", "Province number reference"),
    "trade.import": ("Import", "Buy supplies"),
    "trade.export": ("Export", "Sell supplies"),
    "trade.food": ("Food", "Food"),
    "trade.ammo": ("Ammunition", "Ammunition"),
    "trade.fuel": ("Fuel", "Fuel"),
    "trade.coal": ("Coal", "Coal"),
    "trade.iron": ("Iron", "Iron"),
    "trade.buy_amount": ("Commander, purchase how much?", "How much to buy?"),
    "trade.sell_amount": ("Commander, sell how much?", "How much to sell?"),
    "supply.target": ("Commander, resupply which province?", "Which province should receive supplies?"),
    "supply.gold": ("Commander, how much gold?", "How much gold to deliver?"),
    "supply.food": ("Commander, how much food?", "How much food to deliver?"),
    "supply.ammo": ("Commander, how much ammunition?", "How much ammunition to deliver?"),
    "supply.fuel": ("Commander, how much fuel?", "How much fuel to deliver?"),
    "transfer.mode": ("Commander, how to move forces?", "How should forces move?"),
    "transfer.mode.partial": ("Move some", "Choose generals"),
    "transfer.mode.all": ("Move all", "Move all available generals"),
    "transfer.target": ("Commander, move to which province?", "Which province should receive them?"),
    "transfer.select.partial": ("Commander, which general?", "Which general should move?"),
    "transfer.select.all": ("Commander, who stays?", "Which generals should stay?"),
    "transfer.selection.confirm": ("Confirm selection", "Confirm this roster"),
    "transfer.resource.gold": ("Commander, how much gold?", "How much gold to carry?"),
    "transfer.resource.food": ("Commander, how much food?", "How much food to carry?"),
    "transfer.resource.ammo": ("Commander, how much ammunition?", "How much ammunition to carry?"),
    "transfer.resource.fuel": ("Commander, how much fuel?", "How much fuel to carry?"),
    "biography.unavailable": ("No reliable biography", "No confirmed life record is available"),
    "biography.page": ("Biography", "Life record"),
    "biography.source_fallback": ("Biography text: Traditional Chinese source", "Showing the Traditional Chinese source; a localized biography is pending"),
}


# 日文包刻意採「漢字優先」的短句，避免把未提供的日文字型偽裝成已完成。
# 人名與史料正文仍保留原版漢字；這些詞條可由現有倚天字庫呈現，真正的
# 假名／比例字型與完整人物譯稿另列為後續工作。
JA: dict[str, tuple[str, str]] = {
    "common.confirm": ("司令、実行確認？", "実行確認？"),
    "common.cancel": ("取消", "取消"),
    "settings.title": ("表示設定", "表示設定"),
    "settings.wording": ("表示用語", "表示用語"),
    "settings.wording.original": ("原典用語", "原典用語"),
    "settings.wording.plain": ("現代用語", "現代用語"),
    "settings.theme": ("圖形主題", "画面様式"),
    "settings.theme.retro": ("原版圖形", "古典様式"),
    "settings.theme.modern": ("現代圖形", "現代様式"),
    "other.save": ("遊戯保存", "進行保存"),
    "other.load": ("遊戯読込", "進行読込"),
    "other.save.confirm": ("実行確認？", "現在進行を保存？"),
    "other.load.confirm": ("実行確認？", "進行読込？未保存変更消失"),
    "other.sound": ("効果音", "効果音"),
    "other.command_art": ("指令圖", "指令圖"),
    "other.music": ("音楽", "背景音楽"),
    "other.message_time": ("消息時間", "消息表示時間"),
    "other.message_time.prompt": ("消息時間（1-10）", "消息表示時間（1-10）"),
    "other.spectate": ("観戦", "電脳戦観戦"),
    "other.quit": ("遊戯終了", "保存終了"),
    "other.display": ("表示設定", "表示設定"),
    "other.unavailable": ("未完成", "未実装"),
    "policy.title": ("政策", "政策"),
    "policy.autonomy": ("自治許可", "省自治設定"),
    "policy.production": ("生産配分", "資源生産調整"),
    "policy.production.unavailable": ("鐵、煤、油、糧配分", "黃金生産を鐵、煤、油、糧へ配分"),
    "autonomy.title": ("自治許可", "省自治設定"),
    "autonomy.normal": ("正常", "中央管理"),
    "autonomy.enabled": ("自治", "自治中"),
    "autonomy.prompt": ("省番号入力；ESC終了", "省番号入力；ESC終了"),
    "production.title": ("本省生産", "資源生産配分"),
    "production.gold": ("黃金", "黃金残量"),
    "production.iron": ("鐵礦", "鐵礦"),
    "production.coal": ("煤礦", "煤礦"),
    "production.oil": ("石油", "石油"),
    "production.food": ("糧食", "糧食"),
    "production.select": ("司令、何項調整？", "何項生産を調整？"),
    "production.value": ("司令、何％設定？", "何％に設定？"),
    "recruit.action": ("徴兵", "兵員募集"),
    "recruit.reorganize": ("再編成", "現有兵力再配分"),
    "recruit.infantry": ("步兵", "步兵"),
    "recruit.armour": ("装甲兵", "装甲兵"),
    "recruit.artillery": ("砲兵", "砲兵"),
    "recruit.cavalry": ("騎兵", "騎兵"),
    "recruit.amount": ("司令、何人徴兵？", "何人募集？"),
    "recruit.limit": ("徴兵上限", "募集可能最大"),
    "recruit.cost": ("必要量", "今回必要量"),
    "recruit.gold": ("黃金", "黃金"),
    "recruit.confirm": ("実行確認？", "兵員募集？（Y/N）"),
    "recruit.remaining": ("残存兵力", "未配分兵力"),
    "recruit.general": ("司令", "将領"),
    "recruit.force": ("兵力", "兵力"),
    "command.01": ("移動行動", "部隊移動"),
    "command.02": ("軍事行動", "軍事行動実行"),
    "command.03": ("補給輸送", "補給輸送"),
    "command.04": ("徴税", "税収"),
    "command.05": ("徴兵", "兵員募集"),
    "command.06": ("閲覧", "情報閲覧"),
    "command.07": ("開発", "建設開発"),
    "command.08": ("政策", "省政策設定"),
    "command.09": ("外交", "外交処理"),
    "command.10": ("停戦交渉", "停戦交渉"),
    "command.11": ("秘密行動", "秘密行動実行"),
    "command.12": ("商業活動", "物資売買"),
    "command.13": ("練兵", "部隊訓練"),
    "command.14": ("軍民慰労", "軍民慰労"),
    "command.15": ("其他選項", "其他設定"),
    "diplomacy.loan": ("外国借款", "外国借款申請"),
    "diplomacy.aid": ("外援請求", "国際援助請求"),
    "diplomacy.repay": ("外債償還", "対外借款償還"),
    "diplomacy.loan.prompt": ("司令、借款額？（0-5000）", "黃金借入量？（0-5000）"),
    "diplomacy.loan.credit": ("司令、現在信用度", "現在信用度"),
    "diplomacy.loan.unavailable": ("借款不可", "信用度零、借款不可"),
    "diplomacy.loan.refused": ("各国借款拒否", "借款拒否（一指令消費）"),
    "diplomacy.aid.result": ("援助許可：黃金 +%d、糧食 +%d、彈藥 +%d、燃料 +%d（残存指令 %d）", "援助許可：黃金 +%d、糧食 +%d、彈藥 +%d、燃料 +%d（残存指令 %d）"),
    "diplomacy.aid.refused": ("各国援助拒否", "国際援助不許可（一指令消費）"),
    "diplomacy.repay.prompt": ("司令、償還額？", "対外借款の償還額？"),
    "diplomacy.repay.debt": ("現在外債 %d、黃金 %d", "現在外債 %d、黃金 %d"),
    "diplomacy.repay.none": ("外債無", "償還対象の外債無"),
    "diplomacy.repay.invalid": ("償還額は正数", "零超の額を入力"),
    "diplomacy.repay.result": ("外債償還 %d（残額 %d；信用度 %d；残存指令 %d）", "対外借款償還 %d（残額 %d；信用度 %d；残存指令 %d）"),
    "ceasefire.prompt": ("司令、何省で停戦交渉？", "停戦対象省？"),
    "ceasefire.range": ("省番号 1-%d", "省番号範囲 1-%d"),
    "ceasefire.unavailable": ("使用不可", "本期停戦交渉不可"),
    "ceasefire.no_commander": ("司令は本省外", "司令は本省に不在"),
    "ceasefire.invalid": ("停戦省番号無効", "有効な停戦省を入力"),
    "ceasefire.no_battle": ("戦事無", "対象省に戦事無"),
    "ceasefire.agreed": ("省 %d 停戦同意（残存指令 %d）", "省 %d 停戦同意（残存指令 %d）"),
    "ceasefire.refused": ("省 %d 停戦拒否（残存指令 %d）", "相手は省 %d 停戦拒否（残存指令 %d）"),
    "develop.reclaim": ("開墾", "土地開墾"),
    "develop.arsenal": ("兵工廠建設", "兵工廠建設"),
    "develop.mine": ("金山掘削", "金山開採"),
    "train.confirm": ("司令、練兵実行？", "本省部隊を訓練？"),
    "covert.guerrilla": ("游擊隊", "游擊隊派遣"),
    "covert.student": ("学生運動鼓動", "学生抗議鼓動"),
    "covert.student.target": ("司令、何省で学生運動？", "学生抗議の省？"),
    "view.other": ("他省閲覧", "他省閲覧"),
    "view.owned": ("所属各省閲覧", "所属省閲覧"),
    "view.generals": ("将領閲覧", "本省将領閲覧"),
    "view.province_names": ("省名閲覧", "省番号閲覧"),
    "view.select_prompt": ("他省閲覧", "閲覧省番号入力"),
    "view.choice.overview": ("概況", "省資料閲覧"),
    "view.choice.generals": ("将領閲覧", "該当省将領閲覧"),
    "view.owned.title": ("所属各省概況", "所属省概況"),
    "view.names.title": ("省名閲覧", "省番号対照表"),
    "trade.import": ("輸入", "物資購入"),
    "trade.export": ("輸出", "物資売却"),
    "trade.food": ("糧食", "糧食"),
    "trade.ammo": ("彈藥", "彈藥"),
    "trade.fuel": ("燃料", "燃料"),
    "trade.coal": ("煤礦", "煤礦"),
    "trade.iron": ("鐵礦", "鐵礦"),
    "trade.buy_amount": ("司令、購入量？", "購入量？"),
    "trade.sell_amount": ("司令、売却量？", "売却量？"),
    "supply.target": ("司令、何省へ補給？", "補給対象省？"),
    "supply.gold": ("司令、黃金量？", "輸送黃金量？"),
    "supply.food": ("司令、糧食量？", "輸送糧食量？"),
    "supply.ammo": ("司令、彈藥量？", "輸送彈藥量？"),
    "supply.fuel": ("司令、燃料量？", "輸送燃料量？"),
    "transfer.mode": ("司令、如何移動？", "移動方式？"),
    "transfer.mode.partial": ("部分移動", "将領選択"),
    "transfer.mode.all": ("全部移動", "使用可能将領を全部移動"),
    "transfer.target": ("司令、何省へ移動？", "移動先省？"),
    "transfer.select.partial": ("司令、何将領？", "移動将領？"),
    "transfer.select.all": ("司令、誰を残留？", "残留将領？"),
    "transfer.selection.confirm": ("選択確認", "名簿確認"),
    "transfer.resource.gold": ("司令、黃金量？", "同送黃金量？"),
    "transfer.resource.food": ("司令、糧食量？", "同送糧食量？"),
    "transfer.resource.ammo": ("司令、彈藥量？", "同送彈藥量？"),
    "transfer.resource.fuel": ("司令、燃料量？", "同送燃料量？"),
    "biography.unavailable": ("信頼伝記無", "確認可能な生涯資料無"),
    "biography.page": ("人物自伝", "人物生涯"),
    "biography.source_fallback": ("伝記本文：繁中原文", "繁中原文を表示；本語系訳稿は未完"),
}


PROVINCES = {
    "en": [
        "Nen River", "Jilin", "Hejiang", "Heilongjiang", "Liaobei", "Xing'an",
        "Liaoning", "Songjiang", "Andong", "Jehol", "Hebei", "Chahar", "Mongolia",
        "Suiyuan", "Ningxia", "Shanxi", "Gansu", "Shaanxi", "Henan", "Shandong",
        "Jiangsu", "Anhui", "Zhejiang", "Fujian", "Jiangxi", "Hubei", "Hunan",
        "Guizhou", "Sichuan", "Qinghai", "Xinjiang", "Tibet", "Xikang", "Yunnan",
        "Guangxi", "Guangdong", "Taiwan", "Hainan Island", "Burma",
    ],
    "ja": [
        "嫩江省", "吉林省", "合江省", "黑龍江", "遼北省", "興安省", "遼寧省",
        "松江省", "安東省", "熱河省", "河北省", "察哈爾", "蒙古", "綏遠省",
        "寧夏省", "山西省", "甘肅省", "陜西省", "河南省", "山東省", "江蘇省",
        "安徽省", "浙江省", "福建省", "江西省", "湖北省", "湖南省", "貴州省",
        "四川省", "青海省", "新疆省", "西藏", "西康省", "雲南省", "廣西省",
        "廣東省", "臺灣省", "海南島", "緬甸",
    ],
}

# 目前工作樹提供的是倚天繁中字庫，日文包先以可呈現的繁體漢字短句為邊界。
# 這些替換不宣稱完整日文翻譯，只避免把倚天沒有的日式新字與假名送進執行期。
JA_GLYPH_COMPAT = str.maketrans({
    "で": "於", "な": "何", "に": "於", "の": "之", "は": "是", "へ": "向", "を": "之",
    "処": "處", "労": "勞", "効": "效", "却": "卻", "収": "收", "号": "號", "囲": "圍",
    "国": "國", "売": "賣", "変": "變", "学": "學", "実": "實", "対": "對", "将": "將",
    "属": "屬", "当": "當", "徴": "徵", "戦": "戰", "戯": "戲", "択": "擇", "数": "數",
    "楽": "樂", "様": "樣", "残": "殘", "渉": "涉", "産": "產", "画": "畫", "発": "發",
    "税": "稅", "脳": "腦", "装": "裝", "覧": "覽", "観": "觀", "訳": "譯", "読": "讀",
    "込": "入", "閲": "閱", "頼": "賴",
})


def read_json(path: Path):
    return json.loads(path.read_text(encoding="utf-8"))


def write_json(path: Path, data) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(data, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def format_signature(text: str) -> tuple[str, ...]:
    return tuple(re.findall(r"%[a-zA-Z]", text))


def build_pack(root: Path, language: str, wording: dict[str, tuple[str, str]]) -> None:
    zh_dir = root / "translations" / "zh-Hant"
    out_dir = root / "translations" / language
    source = read_json(zh_dir / "wording.json")
    source_keys = set(source["entries"])
    if source_keys != set(wording):
        missing = sorted(source_keys - set(wording))
        extra = sorted(set(wording) - source_keys)
        raise SystemExit(f"{language}: wording key mismatch missing={missing} extra={extra}")
    entries = {}
    for key, src in source["entries"].items():
        original, plain = wording[key]
        if language == "ja":
            original = original.translate(JA_GLYPH_COMPAT)
            plain = plain.translate(JA_GLYPH_COMPAT)
        for label, value, reference in (("original", original, src["original"]), ("plain", plain, src["plain"])):
            if not value.strip():
                raise SystemExit(f"{language}: {key}/{label} is empty")
            if format_signature(value) != format_signature(reference):
                raise SystemExit(f"{language}: {key}/{label} format placeholders differ")
        entries[key] = {"original": original, "plain": plain}
    write_json(out_dir / "wording.json", {"entries": entries})

    glyph = read_json(zh_dir / "glyphtext.json")
    glyph["language"] = language
    province_entries = glyph["files"]["3.15"]["entries"]
    for i, name in enumerate(PROVINCES[language]):
        province_entries[i]["text"] = name
        province_entries[i]["raw"] = name
    write_json(out_dir / "glyphtext.json", glyph)

    people = read_json(zh_dir / "people.json")
    # SPEC-11 的繁中 authored overlay 是 additive 來源：英／日包必須先
    # 接合這 61 篇，才能對 387 篇成文正文做 source-fallback 或後續翻譯
    # overlay；不能只複製 326 篇 base 而讓語系包悄悄遺漏研究稿。
    authored = read_json(zh_dir / "people-authored.json")
    authored_by_id = {row["id"]: row for row in authored["people"]}
    for person in people["people"]:
        row = authored_by_id.get(person["id"])
        if row is None:
            continue
        if person.get("bio_zh"):
            raise SystemExit(f"{language}: authored overlay 試圖覆蓋既有正文 id={person['id']}")
        person["bio_zh"] = row["bio_zh"]
        person["confidence"] = row["confidence"]
    if sum(bool(person.get("bio_zh")) for person in people["people"]) != 387:
        raise SystemExit(f"{language}: authored overlay 接合後正文數不是 387")
    people["language"] = language
    people["source"] = (
        "translations/zh-Hant/people.json + people-authored.json"
        "（base 保留繁中來源；英／日 machine-draft overlay 另存）"
    )
    for person in people["people"]:
        person["bio_language"] = "zh-Hant"
        person["bio_status"] = "source-fallback"
    write_json(out_dir / "people.json", people)

    overlay_note = "已建立 machine-draft overlay，runtime 會以 fail-closed 方式套用預覽正文。" if (out_dir / "people-biography-overlay.json").is_file() else "尚未建立人物翻譯 overlay，runtime 會顯示繁中 source-fallback。"
    readme = (
        f"# {language} 語系包\n\n"
        "玩家可見的命令、設定、訊息與 39 省名稱已接到此目錄。人物姓名保留遊戲的歷史寫法；"
        f"387 篇可接合自傳的 base 保留繁中來源；{overlay_note}"
        "英／日逐篇譯稿必須另以 `people-biography-overlay.json` 載入，並通過"
        " `docs/spec/28-locale-biography-overlay.md` 的來源、審稿、排版與字型 gates；"
        "machine-draft 只供預覽，未經人審不得改標成 translated。日文包採漢字優先，"
        "假名字型與完整日文譯稿另列後續工作。\n"
    )
    (out_dir / "README.md").write_text(readme, encoding="utf-8")


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--root", type=Path, default=Path(__file__).resolve().parents[1])
    parser.add_argument("--language", choices=("en", "ja"), action="append")
    args = parser.parse_args()
    languages = args.language or ["en", "ja"]
    for language in languages:
        build_pack(args.root, language, EN if language == "en" else JA)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
