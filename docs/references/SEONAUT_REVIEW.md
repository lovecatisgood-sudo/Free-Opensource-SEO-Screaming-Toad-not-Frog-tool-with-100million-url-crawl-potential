# Reference review: SEonaut

Repository: `StJudeWasHere/seonaut`  
Reviewed commit: `880b312c28fab8b0bf7fe4f9449dc4746dbb82ff`  
Review date: 2026-07-30  
License: MIT

## Executive assessment

SEonaut is a credible, mature open-source SEO auditing application and a valuable reference for SEO Auditor. It is substantially more established than the previously reviewed `open-seo-crawler`: development dates to 2022, the repository has hundreds of commits, multiple meaningful contributors, active cross-platform tests, a race-enabled Go test suite, a container build, and a hosted deployment.

It validates Go as a practical language for an SEO crawler. Its issue taxonomy, report queries, HTML parsing fixtures, crawl lifecycle, translations, exports, archive support, and migration history are especially useful sources of product and test ideas.

It is not a safe or scalable foundation to fork unchanged for SEO Auditor. Its security and persistence model targets a conventional authenticated web application, not our local-first, agent-controlled threat model. We should reuse selected MIT-licensed ideas and independently reviewed code, not inherit the architecture wholesale.

## Repository maturity signals

At review time:

- public MIT-licensed repository;
- created in March 2022;
- approximately 900 commits in GitHub history;
- approximately 745 stars and 122 forks;
- multiple contributors, with two major contributors;
- active test and container-build workflows;
- CI runs Go tests with the race detector on Linux, macOS and Windows;
- latest reviewed commit had successful test/build checks and a verified merge commit;
- 38 Go test files and roughly 6,100 lines of test code in the shallow checkout;
- no tagged GitHub releases at review time;
- no visible branch protection through the unauthenticated API;
- no published repository security advisories returned by the API.

These are positive engineering signals, not a guarantee of security.

## What SEonaut does well

### Product coverage

SEonaut includes projects and users, live crawl progress, dashboards, filtered issues, URL exploration, CSV/sitemap/resource exports, WARC-derived archive/replay support, basic-auth crawling, external-link checks, translations, crawl history and Docker deployment.

Its rule catalog covers many normal-site audit needs:

- response errors and redirects;
- titles and descriptions;
- headings and content volume;
- indexability and robots directives;
- canonical relationships;
- hreflang validation;
- sitemap consistency and orphan candidates;
- internal/external links;
- images and alt/size checks;
- depth, duplicate content and dead ends;
- security headers, forms, viewport, DOM size and URL hygiene.

### Engineering patterns worth learning from

- Go packages separate crawler, services, repositories, routes, models and issue reporters.
- Goroutines and channels provide a concise worker/queue model.
- `context` cancellation bounds overall crawl duration.
- The HTTP client disables automatic redirect following so redirect handling remains visible to the crawler.
- Page-level and multi-page rules are separated.
- Migrations document the evolution of the report schema and rule catalog.
- Repository queries power server-side views and exports rather than sending the complete dataset to the browser.
- Tests cover parser behaviour, rule boundaries, queues, robots, services and regressions.
- CI uses `-race` across three operating systems.
- Templates and minimal JavaScript demonstrate that a useful crawler UI need not become a large SPA.

## Ideas we should adopt

Adopt at the behaviour or design level:

1. Separate page-local rules from crawl/graph aggregate rules.
2. Treat issue types as stable records that support efficient queries and historical reporting.
3. Preserve redirect responses rather than allowing the standard client to hide them.
4. Capture TTFB and response metadata during the same guarded request.
5. Maintain a deterministic parser fixture corpus.
6. Run the Go race detector in CI across supported platforms.
7. Use server-side pagination and indexed report queries.
8. Support translations from structured message catalogs.
9. Offer an optional web archive only behind explicit storage budgets.
10. Model external-link status separately from internal traversal.
11. Study its 70-plus migrations as a catalog of real SEO product evolution and edge cases.
12. Preserve a simple, evidence-led interface before adding elaborate visualisations.

Potential code reuse must be assessed function by function, retain the MIT notice, enter `THIRD_PARTY_NOTICES.md`, and receive new security and conformance tests.

## What we should not copy

### Network safety

SEonaut's project validation checks the URL scheme but does not centrally reject loopback, private, link-local, reserved or cloud-metadata targets. Its standard `http.Client` uses the default transport without validated-IP connection pinning. External-link and sitemap requests can also originate from discovered content. This is insufficient for an MCP-controlled or exposed crawler and creates SSRF risk.

SEO Auditor must enforce one target-policy gateway for seeds, redirects, robots, sitemaps, external checks and renderer subresources.

### Binding and web security defaults

The default configuration binds to `0.0.0.0` and documents unencrypted HTTP unless a reverse proxy is added. Several state mutations, including crawl start/stop and project deletion, use GET routes. No CSRF-token framework was found in the inspected route/forms. Session keys are randomly generated at process startup, so restarts invalidate all sessions; cookie options set `HttpOnly` but do not centrally set `SameSite` and set `Secure` only in a crawl-related cookie path.

SEO Auditor will bind to loopback, use POST/DELETE for mutations, require a local session and CSRF/Origin checks, and centralise hardened cookie/security headers.

### Frontier and scaling model

SEonaut stores the queue and visited URL set in memory. It uses two fixed consumers, a random delay up to 1.5 seconds, a fixed 20,000-page service limit, and a two-hour context timeout. Queue state is not a durable frontier, so process recovery and 100,000-URL memory behaviour do not meet our goals.

SEO Auditor will persist frontier state in SQLite with leases, explicit per-host scheduling, adaptive backoff, checkpoints, pause/resume and crash recovery.

### Storage and deployment

SEonaut requires MySQL and recommends Docker Compose. Its default example credentials are intentionally simple and intended for a private Docker network, but the installation is heavier than a local audit tool needs. The application container runs without an explicit non-root user, healthcheck, read-only filesystem or dropped capabilities in the reviewed Dockerfile/Compose files.

SEO Auditor will use SQLite for zero-service local installation and ship a Go binary. Containers, if offered, will be hardened separately.

### Sitemap implementation

One sitemap-index helper is invoked by URL rather than through the crawler's injected HTTP client. That pattern can bypass consistent timeouts, target policy, authentication and test doubles. We will own all sitemap I/O and use a streaming parser over already-guarded response bodies.

### Authentication and secrets

SEonaut supports multiple users and basic-auth crawls, but its minimum password check accepts more than one character. Basic-auth credentials live in process memory for a crawl. Our first release deliberately excludes authenticated crawling; a future version requires a dedicated secret store and stronger policy.

### Release and supply chain

The repository had no tags/releases at review time. Container workflows publish moving branch/latest images, and Docker build stages download a helper binary without an explicit checksum verification in the Dockerfile. SEO Auditor will use tagged immutable releases, checksums/SBOMs, reviewed dependencies and user-initiated updates.

## Architectural lessons for SEO Auditor

SEonaut proves that Go can express the core crawler clearly, but it also shows where a simple in-memory Go crawler reaches its limits. Our differentiator should not merely be “also written in Go.” It should be:

- a durable SQLite frontier instead of an in-memory queue;
- one comprehensive outbound network-security policy;
- explicit crawl completeness and recovery states;
- raw and safely rendered crawl modes;
- versioned evidence-rich rules;
- API, CLI and MCP parity;
- reproducible signed releases and local-first defaults;
- performance measured at 10k/100k URLs rather than assumed.

## Verification limitations

The source was inspected from an inert shallow clone in `/tmp`. No container or application was started. The local environment did not have the Go compiler installed, so tests were not rerun locally; the reviewed GitHub commit reported successful race-enabled test/build checks across Linux, macOS and Windows. Dependency-vulnerability status was not independently reproduced with `govulncheck` during this review.
