#!/usr/bin/env python3
"""Read-only grant/denial probes; never displays DSN/password."""
import os,pathlib,subprocess,json,re
# The known-existing unrelated table is supplied by the authorized operator.
# Validate before reading any credential or opening a connection.
denial_table=os.environ.get('DUEL_DENY_PROBE_TABLE','')
if not re.fullmatch(r'[A-Za-z_][A-Za-z0-9_]*\.[A-Za-z_][A-Za-z0-9_]*',denial_table) or denial_table.split('.')[0] in ('dota_duel','mysql','information_schema','performance_schema','sys'):
    raise SystemExit('Set DUEL_DENY_PROBE_TABLE to a known-existing unrelated schema.table')
quoted_denial_table='.'.join('`'+part+'`' for part in denial_table.split('.'))
raw=pathlib.Path('/etc/dota-duel/mysql-dsn').read_text().strip()
prefix='duel_app:';suffix='@tcp(127.0.0.1:3306)/dota_duel'
if not raw.startswith(prefix) or suffix not in raw:raise SystemExit('Unexpected dedicated credential format')
password=raw[len(prefix):raw.index(suffix)]
env=dict(os.environ,MYSQL_PWD=password)
def query(q):return subprocess.run(['mysql','--no-defaults','--protocol=TCP','--host=127.0.0.1','--user=duel_app','--database=dota_duel','--batch','--skip-column-names'],env=env,input=q,text=True,capture_output=True)
positive=query('SELECT CURRENT_USER(),VERSION(); SELECT version FROM duel_schema_migrations ORDER BY version; SHOW GRANTS FOR CURRENT_USER();')
if positive.returncode:raise SystemExit('Dedicated runtime login/read probe failed')
for sql in ['SELECT COUNT(*) FROM mysql.user;','SHOW CREATE TABLE '+quoted_denial_table+';']:
    if query(sql).returncode==0:raise SystemExit('Unexpected cross-schema permission; stop deployment')
print(positive.stdout.strip())
print(json.dumps({'mysqlSystemSchemaDenied':True,'crossSchemaTableDenied':True,'secretDisplayed':False}))
