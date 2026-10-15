#!/usr/bin/env python3
"""Render and master the user-approved eight cues. Run only in bounded Docker."""
import array
import hashlib
import html
import json
import math
import os
from pathlib import Path
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parents[2]
GAME = ROOT
OUT = GAME / "assets/music/modern"
VERSION = "v.1.0.2-20261015"
RATE = 44100
TARGET = -18.0
CUES = {
    "modern_scene": ("scene", "開局與新聞"),
    "modern_strategy": ("strategy", "政略操作"),
    "modern_main_theme": ("main-theme", "全國地圖主題"),
    "modern_battle_a": ("battle-1", "戰鬥一"),
    "modern_battle_b": ("battle-2", "戰鬥二"),
    "modern_battle_alt": ("battle-alt", "替代戰鬥"),
    "modern_story": ("wall", "人物自傳"),
    "modern_final": ("final", "離開確認與結局"),
}


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def run(args, **kwargs):
    return subprocess.run(args, check=True, capture_output=True, timeout=120, **kwargs)


def write(path, obj):
    assert not path.is_symlink()
    if path.exists():
        assert path.stat().st_uid == os.getuid()
    path.write_text(json.dumps(obj, ensure_ascii=False, indent=2) + "\n")


def loudness(path):
    result = run(["ffmpeg", "-hide_banner", "-nostdin", "-i", str(path),
                  "-af", "loudnorm=I=-18:TP=-2:LRA=11:print_format=json", "-f", "null", "-"], text=True)
    return json.loads(result.stderr[result.stderr.rfind("{"):])


def waveform(path):
    probe = json.loads(run(["ffprobe", "-v", "error", "-show_streams", "-of", "json", str(path)], text=True).stdout)
    stream = probe["streams"][0]
    assert stream["codec_name"] == "vorbis" and stream["channels"] == 2 and int(stream["sample_rate"]) == RATE
    data = run(["ffmpeg", "-v", "error", "-nostdin", "-i", str(path), "-f", "f32le", "-acodec", "pcm_f32le", "-"]).stdout
    samples = array.array("f", data)
    assert samples and all(math.isfinite(x) for x in samples)
    peak = max(abs(x) for x in samples)
    rms = math.sqrt(sum(x*x for x in samples) / len(samples))
    edge = max(abs(samples[-2]-samples[0]), abs(samples[-1]-samples[1]))
    assert .0001 < rms and peak < 1 and edge < .01, (path.name, rms, peak, edge)
    return {"codec": "vorbis", "sample_rate": RATE, "channels": 2, "frames": len(samples)//2,
            "duration_seconds": len(samples)/2/RATE, "rms": rms, "peak_dbfs": 20*math.log10(peak),
            "loop_edge_delta": edge, "decoded_float_sha256": hashlib.sha256(data).hexdigest(),
            "non_silent": True, "clipped": False}


def main():
    assert OUT.is_dir() and OUT.stat().st_uid == os.getuid()
    for path in OUT.iterdir():
        assert not path.is_symlink() and path.stat().st_uid == os.getuid()
    before = {p.name: sha(p) for p in OUT.iterdir() if p.is_file()}
    if (OUT/"provenance.json").exists():
        previous = json.loads((OUT/"provenance.json").read_text())
        if previous.get("version") == VERSION and "prior_runtime_files" in previous:
            before = previous["prior_runtime_files"]
    records, tracks = [], {}
    with tempfile.TemporaryDirectory(prefix="dsds-master-") as temp:
        work = Path(temp)
        run(["go", "build", "-buildvcs=false", "-o", str(work/"renderer"), "./cmd/modern_audio"], cwd=GAME)
        run([str(work/"renderer"), "-out", str(work/"pcm"), "-sample-rate", str(RATE),
             "-effects=false", "-report", str(OUT/"source-pcm.json")], cwd=GAME)
        for cue, (track, title) in CUES.items():
            wav = work/"pcm"/(track+".wav")
            measured = loudness(wav)
            gain = min(TARGET-float(measured["input_i"]), -2.5-float(measured["input_tp"]))
            target = work/(cue+".ogg")
            command = ["ffmpeg", "-hide_banner", "-loglevel", "error", "-nostdin", "-fflags", "+bitexact",
                       "-i", str(wav), "-map_metadata", "-1", "-af", f"volume={gain:.6f}dB",
                       "-c:a", "libvorbis", "-q:a", "5", "-ar", str(RATE), "-ac", "2",
                       "-flags:a", "+bitexact", "-fflags", "+bitexact", "-y", str(target)]
            run(command)
            second = work/(cue+"-repeat.ogg")
            run(command[:-1]+[str(second)])
            assert sha(target) == sha(second), cue
            wave = waveform(target)
            mixed = loudness(target)
            assert abs(float(mixed["input_i"])-TARGET) < .35 and float(mixed["input_tp"]) < -2, (cue, mixed)
            meta = {"file": target.name, "loop_start": 0, "loop_length": 0,
                    "author": "great-era-remake project; original pure-Go procedural composition",
                    "license": "LicenseRef-RRSAL-1.0; see LICENSE", "sha256": sha(target)}
            records.append({"cue": cue, "title": title, "source_track": track,
                            "source_pcm_wav_sha256": sha(wav), "source_lufs": float(measured["input_i"]),
                            "gain_db": round(gain, 6), "integrated_lufs": float(mixed["input_i"]),
                            "true_peak_dbtp": float(mixed["input_tp"]), "bytes": target.stat().st_size,
                            "sha256": meta["sha256"], "encoder_rebuild_identical": True,
                            "waveform": wave, "encoding_command": command[1:-1]+[target.name]})
            tracks[cue] = meta
            (OUT/target.name).write_bytes(target.read_bytes())
            print(f"Mastered {cue}: {mixed['input_i']} LUFS, {mixed['input_tp']} dBTP", flush=True)
    manifest = {"schema": 1, "version": VERSION, "tracks": tracks}
    write(OUT/"manifest.json", manifest)
    sources = [GAME/"internal/ui/audio/procedural.go", GAME/"cmd/modern_audio/main.go",
               GAME/"internal/ui/audio/modern.go", GAME/"go.mod", GAME/"go.sum", GAME/"LICENSE", Path(__file__)]
    provenance = {"schema": 1, "version": VERSION, "status": "eight-cue-mastered-ogg-technical-qa-passed",
                  "composition": "existing original procedural compositions retained; fixed-gain mastering only",
                  "credit": "great-era-remake project; generated by original project source, no named human composer claimed",
                  "copyright_notice": "Copyright (c) 2026 Wang Chun-Yu (wicanr2); see repository LICENSE",
                  "license": "LicenseRef-RRSAL-1.0", "original_assets_read": [], "external_samples": [], "soundfont": None,
                  "user_direction_approval": "2026-10-13: retain these compositions, finish mixing/Ogg/playback QA",
                  "human_listening_of_new_mix": "not performed", "physical_device_listening": "not performed",
                  "public_movie_clearance": "music is project-original; game images/original fonts remain local-only",
                  "encoding": {"codec": "libvorbis", "quality": 5, "sample_rate": RATE, "channels": 2,
                               "target_lufs": TARGET, "true_peak_limit_dbtp": -2, "gain": "constant per cue; no compression or new arrangement",
                               "loop": "entire decoded track; unchanged composition duration; zero endpoints measured"},
                  "sources": [{"file": str(p.relative_to(ROOT)), "sha256": sha(p)} for p in sources],
                  "tools": {name: run(args, text=True).stdout.splitlines()[0] for name,args in {
                      "go":["go","version"], "ffmpeg":["ffmpeg","-version"], "ffprobe":["ffprobe","-version"]}.items()},
                  "prior_runtime_files": before, "manifest_sha256": sha(OUT/"manifest.json"), "tracks": records}
    write(OUT/"provenance.json", provenance)
    write(ROOT/"audio-release-verification.json", provenance)
    cards = []
    for item in records:
        cue = item["cue"]
        cards.append(f'<article><h2>{html.escape(item["title"])}</h2><audio controls preload="none" src="game/assets/music/modern/{cue}.ogg"></audio><p>{item["integrated_lufs"]:.2f} LUFS · {item["true_peak_dbtp"]:.2f} dBTP</p></article>')
    (ROOT/"audio-release-review.html").write_text('<!doctype html><html lang="zh-Hant"><meta charset="utf-8"><meta name="viewport" content="width=device-width"><title>八曲混音試聽</title><style>body{margin:32px auto;padding:0 20px;max-width:1000px;background:#211c19;color:#e7dcc6;font:16px system-ui}h1{color:#ca8c61}main{display:grid;grid-template-columns:repeat(auto-fit,minmax(260px,1fr));gap:16px}article{background:#352c24;border:1px solid #756047;padding:18px}h2{font-size:20px}audio{width:100%}a{color:#dab074}</style><h1>大時代的故事 · 八曲正式混音</h1><p>'+VERSION+'。沿用已確認的原創程序配樂，統一音量。八曲均以 Ogg 播放，地圖與替代戰鬥不再依賴程序後備。</p><p>44.1 kHz、雙聲道，約 -18 LUFS。技術驗收通過；新混音的人耳與實體裝置聽審尚未執行。</p><main>'+''.join(cards)+'</main><p><a href="audio-review.html">改版前技術預覽</a> · <a href="audio-release-verification.json">混音來源與量測</a></p><script>document.querySelectorAll("audio").forEach(a=>{a.loop=true;a.addEventListener("play",()=>document.querySelectorAll("audio").forEach(b=>{if(b!==a)b.pause()}))})</script></html>')
    print("All eight release cues mastered; old preview preserved", flush=True)


if __name__ == "__main__":
    main()
