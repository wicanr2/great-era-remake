# SPEC-25：戰鬥狀態副本寫回 M1

狀態：**READY**  
日期：2026-08-10

## 1. 目的

把 remake 已經實際跑過的戰鬥快照寫回兩份玩家副本：`SAVE(N).DT2` 與
`MEM_WAR.DAT`。這個切片不再等待 DOSBox 取樣；它採用已證實的欄位，並把
尚未解出的欄位保留在原始 bytes 中。

## 2. 證據與明確差異

`docs/re/05-mem-war-record.md` 已確認兩份檔案都是 39 筆、每筆 469 bytes，並確認：

- `+0..+7` 是第一方的四項資源快照（黃金、糧食、彈藥、燃料）；第二方同名
  執行期欄位在相鄰記憶體，並沒有在這 8 bytes 內另存一份。
- `+28..+47` 與 `+48..+67` 是兩方最多 10 個參戰將領 ID。
- `+68..+267` 與 `+268..+467` 是兩方 runtime 將領 ID 清單（各 100 個 u16）。
- `+468` 是攻方來源省份。

原版在立即撤退的已證實樣本不寫回；remake 仍維持這個行為。其他已結束戰鬥
則由 remake 主動保存目前快照，包含原版可能在勝負分出後不經過
`sub_3964E` 的分支。這是有意揭露的 remake 持久化差異，不是宣稱已達原版
寫回時機等價。

## 3. 寫回契約

- `BattleState.ApplyRemakeSnapshot` 只接受合法來源省、最多 10 個且非零的兩方
  參戰將領 ID；輸入不合法就 fail-closed。
- 寫入 `Header`、兩組 roster、兩組 runtime ID 陣列與 `Trailing`；`SlotsA/B`
  及其他未解 bytes 一律從 `Raw` 原樣保留，不把將領 ID 猜成 1-byte 槽位。
- runtime 陣列是 remake 的參戰快照投影：先清空，再按輸入順序填入前 100 槽。
  這個投影不等同於已證實的原版同步時機。
- 執行檔只把結果寫到 `-save` 同目錄的 `.DT2` 與 `MEM_WAR.DAT` 副本，絕不寫入
  `workplace/orig/`。兩份來源檔都未載入時，戰鬥仍可運行但不宣稱已持久化；只載入
  一份時則在結算時回報錯誤，避免半套寫回。

## 4. 驗收

1. 純 Go 測試驗證欄位投影、長度／值域檢查，以及未知 bytes byte-for-byte 保留。
2. `cmd/dsds` 在戰鬥全滅、補給見底、AI 必勝與回合上限收尾後呼叫同一寫回入口；
   立即撤退不呼叫它。
3. Docker 內執行 `go test -count=1 ./...`、Windows `CGO_ENABLED=0` 建置、
   `tools/check_no_cgo.sh`、`tools/deny_scan.sh --all`、`git diff --check`。

