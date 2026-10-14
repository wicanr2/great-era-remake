// ui-audit records source dimensions and real save-state changes for the
// isolated command-room playtest. It never writes original game data.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/game"
	"os"
	"path/filepath"
)

func main() {
	dir := flag.String("game", "", "read-only game directory")
	saved := flag.String("saved", "", "saved DT1")
	flag.Parse()
	result := map[string]any{}
	b, e := os.ReadFile(filepath.Join(*dir, "NEWSDATA.DAT"))
	must(e)
	images, e := assets.DecodeBGISet(b)
	must(e)
	sizes := [][2]int{}
	for _, im := range images {
		sizes = append(sizes, [2]int{im.W, im.H})
	}
	result["news_dimensions"] = sizes
	b, e = os.ReadFile(filepath.Join(*dir, "SAVE(1).DT1"))
	must(e)
	tbl, e := game.ParseSaveProvinces(b)
	must(e)
	p, e := tbl.At(26)
	must(e)
	result["original_province_26"] = p
	if *saved != "" {
		b, e = os.ReadFile(*saved)
		must(e)
		tbl, e = game.ParseSaveProvinces(b)
		must(e)
		p, e = tbl.At(26)
		must(e)
		result["saved_province_26"] = p
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	must(enc.Encode(result))
}
func must(e error) {
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
