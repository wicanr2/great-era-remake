package mobile

// LifecyclePhase 是平台 adapter 可觀測的最小生命週期狀態。它不匯入
// Android／Ebiten；原生 binding 只需把 onPause／onResume／onDestroy 映射成
// Event，再依 Effect 控制遊戲與音訊。
type LifecyclePhase uint8

const (
	PhaseActive LifecyclePhase = iota + 1
	PhaseBackground
	PhaseDestroyed
)

// LifecycleEvent 是平台生命週期的窄邊界。重複事件必須是 no-op，避免背景
// 切回時重複派送輸入或關閉同一個音訊 player。
type LifecycleEvent uint8

const (
	EventPause LifecycleEvent = iota + 1
	EventResume
	EventDestroy
)

// LifecycleEffect 是一次狀態轉換對平台 adapter 的有界指示。
type LifecycleEffect struct {
	Changed     bool
	PauseAudio  bool
	ResumeAudio bool
	CloseAudio  bool
	ResetInput  bool
}

// LifecycleGate 防止重複 pause／resume／destroy 副作用。它不直接派送
// `Action`；呼叫端在 ResetInput 後清除 pointer／touch 按下狀態，再由正常
// 玩家路徑決定下一個 Action。
type LifecycleGate struct {
	phase LifecyclePhase
}

func NewLifecycleGate() *LifecycleGate { return &LifecycleGate{phase: PhaseActive} }

func (g *LifecycleGate) Phase() LifecyclePhase {
	if g == nil || g.phase == 0 {
		return PhaseDestroyed
	}
	return g.phase
}

// AcceptInput 只有在前景 active 時為真。背景／銷毀期間的輸入應由平台層
// 丟棄，不可排隊到恢復後變成重複指令。
func (g *LifecycleGate) AcceptInput() bool { return g != nil && g.phase == PhaseActive }

// Apply 套用生命週期事件；已經處於目標狀態時回傳 Changed=false 的 no-op。
func (g *LifecycleGate) Apply(event LifecycleEvent) LifecycleEffect {
	if g == nil || g.phase == PhaseDestroyed {
		return LifecycleEffect{}
	}
	switch event {
	case EventPause:
		if g.phase != PhaseActive {
			return LifecycleEffect{}
		}
		g.phase = PhaseBackground
		return LifecycleEffect{Changed: true, PauseAudio: true, ResetInput: true}
	case EventResume:
		if g.phase != PhaseBackground {
			return LifecycleEffect{}
		}
		g.phase = PhaseActive
		return LifecycleEffect{Changed: true, ResumeAudio: true, ResetInput: true}
	case EventDestroy:
		g.phase = PhaseDestroyed
		return LifecycleEffect{Changed: true, PauseAudio: true, CloseAudio: true, ResetInput: true}
	default:
		return LifecycleEffect{}
	}
}
