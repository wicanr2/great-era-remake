#!/usr/bin/env python3
"""產生 `modern_campaign` 的決定性、原創、不可發行 MIDI 草稿。

只讀同目錄的 score.json 與標準函式庫；絕不讀 workplace/、原版 MUS/TIM 或分析 MIDI。
草稿用來確認「前奏 → 真空 → 豪情高潮」的弧線，不能取代人耳 QA、授權或正式混音。
"""

from __future__ import annotations

import argparse
import hashlib
import json
import struct
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
DEFAULT_SCORE = ROOT / "docs/music/modern_campaign/score.json"
DEFAULT_OUTPUT = ROOT / "docs/music/modern_campaign/modern_campaign_draft.mid"
BAR = 384


def vlq(value: int) -> bytes:
    if value < 0:
        raise ValueError("MIDI delta-time 不可為負")
    encoded = [value & 0x7F]
    value >>= 7
    while value:
        encoded.append(0x80 | (value & 0x7F))
        value >>= 7
    return bytes(reversed(encoded))


def chunk(kind: bytes, payload: bytes) -> bytes:
    return kind + struct.pack(">I", len(payload)) + payload


def meta(meta_type: int, payload: bytes) -> bytes:
    return bytes((0xFF, meta_type)) + vlq(len(payload)) + payload


def encode_events(events: list[tuple[int, bytes]], end_tick: int) -> bytes:
    ordered = sorted(
        events,
        key=lambda item: (
            item[0],
            0 if item[1][0] & 0xF0 == 0x80 else 1,
            item[1],
        ),
    )
    data = bytearray()
    last_tick = 0
    for tick, event in ordered:
        if tick < last_tick:
            raise ValueError("事件時間倒退")
        data.extend(vlq(tick - last_tick))
        data.extend(event)
        last_tick = tick
    data.extend(vlq(end_tick - last_tick))
    data.extend(meta(0x2F, b""))
    return chunk(b"MTrk", bytes(data))


def conductor_track(score: dict, end_tick: int) -> bytes:
    bpm = score["tempo_bpm"]
    numerator, denominator = score["time_signature"]
    tempo_us = round(60_000_000 / bpm)
    events = [
        (0, meta(0x03, b"modern_campaign conductor")),
        (0, meta(0x01, b"DRAFT - NOT FOR RELEASE")),
        (0, meta(0x06, b"intro 1-4; gather 5-8; advance 9-12; void 13-14; climax 15-20; return 21-24")),
        (0, meta(0x51, tempo_us.to_bytes(3, "big"))),
        (0, meta(0x58, bytes((numerator, {1: 0, 2: 1, 4: 2, 8: 3, 16: 4}[denominator], 24, 8)))),
    ]
    return encode_events(events, end_tick)


def add_note(notes: list[tuple[int, int, int, int]], start: int, pitch: int, duration: int, velocity: int, end_tick: int) -> None:
    if pitch < 0 or pitch > 127 or duration <= 0 or velocity < 1 or velocity > 127:
        raise ValueError("音符值域錯誤")
    if start < 0 or start + duration > end_tick:
        raise ValueError(f"音符超出 loop：{start}+{duration} > {end_tick}")
    notes.append((start, pitch, duration, velocity))


def motif(notes: list[tuple[int, int, int, int]], start: int, shift: int, velocity: int, end_tick: int) -> None:
    pitches = [62, 67, 65, 70, 69, 74]
    durations = [48, 48, 96, 48, 48, 96]
    tick = start
    for pitch, duration in zip(pitches, durations):
        add_note(notes, tick, pitch + shift, duration, velocity, end_tick)
        tick += duration


def chord(notes: list[tuple[int, int, int, int]], start: int, root: int, duration: int, velocity: int, end_tick: int) -> None:
    for pitch in (root, root + 3, root + 7):
        add_note(notes, start, pitch, duration, velocity, end_tick)


def build_tracks(score: dict) -> list[dict]:
    bars = score["bars"]
    end_tick = bars * BAR
    roots = [38, 43, 36, 40, 38, 43, 36, 40, 38, 43, 36, 40, 38, 38,
             38, 43, 36, 40, 38, 43, 36, 40, 38, 40]
    piano: list[tuple[int, int, int, int]] = []
    strings: list[tuple[int, int, int, int]] = []
    brass: list[tuple[int, int, int, int]] = []
    clarinet: list[tuple[int, int, int, int]] = []
    percussion: list[tuple[int, int, int, int]] = []

    for bar, root in enumerate(roots):
        start = bar * BAR
        if bar < 4:
            chord(piano, start, root + 12, BAR * 2, 38, end_tick)
        elif bar in (12, 13):
            add_note(piano, start, root, BAR, 34, end_tick)
        else:
            arp = [root + 12, root + 15, root + 19, root + 22]
            for i in range(8):
                add_note(piano, start + i * 48, arp[i % len(arp)], 42, 54 + (i % 2) * 8, end_tick)

        string_duration = min(BAR * (2 if bar in (12, 13) else 4), end_tick - start)
        add_note(strings, start, root, string_duration, 46 if bar < 14 else 62, end_tick)
        add_note(strings, start, root + 7, string_duration, 40 if bar < 14 else 56, end_tick)

    # 以兩小節一次的銅管呈示建立豪情；真空段完全留白。
    for bar, shift, velocity in ((4, 0, 58), (8, 5, 66), (14, 0, 84), (16, 5, 88),
                                 (18, 0, 92), (20, 5, 82), (22, 0, 74)):
        motif(brass, bar * BAR, shift, velocity, end_tick)
        if bar >= 14:
            motif(brass, bar * BAR, shift - 12, velocity - 12, end_tick)

    # 單簧管在前奏與回望回答，不和高潮銅管搶中頻。
    for bar, shift, velocity in ((0, -12, 48), (2, -7, 44), (6, -12, 50), (10, -7, 54),
                                 (20, -12, 54), (22, -7, 48)):
        motif(clarinet, bar * BAR, shift, velocity, end_tick)

    # 9–12 節制推進；13–14 真空只留下第 14 小節末的一次低鼓心跳；15–24 完整脈衝。
    for bar in range(6, 12):
        start = bar * BAR
        for beat in (1, 3):
            add_note(percussion, start + beat * 96, 38, 24, 46 + bar, end_tick)
    add_note(percussion, 13 * BAR + 288, 45, 48, 38, end_tick)
    for bar in range(14, 24):
        start = bar * BAR
        if bar in (14, 20):
            add_note(percussion, start, 49, 36, 88, end_tick)
        for offset in (96, 192, 288):
            add_note(percussion, start + offset, 38, 24, 62 + (bar % 3) * 5, end_tick)
        if bar % 2 == 1:
            add_note(percussion, start + 288, 45, 36, 56, end_tick)

    return [
        {"name": "Piano command pulse — original, draft only", "channel": 0, "program": 0, "notes": piano},
        {"name": "Low strings resolve and cost — original, draft only", "channel": 1, "program": 48, "notes": strings},
        {"name": "Brass campaign motif — original, draft only", "channel": 2, "program": 60, "notes": brass},
        {"name": "Clarinet distance answer — original, draft only", "channel": 3, "program": 71, "notes": clarinet},
        {"name": "Pulse stem: snare, low drum, restrained — original, draft only", "channel": 9, "program": 0, "notes": percussion},
    ]


def note_track(track: dict, end_tick: int) -> bytes:
    channel = track["channel"]
    events: list[tuple[int, bytes]] = [
        (0, meta(0x03, track["name"].encode("utf-8"))),
        (0, bytes((0xC0 | channel, track["program"]))),
    ]
    for start, pitch, duration, velocity in track["notes"]:
        events.append((start, bytes((0x90 | channel, pitch, velocity))))
        events.append((start + duration, bytes((0x80 | channel, pitch, 0))))
    return encode_events(events, end_tick)


def render(score: dict) -> bytes:
    if score["schema"] != 1 or score["cue"] != "modern_campaign":
        raise ValueError("只接受 schema 1 的 modern_campaign score")
    if score["status"] != "draft-not-for-release":
        raise ValueError("草稿狀態必須保持 draft-not-for-release")
    if score["ppqn"] != 96 or score["bars"] != 24:
        raise ValueError("本草稿固定使用 96 PPQN／24 小節")
    end_tick = score["bars"] * BAR
    tracks = build_tracks(score)
    header = chunk(b"MThd", struct.pack(">HHH", 1, len(tracks) + 1, score["ppqn"]))
    return header + conductor_track(score, end_tick) + b"".join(note_track(t, end_tick) for t in tracks)


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--score", type=Path, default=DEFAULT_SCORE)
    parser.add_argument("--output", type=Path, default=DEFAULT_OUTPUT)
    parser.add_argument("--check", action="store_true", help="只核對既有輸出的 SHA-256，不寫檔")
    args = parser.parse_args()
    score = json.loads(args.score.read_text(encoding="utf-8"))
    rendered = render(score)
    digest = hashlib.sha256(rendered).hexdigest()
    if args.check:
        if not args.output.is_file():
            raise SystemExit(f"缺少輸出檔：{args.output}")
        actual = hashlib.sha256(args.output.read_bytes()).hexdigest()
        if actual != digest:
            raise SystemExit(f"SHA-256 不一致：expected {digest}，actual {actual}")
        print(digest)
        return 0
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_bytes(rendered)
    print(digest)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
