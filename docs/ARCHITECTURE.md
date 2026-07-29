# SEO Auditor system architecture

Status: Final architecture baseline  
Version: 2.0  
Date: 2026-07-30

## 1. Architecture objectives

The system must run locally, remain bounded under hostile input, preserve crawl state durably, and expose one consistent audit model through a web UI, CLI, HTTP API, and MCP. Network fetching is treated as a privileged capability behind one policy-enforcing subsystem.

The design optimises initially for 10 to 100,000 URLs per crawl. It favours correctness, inspectability, and safe recovery over maximum raw throughput.

## 2. Technology baseline

| Area | Choice | Rationale |
|---|---|---|
| Core language/runtime | Go, pinned to a supported stable toolchain | Strong concurrent networking, bounded worker design, native parallelism, race/fuzz tooling, official MCP SDK, and simple cross-platform binaries. |
| Web/render runtime | TypeScript on Node.js | React UI build and an optional isolated Playwright renderer; Node is not the crawl scheduler or security-policy authority. |
| Workspace | Go module plus a pnpm workspace under `web/` | Reproducible Go modules and frontend/renderer lockfile without making the core depend on Node. |
| Local API | Go `net/http` with a small schema/router layer | Minimal attack surface, explicit middleware, standard cancellation/timeouts, and controlled HTTP surface. |
| UI | React + Vite | Local browser application with mature tables and state tooling. |
| Database | SQLite in WAL mode | Local-first, transactional, portable, good indexed query performance. |
| Query/migrations | Reviewed SQL with a narrow Go repository layer | Explicit migrations and tuned SQL without coupling the domain to an ORM. |
| HTML parsing | `golang.org/x/net/html` plus reviewed selector helpers | Tolerant parsing without executing page scripts. |
| Rendering | Node/Playwright Chromium sidecar over framed stdio | Optional modern rendering isolated from the Go process and database. |
| MCP | Official Go MCP SDK over stdio | No listening network port; MCP calls the same Go application services as CLI/API. |
| Validation | Go request types plus generated JSON Schema/OpenAPI | Runtime boundary validation with compile-time internal types. |
| Tests | Go test/race/fuzz, Playwright test, fixture HTTP servers | Unit, integration, UI, race, fuzz, and adversarial network testing. |

Library selection is confirmed during the foundation spike. Any replacement must preserve the security boundaries and shared contracts described here.

## 3. System context

```text
              +------------------+
              | Local web browser|
              +---------+--------+
                        | loopback HTTP + session token
                        v
+-----------+   stdio  +------------------+    jobs    +----------------+
| MCP client|<-------->| MCP adapter      |----------->| Application    |
+-----------+          +------------------+            | services       |
                                                        +---+--------+---+
+-----------+             direct invocation                |        |
| CLI       |-----------------------------------------------+        |
+-----------+                                                        |
                                                          commands   | queries
                                                                     v
  Public web targets <---- guarded fetch/render workers ----+----------------+
                                                            | SQLite + files |
                                                            +----------------+
```

The UI server, CLI, and MCP adapter do not fetch URLs directly. They invoke application services, which enqueue crawl work. All outbound traffic passes through the target-policy and fetch layers.

## 4. Repository layout

```text
cmd/
  seo-auditor/         Local API/UI supervisor and default executable
  seo-auditor-cli/     Scriptable CLI entry point
  seo-auditor-mcp/     stdio MCP entry point
internal/
  contracts/           Public DTOs, IDs, validation and stable error codes
  application/         Use cases and transaction boundaries
  crawler/             Scheduler, frontier, robots, sitemap and orchestration
  fetchpolicy/         URL, DNS, IP, redirect, TLS and response-limit policy
  extractor/           Deterministic raw/rendered document extraction
  rules/               Versioned audit rules and evidence schemas
  renderer/            Sidecar protocol, supervision and browser controls
  database/            Schema, repositories, migrations, backup/restore
  reports/             Comparisons and streaming exporters
  observability/       Structured logs, metrics and diagnostic events
  testfixtures/        Local fixture sites and expected outputs
web/
  app/                 React application
  renderer/            TypeScript Playwright sidecar
  package.json         Frontend/renderer workspace manifest
docs/                  product, architecture, operations and rule documentation
```

Go packages form a directed dependency graph. Domain packages cannot import API, UI, CLI, or MCP entry points. `fetchpolicy` has no dependency on presentation layers or the renderer implementation. The TypeScript renderer has no database credentials and cannot broaden crawl scope.

## 5. Runtime processes

### 5.1 Supervisor/API process

Owns the loopback HTTP server, application-service wiring, job supervision, local session token, and worker lifecycle. It does not parse unbounded response bodies or run browser pages.

### 5.2 Raw crawl workers

Perform guarded DNS resolution, HTTP requests, streaming limits, HTML parsing, and batched result submission. Worker concurrency is bounded globally and per host. A worker receives a normalised job specification, not arbitrary executable configuration.

### 5.3 Render workers

Run Playwright in separate processes. They have lower concurrency, hard timeouts, restricted browser contexts, disabled downloads and permissions, and interception rules for subresources. If a render worker crashes, the crawl records a rendering failure and the supervisor replaces the worker within a restart budget.

### 5.4 Database writer

Serialises batched writes through a dedicated queue/worker so synchronous SQLite operations cannot stall network scheduling or the API event loop. Readers use short-lived read transactions. WAL checkpoints and backups are controlled by database services.

### 5.5 MCP process

Runs as a Go stdio child process of the MCP client. It connects to the local application through a per-install authenticated loopback channel. Version 1 prefers connection to one running supervisor to prevent concurrent database ownership; direct embedded database ownership is deferred unless locking and migrations are proven safe.

## 6. Component responsibilities

### 6.1 Application services

Application services implement use cases such as creating projects, previewing scope, starting/cancelling crawls, querying issues, comparing crawls, and creating artifacts. They own authorization and transaction boundaries and return presentation-neutral DTOs.

### 6.2 Crawl scheduler and frontier

The scheduler:

- stores the frontier in SQLite;
- leases bounded batches to workers;
- enforces global/per-host concurrency and delay;
- orders work by crawl depth, discovery priority, and politeness readiness;
- deduplicates canonical request keys;
- applies include/exclude, host, query, depth, and limit policies before enqueue;
- checkpoints queue and counters;
- recovers expired leases after crashes;
- stops scheduling immediately on pause/cancel/terminal limits.

URL state transitions:

```text
discovered -> accepted -> queued -> leased -> fetched -> extracted -> analysed
                  |          |         |          |
                  +-------> skipped    +-------> retry_wait
                                             or failed
```

Terminal states are `analysed`, `skipped`, and `failed`. A lease has an expiry and attempt number. State transitions are transactional.

### 6.3 Target-policy service

This is the mandatory gateway for every top-level request, redirect, sitemap URL, and renderer subresource.

Validation order:

1. Parse with a standards-compliant URL parser.
2. Require HTTP/HTTPS and reject userinfo.
3. Normalise scheme, host, port, path, and fragment handling.
4. Apply project host/path/query scope.
5. Resolve all A/AAAA records through the controlled resolver.
6. Reject prohibited IP classes or mixed public/private results.
7. Select and pin an allowed address for the connection.
8. Preserve hostname for the `Host` header and TLS SNI/certificate validation.
9. Re-run the entire policy for every redirect.
10. Record safe decision metadata and denial reason.

The renderer interception layer calls the same policy for document and subresource requests. Private-target exceptions, if later implemented, require a locally configured allowlist and can never be passed through MCP.

### 6.4 HTTP fetcher

The fetcher uses streaming I/O and enforces:

- connect, headers, body-idle, and total deadlines;
- header-count and header-byte limits;
- compressed and decompressed body limits;
- decompression-ratio limits;
- redirect policy and maximum hops;
- supported MIME types;
- retry classes and `Retry-After`;
- TLS verification with no insecure fallback;
- safe user-agent and accept headers;
- per-crawl cookie policy, disabled by default.

Bodies are passed to extractors through bounded buffers or temporary managed files. Raw body persistence is disabled by default.

### 6.5 Robots and sitemap services

Robots policies are cached by scheme/host/port with expiry. Fetch failures and parse ambiguities are explicit. The sitemap parser streams XML, limits nested indexes, URL counts and bytes, rejects external entities, and applies target policy before accepting sitemap locations or entries.

### 6.6 Extractor

The extractor is a pure transformation where possible:

```text
FetchRecord + bounded document bytes -> ExtractedPage + discovered link candidates
```

It performs tolerant HTML parsing, base URL resolution, metadata/link/image/hreflang/schema extraction, visible-text normalisation, and hashing. It emits warnings for malformed markup rather than throwing the crawl.

Extracted values retain provenance such as source element, raw value, resolved value, and raw-versus-rendered mode. Text content is bounded before hashing or similarity analysis.

### 6.7 Rule engine

Rules consume typed page, link-graph, sitemap, and crawl aggregates. A rule definition includes:

```go
type RuleDefinition[I any, E any] interface {
    Metadata() RuleMetadata
    Applies(input I, config RuleConfig) bool
    Evaluate(input I, config RuleConfig) []E
    Explain() RuleExplanation
}
```

Rules cannot perform network calls, write files, execute code, or mutate crawl records. Results identify rule ID/version, subject IDs, severity, confidence where applicable, and structured evidence. Aggregate rules run after dependent crawl data is stable.

### 6.8 Comparison and reports

Comparison matches URLs by a documented normalised identity while retaining original URLs. It separates:

- pages added and removed;
- fields changed;
- issues new, fixed, persistent, or severity-changed;
- coverage/configuration differences that make comparison incomplete.

Exporters query with keyset pagination and stream output. XLSX has explicit row/cell safety, formula-injection escaping, sheet-name sanitisation, and size thresholds. Artifacts live in an application-owned directory under opaque IDs.

## 7. Data architecture

### 7.1 Principal entities

| Entity | Purpose |
|---|---|
| `project` | Named audit boundary and lifecycle. |
| `crawl_profile` | Versioned scope, rate, limit, user-agent, and extraction settings. |
| `crawl` | One immutable configuration snapshot plus status and aggregate counters. |
| `url` | Normalised URL identity within a project. |
| `crawl_url` | Per-crawl URL state, depth, discovery source, and fetch outcome. |
| `fetch_attempt` | Request timing, resolved IP class, status, bytes, retry and error. |
| `redirect_hop` | Ordered redirect evidence. |
| `page` | Extracted page-level values and hashes. |
| `heading` | Ordered headings. |
| `link` | Source/destination relationship, anchor, rel and discovery mode. |
| `image` / `page_image` | Image identity and page associations. |
| `hreflang` | Source, language/region, target and validation state. |
| `structured_data` | Format, type summary and bounded validation evidence. |
| `sitemap` / `sitemap_entry` | Sitemap discovery and membership. |
| `issue` | Rule/version, subject, severity, confidence and evidence JSON. |
| `crawl_event` | Operational audit trail and recovery information. |
| `artifact` | Managed exports with checksum, size, format and expiry. |

### 7.2 Database practices

- SQLite foreign keys enabled.
- WAL mode with bounded busy timeout.
- Migrations run under an application lock after backup.
- Batched inserts and prepared statements.
- Composite indexes reflect issue, page, link and comparison queries.
- Large text is bounded and optional; repeated strings may be interned only after profiling.
- Configuration snapshots and rule versions make historical results reproducible.
- Database paths are generated beneath the application data directory, never accepted from API/MCP input.

### 7.3 Crawl consistency

A crawl has a state machine:

```text
created -> validating -> queued -> running <-> paused
                                  |   |  
                                  |   +-> cancelling -> cancelled
                                  +-----> completing -> completed
                                  +-----> failed
```

Only defined transitions are allowed. Completion freezes the configuration and aggregate rule set. A recovered process moves stale `running` crawls into `paused_recovered` before user-controlled resume, unless all work was already committed and completion can be proven.

## 8. API architecture

### 8.1 Local HTTP API

- Prefix: `/api/v1`.
- Bind: loopback only.
- Request and response validation from shared schemas.
- Per-install random session secret stored with user-only permissions.
- Browser receives a same-site, HTTP-only session cookie after loading a one-time bootstrap URL.
- State-changing requests require CSRF protection and `Origin` validation.
- Request body, pagination, and response limits apply globally.
- Errors use stable codes and safe messages; stack traces remain in redacted local logs.

Representative endpoints:

```text
POST   /api/v1/projects
GET    /api/v1/projects
POST   /api/v1/projects/{projectId}/scope-preview
POST   /api/v1/projects/{projectId}/crawls
GET    /api/v1/crawls/{crawlId}
POST   /api/v1/crawls/{crawlId}/pause
POST   /api/v1/crawls/{crawlId}/resume
POST   /api/v1/crawls/{crawlId}/cancel
GET    /api/v1/crawls/{crawlId}/issues
GET    /api/v1/crawls/{crawlId}/pages/{pageId}
POST   /api/v1/comparisons
POST   /api/v1/exports
GET    /api/v1/artifacts/{artifactId}
```

Live progress uses Server-Sent Events with reconnect cursors. Crawl control remains ordinary authenticated HTTP.

### 8.2 MCP interface

The MCP adapter maps each tool to one application use case. It provides no generic HTTP request, SQL, filesystem, shell, or browser tool.

MCP design rules:

- stdio transport only;
- exact JSON schemas with unknown fields rejected;
- opaque UUID/ULID identifiers;
- bounded page sizes and output character counts;
- mutations clearly named and separated from reads;
- `crawl_start` accepts a project/profile or validated configuration subset, never raw internal options;
- no `ignoreRobots`, private-network override, arbitrary headers, cookies, scripts, or arbitrary export paths;
- long jobs return IDs immediately;
- artifacts are produced inside the managed data directory;
- tool errors expose stable codes and remediation, not stack traces.

## 9. Rendering architecture

Rendering is opt-in because it is slower and increases attack surface.

Each page receives a fresh or carefully reset browser context. Controls include:

- Playwright-managed Chromium pinned by lockfile;
- subprocess isolation and an OS sandbox where available;
- navigation and total-page deadlines;
- request interception through target policy;
- blocked downloads, WebRTC, geolocation, notifications, clipboard, camera and microphone;
- no persistent profiles, extensions, saved credentials or host filesystem mounts;
- configurable blocking of media, fonts, tracking and third-party resources;
- maximum subrequests and transferred bytes;
- DOM serialisation and extracted data size limits;
- worker recycling after a page count, memory threshold, crash, or timeout.

Containerisation may strengthen Linux deployments but is not treated as the only security boundary because desktop installs must also be safe.

## 10. Threat model

### 10.1 Assets

- Host machine and local network access.
- Local files and credentials.
- Crawl databases and client site data.
- Application integrity and update channel.
- Availability of the workstation and target websites.

### 10.2 Adversaries and untrusted inputs

- A malicious or compromised target website.
- A URL supplied by an untrusted user or MCP conversation.
- Malformed HTML, XML, compression, redirects, DNS and TLS responses.
- A webpage attempting browser escape, downloads or local-network requests.
- A malicious dependency or upstream release.
- Another local webpage attempting cross-origin requests to the loopback API.

### 10.3 Primary controls

| Threat | Controls |
|---|---|
| SSRF and cloud metadata access | Central IP policy, redirect/subresource revalidation, connection pinning, private ranges denied. |
| DNS rebinding | Validate all answers, pin selected address, reject mixed answers, re-resolve under policy. |
| Cross-site request forgery against localhost | Session token, same-site cookie, CSRF token, Origin checks, no permissive CORS. |
| Browser-based attack | Separate worker, sandbox, fresh contexts, blocked capabilities, resource limits, patched pinned browser. |
| XML/zip/compression bombs | Streaming parsers, external entities disabled, byte/ratio/nesting limits. |
| Filesystem traversal | Managed directories and opaque artifact IDs; no caller-provided paths. |
| Spreadsheet formula injection | Prefix/escape dangerous cell values and document behaviour. |
| Resource exhaustion | Mandatory crawl/depth/time/byte/query/concurrency limits, disk budget, cancellation. |
| Supply-chain compromise | Lockfile, scanning, signed/tagged releases, manual verified updates. |
| Data leakage through logs/MCP | Redaction, response limits, selected fields, raw content off by default. |

## 11. Reliability and recovery

- Worker leases make jobs recoverable after process failure.
- Idempotency keys protect crawl-start and export mutations from accidental duplicates.
- Database writes use transactions; large writes are chunked.
- Startup runs integrity checks after unclean shutdown and preserves the original database before repair attempts.
- Migrations create a verified backup and never auto-delete the previous schema copy.
- Disk-space thresholds pause scheduling before the filesystem is exhausted.
- A crawl can complete with warnings when individual pages fail; systemic failures move the crawl to failed with a reason.
- Shutdown stops leasing, checkpoints counters, drains within a deadline, then terminates workers.

## 12. Testing architecture

### 12.1 Test layers

1. Unit tests for URL normalisation, scope, robots, extraction, rules and comparison.
2. Property/fuzz tests for URLs, redirects, malformed HTML/XML, encodings, headers and decompression.
3. Integration tests against controlled fixture servers for redirects, TLS, rate limits, robots and sitemaps.
4. Adversarial network tests for private IPs, IPv6, mixed DNS, rebinding simulation and metadata paths.
5. Database tests for migrations, crash recovery, lease expiry, concurrent reads and disk limits.
6. Contract tests ensuring API, CLI and MCP return equivalent DTOs.
7. Browser end-to-end tests for project, crawl, issue, comparison and export flows.
8. Endurance and performance tests at 10k and synthetic 100k URL scales.
9. Dependency/static/security scanning in CI.

### 12.2 Fixture strategy

Fixture sites are deterministic and local. They cover valid and malformed markup, duplicate content, canonical/hreflang graphs, redirect chains/loops, sitemap indexes, robots variations, client rendering, infinite URL traps, large responses, encoding errors and slow/rate-limited servers. Expected extraction and rule outputs are version-controlled.

## 13. Deployment and updates

Initial developer distribution uses the pinned Go toolchain plus Node.js/pnpm for web assets and the optional renderer. Raw crawling, CLI, API, and MCP build into Go binaries. Production packages embed compiled web assets; rendered mode includes or downloads a separately checksummed renderer bundle and compatible Chromium.

Release principles:

- source tags and immutable release artifacts;
- generated SBOM and checksums;
- platform CI builds;
- user-initiated update checks that report availability without executing code;
- verified download followed by explicit install/restart;
- database compatibility and rollback notes per release.

There is no background `git pull`, package installation, or self-restart mechanism.

## 14. Architecture decisions and required follow-on ADRs

- ADR-001: Go for the crawler, application services, API, CLI, and MCP. **Accepted 2026-07-30.**
- ADR-002: SQLite driver and dedicated writer strategy.
- ADR-003: Raw HTTP client and DNS/IP pinning implementation.
- ADR-004: MCP embedded mode versus authenticated connection to supervisor.
- ADR-005: Application packaging for Linux, macOS and Windows.
- ADR-006: Near-duplicate algorithm and storage trade-offs.
- ADR-007: Browser sandbox strategy per operating system.
- ADR-008: Rule-pack versioning and documentation citations.

The implementation plan includes short spikes to resolve the highest-risk decisions before feature expansion.

## 15. Scale evolution

The version-1 SQLite architecture must process work in bounded batches even though its supported crawl size is 100,000 URLs. This prevents core algorithms from assuming that the whole frontier, link graph, issue set, or export fits in memory.

Scale evolves through three compatible execution modes:

| Mode | Intended scale | Storage/execution |
|---|---:|---|
| Local audit | Up to 100,000 supported initially | One Go process, SQLite frontier and audit database. |
| Local campaign | 1–10 million targeted post-v1 | Checkpointed segments, SQLite or an evaluated embedded store, deferred aggregates, strict disk budgets. |
| Distributed campaign | 100 million+ north-star | Coordinator, host-owned worker leases, partitioned operational store, immutable object segments, analytical database. |

A crawl campaign has one immutable configuration and global identity. A segment is only a bounded unit of leasing, committing, checkpointing, and compaction. Segment boundaries cannot reset deduplication, depth, per-host politeness, robots policy, retries, graph identity, or rule versions.

The distributed design must preserve host ownership: at any moment one logical scheduler controls request timing for a host, even when many workers participate. This prevents horizontal scale from violating politeness or producing inconsistent robots decisions.

Page-local rules run as records commit. Graph and whole-crawl rules use incremental aggregates where exact semantics are proven; otherwise they run against the completed global dataset. Approximate structures such as Bloom filters may reduce reads but cannot be the authoritative deduplication source because false positives would silently omit pages.

The complete scale design and qualifying benchmark are defined in [SCALE_STRATEGY.md](./SCALE_STRATEGY.md).

## 16. Architecture conformance and change control

Implementation must conform to these non-negotiable boundaries:

1. All outbound requests pass through `fetchpolicy`, including redirects, robots, sitemaps, external checks and renderer subresources.
2. Go application services remain the only authority for crawl scope, state transitions and database writes.
3. The renderer has no database credentials, unrestricted filesystem access, or authority to broaden navigation scope.
4. UI, CLI, API and MCP adapters do not duplicate audit or crawl-domain logic.
5. Version-one frontier and committed results are durable; process memory is never the only source of crawl truth.
6. Public network binding, network MCP, arbitrary command execution, caller-selected filesystem paths and insecure TLS fallback are absent from version one.
7. Segment processing is an internal unit of one global campaign and cannot reset deduplication, graph identity or politeness state.

An ADR is required before changing the core language, primary database, renderer trust boundary, MCP transport, target-policy semantics, public binding model, rule execution model, or campaign consistency guarantees. An experimental spike may evaluate alternatives, but production code cannot cross these boundaries until its ADR is accepted and affected plans are revised.
