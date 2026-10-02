#!/usr/bin/env python3
"""Start/resume isolated B0 QA. Never reset data, replace a listener, or touch 18082."""
import argparse
import json
import pathlib
import socket
import subprocess
import time
import urllib.request

parser = argparse.ArgumentParser()
parser.add_argument('--initialize', action='store_true', help='Allow first initialization in the fixed new B0 directory only')
args = parser.parse_args()
repo = pathlib.Path(__file__).resolve().parent.parent
base = pathlib.Path('/tmp/duel-mysql-test-registry-b0-20261002')
data, sock = base / 'data', base / 'mysql.sock'
runtime = repo.parent / 'qa-runtime' / 'registry-b0'
runtime.mkdir(parents=True, exist_ok=True)
binary = repo / 'bin' / 'dueld-v13-registry-b0'
if not binary.is_file():
    raise SystemExit('Build bin/dueld-v13-registry-b0 first')
mysql = '/opt/homebrew/opt/mysql/bin/mysql'
admin = '/opt/homebrew/opt/mysql/bin/mysqladmin'
mysqld = '/opt/homebrew/opt/mysql/bin/mysqld'
origin = 'http://127.0.0.1:4174'
def ping():
    return subprocess.run([admin, '--no-defaults', '--socket='+str(sock), '-uroot', 'ping'], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL).returncode == 0
def sql(statement=None, file=None):
    command = [mysql, '--no-defaults', '--socket='+str(sock), '-uroot', '--batch', '--skip-column-names']
    return subprocess.run(command, input=(file.read_text() if file else statement), text=True, check=True, capture_output=True).stdout.strip()
def health():
    try:
        with urllib.request.urlopen('http://127.0.0.1:18083/healthz', timeout=1) as response:
            return json.load(response)
    except Exception:
        return None
with socket.socket() as check:
    listening = check.connect_ex(('127.0.0.1', 18083)) == 0
if listening:
    previous = runtime / 'processes.json'
    current = health()
    if previous.is_file() and current and current.get('contractVersion') == 'v1.3-gameplay-rosters':
        print(json.dumps({'alreadyRunning': True, 'health': current, 'state': str(previous)}))
        raise SystemExit(0)
    raise SystemExit('Port18083 is occupied; refusing to replace a listener')
initialized = False
if not (data / 'mysql.ibd').is_file():
    if not args.initialize or data.exists():
        raise SystemExit('Missing B0 data: explicit --initialize and a nonexistent data directory are required; never reset')
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
    else: raise SystemExit('B0 MySQL did not start; inspect its dedicated log')
sql('CREATE DATABASE IF NOT EXISTS dota_duel CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;')
sql('CREATE TABLE IF NOT EXISTS dota_duel.duel_schema_migrations (version INT PRIMARY KEY, applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP);')
versions = set(sql('SELECT version FROM dota_duel.duel_schema_migrations;').splitlines())
for number, filename in enumerate(['001_init.sql', '002_analytics.sql', '003_local_pvp_analytics.sql', '004_hero_registry.sql'], 1):
    if str(number) not in versions:
        sql(file=repo/'migrations'/filename)
import os
env = os.environ.copy()
env.pop('DUEL_MYSQL_DSN_FILE', None)
env.update(DUEL_MYSQL_DSN='root@unix('+str(sock)+')/dota_duel', DUEL_LISTEN='127.0.0.1:18083', DUEL_ALLOWED_ORIGIN=origin, DUEL_TRUSTED_PROXY_IP='')
with (runtime/'go.log').open('ab') as log:
    process = subprocess.Popen([str(binary)], env=env, stdin=subprocess.DEVNULL, stdout=log, stderr=log, start_new_session=True)
for _ in range(50):
    current = health()
    if current: break
    if process.poll() is not None: raise SystemExit('B0 Go exited; inspect its dedicated log')
    time.sleep(.1)
else: raise SystemExit('B0 Go health unavailable; inspect dedicated log before retry')
if current.get('contractVersion') != 'v1.3-gameplay-rosters':
    raise SystemExit('Unexpected backend contract')
state = {'goPid': process.pid, 'mysqlPid': int((base/'mysql.pid').read_text()), 'baseUrl':'http://127.0.0.1:18083/api/v1', 'corsOrigin':origin, 'mysqlSocket':str(sock), 'mysqlData':str(data), 'binary':str(binary), 'initialized':initialized, 'mysqlStarted':mysql_started, 'health':current, 'tcpDatabase':False, 'productionChanged':False}
(runtime/'processes.json').write_text(json.dumps(state, indent=2)+'\n')
print(json.dumps(state, indent=2))
