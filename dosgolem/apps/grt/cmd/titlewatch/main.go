// Command titlewatch 固定步數開到 title，再餵鍵並攔讀取者。
//
//   go run ./apps/grt/cmd/titlewatch -exe /orig/GRT.EXE -root /orig
//   go run ./apps/grt/cmd/titlewatch -exe /orig/SDFA.EXE -queue GRT.EXE -root /orig -basedelta 0x96E0 -ph1 45000000 -ph2 60000000
//
// 不用 ScreenIdle 分段（載入期畫面長期靜止會誤判）：
// 跑 ph1 步（logo 等待迴圈）→ 餵 CR → 跑 ph2 步（title）→ 掛鉤
// sub_11925（AH=0B 輪詢）、sub_10778（AH=3F 讀）、sub_1173D（讀鍵）、
// 0x1177F（按鍵注入）、音效閘門 0x144A6／0x1346C／0x1382C → 餵 CR → 跑 10M 步。
// 印出誰 poll、誰 read（handle＝Arg(0)，0 即 stdin）、閘門、停哪、螢幕簽名。
//
// -basedelta 是 IDA 線性減執行期線性的差：直接載入為 0xEF00；
// SDFA 常駐後 GRT 重定位（PSP 0682），實測為 0x96E0。
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/wicanr2/dosgolem/oracle"
)

// hookAddr 把 IDA 線性位址換成指定 basedelta 下的執行期位址。
// 不用 o.IDA 是因為常駐重定位後 o 內建的差值不再成立。
func hookAddr(baseDelta uint32, ida uint32) oracle.Addr {
	run := ida - baseDelta
	return oracle.Addr{Seg: uint16(run >> 4), Off: uint16(run & 0xF)}
}

func splitComma(s string) []string {
	var out []string
	for _, q := range strings.Split(s, ",") {
		if q = strings.TrimSpace(q); q != "" {
			out = append(out, q)
		}
	}
	return out
}

func main() {
	exe := flag.String("exe", "", "首支程式路徑")
	root := flag.String("root", ".", "原版素材目錄")
	queue := flag.String("queue", "", "接著跑的程式（例 GRT.EXE），逗號分隔")
	delta := flag.Uint64("basedelta", 0xEF00, "IDA 線性減執行期線性")
	ph1 := flag.Uint64("ph1", 25_000_000, "第一段步數（到 logo 等待）")
	ph2 := flag.Uint64("ph2", 55_000_000, "第二段步數（到 title）")
	flag.Parse()
	if *exe == "" {
		fmt.Fprintln(os.Stderr, "要給 -exe")
		os.Exit(2)
	}
	o, err := oracle.Load(*exe, *root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "載入失敗：", err)
		os.Exit(1)
	}
	for _, q := range splitComma(*queue) {
		o.Enqueue(q, "")
	}
	bd := uint32(*delta)

	var polls uint64
	pollCallers := map[uint32]int{}
	o.OnCall(hookAddr(bd, 0x11925), func(o *oracle.Oracle) {
		polls++
		pollCallers[o.Caller().Linear()]++
	})
	// 重定位裁決：trace 實證 AH=0B INT 位於 063C:10D3（線性 7493）。
	// 若遊戲自我重定位，sub_11925 有兩個副本：
	// 未重定位進入點（basedelta 下）與重定位進入點（063C:10C8）。
	// 兩邊計數＋AH=0B log 計數對拍即知真相。
	var pollsTrace uint64
	o.OnCall(oracle.Addr{Seg: 0x063C, Off: 0x10D3}, func(o *oracle.Oracle) {
		pollsTrace++
	})
	var pollsEntry uint64
	o.OnCall(oracle.Addr{Seg: 0x063C, Off: 0x10C8}, func(o *oracle.Oracle) {
		pollsEntry++
	})
	type rd2 struct {
		handle uint16
		a1, a2, a3, a4 uint16
		caller uint32
	}
	reads := []rd2{}
	o.OnCall(hookAddr(bd, 0x10778), func(o *oracle.Oracle) {
		reads = append(reads, rd2{handle: o.Arg(0), a1: o.Arg(1), a2: o.Arg(2), a3: o.Arg(3), a4: o.Arg(4), caller: o.Caller().Linear()})
	})
	keyReads := map[uint32]int{}
	o.OnCall(hookAddr(bd, 0x1173D), func(o *oracle.Oracle) {
		keyReads[o.Caller().Linear()]++
	})
	var injects uint64
	o.OnCall(hookAddr(bd, 0x1177F), func(o *oracle.Oracle) {
		injects++
	})
	gateHits := []string{}
	for _, entry := range []uint32{0x144A6, 0x1346C, 0x1382C} {
		e := entry
		o.OnCall(hookAddr(bd, e), func(o *oracle.Oracle) {
			gateHits = append(gateHits, fmt.Sprintf("gate=%05X callerRT=%05X", e, o.Caller().Linear()))
		})
	}

	if err := o.RunUntil(oracle.Steps(*ph1)); err != nil {
		fmt.Println("第一段失敗：", err)
		os.Exit(1)
	}
	fmt.Println("第一段到，poll:", polls)
	o.Type("\r")
	if err := o.RunUntil(oracle.Steps(*ph2)); err != nil {
		fmt.Println("第二段失敗：", err)
		os.Exit(1)
	}
	px := o.Indexed()
	nz := 0
	for _, v := range px {
		if v != 0 {
			nz++
		}
	}
	fmt.Printf("第二段到：poll=%d read=%d 停在 %s 非零像素=%d\n", polls, len(reads), o.IP(), nz)
	for _, h := range gateHits {
		fmt.Println("  " + h)
	}
	o.Type("\r")
	if err := o.RunUntil(oracle.Steps(10_000_000)); err != nil {
		fmt.Println("餵鍵後停止：", err)
	}
	fmt.Printf("餵鍵後：poll=%d pollTrace10D3=%d pollEntry10C8=%d read=%d 停在 %s\n",
		polls, pollsTrace, pollsEntry, len(reads), o.IP())
	fmt.Println("poll 呼叫端（執行期線性）：")
	for a, n := range pollCallers {
		fmt.Printf("  %05X: %d\n", a, n)
	}
	fmt.Println("sub_1173D 讀鍵呼叫端（執行期線性）：")
	for a, n := range keyReads {
		fmt.Printf("  %05X: %d\n", a, n)
	}
	fmt.Printf("按鍵注入命中 %d 次\n", injects)
	fmt.Println("read（handle／callerRT）：")
	for _, r := range reads {
		fmt.Printf("  handle=%d callerRT=%05X\n", r.handle, r.caller)
	}
}
