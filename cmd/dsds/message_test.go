package main

import "testing"

func TestMessageQueueUsesOriginalFourHundredMillisecondUnits(t *testing.T) {
	q := newMessageQueue(2)
	q.Push("第一則")
	for i := 0; i < 2*messageTicksPerUnit-1; i++ {
		if changed := q.Tick(); changed || q.Current() != "第一則" {
			t.Fatalf("tick %d 過早移除訊息", i)
		}
	}
	if !q.Tick() || q.Active() {
		t.Fatal("原版等待長度結束後沒有移除訊息")
	}
}

func TestMessageQueuePreservesOrderAndDefaultsInvalidUnits(t *testing.T) {
	q := newMessageQueue(99)
	q.Push("一")
	q.Push("二")
	for i := 0; i < 5*messageTicksPerUnit; i++ {
		q.Tick()
	}
	if q.Current() != "二" {
		t.Fatalf("預設時間後 current=%q", q.Current())
	}
}
