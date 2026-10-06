#!/usr/bin/env python3
"""Serial, in-memory production-profile boundary tests; no database/network."""
import json,os,pathlib,subprocess,tempfile
repo=pathlib.Path(__file__).resolve().parent.parent
with tempfile.TemporaryDirectory(prefix='duel-profile-tests-',dir='/tmp') as tmp:
 overlay=pathlib.Path(tmp)/'overlay.json'
 overlay.write_text(json.dumps({'Replace':{
  str(repo/'internal/duel/registry/gameplay-rosters.json'):str(repo/'deploy/production-v13/gameplay-rosters.json'),
  str(repo/'internal/duel/heros22_profile_test.go'):str(repo/'deploy/production-v13/heros22_profile_test.go')}}))
 env=dict(os.environ,GOPROXY='off',GOMAXPROCS='2',GOFLAGS='')
 env.setdefault('GOCACHE',str(pathlib.Path(tmp)/'cache'))
 subprocess.run(['go','test','-p=1','-count=1','-overlay',str(overlay),'-run','^(TestHeros22ProductionBoundary|TestLegacyRequestBytes|TestRegistryMemory|TestRosterHTTPContract)$','./internal/duel'],cwd=repo,env=env,check=True)
