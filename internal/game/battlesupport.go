package game

// 戰鬥中的「有沒有後援」判斷（`sub_534FF` → `sub_53619`，`docs/re/31` §47）。
//
// 這一組控制**兩處必勝結算**與值 16／17 的分流，是決策鏈裡影響最大的
// 單一判斷之一：
//
//	戰力差五倍，**而且有我方鄰省可以支援**——才敢直接判勝負。
//	沒有後援就不冒險速戰，改走打城市或推倒重來。

// ProvinceFlagNoSupport 是省份記錄 `+32` 的 **bit 6**。
//
// `sub_534FF` 掃鄰省時，這一位立起來的省會被跳過（§47）。
// ⚠️ **語意未解**——`docs/re/14` 只解出 bit 2（本回合已處理過）。
// 這裡的命名是照它在本函式裡的作用取的，不是原版的欄位名。
const ProvinceFlagNoSupport uint8 = 0x40

// SupportingNeighbours 是 `sub_534FF`：找出**當前交戰省的鄰省裡，
// 屬於 `leader` 這個勢力而且可用的省**。
//
//	tbl     省份表
//	at      當前交戰的省（原版 `byte_6FFC4`）
//	leader  單位效忠的勢力領袖（`+14`）
//
// 原版的五道篩選，這裡實作四道：
//
//	鄰省編號有效           ✅
//	司令 != 0（非無主省）  ✅
//	司令 == leader         ✅
//	`+32` bit 6 沒立起來    ✅
//	`sub_5A881(鄰) < 100`   ❌ **未實作**——那支未讀（§47）
//
// ⚠️ 少一道篩選的後果是**回報的支援省可能偏多**，也就是必勝結算比原版
// 容易觸發。補上之前不要拿它對原版做行為驗收。
func SupportingNeighbours(tbl *ProvinceTable, at ProvinceID, leader GeneralID) []ProvinceID {
	if tbl == nil || leader == 0 {
		return nil
	}
	p, err := tbl.At(at)
	if err != nil {
		return nil
	}
	var out []ProvinceID
	for _, n := range p.Neighbours {
		np, err := tbl.At(n)
		if err != nil {
			continue
		}
		if np.Commander == 0 || np.Commander != leader {
			continue
		}
		if np.Flags&ProvinceFlagNoSupport != 0 {
			continue
		}
		out = append(out, n)
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
func HasBattleSupport(tbl *ProvinceTable, at ProvinceID, leader GeneralID) bool {
	return len(SupportingNeighbours(tbl, at, leader)) > 0
}
