#!/usr/bin/env python3
"""Prepare, assemble and independently verify local three-platform candidates."""
import hashlib
import json
import os
from pathlib import Path
import plistlib
import re
import shutil
import struct
import subprocess
import sys
import zipfile

ROOT = Path('/work')
STAGE = Path('/stage')
OUT = Path('/output')
VERSION_RE = re.compile(r'v\.[0-9]+\.[0-9]+\.[0-9]+-[0-9]{8}')
DENIED = {'.tpc','.15','.rgb','.mus','.tim','.dat','.glb','.gtb','.dt1','.dt2','.sav','.cps','.bgi','.ogg','.wav','.mid','.midi'}
FULL_LOCAL = os.environ.get('DSDS_PACKAGE_FULL_LOCAL', '0') == '1'
PACKAGE_DIR = 'full-local' if FULL_LOCAL else 'patch'
PACKAGE_SUFFIX = '-full' if FULL_LOCAL else ''
RELEASE_CHANNEL = os.environ.get('DSDS_PACKAGE_RELEASE_CHANNEL', 'candidate')
assert RELEASE_CHANNEL in {'candidate', 'stable'}


def release_status():
    if RELEASE_CHANNEL == 'stable':
        return 'official-local-full-private-not-published' if FULL_LOCAL else 'official-engine-not-published'
    return 'local-full-private-not-published' if FULL_LOCAL else 'local-candidate-not-published'


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def write_json(path, data):
    path.write_text(json.dumps(data, ensure_ascii=False, indent=2)+'\n')


def inventory(folder):
    result={}
    for path in sorted(folder.rglob('*')):
        assert not path.is_symlink(), path
        if path.is_file() and path.name!='RELEASE-MANIFEST.json':
            result[path.relative_to(folder).as_posix()]={'sha256':sha(path),'bytes':path.stat().st_size}
    return result


def check_payload(folder):
    resources = folder
    for subdir in ['Contents/Resources', 'usr/share/great-era']:
        if (folder/subdir).is_dir():resources=folder/subdir
    private = json.loads((STAGE/'private-inputs.json').read_text()) if FULL_LOCAL else {}
    music = json.loads((STAGE/'owned-audio.json').read_text())
    for path in folder.rglob('*'):
        assert not path.is_symlink(),path
        assert path.name not in {'workplace','orig','player'},path
        if path.is_file():
            relative = path.relative_to(resources).as_posix() if path.is_relative_to(resources) else ''
            if relative in private:
                assert sha(path)==private[relative]['sha256'],path
                continue
            if relative in music:
                assert sha(path)==music[relative]['sha256'],path
                continue
            assert path.suffix.lower() not in DENIED,path
            assert path.suffix.lower()!='.png',path
            if path.suffix.lower()=='.exe':
                assert path.name=='dsds.exe',path


def manifest(folder, version, platform, entry):
    check_payload(folder)
    write_json(folder/'RELEASE-MANIFEST.json',{
        'schema':1,'release_version':version,'status':release_status(),'release_channel':RELEASE_CHANNEL,
        'platform':platform,'entrypoint':entry,'contains_original_game_data':FULL_LOCAL,
        'contains_original_audio_or_font':FULL_LOCAL,'requires_user_supplied_legal_game_data':not FULL_LOCAL,
        'distribution':'local use only; original game data/font not cleared for redistribution' if FULL_LOCAL else 'engine and project-original music; legal original data/font must be supplied by user',
        'audio':'eight project-original mastered Ogg cues enabled; user listening accepted; other platform devices unverified; -audio off disables',
        'device_validation':'pending; container and Wine checks are recorded separately',
        'source_manifest':'SHA256SUMS.json in release root','files':inventory(folder)})


def zip_tree(folder, target):
    assert not target.exists(),target
    with zipfile.ZipFile(target,'x',compression=zipfile.ZIP_DEFLATED,compresslevel=9) as z:
        for path in sorted(folder.rglob('*')):
            if not path.is_file():continue
            info=zipfile.ZipInfo(path.relative_to(folder.parent).as_posix(),(1980,1,1,0,0,0))
            info.create_system=3
            info.external_attr=(path.stat().st_mode&0xffff)<<16
            info.compress_type=zipfile.ZIP_DEFLATED
            z.writestr(info,path.read_bytes())


def copy_common(destination):
    shutil.copytree(STAGE/'common',destination,dirs_exist_ok=True)


def collect_licenses():
    destination=STAGE/'common/docs/licenses/runtime'
    destination.mkdir()
    shutil.copy2(Path('/usr/local/go/LICENSE'),destination/'Go-LICENSE.txt')
    modules={}
    for binary in (STAGE/'build').iterdir():
        text=subprocess.check_output(['go','version','-m',str(binary)],text=True)
        for line in text.splitlines():
            fields=line.split()
            if fields and fields[0]=='dep':modules[fields[1]]=fields[2]
    records=[]
    for module,version in sorted(modules.items()):
        escaped=''.join('!'+c.lower() if c.isupper() else c for c in module)
        source=Path('/go/pkg/mod')/(escaped+'@'+version)
        assert source.is_dir(),('missing linked module source',module,version)
        notices=[]
        for path in sorted(source.rglob('*')):
            if not path.is_file() or not path.name.upper().startswith(('LICENSE','NOTICE','COPYING','PATENTS','AUTHORS')):
                continue
            if path.suffix.lower() not in ('','.txt','.md','.rst'):continue
            content=path.read_bytes();content.decode('utf-8')
            name=(module+'@'+version+'/'+str(path.relative_to(source))).replace('/','__')
            (destination/name).write_bytes(content)
            notices.append({'file':'runtime/'+name,'sha256':hashlib.sha256(content).hexdigest()})
        assert notices,('missing module license',module)
        records.append({'module':module,'version':version,'license_files':notices})
    write_json(STAGE/'common/docs/licenses/THIRD-PARTY-SOFTWARE.json',{
        'schema':1,'scope':'Go runtime and actual linked modules across three binaries',
        'go_license':'runtime/Go-LICENSE.txt','modules':records,
        'appimage_runtime':'local-only fixed runtime; source-to-binary public review pending' if FULL_LOCAL else 'not included in public engine ZIP'})


def prepare(version):
    for d in ['build','common','assemble']:(STAGE/d).mkdir()
    common=STAGE/'common'
    for name in ['LICENSE','translations','docs/licenses']:
        source=ROOT/name;target=common/name
        target.parent.mkdir(parents=True,exist_ok=True)
        if source.is_dir():shutil.copytree(source,target)
        else:shutil.copy2(source,target)
    shutil.copy2(ROOT/('docs/release/FULL-LOCAL.md' if FULL_LOCAL else 'docs/release/README.md'),common/'RUNNING.md')
    music=ROOT/'assets/music/modern'
    music_manifest=json.loads((music/'manifest.json').read_text())
    provenance=json.loads((music/'provenance.json').read_text())
    assert music_manifest['version']==version and len(music_manifest['tracks'])==8
    assert provenance['status']=='eight-cue-mastered-ogg-technical-qa-passed'
    assert not provenance['original_assets_read'] and not provenance['external_samples']
    names=['manifest.json','provenance.json','source-pcm.json','acceptance.json']
    for cue,record in music_manifest['tracks'].items():
        source=music/record['file']
        assert source.parent==music and sha(source)==record['sha256'],cue
        names.append(record['file'])
    owned_audio={}
    for name in names:
        source=music/name;target=common/'assets/music/modern'/name
        target.parent.mkdir(parents=True,exist_ok=True);shutil.copy2(source,target)
        owned_audio[target.relative_to(common).as_posix()]={'sha256':sha(source),'bytes':source.stat().st_size,
            'source':str(source),'rights':'LicenseRef-RRSAL-1.0; project-original procedural composition',
            'role':'mastered modern music or provenance'}
    write_json(STAGE/'owned-audio.json',owned_audio)
    private={}
    if FULL_LOCAL:
        game_source=Path('/src/workplace/orig/game')
        font_source=Path('/src/workplace/eten')
        assert game_source.is_dir() and font_source.is_dir()
        for source in sorted(game_source.iterdir()):
            if source.is_dir():
                assert source.name=='.jsdos',source
                continue
            assert not source.is_symlink() and source.is_file(),source
            if source.suffix.lower() in {'.exe','.sav','.bat','.bgi','.cps'}:continue
            target=common/'data/game'/source.name
            target.parent.mkdir(parents=True,exist_ok=True)
            shutil.copy2(source,target)
            private[target.relative_to(common).as_posix()]={'sha256':sha(source),'bytes':source.stat().st_size,
                'source':str(source),'rights':'original game; local-only',
                'role':'pristine initial-state data' if source.suffix.lower() in {'.dt1','.dt2'} else 'original game resource'}
        for name in ['STDFONT.15','SPCFONT.15','ASCFONT.15']:
            source=font_source/name;assert source.is_file() and not source.is_symlink()
            target=common/'data/eten'/name;target.parent.mkdir(parents=True,exist_ok=True)
            shutil.copy2(source,target)
            private[target.relative_to(common).as_posix()]={'sha256':sha(source),'bytes':source.stat().st_size,
                'source':str(source),'rights':'ETEN font; local-only','role':'complete Chinese font'}
        runtime=Path('/src/workplace/hoi4-ui-prototype/runtime-sources')
        runtime_report=json.loads((runtime/'verification.json').read_text())
        for record in runtime_report['full_license_files']:
            source=runtime/record['file'];assert sha(source)==record['sha256']
            target=common/'docs/licenses/appimage'/source.name
            target.parent.mkdir(parents=True,exist_ok=True);shutil.copy2(source,target)
        (common/'LOCAL-ONLY.txt').write_text('本機完整版，含原版遊戲資料與倚天字庫。未取得公開散布權，不可上傳或公開分享。\n配樂為專案原創八曲 Ogg 混音，技術驗收通過；新混音人耳及實體裝置驗收未執行。\n遊戲資料唯讀使用；玩家進度寫系統使用者目錄。\n')
    write_json(STAGE/'private-inputs.json',private)
    source={}
    for name in ['cmd','internal','translations','docs/licenses','docs/release','tools/release','assets/music/modern','go.mod','go.sum','LICENSE','tools/package.sh']:
        path=ROOT/name
        for f in sorted(path.rglob('*')) if path.is_dir() else [path]:
            if f.is_file() and not f.is_symlink():source[f.relative_to(ROOT).as_posix()]=sha(f)
    write_json(STAGE/'source.json',{'version':version,'files':source,
        'tool_images':{x:os.environ[x] for x in ['DSDS_GO_IMAGE','DSDS_MAC_IMAGE','DSDS_APPIMAGE_IMAGE']},
        'dirty_source_snapshot':True,'package_kind':PACKAGE_DIR,'release_channel':RELEASE_CHANNEL,'private_input_files':private,
        'original_game_inputs':[name for name in private if name.startswith('data/')], 'project_original_music':owned_audio})


def assemble(version):
    assert (STAGE/'linux-version.txt').read_text().strip()==version
    collect_licenses()
    (OUT/PACKAGE_DIR).mkdir()
    (OUT/'smoke').mkdir(exist_ok=True)
    windows=STAGE/'assemble/GreatEraRemake'
    copy_common(windows)
    shutil.copy2(STAGE/'build/dsds.exe',windows/'dsds.exe')
    bundled=('-game data/game -eten data/eten ' if FULL_LOCAL else '')+'-modern-audio assets/music/modern -audio modern -locale translations/zh-Hant '
    (windows/'launch.cmd').write_text('@echo off\r\nsetlocal\r\ncd /d "%~dp0"\r\n'
        'set "DSDS_SAVES=%LOCALAPPDATA%\\GreatEraRemake\\Saves"\r\n'
        'if not exist "%DSDS_SAVES%" mkdir "%DSDS_SAVES%"\r\n'
        'dsds.exe '+bundled+'-save "%DSDS_SAVES%\\SAVE(1).DT1" %*\r\n',encoding='utf-8',newline='')
    manifest(windows,version,'windows-amd64','launch.cmd')
    zip_tree(windows,OUT/PACKAGE_DIR/f'great-era-remake-windows-amd64{PACKAGE_SUFFIX}-{version}.zip')
    app=STAGE/'assemble/GreatEraRemake.app'
    mac=app/'Contents/MacOS';resources=app/'Contents/Resources'
    mac.mkdir(parents=True);copy_common(resources)
    shutil.copy2(STAGE/'build/dsds-macos',mac/'dsds')
    (mac/'launch').write_text('#!/bin/sh\nset -eu\ncd "$(dirname "$0")/../Resources"\n'
        'dsds_data="$HOME/Library/Application Support/GreatEraRemake/Saves"\n'
        'mkdir -p "$dsds_data"\nexec ../MacOS/dsds '+bundled+'-save "$dsds_data/SAVE(1).DT1" "$@"\n')
    (mac/'launch').chmod(0o755)
    core=version[2:].split('-')[0]
    with (app/'Contents/Info.plist').open('wb') as f:
        plistlib.dump({'CFBundleExecutable':'launch','CFBundleIdentifier':'org.great-era-remake.dsds',
            'CFBundleName':'Great Era Remake','CFBundlePackageType':'APPL','CFBundleShortVersionString':core,
            'CFBundleVersion':core,'GreatEraReleaseVersion':version,'LSMinimumSystemVersion':'12.0'},f)
    manifest(app,version,'darwin-arm64','Contents/MacOS/launch')
    zip_tree(app,OUT/PACKAGE_DIR/f'great-era-remake-macos-arm64{PACKAGE_SUFFIX}-{version}.zip')
    appdir=STAGE/'AppDir'
    (appdir/'usr/bin').mkdir(parents=True)
    copy_common(appdir/'usr/share/great-era')
    shutil.copy2(STAGE/'build/dsds-linux',appdir/'usr/bin/dsds')
    shutil.copy2(ROOT/'tools/release/dsds.desktop',appdir/'dsds.desktop')
    shutil.copy2(ROOT/'tools/release/dsds.svg',appdir/'dsds.svg')
    (appdir/'AppRun').write_text('#!/bin/sh\nset -eu\n'
        'dsds_app="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"\n'
        'cd "$dsds_app/usr/share/great-era"\n'
        'dsds_data="${XDG_DATA_HOME:-$HOME/.local/share}/GreatEraRemake/Saves"\n'
        'mkdir -p "$dsds_data"\nexec "$dsds_app/usr/bin/dsds" '+bundled+'-save "$dsds_data/SAVE(1).DT1" "$@"\n')
    (appdir/'AppRun').chmod(0o755)
    manifest(appdir,version,'linux-amd64','AppRun')
    if not FULL_LOCAL:
        zip_tree(appdir,OUT/PACKAGE_DIR/f'great-era-remake-linux-x86_64-{version}.zip')


def verify_manifest(folder):
    check_payload(folder)
    record=json.loads((folder/'RELEASE-MANIFEST.json').read_text())
    assert record['files']==inventory(folder)
    assert (folder/record['entrypoint']).is_file()
    return record


def binary_info(path, kind, version):
    data=path.read_bytes()
    assert version.encode() in data,('missing embedded version',kind)
    result={'sha256':sha(path),'bytes':len(data)}
    if kind=='windows':
        assert data[:2]==b'MZ'
        off=struct.unpack_from('<I',data,0x3c)[0]
        assert data[off:off+4]==b'PE\0\0' and struct.unpack_from('<H',data,off+4)[0]==0x8664
        result['architecture']='PE amd64'
        result['runtime_minimum_os']='Windows 10 or Windows Server 2016'
    elif kind=='linux':
        assert data[:5]==b'\x7fELF\x02' and struct.unpack_from('<H',data,18)[0]==62
        result['architecture']='ELF amd64'
        symbols=subprocess.check_output(['objdump','-T',str(path)],text=True)
        versions={tuple(map(int,x.split('.'))) for x in re.findall(r'GLIBC_([0-9.]+)',symbols)}
        assert versions
        result['minimum_glibc']='.'.join(map(str,max(versions)))
    else:
        assert data[:4]==b'\xcf\xfa\xed\xfe' and struct.unpack_from('<I',data,4)[0]==0x100000c
        result['architecture']='Mach-O arm64';offset=32;libraries=[]
        for _ in range(struct.unpack_from('<I',data,16)[0]):
            cmd,size=struct.unpack_from('<II',data,offset)
            assert size>=8 and offset+size<=len(data)
            if cmd in [0xc,0x80000018,0x8000001f,0x80000023]:
                nameoff=struct.unpack_from('<I',data,offset+8)[0]
                name=data[offset+nameoff:offset+size].split(b'\0',1)[0].decode()
                assert name.startswith(('/System/Library/','/usr/lib/')),name
                libraries.append(name)
            if cmd==0x32:
                minos=struct.unpack_from('<I',data,offset+12)[0]
                result['minimum_macos']=f'{minos>>16}.{(minos>>8)&255}.{minos&255}'
            if cmd==0x1d:result['code_signature_command']=True
            offset+=size
        result['system_libraries']=libraries
        assert result.get('minimum_macos')=='12.0.0',result
        result['runtime_minimum_os']='macOS 12 Monterey'
    result['runtime_requirements_source']='https://go.dev/wiki/MinimumRequirements'
    return result


def verify(version):
    report={'schema':1,'release_version':version,'status':release_status(),'release_channel':RELEASE_CHANNEL,
            'source':json.loads((STAGE/'source.json').read_text()),'package_kind':PACKAGE_DIR,
            'distribution':'local-only' if FULL_LOCAL else 'candidate-not-published','archives':{},'binaries':{}}
    reference=os.environ.get('DSDS_REFERENCE_VERSION','')
    if reference:
        previous=json.loads((Path('/src/dist-all')/reference/'SHA256SUMS.json').read_text())
        proof={}
        for kind in ('linux','windows','macos'):
            actual=sha(STAGE/'build'/('reference-'+kind))
            assert actual==previous['binaries'][kind]['sha256'],('source rebuild differs',kind,actual)
            proof[kind]=actual
        report['source_reference_proof']={'version':reference,'three_binaries_byte_identical':True,'binaries':proof,
            'current_version_change':'same runtime source and assets; only buildVersion date changes'}
    for kind,entry in [('windows','GreatEraRemake'),('macos','GreatEraRemake.app')]:
        archive=next((OUT/PACKAGE_DIR).glob('*'+kind+'*.zip'))
        dest=STAGE/('verify-'+kind);dest.mkdir()
        with zipfile.ZipFile(archive) as z:
            assert z.testzip() is None
            for name in z.namelist():
                assert not name.startswith('/') and '..' not in Path(name).parts
            z.extractall(dest)
        folder=dest/entry
        verify_manifest(folder)
        if kind=='macos':
            with (folder/'Contents/Info.plist').open('rb') as f:info=plistlib.load(f)
            assert info['LSMinimumSystemVersion']=='12.0'
            assert info['GreatEraReleaseVersion']==version
        binary=folder/('dsds.exe' if kind=='windows' else 'Contents/MacOS/dsds')
        report['binaries'][kind]=binary_info(binary,kind,version)
        report['archives'][archive.name]={'bytes':archive.stat().st_size,'sha256':sha(archive),'crc':'passed'}
    linux=STAGE/'appimage-extracted/squashfs-root' if FULL_LOCAL else STAGE/'AppDir'
    verify_manifest(linux)
    report['binaries']['linux']=binary_info(linux/'usr/bin/dsds','linux',version)
    image=next((OUT/PACKAGE_DIR).glob('*.AppImage' if FULL_LOCAL else '*linux*.zip'))
    if FULL_LOCAL:
        assert image.read_bytes()[8:11]==b'AI\x02'
        report['appimage_runtime_sha256']=(STAGE/'appimage-runtime.sha256').read_text().split()[0]
    else:
        with zipfile.ZipFile(image) as z:
            assert z.testzip() is None
            assert all(not n.startswith('/') and '..' not in Path(n).parts for n in z.namelist())
        report['appimage_runtime']='not included in public engine ZIP'
    report['archives'][image.name]={'bytes':image.stat().st_size,'sha256':sha(image),'extraction':'passed'}
    report['open_checks']=['Windows physical device','macOS physical device','signing and notarization',
                           'multiple Linux distributions','new-mix human and physical-device listening','public promotion image rights']
    for f in ['full-tests.log','no-cgo.log','linux-version.txt']:
        target=f if FULL_LOCAL else 'patch-'+f
        shutil.copy2(STAGE/f,OUT/'smoke'/target)
    if (OUT/'SHA256SUMS.json').exists():
        previous=json.loads((OUT/'SHA256SUMS.json').read_text())
        assert previous['release_version']==version and PACKAGE_DIR not in previous.get('packages',{})
        assert not previous['archives'].keys() & report['archives'].keys()
        assert previous['binaries']==report['binaries'], 'both package kinds must use identical binaries'
        previous.setdefault('packages',{})[PACKAGE_DIR]=report
        previous['archives'].update(report['archives'])
        write_json(OUT/'SHA256SUMS.json',previous)
    else:
        report['packages']={PACKAGE_DIR:{k:v for k,v in report.items() if k!='packages'}}
        write_json(OUT/'SHA256SUMS.json',report)
    if FULL_LOCAL:write_json(OUT/'full-local/PRIVATE-INPUTS.json',json.loads((STAGE/'private-inputs.json').read_text()))
    print('Verified 3 platform archives, embedded versions, architectures and complete file hashes')


def main():
    phase,version=sys.argv[1:]
    assert VERSION_RE.fullmatch(version)
    assert STAGE.is_dir() and STAGE.stat().st_uid==os.getuid()
    if phase!='prepare':assert OUT.is_dir() and OUT.stat().st_uid==os.getuid()
    {'prepare':prepare,'assemble':assemble,'verify':verify}[phase](version)


if __name__=='__main__':main()
