# Professional product plan

Status: proposed execution baseline
Date: 2026-07-30
Product: SEO Screaming Toad — Not Frog
Community name: DJAI Toad
Planning horizon: stable 2.0 through 3.0

## 1. Product decision

SEO Screaming Toad will become a professional, local-first technical SEO workbench rather than a crawler that merely produces a large issue count. Its product loop is:

```text
Define scope → Crawl → Inspect evidence → Prioritise → Fix → Recrawl → Verify
```

The open-source engine remains useful without payment, accounts, telemetry or third-party APIs. Professional value comes from dependable results, transparent rule logic, safe automation, strong comparisons and the ability to prove that a technical remediation worked.

The product will not claim complete parity or superiority over Screaming Frog. Comparable claims will be limited to workflows measured in the public conformance suite. The theoretical 100M+ segmented architecture remains separate from audit-quality claims.

## 2. Product vision

Make a technical SEO audit behave like a reproducible software test suite: every finding has a cause, evidence, version, limitation and verifiable resolution.

Long-term product promise:

> Give SEO professionals, developers and AI agents one safe local system for discovering technical search problems, understanding their consequences and verifying fixes without surrendering site data to a hosted platform.

## 3. Target users and paid-value jobs

### 3.1 Independent SEO consultant

Needs to audit varied client websites, defend every recommendation, export affected URLs and show before/after results. Will value trustworthy evidence, reusable profiles and professional reports more than an inflated health score.

### 3.2 Agency technical SEO team

Needs repeatable processes across analysts, comparable crawl configurations, segmented views, scheduling, GSC/GA4 enrichment and auditable client deliverables. Will value reduced analyst time and fewer false-positive review hours.

### 3.3 Web developer and release team

Needs exact failing URLs, evidence and regression gates that can run before and after deployment. Will value CLI/API/MCP access, deterministic fixtures and remediation verification.

### 3.4 In-house site operator

Needs historical monitoring across templates, markets and site sections. Will value private local storage, Search Console context, change detection and architecture views.

### 3.5 AI-assisted non-specialist

Needs an agent to configure a bounded crawl, explain results in plain language and propose safe next steps. Will value MCP workflows and approval-gated remediation while retaining visibility into the underlying evidence.

### 3.6 Educator and open-source learner

Needs inspectable rule logic, fixtures and realistic audit examples. Will value a free, understandable system that can be studied and extended.

## 4. User problems to solve

1. Commercial tools can be expensive or opaque for occasional and educational use.
2. A large finding list often mixes deterministic failures, editorial advice and intentional configurations.
3. JavaScript sites can look different in raw responses and rendered DOMs, but many reports hide that distinction.
4. Search Console, analytics, lab performance and field performance are often combined without source or freshness context.
5. Audit handoffs lose the reasoning between evidence, recommendation and implementation.
6. Manual recrawls make it difficult to prove which issues were fixed or introduced.
7. AI agents can reduce technical barriers, but unrestricted crawling, filesystem access or automated site changes create unacceptable risk.

## 5. Product principles

1. **Evidence before score:** show observed facts before conclusions.
2. **Applicability before severity:** intentional or non-applicable conditions must not be presented as universal failures.
3. **Raw and rendered are different sources:** never overwrite one with the other.
4. **Lab and field are different measurements:** label engine, profile, date, freshness and limitations.
5. **Local by default:** external data and retained page content are explicit options.
6. **One domain engine:** UI, CLI, API and MCP return the same findings and versions.
7. **Fixes require verification:** proposed changes are not completed work until a recrawl confirms them.
8. **Bounded extensibility:** custom audits cannot become arbitrary local or page-supplied code execution.
9. **Professional honesty:** a clean technical crawl does not guarantee indexing, rankings, traffic, conversions, rich results or AI citations.

## 6. Product pillars

### Pillar A — trustworthy crawl and indexability

- guarded URL discovery, redirects, robots and sitemap processing;
- durable recovery, explicit terminal reasons and coverage accounting;
- response, canonical, indexability, hreflang and architecture evidence;
- stable list and spider modes with comparable configurations.

### Pillar B — semantic and content diagnostics

- metadata, headings, exact and near-duplicate analysis;
- complete Schema.org vocabulary validation;
- separately versioned Google rich-result profiles;
- language-aware content and template diagnostics;
- bounded custom extraction and custom audits.

### Pillar C — rendered experience and quality signals

- raw/rendered differences, retained DOM and screenshots under explicit budgets;
- console and resource diagnostics;
- axe accessibility evidence;
- Lighthouse lab profiles and separate PageSpeed/CrUX field data.

### Pillar D — search and business context

- Search Console Search Analytics, sitemap and URL Inspection context;
- optional GA4 landing-page enrichment;
- discovery-source union for explainable orphan candidates;
- no inference that missing traffic or API data proves a technical fault.

### Pillar E — professional workflows

- filters, saved views, comparisons, reports and regression budgets;
- PDF, image, AMP, mobile-alternate and authenticated-site coverage;
- architecture visualisation and template/segment analysis;
- scheduled jobs with safe locking and retention;
- agent-assisted explanation, remediation plan and verified recrawl.

## 7. Core product journeys

### 7.1 First professional audit

1. User creates a project or selects a safe preset.
2. Product previews allowed hosts, path rules, robots behavior, URL/disk/time ceilings and rendering mode.
3. User confirms the scope and starts the crawl.
4. Dashboard distinguishes discovered, fetched, rendered, analysed, skipped and failed URLs.
5. Summary separates deterministic errors, recommendations, intentional-review items and coverage limitations.
6. User opens a finding to see applicability, evidence, source mode, rule version, impact, remediation and caveat.
7. User filters affected templates/segments and exports a professional workbook or machine-readable report.

### 7.2 Remediation and verification

1. User selects issues to fix and creates a remediation plan.
2. If source code is available, an AI coding agent may propose changes within the authorised repository.
3. Product or agent builds and tests the site.
4. User runs a targeted recrawl, then a full applicable recrawl.
5. Comparison marks issues fixed, new, persistent or not comparable because coverage changed.
6. Final report links the original evidence to the verified result.

### 7.3 Connected performance investigation

1. User explicitly connects Search Console, PageSpeed/CrUX or GA4.
2. Product displays account, property, date range, quota and freshness before import.
3. Crawl facts remain visually separate from API, lab and field evidence.
4. User prioritises technical work with observed search or business context without treating correlation as causation.

### 7.4 Agent-operated audit through MCP

1. Agent checks product health and lists or creates a project/profile.
2. Agent previews scope and presents material limits to the user.
3. Agent starts a bounded crawl and polls status without blocking.
4. Agent retrieves summary, representative evidence and rule explanations.
5. Agent produces a prioritised report that labels observations, interpretations and limitations.
6. Any code modification or deployment requires separate authority.
7. Agent recrawls and reports only measured before/after results.

## 8. Release plan

### 2.0 stable — qualified local audit core

Purpose: publish the current core through professional release gates.

Must deliver:

- public core conformance suite and coverage matrix;
- 20 authorised beta audits across at least 12 varied sites;
- packaged clean-machine qualification on declared Windows, macOS and Linux targets;
- organisation-owned signing identity, native platform trust where applicable, SBOM, checksums and attestations;
- formal GitHub release with post-download verification;
- explicit current limitations in the UI, README and release notes.

Not included: complete Schema.org/Google profiles, integrations and advanced media/auth workflows.

### 2.1 — semantic correctness and expert rules

Purpose: provide strong structured-data auditing and safe site-specific extensibility.

Must deliver:

- versioned Schema.org vocabulary registry with provenance;
- Google Breadcrumb, Article and Product profiles, followed by remaining selected search features;
- bounded CSS/XPath/text/regex extraction;
- saved custom audit definitions;
- UI, export, API and MCP access to custom results;
- conformance results for every supported vocabulary/profile and extractor boundary.

### 2.2 — rendered quality, accessibility and performance

Purpose: diagnose modern JavaScript sites and user-experience risks.

Must deliver:

- opt-in rendered DOM and screenshot retention;
- raw-versus-rendered field diffs;
- console and failed-resource evidence;
- pinned axe checks with automated-coverage caveats;
- pinned Lighthouse lab profiles;
- optional PageSpeed/CrUX field data kept separate from lab data.

### 2.3 — Search Console and analytics context

Purpose: join technical audit evidence to real search and landing-page observations.

Must deliver:

- OS-backed encrypted secret references;
- Search Console Search Analytics, sitemap and URL Inspection integration;
- GA4 landing-page enrichment;
- explicit property/date/freshness/quota metadata;
- explainable orphan confidence from crawl, sitemap and API sources;
- token revocation and leak-prevention tests.

### 2.4 — advanced crawl and media coverage

Purpose: cover more professional site estates.

Must deliver:

- authenticated crawling with safe methods and isolated sessions;
- bounded PDF and advanced image auditing;
- AMP/mobile-alternate and pagination checks;
- language-aware readability and opt-in spelling/grammar;
- faceted-navigation evidence and recommendations.

### 3.0 — continuous audit and verified remediation

Purpose: turn the crawler into a recurring technical-search operations system.

Must deliver:

- saved views and report recipes;
- OS-scheduler-compatible and later in-app scheduled jobs;
- crawl regression budgets;
- internal link score and architecture visualisations;
- template and segment dashboards;
- approval-gated agent remediation plans;
- build, targeted recrawl, full recrawl and before/after evidence workflow.

## 9. Functional requirement groups

| ID | Requirement | Target |
|---|---|---|
| PQ-01 | Publish machine-readable conformance results and supported/partial/unsupported coverage. | 2.0 |
| PQ-02 | Separate deterministic findings, editorial diagnostics, intentional-review cases and unknowns. | 2.0 |
| PQ-03 | Qualify signed downloadable artifacts on declared clean operating systems. | 2.0 |
| PQ-04 | Validate Schema.org vocabulary independently from search-feature eligibility. | 2.1 |
| PQ-05 | Support safe bounded custom extraction and audit definitions. | 2.1 |
| PQ-06 | Retain rendered evidence only under explicit privacy and disk budgets. | 2.2 |
| PQ-07 | Provide pinned accessibility and performance engines with provenance. | 2.2 |
| PQ-08 | Keep API, lab and field evidence independently identifiable. | 2.2 |
| PQ-09 | Store external credentials outside crawl databases, backups, reports and MCP output. | 2.3 |
| PQ-10 | Explain orphan confidence from observed discovery sources. | 2.3 |
| PQ-11 | Crawl authenticated targets without unsafe form mutation or scope expansion. | 2.4 |
| PQ-12 | Audit PDF, image, AMP, mobile and pagination cases within resource ceilings. | 2.4 |
| PQ-13 | Support scheduled regression checks without unsafe project overlap. | 3.0 |
| PQ-14 | Make architecture metrics explainable and coverage-aware. | 3.0 |
| PQ-15 | Require recrawl evidence before an agent marks a technical condition fixed. | 3.0 |

## 10. Information and UX model

Every finding view must show:

- classification: deterministic error, warning/recommendation, review item or information;
- affected URL/subject and observed value;
- raw, rendered, crawl-graph, API, lab or field source;
- rule/profile/engine version and timestamp where relevant;
- applicability and confidence;
- likely technical consequence;
- remediation with unsafe blanket-action warnings;
- limitation and missing-evidence statement;
- change status compared with a selected baseline.

Primary workspace areas:

1. Projects and profiles
2. Crawl configuration and scope preview
3. Live crawl and resource budgets
4. Overview and prioritised findings
5. Page and link evidence
6. Structured data
7. Rendering and accessibility
8. Performance and field data
9. Search Console and analytics
10. Architecture and segments
11. Comparisons and verification
12. Reports, artifacts and scheduling

Advanced areas should appear only when their source or engine is enabled. The default dashboard must remain useful for a raw, offline audit.

## 11. Product quality metrics

### Correctness

- 100% precision and recall for represented deterministic Must cases before stable 2.0;
- at least 99.5% precision across the deterministic fixture corpus;
- below 2% confirmed live-beta false positives in sampled deterministic findings;
- 100% provenance/source labels for applicable findings.

### Reliability

- no semantic output drift across ten identical fixture runs per supported OS;
- no silent running state after cancellation, crash or exhausted budget;
- exports, UI, API and MCP agree on counts and evidence versions;
- no unresolved high/critical security or data-integrity issue at release.

### Professional usefulness

- at least 80% of beta reviewers can identify evidence and recommended next action without external documentation;
- median time from completed crawl to a filtered, exportable priority list below ten minutes on the reference fixture;
- at least 90% of confirmed repaired fixture defects reported as fixed on recrawl;
- fewer than 5% of beta findings marked “unclear explanation” by reviewers.

### Adoption

- successful signed-download verification rate measured through opt-in release feedback, not telemetry;
- GitHub issue templates produce reproducible reports with profile and version data;
- community contributions add or update fixtures alongside rule changes;
- documentation covers install, first crawl, MCP, limitations and verification on every supported OS.

## 12. Open-source and sustainability model

The crawler, rule engine, local UI, CLI, MCP server and conformance fixtures remain open source. The project should not create an artificially weakened free crawler merely to sell basic correctness.

Possible revenue and sustainability paths:

- paid DJAI technical SEO audits and remediation;
- paid deployment, integration and custom-rule development;
- team training and DJAI Academy courses;
- sponsored feature development;
- optional commercial support with response-time commitments;
- future hosted collaboration or managed scheduling, only under a separate privacy/security PRD.

Paid services must be described separately from measured software capabilities. Community contributions remain credited under the project contribution policy.

## 13. Go-to-market sequence

1. **Evidence preview:** publish the conformance approach and invite authorised beta sites.
2. **Stable core:** publish signed 2.0 binaries, fixture coverage and honest limitations.
3. **Expert proof:** release technical case studies with before/after evidence and configuration files.
4. **Extensibility:** demonstrate custom audits and structured-data profiles in 2.1.
5. **Modern-site depth:** demonstrate raw/rendered, accessibility and performance evidence in 2.2.
6. **Operational workflow:** demonstrate connected data and verified agent remediation only when those releases pass their gates.

Marketing may say “Screaming Frog alternative” as a discovery phrase, but product copy must immediately explain the measured scope and known differences. It must not imply affiliation with Screaming Frog.

## 14. Product risks and mitigations

| Risk | Product response |
|---|---|
| Feature breadth outruns correctness | No feature ships without fixtures, provenance and an exit gate. |
| False positives reduce expert trust | Separate deterministic failures from recommendations and track sampled precision. |
| Integrations leak client data | Optional connections, OS-backed secrets, explicit fields and redaction tests. |
| Rendering increases attack surface | Isolated workers, pinned runtime, denied capabilities and strict budgets. |
| AI overstates recommendations | Agent responses cite evidence and require verified recrawl before completion. |
| Comparison claims create credibility/legal risk | Publish narrow versioned workflow metrics; do not copy proprietary rule text. |
| Open-source maintenance becomes unsustainable | Fund services, support, training and sponsored work without weakening core correctness. |
| Large-scale language distracts from quality | Keep theoretical capacity separate from supported audit and release metrics. |

## 15. Product release gates

A release is ready only when:

1. its product requirements map to tests and evidence;
2. supported, partial and unsupported behavior is documented;
3. fixture and migration tests pass;
4. UI, CLI, API, MCP and exports agree;
5. new retained data has privacy, deletion and budget controls;
6. new network/API paths preserve scope and secret boundaries;
7. clean-machine packages pass for declared platforms;
8. release artifacts verify through checksums, signatures and attestations;
9. known limitations and upgrade behavior are published;
10. public claims do not exceed measured evidence.

## 16. Decisions needed from maintainers

Before stable 2.0:

- name the DJAI organisation that owns release identities;
- select and fund Apple Developer ID and Windows signing arrangements;
- confirm the supported OS/version matrix;
- recruit authorised beta participants and two SEO reviewers;
- assign maintainers for rules, security, release and beta adjudication.

Before 2.3:

- approve the OS secret-store design and OAuth application ownership;
- approve external API cost/quota policy and offline behavior.

Before 3.0:

- decide whether scheduling remains OS-driven or becomes an in-app service;
- define the exact approval boundary for agent-proposed source modifications.

## 17. Related plans

- [Commercial-grade audit quality plan](./COMMERCIAL_QUALITY_PLAN.md)
- [Professional implementation plan](./PROFESSIONAL_IMPLEMENTATION_PLAN.md)
- [Product requirements baseline](./PRD.md)
- [System architecture](./ARCHITECTURE.md)
- [Audit quality roadmap](./QUALITY_ROADMAP.md)
- [Security model](./SECURITY_MODEL.md)
