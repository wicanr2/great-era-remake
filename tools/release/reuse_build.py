"""Reuse the already verified full-package binaries for an identical-source public package."""
from pathlib import Path
import hashlib
import json
import os
import shutil
import subprocess
import sys
import zipfile

stage, output = Path('/stage'), Path('/output')
version = sys.argv[1]
sha = lambda p: hashlib.sha256(p.read_bytes()).hexdigest()
assert stage.stat().st_uid == os.getuid()
release = json.loads((output / 'SHA256SUMS.json').read_text())
current = json.loads((stage / 'source.json').read_text())
previous = release['packages']['full-local']['source']
assert release['release_version'] == current['version'] == previous['version'] == version
for key in ['files', 'tool_images', 'source_commit']:
    assert current[key] == previous[key], key
assert not current['dirty_source_snapshot'] and not previous['dirty_source_snapshot']
proof = dict(version=version, source_commit=current['source_commit'],
             regression='same-source full-local build; not rerun', binaries={}, logs={})
for kind, member, target in [
    ('windows', 'GreatEraRemake/dsds.exe', 'dsds.exe'),
    ('macos', 'GreatEraRemake.app/Contents/MacOS/dsds', 'dsds-macos'),
]:
    archive = next((output / 'full-local').glob('*' + kind + '*.zip'))
    assert sha(archive) == release['archives'][archive.name]['sha256']
    with zipfile.ZipFile(archive) as z:
        data = z.read(member)
    p = stage / 'build' / target
    p.write_bytes(data)
    p.chmod(0o755)
    assert sha(p) == release['binaries'][kind]['sha256']
    proof['binaries'][kind] = sha(p)
archive = next((output / 'full-local').glob('*.AppImage'))
assert sha(archive) == release['archives'][archive.name]['sha256']
extract = stage / 'reuse-full-appimage'
extract.mkdir()
subprocess.run([str(archive), '--appimage-extract'], cwd=extract, check=True,
               stdout=subprocess.DEVNULL, timeout=30)
p = stage / 'build/dsds-linux'
shutil.copy2(extract / 'squashfs-root/usr/bin/dsds', p)
assert sha(p) == release['binaries']['linux']['sha256']
proof['binaries']['linux'] = sha(p)
for name in ['full-tests.log', 'no-cgo.log', 'linux-version.txt']:
    source = output / 'smoke' / name
    shutil.copy2(source, stage / name)
    proof['logs'][name] = sha(source)
assert (stage / 'linux-version.txt').read_text().strip() == version
testlog = (stage / 'full-tests.log').read_text()
assert testlog.count('ok\t') == 19 and 'FAIL' not in testlog
(stage / 'reuse-proof.json').write_text(json.dumps(proof, ensure_ascii=False, indent=2) + '\n')
print('Public package reuses three verified binaries and the same-source 19-package regression.')
