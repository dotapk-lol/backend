#!/usr/bin/env python3
"""Start/resume isolated 46-hero QA. Never reset data, replace a listener, or touch 18082."""
import argparse
import json
import pathlib
import socket
import subprocess
import time
import urllib.request

parser = argparse.ArgumentParser()
parser.add_argument('--initialize', action='store_true', help='Allow first initialization in the fixed new 46-hero directory only')
parser.add_argument('--binary', help='Explicit candidate executable under this repository bin/; a running listener is never replaced')
args = parser.parse_args()
repo = pathlib.Path(__file__).resolve().parent.parent
base = pathlib.Path('/tmp/duel-qa-candidate46-20261002')
data, sock = base / 'data', base / 'mysql.sock'
runtime = repo.parent / 'qa-runtime' / 'candidate46'
runtime.mkdir(parents=True, exist_ok=True)
state_file = runtime / 'processes.json'
prior_state = json.loads(state_file.read_text()) if state_file.is_file() else {}
binary = pathlib.Path(args.binary or prior_state.get('binary') or repo / 'bin' / 'dueld-v13-candidate46-local').resolve()
if binary.parent != (repo / 'bin').resolve():
    raise SystemExit('Candidate executable must belong to this repository bin directory')
if not binary.is_file():
    raise SystemExit('Build bin/dueld-v13-candidate46-local first')
mysql = '/opt/homebrew/opt/mysql/bin/mysql'
admin = '/opt/homebrew/opt/mysql/bin/mysqladmin'
mysqld = '/opt/homebrew/opt/mysql/bin/mysqld'
origin = 'http://127.0.0.1:4185'
def ping():
    return subprocess.run([admin, '--no-defaults', '--socket='+str(sock), '-uroot', 'ping'], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL).returncode == 0
def sql(statement=None, file=None):
    command = [mysql, '--no-defaults', '--socket='+str(sock), '-uroot', '--batch', '--skip-column-names']
    return subprocess.run(command, input=(file.read_text() if file else statement), text=True, check=True, capture_output=True).stdout.strip()
def health():
    try:
        with urllib.request.urlopen('http://127.0.0.1:18084/healthz', timeout=1) as response:
            return json.load(response)
    except Exception:
        return None
with socket.socket() as check:
    listening = check.connect_ex(('127.0.0.1', 18084)) == 0
if listening:
    previous = runtime / 'processes.json'
    current = health()
    if previous.is_file() and current and current.get('contractVersion') == 'v1.3-gameplay-rosters' and prior_state.get('binary') == str(binary):
        print(json.dumps({'alreadyRunning': True, 'health': current, 'state': str(previous)}))
        raise SystemExit(0)
    raise SystemExit('Port18084 is occupied; refusing to replace a listener')
initialized = False
if not (data / 'mysql.ibd').is_file():
    if not args.initialize or data.exists():
        raise SystemExit('Missing 46-hero data: explicit --initialize and a nonexistent data directory are required; never reset')
    base.mkdir(mode=0o700, parents=True, exist_ok=True)
    subprocess.run([mysqld, '--no-defaults', '--initialize-insecure', '--datadir='+str(data), '--lower-case-table-names=2', '--log-error='+str(base/'initialize.log')], check=True)
    initialized = True
mysql_started = False
if not ping():
    with (base/'launcher.log').open('ab') as log:
        subprocess.Popen([mysqld, '--no-defaults', '--datadir='+str(data), '--socket='+str(sock), '--pid-file='+str(base/'mysql.pid'), '--log-error='+str(base/'server.log'), '--skip-networking', '--mysqlx=OFF', '--lower-case-table-names=2', '--innodb-buffer-pool-size=128M', '--max-connections=20'], stdin=subprocess.DEVNULL, stdout=log, stderr=log, start_new_session=True)
    mysql_started = True
    for _ in range(50):
        if ping(): break
        time.sleep(.1)
    else: raise SystemExit('46-hero MySQL did not start; inspect its dedicated log')
sql('CREATE DATABASE IF NOT EXISTS dota_duel CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;')
sql('CREATE TABLE IF NOT EXISTS dota_duel.duel_schema_migrations (version INT PRIMARY KEY, applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP);')
versions = set(sql('SELECT version FROM dota_duel.duel_schema_migrations;').splitlines())
for number, filename in enumerate(['001_init.sql', '002_analytics.sql', '003_local_pvp_analytics.sql', '004_hero_registry.sql'], 1):
    if str(number) not in versions:
        sql(file=repo/'migrations'/filename)
# This SQL is only for this isolated instance; never copied to production migrations.
roster_count = int(sql("SELECT COUNT(*) FROM dota_duel.duel_gameplay_rosters WHERE roster_id='arena-first22-46-v1';"))
if roster_count == 0:
    sql(file=runtime/'local-register.sql')
expected_rosters = json.loads((runtime/'gameplay-rosters.json').read_text())
for item in expected_rosters:
    rid = item['rosterId']
    # Names originate only in the task-owned frozen manifest, never an HTTP request.
    import re
    if not re.fullmatch(r'[a-zA-Z0-9._-]{1,100}', rid):
        raise SystemExit('Invalid roster identifier')
    versions = json.loads(sql("SELECT game_versions FROM dota_duel.duel_gameplay_rosters WHERE roster_id='"+rid+"';"))
    members = [int(x) for x in sql("SELECT hero_id FROM dota_duel.duel_roster_heroes WHERE roster_id='"+rid+"' ORDER BY hero_id;").splitlines()]
    if versions != item['gameVersions'] or members != sorted(item['heroIds']):
        raise SystemExit('Stored candidate roster differs; refusing to overwrite')
import os
env = os.environ.copy()
env.pop('DUEL_MYSQL_DSN_FILE', None)
env.update(DUEL_MYSQL_DSN='root@unix('+str(sock)+')/dota_duel', DUEL_LISTEN='127.0.0.1:18084', DUEL_ALLOWED_ORIGIN=origin, DUEL_TRUSTED_PROXY_IP='')
with (runtime/'go.log').open('ab') as log:
    process = subprocess.Popen([str(binary)], env=env, stdin=subprocess.DEVNULL, stdout=log, stderr=log, start_new_session=True)
for _ in range(50):
    current = health()
    if current: break
    if process.poll() is not None: raise SystemExit('46-hero Go exited; inspect its dedicated log')
    time.sleep(.1)
else: raise SystemExit('46-hero Go health unavailable; inspect dedicated log before retry')
if current.get('contractVersion') != 'v1.3-gameplay-rosters':
    raise SystemExit('Unexpected backend contract')
state = {'goPid': process.pid, 'mysqlPid': int((base/'mysql.pid').read_text()), 'baseUrl':'http://127.0.0.1:18084/api/v1', 'corsOrigin':origin, 'mysqlSocket':str(sock), 'mysqlData':str(data), 'binary':str(binary), 'initialized':initialized, 'mysqlStarted':mysql_started, 'health':current, 'tcpDatabase':False, 'productionChanged':False}
(runtime/'processes.json').write_text(json.dumps(state, indent=2)+'\n')
print(json.dumps(state, indent=2))
