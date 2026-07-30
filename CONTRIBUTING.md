# Contributing to SEO Screaming Toad

Thanks for helping build an open-source technical SEO crawler that practitioners can inspect, automate, and improve together.

## Good first contributions

- Reproduce and document a crawler or audit-rule bug.
- Add fixtures for unusual HTML, redirects, robots, sitemaps, canonicals, hreflang, or structured data.
- Improve accessibility or dense-table usability in the dashboard.
- Test clean installation on Linux, macOS, or Windows.
- Improve MCP client examples and agent workflows.
- Propose a rule only when its evidence, remediation, false-positive risks, and limitations can be tested.

## Before opening an issue

Search existing issues first. Remove credentials, private URLs, client names, proprietary content, cookies, and crawl databases from all examples. Security problems belong in private vulnerability reporting as described in [SECURITY.md](SECURITY.md).

## Development workflow

The recommended sandbox runs as an unprivileged container user:

```bash
docker compose up -d dev
docker compose exec dev bash
make test
```

Useful checks:

```bash
make test
make lint
make test-offline
pnpm typecheck
pnpm test
pnpm build
```

Read [DEVELOPMENT.md](DEVELOPMENT.md) for prerequisites and isolation details.

## Pull requests

1. Keep each pull request focused on one problem.
2. Add or update tests for observable behavior.
3. Document new configuration, MCP tools, audit rules, and limitations.
4. Preserve the guarded networking and local-data boundaries.
5. Run the relevant Go and TypeScript checks.
6. Explain user impact and any remaining limitations in the pull request.

Audit-rule changes must keep findings versioned and evidence-backed. Network changes must test redirects, prohibited addresses, DNS changes, response budgets, cancellation, and error behavior where relevant. Scale changes must distinguish synthetic evidence from live-network results.

## Design principles

- Safety limits are product behavior, not optional obstacles.
- Raw and rendered evidence remain distinguishable.
- Claims must match reproducible evidence.
- MCP mutations must be explicit and bounded.
- Local data does not leave the machine without an operator-requested action.
- Compatibility matters more than cleverness.

By participating, you agree to follow the [Code of Conduct](CODE_OF_CONDUCT.md).

