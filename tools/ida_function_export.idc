// 從 IDA 資料庫匯出單一函式的可回查證據。
//
// 使用：
//   tools/ida.sh raw idat -A \
//     "-S/work/tools/ida_function_export.idc sub_22E25" WAR.EXE.i64
//
// 輸出：workplace/ida/function-sub_22E25.txt
//
// 顯示契約：保留 IDA 原始函式名、線性位址、原始指令與運算元。
// 本工具不改名、不寫註解，也不把語意推測當成證據。
#include <idc.idc>

static main() {
  auto requested, start, finish, ea, f, line, actual;
  Wait();
  if (ARGV.count < 2) {
    Message("用法: -S\"ida_function_export.idc <function-name>\"\n");
    qexit(1);
  }

  requested = ARGV[1];
  start = get_name_ea_simple(requested);
  if (start == BADADDR) {
    Message("找不到函式: %s\n", requested);
    qexit(2);
  }
  finish = get_func_attr(start, FUNCATTR_END);
  if (finish == BADADDR) {
    Message("找不到函式邊界: %s\n", requested);
    qexit(3);
  }

  actual = get_func_name(start);
  f = fopen("/work/function-" + requested + ".txt", "w");
  fprintf(f, "# IDA Pro 單一函式匯出\n");
  fprintf(f, "# 資料庫: WAR.EXE.i64\n");
  fprintf(f, "# 位址空間: IDA linear address\n");
  fprintf(f, "# 原始函式: %s\n", actual);
  fprintf(f, "# 範圍: %Xh..%Xh\n\n", start, finish);

  // 16-bit DOS 載入器的函式邊界可包含非 head byte；逐 byte
  // 詢問 IDA，只輸出 GetDisasm 有內容的原始項目。函式最大只數 KB，
  // 這比依賴 next_head 在舊載入器上的邊界行為更穩定。
  for (ea = start; ea < finish; ea = ea + 1) {
    if (get_item_head(ea) != ea) continue;
    line = GetDisasm(ea);
    if (line != "") fprintf(f, "%08X  %s\n", ea, line);
  }
  fclose(f);
  qexit(0);
}
