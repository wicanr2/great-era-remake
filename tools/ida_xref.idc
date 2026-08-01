// 查一個全域變數的所有交叉參考：誰讀它、誰寫它、寫的是什麼。
//
//   tools/ida.sh raw idat -A "-S/work/tools/ida_xref.idc word_64944" WAR.EXE.i64
//   cat workplace/ida/xref-word_64944.txt
//
// 為什麼要這個（2026-08-01）：
//
// 這個專案的 17 個分析腳本全部在 grep `.asm` 文字檔，而 `.asm` 是攤平的——
// 它沒有交叉參考圖。想知道「word_64944 到底是什麼」，只能從呼叫端的參數
// 順序反推，那是**間接證據**，而 docs/re/31 §16 就是這樣把它推錯的
// （推成「第二方的代表單位」，§35 出現反證後退回未解）。
//
// IDA 的資料庫裡本來就有完整的 xref 圖。這支腳本把它取出來，
// 讓「誰寫過這個位址、寫了什麼值」變成一個可以直接問的問題。
//
// 輸出到 /work/xref-<符號>.txt（也就是 workplace/ida/，已 gitignore）。

#include <idc.idc>

// 往回找最近的立即數來源，用來猜「寫進去的值是什麼」。
// 只看同一個基本區塊的前幾行，找不到就回空字串——**不要硬猜**。
static guess_value(ea) {
    auto i, p, d;
    p = ea;
    for (i = 0; i < 6; i++) {
        p = prev_head(p, 0);
        if (p == BADADDR) break;
        d = GetDisasm(p);
        // mov <reg>, <立即數> 這種形式
        if (strstr(d, "mov ") == 0 && strstr(d, "[") < 0) return d;
        // 位元操作也算「值的來源」
        if (strstr(d, "or ") == 0 || strstr(d, "and ") == 0 || strstr(d, "xor ") == 0)
            return d;
    }
    return "";
}

static dump(name) {
    auto ea, x, f, out, d, fn, kind, nr, nw, no, t, src;

    ea = get_name_ea_simple(name);
    if (ea == BADADDR) {
        Message("[ida_xref] 找不到符號 %s\n", name);
        return 0;
    }

    out = fopen("/work/xref-" + name + ".txt", "w");
    fprintf(out, "# %s @ %x\n", name, ea);
    fprintf(out, "# 由 tools/ida_xref.idc 產生（IDA 的交叉參考圖，不是 grep .asm）\n\n");

    // 這個位址本身的型別與大小，IDA 已經推導好了。
    fprintf(out, "型別/大小: %s (item size %d)\n\n", GetDisasm(ea), get_item_size(ea));

    nr = 0; nw = 0; no = 0;
    fprintf(out, "## 交叉參考\n\n");
    for (x = get_first_dref_to(ea); x != BADADDR; x = get_next_dref_to(ea, x)) {
        d  = GetDisasm(x);
        fn = get_func_name(x);
        if (fn == "") fn = "(不在函式內)";

        // 讀寫判定：**用 IDA 自己標的 xref 型別**（XrefType()）。
        //
        // 這裡連錯兩次，兩次都是自己重造 IDA 已經有的東西：
        //
        //   第一版  strstr(d, "mov " + name + ",")
        //           → 漏掉 "mov     word_64944, dx"，助憶碼後面是多個空格
        //   第二版  print_operand(x, 0) == name
        //           → 把 "push word_64942" 判成寫，因為 push 的第 0 個
        //             運算元是**來源**不是目的
        //
        // 指令語意不該由呼叫端猜。IDA 在建資料庫時就已經標好每一條
        // 交叉參考是讀還是寫（dr_R / dr_W / dr_O），直接問它。
        //
        //   dr_O = 1 取位址   dr_W = 2 寫   dr_R = 3 讀
        t = XrefType();
        if (t == 2)      { kind = "寫";   nw++; }
        else if (t == 1) { kind = "取址"; no++; }
        else             { kind = "讀";   nr++; }

        src = "";
        if (kind == "寫") src = guess_value(x);

        fprintf(out, "%-4s %-14s %8x  %s\n", kind, fn, x, d);
        if (src != "") fprintf(out, "                            ↑ 值可能來自: %s\n", src);
    }

    fprintf(out, "\n讀 %d 處、寫 %d 處、取位址 %d 處。\n", nr, nw, no);
    fprintf(out, "\n⚠️ 這是 IDA 標的直接參考。**透過指標的間接寫入抓不到**\n");
    fprintf(out, "   （ptr = &x 之後 es:[di] = v 那種）。看到寫入數為 0 時，\n");
    fprintf(out, "   先看「取位址」那幾筆——間接寫入一定是從那裡開始的。\n");
    fclose(out);
    Message("[ida_xref] %s: 讀 %d 寫 %d 取址 %d\n", name, nr, nw, no);
    return 1;
}

static main() {
    auto i;
    Wait();
    if (ARGV.count < 2) {
        Message("[ida_xref] 用法: -S\"ida_xref.idc <符號名> [符號名...]\"\n");
        qexit(1);
    }
    for (i = 1; i < ARGV.count; i++) dump(ARGV[i]);
    qexit(0);
}
