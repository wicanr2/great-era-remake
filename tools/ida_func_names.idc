// 列出函式名含鍵盤相關子串的函式（找 ReadKey 類例程用）。
//   tools/ida.sh query tools/ida_func_names.idc GRT.EXE.i64
// 輸出到 /work/func-names.txt（→ workplace/ida/user-output/）。
#include <idc.idc>

static main() {
  auto i, qty, ea, name, f, n;
  Wait();
  f = fopen("/work/func-names.txt", "w");
  if (f == 0) {
    Message("開檔失敗\n");
    qexit(2);
  }
  n = 0;
  qty = get_func_qty();
  for (i = 0; i < qty; i++) {
    ea = getn_func(i);
    name = get_func_name(ea);
    if (name != "") {
      fprintf(f, "%Xh %s\n", ea, name);
      n++;
    }
  }
  fclose(f);
  Message("匯出 %d 個函式名\n", n);
  qexit(0);
}
