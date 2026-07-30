# Operations guide

## Install and update

The verified development target is Linux through Docker Compose. Install Docker with Compose support, check out a pinned commit, then run:

```bash
docker compose up -d dev
docker compose exec -T dev make test
docker compose exec -T dev go build -o bin/seo-auditor ./cmd/seo-auditor
docker compose exec -T dev go build -o bin/seo-auditor-cli ./cmd/seo-auditor-cli
docker compose exec -T dev go build -o bin/seo-auditor-mcp ./cmd/seo-auditor-mcp
```

Before updating, stop active crawls and make a managed backup. Check out the new pinned revision, rebuild the container, run the test suite, and start the application. Database migrations run transactionally on startup. Restore the pre-update database and previous binary if startup or integrity checks fail.

Native macOS and Windows builds are design targets, not runtime-qualified releases yet. Raw mode only needs the Go binary. Rendered mode also needs the pinned Node dependencies and Playwright Chromium runtime.

## Quick start

Start the local application with `go run ./cmd/seo-auditor`; it binds to loopback only. Use the browser UI, or create and run a profile with the CLI:

```bash
seo-auditor-cli project-create --name Example
seo-auditor-cli profile-create --project PROJECT_ID --name Default --url https://example.com/ --max-urls 10000
seo-auditor-cli crawl-start --project PROJECT_ID --profile PROFILE_ID
seo-auditor-cli crawl-status --crawl CRAWL_ID
seo-auditor-cli audit-summary --crawl CRAWL_ID
```

Profiles control the seed URL, host/subdomain scope, excluded path expressions, raw or rendered mode, response compression, concurrency, depth, URL ceiling, duration, response-size and retry budgets. Preview unfamiliar scope rules before starting. A project owns its profiles and crawl history.

The default response-compression mode requests gzip and enforces compressed, decoded-size and expansion-ratio limits. If a verified CDN incompatibility returns an erroneous response whenever `Accept-Encoding` is present, create the profile with `--response-compression disabled` or select “Disabled for incompatible CDNs” in the UI. Disabled mode omits the negotiation header but retains response-size, timeout, redirect, scope, DNS/IP, robots and disk controls. It is explicit and never retries or bypasses a genuine 403 response.

## CLI and reports

`seo-auditor-cli` emits JSON on stdout and errors as JSON on stderr. Run it without a command for the command list. Important read workflows include `crawl-status`, `crawl-timeline`, `audit-summary`, `page-list`, `page-get`, `issue-list`, and `issue-explain`. Lifecycle commands are `crawl-pause`, `crawl-resume`, and `crawl-cancel`. `report-export` supports CSV, NDJSON and XLSX managed artifacts. `diagnostic-create` writes a metadata-only support artifact; use `artifact-get` to inspect its managed path.

For bounded list mode, run `seo-auditor-cli crawl --urls 'https://example.com/a,https://example.org/b' --max-urls 10000`. List order is retained, normalized duplicates are removed, and the allowed-host set is derived from the supplied URLs. The local API accepts the equivalent `urls` array. A list is limited to 10,000 seeds and still obeys all target guards, robots rules, crawl limits and per-host politeness. Standalone mode also accepts `--response-compression gzip|disabled`.

## MCP

Configure an MCP client to execute `seo-auditor-mcp` over stdio. The server never opens an MCP network listener. Tools cover health, projects/profiles, crawl control and timeline, summaries, pages, links, issues, comparisons, reports and metadata-only diagnostics. Treat crawl-start and lifecycle tools as mutations requiring explicit operator intent. Keep protocol stdout reserved for MCP frames; logs belong on stderr. Complete configuration and agent workflow examples are in [MCP.md](./MCP.md).

## Storage, privacy and retention

The data directory contains SQLite/WAL state and the artifact directory contains reports, backups and diagnostics. Page URLs, extracted text metadata, headers, links and audit evidence can be sensitive. Data stays local unless the operator moves or shares it. Diagnostics exclude crawled URLs, content, headers and issue evidence by design, but should still be reviewed before sharing. Trash is recoverable until the operator removes its retained records. Establish an organization-specific retention period and securely remove expired database backups and reports.

## Backup, restore and troubleshooting

Create managed backups only after checking crawl status. Backups use SQLite's online backup API and are integrity-checked. Keep at least one copy outside the active data directory. Restoration is an offline operation: stop the app, preserve the failed database, replace it with a verified backup, then start the same or a schema-compatible binary.

For failures, inspect `crawl-status` and `crawl-timeline`, then create a diagnostic artifact. Common causes are robots denial, DNS answers containing prohibited addresses, scope exclusions, TLS failures, response budgets, renderer deadlines, or exhausted crawl limits. Never disable target guards to work around an error. Rendered-mode failures fall back to separately labelled raw evidence.

## Responsible crawling

Only crawl authorized targets. Identify the crawler honestly, honor robots policy, use low per-host concurrency, and schedule large audits away from peak traffic. A high URL ceiling is not permission to consume server capacity. Stop when a site shows instability or asks you to stop. Findings are technical observations, not guarantees about indexing, rankings, accessibility, or security.
