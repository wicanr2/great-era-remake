// 反組譯 IDA 資料庫裡的任意位址範圍（包含 .asm 收合成一行不印的函式，
// 例如 WAR.EXE 的 CRT READKEY/KEYPRESSED）。
//
//   tools/ida.sh query tools/ida_range.idc WAR.EXE.i64 62290-62500
//   cat workplace/ida/user-output/range-62290-62500.txt
//
// 輸出到 /work/range-<起>-<迄>.txt（query 模式下會被收進 user-output/）。
// 每一行：位址＋反組譯＋直接交叉參考註記。位址是 IDA 線性位址。

#include <idc.idc>

static dump_range(spec) {
    auto i, start, fin, ea, out, d, x, c, n;
    i = strstr(spec, "-");
    if (i < 0) {
        Message("[ida_range] 範圍格式錯誤 %s（要 起-迄 十六進位）\n", spec);
        return 0;
    }
    start = xtol(substr(spec, 0, i));
    fin = xtol(substr(spec, i + 1, strlen(spec)));
    out = fopen("/work/range-" + spec + ".txt", "w");
    fprintf(out, "# IDA 線性 %x-%x（tools/ida_range.idc，非 grep .asm）\n\n", start, fin);
    for (ea = start; ea < fin; ) {
        d = GetDisasm(ea);
        if (d == "") {
            fprintf(out, "%05x  db %02xh\n", ea, Byte(ea));
            ea = ea + 1;
        } else {
            fprintf(out, "%05x  %s", ea, d);
            // 直接程式碼參考：call 進來的有誰、跳去哪
            c = 0;
            for (x = RfirstB(ea); x != BADADDR; x = RnextB(ea, x)) {
                if (c == 0) fprintf(out, "    ; ← ");
                else fprintf(out, ", ");
                fprintf(out, "%s+%x", get_func_name(x), x - get_func_attr(x, FUNCATTR_START));
                c = c + 1;
                if (c > 3) { fprintf(out, ",…"); break; }
            }
            fprintf(out, "\n");
            ea = ea + ItemSize(ea);
        }
    }
    fclose(out);
    Message("[ida_range] %s 完成\n", spec);
    return 1;
}

static main() {
    auto i;
    Wait();
    if (ARGV.count < 2) {
        Message("[ida_range] 用法: -S\"ida_range.idc <起-迄> […]\"（十六進位）\n");
        qexit(1);
    }
    for (i = 1; i < ARGV.count; i++) dump_range(ARGV[i]);
    qexit(0);
}
