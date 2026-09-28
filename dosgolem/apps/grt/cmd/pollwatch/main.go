// Command pollwatch 攔 GRT.EXE 的按鍵輪詢與中斷向量存取，印呼叫端。
//
//   go run ./apps/grt/cmd/pollwatch -exe /orig/GRT.EXE -root /orig -steps 20000000
//
// 只讀原版，不寫任何東西。位址是 GRT.EXE 的 IDA 線性位址
// （`docs/playtest/43` §4 有靜態對照）。
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/wicanr2/dosgolem/oracle"
)

func main() {
	exe := flag.String("exe", "", "GRT.EXE 路徑")
	root := flag.String("root", ".", "原版素材目錄")
	steps := flag.Uint64("steps", 20_000_000, "再跑幾道指令")
	typeText := flag.String("type", "", "預填 DOS stdin")
	adlib := flag.Bool("adlib", false, "打開 OPL2 存在偵測")
	flag.Parse()
	if *exe == "" {
		fmt.Fprintln(os.Stderr, "要給 -exe")
		os.Exit(2)
	}
	o, err := oracle.LoadWith(*exe, *root, oracle.Options{AdLib: *adlib})
	if err != nil {
		fmt.Fprintln(os.Stderr, "載入失敗：", err)
		os.Exit(1)
	}
	if *typeText != "" {
		o.Type(*typeText)
	}

	pollCallers := map[uint32]int{}
	var pollHits uint64
	o.OnCall(o.IDA(0x11925), func(o *oracle.Oracle) {
		pollHits++
		c := o.Caller()
		pollCallers[o.ToIDA(c)]++
	})

	gateHits := []string{}
	for _, entry := range []uint32{0x144A6, 0x1346C, 0x1382C} {
		e := entry
		o.OnCall(o.IDA(e), func(o *oracle.Oracle) {
			gateHits = append(gateHits, fmt.Sprintf("gate=%05X caller=%05X", e, o.ToIDA(o.Caller())))
		})
	}
	vecHits := []string{}
	for _, entry := range []uint32{0x10394, 0x103A4} {
		e := entry
		o.OnCall(o.IDA(e), func(o *oracle.Oracle) {
			vecHits = append(vecHits, fmt.Sprintf("entry=%05X caller=%05X arg0=%04X ax=%04X",
				e, o.ToIDA(o.Caller()), o.Arg(0), o.AX()))
		})
	}

	if err := o.RunUntil(oracle.Steps(*steps)); err != nil {
		fmt.Println("停止：", err)
	}
	fmt.Printf("sub_11925 命中 %d 次，停在 %s\n", pollHits, o.IP())
	addrs := make([]uint32, 0, len(pollCallers))
	for a := range pollCallers {
		addrs = append(addrs, a)
	}
	sort.Slice(addrs, func(i, j int) bool { return addrs[i] < addrs[j] })
	fmt.Println("輪詢呼叫端（IDA 線性：次數）：")
	for _, a := range addrs {
		fmt.Printf("  %05X: %d\n", a, pollCallers[a])
	}
	fmt.Println("向量存取（entry／caller／arg0／AX）：")
	for _, h := range vecHits {
		fmt.Println("  " + h)
	}
	fmt.Println("音效閘門（gate／caller）：")
	for _, h := range gateHits {
		fmt.Println("  " + h)
	}
}
