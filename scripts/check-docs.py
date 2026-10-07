#!/usr/bin/env python3
"""Check paired Markdown, navigation, local links and package file inventory."""
from pathlib import Path
import json
import re
import subprocess
import sys
from urllib.parse import unquote

root = Path(__file__).resolve().parents[1]
listed = subprocess.check_output(
    ['git', 'ls-files', '--cached', '--others', '--exclude-standard'],
    cwd=root, text=True).splitlines()
files = {name for name in listed if (root / name).is_file()}
markdown = sorted(name for name in files if name.endswith('.md'))
errors = []
for name in markdown:
    page = root / name
    english = page.with_name(page.name.replace('.zh-CN.md', '.md'))
    chinese = english.with_name(english.stem + '.zh-CN.md')
    header = (f'[English]({english.name}) | [简体中文]({chinese.name}) | '
              '[Website / 官网](https://dotapk.lol)')
    text = page.read_text(encoding='utf-8')
    if not english.is_file() or not chinese.is_file():
        errors.append(f'{name}: missing language partner')
    if not text.startswith(header + '\n\n'):
        errors.append(f'{name}: missing standard first-line navigation')
    if text.count('```') % 2:
        errors.append(f'{name}: unbalanced code fences')
    for target in re.findall(r'\[[^\]]*\]\(([^)]+)\)', text):
        target = target.strip().split(' "', 1)[0].strip('<>')
        if '://' in target or target.startswith(('#', 'mailto:')):
            continue
        path = unquote(target.split('#', 1)[0])
        if path and not (page.parent / path).exists():
            errors.append(f'{name}: broken relative link {target}')
    if re.search(r'/Users/[^/\s]+|/home/[^/\s]+', text):
        errors.append(f'{name}: private workstation path')
package = root / 'package.json'
if package.exists():
    data = json.loads(package.read_text())
    if data.get('name') == '@dotapk/heros':
        allowed = set(data.get('files', []))
        for name in markdown:
            if not name.startswith('.github/') and name not in allowed:
                errors.append(f'{name}: missing package files entry')
        for name in allowed:
            if not (root / name).exists():
                errors.append(f'{name}: package files entry does not exist')
if errors:
    print('\n'.join(errors), file=sys.stderr)
    raise SystemExit(1)
print(f'PASS {len(markdown)} Markdown files / {len(markdown)//2} language pairs, '
      'first-line navigation, relative links and package inventory')
