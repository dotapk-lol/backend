#!/usr/bin/env python3
"""Reproduce the pinned production binary locally; never connects to a database/server."""
import argparse
import hashlib
import json
import os
import pathlib
import subprocess
import tempfile

SOURCE = 'aa04e01f25ef2226b213cfc64e68c6c96fd0e18c'
MANIFEST_SHA = 'b7df50b06e88466580d19c6a0beae241d78ef0386d72c847141eeea3642f79be'
BINARY_SHA = '9962a885ccd42ee8cc8fbd878bef1724b552c35906c8daaca71f4f2536e3c0d3'
repo = pathlib.Path(__file__).resolve().parent.parent
parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--output', type=pathlib.Path, default=repo/'bin/dueld-production-v13-linux-amd64')
args = parser.parse_args()
manifest = repo/'deploy/production-v13/gameplay-rosters.json'
if hashlib.sha256(manifest.read_bytes()).hexdigest() != MANIFEST_SHA:
    raise SystemExit('Pinned production manifest differs; do not enable another roster/build')
version = subprocess.check_output(['go','version'],text=True).strip()
if version.split()[2] != 'go1.26.1':
    raise SystemExit('Exact reproduction requires Go1.26.1; production is not changed')
output = args.output.resolve()
output.parent.mkdir(parents=True, exist_ok=True)
source_paths = subprocess.check_output(['git','-C',str(repo),'ls-tree','-r','--name-only',SOURCE,'--','go.mod','go.sum','cmd','internal'],text=True).splitlines()
for name in source_paths:
    pinned=subprocess.check_output(['git','-C',str(repo),'show',SOURCE+':'+name])
    if not (repo/name).is_file() or (repo/name).read_bytes()!=pinned:
        raise SystemExit('Working source differs from pinned production snapshot: '+name)
expected_go={name for name in source_paths if name.endswith('.go')}
actual_go={str(p.relative_to(repo)) for directory in ['cmd','internal'] for p in (repo/directory).rglob('*.go')}
if actual_go!=expected_go:
    raise SystemExit('Unexpected Go source files outside pinned snapshot')
with tempfile.TemporaryDirectory(prefix='duel-production-v13-build-',dir='/tmp') as tmp:
    overlay = pathlib.Path(tmp)/'overlay.json'
    overlay.write_text(json.dumps({'Replace':{str(repo/'internal/duel/registry/gameplay-rosters.json'):str(manifest)}}))
    env=dict(os.environ,CGO_ENABLED='0',GOOS='linux',GOARCH='amd64',GOPROXY='off',GOFLAGS='',GOMAXPROCS='2')
    env.setdefault('GOCACHE',str(pathlib.Path(tmp)/'cache'))
    subprocess.run(['go','build','-p=1','-buildvcs=false','-overlay',str(overlay),'-trimpath','-o',str(output),'./cmd/dueld'],cwd=repo,env=env,check=True)
actual=hashlib.sha256(output.read_bytes()).hexdigest()
if actual != BINARY_SHA:
    raise SystemExit('Reproduced binary differs from production: '+actual)
print(json.dumps({'result':'PASS','sourceCommit':SOURCE,'manifestSha256':MANIFEST_SHA,'binarySha256':actual,'output':str(output),'productionChanged':False,'databaseAccess':False}))
