# Professional implementation plan

Status: implementation baseline substantially delivered; external release gates and explicitly partial capabilities remain
Date: 2026-07-30
Delivery scope: stable 2.0 through 3.0
Depends on: `PROFESSIONAL_PRODUCT_PLAN.md` and `COMMERCIAL_QUALITY_PLAN.md`

## 1. Implementation outcome

Deliver the professional product in independently releasable increments while preserving the current Go scheduling/security authority, isolated browser boundary, local-first storage and common UI/CLI/API/MCP domain services.

Every increment must include:

```text
Contract → Fixture → Implementation → UI/API/MCP → Documentation → Qualification evidence
```

Feature implementation without its fixtures, limitations and release evidence is incomplete.

## 2. Delivery rules

1. Protect the outbound network boundary: only the guarded fetch/render adapters may contact crawl targets.
2. Keep raw, rendered, crawl-graph, API, lab and field evidence distinct in contracts and storage.
3. Version rule metadata, evidence schemas, external profiles and engine manifests.
4. Add migrations before application code depends on new stored data; migration tests cover legacy databases.
5. Keep database writes behind the current writer/application-service boundaries.
6. Do not store credentials in crawl profiles, SQLite crawl data, backups, exports, diagnostics or MCP responses.
7. Add all professional capabilities through the same application services used by UI, CLI, API and MCP.
8. Enforce evidence, time, memory, match-count, disk and concurrency budgets at contract boundaries.
9. Prefer no finding over a false positive when rule data is incomplete.
10. Do not publish release or comparison claims until their evidence job passes.

## 3. Workstreams

| Workstream | Responsibility | Principal paths |
|---|---|---|
| WS1 Conformance | Fixture sites, manifests, expected findings, coverage and differential reports | `internal/conformance`, `internal/testfixtures/sites`, `docs/conformance` |
| WS2 Rule data | Schema.org and search-feature profiles, provenance and update checks | `internal/ruledata`, `internal/rules` |
| WS3 Custom audits | Safe selectors, conditions, saved definitions and outputs | `internal/customaudit`, contracts/database/application |
| WS4 Render quality | DOM/screenshots, raw/rendered diffs, console/resources | `web/renderer`, `internal/renderer`, artifacts/database |
| WS5 Accessibility/performance | axe, Lighthouse, PageSpeed and CrUX adapters | new isolated workers and application services |
| WS6 Search integrations | Secret store, Search Console and GA4 | `internal/secretstore`, `internal/integrations` |
| WS7 Media/auth | Authenticated sessions, PDF/image/AMP/mobile/pagination | fetch policy, extractor and new media packages |
| WS8 Architecture/workflow | Link scores, segments, visualisation, schedules and regressions | database, reports, web app, MCP |
| WS9 Release engineering | Native packages, signing, attestations and clean-machine qualification | `.github/workflows`, `scripts`, `docs/release-evidence` |
| WS10 Beta operations | Authorisation, sampling, adjudication and case studies | `docs/beta`, managed private evidence |

## 4. Dependency order

```text
Conformance contracts
    ├── Core release qualification ──> signed 2.0
    ├── Rule-data registry ──────────> structured data
    ├── Custom-audit contracts ──────> custom audits
    └── Evidence-source model
            ├── rendered retention ──> axe + Lighthouse
            ├── secret store ────────> GSC + GA4 + auth
            └── artifact retention ──> PDFs + screenshots

Stable evidence model
    └── architecture metrics + scheduling + agent verification ──> 3.0
```

The conformance manifest, evidence-source model and release identity are critical-path decisions. Search integrations cannot start production implementation before the secret-store contract is approved. Authenticated crawling cannot ship before safe-method and mutation tests pass.

## 5. Phase 0 — repository and conformance foundation

Target: weeks 1–3
Release: 2.0 stable candidate

### 5.1 Contracts

- Define `FixtureCase`, `ExpectedFinding`, `ExpectedAbsence`, `ExpectedCoverage` and `ConformanceRun` schemas.
- Define canonical ordering and stable identity for finding comparison.
- Define classification enum: deterministic, recommendation, review and information.
- Define evidence-source enum: raw, rendered, graph, sitemap, external API, lab and field.
- Add schema-version migration policy for fixture manifests.

### 5.2 Fixture server

- Build deterministic in-process sites for responses/redirects, robots, sitemaps, canonicals, indexability, metadata, headings, links, duplicates, hreflang and raw/rendered differences.
- Support stable route names and deterministic timestamps/content.
- Make JavaScript fixtures use the existing isolated renderer protocol.
- Prevent fixture tests from requiring public DNS or internet access.

### 5.3 Conformance runner

- Start a fixture project with an immutable profile.
- Run the crawler to a known terminal state.
- Compare actual findings, explicit absences and coverage with the manifest.
- Produce JSON and human-readable HTML/Markdown summaries.
- Calculate precision/recall only for correctly classified deterministic cases.
- Fail CI on drift, missing evidence version or unstable order.

### 5.4 Coverage catalog

- Map every public rule/filter to supported, partial or unsupported cases.
- Link each supported claim to fixture IDs.
- Render the matrix into documentation and a release artifact.

### 5.5 Acceptance

- All current AUD-01 through AUD-13 families have positive, negative, malformed and non-applicable cases where meaningful.
- Ten repeated local runs produce identical semantic output.
- The full offline Go/TypeScript suites pass.
- No product capability appears as supported without a fixture link.

## 6. Phase 1 — stable 2.0 release qualification

Target: weeks 2–8, overlapping Phase 0
Release: 2.0 stable

### 6.1 Release workflow

- Add a tag-triggered GitHub Actions workflow with minimal permissions.
- Build UI and Go binaries from a clean protected commit.
- Run tests, vet, type checking, production builds and conformance suite.
- Generate CycloneDX SBOM and relative checksums.
- Generate GitHub artifact attestations and keyless checksum signature.
- Build/sign/notarise macOS artifacts on a native runner.
- Build and Authenticode-sign Windows artifacts on a native runner.
- Create a draft release, require protected-environment approval, then run post-download verification.

### 6.2 Packaging

- Decide archive/installer formats per platform.
- Bundle only required binaries, documentation, examples and renderer dependencies.
- Make rendered-mode browser installation explicit and testable.
- Add `verify-release` scripts for checksum, signature, attestation and version output.
- Document uninstall and local-data retention paths.

### 6.3 Clean-machine qualification

- Automate smoke journeys where possible: launch, health, project/profile, small fixture crawl, issue query, export, MCP health and shutdown.
- Manually verify first-launch prompts, non-admin use, spaces/Unicode paths, firewall behavior, sleep/resume and application removal.
- Record OS build, architecture, artifact checksum and observed exceptions.

### 6.4 Closed beta

- Create authorisation and audit-profile templates.
- Recruit at least 12 owners covering the agreed site matrix.
- Run 20 audits with conservative site-specific limits.
- Sample findings and clean pages per the commercial quality plan.
- Adjudicate disputed high-impact findings with two SEO reviewers.
- Fix material crawler/rule defects and rerun affected audits.
- Store only anonymised or explicitly publishable evidence in the repository.

### 6.5 Differential benchmark

- Run Toad and Screaming Frog 24.3 against owned fixture estates with documented equivalent settings.
- Normalise URL identities and compare crawl/indexability/graph outcomes.
- Review each difference against fixture intent and primary sources.
- Publish aggregate results, versions, settings and limitations.

### 6.6 Acceptance

- Deterministic accuracy gates pass.
- No unresolved high/critical security, integrity or release risk.
- Twenty beta audits and required samples are dispositioned.
- All declared OS artifacts pass clean-machine tests.
- Checksums, signatures, native trust and attestations verify after download.
- A formal GitHub 2.0 release exists with honest known limitations.

## 7. Phase 2 — structured data and custom audits

Target: 6–9 weeks after stable foundation
Release: 2.1

### 7.1 Versioned rule-data registry

- Define signed or checksum-pinned immutable data bundles.
- Store source URL, retrieved date, source version/licence, transform version and bundle checksum.
- Load only bundled reviewed versions at runtime.
- Add a generator that creates reviewable diffs; never update rules during a crawl.
- Expose active bundle versions in diagnostics and every applicable finding.

### 7.2 Schema.org validation

- Extend extracted statements for JSON-LD, microdata and RDFa without unbounded script/body retention.
- Validate type/property existence, case, superseded vocabulary and relevant domain/range relationships.
- Keep parser, structural and vocabulary findings separate.
- Add fixtures for graphs, nested objects, arrays, references, pending vocabulary and malformed formats.

### 7.3 Google search-feature profiles

- Create profile contracts for applicability, required/recommended properties and nested types.
- Implement Breadcrumb, Article and Product first.
- Add visible-content consistency checks only where evidence is available and deterministic.
- Label eligibility diagnostics without promising rich-result display.
- Expand profile coverage through separately reviewed data updates.

### 7.4 Custom extraction engine

- Implement bounded CSS and XPath selectors over existing parsed documents.
- Add text/regex matching with a safe engine or explicit time/work budget.
- Add typed extraction values and raw/rendered mode selection.
- Add conditions and custom finding definitions without arbitrary code.
- Persist definitions by schema version; export/import non-secret JSON.
- Store only bounded matches and explicit truncation evidence.

### 7.5 Surface integration

- Add application services and contracts before UI/API/MCP adapters.
- Add configuration and results screens with selector preview.
- Add columns/sheets to reports without breaking older exports.
- Add MCP tools for listing definitions/results and running only pre-approved bounded definitions.

### 7.6 Acceptance

- All structured-data formats and initial Google profiles pass positive/negative fixtures.
- Adversarial selectors/regex and oversized matches stay inside budgets.
- UI, export, API and MCP return identical versioned results.
- Existing crawl profiles/databases migrate and reopen successfully.

## 8. Phase 3 — rendered evidence, accessibility and performance

Target: 7–10 weeks
Release: 2.2

### 8.1 Artifact retention model

- Add retention policy fields: enabled types, maximum per-page bytes, crawl bytes and retention duration.
- Store DOM/screenshots as managed opaque artifacts, not arbitrary paths.
- Record checksum, MIME type, viewport, engine version and page/crawl IDs.
- Add deletion, expiry and orphan-artifact reconciliation.
- Add selector/query/header/form redaction before persistence.

### 8.2 Rendered diagnostics

- Capture bounded final DOM, screenshots, console messages and failed/blocked resources.
- Compute raw/rendered differences for critical head fields, headings, links, schema and visible text.
- Add explicit partial/failed rendering states and do not silently substitute raw data.

### 8.3 axe worker

- Pin the axe version in the renderer lockfile and engine manifest.
- Run inside the current isolated browser boundary.
- Store rule, impact, WCAG tags, target selector and bounded evidence.
- Map findings into a separate accessibility category with automated-coverage limitations.

### 8.4 Lighthouse worker

- Create a separate isolated worker/profile contract.
- Pin engine/runtime and define mobile/desktop lab profiles.
- Store environment, throttling, timestamp and variability evidence.
- Use representative template sampling and explicit per-crawl run ceilings.

### 8.5 PageSpeed/CrUX adapter

- Implement optional API requests behind user-supplied credential references and quotas.
- Store field freshness and origin/URL scope.
- Keep field data separate from lab runs in database and reports.

### 8.6 Acceptance

- Retention limits survive malicious DOMs, images and console floods.
- Redaction and deletion tests pass.
- Repeated lab fixtures remain inside documented tolerances.
- UI and exports cannot visually or structurally confuse raw, rendered, lab and field evidence.
- Accessibility copy never claims complete WCAG conformance.

## 9. Phase 4 — Search Console, GA4 and secrets

Target: 5–8 weeks
Release: 2.3

### 9.1 Secret-store abstraction

- Define opaque credential references with create/use/revoke status.
- Implement Windows Credential Manager, macOS Keychain and Linux Secret Service adapters where available.
- Define a secure unsupported-platform behavior; never fall back silently to plaintext.
- Ensure database backups and profile exports contain references only.
- Add log, diagnostic, crash and MCP redaction tests.

### 9.2 OAuth and integrations

- Create organisation-owned OAuth applications and least-privilege scopes.
- Implement explicit account/property selection.
- Add Search Analytics, sitemap and URL Inspection jobs with quota/rate control.
- Add GA4 landing-page metrics with date/timezone/attribution context.
- Persist source, retrieval time, coverage, quota and API errors.
- Make revoked/expired access fail closed without affecting ordinary crawls.

### 9.3 Orphan model

- Union crawl, sitemap, Search Console and GA4 inventories.
- Define confidence based on which sources observed a URL and crawl completeness.
- Never call a URL an orphan solely because the crawler did not discover it.

### 9.4 Acceptance

- Tokens cannot appear in database snapshots, reports, diagnostics, logs or MCP responses.
- Revocation, expiry and wrong-property tests pass.
- API results always show freshness, property and limitation data.
- Offline use remains fully functional.

## 10. Phase 5 — authenticated and media-rich estates

Target: 8–12 weeks
Release: 2.4

### 10.1 Authenticated crawling

- Support reviewed header/basic/session recipes using secret references.
- Use isolated cookie jars per crawl and purge them according to retention policy.
- Default to GET/HEAD; block POST/PUT/PATCH/DELETE and form submission.
- Revalidate every authenticated redirect and renderer subresource through fetch policy.
- Warn about logout links and mutation-looking URLs; allow explicit exclusions.

### 10.2 PDF audit

- Add content-type and magic-byte confirmation.
- Parse metadata, title, page count, text availability and links under byte/time/page limits.
- Treat encrypted, damaged and image-only PDFs as explicit partial states.
- Never run embedded actions or external references.

### 10.3 Image audit

- Record intrinsic dimensions, encoded bytes, format and responsive variants.
- Analyse broken resources, oversized assets, modern formats, lazy loading and bounded perceptual hashes.
- Keep decorative/informative and alt-quality judgments as review/assisted findings.

### 10.4 Additional SEO relationships

- Add AMP and mobile-alternate reciprocity/status/canonical/indexability rules.
- Add pagination and faceted-navigation graph evidence.
- Add language attribution, readability and opt-in dictionaries/grammar engines.

### 10.5 Acceptance

- Hostile PDF/image fixtures cannot exceed parser budgets or escape processing boundaries.
- Authentication tests prove no unsafe method or private-scope expansion.
- Session credentials and cookies never enter exports/diagnostics.
- Every new rule has positive, negative, malformed and non-applicable fixtures.

## 11. Phase 6 — continuous audits, architecture and agents

Target: 6–10 weeks
Release: 3.0

### 11.1 Architecture analysis

- Implement exact small-graph fixtures for internal PageRank/link score.
- Add depth, hub, isolated-cluster and canonical/hreflang graph summaries.
- Add segment/template assignment rules with explicit provenance.
- Precompute bounded summaries; do not send millions of nodes to the browser.
- Add progressive graph sampling and exported full-edge datasets.

### 11.2 Scheduling and retention

- Define serialisable schedule/job contracts.
- Support documented OS scheduler commands first.
- Add project-level locking and missed/overlapping-run policy.
- Add crawl/artifact retention rules with preview and recoverable deletion where practical.
- Add notification adapters only through separately approved destinations/secrets.

### 11.3 Regression budgets

- Allow users to select deterministic rules, segments and thresholds.
- Refuse comparisons when scope/profile differences make the result invalid unless explicitly acknowledged.
- Return machine-readable pass/fail/indeterminate results to CLI, API and MCP.

### 11.4 Agent remediation workflow

- Add read-only remediation-plan generation from selected issue evidence.
- Require explicit repository/change authority outside the crawler.
- Record proposed change references, build/test result and recrawl IDs.
- Mark a condition fixed only when the comparable recrawl proves it.
- Keep deployment outside the crawler unless a future separately approved product requirement adds it.

### 11.5 Acceptance

- Graph metrics match exact fixtures and remain bounded at large scale.
- Scheduled jobs cannot concurrently own the same crawl/project database unsafely.
- Regression outputs distinguish pass, fail and incomparable.
- Agent workflows cannot convert a proposal into a verified result without recrawl evidence.

## 12. Data and migration plan

Expected new entities:

- `rule_data_bundle`
- `finding_classification`
- `custom_audit_definition` and `custom_audit_result`
- `page_artifact` and `artifact_retention_policy`
- `engine_run` for render/axe/Lighthouse metadata
- `external_connection_reference` and `external_observation`
- `content_segment` and `architecture_metric`
- `schedule_definition`, `scheduled_run` and `regression_budget`
- `remediation_record` and `verification_link`

Migration requirements:

- forward-only numbered SQL migrations;
- migration tests from the oldest supported production schema and current RC4 schema;
- no credential material in migrations;
- explicit defaults distinguishing “not collected” from empty/clean;
- backup before migration and read-only compatibility diagnostics;
- rollback through backup/restore rather than destructive down migrations.

## 13. Contract and surface plan

For every new capability:

1. add internal typed contracts and validation;
2. add application service and permission classification;
3. add database repository/migration;
4. add CLI/API behavior and OpenAPI schema;
5. add MCP tool only when the operation is safe and useful to agents;
6. add web UI after stable query/pagination behavior exists;
7. add streaming export with version/source columns;
8. add documentation and fixture links.

MCP additions should favour small composable tools:

- list active rule/profile/engine versions;
- list conformance coverage;
- manage bounded custom definitions;
- query structured/accessibility/performance observations;
- inspect external-data freshness without exposing tokens;
- create/read remediation and verification records.

Long jobs return opaque IDs and are polled. No MCP tool accepts arbitrary filesystem paths, unrestricted network targets, credentials or executable code.

## 14. Test strategy

### Unit

- parsers, validators, budgets, classification and evidence schemas;
- deterministic graph and comparison logic;
- secret redaction and credential-reference behavior.

### Fixture/integration

- complete site-level fixtures for every rule and source mode;
- database lifecycle and legacy migration tests;
- worker crash, timeout, cancellation and restart tests;
- API/MCP contract agreement.

### Security/adversarial

- SSRF, DNS rebinding, redirect and renderer subresource checks;
- oversized/compressed HTML, DOM, schema, PDF, image and extraction matches;
- malicious selectors/regex and artifact names;
- OAuth revocation, token/log/export leakage and session isolation;
- unsafe HTTP methods and mutation-looking authenticated links.

### Performance

- representative 10k and 100k regression profiles;
- bounded memory/disk measurements for each optional engine;
- database query budgets for primary dashboard views;
- sampled Lighthouse/axe/render concurrency tests;
- graph aggregation and large export tests.

### Platform

- Linux, Windows and macOS CI for core tests;
- native package/install/signature smoke tests;
- clean-machine manual qualification before stable releases;
- rendered-mode qualification only on explicitly declared platforms.

## 15. CI and release gates

Required pull-request checks:

1. formatting/lint/type checks;
2. Go and TypeScript unit/integration suites;
3. network-disabled test profile;
4. database migration and recovery tests;
5. conformance suite and coverage-claim validation;
6. security/static/dependency scans;
7. production UI and binary builds;
8. documentation link and claim checks.

Required tag checks:

1. all pull-request gates at the tagged commit;
2. native platform builds;
3. fixture smoke crawl from packaged artifact;
4. SBOM/checksum/signature/attestation generation;
5. native signing/notarisation verification;
6. clean-install smoke evidence;
7. draft release notes and limitations;
8. human protected-environment approval;
9. post-publish download verification.

## 16. Definition of done

A work item is done only when:

- requirements and non-goals are clear;
- contracts and budgets are reviewed;
- positive, negative, malformed, boundary and non-applicable fixtures pass;
- security/privacy effects are tested;
- legacy data migrates successfully;
- UI, CLI, API, MCP and exports agree where applicable;
- accessibility and keyboard use are checked for UI changes;
- documentation identifies evidence source, version and limitation;
- diagnostics contain no crawled content or secrets beyond approved metadata;
- `git diff --check`, full applicable tests and production builds pass;
- release/claim coverage is updated.

## 17. Staffing and execution model

Recommended minimum team:

- one crawler/rules engineer;
- one browser/performance/accessibility engineer;
- one frontend/reporting engineer;
- one release/security engineer;
- one SEO quality lead;
- one community/documentation maintainer, part-time or shared.

Use two-week iterations. Each iteration should include one vertical capability slice and one quality/release slice. Avoid separate months of backend work followed by delayed UI/MCP integration.

Decision owners:

| Decision | Accountable owner |
|---|---|
| Rule meaning and fixture truth | SEO quality lead |
| Fetch/render/security boundary | Security maintainer |
| Schema and migrations | Storage maintainer |
| Public contracts and compatibility | Application maintainer |
| Signing and release identity | DJAI release owner |
| Claims and release scope | Product owner |

## 18. First four iterations

### Iteration 1

- conformance schemas and fixture server skeleton;
- AUD-01/AUD-04/AUD-05 response/indexability fixtures;
- evidence-source/classification contracts;
- signing identity and supported-platform decision record.

### Iteration 2

- sitemap, robots, hreflang and raw/rendered fixtures;
- conformance JSON/HTML report and coverage catalog;
- draft tag/release workflow with attestations;
- Windows/macOS clean-machine runbooks.

### Iteration 3

- remaining core rule fixtures and repeatability gate;
- first five authorised beta audits;
- packaged smoke-test harness;
- Schema.org bundle format and provenance generator.

### Iteration 4

- correct beta/conformance defects and rerun;
- native signing/notarisation proof;
- first owned-fixture differential benchmark;
- stable release review or explicit pre-release decision;

## 19. Evidence artifacts to retain

- conformance manifest and report version;
- coverage matrix and claim mapping;
- test and production-build summaries;
- dependency scans and SBOM;
- clean-machine qualification records;
- release artifact hashes, signatures and attestations;
- anonymised beta sample/adjudication records;
- comparison version/configuration/result summary;
- migration, recovery and performance benchmark results;
- signed maintainer release decision and dated exceptions.

## 20. Immediate next action

The locally implementable foundation and professional vertical slices are now in the working tree. The next release actions require evidence or authority outside this workspace: authorized closed-beta audits, clean-machine macOS/Windows runs, organisation-owned native signing identities and execution of the protected GitHub candidate workflow.

Capability expansion should proceed one fixture-backed slice at a time. The highest-value remaining slices are additional reviewed Google search-feature profiles, Search Console URL Inspection/sitemap inventory, a separately isolated local Lighthouse worker, a full bounded PDF parser, perceptual image analysis, and exact small-graph/regression fixtures. These remain future work and must not be represented as complete merely because adjacent adapters exist.

## 20.1 Implementation disposition (2026-07-30)

| Area | Disposition | Honest boundary |
|---|---|---|
| Conformance and provenance | Implemented | 17 rule families and 21 represented positive findings; not a commercial-parity proof |
| Schema.org | Implemented for bounded JSON-LD syntax, structure, vocabulary, superseded terms and advisory domain/range | Bundled Schema.org 30.0; not every semantic constraint or remote context |
| Google search features | Implemented for Breadcrumb, Article and merchant Product | Additional Google profiles require separately reviewed versioned data |
| Custom CSS/XPath audits | Implemented | Safe bounded subset; no arbitrary scripts or network access |
| Rendered evidence and axe | Implemented | Automated accessibility coverage is partial by nature |
| PageSpeed and CrUX | Implemented as opt-in Google API adapters | PageSpeed supplies Lighthouse lab evidence; a local Lighthouse worker is not implemented |
| Search Console and GA4 | Implemented for Search Analytics and GA4 landing-page metrics | URL Inspection, sitemap inventory and cross-source orphan confidence remain future work |
| Authentication | Implemented for host-bound raw bearer/basic/cookie crawling | Authenticated rendered crawling is rejected until equivalently mediated |
| DOM/screenshots/artifacts | Implemented with redaction, budgets, expiry and orphan reconciliation | Screenshots are opt-in because pixels can contain personal data |
| PDF/image/mobile/AMP | Implemented as bounded readiness and delivery diagnostics | Full PDF parsing, perceptual image analysis and all reciprocity cases remain partial |
| Architecture and schedules | Implemented as bounded link metrics/table and local recurring jobs | Not a full interactive graph engine or unattended cloud scheduler |
| MCP | Implemented with 36 bounded tools | Secrets are enrolled outside MCP; tools accept references only |
| Release automation | Implemented for candidate archives, SBOM, checksums, provenance attestation and draft prerelease | Native signing/notarization, clean-machine qualification and a formal stable release require external identities/runners |

## 21. Related plans

- [Professional product plan](./PROFESSIONAL_PRODUCT_PLAN.md)
- [Commercial-grade audit quality plan](./COMMERCIAL_QUALITY_PLAN.md)
- [Existing implementation baseline](./IMPLEMENTATION_PLAN.md)
- [System architecture](./ARCHITECTURE.md)
- [Security model](./SECURITY_MODEL.md)
- [Release and reproducibility](./RELEASE.md)
