package game

// 原版的亂數產生器：Turbo Pascal 6.0 的 32-bit LCG。
//
// `docs/playtest/12` 證實徵稅的收入帶亂數（同省同操作結果不同），
// 而 `CLAUDE.md` §9 要求「截圖驗收要帶固定亂數種子」——
// 沒有可重現的亂數，任何帶隨機的規則都無法與原版逐次對照。
//
// 這一支把那個門檻拆掉了：**序列完全可重現**（`docs/re/17`）。

// RandMultiplier 是 Turbo Pascal 的 LCG 乘數。
//
// 從 `WAR.EXE` 的 runtime 段反組譯出來（`docs/re/17` §2）——
// 那段程式用 16-bit 指令手工展開 32-bit 乘法，移位序列
// 逐位對上 `0x08088405`：
//
//	seed_lo × 0x0808：cx <<= 3 三次，然後 ch += cl（= ×0x101）
//	seed_hi × 0x8405：×1 + ×4 + ×0x400 + ×0x8000
const RandMultiplier uint32 = 0x08088405 // 134,775,813

// RandSeedAddr 是 `RandSeed` 在 `WAR.EXE` 資料段的位址。
//
// 記著它是為了**動態驗證**：真要與原版逐次對照，
// 可以在 DOSBox 裡讀這一格比對序列。
const RandSeedAddr = 0x072A

// Rand 是原版的亂數產生器。零值可用（種子 0，與原版沒跑 Randomize 時一致）。
//
// **不是 thread-safe**，也刻意不做——原版是單執行緒的，
// 而規則層要的正是「同樣的操作序列產生同樣的結果」。
type Rand struct {
	seed uint32
}

// NewRand 用指定的種子建一個產生器。
//
// 原版的種子來自 `Randomize`（`int 21h` / `AH=2Ch` 取系統時間，
// `CX:DX` 直接寫進 `RandSeed`），所以每次開遊戲都不同。
// **remake 的測試與截圖驗收一律用固定種子。**
func NewRand(seed uint32) *Rand { return &Rand{seed: seed} }

// Seed 取得目前的種子，方便存檔或除錯時記錄。
func (r *Rand) Seed() uint32 { return r.seed }

// SetSeed 直接設種子。
func (r *Rand) SetSeed(s uint32) { r.seed = s }

// next 推進一步並回傳新的種子，語意照原版：
//
//	mov  ax, [072Ah]        ; seed 低 16
//	mov  bx, [072Ch]        ; seed 高 16
//	...32-bit 乘法展開...
//	add  ax, 1
//	adc  dx, 0
//	mov  [072Ah], ax
//	mov  [072Ch], dx
func (r *Rand) next() uint32 {
	r.seed = r.seed*RandMultiplier + 1
	return r.seed
}

// Word 對應 `Random(range: Word): Word`，回 `[0, n)`。
//
// Turbo Pascal 的取值方式是「新種子（無號）乘上 range，取高 32 位」
// ——等價於 `seed / 2^32 * n`，而不是取餘數。**這個差別會影響分佈**，
// 所以不能寫成 `seed % n`。
func (r *Rand) Word(n uint16) uint16 {
	if n == 0 {
		return 0
	}
	return uint16((uint64(r.next()) * uint64(n)) >> 32)
}

// Int 是 `Word` 的方便包裝，n <= 0 時回 0。
func (r *Rand) Int(n int) int {
	if n <= 0 || n > 0xFFFF {
		return 0
	}
	return int(r.Word(uint16(n)))
}

// Real 對應無參數的 `Random: Real`，回 `[0, 1)`。
//
// 原版那條路徑走 8087 模擬（`int 35h`／`39h`／`3Dh`），
// 值是 `seed / 2^32`。
func (r *Rand) Real() float64 {
	return float64(r.next()) / 4294967296.0
}
