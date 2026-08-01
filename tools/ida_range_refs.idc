// 掃全部指令，找「運算元的數值落在指定位址範圍內」的地方。
//
// 為什麼要這支：`tools/ida_xref.idc` 只看 IDA 的 xref 圖，而 xref 圖只涵蓋
// **IDA 認得出來的參考**。像 `mov ax, 6EFAAh`（把位址當純數字算）或
// 位址被拆成 `mov di, 6EF00h` + `add di, 0AAh` 這類，xref 圖是空的。
//
//   tools/ida.sh raw idat -A "-S/work/tools/ida_range_refs.idc 6EFAA 6F532" WAR.EXE.i64
//   → workplace/ida/rangerefs-6EFAA-6F532.txt
//
// ⚠️ headless 的 print/Message 不進 stdout，一律寫檔（CLAUDE.md §4.1）。
#include <idc.idc>

static main() {
  auto lo, hi, f, ea, seg, i, v, n, mn, fn;
  Wait();
  if (ARGV.count < 3) { Message("用法: -S\"ida_range_refs.idc <lo> <hi>\"\n"); qexit(1); }
  lo = xtol(ARGV[1]); hi = xtol(ARGV[2]);
  f = fopen("/work/rangerefs-" + ARGV[1] + "-" + ARGV[2] + ".txt", "w");
  fprintf(f, "# 位址範圍 %Xh..%Xh 的立即數／運算元參考\n", lo, hi);
  fprintf(f, "# 由 tools/ida_range_refs.idc 產生（掃全部指令，不是 xref 圖）\n\n");
  n = 0;
  for (seg = get_first_seg(); seg != BADADDR; seg = get_next_seg(seg)) {
    for (ea = seg; ea < get_segm_end(seg); ea = next_head(ea, get_segm_end(seg))) {
      if (!is_code(get_full_flags(ea))) continue;
      for (i = 0; i < 4; i++) {
        v = get_operand_value(ea, i);
        if (v == -1 || v < lo || v >= hi) continue;
        mn = print_insn_mnem(ea);
        if (mn == "") continue;
        fn = get_func_name(ea);
        fprintf(f, "%-18s %-6X  %s %s   → 偏移 +%d\n",
                fn, ea, mn, print_operand(ea, i), v - lo);
        n = n + 1;
        break;
      }
    }
  }
  fprintf(f, "\n合計 %d 處。\n", n);
  fprintf(f, "\n⚠️ 立即數與位址同值也會中（例如純數字 %d）。用「偏移」欄判斷是不是真的落在結構裡。\n", lo);
  fclose(f);
  Message("[range_refs] %Xh..%Xh: %d 處\n", lo, hi, n);
  qexit(0);
}
