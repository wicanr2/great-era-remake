// 對一段位址範圍**逐 byte** 問 IDA 的 xref 圖：誰參考了這塊記憶體裡的任何一個位置。
//
//   tools/ida.sh raw idat -A "-S/work/tools/ida_range_xref.idc 6EFAA 6F532" WAR.EXE.i64
//   cat workplace/ida/rangexref-6EFAA-6F532.txt
//
// 為什麼要這支（2026-08-02）：
//
// `ida_xref.idc` 只能查**一個具名符號**。但一塊結構化的記憶體通常只有起點
// 有名字，中間的 `base + i*stride + col` 會被 IDA 標成對「中間某個位址」的參考——
// 那些位址沒有符號名，於是查基址什麼都查不到，看起來像「沒人用這塊資料」。
//
// ⛔ **不要改用 grep `.asm` 找裸位址。** 16-bit 反組譯顯示的是 `segment:offset`，
//    IDA 的線性位址（`6EFAA`）**根本不會出現在文字裡**。實測：整份 WAR.EXE.asm
//    連一個 5 位十六進位常數都沒有。用它下「沒有參考」的結論是假陰性。
//
// 這支比「掃全部指令的立即數」快兩個數量級：1416 次查詢 vs 二十萬條指令。

#include <idc.idc>

static main() {
    auto lo, hi, ea, x, out, n, nref, k, t;
    Wait();
    if (ARGV.count < 3) { qexit(1); }
    lo = xtol(ARGV[1]); hi = xtol(ARGV[2]);
    out = fopen("/work/rangexref-" + ARGV[1] + "-" + ARGV[2] + ".txt", "w");
    fprintf(out, "# %s..%s 逐 byte 的交叉參考\n", ARGV[1], ARGV[2]);
    fprintf(out, "# 由 tools/ida_range_xref.idc 產生\n\n");
    n = 0; nref = 0;
    for (ea = lo; ea < hi; ea = ea + 1) {
        k = 0;
        for (x = get_first_dref_to(ea); x != BADADDR; x = get_next_dref_to(ea, x)) {
            t = XrefType();
            fprintf(out, "%-6s +%-5d  %-16s %-6X  %s\n",
                    t == 1 ? "取址" : (t == 2 ? "寫" : "讀"),
                    ea - lo, get_func_name(x), x, GetDisasm(x));
            k = k + 1; nref = nref + 1;
        }
        if (k > 0) n = n + 1;
    }
    fprintf(out, "\n範圍內有 %d 個位址被參考，合計 %d 筆。\n", n, nref);
    fclose(out);
    qexit(0);
}
