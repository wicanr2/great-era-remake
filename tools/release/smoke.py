#!/usr/bin/env python3
"""Run only in a bounded Docker image with Xvfb, xdotool and ImageMagick.

Usage: smoke.py VERSION linux|windows full-local|patch
Mount release root at /release (writable), source at /src (read-only).
Linux additionally uses PulseAudio/parec; Windows uses the existing Wine image.
"""
from pathlib import Path
import array, hashlib, json, math, os, signal, subprocess, sys, time, zipfile

VERSION, PLATFORM, KIND=sys.argv[1:]
assert PLATFORM in {'linux','windows'} and KIND in {'full-local','patch'}
RELEASE=Path('/release'); OUT=RELEASE/'smoke'/f'{KIND}-{PLATFORM}'
assert not OUT.exists(); OUT.mkdir()
assert OUT.stat().st_uid==os.getuid()
ENV={**os.environ,'DISPLAY':':96','LIBGL_ALWAYS_SOFTWARE':'1',
     'XDG_CONFIG_HOME':'/tmp/smoke-home/config','XDG_DATA_HOME':'/tmp/smoke-home/data',
     'XDG_RUNTIME_DIR':'/tmp/smoke-runtime','WINEPREFIX':'/tmp/smoke-wine',
     'XDG_CACHE_HOME':'/tmp/smoke-cache','ALSA_CONFIG_PATH':'/tmp/smoke-alsa.conf',
     'WINEDEBUG':'-all','WINEDLLOVERRIDES':'mscoree,mshtml=d'}
for p in [ENV['XDG_CONFIG_HOME'],ENV['XDG_DATA_HOME'],ENV['XDG_RUNTIME_DIR'],ENV['XDG_CACHE_HOME']]:Path(p).mkdir(mode=0o700,parents=True)
procs=[]; events=[]; shots={}; window=None
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
def cmd(*args,check=True,timeout=30):
    return subprocess.run(args,env=ENV,stdout=subprocess.PIPE,stderr=subprocess.PIPE,text=True,check=check,timeout=timeout).stdout.strip()
def spawn(args,name,cwd='/tmp'):
    f=(OUT/(name+'.log')).open('w');p=subprocess.Popen(args,env=ENV,cwd=cwd,stdout=f,stderr=f,start_new_session=True);procs.append((p,f));return p
def key(k):
    events.append({'key':k});cmd('xdotool','keydown','--clearmodifiers',k);time.sleep(.2);cmd('xdotool','keyup',k);time.sleep(.5)
def click(x,y):
    events.append({'click':[x,y]});cmd('xdotool','mousemove','--window',window,str(x),str(y),'mousedown','1');time.sleep(.25);cmd('xdotool','mouseup','1');time.sleep(.75)
def shot(name):
    p=OUT/(name+'.png');cmd('import','-display',':96','-window',window,str(p));assert float(cmd('identify','-format','%[standard-deviation]',str(p)))>2000;shots[name]=sha(p)
def pref():return json.loads(PREF.read_text())
def select(row,mode=None,track=None):
    modern='1280x800' in cmd('xdotool','getwindowgeometry',window)
    click(840,24);click(840,(79 if modern else 69)+36*row);deadline=time.monotonic()+45
    while time.monotonic()<deadline:
        if PREF.is_file():
            p=pref()
            if (mode is None or p.get('audio')==mode) and (track is None or p.get('music_track')==track):return
        time.sleep(.2)
    shot('selection-failed');raise TimeoutError('music selection')
def launch(name):
    global window
    p=spawn(ARGS,name);deadline=time.monotonic()+70
    while time.monotonic()<deadline:
        assert p.poll() is None,(OUT/(name+'.log')).read_text(errors='replace')[-3000:]
        result=cmd('xdotool','search','--onlyvisible','--name','^大時代的故事$',check=False)
        if result:
            window=result.splitlines()[-1];cmd('xdotool','windowfocus',window);cmd('xdotool','windowmove',window,'0','0')
            probe=OUT/'probe.png';cmd('import','-display',':96','-window',window,str(probe))
            if float(cmd('identify','-format','%[standard-deviation]',str(probe)))>2000:
                assert any(size in cmd('xdotool','getwindowgeometry',window) for size in ['1280x768','1280x800']);return p
        time.sleep(.2)
    raise TimeoutError('packaged raw binary')
def quit(p):key('F10');key('y');p.wait(timeout=25);assert p.returncode==0
def record(name='default-hd'):
    time.sleep(1);path=OUT/(name+'.pcm');f=path.open('wb')
    p=subprocess.Popen(['parec','--device=qa.monitor','--format=s16le','--rate=44100','--channels=2'],env=ENV,stdout=f,stderr=subprocess.DEVNULL)
    try:
        deadline=time.monotonic()+15
        while time.monotonic()<deadline and path.stat().st_size<44100*4*3:time.sleep(.1)
    finally:p.terminate();p.wait(timeout=5);f.close()
    data=path.read_bytes();a=array.array('h');a.frombytes(data[:len(data)//2*2]);assert len(a)>=44100*2*3
    rms=math.sqrt(sum(x*x for x in a)/len(a));assert rms>20
    return {'rms':rms,'frames':len(a)//2,'sha256':sha(path)}
def stop(sig,frame):raise KeyboardInterrupt('bounded smoke stopped')
signal.signal(signal.SIGTERM,stop)
try:
    sums=json.loads((RELEASE/'SHA256SUMS.json').read_text())
    archive=next((RELEASE/KIND).glob('*'+('windows' if PLATFORM=='windows' else 'linux')+'*'))
    assert sha(archive)==sums['archives'][archive.name]['sha256']
    if archive.suffix=='.AppImage':
        subprocess.run([str(archive),'--appimage-extract'],cwd=OUT,env=ENV,stdout=subprocess.DEVNULL,check=True,timeout=45)
        folder=OUT/'squashfs-root'
    else:
        with zipfile.ZipFile(archive) as z:
            assert z.testzip() is None
            for item in z.infolist():
                assert not Path(item.filename).is_absolute() and '..' not in Path(item.filename).parts
                assert ((item.external_attr>>16)&0o170000)!=0o120000
            z.extractall(OUT)
            for item in z.infolist():(OUT/item.filename).chmod((item.external_attr>>16)&0o777)
        folder=OUT/('GreatEraRemake' if PLATFORM=='windows' else 'AppDir')
    manifest=json.loads((folder/'RELEASE-MANIFEST.json').read_text())
    assert manifest['release_version']==VERSION
    binary=folder/('dsds.exe' if PLATFORM=='windows' else 'usr/bin/dsds')
    spawn(['Xvfb',':96','-screen','0','1440x900x24','-nolisten','tcp'],'xvfb');time.sleep(.5)
    if PLATFORM=='linux':
        ENV['PULSE_SERVER']='unix:/tmp/smoke-runtime/native'
        spawn(['pulseaudio','--daemonize=no','--exit-idle-time=-1','--use-pid-file=no','-n',
               '--load=module-native-protocol-unix socket=/tmp/smoke-runtime/native auth-anonymous=1',
               '--load=module-null-sink sink_name=qa format=s16le rate=44100 channels=2'],'pulse')
        for _ in range(60):
            try:cmd('pactl','info');break
            except subprocess.CalledProcessError:time.sleep(.1)
        ENV.pop('ALSA_CONFIG_PATH',None)
        ARGS=[str(binary)];PREF=Path(ENV['XDG_CONFIG_HOME'])/'dsds/prefs.json';SAVE=Path(ENV['XDG_DATA_HOME'])/'GreatEraRemake/Saves/campaign.gesave'
    else:
        Path(ENV['ALSA_CONFIG_PATH']).write_text('pcm.!default { type null }\nctl.!default { type hw card 0 }\n')
        Path(ENV['WINEPREFIX']).mkdir(mode=0o700)
        cmd('wineboot','-u',timeout=75)
        cmd('wine','reg','add',r'HKCU\Software\Wine\Drivers','/v','Audio','/d','alsa','/f')
        local=cmd('wine','cmd','/c','echo','%LOCALAPPDATA%');roaming=cmd('wine','cmd','/c','echo','%APPDATA%')
        SAVE=Path(cmd('winepath','-u',local))/'GreatEraRemake/Saves/campaign.gesave';PREF=Path(cmd('winepath','-u',roaming))/'dsds/prefs.json'
        ARGS=['wine',str(binary)]
    if KIND=='patch':
        convert=lambda s:cmd('winepath','-w',s) if PLATFORM=='windows' else s
        ARGS+=['-game',convert('/src/workplace/orig/game'),'-eten',convert('/src/workplace/eten')]
    cmd('setxkbmap','us');p=launch('first-zero-args' if KIND=='full-local' else 'engine-import');shot('01-default-hd')
    technical_audio=record() if PLATFORM=='linux' else {'scope':'Wine ALSA null; physical listening unverified'}
    key('Return');time.sleep(2);shot('01-modern-map')
    assert '1280x800' in cmd('xdotool','getwindowgeometry',window)
    for k,n in [('F2','02-original-art'),('F4','03-original-ui'),('F2','04-hd-art-original-ui'),('F4','05-command-ui')]:key(k);shot(n)
    military_audio={}
    if KIND=='full-local':
        for row,track in [(12,'huan-wo-he-shan'),(13,'national-revolution'),(14,'eight-hundred-heroes')]:
            select(row,track=track);shot('music-'+track)
            if PLATFORM=='linux':military_audio[track]=record(track)
        select(1,mode='retro');shot('06-retro-music');assert pref()['music_track']=='eight-hundred-heroes'
        select(2,mode='off');select(0,mode='modern');assert pref()['music_track']=='eight-hundred-heroes'
    else:
        before=pref();click(840,24);click(840,79+36*12);time.sleep(4)
        assert pref()==before,'missing military song changed public-package preferences'
        shot('missing-military-retained');select(1,mode='retro');shot('06-retro-music');select(0,mode='modern')
    select(8,track='battle-2');select(2,mode='off');shot('07-muted-manual-track')
    click(72,269);click(510,631);shot('08-recruit');click(300,731);shot('08-confirm');key('Escape');shot('08-adjust');key('Escape');shot('08-return-map')
    key('ctrl+s');click(810,488);time.sleep(2);assert SAVE.is_file();saved=sha(SAVE);quit(p)
    p=launch('restart-zero-args');shot('09-restart');assert pref()['audio']=='off' and pref()['music_track']=='battle-2';key('c');time.sleep(2);shot('09-continued')
    select(0,mode='modern');shot('10-hd-music-restored');select(2,mode='off');quit(p);assert sha(SAVE)==saved
    for relative,r in manifest['files'].items():assert sha(folder/relative)==r['sha256'],relative
    with zipfile.ZipFile(SAVE) as z:
        state=z.read('state.json');metadata=json.loads(z.read('manifest.json'))
        assert hashlib.sha256(state).hexdigest()==metadata['StateSHA256']
    import shutil;shutil.copy2(SAVE,OUT/'campaign.gesave')
    report={'version':VERSION,'platform':PLATFORM,'package_kind':KIND,'seed':1,'normal_player_ui':True,
      'raw_binary_from_unrelated_cwd':True,'zero_game_arguments':KIND=='full-local','launch_arguments':ARGS,
      'default_art_interface':'HD/Modern/high/mouse','default_audio':'modern','audio':technical_audio,'military_audio':military_audio,
      'music_switches':['retro','modern','manual battle-2','off','restart remembered','modern restored'],
      'save_sha256':saved,'actual_save_location':str(SAVE),'preferences':pref(),'packaged_inputs_unchanged':True,
      'binary_sha256':sha(binary),'archive_sha256':sha(archive),'screenshots':shots,'events':events,
      'scope':'Linux container Pulse routing' if PLATFORM=='linux' else 'Wine auxiliary GUI and ALSA null; no physical Windows claim'}
    (OUT/'gui-report.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n');print(json.dumps(report,ensure_ascii=False),flush=True)
finally:
    for p,f in reversed(procs):
        if p.poll() is None:
            try:os.killpg(p.pid,signal.SIGTERM);p.wait(timeout=6)
            except (ProcessLookupError,subprocess.TimeoutExpired):
                if p.poll() is None:os.killpg(p.pid,signal.SIGKILL);p.wait(timeout=5)
        f.close()
    if PLATFORM=='windows':subprocess.run(['wineserver','-k'],env=ENV,check=False,timeout=15)
