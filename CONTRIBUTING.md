[English](CONTRIBUTING.md) | [简体中文](CONTRIBUTING.zh-CN.md) | [Website / 官网](https://dotapk.lol)

# Contributing

Start from current main in a separate branch and read [API](docs/API.md), [development](docs/DEVELOPMENT.md) and [architecture](docs/ARCHITECTURE.md). Explain the trigger, resulting behavior and actual validation.

- Update English/Chinese Markdown together, with language/website links first and valid relative links. Do not translate/duplicate schemas, profile data or the standard English MIT legal text.
- Preserve strict shapes, anonymous authorization, idempotency/immutable reports and historical wire identity. Never promote client_reported into peer_agreement.
- Keep fixed registry allocations and explicit embed/SQL build bindings. Catalog presence is not playable status; current frontend22 remains unchanged unless separately accepted.
- Explain schema/view changes: migration, historical compatibility, grain/denominator and excluded anomalies. Do not add a database to evade existing MySQL definitions.
- Run affected serial tests; destructive integration only uses dedicated disposable allowlisted sockets, never production/retained QA or a shared parallel test instance.
- Do not commit DSNs, tokens, player/match identifiers, IP/room codes, raw reports, traceable receipts or backups. Manual public aggregates follow heros/balance-data policy.
- Project-owned code/docs use [MIT](LICENSE). Preserve notices and third-party licenses; do not introduce unauthorized media.

Check `git diff --check`, paired docs/links and relevant tests. Deployment/security/DB/CI privilege changes need explicit scope; coordinate concurrency and never force-push over history. Prose cleanup needs no production operation or deployment.
