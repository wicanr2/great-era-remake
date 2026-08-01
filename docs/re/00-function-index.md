# 已解函式索引（自動產生，勿手改）

> `./tools/py.sh tools/gen_func_index.py` 重新產生。

**動手讀任何 `sub_` 之前先在這裡 grep 它。**
`CLAUDE.md` §7.1「動手挖之前先查手上已有的」——這份索引就是為了讓那一條真的做得到。

目前 **393** 個函式在文件裡有紀錄。

| 函式 | 目前的結論（摘自文件）| 出處 | 提及次數 |
|---|---|---|---:|
| `sub_10002` | 15 \| 其他選項 \| 存讀檔、授權自治、提升武裝等 \| `sub_10002` \| confirmed \| | `docs/mechanics/10-political.md`（另 1 份） | 2 |
| `sub_10193` | `sub_10193`（玩家的指令迴圈）每執行完一個指令就： | `docs/mechanics/10-political.md`（另 7 份） | 16 |
| `sub_10541` | if 覆滅: 回 1；sub_10541(第一期 ? 20h : 40h) | `docs/re/16-game-over.md` | 3 |
| `sub_1060A` | **`byte_7003A`**（`sub_1060A` 的觀戰路徑，只判天下統一）， | `docs/playtest/13-new-game-settings.md`（另 5 份） | 8 |
| `sub_10AB5` | `sub_10AB5` 那組（MARK-A → MARK-C）還沒細讀，**先不下定論**。 | `docs/re/01-war-exe-turbo-pascal.md`（另 4 份） | 9 |
| `sub_11380` | **20h（32）** \| `sub_11380`、`sub_52222`、`sub_524FC`、`sub_5CEAE` \| 政略 + 戰鬥 \| | `docs/re/21-develop-and-general-state.md` | 1 |
| `sub_115BD` | `sub_115BD` \| `mov [di+7A7Eh], cl` \| | `docs/re/20-field-writers.md`（另 1 份） | 2 |
| `sub_11C04` | `Random(2)` \| 6 \| `sub_11C04`(3) 等 \| | `docs/mechanics/70-ai.md` | 1 |
| `sub_126D5` | `sub_126D5` \| `mov 0EA60h`（60,000）\| 設到上限 \| | `docs/re/20-field-writers.md` | 1 |
| `sub_12AAD` | **1** \| `sub_12AAD`、`sub_13484`、`sub_5375E`、`sub_53B1E` \| 政略 + 戰鬥 \| | `docs/re/21-develop-and-general-state.md` | 1 |
| `sub_12FB2` | `sub_12FB2` \| 設 0 \| | `docs/re/21-develop-and-general-state.md` | 1 |
| `sub_13484` | **1** \| `sub_12AAD`、`sub_13484`、`sub_5375E`、`sub_53B1E` \| 政略 + 戰鬥 \| | `docs/re/21-develop-and-general-state.md` | 1 |
| `sub_13710` | if sub_13710(當前省) 且 父[-0Ch] + 5 ≥ 清單長度 且 byte_6FE88 != 1: 停 | `docs/re/30-sortie-preparation.md`（另 1 份） | 4 |
| `sub_1374F` | sub_1374F(狀態) 將領 +17（兵力）:= ax | `docs/re/32-seasonal-interior.md` | 1 |
| `sub_13786` | if 司令 == 0 或 命令數 == 0: sub_13786(); return | `docs/re/13-turn-structure-and-roster.md`（另 1 份） | 3 |
| `sub_137AB` | if byte_6FE83 != 0: sub_137AB(...) | `docs/mechanics/70-ai.md`（另 1 份） | 2 |
| `sub_1398D` | 4. `sub_1398D`：勢力領袖遷都 — confirmed | `docs/re/13-turn-structure-and-roster.md`（另 1 份） | 4 |
| `sub_13B44` | ⚠️ 與 `docs/re/13` §1「清單依攻擊力降序」**不衝突**：那份是 `sub_13B44` | `docs/playtest/07-field-naming-by-screen.md`（另 3 份） | 4 |
| `sub_13D23` | confirmed：`sub_13D23` 呼叫 `sub_306CF` 之後 `mov ss:[di-236h], ax` / | `docs/mechanics/70-ai.md`（另 6 份） | 11 |
| `sub_13DDC` | if !sub_5C7FE(司令): sub_13DDC() | `docs/re/13-turn-structure-and-roster.md` | 1 |
| `sub_13EA8` | if sub_13EA8(狀態, 省) < 3: return ; 某個計數 < 3 | `docs/re/32-seasonal-interior.md` | 2 |
| `sub_13F4E` | `sub_13F4E(狀態, 省)` = 掃鄰省找符合條件的 — **confirmed（結構）** | `docs/mechanics/70-ai.md`（另 2 份） | 11 |
| `sub_1401E` | if sub_1401E(將領, 鄰省) 且 模式 == 1: → 選中它 | `docs/mechanics/70-ai.md` | 3 |
| `sub_1408F` | `sub_1408F` 把同一格當 **Pascal 陣列的長度欄**用：`arr[0]` 是元素數， | `docs/mechanics/70-ai.md`（另 4 份） | 11 |
| `sub_14761` | 6g. `sub_14761`：這個省有沒有滿員／缺員的將領 — **confirmed** | `docs/mechanics/70-ai.md`（另 2 份） | 6 |
| `sub_14860` | 3. `sub_14860(狀態, 省)`：算兩個威脅指標 — confirmed（結構） | `docs/re/11-rescue-and-pathfinding.md`（另 1 份） | 3 |
| `sub_149C8` | 2. `sub_149C8`：候選是「我方的、前線的、快撐不住的」省 — confirmed | `docs/re/11-rescue-and-pathfinding.md`（另 1 份） | 2 |
| `sub_14B0D` | ⚠️ 決策鏈 A 的兩個呼叫都用 `arg_4 = 3`，會**跳過 `sub_14B0D` 與 `sub_14BEE`**； | `docs/re/28-ai-chain-a.md`（另 1 份） | 2 |
| `sub_14BEE` | ⚠️ 決策鏈 A 的兩個呼叫都用 `arg_4 = 3`，會**跳過 `sub_14B0D` 與 `sub_14BEE`**； | `docs/re/28-ai-chain-a.md`（另 1 份） | 2 |
| `sub_14CCF` | 5. `sub_14CCF`：挑的不是目的地，是**通往它的第一步** — confirmed | `docs/re/11-rescue-and-pathfinding.md`（另 2 份） | 3 |
| `sub_14E85` | 目標 = sub_14E85(狀態, 0, 1, 當前省) | `docs/re/11-rescue-and-pathfinding.md`（另 1 份） | 6 |
| `sub_14EBA` | 3 \| `sub_14EBA` \| **黃金最多**的省（`docs/re/12` §5）\| — \| | `docs/mechanics/70-ai.md`（另 2 份） | 7 |
| `sub_14F9A` | `sub_14F9A` 在調動成功後把**目標省的司令欄位改成我方勢力**。 | `docs/mechanics/20-military.md`（另 8 份） | 37 |
| `sub_15018` | 2. **`+16 = 1`** 與 `sub_15018` 檢查的 `+16 & 1` 對上： | `docs/mechanics/30-combat.md`（另 7 份） | 19 |
| `sub_150FC` | 回 0 \| **沒有可攻打的鄰省** \| `sub_15018` → `sub_15667` → `sub_150FC(1)` → `sub_1527A` → `sub_15925(0)` → *`sub_15A9… | `docs/mechanics/70-ai.md`（另 5 份） | 31 |
| `sub_1527A` | 回 0 \| **沒有可攻打的鄰省** \| `sub_15018` → `sub_15667` → `sub_150FC(1)` → `sub_1527A` → `sub_15925(0)` → *`sub_15A9… | `docs/mechanics/70-ai.md`（另 4 份） | 18 |
| `sub_1541E` | 回非 0 \| **有可攻打的鄰省** \| `sub_1541E` → `sub_150FC(2)` → `sub_15925(1)` → *`sub_15A9A`* → *`sub_15CE3`* \| | `docs/mechanics/70-ai.md`（另 4 份） | 20 |
| `sub_15667` | 回 0 \| **沒有可攻打的鄰省** \| `sub_15018` → `sub_15667` → `sub_150FC(1)` → `sub_1527A` → `sub_15925(0)` → *`sub_15A9… | `docs/mechanics/70-ai.md`（另 4 份） | 18 |
| `sub_1588C` | `sub_1588C`：放行無主省與「自己人太少」的省 — **confirmed**（2026-08-01 訂正） | `docs/mechanics/70-ai.md`（另 3 份） | 8 |
| `sub_15925` | 回 0 \| **沒有可攻打的鄰省** \| `sub_15018` → `sub_15667` → `sub_150FC(1)` → `sub_1527A` → `sub_15925(0)` → *`sub_15A9… | `docs/mechanics/70-ai.md`（另 8 份） | 32 |
| `sub_15A9A` | 回 0 \| **沒有可攻打的鄰省** \| `sub_15018` → `sub_15667` → `sub_150FC(1)` → `sub_1527A` → `sub_15925(0)` → *`sub_15A9… | `docs/mechanics/70-ai.md`（另 3 份） | 16 |
| `sub_15CE3` | 回非 0 \| **有可攻打的鄰省** \| `sub_1541E` → `sub_150FC(2)` → `sub_15925(1)` → *`sub_15A9A`* → *`sub_15CE3`* \| | `docs/mechanics/70-ai.md`（另 4 份） | 19 |
| `sub_15F3C` | ⚠️ **清單的順序還沒解**（在 `sub_15F3C` 裡填的）。 | `docs/mechanics/20-military.md`（另 6 份） | 13 |
| `sub_16047` | 3r. ✅ 步驟 5：柿子挑軟的捏 `sub_16047` — **confirmed** | `docs/mechanics/70-ai.md`（另 2 份） | 9 |
| `sub_161EB` | 3. `sub_161EB`：湊兵前的初始化 — **confirmed（欄位）** | `docs/re/30-sortie-preparation.md` | 1 |
| `sub_1629E` | +21 \| **兵種編號**（強證據；值域含 1/4/5/6）。`== 1` 可穿越長城 \| `sub_4A583`、`sub_1629E` \| | `docs/mechanics/30-combat.md`（另 2 份） | 8 |
| `sub_163BE` | **`sub_163BE` 回非 0 就停**。它裡面依模式分兩套停止條件， | `docs/re/30-sortie-preparation.md` | 4 |
| `sub_1655C` | 上面那套是高難度（`sub_1655C`）。低難度走 `sub_16A9E`，標準完全不同： | `docs/mechanics/70-ai.md`（另 2 份） | 14 |
| `sub_16646` | `sub_16A9E` 的四支子函式未讀：`sub_16646`／`sub_16702`／`sub_16962`／`sub_16905`（§5c）。 | `docs/re/30-sortie-preparation.md`（另 1 份） | 5 |
| `sub_16702` | `sub_16702` 未讀，所以低難度**挑人的篩選**（有沒有兵種 4 那條）也未定。 | `docs/re/30-sortie-preparation.md`（另 1 份） | 7 |
| `sub_16905` | 大概是 `sub_16905`（`sub_16A9E` 的收尾，未讀）做同樣的事，**未驗**。 | `docs/re/30-sortie-preparation.md`（另 1 份） | 7 |
| `sub_16962` | `sub_16962` 判斷、`sub_16905` 收尾）。但控制流本身讀得出來： | `docs/re/30-sortie-preparation.md`（另 1 份） | 7 |
| `sub_16A9E` | 上面那套是高難度（`sub_1655C`）。低難度走 `sub_16A9E`，標準完全不同： | `docs/mechanics/70-ai.md`（另 1 份） | 10 |
| `sub_16B44` | 3x. ✅ 補給門檻 `sub_16B44` — **confirmed**（六步裡出現四次） | `docs/mechanics/70-ai.md`（另 2 份） | 13 |
| `sub_16D98` | `sub_5AE48` \| `sub_39B6E`（**戰鬥主函式**）、`sub_4F883`、`sub_16D98` \| 戰鬥端 \| | `docs/re/29-ai-does-prepare-attacks.md` | 1 |
| `sub_16E77` | `sub_16E77`（148 行）。**這是第一條解出來的「電腦為什麼選這個指令」**—— | `docs/mechanics/70-ai.md`（另 1 份） | 6 |
| `sub_16F89` | 3w. ✅ 優勢比門檻 `sub_16F89` — **confirmed** | `docs/mechanics/70-ai.md`（另 2 份） | 10 |
| `sub_17019` | 3u. ✅ 軟柿子判定 `sub_17019` / `sub_306CF` — **confirmed** | `docs/mechanics/70-ai.md`（另 2 份） | 12 |
| `sub_170C0` | 3v. ✅ 「被包圍就突圍」`sub_5A9F6` / `sub_170C0` — **confirmed** | `docs/mechanics/70-ai.md`（另 1 份） | 6 |
| `sub_17135` | 3t. ✅ 步驟 4 的目標挑選是**三層優先序** `sub_17135` — **confirmed** | `docs/mechanics/70-ai.md`（另 1 份） | 7 |
| `sub_17236` | 步驟 6：sub_5A9F6(當前省) 為真 → sub_17236 → sub_174C9（出兵） | `docs/mechanics/70-ai.md`（另 1 份） | 8 |
| `sub_173B5` | `父[-28h]` 那 10 格是出兵名單的副本 — **confirmed**（`sub_173B5`） | `docs/re/30-sortie-preparation.md` | 4 |
| `sub_17437` | 5b. `sub_17437`：湊到兩倍了嗎 — **confirmed（公式）** | `docs/re/30-sortie-preparation.md`（另 1 份） | 4 |
| `sub_174C9` | 步驟 6：sub_5A9F6(當前省) 為真 → sub_17236 → sub_174C9（出兵） | `docs/mechanics/70-ai.md`（另 7 份） | 40 |
| `sub_17ADA` | 步驟 3（`sub_17ADA+314`）\| 掃到的每一個**敵對鄰省** \| | `docs/mechanics/70-ai.md`（另 6 份） | 21 |
| `sub_18509` | （開頭就跳走了），要 `sub_18509` 那個呼叫端才走得到——**仍是假說**。 | `docs/re/30-sortie-preparation.md`（另 1 份） | 5 |
| `sub_18903` | `sub_18903` \| **`add byte ptr [di+7A7Eh], 14h`（+20）** \| | `docs/re/20-field-writers.md`（另 1 份） | 5 |
| `sub_189C8` | 6 \| 地價 `+17` 增加（`sub_189C8`）\| **地價 < 200 且不是「兵多沒錢」** \| | `docs/re/32-seasonal-interior.md` | 2 |
| `sub_18A67` | 7. ⭐ `sub_18A67` = 「這個將領滿員了嗎」——四個滿員數的程式碼證據 | `docs/re/32-seasonal-interior.md` | 6 |
| `sub_18AED` | **`sub_18F10` 已讀**（30 行，轉手呼叫 `sub_18AED`，見 §8）； | `docs/re/32-seasonal-interior.md` | 7 |
| `sub_18CD5` | sub_18CD5(bp) ; 設 狀態[-231h] = 1 | `docs/re/32-seasonal-interior.md` | 3 |
| `sub_18D45` | 花多少：`sub_18D45` — **confirmed** | `docs/re/32-seasonal-interior.md` | 2 |
| `sub_18F10` | **`sub_18F10` 已讀**（30 行，轉手呼叫 `sub_18AED`，見 §8）； | `docs/re/32-seasonal-interior.md` | 5 |
| `sub_18F5B` | `sub_18F5B`／`sub_191A6` 大量把 `-232h` 拿去做乘法——**與「命令數」讀不通**， | `docs/re/13-turn-structure-and-roster.md`（另 1 份） | 8 |
| `sub_191A6` | `sub_18F5B`／`sub_191A6` 大量把 `-232h` 拿去做乘法——**與「命令數」讀不通**， | `docs/re/13-turn-structure-and-roster.md`（另 1 份） | 8 |
| `sub_19443` | `sub_19443`：把花費分配到部隊 — **confirmed（結構）** | `docs/re/32-seasonal-interior.md` | 3 |
| `sub_194F4` | 原文：全檔只有 `sub_194F4` 一處寫，而且是 `sub [di-6229h], cx` | `docs/re/21-develop-and-general-state.md`（另 2 份） | 4 |
| `sub_19563` | 3 \| `sub_19563` \| **黃金 ≥ 5,000** \| | `docs/re/32-seasonal-interior.md` | 11 |
| `sub_197A9` | if 命令數 > 0 且 狀態[-6] == 0 且 !sub_197A9(): | `docs/re/13-turn-structure-and-roster.md` | 2 |
| `sub_197DB` | ⚠️ 但**黃金支出接了卻不知道換到什麼**（`sub_19CD4` 與 `sub_197DB` 只看到 | `docs/re/32-seasonal-interior.md`（另 1 份） | 4 |
| `sub_19918` | 11 \| `sub_19918` \| 將領 `+16` bit 2 **未設**才進這一支 \| | `docs/re/32-seasonal-interior.md`（另 1 份） | 6 |
| `sub_19B89` | sub_19B89() | `docs/re/13-turn-structure-and-roster.md`（另 1 份） | 2 |
| `sub_19CD4` | 1 \| 黃金 −5,000（`sub_19CD4`）\| **兵工廠數（`+18`）== 0** 才做 \| | `docs/re/32-seasonal-interior.md`（另 1 份） | 5 |
| `sub_19D6A` | `sub_19D6A`／`sub_3964E` 各 6 次最多。逐一對出來才知道 | `docs/re/17-random.md`（另 1 份） | 2 |
| `sub_1A100` | `sub_1A100` \| `+= 20`，夾到 100 \| **電腦內政「休養」** \| | `docs/re/19-loyalty.md`（另 2 份） | 7 |
| `sub_1A729` | `Random(3)` \| 2 \| `sub_1A729`、`sub_1AA28` \| | `docs/mechanics/70-ai.md` | 1 |
| `sub_1AA28` | `Random(5)` \| 6 \| `sub_36D74`(3)、`sub_1AA28`、`sub_1FC1C`、`sub_35B19` \| | `docs/mechanics/70-ai.md`（另 1 份） | 3 |
| `sub_1AC01` | sub_1AC01() | `docs/re/13-turn-structure-and-roster.md`（另 1 份） | 2 |
| `sub_1ACCC` | `sub_1ACCC`（**電腦的一個回合**）\| **138 支** \| **零命中** \| | `docs/re/12-ai-chain-has-no-attack.md`（另 9 份） | 18 |
| `sub_1B089` | `sub_1B089` \| 1 \| 觀戰 \| | `docs/re/27-screen-map.md` | 1 |
| `sub_1B399` | `sub_1B399` \| 15 \| 司令欲儲存進度於那個檔案？司令？檔案儲存成功 \| | `docs/re/27-screen-map.md` | 2 |
| `sub_1B74B` | 反追 `sub_1B74B`（處理 `save(%d).dt1` 的函式）發現它用的是 | `docs/playtest/02-load-save-strategy-phase.md`（另 1 份） | 3 |
| `sub_1BC5F` | `sub_1BC5F` \| 3 \| 司令？ \| | `docs/re/27-screen-map.md` | 1 |
| `sub_1BDC6` | `sub_1BDC6` \| 3 \| 司令？ \| | `docs/re/27-screen-map.md` | 1 |
| `sub_1BF58` | `byte_7003B` 全 `WAR.EXE` **只有一個寫入點**：`sub_1BF58`，而那是個 toggle： | `docs/mechanics/70-ai.md`（另 1 份） | 3 |
| `sub_1C0BF` | `sub_1C0BF` \| 3 \| 司令？ \| | `docs/re/27-screen-map.md` | 1 |
| `sub_1C220` | `sub_1C220` \| 3 \| 司令？ \| | `docs/re/27-screen-map.md` | 1 |
| `sub_1C38C` | **輸出 ON/OFF 字串的設定開關**。`sub_1BF58` 由 `sub_1C38C` 呼叫， | `docs/mechanics/70-ai.md`（另 2 份） | 3 |
| `sub_1C606` | **「休息」有對應的函式**：`sub_1C606` 印「司令〈某〉欲休息嗎？」 | `docs/mechanics/10-political.md`（另 9 份） | 12 |
| `sub_1C916` | 13 \| 練兵 \| 提升**士兵戰技** \| `sub_1C916` \| confirmed \| | `docs/mechanics/10-political.md`（另 8 份） | 10 |
| `sub_1CB96` | **進口（買）** \| `sub_1D703` \| 糧食、彈藥、燃料 \| `sub_1CB96`／`sub_1CF62`／`sub_1D32E` \| | `docs/mechanics/40-economy.md`（另 1 份） | 3 |
| `sub_1CF62` | **進口（買）** \| `sub_1D703` \| 糧食、彈藥、燃料 \| `sub_1CB96`／`sub_1CF62`／`sub_1D32E` \| | `docs/mechanics/40-economy.md`（另 1 份） | 3 |
| `sub_1D32E` | **進口（買）** \| `sub_1D703` \| 糧食、彈藥、燃料 \| `sub_1CB96`／`sub_1CF62`／`sub_1D32E` \| | `docs/mechanics/40-economy.md`（另 1 份） | 3 |
| `sub_1D703` | 12 \| 商業活動 \| **買賣資源**：進口 3 種、出口 5 種，**匯率全解**（彈藥 10／煤鐵 20／糧食燃料 30 換 1 金，買賣同價）\| `sub_1D703`、`sub_1ED95` \| conf… | `docs/mechanics/10-political.md`（另 2 份） | 6 |
| `sub_1DAA8` | **出口（賣）** \| `sub_1ED95` \| 糧食、彈藥、煤礦、鐵礦、燃料 \| `sub_1DAA8`／`sub_1DE6E`／`sub_1E234`／`sub_1E5FA`／`sub_1E9C0` \| | `docs/mechanics/40-economy.md`（另 1 份） | 3 |
| `sub_1DE6E` | **出口（賣）** \| `sub_1ED95` \| 糧食、彈藥、煤礦、鐵礦、燃料 \| `sub_1DAA8`／`sub_1DE6E`／`sub_1E234`／`sub_1E5FA`／`sub_1E9C0` \| | `docs/mechanics/40-economy.md`（另 1 份） | 3 |
| `sub_1E234` | **出口（賣）** \| `sub_1ED95` \| 糧食、彈藥、煤礦、鐵礦、燃料 \| `sub_1DAA8`／`sub_1DE6E`／`sub_1E234`／`sub_1E5FA`／`sub_1E9C0` \| | `docs/mechanics/40-economy.md`（另 1 份） | 3 |
| `sub_1E5FA` | **出口（賣）** \| `sub_1ED95` \| 糧食、彈藥、煤礦、鐵礦、燃料 \| `sub_1DAA8`／`sub_1DE6E`／`sub_1E234`／`sub_1E5FA`／`sub_1E9C0` \| | `docs/mechanics/40-economy.md`（另 1 份） | 3 |
| `sub_1E9C0` | **出口（賣）** \| `sub_1ED95` \| 糧食、彈藥、煤礦、鐵礦、燃料 \| `sub_1DAA8`／`sub_1DE6E`／`sub_1E234`／`sub_1E5FA`／`sub_1E9C0` \| | `docs/mechanics/40-economy.md`（另 1 份） | 3 |
| `sub_1ED95` | 12 \| 商業活動 \| **買賣資源**：進口 3 種、出口 5 種，**匯率全解**（彈藥 10／煤鐵 20／糧食燃料 30 換 1 金，買賣同價）\| `sub_1D703`、`sub_1ED95` \| conf… | `docs/mechanics/10-political.md`（另 2 份） | 4 |
| `sub_1F380` | 提升武裝 \| `sub_1F380` \| **武裝程度**，要花錢 \| | `docs/mechanics/60-personnel.md`（另 1 份） | 3 |
| `sub_1FA1C` | 12 商業活動 \| `sub_1FA1C` \| 未碰 \| | `docs/playtest/10-recruit-economics.md`（另 3 份） | 4 |
| `sub_1FC1C` | 11 \| 秘密行動 \| **機率全解**：游擊隊 40%／學潮 20%，錢在判定前就扣 \| `sub_1FC1C`、`sub_20447` \| confirmed \| | `docs/mechanics/10-political.md`（另 4 份） | 7 |
| `sub_20447` | 11 \| 秘密行動 \| **機率全解**：游擊隊 40%／學潮 20%，錢在判定前就扣 \| `sub_1FC1C`、`sub_20447` \| confirmed \| | `docs/mechanics/10-political.md`（另 4 份） | 7 |
| `sub_20B28` | `sub_20447` \| `× 0.8` \| 秘密行動底下（`sub_20B28` = 指令 11）\| | `docs/re/19-loyalty.md`（另 3 份） | 6 |
| `sub_20CF0` | 10 \| 談判停火 \| **判定全解**：比雙方戰力，佔上風 70%／劣勢 20% \| `sub_211D5`、`sub_20CF0` \| confirmed \| | `docs/mechanics/10-political.md`（另 1 份） | 2 |
| `sub_20E05` | 談判停火 \| `sub_211D5`、`sub_20E05` \| 「司令不在本省／欲在何省談判停火？／並無戰事／同意／拒絕」\| | `docs/mechanics/50-diplomacy.md`（另 1 份） | 3 |
| `sub_211D5` | 10 \| 談判停火 \| **判定全解**：比雙方戰力，佔上風 70%／劣勢 20% \| `sub_211D5`、`sub_20CF0` \| confirmed \| | `docs/mechanics/10-political.md`（另 3 份） | 7 |
| `sub_2164A` | 9 \| 外交 \| **三支公式全解**：貸款門檻、外援 70%＋史實禁運、還債對稱 \| `sub_2164A` 等 \| confirmed \| | `docs/mechanics/10-political.md`（另 2 份） | 7 |
| `sub_21D1D` | 請求外援 \| `sub_21D1D` \| 「各國均拒絕提供援助／同意援助共援助／黃金 糧食 彈藥 燃料」\| | `docs/mechanics/50-diplomacy.md`（另 2 份） | 6 |
| `sub_223ED` | `sub_223ED` 有「付錢 + 加信用度」，但**扣減外債餘額那一步還沒讀到**。 | `docs/mechanics/50-diplomacy.md`（另 1 份） | 5 |
| `sub_2296C` | 償還外債 \| `sub_223ED`／`sub_2296C` \| 「目前**外債**／信用度為」\| | `docs/re/27-screen-map.md`（另 2 份） | 4 |
| `sub_22E25` | 整個月」，那是錯的**：`sub_22E25` 是授權自治不是開發。全檔只有它設 bit 5， | `docs/mechanics/10-political.md`（另 4 份） | 9 |
| `sub_23511` | 8 \| 政策 \| **產能分配百分比**（`docs/re/26`）\| `sub_23511` \| confirmed \| | `docs/mechanics/10-political.md`（另 6 份） | 21 |
| `sub_23FF6` | `sub_23FF6` 是**指令 8「政策」**。開發是 `sub_24D59`。 | `docs/playtest/06-dosbox-restored.md`（另 5 份） | 7 |
| `sub_241D0` | 全檔只有 `sub_22E25` 設 bit 5；開發的三支（`sub_24D59`／`sub_241D0`／ | `docs/mechanics/40-economy.md`（另 2 份） | 4 |
| `sub_24535` | → **部分有解**：挖金礦 `sub_24535` 讀的正是一張**獨立於省份記錄之外**的 | `docs/mechanics/40-economy.md`（另 3 份） | 6 |
| `sub_248A8` | `sub_248A8`）用 `tools/field_writes.py` 掃過，**一個都不碰它**。 | `docs/mechanics/40-economy.md`（另 3 份） | 4 |
| `sub_24D59` | `sub_23FF6` 是**指令 8「政策」**。開發是 `sub_24D59`。 | `docs/playtest/06-dosbox-restored.md`（另 8 份） | 9 |
| `sub_25527` | 步兵 \| `sub_260C7` \| `sub_25B2B` \| `sub_25527` \| | `docs/mechanics/20-military.md`（另 1 份） | 2 |
| `sub_25B2B` | 步兵 \| `sub_260C7` \| `sub_25B2B` \| `sub_25527` \| | `docs/mechanics/20-military.md`（另 1 份） | 3 |
| `sub_260C7` | 步兵 \| `sub_260C7` \| `sub_25B2B` \| `sub_25527` \| | `docs/mechanics/20-military.md`（另 1 份） | 3 |
| `sub_26931` | 砲兵 \| `sub_271F8` \| `sub_26D69` \| `sub_26931` \| | `docs/mechanics/20-military.md`（另 1 份） | 2 |
| `sub_26D69` | 砲兵 \| `sub_271F8` \| `sub_26D69` \| `sub_26931` \| | `docs/mechanics/20-military.md`（另 1 份） | 3 |
| `sub_271F8` | 砲兵 \| `sub_271F8` \| `sub_26D69` \| `sub_26931` \| | `docs/mechanics/20-military.md`（另 1 份） | 3 |
| `sub_27992` | 裝甲兵 \| `sub_28259` \| `sub_27DCA` \| `sub_27992` \| | `docs/mechanics/20-military.md`（另 1 份） | 2 |
| `sub_27DCA` | 裝甲兵 \| `sub_28259` \| `sub_27DCA` \| `sub_27992` \| | `docs/mechanics/20-military.md`（另 1 份） | 3 |
| `sub_28259` | `sub_28259`（徵兵執行，630 行）末尾： | `docs/playtest/10-recruit-economics.md`（另 5 份） | 8 |
| `sub_289F3` | 騎兵 \| `sub_29494` \| `sub_28FF7` \| `sub_289F3` \| | `docs/mechanics/20-military.md`（另 1 份） | 2 |
| `sub_28FF7` | 騎兵 \| `sub_29494` \| `sub_28FF7` \| `sub_289F3` \| | `docs/mechanics/20-military.md`（另 1 份） | 3 |
| `sub_29494` | 騎兵 \| `sub_29494` \| `sub_28FF7` \| `sub_289F3` \| | `docs/mechanics/20-military.md`（另 1 份） | 3 |
| `sub_29B46` | 5 \| 徵兵 \| 四兵種各一支，10 兵 = 1 金 \| `sub_29B46` \| confirmed \| | `docs/mechanics/10-political.md`（另 3 份） | 5 |
| `sub_29DDF` | `sub_29B46`（徵兵選單）與 `sub_29DDF`（整編選單）並列四項： | `docs/re/27-screen-map.md`（另 2 份） | 4 |
| `sub_2A11B` | 5 \| **徵兵** \| `sub_2A11B` \| \| 13 \| 練兵 \| `sub_1C916` \| | `docs/re/13-turn-structure-and-roster.md`（另 1 份） | 2 |
| `sub_2A941` | `sub_2A941`：編號／省名／省長／狀態／黃金／糧食／彈藥／燃料／煤礦／鐵礦／ | `docs/re/27-screen-map.md` | 2 |
| `sub_2AB96` | `sub_2AB96` \| 2 \| 萬萬 \| | `docs/re/27-screen-map.md` | 1 |
| `sub_2B344` | **「查閱將領」那一頁實際印出來的清單**（`sub_2B344`，`docs/re/27`）。 | `docs/mechanics/60-personnel.md`（另 1 份） | 5 |
| `sub_2B567` | `sub_2B567` \| 5 \| －－－步兵砲兵裝甲兵騎兵 \| | `docs/re/27-screen-map.md` | 1 |
| `sub_2BBA2` | 6 \| 查閱 \| 換省檢視／查閱將領 \| `sub_2BBA2`、`sub_2C351` \| confirmed \| | `docs/mechanics/10-political.md`（另 1 份） | 2 |
| `sub_2C27C` | `sub_2C27C` \| 2 \| 按換頁 \| | `docs/re/27-screen-map.md` | 1 |
| `sub_2C351` | 6 \| 查閱 \| 換省檢視／查閱將領 \| `sub_2BBA2`、`sub_2C351` \| confirmed \| | `docs/mechanics/10-political.md`（另 2 份） | 3 |
| `sub_2C6C6` | 1. **讀 `sub_2C6C6`（440 行）裡的 Real 運算與 `Random` 呼叫點** | `docs/playtest/12-tax-command.md`（另 7 份） | 12 |
| `sub_2CB6B` | 3 \| 運補 \| `sub_2CB6B` \| \| 11 \| 秘密行動 \| `sub_20B28` \| | `docs/re/13-turn-structure-and-roster.md`（另 1 份） | 2 |
| `sub_2D238` | 2 \| 軍事行動 \| 攻打，見 `20-military.md` \| `sub_2D238` \| confirmed \| | `docs/mechanics/10-political.md`（另 1 份） | 2 |
| `sub_2D812` | sub_39B6E **零命中** 直接呼叫者：['sub_2D812', 'sub_3562B', 'sub_368A8'] | `docs/re/12-ai-chain-has-no-attack.md`（另 4 份） | 11 |
| `sub_2DEFC` | sub_2DEFC **零命中** 直接呼叫者：['sub_10193'] | `docs/re/12-ai-chain-has-no-attack.md`（另 4 份） | 11 |
| `sub_2E425` | 1 \| 調動行動 \| 整個將領連人帶兵換省，八個模式 \| `sub_2E425`、`sub_2E9A1` \| confirmed \| | `docs/mechanics/10-political.md`（另 1 份） | 2 |
| `sub_2E676` | `sub_2E676` \| 2 \| 無法調動 \| | `docs/re/27-screen-map.md` | 1 |
| `sub_2E9A1` | 1 \| 調動行動 \| 整個將領連人帶兵換省，八個模式 \| `sub_2E425`、`sub_2E9A1` \| confirmed \| | `docs/mechanics/10-political.md`（另 1 份） | 2 |
| `sub_2F2F4` | 1 \| 調動行動 \| `sub_2F2F4` \| \| 9 \| 外交 \| `sub_2296C` \| | `docs/re/13-turn-structure-and-roster.md`（另 1 份） | 2 |
| `sub_2F74C` | +22–24 \| 各 3 \| 未解，**同一支函式**（`sub_2F74C`）存取的三元組 \| | `docs/spec/02-generals.md`（另 1 份） | 2 |
| `sub_2F8ED` | +25–27 \| 各 3 \| 未解，**同一支函式**（`sub_2F8ED`）存取的三元組 \| | `docs/spec/02-generals.md`（另 1 份） | 2 |
| `sub_2FCA9` | 會呼叫 `sub_2FCA9`，看起來是**每季一次**） | `docs/re/16-game-over.md`（另 1 份） | 2 |
| `sub_30003` | 年度結算 `sub_30003`（`docs/re/25` §2）逐省做： | `docs/mechanics/40-economy.md`（另 8 份） | 13 |
| `sub_303CD` | **[2026-08-01 解決]** `sub_303CD` 讀出來了（`docs/re/18`）： | `docs/playtest/12-tax-command.md`（另 3 份） | 6 |
| `sub_30526` | **交戰中的省、無主的省都跳過。** `sub_30526` 的內容未讀，列為未解。 | `docs/re/25-year-end.md` | 4 |
| `sub_306CF` | `sub_5A881`（數量）與 `sub_306CF`（兵力總和）的篩選條件一字不差： | `docs/mechanics/20-military.md`（另 7 份） | 18 |
| `sub_30794` | 3d. `sub_30794` 不是第二筆收入 | `docs/re/18-tax-formula.md`（另 1 份） | 3 |
| `sub_3083A` | 3 \| 運補 \| **搬運全解**：四種資源各一支函式，**超過上限退回來源省** \| `sub_3083A`–`sub_30AD4` \| confirmed \| | `docs/mechanics/10-political.md`（另 1 份） | 4 |
| `sub_30918` | `sub_30918` \| 糧食 \| `+2` \| | `docs/mechanics/40-economy.md` | 1 |
| `sub_309F6` | `sub_309F6` \| 彈藥 \| `+4` \| | `docs/mechanics/40-economy.md` | 1 |
| `sub_30AD4` | 3 \| 運補 \| **搬運全解**：四種資源各一支函式，**超過上限退回來源省** \| `sub_3083A`–`sub_30AD4` \| confirmed \| | `docs/mechanics/10-political.md`（另 1 份） | 2 |
| `sub_31AA6` | `sub_31AA6` \| −10 \| | `docs/re/21-develop-and-general-state.md`（另 1 份） | 2 |
| `sub_3231A` | `sub_3231A` 依序問四次：「司令欲調動多少**黃金**／**糧食**／**彈藥**／**燃料**？」 | `docs/mechanics/40-economy.md`（另 1 份） | 3 |
| `sub_33888` | `sub_33888` \| 13 \| 司令無省長無數兵工廠數兵力將領數人民忠誠度萬 \| | `docs/re/27-screen-map.md`（另 1 份） | 2 |
| `sub_34105` | sub_34105(目標省) ; 重畫 | `docs/mechanics/70-ai.md`（另 1 份） | 3 |
| `sub_3412B` | 指令裡只剩偏移」的相對形式。**慰勞軍民（`sub_3412B`）也會加民心**， | `docs/re/19-loyalty.md`（另 6 份） | 12 |
| `sub_349C1` | `sub_349C1` 回的是 **1 / 人力權重**（`docs/re/10` §3），所以 | `docs/mechanics/70-ai.md`（另 4 份） | 15 |
| `sub_34B0B` | `sub_34B0B` \| 2 \| 陣亡陣亡 \| | `docs/re/27-screen-map.md` | 1 |
| `sub_34D5D` | `sub_34D5D` \| `mov byte ptr [di+7A7Eh], 64h`（設 100）\| | `docs/re/20-field-writers.md` | 1 |
| `sub_34F5A` | `sub_34F5A` \| 4 \| 繼任司令繼任司令 \| | `docs/re/27-screen-map.md` | 2 |
| `sub_35005` | `sub_35005` \| 2 \| 無人繼任 \| | `docs/re/27-screen-map.md` | 1 |
| `sub_353C4` | `sub_353C4` \| 4 \| 選擇一位將領繼任 \| | `docs/re/27-screen-map.md` | 2 |
| `sub_3562B` | sub_39B6E **零命中** 直接呼叫者：['sub_2D812', 'sub_3562B', 'sub_368A8'] | `docs/re/12-ai-chain-has-no-attack.md`（另 2 份） | 4 |
| `sub_35B19` | `sub_35B19` 的「中國／日軍」對立），所以能談的是那四個。 | `docs/mechanics/50-diplomacy.md`（另 2 份） | 3 |
| `sub_35E89` | `Random(6)` \| 2 \| `sub_24535`、`sub_35E89` \| | `docs/mechanics/70-ai.md` | 1 |
| `sub_35F51` | `sub_35F51` \| 4 \| 離開投靠改編為 \| | `docs/re/27-screen-map.md` | 2 |
| `sub_36494` | 日本在這個遊戲裡是敵對勢力不是外交對象（`sub_36494`「日軍海外部隊正逼近」、 | `docs/mechanics/50-diplomacy.md`（另 1 份） | 3 |
| `sub_366B8` | `sub_3562B` / `sub_366B8` / `sub_10AB5` \| — \| 未讀 \| | `docs/re/29-ai-does-prepare-attacks.md` | 1 |
| `sub_368A8` | sub_39B6E **零命中** 直接呼叫者：['sub_2D812', 'sub_3562B', 'sub_368A8'] | `docs/re/12-ai-chain-has-no-attack.md` | 2 |
| `sub_36BC7` | `sub_36BC7(省)`：一張硬編碼的五省清單 — **confirmed（清單）** | `docs/mechanics/70-ai.md`（另 3 份） | 10 |
| `sub_36D74` | `Random(5)` \| 6 \| `sub_36D74`(3)、`sub_1AA28`、`sub_1FC1C`、`sub_35B19` \| | `docs/mechanics/70-ai.md` | 1 |
| `sub_371B9` | `sub_371B9` \| 3 \| 日本電報閣下 \| | `docs/re/27-screen-map.md` | 1 |
| `sub_37A81` | `sub_37A81`「美國援助軍事物資中」是**事件**不是玩家指令—— | `docs/mechanics/50-diplomacy.md`（另 1 份） | 4 |
| `sub_37EAE` | 欠餉／欠糧 \| `sub_37EAE` \| 欠餉 → 士氣降低；欠糧 → 體力降低 \| | `docs/mechanics/60-personnel.md`（另 3 份） | 13 |
| `sub_383EF` | `sub_383EF` \| `and al, 0FBh`，對全 39 省 \| 回合重置 \| | `docs/re/14-acted-flag-and-last-step.md`（另 1 份） | 2 |
| `sub_38444` | `sub_30003` 由 `sub_38444+97` 呼叫——**回合／月份的自動更新**。 | `docs/re/23-rescan-with-fixed-pattern.md`（另 1 份） | 2 |
| `sub_38DFE` | `WAR.EXE` 裡**只有一支函式**比較 `'6'`：`sub_38DFE`（`0x38DFE`， | `docs/mechanics/30-combat.md`（另 1 份） | 3 |
| `sub_3964E` | **傳入的是司令 ID 已經是 confirmed**——§5 從另一條路（`sub_3964E` 用同一個 | `docs/mechanics/70-ai.md`（另 8 份） | 15 |
| `sub_39B6E` | 戰鬥主迴圈 `sub_39B6E` 逐一處理戰場單位： | `docs/mechanics/30-combat.md`（另 9 份） | 25 |
| `sub_3A320` | **戰鬥中的 AI** \| ✅ **骨架已解**（`docs/re/31`）：`sub_3F698` 主入口 → `sub_3AB99` 決策鏈（與政略同構：兩分支、優先序、已決定旗標）→ 分派到 **13 種行動… | `docs/mechanics/70-ai.md`（另 1 份） | 10 |
| `sub_3A766` | 這解釋了整條鏈的形狀：`sub_3A766`（§6）檢查那些城市格有沒有被攻方站住 | `docs/re/31-battle-ai-chain.md` | 2 |
| `sub_3A817` | `sub_3A817` \| 0 \| **5.0** \| 第二方被壓到第一方的 1/5 以下 \| 成立才繼續 \| | `docs/re/31-battle-ai-chain.md`（另 1 份） | 9 |
| `sub_3A885` | if !決定: sub_3A885 ; 12 | `docs/re/31-battle-ai-chain.md` | 2 |
| `sub_3A8C8` | if !決定: sub_3A8C8 ; 19 | `docs/re/31-battle-ai-chain.md` | 2 |
| `sub_3A8F7` | `sub_3A8F7` 傳的是 `AX=5280h BX=1EB8h DX=2B85h`， | `docs/re/31-battle-ai-chain.md` | 8 |
| `sub_3A94E` | if !決定 且 byte_6FFCA & 4: sub_3A94E ; 16 / 17 | `docs/re/31-battle-ai-chain.md` | 4 |
| `sub_3A988` | **7**（80h）\| `sub_3A988` \| `word[796h + i×2] & 2000h` \| | `docs/re/31-battle-ai-chain.md` | 3 |
| `sub_3A9C9` | **6**（40h）\| `sub_3A9C9` \| `將領[word_64944].+4 == byte_6FFC4` \| | `docs/re/31-battle-ai-chain.md` | 3 |
| `sub_3A9F4` | sub_3A9F4 ; 2 | `docs/re/31-battle-ai-chain.md` | 2 |
| `sub_3AA51` | 3 \| `sub_3C89A` \| `sub_3AA51`／**預設** \| | `docs/re/31-battle-ai-chain.md`（另 1 份） | 8 |
| `sub_3AABA` | if !決定 且 byte_6FFCA & 4: sub_3AABA ; 4 | `docs/re/31-battle-ai-chain.md` | 2 |
| `sub_3AAF4` | sub_3AAF4 ; 不設值，收尾 | `docs/re/31-battle-ai-chain.md` | 1 |
| `sub_3AB99` | **戰鬥中的 AI** \| ✅ **骨架已解**（`docs/re/31`）：`sub_3F698` 主入口 → `sub_3AB99` 決策鏈（與政略同構：兩分支、優先序、已決定旗標）→ 分派到 **13 種行動… | `docs/mechanics/70-ai.md`（另 1 份） | 5 |
| `sub_3AC7B` | 分派表裡最短的一支：`sub_3B19C` 只有 15 行，轉手呼叫 `sub_3AC7B`： | `docs/re/31-battle-ai-chain.md` | 2 |
| `sub_3ACDF` | `sub_3ACDF`（460 行）讀完了。它是**六方向鄰接掃描**，尾段這一段解開 `+10`： | `docs/re/31-battle-ai-chain.md` | 3 |
| `sub_3B079` | 追 `sub_3B079`（100 行，`sub_3B31D` 對每個單位呼叫它）找到那一行： | `docs/re/31-battle-ai-chain.md` | 4 |
| `sub_3B19C` | 第二方 × 5 ≤ 第一方 → 決策值 A=11 → sub_3B19C(0) → 第一方勝 | `docs/mechanics/30-combat.md`（另 1 份） | 15 |
| `sub_3B1B1` | `sub_3B1B1`：把周圍的格依**防禦力**排序 | `docs/re/31-battle-ai-chain.md` | 5 |
| `sub_3B31D` | 2 \| `sub_3B31D` \| **佈防**——以主力為中心把周圍三圈的格發給沒去處的單位 \| | `docs/mechanics/30-combat.md`（另 1 份） | 8 |
| `sub_3B400` | 21. ⭐⭐ `sub_3B492` + `sub_3B400`：**挑最近的敵方城市** | `docs/re/31-battle-ai-chain.md`（另 1 份） | 4 |
| `sub_3B492` | sub_3B492(狀態, 0, &評分表, &城市清單, 城市數) ; ← **367 行，每個城市算一個 32-bit 分數** | `docs/re/31-battle-ai-chain.md`（另 1 份） | 6 |
| `sub_3B780` | sub_3B780(狀態, &var_22, 城市清單[i]) ; ← 分數 0 的城市 | `docs/re/31-battle-ai-chain.md` | 1 |
| `sub_3B8B0` | sub_3B8B0(狀態, 1, u) ; 一個都走不到 | `docs/re/31-battle-ai-chain.md` | 2 |
| `sub_3B9D8` | if 選中城市 == 0: sub_3B9D8(狀態, 1, 旗標陣列) ; 371 行，未讀 | `docs/re/31-battle-ai-chain.md` | 2 |
| `sub_3BCED` | 25. `sub_3BCED`：候選清單怎麼來的 — **部分** | `docs/re/31-battle-ai-chain.md` | 7 |
| `sub_3BF6A` | 剩下三支未讀：`sub_3BF6A`（175，城市空著）、`sub_3B9D8`（371，沒選到城市）、 | `docs/re/31-battle-ai-chain.md` | 6 |
| `sub_3C0D5` | if u != 0 且 將領[u].+8 == 0: sub_3C0D5(bp) ; ← **城市有守軍** | `docs/re/31-battle-ai-chain.md` | 5 |
| `sub_3C213` | 22. 目標選定之後：`sub_3C213` 的三向分流 | `docs/re/31-battle-ai-chain.md` | 5 |
| `sub_3C26A` | ⚠️ 幾支函式**寫不只一個值**（`sub_47EAA` 寫 1 與 3、`sub_3C26A` 寫 3 與 4）， | `docs/re/31-battle-ai-chain.md`（另 1 份） | 11 |
| `sub_3C40C` | sub_3C40C(狀態, &var_22) | `docs/re/31-battle-ai-chain.md` | 1 |
| `sub_3C633` | sub_3C633(狀態) | `docs/re/31-battle-ai-chain.md` | 1 |
| `sub_3C777` | 26. `sub_3C777`（143 行）：收尾補救，而且**保存-呼叫-還原** | `docs/re/31-battle-ai-chain.md` | 4 |
| `sub_3C89A` | 3 \| `sub_3C89A` \| **打城市**（預設行動）\| | `docs/mechanics/30-combat.md`（另 1 份） | 11 |
| `sub_3CA09` | 4 \| `sub_3CA09` \| **打敵方主力周邊** \| | `docs/mechanics/30-combat.md`（另 1 份） | 11 |
| `sub_3CBD9` | +0 \| byte \| `sub_3CBD9`／`sub_3D57B`／`sub_3DD05`／`sub_3DED9` \| **2** \| | `docs/re/31-battle-ai-chain.md` | 5 |
| `sub_3D0AC` | ⚠️ `CONTEXT.md` 把 `+9` 記成「所屬省」。這裡 `sub_3D0AC` 在**戰鬥中**寫它， | `docs/re/31-battle-ai-chain.md` | 17 |
| `sub_3D261` | `sub_3D261`（191 行）\| `+10`(word)、`+13`(byte) \| | `docs/re/31-battle-ai-chain.md` | 3 |
| `sub_3D57B` | 13 \| `sub_3D57B` \| **預設**（五步都沒決定時）\| | `docs/re/31-battle-ai-chain.md` | 5 |
| `sub_3D823` | +0 \| byte \| `sub_3D823`／`sub_3E24F` \| **3** \| | `docs/re/31-battle-ai-chain.md` | 5 |
| `sub_3DAEF` | if 將領[u].+12 == 0FFh: sub_3DAEF(...) | `docs/re/31-battle-ai-chain.md` | 4 |
| `sub_3DD05` | +0 \| byte \| `sub_3CBD9`／`sub_3D57B`／`sub_3DD05`／`sub_3DED9` \| **2** \| | `docs/re/31-battle-ai-chain.md` | 6 |
| `sub_3DED9` | +0 \| byte \| `sub_3CBD9`／`sub_3D57B`／`sub_3DD05`／`sub_3DED9` \| **2** \| | `docs/re/31-battle-ai-chain.md` | 13 |
| `sub_3E24F` | `sub_3E24F`／`sub_3E691` 未讀。**這是下一步最有價值的目標。** | `docs/re/31-battle-ai-chain.md` | 3 |
| `sub_3E691` | `sub_3E24F`／`sub_3E691` 未讀。**這是下一步最有價值的目標。** | `docs/re/31-battle-ai-chain.md` | 7 |
| `sub_3E93A` | 2 \| `sub_3DED9`、`sub_3E93A`、`sub_48A95` \| 11 支 \| | `docs/re/31-battle-ai-chain.md` | 1 |
| `sub_3ED4F` | 變數 \| `sub_3C777`(cl)、`sub_3D0AC`(cx)、`sub_3ED4F`(cx) \| — \| | `docs/re/31-battle-ai-chain.md` | 1 |
| `sub_3EF98` | if byte_6FFCA & 4: sub_3EF98 | `docs/re/31-battle-ai-chain.md` | 2 |
| `sub_3F0EF` | if byte_6AA84 & 1Fh != 2: sub_3F0EF | `docs/re/31-battle-ai-chain.md` | 4 |
| `sub_3F623` | if byte_6AA84 & 1Fh != 2: sub_3F623 | `docs/re/31-battle-ai-chain.md` | 9 |
| `sub_3F698` | **戰鬥中的 AI** \| ✅ **骨架已解**（`docs/re/31`）：`sub_3F698` 主入口 → `sub_3AB99` 決策鏈（與政略同構：兩分支、優先序、已決定旗標）→ 分派到 **13 種行動… | `docs/mechanics/70-ai.md`（另 2 份） | 5 |
| `sub_41513` | +5 \| **戰場格編號 0..195**，`0xFF` = 不在場上 \| `sub_50FF5`、`sub_41513`（`docs/re/07` §2）\| | `docs/mechanics/30-combat.md`（另 5 份） | 11 |
| `sub_4166E` | `cl`（`or al, 4`）\| `sub_41513`、`sub_4166E`、`sub_4E239` \| **只在戰鬥中** \| | `docs/re/21-develop-and-general-state.md`（另 1 份） | 2 |
| `sub_42B56` | sub_42B56(arg_4) ; 顯示訊息後結束 | `docs/re/09-ranged-attack.md` | 3 |
| `sub_42BC3` | `sub_42BC3(target, unit)` = 目標在不在攻擊範圍內 — **confirmed** | `docs/re/09-ranged-attack.md`（另 1 份） | 5 |
| `sub_42C8F` | `sub_42C8F` 開頭是一段遠程攻擊的守門： | `docs/mechanics/30-combat.md`（另 3 份） | 8 |
| `sub_4798B` | 1 \| `sub_4798B`、`sub_47EAA` \| 8 支（含 `sub_3D0AC`、`sub_3D823`、`sub_3DED9`）\| | `docs/re/31-battle-ai-chain.md` | 2 |
| `sub_47B6D` | 將領[u].+12 = sub_47B6D(狀態) ; 目標格 | `docs/re/31-battle-ai-chain.md` | 2 |
| `sub_47EAA` | ⚠️ 幾支函式**寫不只一個值**（`sub_47EAA` 寫 1 與 3、`sub_3C26A` 寫 3 與 4）， | `docs/re/31-battle-ai-chain.md` | 11 |
| `sub_483A5` | 傳的正是 **0**。另一個呼叫端 `sub_483A5+1BC` 還沒讀。 | `docs/re/08-combat-damage.md` | 2 |
| `sub_48A95` | 2 \| `sub_3DED9`、`sub_3E93A`、`sub_48A95` \| 11 支 \| | `docs/re/31-battle-ai-chain.md` | 1 |
| `sub_4A1C0` | `sub_4A1C0` 移動一格時做 `+7 -= 成本`，呼叫端先檢查 `+7 >= 成本` | `docs/mechanics/30-combat.md`（另 1 份） | 3 |
| `sub_4A470` | 成本表：`sub_4A470` 填，值來自 `byte_91E` — **confirmed** | `docs/re/07-battlefield-geometry.md` | 2 |
| `sub_4A583` | +21 \| **兵種編號**（強證據；值域含 1/4/5/6）。`== 1` 可穿越長城 \| `sub_4A583`、`sub_1629E` \| | `docs/mechanics/30-combat.md`（另 1 份） | 3 |
| `sub_4A841` | （`sub_4A841+383`、`sub_42C8F+6EC`）還沒讀。 | `docs/playtest/04-battlesim-dryrun.md`（另 1 份） | 2 |
| `sub_4ABFD` | 呼叫端 `sub_4ABFD` 先過三道 gate： | `docs/re/07-battlefield-geometry.md` | 1 |
| `sub_4E055` | sub_4E055 sub_4E239 sub_4E413 sub_4E579 sub_4E627 sub_4E65C | `docs/re/04-battle-module.md` | 1 |
| `sub_4E239` | `cl`（`or al, 4`）\| `sub_41513`、`sub_4166E`、`sub_4E239` \| **只在戰鬥中** \| | `docs/re/21-develop-and-general-state.md`（另 2 份） | 3 |
| `sub_4E413` | sub_4E055 sub_4E239 sub_4E413 sub_4E579 sub_4E627 sub_4E65C | `docs/re/04-battle-module.md` | 1 |
| `sub_4E579` | sub_4E055 sub_4E239 sub_4E413 sub_4E579 sub_4E627 sub_4E65C | `docs/re/04-battle-module.md` | 1 |
| `sub_4E627` | sub_4E055 sub_4E239 sub_4E413 sub_4E579 sub_4E627 sub_4E65C | `docs/re/04-battle-module.md` | 1 |
| `sub_4E65C` | 位移表由 `sub_4E65C` 填、邊界規則在 `sub_510E0`（`docs/re/07` §6）： | `docs/mechanics/30-combat.md`（另 3 份） | 5 |
| `sub_4E7D0` | sub_4E7D0 sub_4EA34 sub_4EB20 sub_4EC99* sub_4ED9B sub_4EECE | `docs/re/04-battle-module.md` | 1 |
| `sub_4EA34` | sub_4E7D0 sub_4EA34 sub_4EB20 sub_4EC99* sub_4ED9B sub_4EECE | `docs/re/04-battle-module.md` | 1 |
| `sub_4EB20` | sub_4E7D0 sub_4EA34 sub_4EB20 sub_4EC99* sub_4ED9B sub_4EECE | `docs/re/04-battle-module.md` | 1 |
| `sub_4EC99` | │ │ └─ +12A → sub_4EC99 讀 warpos.dat | `docs/re/04-battle-module.md` | 2 |
| `sub_4ED9B` | `WAR.EXE` 的 `sub_4ED9B` 是讀這個檔的函式，反組譯直接給出答案： | `docs/spec/04-battlefield-tiles.md`（另 1 份） | 3 |
| `sub_4EECE` | sub_4E7D0 sub_4EA34 sub_4EB20 sub_4EC99* sub_4ED9B sub_4EECE | `docs/re/04-battle-module.md` | 1 |
| `sub_4EF86` | sub_4EF86 sub_4EFBE* sub_4F0B0 sub_4F0E2 sub_4F2AC sub_4F468* | `docs/re/04-battle-module.md` | 1 |
| `sub_4EFBE` | │ └─ +170 → sub_4EFBE 讀 tername.dat | `docs/re/04-battle-module.md` | 3 |
| `sub_4F0B0` | sub_4EF86 sub_4EFBE* sub_4F0B0 sub_4F0E2 sub_4F2AC sub_4F468* | `docs/re/04-battle-module.md` | 1 |
| `sub_4F0E2` | +9 \| **所屬省編號**（部署時決定從哪一側進場）\| `sub_4F0E2` 寫進 `byte_6AB65`，`sub_41513` 比對 \| | `docs/mechanics/30-combat.md`（另 6 份） | 9 |
| `sub_4F2AC` | `+30` 士氣 \| `sub_4F2AC` \| **+cl（戰鬥中唯一的加項）** \| | `docs/re/21-develop-and-general-state.md`（另 1 份） | 2 |
| `sub_4F468` | `sub_4F468` 讀檔、`sub_3964E` 寫檔，兩支用**同一組偏移**在記錄與全域變數 | `docs/re/05-mem-war-record.md`（另 1 份） | 4 |
| `sub_4F5F3` | sub_4F5F3 sub_4F6FF sub_4F883 sub_4F8E5 sub_4F952 sub_4FA01 | `docs/re/04-battle-module.md` | 1 |
| `sub_4F6FF` | sub_4F5F3 sub_4F6FF sub_4F883 sub_4F8E5 sub_4F952 sub_4FA01 | `docs/re/04-battle-module.md` | 1 |
| `sub_4F883` | `sub_5AE48` \| `sub_39B6E`（**戰鬥主函式**）、`sub_4F883`、`sub_16D98` \| 戰鬥端 \| | `docs/re/29-ai-does-prepare-attacks.md`（另 1 份） | 3 |
| `sub_4F8E5` | sub_4F5F3 sub_4F6FF sub_4F883 sub_4F8E5 sub_4F952 sub_4FA01 | `docs/re/04-battle-module.md` | 1 |
| `sub_4F952` | sub_4F5F3 sub_4F6FF sub_4F883 sub_4F8E5 sub_4F952 sub_4FA01 | `docs/re/04-battle-module.md` | 1 |
| `sub_4FA01` | sub_4F5F3 sub_4F6FF sub_4F883 sub_4F8E5 sub_4F952 sub_4FA01 | `docs/re/04-battle-module.md` | 1 |
| `sub_4FAFC` | sub_4FAFC sub_4FC19 sub_4FC8B sub_4FCCC sub_4FE86 sub_4FEB7 | `docs/re/04-battle-module.md` | 1 |
| `sub_4FC19` | sub_4FAFC sub_4FC19 sub_4FC8B sub_4FCCC sub_4FE86 sub_4FEB7 | `docs/re/04-battle-module.md` | 1 |
| `sub_4FC8B` | sub_4FAFC sub_4FC19 sub_4FC8B sub_4FCCC sub_4FE86 sub_4FEB7 | `docs/re/04-battle-module.md` | 1 |
| `sub_4FCCC` | 196×196 成本矩陣誰讀 \| `sub_4FCCC` 建了 38,416 B（`dword_6A454`），用途未追 \| 未讀 \| | `docs/mechanics/30-combat.md`（另 2 份） | 3 |
| `sub_4FE86` | sub_4FAFC sub_4FC19 sub_4FC8B sub_4FCCC sub_4FE86 sub_4FEB7 | `docs/re/04-battle-module.md` | 1 |
| `sub_4FEB7` | sub_4FAFC sub_4FC19 sub_4FC8B sub_4FCCC sub_4FE86 sub_4FEB7 | `docs/re/04-battle-module.md` | 1 |
| `sub_4FEF0` | **有鐵路**（`sub_4FEF0`，覆蓋一切）\| **2** \| | `docs/re/07-battlefield-geometry.md`（另 1 份） | 4 |
| `sub_4FF3B` | sub_4FEF0 sub_4FF3B | `docs/re/04-battle-module.md` | 1 |
| `sub_50151` | 而 `sub_50151` 是這樣解碼的： | `docs/re/07-battlefield-geometry.md`（另 1 份） | 12 |
| `sub_501A3` | `sub_503BB`（375 行）與 `sub_501A3`（348 行）看起來嚇人，內容其實是**兩張表** | `docs/re/08-combat-damage.md` | 5 |
| `sub_503BB` | `sub_503BB`（375 行）與 `sub_501A3`（348 行）看起來嚇人，內容其實是**兩張表** | `docs/re/08-combat-damage.md`（另 1 份） | 7 |
| `sub_50600` | var_4 = Round( Real(sub_5A0B9(F)) × sub_50600(F) ÷ sub_50655(E) ) | `docs/re/08-combat-damage.md` | 5 |
| `sub_50655` | var_4 = Round( Real(sub_5A0B9(F)) × sub_50600(F) ÷ sub_50655(E) ) | `docs/re/08-combat-damage.md` | 5 |
| `sub_506B0` | `sub_506B0` 算出每格的基準成本（`docs/re/07` §7）： | `docs/mechanics/30-combat.md`（另 3 份） | 7 |
| `sub_50783` | `sub_50783` 每格呼叫 `sub_506B0` 算出基準成本填進 `byte_91E`， | `docs/re/07-battlefield-geometry.md` | 1 |
| `sub_507C7` | `sub_507C7`（兩支戰損函式都會在特定條件下呼叫它） | `docs/re/08-combat-damage.md` | 1 |
| `sub_50FF5` | +5 \| **戰場格編號 0..195**，`0xFF` = 不在場上 \| `sub_50FF5`、`sub_41513`（`docs/re/07` §2）\| | `docs/mechanics/30-combat.md`（另 2 份） | 9 |
| `sub_510E0` | 位移表由 `sub_4E65C` 填、邊界規則在 `sub_510E0`（`docs/re/07` §6）： | `docs/mechanics/30-combat.md`（另 2 份） | 11 |
| `sub_51399` | sub_51399(0, A.格, B.格) | `docs/re/08-combat-damage.md` | 1 |
| `sub_517BE` | sub_517BE(A, B) | `docs/re/08-combat-damage.md` | 1 |
| `sub_51844` | `arg_C != 0 且 arg_8 == 0` 時還會先跑 `sub_51844`（98 行，未讀）。 | `docs/re/08-combat-damage.md` | 2 |
| `sub_5190E` | W(u) = sub_5190E(u.+21) ; 第二張兵種權重表 | `docs/re/08-combat-damage.md`（另 2 份） | 7 |
| `sub_51972` | if var_8 × 4 <= var_4: sub_51972 | `docs/re/08-combat-damage.md`（另 1 份） | 6 |
| `sub_51B94` | → `sub_51B94`。兩支的戰損算法不同，**都還沒讀**。 | `docs/mechanics/30-combat.md`（另 1 份） | 4 |
| `sub_51D68` | 順帶訂正：`sub_51D68` 的 E／F | `docs/mechanics/30-combat.md`（另 3 份） | 8 |
| `sub_51EC0` | sub_51EC0(1, B) / sub_51EC0(1, A) | `docs/re/08-combat-damage.md`（另 1 份） | 2 |
| `sub_51F19` | ⚠️ `sub_51F19` **也會扣 `+17`**（linear 0x126699）。所以兵力不只在 | `docs/re/08-combat-damage.md`（另 1 份） | 4 |
| `sub_52129` | sub_52129(0, id) | `docs/mechanics/30-combat.md` | 1 |
| `sub_52222` | **20h（32）** \| `sub_11380`、`sub_52222`、`sub_524FC`、`sub_5CEAE` \| 政略 + 戰鬥 \| | `docs/re/21-develop-and-general-state.md` | 2 |
| `sub_524FC` | **20h（32）** \| `sub_11380`、`sub_52222`、`sub_524FC`、`sub_5CEAE` \| 政略 + 戰鬥 \| | `docs/re/21-develop-and-general-state.md` | 2 |
| `sub_528EE` | sub_528EE(2, A, B) | `docs/re/08-combat-damage.md` | 1 |
| `sub_52EEA` | sub_52EEA(結果, A, B) | `docs/re/08-combat-damage.md` | 1 |
| `sub_5301B` | ⚠️ **加成有開關。** `arg_A` 為 0 時整段跳過，而 `sub_5301B` 那條呼叫路徑 | `docs/re/08-combat-damage.md`（另 2 份） | 5 |
| `sub_530B4` | `sub_530B4` 算完之後對 A 與 B 的 `+17`（兵力）**各做一次減法** | `docs/mechanics/30-combat.md`（另 3 份） | 9 |
| `sub_534FF` | +14 \| **效忠的勢力領袖**（將領 ID，u16）。與省份記錄 `+20` 司令直接比較 \| `sub_534FF`、`sub_5A881`（`docs/re/07` §9）\| | `docs/mechanics/30-combat.md`（另 2 份） | 5 |
| `sub_53619` | 那是**發動結算**不是脫離戰鬥。`sub_53619` 只是成立後的另一道確認。 | `docs/re/31-battle-ai-chain.md` | 1 |
| `sub_5375E` | **1** \| `sub_12AAD`、`sub_13484`、`sub_5375E`、`sub_53B1E` \| 政略 + 戰鬥 \| | `docs/re/21-develop-and-general-state.md` | 1 |
| `sub_53B1E` | **1** \| `sub_12AAD`、`sub_13484`、`sub_5375E`、`sub_53B1E` \| 政略 + 戰鬥 \| | `docs/re/21-develop-and-general-state.md` | 1 |
| `sub_53C28` | `sub_53C28` 檢查單位周圍六格有沒有**同一方**（`+8` 相同）的單位， | `docs/mechanics/30-combat.md`（另 1 份） | 2 |
| `sub_53DA9` | +8 \| **攻方 1／守方 0**（強證據）\| `sub_41513` 寫 1、`sub_53DA9` 用它選顯示旗標（`docs/re/07` §8）\| | `docs/mechanics/30-combat.md`（另 1 份） | 9 |
| `sub_540B2` | sub_540B2 sub_54391 sub_546D1 sub_543C2 sub_54CFD | `docs/mechanics/30-combat.md` | 3 |
| `sub_54391` | 3i. 誰是攻方：`sub_54391` 與 `sub_54CFD` — **confirmed** | `docs/mechanics/30-combat.md`（另 1 份） | 6 |
| `sub_543C2` | `sub_543C2` 對**兩張都**做回合初始化（`+16=1`、`+7=+6`、`+13=0`）， | `docs/mechanics/30-combat.md`（另 1 份） | 5 |
| `sub_5446D` | 3d. 回合初始化：`sub_5446D` 與 `sub_545B0` — **confirmed** | `docs/mechanics/30-combat.md`（另 1 份） | 14 |
| `sub_545B0` | 3d. 回合初始化：`sub_5446D` 與 `sub_545B0` — **confirmed** | `docs/mechanics/30-combat.md`（另 1 份） | 8 |
| `sub_546D1` | `sub_546D1`（九步結算的第三步）： | `docs/mechanics/40-economy.md`（另 3 份） | 5 |
| `sub_54826` | `sub_54826` 裡有一段完整的算式： | `docs/mechanics/30-combat.md`（另 4 份） | 12 |
| `sub_54CFD` | 3i. 誰是攻方：`sub_54391` 與 `sub_54CFD` — **confirmed** | `docs/mechanics/30-combat.md`（另 1 份） | 5 |
| `sub_54DAC` | 3f. 佔領：`sub_54DAC` — **confirmed** | `docs/mechanics/30-combat.md`（另 2 份） | 9 |
| `sub_54E3B` | if arg_0 != 0: sub_54E3B(word_64942, word_64944) | `docs/re/31-battle-ai-chain.md`（另 1 份） | 8 |
| `sub_55632` | sub_55632(&清單, 1, 主將格) ; 收集 **37 格** | `docs/re/31-battle-ai-chain.md` | 13 |
| `sub_55AB7` | sub_55BCC(a, b) = sub_510E0(a, b) 或 sub_55AB7(a, b) | `docs/re/31-battle-ai-chain.md` | 3 |
| `sub_55BCC` | `sub_55BCC(a, b)` 是兩個判斷的 or： | `docs/re/31-battle-ai-chain.md` | 5 |
| `sub_55CEC` | `sub_55CEC(unit, cell, n)` 用六鄰位移表數**該格周圍的敵方單位數**， | `docs/mechanics/30-combat.md`（另 2 份） | 13 |
| `sub_55FBE` | 決策鏈開頭先跑 `sub_55FBE`： | `docs/re/31-battle-ai-chain.md` | 5 |
| `sub_567B9` | 10. `sub_567B9` = **移動單位** — **confirmed（欄位）** | `docs/re/31-battle-ai-chain.md` | 25 |
| `sub_579F4` | 旗標[sub_579F4(mode, u)] = 1 ; 標記「這個單位有任務了」 | `docs/re/31-battle-ai-chain.md` | 3 |
| `sub_57B15` | sub_57B15(arg_2, arg_4) ; 遠程攻擊 | `docs/re/09-ranged-attack.md` | 2 |
| `sub_58172` | sub_58172(單位) | `docs/re/07-battlefield-geometry.md` | 1 |
| `sub_58209` | 27. ⭐⭐ `sub_58209`（289 行）：照距離排序，逐一試路 | `docs/re/31-battle-ai-chain.md` | 8 |
| `sub_5867E` | if 單位.+21 != 1 and sub_5867E(格) != 0: | `docs/re/07-battlefield-geometry.md` | 3 |
| `sub_58854` | ⚠️ 仍未證實：`sub_58854`（494 行，產生格子清單的那支）還沒讀完。 | `docs/formats/05-tpc-tilesets.md`（另 1 份） | 5 |
| `sub_58D4A` | if (byte_6AA84 & 0C0h) != 0C0h 或 sub_58D4A() == 0 | `docs/re/31-battle-ai-chain.md`（另 1 份） | 8 |
| `sub_58DF9` | 敵方單位數 = sub_58DF9(0) | `docs/re/31-battle-ai-chain.md` | 2 |
| `sub_59CBF` | sub_37EAE 2字 #210 欠糧 sub_59CBF 2字 #64 檔案 | `docs/re/24-glyph-draw-family.md`（另 1 份） | 4 |
| `sub_5A031` | +7 = +6 sub_5A031(id) | `docs/mechanics/30-combat.md` | 1 |
| `sub_5A0B9` | （`sub_5A0B9`，`docs/re/08` §4e）： | `docs/mechanics/70-ai.md`（另 8 份） | 30 |
| `sub_5A3B2` | sub_5A3B2(5, B) / sub_5A3B2(5, A) | `docs/re/08-combat-damage.md` | 1 |
| `sub_5A467` | ⚠️ 這與 `AddResource`（`sub_5A467`）的「直接截斷」是**不同的行為**。 | `docs/mechanics/40-economy.md`（另 2 份） | 13 |
| `sub_5A4BC` | **但不是。** 每個分支做的事是把一組座標傳給 `sub_5A4BC`： | `docs/re/04-battle-module.md`（另 1 份） | 2 |
| `sub_5A541` | sub_5A541(當前省, [arg_0+8]) | `docs/mechanics/30-combat.md` | 1 |
| `sub_5A881` | `sub_5A881`（數量）與 `sub_306CF`（兵力總和）的篩選條件一字不差： | `docs/mechanics/20-military.md`（另 6 份） | 15 |
| `sub_5A907` | enemy = sub_5A907(該省司令, …, 該省) 收集到的敵對鄰省數 | `docs/mechanics/70-ai.md`（另 3 份） | 8 |
| `sub_5A9F6` | 3v. ✅ 「被包圍就突圍」`sub_5A9F6` / `sub_170C0` — **confirmed** | `docs/mechanics/70-ai.md`（另 2 份） | 16 |
| `sub_5ADA6` | sub_5ADA6(當前省) | `docs/mechanics/30-combat.md` | 1 |
| `sub_5AE48` | 戰鬥端 \| `sub_5AE48`／`sub_4F0E2` ← **戰鬥主函式 `sub_39B6E`** \| 讀 + 寫 \| | `docs/mechanics/70-ai.md`（另 1 份） | 3 |
| `sub_5B76E` | `sub_5B76E(a, b)` = a 是不是 b 的鄰省 — **confirmed** | `docs/mechanics/70-ai.md`（另 3 份） | 9 |
| `sub_5B7DC` | `sub_5B7DC(省編號)` 就是「找可攻打的鄰省」，回傳第一個符合的省編號（0 = 沒有）： | `docs/mechanics/20-military.md`（另 7 份） | 23 |
| `sub_5B983` | mine = ProvinceStrength(當前省) ; sub_5B983 | `docs/mechanics/70-ai.md`（另 5 份） | 16 |
| `sub_5BA01` | ⚠️ 仍未讀：`sub_5BA56`（模式 10，344 行，還呼叫 `sub_5BA01`）。 | `docs/re/15-pathfinding-weights.md` | 1 |
| `sub_5BA56` | `arg_4` 選權重：`10` 走 `sub_5BA56`、`11` 走 `sub_5BD7B`。 | `docs/mechanics/70-ai.md`（另 3 份） | 6 |
| `sub_5BD7B` | `arg_4` 選權重：`10` 走 `sub_5BA56`、`11` 走 `sub_5BD7B`。 | `docs/mechanics/70-ai.md`（另 2 份） | 6 |
| `sub_5BEFA` | 只是換到省份層。⚠️ 兩支權重函式與 `sub_5BEFA` 都**未讀**。 | `docs/mechanics/70-ai.md` | 2 |
| `sub_5BF74` | `sub_15CE3` \| **未讀**（302 行，也呼叫 `sub_5BF74` 與 `sub_5A907`）\| | `docs/mechanics/70-ai.md`（另 3 份） | 11 |
| `sub_5C155` | `sub_5C155` \| 1 \| 日軍 \| | `docs/re/27-screen-map.md` | 1 |
| `sub_5C742` | `sub_5C742` \| 1 \| 五省聯軍 \| | `docs/re/27-screen-map.md` | 1 |
| `sub_5C7FE` | 條件 1：司令是不是十大勢力之一（`sub_5C7FE`） | `docs/mechanics/70-ai.md`（另 5 份） | 19 |
| `sub_5C9C6` | `sub_5C9C6` \| 7 \| 第頁第行倒數第行 \| | `docs/re/27-screen-map.md` | 1 |
| `sub_5CEAE` | **20h（32）** \| `sub_11380`、`sub_52222`、`sub_524FC`、`sub_5CEAE` \| 政略 + 戰鬥 \| | `docs/re/21-develop-and-general-state.md` | 1 |
| `sub_5D670` | `sub_5D670` \| `sub_5D837` \| `sub_5D739` \| 1 字 \| `1Eh` = 30 \| `1.15` \| | `docs/re/24-glyph-draw-family.md` | 3 |
| `sub_5D695` | `sub_5D695` \| `sub_5D9DA` \| `sub_5D895` \| 2 字 \| `3Ch` = 60 \| `2.15` \| | `docs/re/24-glyph-draw-family.md` | 3 |
| `sub_5D6BE` | `sub_5D6BE` \| `sub_5DB7D` \| `sub_5DA38` \| 3 字 \| `5Ah` = 90 \| `3.15`、`MAN*15` \| | `docs/re/24-glyph-draw-family.md` | 3 |
| `sub_5D710` | `sub_5D710` \| `sub_5DEEF` \| `sub_5DD85` \| 5 字 \| `96h` = 150 \| `FAN(N).15`、`JAPAN*.15`、`TN15.N` \| | `docs/re/24-glyph-draw-family.md` | 2 |
| `sub_5D739` | `sub_5D670` \| `sub_5D837` \| `sub_5D739` \| 1 字 \| `1Eh` = 30 \| `1.15` \| | `docs/re/24-glyph-draw-family.md` | 1 |
| `sub_5D837` | `sub_5D670` \| `sub_5D837` \| `sub_5D739` \| 1 字 \| `1Eh` = 30 \| `1.15` \| | `docs/re/24-glyph-draw-family.md` | 1 |
| `sub_5D895` | `sub_5D695` \| `sub_5D9DA` \| `sub_5D895` \| 2 字 \| `3Ch` = 60 \| `2.15` \| | `docs/re/24-glyph-draw-family.md` | 1 |
| `sub_5D9DA` | `sub_5D695` \| `sub_5D9DA` \| `sub_5D895` \| 2 字 \| `3Ch` = 60 \| `2.15` \| | `docs/re/24-glyph-draw-family.md` | 1 |
| `sub_5DA38` | `sub_5D6BE` \| `sub_5DB7D` \| `sub_5DA38` \| 3 字 \| `5Ah` = 90 \| `3.15`、`MAN*15` \| | `docs/re/24-glyph-draw-family.md` | 2 |
| `sub_5DB7D` | `sub_5D6BE` \| `sub_5DB7D` \| `sub_5DA38` \| 3 字 \| `5Ah` = 90 \| `3.15`、`MAN*15` \| | `docs/re/24-glyph-draw-family.md` | 1 |
| `sub_5DD85` | `sub_5D710` \| `sub_5DEEF` \| `sub_5DD85` \| 5 字 \| `96h` = 150 \| `FAN(N).15`、`JAPAN*.15`、`TN15.N` \| | `docs/re/24-glyph-draw-family.md` | 1 |
| `sub_5DEEF` | `sub_5D710` \| `sub_5DEEF` \| `sub_5DD85` \| 5 字 \| `96h` = 150 \| `FAN(N).15`、`JAPAN*.15`、`TN15.N` \| | `docs/re/24-glyph-draw-family.md` | 1 |
| `sub_5E4FC` | 2. 字模載入：`sub_5E4FC` | `docs/re/01-war-exe-turbo-pascal.md` | 1 |
| `sub_5EB1B` | IOResult <> 0 → sub_5EB1B(2) ← 跳「請插入磁片」提示 | `docs/re/01-war-exe-turbo-pascal.md` | 3 |
