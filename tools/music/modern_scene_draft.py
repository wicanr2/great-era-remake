#!/usr/bin/env python3
"""產生 `modern_scene` 的決定性、原創、不可發行 MIDI 草稿。

此工具只讀同目錄 score.json；它絕不讀 workplace/、原版 MUS/TIM 或分析 MIDI。
輸出是用於人類編曲審稿的 Standard MIDI File，不能取代人耳 QA、權利清查或 Ogg 混音。
"""

from __future__ import annotations

import argparse
import hashlib
import json
import struct
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
DEFAULT_SCORE = ROOT / "docs/music/modern_scene/score.json"
DEFAULT_OUTPUT = ROOT / "docs/music/modern_scene/modern_scene_draft.mid"


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


def conductor_track(score: dict) -> bytes:
    bpm = score["tempo_bpm"]
    numerator, denominator = score["time_signature"]
    tempo_us = round(60_000_000 / bpm)
    denominator_power = {1: 0, 2: 1, 4: 2, 8: 3, 16: 4}[denominator]
    events = [
        (0, meta(0x03, b"modern_scene conductor")),
        (0, meta(0x01, b"DRAFT - NOT FOR RELEASE")),
        (0, meta(0x51, tempo_us.to_bytes(3, "big"))),
        (0, meta(0x58, bytes((numerator, denominator_power, 24, 8)))),
    ]
    return encode_events(events, score["loop"]["length_ticks"])


def note_track(track: dict, end_tick: int) -> bytes:
    channel = track["channel"]
    if not 0 <= channel <= 15:
        raise ValueError("MIDI channel 超出範圍")
    events: list[tuple[int, bytes]] = [
        # MIDI text meta events are byte strings; UTF-8 keeps the score title readable
        # without changing any musical event bytes.
        (0, meta(0x03, track["name"].encode("utf-8"))),
        (0, bytes((0xC0 | channel, track["program"]))),
    ]
    for start, pitch, duration, velocity in track["notes"]:
        if not (0 <= pitch <= 127 and 1 <= duration and 1 <= velocity <= 127):
            raise ValueError("score.json 的 note 欄位超出 MIDI 值域")
        if start < 0 or start + duration > end_tick:
            raise ValueError("音符超出八小節 loop 範圍")
        # 同 tick 先放 Note Off，避免重音重疊時掛音。
        events.append((start, bytes((0x90 | channel, pitch, velocity))))
        events.append((start + duration, bytes((0x80 | channel, pitch, 0))))
    return encode_events(events, end_tick)


def encode_events(events: list[tuple[int, bytes]], end_tick: int) -> bytes:
    # Note Off (0x8n) 先於 Note On；其餘同 tick 按位元組穩定排序，輸出固定。
    ordered = sorted(events, key=lambda item: (item[0], 0 if item[1][0] & 0xF0 == 0x80 else 1, item[1]))
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


def render(score: dict) -> bytes:
    if score["schema"] != 1 or score["cue"] != "modern_scene":
        raise ValueError("只接受 schema 1 的 modern_scene score")
    if score["status"] != "draft-not-for-release":
        raise ValueError("草稿狀態必須保持 draft-not-for-release")
    if score["ppqn"] != 96:
        raise ValueError("本草稿固定使用 96 PPQN")
    if score["loop"]["bars"] != 8 or score["loop"]["length_ticks"] != 3072:
        raise ValueError("本草稿固定為 8 小節／3,072 tick")
    header = chunk(b"MThd", struct.pack(">HHH", 1, len(score["tracks"]) + 1, score["ppqn"]))
    tracks = [conductor_track(score)]
    tracks.extend(note_track(track, score["loop"]["length_ticks"]) for track in score["tracks"])
    return header + b"".join(tracks)


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
