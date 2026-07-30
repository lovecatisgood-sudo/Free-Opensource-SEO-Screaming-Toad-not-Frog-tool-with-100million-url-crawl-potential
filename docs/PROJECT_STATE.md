# Project state checkpoint

Checkpoint date: 2026-07-30 (Asia/Ho_Chi_Minh)  
Branch: `main`  
Previous checkpoint commit: `ec23406`

## Current outcome

SEO Auditor's locally executable v2.0 implementation is complete as release candidate `2.0.0-rc.4`. The repository contains the approved PRD, architecture and implementation plan plus the implemented Go crawler, SQLite persistence, raw and isolated rendered audits, web UI, local API, CLI, MCP server, reports, recovery, scale prototypes and release tooling.

Do not describe the product as a stable v2.0 release yet. External release qualification remains blocked on additional closed-beta SEO sampling, clean-machine macOS and Windows runtime tests, and a release-signing identity.

### Professional-quality implementation checkpoint

The quality roadmap is now substantially implemented in the working tree. The deterministic conformance baseline covers AUD-01 through AUD-17 with 21 expected findings, zero unexpected/missing findings, and precision/recall 1.0 for represented cases. New audit depth includes Schema.org 30.0 vocabulary checks, pinned Google Breadcrumb/Article/merchant Product profiles, bounded CSS/XPath custom audits, mobile/AMP signals, image-delivery checks and bounded PDF readiness diagnostics.

Rendered audits can retain redacted DOM and opt-in screenshots as managed expiring artifacts, store console/resource failures, and run pinned axe-core 4.11.0 checks with explicit automated-coverage limitations. Optional PageSpeed Insights, CrUX, Search Console and GA4 adapters preserve lab/field/external evidence separately. Transport errors are redacted, OAuth credentials can refresh and rotate inside the platform credential store, and no credential value is stored in profiles, SQLite crawl data, exports, diagnostics or MCP responses.

Authenticated raw crawling now supports host-bound bearer, basic and cookie credentials through opaque OS secret references. Redirects cannot carry credentials outside the configured host/subdomain boundary. Authenticated rendered crawling remains deliberately rejected until equivalent browser-boundary mediation is implemented and qualified.

The dashboard and API now include integration observations, recurring audits and a bounded internal-link architecture view with segment, depth, inlink/outlink, score and orphan evidence. Scheduled audits reuse stored guarded profiles and run only while the local application is open. The MCP surface contains 36 bounded tools and accepts credential references—but never credential values—for authenticated profiles and provider queries.

Release automation now includes a tag/manual candidate workflow that qualifies Go/TypeScript/conformance gates, builds deterministic cross-platform archives, generates CycloneDX SBOM and checksums, creates GitHub artifact attestations with OIDC, verifies downloaded artifacts and prepares a protected draft prerelease. Local Bash and PowerShell checksum/attestation verification scripts are included. Native Apple Developer ID/notarization and Windows Authenticode cannot be completed until DJAI supplies organization-owned identities; the workflow therefore remains explicitly a release-candidate path, not proof of a signed stable release.

Repeated 5–10-million campaigns and live 100M qualification were removed from v2.0 scope by product decision. Approved language is: “Architecturally designed for segmented campaigns beyond 100 million URLs; this is a theoretical scalability target, not a tested, supported or guaranteed capacity.”

Product development is now quality-first rather than capacity-first. `docs/QUALITY_ROADMAP.md` defines the ordered audit-depth programme. Its claim boundary keeps URL capacity separate from audit quality and prohibits parity/superiority claims until a published fixture suite supports them.

`docs/COMMERCIAL_QUALITY_PLAN.md` now turns that direction into an execution programme. It separates stable-release qualification from longer capability work, defines a public fixture/conformance laboratory, measurable accuracy and beta gates, a 20-audit authorised-site matrix, clean-machine and signing requirements, a differential Screaming Frog 24.3 benchmark, six delivery waves and a first-30-days sequence. The plan treats transparent rules, reproducible evidence and the agent-assisted fix/recrawl loop as the intended differentiation; MCP alone is not represented as unique because Screaming Frog 24 added MCP support.

Two companion execution documents now complete that planning set. `docs/PROFESSIONAL_PRODUCT_PLAN.md` defines the professional audiences, product promise, pillars, end-to-end journeys, release scope from stable 2.0 through 3.0, functional requirements, success metrics, open-source sustainability model, launch sequence and maintainer decisions. `docs/PROFESSIONAL_IMPLEMENTATION_PLAN.md` maps that product plan into ten workstreams, dependency order, seven phased releases, migrations, common surface contracts, security/performance/platform testing, CI gates, definition of done and the first four two-week iterations.

The Phase 0 conformance foundation is implemented. `internal/conformance` defines versioned manifests, expected findings/absences, deterministic comparison, stable reports, precision/recall and rule-family baseline coverage. `internal/testfixtures/sites` serves deterministic local fixtures, and `cmd/seo-auditor-conformance` produces Markdown or JSON and fails non-zero on drift. The `core-rules.v1` fixture has a clean control plus positive cases across every AUD-01 through AUD-17 family. This is baseline fixture coverage, not complete rule validation or parity evidence.

Finding classification and evidence source are now first-class data. Migration 009 adds checked/indexed `classification` and `evidence_source` columns and backfills legacy issues. New raw/rendered findings write provenance directly; finalised aggregate issues are normalised to raw, rendered, graph or sitemap evidence. Database queries, reports, API/MCP JSON and the dashboard issue table/explanation drawer expose both fields. `docs/CONFORMANCE.md` documents the contract and remaining boundary, malformed, graph, rendered and adversarial fixture work.

The structured-data slice is implemented in the working tree. AUD-13 separates syntax, structure, context and vocabulary diagnostics using a checksum-pinned Schema.org 30.0 registry (933 types and 1,521 properties). AUD-14 adds separately versioned Breadcrumb, Article and merchant Product profile diagnostics from Google Search Central documentation reviewed 2026-07-30. Extraction retains bounded, sorted evidence and per-node shapes without storing arbitrary values. These diagnostics do not promise rich-result eligibility or display.

The public interface is now branded **SEO Screaming Toad — Not Frog**, with **DJAI Toad** as an alternate community name. The React dashboard uses the supplied toad and DJAI artwork in a dense technical workbench with project navigation, crawl configuration, progress, evidence tables, comparisons and a separate DJAI services rail. The public capacity phrase remains explicitly theoretical.

The stdio MCP server is agent-operable from a fresh workspace. Beyond the original required tools, it now exposes simplified profile creation/listing, page and link listing, crawl timeline, pause/resume and metadata-only diagnostics. MCP tool annotations label read-only, additive and open-world operations; `health_get` verifies the local API connection. Configuration and workflow guidance are in `docs/MCP.md`.

The public repository README now includes eight dashboard screenshots captured from a real local DJAI Academy verification crawl, an expert-oriented product tour, DJAI Academy trainer/community co-creation credit, service links and the Siamese Cat Cafe link. During capture, a raw-only summary compatibility bug was found and fixed: the dashboard now safely handles an absent rendering-status distribution.

The dashboard footer now carries the approved attribution: **Siamese Cat Dev from DJAI Academy & With DJAI Community**. Siamese Cat Dev links to `https://github.com/lovecatisgood-sudo`, DJAI Academy to `https://djai.academy`, and DJAI Community to `https://school.djai.academy`. A separate **Support this project on GitHub ★** call-to-action links to the public SEO Screaming Toad repository. The generated embedded web assets include this footer.

The public repository's `main` branch is protected. Pull requests must be current, all five existing CI checks must pass, and conversations must be resolved; the rules apply to administrators and prohibit force pushes and deletion. Required external approval remains zero so the solo maintainer can merge after the automated gates pass.

## Course-authoring state

An English and a separately authored Thai community course now exist under `docs/course/`. Both use five modules and 12 lesson folders, with a quiz after every module and a final 10-question quiz. The core course is designed for a fast-paced 60-minute delivery. An optional real-world DJAI Academy audit lesson sits outside the core hour.

English course root:

- `docs/course/Course 4: SEO & ASO with AI Agent - top rank your site on search/`
- `COURSE_OVERVIEW.md` records the timing, learning outcome and claim guardrails.
- `SCREENSHOT_PLAN.md` maps 26 planned core-course visual assets: eight existing dashboard captures and 18 planned new captures. The planned local Screaming Frog comparison uses installed SEO Spider 24.3, the same authorized target, comparable raw scope and a maximum of 500 URLs. These additional core-course captures have not yet been taken.
- Each of the 12 lessons has its own folder and `SCRIPT.md`; every module has `MODULE_QUIZ.md`; `FINAL_QUIZ.md` contains the final assessment.

Thai course root:

- `docs/course/Thai Course 4 - SEO & ASO with AI Agent - ดันเว็บไซต์ให้เด่นบน Search/`
- The Thai scripts are original Thai teaching copy based only on the course outline, not sentence-by-sentence translations. They use a natural male-instructor voice, Thai production directions and the technical English terms students will see in the product.
- It contains 12 lesson folders, five Thai module quizzes, a Thai final quiz and an original Thai version of the real-world bonus lesson.
- The Thai bonus references the verified English-course screenshots and workbook instead of duplicating binary assets.

Course-wide claim boundaries remain explicit: “top rank” is an aspiration rather than a guarantee; Toad covers much of the core normal technical-audit workflow but does not claim complete Screaming Frog parity; 100M+ remains theoretical segmented-architecture potential.

## Course real-world DJAI audit

Target: `https://www.djai.academy/`

Project: `DJAI Academy course reality audit 2026-07-30`

Profile: raw HTML, same host only, compression disabled, 100,000-URL ceiling

Crawl ID: `crawl_425921818cae33f4a03a417f7880cd3e`

The authorized course crawl completed in about 31 seconds after the same-host frontier emptied. The 100,000 setting was only a safety ceiling; the achieved crawl size was 287 discovered/fetched resources, 222 analysed HTML pages, zero failures and only 2xx fetched responses. The sitemap was available and recorded as `ok`.

Rule totals were 8 errors, 66 warnings and 20 informational observations, for 94 findings:

- AUD-05: eight 200-status blog category-filter URLs with `noindex`;
- AUD-02: 42 warnings and 13 informational metadata observations;
- AUD-08: 24 near-duplicate similarity observations concentrated in related tool families;
- AUD-03: seven informational title/H1 equality observations.

Practical review deliberately differs from raw severity. All eight AUD-05 URLs used `noindex, follow`, canonicalized to the appropriate Thai or English blog hub, and were omitted from the sitemap. This is consistent with intentional duplicate-control for interactive filters unless the product intends category pages to rank. No confirmed high-priority technical failure was found in this raw crawl. The most useful medium-priority review is whether the 24 similar QR/image/PDF/media tool pages are sufficiently differentiated for their distinct intents. Metadata-length observations are lower priority and their thresholds are editorial diagnostics rather than ranking rules.

Published course evidence:

- English bonus lesson: `Bonus Lesson - Real World DJAI Academy Audit/SCRIPT.md`;
- evidence-backed Markdown audit: `Bonus Lesson - Real World DJAI Academy Audit/AUDIT_REPORT.md`;
- seven genuine dashboard captures under its `ASSETS/screenshots/` folder;
- managed XLSX workbook: `ASSETS/reports/djai-academy-live-audit-2026-07-30.xlsx`;
- workbook SHA-256: `0255dbd3f7dab4fe32add3777ce122cff2bce3c31a59affac43f4c7ef82ae44b`;
- reproducible capture helper: `scripts/capture-course-bonus.mjs`.

The live course database remains ignored at `.data/course-djai-live/seo-auditor.db`. The isolated supervisor used loopback port 7342 and was shut down cleanly after capture. No automated crawls were run against Bangkok Post, Thai Rath or Major Cineplex because third-party authorization was not established.

## Verification evidence

- Full Go test suite and `go vet` passed.
- The MCP schema, lifecycle scenario, authenticated loopback API integration and network-disabled test profile passed; a production `seo-auditor-mcp` binary also built successfully.
- TypeScript tests, type checking and production builds passed.
- Opt-in real Chromium mediation, hung-page and crash-boundary tests passed.
- RC4 was built from `5213ad6990574abc65092bed9bc8dbaa80847eee` for Linux amd64/arm64, macOS amd64/arm64 and Windows amd64.
- RC4 includes a CycloneDX SBOM and verified `SHA256SUMS`.
- RC4 is unsigned.
- Release artifacts are local at `.artifacts/release/2.0.0-rc.4/`.
- The Git worktree was clean before this checkpoint was added.
- After the quality-first scope change, focused extractor/rules/database/renderer tests, the full Go suite, the network-disabled full Go suite, TypeScript tests and TypeScript production builds passed in the development container.
- The offline gate exposed and fixed a pre-existing real-DNS dependency in the list-mode test. Application services now have a test-only resolver seam while production continues to default to the guarded system resolver.
- The branded React workbench passed application type-checking, frontend tests, production compilation, the full embedded-UI Go suite, and visual inspection at 1600×1000 and 430×900 viewports.
- The updated attribution/footer passed all three frontend tests, TypeScript type checking, production compilation and an embedded-bundle content check. Go was not installed in the current shell, so no new Go test run was performed for this footer-only checkpoint.
- For the course audit, the repository-local Go toolchain successfully built `seo-auditor`, `seo-auditor-cli` and `seo-auditor-mcp` with workspace-local caches.
- The completed course crawl, summary, issue evidence, lifecycle and workbook export were read back through the local API/CLI. The workbook passed ZIP/XLSX archive integrity validation and its checksum matches the application artifact metadata.
- All seven bonus screenshots were visually inspected. The capture helper passed `node --check`; its final crop path does not mutate dashboard content or bypass the application's content-security policy.
- The Thai structure check found exactly 12 numbered lesson folders and five module-quiz files, and its shared English bonus assets resolve to existing files.
- No full Go or frontend test suite was rerun for the course-only documentation and generated-asset additions.
- During commercial-quality planning, the repository-local Go 1.26.5 toolchain was confirmed and the full Go suite passed with workspace-local caches and loopback fixture sockets enabled. Both TypeScript workspaces passed all 18 tests and type checking by invoking their pinned local binaries. The earlier footer-checkpoint note that Go was unavailable describes that prior shell only and is not the current environment state.
- The professional implementation pass completed full Go tests and vet, frontend and renderer type checks/tests/production builds, and the 17-family conformance run (21 true positives, zero false positives and zero false negatives for represented cases). Opt-in real Chromium mediation, screenshot, axe, state-isolation and hung-page tests passed.
- A `2.0.0-rc.quality` release dry run cross-compiled all three Go executables for Linux amd64/arm64, macOS amd64/arm64 and Windows amd64. Its SBOM, relative SHA-256 manifest and every packaged file verified locally. This proves cross-compilation and package integrity only; macOS/Windows runtime trust and native signing remain external gates.

## Scale state

The 1,000,000- and 5,000,000-URL synthetic campaigns passed on the production persistence, extraction, rules, graph, recovery and verification paths. These runs did not use live guarded network fetching.

Final 5M reconciliation:

- status: `completed`;
- discovered, committed pages and unique URL identities: 5,000,000 each;
- outstanding work and missing required fields: 0;
- links: 4,999,999;
- issues: 15,884,167;
- completed/invalid segments: 50/0;
- SQLite integrity: `ok`;
- final checkpointed storage: 11,127,939,072 bytes.

The run exposed and fixed immature storage projection and unbounded near-duplicate candidate generation. A read-only `seo-auditor-scale verify` command now reproduces campaign reconciliation. Full evidence is in `docs/benchmarks/2026-07-30-segmented-campaign.md`.

The Stage C coordinator and immutable-segment prototypes are implemented. Additional Stage B/D campaigns are optional research rather than release gates.

## DJAI Academy live screening

Target: `https://djai.academy`  
Canonical crawl seed: `https://www.djai.academy/`  
Crawl ID: `crawl_56d3870acb78b7b8222e189f8007c156`

The bounded raw audit completed with 285 fetched URLs, 220 analysed HTML pages, zero failures and only 2xx fetched responses. It recorded 75 warnings and 28 informational observations:

- 24 non-reciprocal hreflang observations on eight category-filter URLs;
- 20 duplicate title/description observations on the main and filtered blog views;
- 18 long and 4 short titles;
- 7 long and 13 short meta descriptions;
- 2 near-duplicate template pairs;
- 7 informational title/H1 matches;
- 8 filtered blog URLs intentionally absent from the sitemap.

Local ignored evidence:

- database: `.data/live-djai-www/seo-auditor.db`;
- issue export: `.data/live-djai-www/artifacts/artifact_725cc7fb159cb6db7c12e65a56c2adbb.csv`;
- page export: `.data/live-djai-www/artifacts/artifact_77ccb6cef0cdeb254a651091f2d3955d.csv`.

Hostinger returned HTTP 403 whenever the client explicitly sent an `Accept-Encoding` header, including `gzip` and `identity`; the same requests returned 200 when that header was omitted. SEO Auditor now provides an explicit `gzip|disabled` response-compression profile setting in the UI, CLI, API and stored configuration. Disabled mode omits the header while preserving robots, scope, DNS/IP, redirect, byte and politeness controls; it does not retry or bypass genuine 403 responses. A live 10-URL regression completed 10/10 pages with zero failures and a clean `limit_reached` terminal state.

## Resume order

1. Read `docs/PRD.md`, `docs/ARCHITECTURE.md`, `docs/IMPLEMENTATION_PLAN.md`, `docs/QUALITY_ROADMAP.md`, `docs/COMMERCIAL_QUALITY_PLAN.md`, `docs/PROFESSIONAL_PRODUCT_PLAN.md`, `docs/PROFESSIONAL_IMPLEMENTATION_PLAN.md`, this checkpoint, both course overview files and the segmented campaign report.
2. Review the uncommitted English/Thai course scripts, seven bonus screenshots, workbook and capture helper before staging or committing them.
3. If continuing course production, capture only the remaining items in the English `SCREENSHOT_PLAN.md`; do not recapture or duplicate the completed bonus assets without a reason.
4. Obtain more authorized beta targets and retain comparison evidence. Do not run large crawls against third-party sites without explicit owner authorization and an approved rate/scope/window.
5. Obtain clean macOS/Windows environments and organisation-owned Apple/Windows signing identities before promoting a candidate to stable.
6. Expand the deliberately partial expert capabilities only with reviewed fixtures: additional Google search-feature profiles, full PDF parsing, perceptual image analysis, Search Console URL Inspection/sitemap enrichment and local Lighthouse isolation.
7. Treat any additional multi-million or distributed campaigns as separately approved optional research.

Current Git state at this checkpoint: local `main` is one commit ahead of `origin/main` at `ec23406`; the course tree, capture helper, commercial-quality plan, professional product/implementation plans, documentation index and this state update are uncommitted. The previously committed footer work remains clean. Because `main` is protected, publish new work through a branch and pull request rather than a direct push.

The ignored `open-seo-crawler/` reference remains untrusted and must not be installed, executed or copied. SEonaut was used only as design inspiration; project source remains original.
