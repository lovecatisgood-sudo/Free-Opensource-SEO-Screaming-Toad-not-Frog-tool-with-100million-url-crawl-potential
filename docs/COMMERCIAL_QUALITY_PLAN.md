# Commercial-grade audit quality plan

Status: proposed execution baseline
Date: 2026-07-30
Product: SEO Screaming Toad — Not Frog / DJAI Toad

## 1. Decision

The product should not try to win by copying every tab in a mature commercial crawler. It should first become independently trustworthy on the SEO decisions experts make most often, then add breadth in measured waves.

The target is:

> A safe, local-first, evidence-backed open-source technical SEO crawler whose results are reproducible enough for professional audits and whose agent workflow can take an authorised site from discovery through explanation, remediation and verified recrawl.

“As good as” must be demonstrated per workflow. “Better” may be claimed only for a named, measured dimension such as audit reproducibility, rule transparency, privacy, evidence traceability or agent-assisted verification. Do not make an overall parity or superiority claim until the public conformance suite supports it.

Screaming Frog SEO Spider 24.3 is the current comparison baseline as of this plan. Its official release history records MCP support in 24.0, so MCP availability alone is no longer a differentiator. Its current documented feature surface also includes structured-data and Google rich-result validation, custom extraction/search/JavaScript, Lighthouse and PageSpeed data, axe accessibility checks, Search Console and GA4 integrations, scheduling, semantic similarity and architecture visualisations. The relevant product opportunity is a more transparent and safer end-to-end audit workflow, not a feature-count slogan.

## 2. What professional quality means

Each rule and integration must be judged on six dimensions:

| Dimension | Required evidence |
|---|---|
| Correctness | Versioned positive, negative, boundary and malformed fixtures with reviewed expected results. |
| Completeness | A published coverage matrix showing supported, partial and unsupported cases. |
| Explainability | Applicability, observed evidence, confidence, impact, remediation, limitations and source version. |
| Reproducibility | Pinned engines/profiles and deterministic output from the same fixture and configuration. |
| Safety | Bounded resource use, guarded networking, secret isolation and failure that does not widen scope. |
| Workflow value | Filtering, prioritisation, export, comparison, MCP access and fix/recrawl verification. |

The product is ready for professional use when an expert can answer all of these questions from the report:

1. What exactly was tested, and was it raw HTML, rendered DOM, API, lab or field evidence?
2. Why did the rule apply, and what evidence triggered it?
3. Is this an error, a recommendation, a possible intentional configuration or an unknown?
4. What is the likely user/search impact without pretending to predict rankings?
5. What should be changed, and what could go wrong if it is changed blindly?
6. Did a recrawl prove that the technical condition was fixed?

## 3. Current baseline

The release candidate already has a strong foundation:

- guarded fetch and render boundaries, crawl budgets and durable SQLite recovery;
- raw and rendered findings stored separately;
- 13 versioned audit-rule families with structured evidence;
- exact and bounded near-duplicate analysis;
- web UI, CLI, HTTP API and 27 MCP tools over common application services;
- comparisons, exports, campaign segmentation, SBOMs and checksums;
- cross-compilation for Linux, macOS and Windows;
- a real authorised DJAI Academy audit and synthetic 1M/5M campaign evidence.

The remaining weakness is not the crawl loop. It is the breadth of audit semantics, the quantity of ground-truth validation and public release trust.

Fresh verification on 2026-07-30 found that the repository-local Go 1.26.5 toolchain is usable. The full Go suite passed when its cache was placed inside the workspace and loopback fixture tests were permitted to bind sockets. Both TypeScript workspaces passed all 18 tests and type checking. The earlier statement that Go was unavailable should therefore be treated as stale.

## 4. Two parallel programmes

Work must proceed through two parallel programmes. New feature work must not delay the stable core indefinitely, and the stable release must not be represented as complete commercial parity.

### Programme A — qualify and publish the stable core

This resolves the current release blockers. It is a 2.0 stable gate, not optional roadmap work.

### Programme B — deepen professional audit quality

This delivers the remaining capabilities in quality-ranked waves. Each wave ships only when its fixture, documentation and evidence gates pass.

## 5. Programme A: stable release qualification

### A1. Public conformance fixture laboratory

Create `internal/testfixtures/sites/` containing deterministic HTTP fixture sites rather than isolated parser strings. Every fixture should expose both a raw response and, where applicable, a rendered state.

Initial fixture families:

1. response codes, redirect chains/loops, retry behaviour and malformed headers;
2. robots policies, meta/X-Robots conflicts and canonical chains;
3. sitemap indexes, alternate/image/video entries, malformed XML and coverage conflicts;
4. canonicals, pagination, mobile alternates, AMP and query duplicates;
5. hreflang clusters with reciprocal, regional, canonical and indexability failures;
6. raw-versus-rendered titles, canonicals, links, structured data and content;
7. duplicate, near-duplicate, thin/template and language-aware content cases;
8. structured data in JSON-LD, microdata and RDFa;
9. images, PDFs, accessibility and performance cases as those engines land;
10. crawl traps, authentication, session boundaries and adversarial resource limits.

Each fixture case needs:

- a stable case ID;
- the intended SEO interpretation;
- expected findings and explicit expected absences;
- rule/profile/engine versions;
- raw and rendered evidence snapshots;
- a licence suitable for redistribution;
- a regression owner and provenance note.

Build a conformance runner that emits machine-readable JSON, a human HTML report and a coverage matrix. CI must fail on an unexpected finding, a missing expected finding, unstable ordering or an unapproved snapshot change.

### A2. Accuracy gates

Use the fixture corpus as ground truth. Do not use agreement with another crawler as truth by itself.

For stable 2.0:

- 100% recall and 100% precision for deterministic Must-level core cases represented in the fixture suite;
- at least 99.5% precision across all deterministic rule cases;
- no nondeterministic output across 10 repeated runs on each supported OS;
- every supported rule has at least one positive, negative, malformed and non-applicable case;
- every finding schema validates and every rule explanation includes a limitation;
- raw/rendered provenance is present on 100% of applicable findings;
- cancellation, disk ceiling, restart recovery and export consistency pass on all supported OSs.

For live beta results, where perfect ground truth is not available:

- sample at least 30 findings per applicable core rule across the beta portfolio, or all findings if fewer;
- sample at least 20 clean pages per applicable rule to look for false negatives;
- require no unresolved high/critical product or security defect;
- require less than 2% confirmed false positives across sampled deterministic findings;
- document every material false negative and its release disposition.

Editorial diagnostics such as title length and inferred intent must be reported separately from deterministic errors and must not be included in the deterministic precision score.

### A3. Closed-beta audit matrix

Use only sites owned by participants or covered by explicit written authorisation. Start with 20 audits across at least 12 distinct sites:

| Site type | Minimum audits | Required stress case |
|---|---:|---|
| Small static/business | 2 | clean baseline plus common metadata defect |
| WordPress/news/editorial | 3 | categories, pagination, media and large link graph |
| Shopify/WooCommerce/e-commerce | 3 | products, variants, faceting and Product schema |
| Multilingual | 3 | reciprocal hreflang and locale routing |
| React/Next/Vue JavaScript | 3 | raw/rendered disagreement and blocked resources |
| Migration/redirect estate | 2 | chains, loops, canonicals and retired URLs |
| Large sitemap/site | 2 | at least 100k authorised URLs or the full authorised estate |
| Throttled/WAF-protected | 1 | Retry-After, low rate and stable cancellation |
| Authenticated staging | 1 | only after the authenticated-crawl security gate exists |

For each run retain the immutable profile, owner authorisation, scope, terminal reason, host-load observation, anonymised finding review, false-positive/negative sample, exports and recrawl result. At least two independent SEO reviewers should adjudicate disputed high-impact findings.

### A4. Differential professional benchmark

Create a versioned comparison matrix against Screaming Frog 24.3 on the same owned fixture sites and configuration. Compare outcomes, not UI labels:

- discovered URL set and exclusion reasons;
- final status/redirect graph;
- indexability and canonical classification;
- sitemap and robots coverage;
- hreflang clusters;
- structured-data results;
- raw/rendered differences;
- duplicate/near-duplicate groups;
- exports and crawl comparisons.

Classify every difference as Toad defect, comparator defect, configuration mismatch, interpretation difference or expected scope difference. A reviewer decides from standards, primary documentation and fixture intent. Publish aggregate results and fixture definitions; do not redistribute proprietary binaries, private site data or copied proprietary rule text.

### A5. Clean-machine runtime qualification

Test packaged artifacts, not a source checkout.

Required matrix:

- Windows 11 x64, standard non-admin user, clean VM;
- macOS current and previous major versions on Apple Silicon;
- macOS x64 where hardware or a trusted runner remains available;
- Ubuntu LTS x64 and arm64;
- raw-only install/start/crawl/export/MCP/uninstall path on every platform;
- rendered-mode installation, browser provisioning, crash recovery and permissions on declared rendered-mode platforms.

Each run verifies checksum/signature instructions, first launch, loopback binding, spaces/non-ASCII paths, locked files, sleep/resume, low disk, firewall prompt, Unicode exports and removal of application data. Publish a dated qualification record with OS build, architecture, artifact hash and known exceptions.

### A6. Signing and release identity

Adopt three complementary trust mechanisms:

1. GitHub Actions OIDC artifact attestations for the release archives, `SHA256SUMS` and SBOM.
2. Keyless Sigstore/cosign signatures for the checksum manifest, with verification instructions.
3. Native platform trust where users expect it:
   - Apple Developer ID signing, hardened runtime, notarisation and stapling for macOS packages;
   - Authenticode signing for Windows through an organisation-controlled certificate or managed trusted-signing service.

Keys and account credentials must be owned by a DJAI release organisation, not a contributor’s personal workstation. Use protected GitHub environments, tag protection, least-privilege workflow permissions and two maintainers for release approval when maintainership permits. A compromised identity must have a documented revocation and replacement procedure.

### A7. Formal GitHub release

Create a tag-driven release workflow that:

1. accepts only a protected semantic-version tag pointing to a main-branch commit;
2. rebuilds and tests the UI, Go services and MCP server;
3. builds packages on native OS runners where platform signing requires it;
4. runs clean-install smoke tests;
5. generates SBOM, checksums, signatures and provenance attestations;
6. verifies every artifact before upload;
7. creates a draft GitHub release with release notes and known limitations;
8. requires human approval before publication;
9. performs a post-publish download and verification test.

Stable 2.0 publication is allowed only after A1–A7 are evidenced. If native Apple or Windows identity procurement is delayed, publish a clearly labelled beta/pre-release; do not relabel unsigned binaries as stable signed downloads.

## 6. Programme B: capability and quality roadmap

### Wave 1 — semantic correctness and expert extensibility

Target: make core technical findings trustworthy and adaptable.

#### 1. Complete structured-data validation

- Vendor a provenance-recorded, versioned Schema.org vocabulary snapshot.
- Validate known types/properties, domains/ranges, superseded vocabulary and case policy for JSON-LD, microdata and RDFa.
- Keep syntax, Schema.org vocabulary and search-engine profile findings separate.
- Add versioned Google rich-result profiles, beginning with Breadcrumb, Article and Product, then the remaining currently supported search-gallery types.
- Treat required properties as errors and recommended properties as warnings only when the relevant profile applies.
- Compare raw markup with rendered markup and visible page content where a requirement depends on visible reality.
- Add an update bot that opens a reviewable profile-change pull request; never silently fetch rule definitions during a crawl.

#### 2. Custom extraction and custom audits

- Bounded CSS, XPath, text and regular-expression extractors.
- Typed outputs: string, number, boolean, URL and list.
- Conditions over status, content type, raw/rendered mode and extracted values.
- Saved, importable JSON definitions with schema version and human-readable provenance.
- Match-count, value-size, regex-time, selector-complexity and total-work ceilings.
- No arbitrary native code. Custom JavaScript, if ever added, must run in a separate opt-in sandbox with no secrets, filesystem or unrestricted network access.
- Visual selector preview can follow after the safe engine and fixtures are complete.

Exit gate: the complete structured-data and custom-audit fixture suites pass; adversarial patterns cannot escape work/evidence limits; custom results are available through UI, exports, API and MCP.

### Wave 2 — rendered evidence, accessibility and performance

Target: explain what a user and a modern search renderer actually receive.

#### 3. Rendered DOM and screenshot evidence

- Explicit opt-in retention policy with per-page, per-crawl and disk ceilings.
- Compressed DOM snapshots, viewport screenshots, console errors and failed/blocked resource summaries.
- Redaction hooks for selectors, query values, headers and form fields.
- Hash and timestamp every retained artifact and link it to its page, renderer version and profile.
- Add raw-versus-rendered diffs for metadata, headings, links, structured data, robots and visible text.

#### 4. axe accessibility

- Run a version-pinned axe engine only in isolated rendered mode.
- Store rule ID, impact, WCAG tags, selector/location and bounded evidence.
- Clearly state that automated checks do not constitute complete WCAG conformance.
- Add keyboard/focus/manual-review prompts for conditions automation cannot decide.

#### 5. Lighthouse and PageSpeed/CrUX

- Run pinned local Lighthouse/Insight audits with named mobile/desktop and throttling profiles.
- Store lab engine version, environment, profile and run-to-run variance.
- Optionally query PageSpeed Insights and CrUX field data; retain API and field evidence separately from Lighthouse lab results.
- Group by templates and run representative samples rather than forcing a costly lab run on every URL.
- Never convert a Lighthouse SEO score into an overall SEO-health score.

Exit gate: raw/rendered/lab/field labels are impossible to confuse in storage, UI or export; repeated lab fixtures stay inside a documented tolerance; secrets and page content are absent from diagnostics.

### Wave 3 — search performance and analytics context

Target: connect technical conditions to observed search and business data without confusing correlation with causation.

#### 6. Search Console

- Least-privilege OAuth for Search Analytics, sitemap data and URL Inspection.
- Separate discovered-by-crawl, sitemap and API inventories to create explainable orphan candidates.
- Record API freshness, quota, permission and sampling limitations.
- Encrypt refresh tokens in a dedicated OS-backed secret store; never include them in DB backups, exports, logs or MCP output.

#### 7. GA4

- Optional landing-page sessions, engagement and conversion enrichment.
- Explicit property, stream, date range, timezone and attribution context.
- No claim that analytics absence proves an SEO defect.

Exit gate: revocation fails closed, account selection is explicit, token-leak tests pass, and every metric carries source/freshness context.

### Wave 4 — crawl breadth and media depth

Target: cover the remaining high-frequency professional workflows.

- Authenticated crawling through OS-backed secrets, explicit login recipes, session isolation and mutation-risk warnings. Default to GET/HEAD and block unsafe form actions.
- PDF status, metadata, title, page count, text availability, link and indexability checks with decompression and parser limits.
- Advanced image checks for dimensions, bytes, format, broken resources, responsive variants, lazy loading and bounded duplicate hashing. Treat visual alt quality as human/AI-assisted review, not deterministic fact.
- AMP and mobile-alternate reciprocity, canonical, status and indexability rules.
- Pagination and faceted-navigation analysis based on graph and canonical evidence.
- Readability, language detection and opt-in spelling/grammar with language and model/dictionary provenance.

Exit gate: hostile binary/media fixtures remain bounded; authenticated crawls cannot widen scope or expose secrets; each new rule family has complete positive/negative fixtures.

### Wave 5 — architecture, scheduling and remediation workflow

Target: make large audits easier to interpret and act on than a flat issue list.

- Internal PageRank/link score with documented formula and crawl-coverage caveats.
- Template/segment views, crawl-depth distribution, hubs, isolated clusters and canonical/hreflang graph views.
- Saved views and report recipes as versioned portable configuration.
- OS-scheduler-compatible jobs first; an in-app scheduler only after secret, locking and retention behavior is proven.
- Crawl-to-crawl regression budgets: fail a scheduled audit when selected Must findings regress beyond a threshold.
- Agent remediation plans that cite files/evidence where source access exists, require user approval for changes, and always end in build, recrawl and before/after verification.

Exit gate: graph calculations have small exact fixtures and large bounded tests; scheduled jobs cannot overlap the same project unsafely; agent output never represents an unverified proposed fix as complete.

## 7. Architecture changes

Preserve the current security boundaries and add these components behind application services:

```text
Versioned rule-data registry
  ├─ Schema.org vocabulary snapshots
  ├─ Google search-feature profiles
  ├─ axe/Lighthouse engine manifests
  └─ provenance and update policy

Evidence engines
  ├─ deterministic raw extractor/rules
  ├─ isolated rendered DOM/axe worker
  ├─ isolated Lighthouse worker
  ├─ bounded PDF/image analyzers
  └─ optional external API adapters

Conformance laboratory
  ├─ fixture-site server
  ├─ expected finding manifests
  ├─ cross-platform runner
  ├─ differential comparator
  └─ coverage/precision report

Credential boundary
  ├─ OS-backed encrypted secret store
  ├─ least-privilege OAuth adapters
  └─ redaction and revocation tests
```

Recommended new packages:

- `internal/ruledata` for immutable, versioned vocabularies/profiles;
- `internal/customaudit` for bounded extraction and conditions;
- `internal/accessibility` and `internal/performance` for engine-neutral contracts;
- `internal/integrations/searchconsole`, `pagespeed` and `ga4`;
- `internal/secretstore` for platform-backed credential references;
- `internal/mediaaudit` for PDF/image analysis;
- `internal/architecture` for link scoring and graph summaries;
- `internal/conformance` plus `internal/testfixtures/sites` for ground truth.

The Go application remains the only scheduler, scope authority and database writer. Node/browser tools remain isolated workers. External APIs enrich stored crawl facts; they do not alter crawl scope or silently change deterministic rule results.

## 8. Delivery order and estimates

Estimates are planning ranges after design, not promises.

| Phase | Scope | Team calendar with 4–6 contributors | Primary release |
|---|---|---:|---|
| 0 | Conformance harness, release workflow, signing setup, beta recruitment | 3–5 weeks | 2.0 stable candidate |
| 1 | 20 beta audits, clean-machine matrix, signed/notarised packages, formal release | 3–6 weeks, overlaps Phase 0 | 2.0 stable |
| 2 | Full Schema.org/Google profiles and custom audits | 6–9 weeks | 2.1 |
| 3 | DOM retention, screenshots, axe, Lighthouse, PSI/CrUX | 7–10 weeks | 2.2 |
| 4 | Search Console, GA4 and secret store | 5–8 weeks | 2.3 |
| 5 | Auth, PDF/image, AMP/mobile, content depth | 8–12 weeks | 2.4 |
| 6 | Scheduling, visualisation and verified agent remediation | 6–10 weeks | 3.0 |

With a focused 4–6 person team, the broad roadmap is roughly 8–12 calendar months. A single maintainer should expect roughly 18–30 months because platform qualification, integration maintenance, fixture curation and beta adjudication cannot be compressed like ordinary feature coding.

Suggested ownership:

- crawler/rules engineer;
- renderer/performance/accessibility engineer;
- frontend/reporting engineer;
- release/security engineer;
- SEO test lead who owns fixture truth and beta adjudication;
- community maintainer for documentation, triage and authorised beta recruitment.

## 9. Release and claim ladder

Use increasingly strong language only as evidence accumulates:

1. **Now:** “Safe, local-first, evidence-backed open-source technical SEO crawler.”
2. **After stable gates:** “Professionally qualified on the published core conformance suite and cross-platform release matrix.”
3. **After workflow parity gates:** “Covers the published core technical-audit workflows represented in our comparison suite.”
4. **Only after independent repeatable evidence:** a narrowly scoped comparison claim naming the product versions, workflow, fixture release and metric.

Never claim that a clean crawl guarantees indexing, ranking, traffic, rich results, AI citations or revenue. Continue using the approved theoretical 100M+ language; capacity does not establish audit quality.

## 10. Expert acceptance scorecard

A milestone is commercially credible when all applicable rows are green:

| Area | Gate |
|---|---|
| Core correctness | Deterministic fixture thresholds in A2 pass. |
| Breadth | Coverage matrix has no unlabelled partial/unsupported workflow. |
| Live usefulness | 20 authorised beta audits reviewed and material errors dispositioned. |
| Cross-platform | Clean-machine package tests pass on the declared support matrix. |
| Supply chain | Checksums, SBOM, signatures, attestations and native trust checks verify. |
| Security/privacy | No unresolved high/critical issue; secrets and retained content obey explicit budgets. |
| Reports | UI, CSV/XLSX, API and MCP agree on counts, evidence and versions. |
| Remediation | Before/after recrawl proves the technical condition changed. |
| Documentation | Limitations, profiles, engines and evidence source are visible to the user. |
| Claims | Public copy is no stronger than the published evidence. |

## 11. First 30 days

Week 1:

- create the conformance manifest format and runner;
- port existing rule tests into the first complete fixture site;
- add a feature/coverage matrix with current partial and unsupported labels;
- decide the DJAI organisation release identity and platform signing budget.

Week 2:

- add robots, redirects, canonical, sitemap, hreflang and raw/rendered fixtures;
- add tag-driven draft-release workflow, SBOM/checksum attestation and post-download verification;
- prepare clean Windows/macOS VM runbooks;
- recruit authorised beta owners using the fixed data-retention and anonymisation form.

Week 3:

- run the first five varied beta audits and adjudicate samples;
- run the first cross-platform packaged-artifact tests;
- implement Schema.org snapshot provenance/update tooling;
- publish the initial conformance report as a CI artifact.

Week 4:

- fix all confirmed core false positives/negatives;
- rerun and publish the first differential comparison on owned fixtures;
- complete the remaining stable-release evidence or keep the build labelled pre-release;
- begin Breadcrumb, Article and Product profiles only after the core gates remain green.

## 12. External standards and comparison references

- Screaming Frog [release history](https://www.screamingfrog.co.uk/seo-spider/release-history/), [user guide](https://www.screamingfrog.co.uk/seo-spider/user-guide/) and [configuration reference](https://www.screamingfrog.co.uk/seo-spider/user-guide/configuration/) define the comparison product surface used in this plan.
- Google documents that generic Schema.org validation and Google rich-result eligibility are separate validation layers, and that valid markup does not guarantee a rich result: [structured-data validation](https://developers.google.com/search/docs/appearance/structured-data) and [general guidelines](https://developers.google.com/search/docs/appearance/structured-data/sd-policies).
- GitHub documents OIDC-backed [artifact attestations](https://docs.github.com/en/actions/concepts/security/artifact-attestations) and their provenance—not security-guarantee—boundary.
- Apple requires the appropriate Developer ID identity for direct-distribution signing and describes [notarisation](https://developer.apple.com/documentation/security/notarizing-macos-software-before-distribution) as a separate malware and signing check.

These references are inputs to versioned local rule/profile data. Production crawls must remain deterministic and must not depend on live documentation pages.
