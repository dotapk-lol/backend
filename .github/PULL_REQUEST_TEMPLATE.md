[English](PULL_REQUEST_TEMPLATE.md) | [简体中文](PULL_REQUEST_TEMPLATE.zh-CN.md) | [Website / 官网](https://dotapk.lol)

# Pull request

## Problem and resulting behavior

Describe the concrete trigger and what changes for the user/developer.

## Validation

List actual commands and what they prove; distinguish source/module tests from real browser/network acceptance.

- [ ] English/Chinese guides updated together; language/website headers and relative links checked.
- [ ] `python3 scripts/check-docs.py` and `git diff --check` pass.
- [ ] Applicable focused checks completed serially; untested limits stated.
- [ ] Stable IDs, released22/88 and paused status preserved, or a separate accepted scope explains the change.
- [ ] No credentials, private paths, per-player/game reports, traceable QA data or unauthorized media.
- [ ] Project MIT and third-party rights preserved; no implicit deployment/DB/CI privilege change.

See [contributing](../CONTRIBUTING.md). Standard LICENSE remains English; Chinese guides explain scope without replacing legal terms.
