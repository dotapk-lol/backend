#!/usr/bin/env python3
"""Restore this task's existing QA data, never initialize/reset/install a service."""
import json, os, pathlib, shutil, subprocess, time, urllib.request
root=pathlib.Path(__file__).resolve().parents[2]
data=pathlib.Path('/tmp/duel-mysql-test-20261002')
sock=data/'mysql.sock'
runtime=root/'qa-runtime'
runtime.mkdir(exist_ok=True)
mysqladmin='/opt/homebrew/bin/mysqladmin'
mysqld='/opt/homebrew/bin/mysqld'
binary=root/'candidate-v12/bin/dueld-local'
for path in [data/'mysql.ibd',data/'ibdata1',data/'dota_duel/duel_matches.ibd',binary]:
    if not path.is_file():raise SystemExit('Required existing QA file missing: '+str(path)+'; refusing to initialize')
def ping():
    return subprocess.run([mysqladmin,'--no-defaults','--socket='+str(sock),'-uroot','ping'],stdout=subprocess.DEVNULL,stderr=subprocess.DEVNULL).returncode==0
def health():
    try:
        with urllib.request.urlopen('http://127.0.0.1:18082/healthz',timeout=2) as r:return json.load(r)
    except Exception:return None
mysql_started=False
go_started=False
if not ping():
    # A closed data directory can be copied consistently; never replace the backup.
    opened=subprocess.run(['/usr/sbin/lsof','+D',str(data)],capture_output=True,text=True)
    if opened.returncode==0 and opened.stdout.strip():raise SystemExit('QA data directory is still open; inspect existing process before restoring')
    backup=runtime/'mysql-before-recovery-20261002'
    if not backup.exists():shutil.copytree(data,backup,ignore=shutil.ignore_patterns('*.sock','*.sock.lock'))
    args=[mysqld,'--no-defaults','--datadir='+str(data),'--socket='+str(sock),'--pid-file='+str(data/'mysql.pid'),'--log-error='+str(data/'server.log'),'--skip-networking','--mysqlx=OFF','--lower-case-table-names=2','--innodb-buffer-pool-size=128M','--max-connections=20']
    with (runtime/'mysql-launch.log').open('ab') as log:
        p=subprocess.Popen(args,stdin=subprocess.DEVNULL,stdout=log,stderr=subprocess.STDOUT,start_new_session=True,close_fds=True)
    (runtime/'mysql-launch.pid').write_text(str(p.pid)+'\n')
    for _ in range(50):
        if ping():break
        if p.poll() is not None:raise SystemExit('MySQL restore failed; see existing server.log; no initialization was attempted')
        time.sleep(.2)
    else:raise SystemExit('MySQL readiness timed out; inspect server.log')
    mysql_started=True
current=health()
if current is None:
    # Refuse to replace another service on the requested loopback port.
    listening=subprocess.run(['/usr/sbin/lsof','-nP','-iTCP:18082','-sTCP:LISTEN','-t'],capture_output=True,text=True)
    if listening.stdout.strip():raise SystemExit('Port18082 is occupied but not healthy; inspect before replacing')
    env=dict(os.environ,DUEL_MYSQL_DSN='root@unix('+str(sock)+')/dota_duel',DUEL_ALLOWED_ORIGIN='http://127.0.0.1:4173',DUEL_LISTEN='127.0.0.1:18082')
    with (runtime/'go-server.log').open('ab') as log:
        p=subprocess.Popen([str(binary)],env=env,cwd=root,stdin=subprocess.DEVNULL,stdout=log,stderr=subprocess.STDOUT,start_new_session=True,close_fds=True)
    (runtime/'go-server.pid').write_text(str(p.pid)+'\n')
    for _ in range(50):
        current=health()
        if current:break
        if p.poll() is not None:raise SystemExit('Go restore failed; see qa-runtime/go-server.log')
        time.sleep(.2)
    else:raise SystemExit('Go readiness timed out')
    go_started=True
if current.get('contractVersion')!='v1.2-abort-reconciliation' or not current.get('ok'):raise SystemExit('Unexpected QA health/version: '+json.dumps(current))
print(json.dumps({'mysqlStarted':mysql_started,'goStarted':go_started,'dataDirectory':str(data),'databaseNetwork':'Unix socket only; TCP disabled','api':'http://127.0.0.1:18082','health':current,'existingDataReused':True,'initialized':False},indent=2))
