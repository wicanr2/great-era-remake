package main

// 原版 byte_6FE85 的每一級會呼叫 DELAY(0x190)，也就是 400 ms。
// Ebiten 預設每秒 60 次 Update，因此每級是 24 ticks。
const messageTicksPerUnit = 24

type messageQueue struct {
	units     int
	remaining int
	items     []string
}

func newMessageQueue(units int) *messageQueue {
	q := &messageQueue{}
	q.SetUnits(units)
	return q
}

func (q *messageQueue) SetUnits(units int) {
	if units < 1 || units > 10 {
		units = 5
	}
	q.units = units
}

func (q *messageQueue) Push(msg string) {
	if msg == "" {
		return
	}
	q.items = append(q.items, msg)
	if len(q.items) == 1 {
		q.remaining = q.units * messageTicksPerUnit
	}
}

func (q *messageQueue) Active() bool { return q != nil && len(q.items) > 0 }

func (q *messageQueue) Current() string {
	if !q.Active() {
		return ""
	}
	return q.items[0]
}

// Tick 回報畫面是否需要重畫（當前訊息消失或切到下一則）。
func (q *messageQueue) Tick() bool {
	if !q.Active() {
		return false
	}
	q.remaining--
	if q.remaining > 0 {
		return false
	}
	q.items = q.items[1:]
	if len(q.items) > 0 {
		q.remaining = q.units * messageTicksPerUnit
	}
	return true
}
