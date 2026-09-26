# SPEC-40：Modern 人物檔案肖像登錄 P0

> **狀態：READY（資料契約、權利 gate 與無圖後備畫面；不含批次蒐圖）**  
> 日期：2026-08-12  
> 範圍：Modern 人物檔案可選肖像的離線資料登錄、權利查證、完整性驗證與發行 notice。

## 1. 目的與邊界

人物肖像是 remake 的文化詮釋層，不是原版規則、不是將領數值，也不是人物自傳翻譯的一部分。
使用者已確認 A+B 素材政策：可驗證公有領域／CC0，以及附完整作者、授權、來源與署名資訊的
CC BY／CC BY-SA 圖像。所有圖像必須隨包離線可用；不得在遊戲執行期向網路請求圖片。

本切片只建立一份語系無關的登錄，讓 Modern 人物檔案能安全顯示「已審核肖像」或中性檔案卡
後備畫面（fallback）。它不新增人物、不補寫歷史、不改變玩法或存檔，也不輸入、轉存、散布
原版 `HEAD1/2.TPC` 或任何其他原版美術。

## 2. 檔案與資料契約

登錄檔固定為 `translations/shared/portraits.json`；圖檔固定置於
`translations/shared/portraits/`，P0 只接受 `.jpg`／`.jpeg`；WebP 要等到有鎖版純 Go
decoder 與獨立測試才可另立規格。禁止 `.png`，以維持原版展示截圖白名單只限
`docs/images/*.png` 的既有 deny 規則。

```json
{
  "schema_version": "1",
  "portraits": [
    {
      "person_id": 1,
      "status": "available",
      "asset": "portraits/0001-example.jpg",
      "asset_sha256": "64 個小寫十六進位字元",
      "identity": {
        "name": "蔣中正",
        "verification": "confirmed",
        "reference_url": "https://www.wikidata.org/wiki/Q16574"
      },
      "rights": {
        "kind": "public-domain",
        "license_url": "https://…",
        "author": "作者或 Unknown author（不可留白）",
        "attribution": "玩家可見的署名文字",
        "source_page": "https://commons.wikimedia.org/wiki/File:…",
        "retrieved_at": "2026-08-12"
      },
      "review": {
        "identity": "confirmed",
        "rights": "confirmed",
        "reviewed_at": "2026-08-12"
      },
      "crop": { "x": 0, "y": 0, "width": 1, "height": 1 }
    }
  ]
}
```

`person_id` 指向現有 `PeopleDB` 的 canonical 人物 id，不跟著繁中／英／日文字複製。跨期別的
同一人物共用同一筆肖像；`#274 無省長` 永遠不得有肖像。`crop` 是原始圖像寬高正規化後的
半開矩形，四個數皆在 `[0,1]`，且 `x + width`、`y + height` 不得超過 `1`。

缺少該人物登錄不是錯誤：renderer 必須顯示中性檔案卡。`status=available` 則是失敗即關閉：
缺欄位、重複 id、跨目錄路徑、錯誤副檔名、雜湊不符、未知人物、未確認身分／權利，或不在
允許授權列舉中的資料都不得顯示圖片。

## 3. 權利與發行 gate

`rights.kind` 只允許：

| 值 | 條件 |
|---|---|
| `public-domain` | 有可回查的公有領域依據與來源頁；不以「年代看起來很久」推定 |
| `cc0-1.0` | 來源頁明示 CC0 1.0；仍記錄作者、來源與擷取日 |
| `cc-by-4.0` | 完整作者、來源頁、授權 URL 與玩家可見署名文字必填 |
| `cc-by-sa-4.0` | 同 CC BY，另把相同方式分享條件列入發行 notice |

每個 `available` 圖檔必須以 SHA-256 對登錄驗證，並由產生器彙整為
`docs/licenses/portraits-NOTICE.md`。notice 至少逐筆列出人物、檔名、作者、來源頁、
授權、署名文字、擷取日與 SHA-256。桌面 ZIP、macOS app 與 Linux AppImage 若有任何肖像，
必須一併帶入圖檔、登錄及 notice；否則打包失敗。P0 在人物檔案肖像欄下顯示短署名，完整
資料在隨包 notice；不以只有開發文件的署名取代玩家可見資訊。H4 的「資料與授權」唯讀控制項
屬於後續畫面規格，不得把它誤稱為本 P0 已完成的互動。

## 4. 呈現與輸入

Modern 人物檔案保留全域導覽，將肖像放在身份區的固定欄位，文字區不因有無圖片改變行寬。
有通過 gate 的肖像時顯示裁切後圖像、人物名與短署名；沒有肖像或載入失敗時顯示標示為
「尚未找到可驗證肖像」的中性檔案卡，不能以 AI 生成臉孔、隨機照片或原版圖像代替。完整
「資料與授權」唯讀頁與其鍵盤、滑鼠、觸控 action 留待 H4 總畫面規格定案後另接，P0 不假裝
已有這個入口。

640×350 `retro` 自傳不顯示本切片的圖片，避免把高解析 Modern 差異誤加回原味畫面。

## 5. 驗收

1. 登錄 parser 測試覆蓋：合法公有領域、CC0、CC BY、CC BY-SA；拒絕缺署名、未知授權、
   缺失或錯誤雜湊、路徑穿越、未知／排除人物與不合法 crop。
2. renderer 測試覆蓋：有肖像、缺登錄、檔案不見、雜湊失敗四種狀態；後三者均走同一中性
   檔案卡，不留下空白或顯示不可信圖像。
3. notice 產生器對每筆 `available` 產出完整欄位；release 包驗證登錄、圖檔、notice 一致。
4. 全量 `go test -count=1 ./...`、`tools/check_no_cgo.sh`、`tools/deny_scan.sh --all` 與
   `git diff --check` 通過。圖像不取自 `workplace/orig/`，也不出現在 `docs/images/` 以外的
   PNG 路徑。
