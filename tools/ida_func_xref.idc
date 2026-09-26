// 匯出 IDA 資料庫中指定函式的直接 code xref。
//
// 顯示契約：保留目標函式原名、線性位址、呼叫端原名與 callsite；
// 不改名、不寫註解、不把語意推測寫回資料庫。
//
// 使用（由 tools/ida.sh query 從 WAR.EXE 建立一次性 DB）：
//   tools/ida.sh query tools/ida_func_xref.idc WAR.EXE.i64 sub_35005
// 輸出：workplace/ida/user-output/funcxref-sub_35005.txt
#include <idc.idc>

static dump(requested) {
  auto target, x, out, fn, n;
  target = get_name_ea_simple(requested);
  if (target == BADADDR) {
    Message("找不到函式: %s\n", requested);
    return 0;
  }

  out = fopen("/work/funcxref-" + requested + ".txt", "w");
  fprintf(out, "# IDA Pro 函式 code xref 匯出\n");
  fprintf(out, "# 目標原始函式: %s\n", get_func_name(target));
  fprintf(out, "# 目標位址空間: IDA linear address\n");
  fprintf(out, "# 目標位址: %08Xh\n\n", target);
  fprintf(out, "## 直接 code xref to\n\n");
  n = 0;
  // IDA Pro 9.4 的 IDC API 已把 flags 參數移除；呼叫一參數版本，
  // 否則 headless 執行會在腳本行號處中止而不產生匯出檔。
  for (x = get_first_cref_to(target);
       x != BADADDR;
       x = get_next_cref_to(target, x)) {
    fn = get_func_name(x);
    if (fn == "") fn = "(不在函式內)";
    fprintf(out, "%-18s %-8Xh  %s\n", fn, x, GetDisasm(x));
    n = n + 1;
  }
  fprintf(out, "\n直接 code xref 合計 %d 筆。\n", n);
  fprintf(out, "\n⚠️ 這只涵蓋 IDA 建立的直接 code xref；間接函式指標呼叫不在此列。\n");
  fclose(out);
  Message("[ida_func_xref] %s: %d 筆\n", requested, n);
  return 1;
}

static main() {
  auto i;
  Wait();
  if (ARGV.count < 2) {
    Message("用法: -S\"ida_func_xref.idc <function-name> [function-name...]\"\n");
    qexit(1);
  }
  for (i = 1; i < ARGV.count; i++) dump(ARGV[i]);
  qexit(0);
}
