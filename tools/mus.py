#!/usr/bin/env python3
"""《大時代的故事》`.MUS` / `.TIM` 音樂格式解碼器。

`.TIM` = 每首曲子自帶的 AdLib(OPL2)音色庫，欄位順序與 Ad Lib `.BNK` 的
instrument record 相同，但每個欄位由 UINT8 加寬成 UINT16LE。
`.MUS` = MIDI 風格的事件串（delta byte + status + data），由常駐於
INT 66h 的外部驅動程式（`SDFA.EXE`）解譯。

用法：
    tools/py.sh tools/mus.py tim        # 傾印 8 個 .TIM 的音色庫
    tools/py.sh tools/mus.py header     # 傾印 8 個 .MUS 的檔頭
    tools/py.sh tools/mus.py events     # 解析事件串並統計
    tools/py.sh tools/mus.py midi       # 匯出 Standard MIDI File 到 workplace/audio/midi/
    tools/py.sh tools/mus.py notes NAME # 列出單曲的音符（含音高名）
    tools/py.sh tools/mus.py style      # 編制、調式吻合度、音級分佈

輸出一律寫到 workplace/audio/（已 gitignore）；原版素材唯讀。
"""

from __future__ import annotations

import struct
import sys
from collections import Counter
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
AUDIO = ROOT / "workplace" / "audio"

SONGS = ["MAINTHEM", "BATTLE1", "BATTLE2", "SCENE", "FINAL", "BT02", "WALL", "STRATEGY"]

# ---------------------------------------------------------------------------
# .TIM
# ---------------------------------------------------------------------------

NAME_LEN = 9
REC_WORDS = 28
REC_LEN = REC_WORDS * 2

# Ad Lib .BNK instrument record 的欄位順序（modulator 13 + carrier 13 + 2 波形）。
# carrier 的 feedback / connection 在 OPL 上不存在（那是 channel 級參數），
# 檔案裡那兩格是未初始化的殘值。
OP_FIELDS = ["ksl", "multi", "feedback", "attack", "sustain", "eg", "decay",
             "release", "level", "am", "vib", "ksr", "connection"]
FIELD_NAMES = [f"mod.{f}" for f in OP_FIELDS] + [f"car.{f}" for f in OP_FIELDS] + \
              ["mod.wave", "car.wave"]
# OPL2 各欄位的合法上限，用來驗證欄位順序推論是否成立。
FIELD_MAX = {"ksl": 3, "multi": 15, "feedback": 7, "attack": 15, "sustain": 15,
             "eg": 1, "decay": 15, "release": 15, "level": 63, "am": 1, "vib": 1,
             "ksr": 1, "connection": 1, "wave": 3}
# 這兩格 OPL 用不到，檔案裡是殘值，不參與驗證。
UNUSED_IDX = {OP_FIELDS.index("feedback") + 13, OP_FIELDS.index("connection") + 13}


def parse_tim(data: bytes) -> dict:
    version, count, param_off = struct.unpack_from("<HHH", data, 0)
    names = []
    for i in range(count):
        raw = data[6 + i * NAME_LEN: 6 + (i + 1) * NAME_LEN]
        names.append(raw.split(b"\0")[0].decode("ascii", "replace"))
    insts = []
    for i in range(count):
        off = param_off + i * REC_LEN
        insts.append(list(struct.unpack_from("<%dH" % REC_WORDS, data, off)))
    return {
        "version": version, "count": count, "param_off": param_off,
        "names": names, "params": insts,
        "expected_off": 6 + count * NAME_LEN,
        "expected_size": 6 + count * NAME_LEN + count * REC_LEN,
        "actual_size": len(data),
    }


def field_violations(params: list[int]) -> list[str]:
    bad = []
    for idx, val in enumerate(params):
        if idx in UNUSED_IDX:
            continue
        name = FIELD_NAMES[idx]
        key = name.split(".")[1]
        if key.startswith("wave"):
            key = "wave"
        if val > FIELD_MAX[key]:
            bad.append(f"{name}={val}>{FIELD_MAX[key]}")
    return bad


def cmd_tim() -> None:
    bank: dict[str, tuple[int, ...]] = {}
    conflicts = []
    total_viol = 0
    for song in SONGS:
        data = (AUDIO / f"{song}.TIM").read_bytes()
        t = parse_tim(data)
        ok_off = "OK" if t["param_off"] == t["expected_off"] else "MISMATCH"
        ok_sz = "OK" if t["expected_size"] == t["actual_size"] else "MISMATCH"
        print(f"\n=== {song}.TIM  ver={t['version']}  count={t['count']}  "
              f"param_off=0x{t['param_off']:X}({ok_off})  size={t['actual_size']}({ok_sz})")
        for name, p in zip(t["names"], t["params"]):
            viol = field_violations(p)
            total_viol += len(viol)
            key = tuple(v for i, v in enumerate(p) if i not in UNUSED_IDX)
            if name in bank and bank[name] != key:
                conflicts.append((song, name))
            bank.setdefault(name, key)
            flag = "  !! " + ",".join(viol) if viol else ""
            print(f"  {name:<10} " + " ".join(f"{v:>3}" if i not in UNUSED_IDX else "  ."
                                              for i, v in enumerate(p)) + flag)
    print(f"\n欄位值域違規總數：{total_viol}（0 = 欄位順序推論與 OPL2 值域一致）")
    print(f"不重複音色名：{len(bank)}；跨檔同名但參數不同：{len(conflicts)} {conflicts}")
    print("音色名清單：", ", ".join(sorted(bank)))


# ---------------------------------------------------------------------------
# .MUS
# ---------------------------------------------------------------------------

HDR_LEN = 0x46

# status 高 nibble → data byte 數（AdLib MUS 規格）。
# 8x/9x = NoteOff/NoteOn（2）；Ax = 音量變更（1，不是標準 MIDI 的 poly aftertouch）；
# Bx = 控制器（1）；Cx = 音色切換（1）；Dx = 通道壓力（1）；Ex = 弯音（2）。
DATA_LEN = {0x8: 2, 0x9: 2, 0xA: 1, 0xB: 1, 0xC: 1, 0xD: 1, 0xE: 2}

NOTE_NAMES = ["C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"]


def note_name(n: int) -> str:
    return f"{NOTE_NAMES[n % 12]}{n // 12 - 1}"


def parse_mus(data: bytes) -> dict:
    """欄位名稱依 AdLib MUS 格式規格（VGMPF）。"""
    hdr = {
        "major": data[0], "minor": data[1],
        "tuneId": struct.unpack_from("<i", data, 2)[0],
        "tuneName": data[6:0x24].split(b"\0")[0].decode("ascii", "replace"),
        "tickBeat": data[0x24],          # 每拍 tick 數
        "beatMeasure": data[0x25],       # 每小節拍數
        "totalTick": struct.unpack_from("<i", data, 0x26)[0],
        "dataSize": struct.unpack_from("<i", data, 0x2A)[0],
        "nrCommand": struct.unpack_from("<i", data, 0x2E)[0],
        "soundMode": data[0x3A],         # 0=melodic(9 聲部) 1=percussive(6+5)
        "pitchBRange": data[0x3B],       # 弯音範圍（半音）
        "basicTempo": struct.unpack_from("<H", data, 0x3C)[0],
    }
    # tick 速率 = (tempo / 60) * tickBeat（Hz），秒數 = totalTick / 該速率
    rate = hdr["basicTempo"] / 60.0 * hdr["tickBeat"]
    hdr["tickRate"] = rate
    hdr["seconds"] = hdr["totalTick"] / rate if rate else 0.0
    # 相容舊欄位名
    hdr.update({"w26": hdr["totalTick"], "w2A": hdr["dataSize"],
                "w2E": hdr["nrCommand"], "b3A": hdr["soundMode"],
                "b3B": hdr["pitchBRange"], "b3C": hdr["basicTempo"]})
    events = []          # (abs_tick, status, data bytes)
    pos = HDR_LEN
    tick = 0
    running = None
    trailing = None
    while pos < len(data):
        # 0xF8 是 real-time byte：**沒有 delta 前綴、沒有 data、不推進 tick**。
        # 依據：修掉這條之後 8 首曲子的事件數與總 tick 全部與檔頭 @0x2E／@0x26 相符。
        if data[pos] == 0xF8:
            events.append((tick, 0xF8, b""))
            pos += 1
            continue
        delta = data[pos]
        pos += 1
        if pos >= len(data):
            trailing = ("delta at EOF", pos - 1)
            break
        tick += delta
        b = data[pos]
        if b & 0x80:
            status = b
            pos += 1
        else:
            if running is None:
                trailing = ("data byte without running status", pos)
                break
            status = running
        hi = status >> 4
        if status == 0xF0:                      # SysEx：讀到 F7
            end = data.find(0xF7, pos)
            if end < 0:
                trailing = ("unterminated sysex", pos)
                break
            events.append((tick, status, data[pos:end]))
            pos = end + 1
            running = None
        elif status >= 0xF1:                    # F1..FF：無 data
            events.append((tick, status, b""))
            running = None
            if status == 0xFC:                  # 停止
                break
        elif hi in DATA_LEN:
            n = DATA_LEN[hi]
            events.append((tick, status, data[pos:pos + n]))
            pos += n
            running = status
        else:
            trailing = (f"unknown status 0x{status:02X}", pos)
            break
    return {"hdr": hdr, "events": events, "end_pos": pos, "size": len(data),
            "trailing": trailing, "ticks": tick}


def cmd_header() -> None:
    print(f"{'曲名':<10}{'ver':>5}{'tick/拍':>8}{'拍/小節':>8}{'totalTick':>10}"
          f"{'dataSize':>9}{'nrCmd':>7}{'模式':>5}{'弯音':>5}{'tempo':>6}"
          f"{'tick/s':>8}{'秒':>7}  其餘位元組")
    for song in SONGS:
        data = (AUDIO / f"{song}.MUS").read_bytes()
        m = parse_mus(data)["hdr"]
        head = bytearray(data[:HDR_LEN])
        for o in list(range(0, 2)) + list(range(0x24, 0x32)) + [0x3A, 0x3B, 0x3C, 0x3D]:
            head[o] = 0
        rest = "全零" if not any(head) else "有殘值:" + head.hex()
        ok = "OK" if m["dataSize"] == len(data) - HDR_LEN else "!!"
        print(f"{song:<10}{m['major']}.{m['minor']:<3}{m['tickBeat']:>8}"
              f"{m['beatMeasure']:>8}{m['totalTick']:>10}{m['dataSize']:>9}({ok})"
              f"{m['nrCommand']:>7}{m['soundMode']:>5}{m['pitchBRange']:>5}"
              f"{m['basicTempo']:>6}{m['tickRate']:>8.0f}{m['seconds']:>7.1f}  {rest}")


def cmd_events() -> None:
    for song in SONGS:
        data = (AUDIO / f"{song}.MUS").read_bytes()
        m = parse_mus(data)
        ev = m["events"]
        stat = Counter(e[1] & 0xF0 if e[1] < 0xF0 else e[1] for e in ev)
        chans = sorted({e[1] & 0x0F for e in ev if e[1] < 0xF0})
        notes = [e for e in ev if (e[1] >> 4) == 9 and e[2][1] != 0]
        pitches = [e[2][0] for e in notes]
        progs = sorted({(e[1] & 0xF, e[2][0]) for e in ev if (e[1] >> 4) == 0xC})
        tail = m["size"] - m["end_pos"]
        h = m["hdr"]
        chk_n = "OK" if len(ev) == h["w2E"] else f"檔頭說 {h['w2E']}"
        chk_t = "OK" if m["ticks"] == h["w26"] else f"檔頭說 {h['w26']}"
        print(f"\n=== {song}.MUS  {m['size']} B  事件 {len(ev)}({chk_n})  "
              f"總 tick {m['ticks']}({chk_t})  "
              f"解析到 0x{m['end_pos']:X}（剩 {tail} B）  異常={m['trailing']}")
        print("  status 分佈：", " ".join(f"{s:02X}:{c}" for s, c in sorted(stat.items())))
        print(f"  使用聲部：{chans}")
        print(f"  NoteOn 數 {len(notes)}；音高 {min(pitches)}({note_name(min(pitches))})"
              f"–{max(pitches)}({note_name(max(pitches))})")
        print(f"  ProgramChange (ch,prog)：{progs}")
        exx = [e for e in ev if (e[1] >> 4) == 0xE]
        vals = Counter((e[2][0], e[2][1]) for e in exx)
        print(f"  Ex 事件 {len(exx)} 種類 {len(vals)}：{list(vals.items())[:8]}")
        axx = [e for e in ev if (e[1] >> 4) == 0xA]
        print(f"  Ax 事件 {len(axx)}，值域 "
              f"{min((e[2][0] for e in axx), default=-1)}–{max((e[2][0] for e in axx), default=-1)}")
        syx = [e for e in ev if e[1] == 0xF0]
        # F0 7F 00 <整數> <分數/128> F7 = 速度倍率（0x01,0x00 = 1.00 = 原速）
        spd = [(t, p[2] + p[3] / 128.0) for t, _, p in syx if len(p) == 4]
        print(f"  速度控制 SysEx {len(syx)} 次，倍率 {sorted({round(v, 3) for _, v in spd})}")


def _vlq(n: int) -> bytes:
    out = bytes([n & 0x7F])
    n >>= 7
    while n:
        out = bytes([(n & 0x7F) | 0x80]) + out
        n >>= 7
    return out


def to_midi(song: str) -> bytes:
    """把 .MUS 轉成 Standard MIDI File format 0，供試聽與轉譜。

    remake 差異（**不是原版語意**）：Ax 自訂音量事件轉成 MIDI CC7；
    Ex 保留為 pitch bend；音色沿用 .TIM 的名字對照不到 GM，故不寫 ProgramChange。
    """
    data = (AUDIO / f"{song}.MUS").read_bytes()
    m = parse_mus(data)
    trk = bytearray()
    prev = 0
    for tick, status, payload in m["events"]:
        if status >= 0xF0:
            continue
        hi, ch = status >> 4, status & 0xF
        if hi == 0xA:                            # 自訂音量 → CC7
            out = bytes([0xB0 | ch, 7, min(payload[0], 127)])
        elif hi in (0x8, 0x9, 0xE):
            out = bytes([status]) + bytes(payload)
        elif hi in (0xB, 0xC, 0xD):
            out = bytes([status]) + bytes(payload)
        else:
            continue
        trk += _vlq(tick - prev) + out
        prev = tick
    trk += _vlq(0) + b"\xff\x2f\x00"
    hdr = b"MThd" + struct.pack(">IHHH", 6, 0, 1, 96)
    return hdr + b"MTrk" + struct.pack(">I", len(trk)) + bytes(trk)


def cmd_midi() -> None:
    out = AUDIO / "midi"
    out.mkdir(exist_ok=True)
    for song in SONGS:
        p = out / f"{song}.mid"
        p.write_bytes(to_midi(song))
        print(f"{p} {p.stat().st_size} B")
    print("（MIDI 是分析用中介檔，放在 gitignore 的 workplace/ 底下，不進版控）")


def cmd_notes(song: str, limit: int = 120) -> None:
    data = (AUDIO / f"{song}.MUS").read_bytes()
    m = parse_mus(data)
    pcs: Counter[int] = Counter()
    shown = 0
    for tick, status, payload in m["events"]:
        if status >= 0xF0 or (status >> 4) != 9:
            continue
        n, v = payload
        if v == 0:
            continue
        pcs[n % 12] += 1
        if shown < limit:
            print(f"  t={tick:<6} ch{status & 0xF:<3} {note_name(n):<5} vel={v}")
            shown += 1
    print(f"\n{song} 音級分佈（0=C）：")
    for pc, c in sorted(pcs.items()):
        print(f"  {NOTE_NAMES[pc]:<3}{c:>5}")


MAJOR = [0, 2, 4, 5, 7, 9, 11]
GONG = [0, 2, 4, 7, 9]          # 宮五聲（= 西洋大調五聲）


def _fit(pcs: Counter[int], scale: list[int]) -> tuple[float, int]:
    tot = sum(pcs.values()) or 1
    return max(((sum(c for p, c in pcs.items() if p in {(r + i) % 12 for i in scale}) / tot, r)
                for r in range(12)))


def cmd_style() -> None:
    """調式與編制分析——`docs/design/30-audio-direction.md` §1 那幾張表的來源。"""
    print(f"{'曲名':<10}{'秒':>7}{'BPM':>5}{'模式':>5}{'聲部':>5}{'音色':>5}  編制")
    for song in SONGS:
        m = parse_mus((AUDIO / f"{song}.MUS").read_bytes())
        t = parse_tim((AUDIO / f"{song}.TIM").read_bytes())
        ch = sorted({s & 0xF for _, s, _ in m["events"] if s < 0xF0})
        h = m["hdr"]
        print(f"{song:<10}{h['seconds']:>7.1f}{h['basicTempo']:>5}"
              f"{'打擊' if h['soundMode'] else '旋律':>5}{len(ch):>5}{t['count']:>5}  "
              f"{','.join(t['names'])}")

    print("\n音級分佈與調式吻合度（僅旋律聲部 0–5）")
    for song in SONGS:
        m = parse_mus((AUDIO / f"{song}.MUS").read_bytes())
        pcs = Counter(p[0] % 12 for _, s, p in m["events"]
                      if s < 0xF0 and (s >> 4) == 9 and p[1] != 0 and (s & 0xF) < 6)
        tot = sum(pcs.values()) or 1
        fm, fp = _fit(pcs, MAJOR), _fit(pcs, GONG)
        print(f"{song:<10} n={tot:<5} 大調七聲 {fm[0]:.2f}@{NOTE_NAMES[fm[1]]:<3}"
              f" 五聲 {fp[0]:.2f}@{NOTE_NAMES[fp[1]]:<3} | "
              + " ".join(f"{NOTE_NAMES[i]}:{pcs.get(i, 0) * 100 // tot:>2}%" for i in range(12)))


def main(argv: list[str]) -> int:
    cmd = argv[1] if len(argv) > 1 else "tim"
    if cmd == "tim":
        cmd_tim()
    elif cmd == "header":
        cmd_header()
    elif cmd == "events":
        cmd_events()
    elif cmd == "midi":
        cmd_midi()
    elif cmd == "style":
        cmd_style()
    elif cmd == "notes":
        cmd_notes(argv[2] if len(argv) > 2 else "MAINTHEM")
    else:
        print(__doc__)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv))
