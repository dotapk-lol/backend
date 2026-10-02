#!/usr/bin/env python3
"""Approved one-time DotaPK identity provisioning on the selected server.
Requires existing local MySQL administrator authentication. Never prints secrets.
Refuses to overwrite an existing identity or credential file.
"""
import os,pathlib,secrets,subprocess,sys

def query(sql):
    p=subprocess.run(['mysql','--batch','--skip-column-names'],input=sql,text=True,capture_output=True)
    if p.returncode:raise SystemExit('MySQL administrative operation failed; inspect locally without revealing secret input')
    return p.stdout.strip()
if os.geteuid()!=0:raise SystemExit('Run as the authorized host administrator')
if query("SELECT @@global.general_log;")!='0':raise SystemExit('General query logging is enabled; stop for review before handling a new secret')
if query("SELECT COUNT(*) FROM information_schema.SCHEMATA WHERE SCHEMA_NAME='dota_duel';")!='1':raise SystemExit('Dedicated dota_duel schema must already exist')
if query("SELECT COUNT(*) FROM mysql.user WHERE User='duel_app';")!='0':raise SystemExit('duel_app already exists; inspect grants and stored configuration without overwriting')
folder=pathlib.Path('/etc/dota-duel')
folder.mkdir(mode=0o700,exist_ok=True)
if folder.stat().st_uid!=0 or folder.stat().st_mode & 0o077:raise SystemExit('Credential directory must be root-owned mode0700')
password=secrets.token_urlsafe(32)
path=folder/'mysql-dsn'
fd=os.open(path,os.O_WRONLY|os.O_CREAT|os.O_EXCL,0o600)
with os.fdopen(fd,'w') as f:f.write('duel_app:'+password+'@tcp(127.0.0.1:3306)/dota_duel?charset=utf8mb4&parseTime=true&loc=UTC\n')
# Token alphabet has no SQL quote/backslash; input is private stdin, not argv.
query("CREATE USER 'duel_app'@'127.0.0.1' IDENTIFIED BY '"+password+"'; GRANT SELECT,INSERT,UPDATE,DELETE ON dota_duel.* TO 'duel_app'@'127.0.0.1';")
print(query("SHOW GRANTS FOR 'duel_app'@'127.0.0.1';"))
print('Dedicated credential stored root:root0600; secret not displayed')
