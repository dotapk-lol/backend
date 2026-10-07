#!/usr/bin/env python3
"""Build current source with explicit roster/protocol overlays; no DB/network access."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess
import tempfile

repo = Path(__file__).resolve().parent.parent
parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--rosters', type=Path, required=True)
parser.add_argument('--features', type=Path, required=True)
parser.add_argument('--output', type=Path, required=True)
parser.add_argument('--goos', default=subprocess.check_output(['go', 'env', 'GOOS'], text=True).strip())
parser.add_argument('--goarch', default=subprocess.check_output(['go', 'env', 'GOARCH'], text=True).strip())
args = parser.parse_args()
rosters, features = args.rosters.resolve(), args.features.resolve()
roster_data, feature_data = json.loads(rosters.read_text()), json.loads(features.read_text())
if set(feature_data) != {'roomSelectionVersions'} or not isinstance(feature_data['roomSelectionVersions'], list):
    raise SystemExit('Expected explicit roomSelectionVersions list')
versions = feature_data['roomSelectionVersions']
bindings = {}
for roster in roster_data:
    for version in roster['gameVersions']:
        if version in bindings:
            raise SystemExit('Duplicate roster version binding')
        bindings[version] = roster
if len(set(versions)) != len(versions) or any(version not in bindings or 1 not in bindings[version]['heroIds'] for version in versions):
    raise SystemExit('Selection versions must uniquely bind a roster containing default hero1')
output = args.output.resolve()
output.parent.mkdir(parents=True, exist_ok=True)
with tempfile.TemporaryDirectory(prefix='duel-profile-build-', dir='/tmp') as temporary:
    overlay = Path(temporary)/'overlay.json'
    overlay.write_text(json.dumps({'Replace': {
        str(repo/'internal/duel/registry/gameplay-rosters.json'): str(rosters),
        str(repo/'internal/duel/registry/protocol-features.json'): str(features),
    }}))
    env = dict(os.environ, CGO_ENABLED='0', GOOS=args.goos, GOARCH=args.goarch,
               GOPROXY='off', GOFLAGS='', GOMAXPROCS='2')
    env.setdefault('GOCACHE', str(Path(temporary)/'cache'))
    subprocess.run(['go', 'build', '-p=1', '-buildvcs=false', '-overlay', str(overlay),
                    '-trimpath', '-o', str(output), './cmd/dueld'], cwd=repo, env=env, check=True)
print(json.dumps({'result': 'PASS', 'target': args.goos+'/'+args.goarch,
                  'rostersSha256': hashlib.sha256(rosters.read_bytes()).hexdigest(),
                  'featuresSha256': hashlib.sha256(features.read_bytes()).hexdigest(),
                  'binarySha256': hashlib.sha256(output.read_bytes()).hexdigest(),
                  'databaseAccess': False, 'productionChanged': False}))
