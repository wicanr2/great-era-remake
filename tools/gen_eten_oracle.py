"""用 tools/eten.py 產生不含字模內容的 Big5→倚天索引 oracle。"""

from pathlib import Path

import eten

OUT = Path("internal/assets/testdata/eten-index-oracle.tsv")


def main() -> int:
    rows = ["# big5\tfont\tindex"]
    for hi, lo in eten.iter_big5():
        hit = eten.big5_to_font(hi, lo)
        if hit is None:
            rows.append(f"{hi:02X}{lo:02X}\t-\t-")
        else:
            rows.append(f"{hi:02X}{lo:02X}\t{hit[0]}\t{hit[1]}")
    OUT.parent.mkdir(parents=True, exist_ok=True)
    OUT.write_text("\n".join(rows) + "\n", "ascii")
    print(f"寫入 {OUT}：{len(rows)-1} 個合法 Big5 雙位元組")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
