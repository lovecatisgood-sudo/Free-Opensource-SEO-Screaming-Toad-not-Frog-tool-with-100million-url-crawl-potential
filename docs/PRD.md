# Product requirements document: SEO Auditor

Status: Final product baseline  
Version: 2.0  
Date: 2026-07-30  
Owner: Product and engineering  
Product codename: SEO Auditor

Public product name: SEO Screaming Toad — Not Frog
Alternate community name: DJAI Toad

## 1. Executive summary

SEO Auditor is a secure, local-first website crawler and technical SEO audit application. It gives site owners, SEO professionals, and developers the recurring crawl, diagnostics, comparison, and export workflows they use most often without sending crawl data to a third-party service. It includes a local web interface, a scriptable CLI, and a constrained Model Context Protocol (MCP) server so an AI agent can run and interpret audits through explicit tools.

The first production release will provide dependable technical audits for typical small and medium sites and bounded large-site crawls. The initial scale target is 10 to 100,000 discovered URLs per crawl on suitable hardware. The product does not promise immediate feature parity with mature enterprise crawlers or million-URL operation.

Security is part of the product contract. The default installation listens only on loopback, MCP uses stdio, arbitrary private-network targets are blocked, redirects are revalidated, TLS errors are reported rather than bypassed, crawl limits are mandatory, and no automatic upstream code execution or self-update endpoint exists.

## 2. Problem statement

Technical SEO audits require collecting and relating data across HTTP responses, rendered and raw HTML, internal links, directives, canonical tags, sitemaps, structured data, and repeated crawls. Existing tools can be expensive, cloud-dependent, difficult to automate safely, or too broad for teams that need a focused local workflow.

Users need a crawler that:

- produces trustworthy evidence rather than opaque scores;
- remains responsive and bounded on malformed or adversarial sites;
- preserves crawl data locally for comparison;
- supports raw HTML and optional JavaScript-rendered pages;
- exports results in common formats;
- can be controlled safely by users, scripts, and MCP clients;
- explains findings without claiming that every heuristic is universally correct.

## 3. Product vision

Make a technical site audit as repeatable and inspectable as running a test suite: configure a bounded scope, crawl, review evidence, export or compare, fix the site, and crawl again.

## 4. Goals

### 4.1 Product goals

1. Cover the highest-value technical SEO workflows used in routine audits.
2. Make each issue traceable to URLs, captured evidence, and an explicit rule version.
3. Support raw HTTP crawling and isolated JavaScript rendering.
4. Persist crawls so users can reopen and compare them.
5. Provide UI, CLI, JSON, and MCP access to the same domain logic.
6. Be safe to install and run on a personal workstation by default.
7. Remain useful without accounts, API keys, telemetry, or cloud services.
8. Respect websites through robots controls, rate limits, transparent user agents, and bounded concurrency.
9. Keep the architecture theoretically extensible to segmented campaigns above 100 million unique URLs without making that an implemented, supported or release-blocking capacity claim.

### 4.2 Initial success metrics

- At least 95% of a curated 500-case conformance fixture set produces the expected extraction and issue results before public beta.
- A 10,000-URL static site crawl completes without unbounded memory growth, deadlock, or database corruption.
- No high- or critical-severity findings in the release security review.
- Every finding returned by the UI, CLI, or MCP includes a stable rule ID and affected URL count.
- A user can install, start a crawl, inspect the summary, and export issues without external credentials.
- Interrupted crawls recover to an explicit resumable or terminal state; they never remain silently “running.”
- Repeated crawls with identical fixtures and configuration produce semantically identical results.

### 4.3 Theoretical scale direction

SEO Auditor's segmentation, durable identity, coordinator and immutable-result designs could theoretically be extended to crawl campaigns exceeding 100 million unique URLs. This is an architectural hypothesis, not a v2.0 requirement, supported limit, delivery commitment or performance guarantee.

Public language must say that the design has theoretical potential beyond 100 million URLs and that the capacity is untested and unsupported. The project does not claim that one campaign can currently complete that volume. Any future capacity claim would require a separately approved PRD and evidence program.

## 5. Non-goals for version 1

- Complete feature parity with Screaming Frog, Sitebulb, Ahrefs, or Semrush.
- Guaranteed million-URL crawling.
- Advertising a tested, verified, supported or guaranteed 100M+ capacity.
- A hosted multi-tenant SaaS.
- Rank tracking, backlink indexes, or keyword-volume databases.
- Automatic modification of audited websites.
- Automatic deployment of recommendations.
- Executing arbitrary user-provided JavaScript during a crawl.
- Login-form automation or crawling authenticated applications in the initial release.
- Circumventing robots.txt, CAPTCHAs, authentication, anti-bot controls, or rate limiting.
- A universal SEO health score presented as a search-ranking prediction.

## 6. Users and primary jobs

### 6.1 SEO practitioner

Needs to crawl a client site, identify technical problems, filter affected URLs, export evidence, and compare the result after remediation.

### 6.2 Developer or site operator

Needs reproducible failure lists, source/target link evidence, and machine-readable output suitable for CI or release checks.

### 6.3 Agency analyst

Needs named projects, reusable crawl profiles, historical comparisons, and spreadsheet exports without leaking client data to an external service.

### 6.4 AI-assisted operator

Needs an MCP client to start a bounded crawl, poll status, inspect issues and representative evidence, and export a report without receiving arbitrary filesystem or network access.

## 7. Guiding principles

1. **Evidence before advice:** findings show the observed value and rule that produced them.
2. **Safe defaults:** local binding, private-network denial, TLS validation, limits, and robots compliance are defaults.
3. **One engine:** UI, CLI, API, and MCP call the same application services and rule engine.
4. **Deterministic core:** parsing and rules must be testable without the network.
5. **Progressive capability:** raw HTML is the default; rendering and integrations are explicit options.
6. **No silent recovery:** retries, TLS failures, skipped URLs, and incomplete coverage are reported.
7. **Local ownership:** data stays local unless the user explicitly exports it.

## 8. Product scope and priorities

Priorities use MoSCoW: Must, Should, Could, Won't for version 1.

### 8.1 Projects and configuration

| ID | Requirement | Priority |
|---|---|---|
| PRJ-01 | Create, rename, archive, and delete a local project. | Must |
| PRJ-02 | Store seed URL, allowed hosts, include/exclude rules, limits, user agent, rendering mode, and explicit response-compression mode in a reusable crawl profile. | Must |
| PRJ-03 | Preview the effective crawl scope before starting. | Must |
| PRJ-04 | Import and export a non-secret crawl profile as JSON. | Should |
| PRJ-05 | Support multiple profiles per project. | Should |

### 8.2 Crawl engine

| ID | Requirement | Priority |
|---|---|---|
| CRW-01 | Crawl HTTP and HTTPS URLs using breadth-aware scheduling and URL deduplication. | Must |
| CRW-02 | Restrict traversal to configured hosts and path rules. | Must |
| CRW-03 | Enforce maximum URLs, depth, response bytes, duration, concurrency, redirects, and per-host request rate. | Must |
| CRW-04 | Respect robots.txt by default and record robots decisions. | Must |
| CRW-05 | Support sitemap discovery from robots.txt and common sitemap locations. | Must |
| CRW-06 | Fetch and recursively parse bounded sitemap indexes and URL sets. | Must |
| CRW-07 | Follow redirects within configured limits and record complete redirect chains. | Must |
| CRW-08 | Retry selected transient failures with bounded exponential backoff and jitter. | Must |
| CRW-09 | Pause, resume, and cancel a crawl without corrupting persisted data. | Must |
| CRW-10 | Stream progress and expose queued, active, completed, skipped, and failed counts. | Must |
| CRW-11 | Detect common crawler traps, repeated path segments, calendar spaces, session/tracking parameters, and excessive query variants. | Must |
| CRW-12 | Support optional Playwright rendering with separate worker and resource budgets. | Should |
| CRW-13 | Capture raw-versus-rendered differences for selected page fields. | Should |
| CRW-14 | Support list mode from pasted or uploaded URLs. | Should |
| CRW-15 | Support authenticated crawling. | Won't |

### 8.3 Extraction

| ID | Requirement | Priority |
|---|---|---|
| EXT-01 | Record requested URL, final URL, method, status, content type, byte size, response time, redirect chain, and selected response headers. | Must |
| EXT-02 | Extract title, meta description, meta robots, X-Robots-Tag, canonical, H1/H2, word count, language, viewport, and HTML hash. | Must |
| EXT-03 | Extract internal/external links with source URL, destination URL, anchor text, rel attributes, and discovery source. | Must |
| EXT-04 | Extract images, source URLs, alt presence/value, dimensions when declared, and page associations. | Must |
| EXT-05 | Extract hreflang values and reciprocal relationships. | Must |
| EXT-06 | Extract JSON-LD, microdata, and RDFa types without executing page scripts. | Must |
| EXT-07 | Detect Open Graph and common social metadata. | Should |
| EXT-08 | Extract pagination signals and mobile-related metadata. | Should |
| EXT-09 | Support bounded custom extraction with CSS selectors and XPath. | Should |
| EXT-10 | Extract text from PDFs. | Could |

### 8.4 Audit rules

| ID | Requirement | Priority |
|---|---|---|
| AUD-01 | Evaluate response codes, broken internal links, redirect chains/loops, and redirect targets. | Must |
| AUD-02 | Evaluate missing, duplicate, empty, short, and long titles and descriptions using configurable thresholds. | Must |
| AUD-03 | Evaluate missing/multiple H1s and title/H1 duplication. | Must |
| AUD-04 | Evaluate canonical absence, invalid targets, conflicting canonicals, chains, and non-indexable canonical targets. | Must |
| AUD-05 | Determine indexability from response, directives, canonical state, and crawl configuration. | Must |
| AUD-06 | Evaluate sitemap coverage, non-200 entries, redirects, non-indexable entries, and crawled URLs absent from sitemaps. | Must |
| AUD-07 | Evaluate robots blocks and report scope limitations. | Must |
| AUD-08 | Detect exact duplicates and configurable near-duplicate content. | Must |
| AUD-09 | Evaluate hreflang validity, language/region syntax, targets, and return tags. | Must |
| AUD-10 | Evaluate image alt coverage and broken image resources. | Must |
| AUD-11 | Evaluate depth, orphan candidates, low-inlink pages, and internal nofollow links. | Must |
| AUD-12 | Evaluate mixed content and selected security headers as technical observations, not ranking guarantees. | Must |
| AUD-13 | Evaluate structured-data syntax and known type/property consistency. | Should |
| AUD-14 | Validate selected Google rich-result profiles from versioned local rule data. | Should |
| AUD-15 | Run accessibility checks in rendered mode using an isolated, version-pinned engine. | Could |
| AUD-16 | Integrate Lighthouse or PageSpeed data. | Could |

Every audit rule must define a stable ID, title, severity, category, applicability conditions, evidence schema, remediation guidance, limitations, and rule version.

### 8.5 Results and reports

| ID | Requirement | Priority |
|---|---|---|
| REP-01 | Provide crawl overview, coverage, response-code distribution, indexability, and issue severity counts. | Must |
| REP-02 | Filter, sort, and search pages and issues without loading the full crawl into browser memory. | Must |
| REP-03 | Show page details, response metadata, extracted fields, issues, inlinks, outlinks, images, schema, and hreflang. | Must |
| REP-04 | Show every issue’s explanation, evidence, affected URLs, and configurable threshold where relevant. | Must |
| REP-05 | Export pages and issues to streaming CSV and JSON/NDJSON. | Must |
| REP-06 | Export a multi-sheet XLSX workbook for bounded result sets. | Should |
| REP-07 | Compare two completed crawls with added, removed, fixed, new, and changed findings. | Must |
| REP-08 | Save named report views and filters. | Could |
| REP-09 | Produce a self-contained HTML executive report. | Could |

### 8.6 CLI, API, and MCP

| ID | Requirement | Priority |
|---|---|---|
| INT-01 | Provide CLI commands for project management, crawling, status, issues, page lookup, comparison, and export. | Must |
| INT-02 | Provide a versioned loopback HTTP API used by the local UI. | Must |
| INT-03 | Provide an MCP server over stdio; network MCP is disabled in version 1. | Must |
| INT-04 | MCP tools use explicit schemas, bounded pagination, and opaque IDs. | Must |
| INT-05 | Starting or cancelling a crawl requires an explicit MCP tool call; read tools cannot mutate state. | Must |
| INT-06 | MCP tools cannot accept arbitrary output paths; exports return managed artifact IDs and approved paths. | Must |
| INT-07 | Long-running crawl tools return quickly with crawl IDs and separate status/result calls. | Must |
| INT-08 | Provide an OpenAPI document for the local HTTP API. | Should |

Required MCP tools for version 1:

- `project_create`
- `project_list`
- `crawl_preview_scope`
- `crawl_start`
- `crawl_status`
- `crawl_cancel`
- `crawl_list`
- `audit_summary`
- `issue_list`
- `issue_explain`
- `page_get`
- `crawl_compare`
- `report_export`
- `artifact_get`

Implemented agent-usability extensions:

- `profile_create`
- `profile_list`
- `crawl_pause`
- `crawl_resume`
- `crawl_timeline`
- `page_list`
- `link_list`
- `diagnostic_create`

## 9. Core user journeys

### 9.1 First audit

1. User opens the local application.
2. User creates a project and enters a public website URL.
3. Product resolves and validates the target, displays allowed hosts and default limits, and previews robots/sitemap scope.
4. User starts the crawl.
5. Product shows live status, warnings, coverage, and an explicit stop control.
6. Product completes and opens the audit summary.
7. User inspects an issue, affected URLs, evidence, and remediation guidance.
8. User exports selected results.

### 9.2 Recrawl and compare

1. User opens an existing project and selects its profile.
2. User starts a new crawl after site changes.
3. Product compares the completed crawl against a chosen baseline.
4. User reviews fixed, new, and changed issues and exports the comparison.

### 9.3 MCP-driven audit

1. MCP client calls `crawl_preview_scope`.
2. Client presents scope and limits to the user when needed.
3. Client calls `crawl_start` and receives an opaque crawl ID.
4. Client polls `crawl_status` with bounded frequency.
5. Client requests `audit_summary` and paginated `issue_list` results.
6. Client requests representative `page_get` evidence before recommending changes.
7. Client optionally requests a managed export artifact.

## 10. UX requirements

- The interface must distinguish configured limits, discovered URLs, fetched URLs, skipped URLs, and analysed pages.
- “Unlimited” is not offered; users may select a high explicit limit.
- Every incomplete crawl shows the reason: cancelled, time limit, URL limit, disk limit, crash recovery, robots restriction, or error.
- Severity is visually distinct from confidence and coverage.
- Recommendations must include caveats where business intent could change the correct action.
- Tables must be virtualised or server-paginated.
- Destructive actions require confirmation and identify the exact project/crawl affected.
- Keyboard navigation and WCAG 2.1 AA contrast are required for primary workflows.
- The UI never exposes stack traces, secrets, local absolute paths, or arbitrary response bodies by default.

## 11. Security and privacy requirements

### 11.1 Network boundaries

- The HTTP server binds to `127.0.0.1` and `::1` only by default.
- Remote/LAN binding is unavailable in version 1 unless implemented later with authentication, TLS, and explicit configuration.
- MCP communicates over stdio only.
- Allowed target schemes are `http` and `https`.
- Userinfo in URLs is rejected.
- Initial DNS results and every redirect target are resolved and checked.
- Loopback, private, link-local, multicast, reserved, carrier-grade NAT, and cloud metadata addresses are denied by default for IPv4 and IPv6.
- Mixed public/private DNS answers are denied.
- DNS rebinding is mitigated by connecting only to validated resolved addresses while preserving host/SNI semantics.
- A project may allow private targets only through an explicit advanced setting and per-host allowlist; this capability is excluded from MCP in version 1.

### 11.2 Fetch safety

- TLS certificates and hostnames are verified. There is no automatic insecure fallback.
- Response headers and streaming byte limits are enforced before buffering content.
- Gzip mode uses decompression-ratio and total decompressed-size limits to protect against compressed bombs. An explicit disabled mode may omit `Accept-Encoding` for verified CDN incompatibilities while retaining identity-response byte limits; it never retries or bypasses access-control responses.
- Redirect count, header size, URL length, document size, and request duration are bounded.
- Unsupported protocols and content are recorded and skipped.
- Cookies are disabled by default and scoped to a crawl when enabled later.
- Page-provided downloads, service workers, dialogs, permissions, and browser persistence are disabled in rendering workers.

### 11.3 Application and data safety

- No update, restart, shell, package-install, or arbitrary-command endpoint exists.
- The application does not mutate its own source code.
- No telemetry is sent by default.
- Logs redact URL credentials, query values marked sensitive, and authorization/cookie headers.
- Exports use generated artifact IDs; path traversal and caller-selected arbitrary destinations are prohibited through API/MCP.
- Database writes are transactional and schema migrations are reversible or backed up.
- Project and crawl deletion is scoped, confirmed, logged locally, and recoverable through a short trash-retention period where feasible.
- Content Security Policy and anti-CSRF controls protect the local web interface.
- API state mutations require a per-install local session token, even on loopback.

### 11.4 Supply-chain safety

- Runtime and direct dependencies are pinned through a committed lockfile.
- CI runs dependency, license, secret, static-analysis, and container/filesystem scans where applicable.
- Releases are built in CI, checksummed, signed when infrastructure permits, and tied to source tags.
- Updates are user-initiated downloads with release verification; there is no unattended `git pull`.

## 12. Responsible crawling requirements

- Default user agent identifies SEO Auditor and provides a documentation URL.
- Default concurrency is conservative and per-host delay is nonzero.
- `Retry-After` is respected.
- Repeated 429/503 responses automatically reduce concurrency.
- Robots.txt is respected by default; disabling robots compliance requires a visible local-only advanced choice and is unavailable through MCP version 1.
- The UI reminds users to crawl only sites they are authorised to assess.
- Crawl cancellation stops new scheduling immediately and winds down active requests.

## 13. Data and retention

Stored data includes projects, profiles, crawl state, URLs, fetch attempts, redirects, extracted fields, links, directives, structured data summaries, issues, rule versions, events, and artifacts. Raw HTML is off by default; when enabled it has a separate size limit and retention control.

Default retention:

- Completed crawl structured data: retained until deleted.
- Raw response bodies: not stored.
- Rendered screenshots: not stored in version 1.
- Temporary export artifacts: seven days.
- Operational logs: 14 days with size rotation.
- Trash for deleted crawls: seven days where disk capacity permits.

## 14. Performance and reliability targets

- Default raw crawl concurrency: 5 per host, configurable up to a safe global cap.
- Default delay: 250 ms per host, adjusted by backoff.
- UI status latency: under two seconds during normal crawling.
- Resume checkpoint: at least every 100 processed URLs or five seconds.
- Database integrity must survive forced process termination in automated tests.
- Memory target for a 100,000-URL raw crawl: bounded primarily by active work and caches; URL/link records live in SQLite rather than an all-in-memory graph.
- Export operations stream data and do not require loading a complete crawl into memory.
- Renderer concurrency defaults to one and has an explicit maximum.

Performance varies by site, network, rendering mode, hardware, robots policy, and configured delay. The product reports actual request rate and limiting factors.

## 15. Compatibility

Initial supported environments:

- Linux x64 and arm64.
- macOS Apple Silicon and x64.
- Windows 11 x64.
- The project-pinned stable Go toolchain for source installations.
- Node.js is required from source only for building the web UI and optional JavaScript renderer; packaged releases bundle the required assets/runtime.
- Current Chromium supplied by the pinned Playwright version for rendered mode.

The local UI supports current stable Chrome, Edge, Firefox, and Safari. Packaged desktop distribution is a post-MVP decision; the first implementation may ship as a local service plus browser UI and CLI.

## 16. Observability

- Structured local logs with component, crawl ID, event, severity, and safe metadata.
- Per-crawl event timeline for start, pause, resume, backoff, limit reached, cancel, recovery, and completion.
- Counters for requests, status groups, retries, bytes, queue states, skips by reason, and rendering failures.
- Diagnostic bundle export excludes crawled content and secrets by default.
- No external telemetry unless a future opt-in proposal is separately approved.

## 17. Acceptance criteria for version 1

Version 1 is releasable only when:

1. All Must requirements are implemented or explicitly removed through a PRD revision.
2. The security test suite demonstrates private-address blocking, redirect revalidation, DNS-rebinding defenses, TLS enforcement, size limits, and path protection.
3. UI, CLI, HTTP API, and MCP return consistent results for the same crawl.
4. Raw and rendered fixture crawls pass extraction and rule conformance tests.
5. A 10,000-URL endurance crawl and a 100,000-URL synthetic storage test pass documented resource thresholds.
6. Cancellation, crash recovery, database backup, migration, and deletion flows pass end-to-end tests.
7. Exported CSV, NDJSON, and supported XLSX reports match the documented schemas.
8. The product includes an operator guide, rule catalog, data-retention explanation, responsible-use policy, and security model.
9. A clean machine can install and run the signed/tagged release using documented steps.
10. No high or critical known dependency vulnerabilities remain without a documented, time-bounded exception.

## 18. Risks and mitigations

| Risk | Impact | Mitigation |
|---|---|---|
| SSRF or access to internal services | Critical | Central target policy, IP validation, redirect validation, pinned connections, adversarial tests. |
| Crawler traps cause unbounded work | High | Mandatory limits, trap heuristics, query budgets, live queue visibility, cancellation. |
| JavaScript renderer compromise or resource exhaustion | High | Isolated worker process, browser sandbox where supported, blocked downloads/permissions, CPU/memory/time budgets. |
| False-positive SEO recommendations | High | Applicability rules, evidence, configurable thresholds, confidence and limitations. |
| Database growth and poor large-crawl performance | High | Normalised SQLite schema, indexes, WAL, batching, streaming exports, load tests. |
| Supply-chain compromise | High | Lockfile, scanning, reviewed updates, checksums/signing, no auto-pull. |
| Scope expansion toward full parity | Medium | Versioned roadmap and explicit non-goals. |
| Sites block or throttle the crawler | Medium | Honest user agent, adaptive backoff, respectful defaults, clear incomplete-coverage reporting. |
| Rule definitions become stale | Medium | Versioned rule packs, citations/last-reviewed metadata, conformance fixtures. |

## 19. Future roadmap candidates

- Search Console and analytics integrations with least-privilege OAuth.
- Lighthouse and accessibility audits.
- Safe custom extraction and saved extractor libraries.
- Authenticated crawling with an isolated secrets store.
- Scheduled crawls through an OS-level scheduler without self-update privileges.
- Crawl visualisations and internal-link optimisation views.
- PDF extraction.
- Signed desktop packages.
- Authenticated remote MCP/HTTP deployment for teams.
- Optional segmented single-machine performance research beyond the completed 5-million synthetic campaign.
- Optional distributed-campaign research informed by the theoretical 100M+ architecture note.
- Partitioned analytical storage and object-backed immutable result segments for very large campaigns.
- Plugin API for trusted, signed rule packs.

These candidates are not commitments until promoted into a revised PRD.

## 20. Reference-repository policy

The downloaded `open-seo-crawler` repository may be used as an MIT-licensed reference. Reuse must follow these rules:

1. Preserve required copyright and MIT license notices for copied or substantially derived code.
2. Review and rewrite security-sensitive network, process, update, storage, and route logic.
3. Prefer importing behavioural ideas, fixtures, rule taxonomies, and edge cases over copying the single-file server design.
4. Track reused source in `THIRD_PARTY_NOTICES.md` with upstream path and commit.
5. Do not reuse branding, trade dress, or claims of affiliation.
6. Every reused function must receive tests and the same review as new code.

## 21. Product change control

This document is the final product baseline for initial implementation. A change requires an explicit PRD revision when it:

- adds or removes a Must requirement;
- changes the version-one supported scale or platforms;
- weakens a security, privacy, responsible-crawling, evidence, or recovery guarantee;
- changes stored data or default retention materially;
- changes the meaning of a public capacity or quality claim;
- promotes a future-roadmap item into committed scope.

Implementation discoveries may refine internal design without a PRD revision only when externally observable behaviour and acceptance criteria remain unchanged. Deferred Must requirements cannot be silently reclassified; the PRD and implementation baseline must be revised together.
