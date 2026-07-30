# SEO Auditor implementation plan

Status: Final implementation baseline  
Version: 2.0  
Date: 2026-07-30  
Planning unit: one focused engineering stream assisted by Codex; elapsed time depends on review availability and scope stability.

## 1. Delivery strategy

Build a vertical, secure crawl path before adding breadth. Every milestone must leave the repository testable and documented. Security controls are implemented alongside fetching, not retrofitted after the crawler works.

The delivery order is:

```text
Foundation -> guarded fetch -> durable crawl -> extraction/rules
           -> reports/API/UI -> MCP -> rendering -> hardening/release
```

Target duration for a strong version 1 is approximately 14–20 focused engineering weeks. A useful alpha appears earlier, after the raw crawler and essential rules. Dates should be set only after the foundation and guarded-fetch spikes establish actual velocity.

## 2. Workstreams

| Workstream | Responsibility |
|---|---|
| Platform | Monorepo, builds, configuration, release and CI. |
| Security | Target policy, local API security, renderer isolation, supply chain. |
| Crawl engine | Frontier, scheduling, robots, sitemaps, fetching, recovery. |
| Analysis | Extraction, graph construction, audit rules, comparisons. |
| Product surfaces | API, UI, CLI, MCP and exports. |
| Quality | Fixtures, conformance, fuzzing, endurance, documentation. |

One person may own several workstreams, but reviews should follow these boundaries.

## 3. Milestones

### Milestone 0 — Foundation and risk spikes

Estimated effort: 1–2 weeks  
Exit: architecture choices validated; repository can build and test on supported development platforms.

Tasks:

- Initialise the Go module, `cmd/`/`internal/` package boundaries, and the `web/` pnpm workspace from the architecture document.
- Pin Go, Node, and package-manager versions; commit `go.mod`, `go.sum`, and the frontend/renderer lockfile.
- Add Go formatting, vet/static analysis, race tests, fuzz-test harnesses, frontend TypeScript strict mode, linting, and coverage reporting.
- Add CI for Linux, macOS and Windows where practical.
- Establish security scanning, dependency review, secret scanning and license inventory.
- Define shared ID, error, pagination and job-status contracts.
- Create application data-directory resolution with safe permissions.
- Record accepted ADR-001 and spike ADR-002 through ADR-004.
- Prototype SQLite WAL, writer worker, migrations and 100k synthetic URL inserts.
- Prototype HTTP DNS resolution plus validated-IP connection pinning with correct TLS SNI.
- Prototype the official Go MCP SDK over stdio with one read-only health tool.
- Create `THIRD_PARTY_NOTICES.md` and reference-reuse checklist.

Tests/gates:

- Clean Go and web/renderer builds plus unit tests pass from a fresh clone.
- Lockfile is reproducible.
- Synthetic database spike meets documented throughput and memory observations.
- Fetch spike proves public HTTPS works while loopback/private/link-local targets fail.
- ADRs record rejected alternatives and consequences.

### Milestone 1 — Guarded HTTP and scope engine

Estimated effort: 2 weeks  
Exit: a CLI can safely fetch one public URL through the production target policy and store a bounded result.

Tasks:

- Implement strict URL parsing, normalisation and request-key generation.
- Implement host/path/query include and exclude rules.
- Implement IPv4/IPv6 prohibited-range classification.
- Implement controlled DNS resolution, mixed-answer rejection and connection pinning.
- Revalidate each redirect and enforce redirect limits.
- Implement streaming response limits, decompression limits and timeouts.
- Implement TLS verification and structured failure reporting.
- Implement per-host rate limiter, bounded retries and `Retry-After` handling.
- Implement safe user-agent defaults and header policy.
- Build fixture servers for redirects, TLS, slow bodies, oversized bodies, compression and rate limiting.
- Add URL and network-policy fuzz/property tests.

Tests/gates:

- SSRF suite covers loopback, RFC1918, link-local, CGNAT, multicast, reserved, IPv4-in-IPv6 and common metadata targets.
- Redirects from public to prohibited targets are blocked.
- DNS-rebinding simulation cannot change the connected address outside validated answers.
- TLS hostname and certificate failures never trigger an insecure retry.
- Byte, ratio and deadline limits stop hostile fixture responses predictably.

### Milestone 2 — Durable raw crawl engine

Estimated effort: 2–3 weeks  
Exit: bounded multi-page raw crawls survive cancellation and process restart.

Tasks:

- Create project, profile, crawl, URL, frontier, fetch-attempt and event tables.
- Implement crawl state machine and transactional transitions.
- Implement persistent frontier, URL deduplication, depth and discovery provenance.
- Implement worker leasing, expiry, retries and crash recovery.
- Implement global/per-host concurrency and adaptive throttling.
- Implement robots.txt fetch, parse, cache and decision evidence.
- Implement streaming sitemap and sitemap-index discovery/parsing.
- Implement mandatory URL/depth/time/disk/query/redirect limits.
- Implement pause, resume, cancel and graceful shutdown.
- Add CLI commands for project creation, scope preview, crawl start/status/cancel.
- Add progress counters and structured crawl events.
- Detect common crawler traps and repeated URL spaces.

Tests/gates:

- Fixture site crawl produces the expected URL set, robots decisions and sitemap membership.
- Forced termination recovers leases without duplicate committed page results.
- URL, depth, time and disk limits produce explicit terminal or paused states.
- Cancellation stops new leases immediately and drains active requests within its deadline.
- A 10,000-URL static fixture completes without unbounded memory growth.

### Milestone 3 — Extraction and essential audit rules

Estimated effort: 3 weeks  
Exit: completed crawls produce a trustworthy technical audit accessible through CLI/JSON.

Tasks:

- Implement bounded tolerant HTML parsing and base URL handling.
- Extract response fields, metadata, headings, directives, canonicals, links, images, hreflang, social metadata and language/viewport signals.
- Extract JSON-LD, microdata and RDFa summaries without executing code.
- Store page, heading, link, image, hreflang, sitemap and structured-data records.
- Implement visible-text normalisation, exact hashes and initial near-duplicate strategy.
- Define rule registry, rule metadata/evidence contracts and versioning.
- Implement Must rules AUD-01 through AUD-12.
- Implement crawl aggregate phase and indexability model.
- Produce a generated rule catalog from source metadata.
- Port useful behavioural cases from the reference repository as independently reviewed fixtures; record derivation notices when code is reused.

Tests/gates:

- Curated extraction fixtures pass across malformed HTML, encodings and base/canonical edge cases.
- Each rule has positive, negative, not-applicable and threshold-boundary tests.
- Every issue contains stable rule/version IDs and valid typed evidence.
- Re-running the same fixture with the same profile is semantically deterministic.
- The rule engine performs no network or filesystem operations.

### Milestone 4 — Query API, reports and crawl comparison

Estimated effort: 2 weeks  
Exit: data can be queried and exported without loading entire crawls into memory.

Tasks:

- Implement application services and repositories for summaries, pages, links and issues.
- Add keyset pagination, filter schemas, sort allowlists and query indexes.
- Implement configuration-aware crawl comparison.
- Implement streaming CSV and NDJSON exporters.
- Implement bounded XLSX export with formula-injection protection.
- Implement managed artifact lifecycle, checksums and expiry cleanup.
- Implement `/api/v1` endpoints and generated OpenAPI document.
- Add local session bootstrap, same-site cookie, CSRF token and Origin validation.
- Add Server-Sent Events with cursor-based reconnect for progress.
- Add database backup, migration verification and crawl trash/restore services.

Tests/gates:

- API rejects unknown fields, oversized bodies, invalid transitions and unbounded pagination.
- CSRF/cross-origin fixture pages cannot mutate the loopback API.
- Exports match schemas and escape spreadsheet formulas.
- Comparison distinguishes data changes from profile/coverage differences.
- 100k synthetic result queries and exports stay within measured memory bounds.

### Milestone 5 — Local web interface

Estimated effort: 2–3 weeks  
Exit: all primary version-1 workflows are usable from the browser.

Tasks:

- Build application shell, navigation and project list.
- Build project/profile editor with scope preview and safe defaults.
- Build crawl start confirmation, progress, pause/resume/cancel and warnings.
- Build summary dashboard with coverage and issue distributions.
- Build server-paginated page and issue tables with saved URL state.
- Build issue explanation/evidence and page-detail views.
- Build crawl history, comparison selection and comparison results.
- Build export/artifact controls and deletion confirmations.
- Add empty, loading, partial, limit-reached, recovered and failure states.
- Implement keyboard navigation and automated accessibility checks for primary flows.

Tests/gates:

- End-to-end tests cover first crawl, cancellation, recovery, issue review, comparison and export.
- The UI never displays a crawl as complete when a limit or systemic error stopped it.
- Tables remain responsive on synthetic large datasets.
- Primary workflows meet documented keyboard and contrast requirements.

### Milestone 6 — MCP and CLI completion

Estimated effort: 1–2 weeks  
Exit: an MCP client can safely conduct and interpret a bounded audit using the same services as the UI.

Tasks:

- Implement all required MCP tools from the PRD over stdio.
- Reject unknown fields and enforce bounded pagination/output sizes.
- Ensure read tools and mutation tools have distinct handlers and names.
- Prevent MCP access to private-target overrides, arbitrary headers, robots bypass, arbitrary output paths and raw filesystem operations.
- Add crawl polling guidance and structured progress responses.
- Complete equivalent CLI commands and JSON output modes.
- Add contract tests comparing UI/API-facing DTOs, CLI JSON and MCP results.
- Write MCP installation, configuration and tool-usage documentation.

Tests/gates:

- MCP protocol conformance and initialize/list/call flows pass.
- Tool schemas have descriptions, limits and stable errors.
- Duplicate `crawl_start` calls with an idempotency key create one crawl.
- Malicious tool arguments cannot reach private networks or arbitrary files.
- A scripted MCP scenario completes preview, crawl, summary, issue evidence and export.

### Milestone 7 — Optional JavaScript rendering

Estimated effort: 2–3 weeks  
Exit: selected crawls can safely extract rendered content with explicit coverage and resource costs.

Tasks:

- Resolve ADR-007 for each supported operating system.
- Implement the framed stdio protocol between the Go supervisor and TypeScript/Playwright render workers.
- Configure fresh contexts, capability restrictions and request interception.
- Route document and subresource URLs through target policy.
- Enforce navigation, request-count, byte, CPU/time and memory/recycle budgets.
- Add rendered DOM extraction and raw-versus-rendered provenance.
- Implement rendering failure taxonomy and fallback policy.
- Add settings, progress and result differences to CLI/API/UI.
- Add deterministic client-rendered fixture sites and hostile subresource fixtures.

Tests/gates:

- Rendered fixtures expose expected client-generated metadata/content/links.
- Attempts to access private URLs, download files, request permissions or persist browser state fail.
- Hung pages time out and workers recycle without stalling the crawl.
- Raw and rendered evidence cannot be confused in issues or exports.
- Renderer crashes are bounded by restart budgets and visible in coverage.

### Milestone 8 — Hardening, beta and release

Estimated effort: 2 weeks  
Exit: all PRD release criteria pass and a reproducible versioned release is available.

Tasks:

- Run full conformance, fuzz, adversarial, recovery, endurance and UI suites.
- Profile and tune database indexes, batches, caches and worker counts.
- Conduct manual threat-model review and remediate findings.
- Audit dependencies, licenses, secrets and release contents.
- Produce SBOM, checksums and signed artifacts where infrastructure permits.
- Validate clean installation on Linux, macOS and Windows.
- Write user guide, operator guide, MCP guide, security model, privacy/retention guide, rule catalog and troubleshooting guide.
- Create migration/backup/rollback procedures.
- Run a closed beta on varied authorised sites and triage false positives.
- Freeze schemas and rule IDs for version 1.
- Tag and publish release notes with known limitations.

Tests/gates:

- All version-1 acceptance criteria in the PRD pass.
- No unapproved high or critical security/dependency finding remains.
- 10k endurance crawl and 100k synthetic storage/query test meet documented thresholds.
- Upgrade from the preceding beta preserves projects and crawls.
- Release artifacts reproduce from the source tag and checksums verify.

## 4. Dependency graph

| Deliverable | Depends on |
|---|---|
| Durable frontier | Database spike, shared contracts, target policy |
| Sitemap/robots | Guarded fetch, URL/scope engine |
| Extraction | Bounded fetch records, contracts |
| Aggregate audit rules | Extraction and crawl/link/sitemap persistence |
| API/query layer | Database repositories and application services |
| UI | Versioned API and event contracts |
| MCP | Application services and shared schemas |
| Rendering | Target policy, extractor, worker supervision |
| Comparison/export | Stable page/issue data model |
| Release packaging | Core feature stability and platform tests |

Rendering can proceed after the raw extraction contract stabilises, but it must not bypass the target-policy workstream.

## 5. Initial backlog by priority

### P0 — Required for first usable alpha

- Go module, web workspace, and CI foundation.
- Secure URL/DNS/IP/redirect policy.
- Bounded HTTP fetcher.
- SQLite crawl/frontier persistence.
- Robots, sitemaps, rate limits and cancellation.
- Essential extraction and rules.
- CLI crawl/status/issues/export.

### P1 — Required for version 1

- Full Must-rule set.
- Loopback API security.
- UI primary workflows.
- Crawl comparison.
- CSV, NDJSON and XLSX reports.
- MCP stdio tools.
- Crash recovery and migration/backup.
- JavaScript rendering, provided its security gates pass.

### P2 — Candidate immediately after version 1

- Custom CSS/XPath extraction.
- Structured-data rich-result validation.
- Accessibility/Lighthouse integrations.
- Saved report views.
- Self-contained HTML reports.
- Signed desktop packaging improvements.

## 6. Quality plan

### 6.1 Coverage policy

Coverage percentages are secondary to risk coverage, but minimums provide regression signals:

- Fetch policy, URL scope, state machines and rule engine: 90% branch coverage.
- Extractors and database repositories: 85% branch coverage.
- Application services, API, CLI and MCP handlers: 80% branch coverage.
- UI: critical-path component and end-to-end coverage rather than a global percentage target.

Security-relevant branches require explicit tests even if aggregate thresholds already pass.

### 6.2 Required CI checks

- Go formatting, vet/static analysis, and frontend linting.
- Go and TypeScript type checking.
- Unit, race, fuzz-seed, and integration tests.
- Migration tests.
- API/MCP schema compatibility tests.
- UI build and primary end-to-end tests.
- Dependency and license audit.
- Secret and static-security scan.
- Generated documentation/rule catalog freshness.
- Platform smoke tests.

### 6.3 Performance baselines

Record hardware, runtime, profile, fixture, database size, peak RSS, pages/sec and p95 fetch/extract/write latency. Baselines should detect regressions rather than promise universal crawl speed.

## 7. Security review checklist

Before alpha, beta and release, confirm:

- Every outbound code path imports the central target policy.
- Every redirect and renderer request is revalidated.
- Prohibited IP tables cover IPv4, IPv6 and mapped forms.
- TLS verification cannot be disabled through API, CLI or MCP.
- API binds only to loopback and rejects hostile origins.
- MCP exposes no generic shell, filesystem, HTTP, SQL or browser primitive.
- Response, XML, archive, export and database sizes are bounded.
- Logs and errors redact sensitive headers and configured query fields.
- Artifact and database paths cannot be selected by callers.
- Spreadsheet exports mitigate formula injection.
- Browser workers have no persistent profiles or unnecessary permissions.
- Dependencies are pinned and release inputs are traceable.
- No auto-update, auto-install or self-restart endpoint exists.

## 8. Documentation deliverables

- Installation and update guide.
- Quick-start and first-audit tutorial.
- Crawl configuration reference.
- Rule catalog with evidence and limitations.
- UI, CLI and MCP references.
- Security architecture and threat model.
- Privacy, storage and retention guide.
- Responsible crawling policy.
- Backup, restore and migration guide.
- Troubleshooting and diagnostic-bundle guide.
- Third-party notices and reference-reuse ledger.

## 9. Definition of done for any feature

A feature is done when:

1. Its requirement and acceptance behaviour are clear.
2. Domain logic lives outside UI/API/MCP adapters.
3. Inputs and outputs have runtime-validated schemas.
4. Security boundaries and resource limits are enforced.
5. Success, failure, cancellation and recovery paths are tested where applicable.
6. Data migrations and indexes are included where required.
7. Logs are useful and redacted.
8. CLI/API/MCP/UI consistency is maintained for shared capabilities.
9. User and operator documentation is updated.
10. The change passes CI and review with no unresolved high-risk finding.

## 10. Project controls

- Use small vertical changes with tests; avoid a single large crawler implementation commit.
- Record material architectural choices as ADRs before dependent work expands.
- Keep rule IDs and public schema changes reviewed and versioned.
- Maintain a risk register with owner, mitigation and next review date.
- Demonstrate the working vertical slice at every milestone.
- Re-estimate after Milestones 0, 2, 5 and 7.
- Do not promote P2 work while a P0/P1 security or data-integrity gate is failing.

## 11. First implementation slice

The first code slice after approval should be deliberately small:

1. Initialise the monorepo and quality tooling.
2. Define Go URL, target-decision, fetch-result and safe-error types plus their public JSON schemas.
3. Implement prohibited IP classification with exhaustive tests.
4. Implement public-target DNS validation and pinned HTTP connection spike.
5. Expose one Go CLI command that fetches a public URL with strict limits and prints structured JSON.
6. Prove that loopback, private, link-local, metadata and public-to-private redirects are blocked.

This slice validates the most consequential security boundary before database, UI or audit-rule volume makes it expensive to change.

## 12. Optional post-version-1 scale research

The following stages document how the architecture could theoretically extend beyond current tested volumes. They are not v2.0 release gates, scheduled commitments or supported-capacity promises. Further work requires a separately approved scope.

### Scale Stage A — Segment-ready local engine

Included in version-1 architecture:

- no complete frontier or graph held only in memory;
- keyset pagination and streaming exports;
- bounded database batches and worker leases;
- checkpointable crawl counters and host state;
- page-local rules independent of total crawl size;
- explicit global versus segment limits.

### Scale Stage B — Million-URL local campaigns

- Add automatic segment continuation and compaction.
- Add disk forecasts, quotas and safe-pause thresholds.
- Validate incremental aggregates and deferred whole-crawl rules.
- Tune URL/link partitions and indexes at 1 million, then 5 million and 10 million URLs.
- Add campaign ETA, segment history and recovery diagnostics.
- Publish repeatable performance reports for supported hardware profiles.

### Scale Stage C — Distributed campaign prototype

- Separate coordinator, fetch workers, analysis workers and artifact services.
- Add durable distributed leases and idempotent result commits.
- Partition URL and link data while preserving global identity.
- Assign one logical politeness owner per host.
- Store immutable result segments in object storage.
- Evaluate PostgreSQL or a purpose-built operational frontier and ClickHouse-class analytical storage.
- Prove worker loss, coordinator failover, replay and deduplication semantics.

### Scale Stage D — Optional 100M research protocol

- Complete more than 100 million unique guarded fetches in one campaign.
- Retain the required audit field set and global link relationships.
- Pass known-result extraction and rule samples throughout the dataset.
- Demonstrate controlled crash/restart recovery without lost or duplicate committed URLs.
- Publish hardware, configuration, duration, request rate, memory, storage and failure counts.
- Publish the benchmark generator, expected invariants and verification tools.
- Repeat the benchmark on the release candidate advertised as supporting 100M+.

Stage D is deferred indefinitely and is not required for v2.0. Approved public language is limited to: “Architecturally designed for segmented campaigns beyond 100 million URLs; this is a theoretical scalability target, not a tested, supported or guaranteed capacity.”

## 13. Requirement-to-milestone traceability

| Requirement group | Primary delivery | Verification/final gate |
|---|---|---|
| PRJ-01–PRJ-03 | Milestones 2 and 5 | Milestone 8 end-to-end project/profile flows |
| PRJ-04–PRJ-05 | Milestones 4 and 5 | Milestone 8 profile round-trip and multi-profile tests |
| CRW-01–CRW-11 | Milestones 1 and 2 | Milestone 8 conformance, recovery and endurance suites |
| CRW-12–CRW-13 | Milestone 7 | Milestone 8 renderer security and evidence tests |
| CRW-14 | Milestone 2 | Milestone 8 list-mode scope and order tests |
| CRW-15 | Not delivered; explicit version-one Won't | PRD non-goal remains documented |
| EXT-01–EXT-06 | Milestone 3 | Milestone 8 extraction fixture corpus |
| EXT-07–EXT-08 | Milestone 3 | Milestone 8 extraction fixture corpus |
| EXT-09 | P2 after version one unless promoted | Revised PRD required if promoted |
| EXT-10 | P2 after version one | Revised PRD required if promoted |
| AUD-01–AUD-12 | Milestone 3 | Milestone 8 rule conformance suite |
| AUD-13–AUD-14 | Milestones 3 and 8 as Should scope | Rule fixtures and versioned validation data |
| AUD-15–AUD-16 | P2 after version one | Revised PRD required if promoted |
| REP-01–REP-05 | Milestones 4 and 5 | Milestone 8 query/export/UI end-to-end tests |
| REP-06–REP-07 | Milestones 4 and 5 | Milestone 8 XLSX safety and comparison tests |
| REP-08–REP-09 | P2 after version one | Revised PRD required if promoted |
| INT-01–INT-02 | Milestones 2, 4 and 6 | Milestone 8 API/CLI contract tests |
| INT-03–INT-07 | Milestone 6 | Milestone 8 MCP conformance and adversarial tests |
| INT-08 | Milestone 4 | Generated OpenAPI freshness check |
| Security/privacy requirements | Milestones 0, 1, 4, 7 and 8 | Security checklist plus no high/critical release finding |
| Theoretical 100M+ architecture direction | Optional post-version-one research | Not a release gate or supported-capacity claim |

## 14. Milestone approval records

Each milestone closes with a short record containing:

- delivered scope and requirement IDs;
- test, security and performance evidence;
- known limitations and deferred work;
- schema, migration and compatibility notes;
- risks added, closed or accepted;
- decision to proceed, repeat, narrow or revise the baseline.

Milestone dates and estimates may change without revising the product plan. Scope, acceptance criteria, security boundaries and supported-capacity claims may not.
