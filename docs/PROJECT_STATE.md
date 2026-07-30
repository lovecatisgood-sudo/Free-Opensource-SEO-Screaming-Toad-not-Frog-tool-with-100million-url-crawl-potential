# Project state checkpoint

Checkpoint date: 2026-07-30 (Asia/Ho_Chi_Minh)  
Branch: `main`  
Previous checkpoint commit: `f96d557`

## Current outcome

SEO Auditor's locally executable v2.0 implementation is complete as release candidate `2.0.0-rc.4`. The repository contains the approved PRD, architecture and implementation plan plus the implemented Go crawler, SQLite persistence, raw and isolated rendered audits, web UI, local API, CLI, MCP server, reports, recovery, scale prototypes and release tooling.

Do not describe the product as a stable v2.0 release yet. External release qualification remains blocked on additional closed-beta SEO sampling, clean-machine macOS and Windows runtime tests, and a release-signing identity.

Repeated 5–10-million campaigns and live 100M qualification were removed from v2.0 scope by product decision. Approved language is: “Architecturally designed for segmented campaigns beyond 100 million URLs; this is a theoretical scalability target, not a tested, supported or guaranteed capacity.”

Product development is now quality-first rather than capacity-first. `docs/QUALITY_ROADMAP.md` defines the ordered audit-depth programme. Its claim boundary keeps URL capacity separate from audit quality and prohibits parity/superiority claims until a published fixture suite supports them.

The first Q1 slice is implemented in the working tree: AUD-13 detects malformed JSON-LD, structurally invalid `@type`/`@context` values, and compact structured-data types without an observed Schema.org context. Extraction now retains bounded, sorted type/property/context evidence with an explicit truncation flag. Full Schema.org vocabulary and Google rich-result profile validation remain separate Q1 work; current AUD-13 findings do not imply either form of validation.

The public interface is now branded **SEO Screaming Toad — Not Frog**, with **DJAI Toad** as an alternate community name. The React dashboard uses the supplied toad and DJAI artwork in a dense technical workbench with project navigation, crawl configuration, progress, evidence tables, comparisons and a separate DJAI services rail. The public capacity phrase remains explicitly theoretical.

## Verification evidence

- Full Go test suite and `go vet` passed.
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

1. Read `docs/PRD.md`, `docs/ARCHITECTURE.md`, `docs/IMPLEMENTATION_PLAN.md`, `docs/QUALITY_ROADMAP.md`, this checkpoint and the segmented campaign report.
2. Confirm `git status --short --branch` is clean and start the capability-free development container.
3. Continue Q1 with a provenance-recorded, versioned Schema.org vocabulary snapshot, then implement AUD-14 as separate Google search-feature profiles beginning with Breadcrumb, Article and Product.
4. Obtain more authorized beta targets and retain comparison evidence.
5. Obtain clean macOS/Windows environments and a signing identity before promoting RC4 to stable.
6. Treat any additional multi-million or distributed campaigns as separately approved optional research.

The ignored `open-seo-crawler/` reference remains untrusted and must not be installed, executed or copied. SEonaut was used only as design inspiration; project source remains original.
