package game

// 戰鬥中的「有沒有後援」判斷（`sub_534FF` → `sub_53619`，`docs/re/31` §47）。
//
// 這一組控制**兩處必勝結算**與值 16／17 的分流，是決策鏈裡影響最大的
// 單一判斷之一：
//
//	戰力差五倍，**而且有我方鄰省可以支援**——才敢直接判勝負。
//	沒有後援就不冒險速戰，改走打城市或推倒重來。
//
// ⚠️ **`sub_534FF` 本身不在這裡實作。** 它是 `ProvinceTable.ReinforcementSources`
// （`province.go`），從 `docs/re/07` §9 解出來的，比這裡早。
// 本檔只做「戰鬥決策鏈怎麼問它」這一層。
//
// 這件事踩過一次：初版在這裡把 `sub_534FF` 重寫了一遍，而且抄漏兩道
// ——無主省原版是**停止掃描**不是跳過，`+32` bit 6 的語意也已經解出來是
// 「那個省正在打仗」。`CLAUDE.md` §7.4 講的就是這個：
// **一條規則只留一份實作，重抄的那份一定會漂。**

// SupportingNeighbours 是決策鏈問的那一句：**當前交戰省的鄰省裡，
// 有哪些屬於 `leader` 這個勢力而且能出兵**。
//
//	tbl     省份表
//	at      當前交戰的省（原版 `byte_6FFC4`）
//	leader  單位效忠的勢力領袖（`+14`）
//	units   將領表，數「可用將領數」那道篩選要用
//
// 直接轉給 `ReinforcementSources`——原版就是同一支 `sub_534FF`，
// 政略的增援與戰鬥的後援判斷共用它。
func SupportingNeighbours(tbl *ProvinceTable, at ProvinceID, leader GeneralID,
	units []CombatUnit) []ProvinceID {
	if tbl == nil || leader == 0 {
		return nil
	}
	out, err := tbl.ReinforcementSources(at, leader, units)
	if err != nil {
		return nil
	}
	return out
}

// HasBattleSupport 回答「這一方在當前交戰省有沒有可用的鄰省支援」。
//
// 這是 `sub_53619` 的**正相**——原版回的是反相（有支援回 0），
// 所以接到 `BattleAIInput.Sub53619` 時要取反。
//
// 刻意提供正相版本：`Sub53619` 那個名字保留原版的反相語意是為了對照
// 組合語言，但呼叫端讀「有沒有支援」比讀「sub_53619 為不為零」清楚得多。
func HasBattleSupport(tbl *ProvinceTable, at ProvinceID, leader GeneralID,
	units []CombatUnit) bool {
	return len(SupportingNeighbours(tbl, at, leader, units)) > 0
}
