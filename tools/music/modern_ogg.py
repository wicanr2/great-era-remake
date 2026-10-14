#!/usr/bin/env python3
"""Rebuild the existing procedural preview; run inside the project Docker image."""
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile

ROOT = Path(__file__).resolve().parents[2]
CUES = {
    "modern_scene": "scene", "modern_strategy": "strategy",
    "modern_battle_a": "battle-1", "modern_battle_b": "battle-2",
    "modern_story": "wall", "modern_final": "final",
}
RATE = 44100


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def run(args, **kwargs):
    return subprocess.run(args, check=True, **kwargs)


def main():
    out = Path(sys.argv[1]).resolve()
    if not out.is_dir() or out.stat().st_uid != os.getuid():
        raise ValueError("Output must be an existing directory owned by the current user")
    for entry in out.iterdir():
        if entry.is_symlink() or entry.stat().st_uid != os.getuid():
            raise ValueError(f"Unsafe existing output: {entry.name}")
    with tempfile.TemporaryDirectory(prefix="dsds-audio-") as temp:
        work = Path(temp)
        exe = work / "renderer"
        run(["go", "build", "-buildvcs=false", "-o", str(exe), "./cmd/modern_audio"], cwd=ROOT)
        wav = work / "wav"
        run([str(exe), "-out", str(wav), "-sample-rate", str(RATE),
             "-effects=false", "-report", str(out / "source-pcm.json")], cwd=ROOT)
        tracks = {}
        outputs = []
        for cue, track in list(CUES.items()) + [
                ("fallback_main-theme", "main-theme"), ("fallback_battle-alt", "battle-alt")]:
            target = out / (cue + ".ogg")
            command = ["ffmpeg", "-hide_banner", "-loglevel", "error", "-fflags", "+bitexact",
                       "-i", str(wav / (track + ".wav")), "-map_metadata", "-1",
                       "-c:a", "libvorbis", "-q:a", "4", "-ar", str(RATE),
                       "-flags:a", "+bitexact", "-fflags", "+bitexact", "-y", str(target)]
            run(command)
            second = work / target.name
            run(command[:-1] + [str(second)])
            if digest(target) != digest(second):
                raise ValueError(f"Non-deterministic Ogg output: {cue}")
            outputs.append({"cue": cue, "source_track": track, "file": target.name,
                            "sha256": digest(target), "wav_sha256": digest(wav / (track + ".wav")),
                            "bytes": target.stat().st_size, "encoder_rebuild_identical": True,
                            "runtime_manifest": cue in CUES})
            if cue in CUES:
                tracks[cue] = {
                    "file": target.name, "loop_start": 0, "loop_length": 0,
                    "author": "great-era-remake procedural generator; formal composer credit pending",
                    "license": "LicenseRef-RRSAL-1.0 (see repo LICENSE); technical preview only",
                    "sha256": digest(target),
                }
        manifest = {"schema": 1, "version": "modern-music-tech-preview-reproducible-1", "tracks": tracks}
        (out / "manifest.json").write_text(json.dumps(manifest, ensure_ascii=False, indent=2) + "\n")
        sources = ["internal/ui/audio/procedural.go", "cmd/modern_audio/main.go",
                   "tools/music/modern_ogg.py", "tools/make_modern_ogg.sh", "go.mod", "go.sum", "LICENSE"]
        provenance = {
            "schema": 1, "status": "technical-preview-not-release-cleared",
            "composition": "existing original pure-Go procedural composition; no new arrangement",
            "original_assets_read": [], "external_samples": [], "soundfont": None,
            "formal_composer": None, "human_listening": "pending", "device_listening": "pending",
            "public_movie_clearance": "pending", "license_record": "repository LICENSE; no new rights granted",
            "sources": [{"file": f, "sha256": digest(ROOT / f)} for f in sources],
            "tools": {name: subprocess.check_output(args, text=True).splitlines()[0] for name, args in {
                "go": ["go", "version"], "python": ["python3", "--version"],
                "ffmpeg": ["ffmpeg", "-version"], "ffprobe": ["ffprobe", "-version"]}.items()},
            "encoding": {"codec": "libvorbis", "quality": 4, "sample_rate": RATE, "channels": 2,
                         "metadata": "stripped", "bitexact": True,
                         "command": "ffmpeg -i SOURCE.wav -map_metadata -1 -c:a libvorbis -q:a 4 -ar 44100 -flags:a +bitexact -fflags +bitexact OUTPUT.ogg",
                         "reference": "https://ffmpeg.org/ffmpeg-formats.html#Format-Options"},
            "manifest_sha256": digest(out / "manifest.json"), "outputs": outputs,
            "fallback_note": "main-theme and battle-alt previews are excluded from the six-cue runtime manifest",
        }
        (out / "provenance.json").write_text(json.dumps(provenance, ensure_ascii=False, indent=2) + "\n")
        print(f"Generated {len(tracks)} manifest cues and 2 fallback auditions; all 8 encoder rebuilds identical")


if __name__ == "__main__":
    main()
