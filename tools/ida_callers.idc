// 列一個函式符號的所有程式碼交叉參考（呼叫端）：誰 call 它、在哪裡。
//
//   tools/ida.sh query tools/ida_callers.idc GRT.EXE.i64 sub_11A25 sub_10D83
//   cat workplace/ida/user-output/callers-sub_11A25.txt
//
// 跟 tools/ida_xref.idc 互補：那支只看資料參考（dref），這支只看
// 程式碼參考（cref/RfirstB），用來找「鍵盤讀取是被哪個迴圈呼叫的」。
// 輸出到 /work/callers-<符號>.txt（query 模式下會被收進 user-output/）。
//
// 注意：add di, 常數這類算術不會建立 xref，直接 xref 也抓不到取址後的
// 間接呼叫；零命中時先懷疑工具邊界，再追取址端。

#include <idc.idc>

static dump(name) {
    auto ea, x, out, fn, base;
    ea = get_name_ea_simple(name);
    if (ea == BADADDR) {
        Message("[ida_callers] 找不到符號 %s\n", name);
        return 0;
    }
    out = fopen("/work/callers-" + name + ".txt", "w");
    fprintf(out, "# %s @ %x\n", name, ea);
    fprintf(out, "# 由 tools/ida_callers.idc 產生（IDA 程式碼交叉參考，不是 grep .asm）\n\n");
    fprintf(out, "## 呼叫端\n\n");
    for (x = RfirstB(ea); x != BADADDR; x = RnextB(ea, x)) {
        fn = get_func_name(x);
        if (fn == "") fn = "(不在函式內)";
        base = get_func_attr(x, FUNCATTR_START);
        fprintf(out, "%-14s +0x%x  %x  %s\n", fn, x - base, x, GetDisasm(x));
    }
    fprintf(out, "\n直接程式碼參考結束。間接呼叫抓不到，見檔頭注意。\n");
    fclose(out);
    Message("[ida_callers] %s 完成\n", name);
    return 1;
}

static main() {
    auto i;
    Wait();
    if (ARGV.count < 2) {
        Message("[ida_callers] 用法: -S\"ida_callers.idc <符號名> [符號名...]\"\n");
        qexit(1);
    }
    for (i = 1; i < ARGV.count; i++) dump(ARGV[i]);
    qexit(0);
}
