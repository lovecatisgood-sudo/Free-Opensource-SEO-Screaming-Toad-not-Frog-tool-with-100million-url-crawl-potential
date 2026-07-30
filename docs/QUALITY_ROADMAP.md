# Audit quality roadmap

Status: approved product direction as of 2026-07-30.

## Objective

Increase the completeness, diagnostic usefulness and reproducibility of the SEO audit. Crawl capacity is no longer the primary product-development priority. The theoretical 100M+ architecture remains an explicitly unverified design direction and is not a quality or support claim.

The target is strong coverage for ordinary public-site technical SEO audits while retaining the project's local-first, bounded and evidence-backed security model. Feature count alone is not success: every finding needs applicability conditions, versioned evidence, remediation, limitations, negative fixtures and deterministic output.

## Quality principles

1. Keep observed evidence separate from recommendations and search-engine outcomes.
2. Keep raw, rendered, API, lab and field evidence separately labelled.
3. Validate syntax, Schema.org vocabulary and Google search-feature eligibility as separate layers.
4. Prefer no finding over a false positive when local rule data is incomplete.
5. Never imply that valid markup guarantees indexing, ranking or a rich result.
6. Keep external integrations optional, least-privilege and local by default.
7. Preserve crawl budgets, SSRF controls and renderer mediation when adding depth.

## Delivery order

### Q1 — Structured-data foundations

- AUD-13 JSON-LD syntax and structural validation.
- Bounded context, type and property evidence without storing full scripts by default.
- Versioned Schema.org vocabulary snapshot with provenance and update tests.
- AUD-14 separately versioned Google search-feature profiles, starting with Breadcrumb, Article and Product.
- Raw/rendered findings remain distinct.

Exit gate: malformed and valid fixtures for JSON-LD, microdata and RDFa; no rich-result guarantee language; deterministic offline validation.

### Q2 — Custom audit framework

- Bounded CSS selector, XPath and text/regex search.
- Saved extractor and custom-check definitions.
- Output-size, execution-time and match-count ceilings.
- No page-supplied or user-supplied native code execution.

Exit gate: adversarial selectors and oversized matches cannot bypass resource or evidence limits.

### Q3 — Performance and accessibility

- Isolated, version-pinned axe checks in rendered mode.
- Local Lighthouse lab runs with explicit device and throttling profiles.
- Optional PageSpeed/CrUX field data stored separately from lab results.

Exit gate: reports always identify engine version, profile, timestamp and lab-versus-field source.

### Q4 — Search and analytics enrichment

- Least-privilege Search Console Search Analytics and URL Inspection integration.
- Optional GA4 landing-page enrichment.
- API/list/sitemap/crawl discovery-source union for stronger orphan candidates.

Exit gate: revoked tokens fail closed; credentials never enter crawl exports, diagnostics or MCP responses.

### Q5 — Content and architecture depth

- Configurable near-duplicate groups and bounded semantic similarity.
- Readability plus opt-in spelling and grammar checks with language attribution.
- Internal link score, template/segment views and architecture visualisations.
- Explainable orphan confidence based on observed discovery sources.

### Q6 — Investigation and workflow breadth

- Stored rendered DOM under explicit retention budgets, screenshots and console diagnostics.
- Safe authenticated crawling with a dedicated encrypted secret store and mutation-risk warnings.
- PDF text/metadata, advanced image, pagination, mobile-alternate and AMP checks.
- Saved report views and OS-scheduler-compatible audit jobs.

## Priority score

Work is ordered by audit impact, false-negative reduction, user frequency, implementation risk and ability to operate offline. The first five priorities are structured-data validation, custom audits, Search Console/PageSpeed enrichment, accessibility and rendered-page diagnostics.

## Claim boundary

Approved positioning: “A safe, local-first, evidence-backed open-source technical SEO crawler designed for large segmented campaigns.”

Do not claim parity with or superiority to a named commercial crawler until a published, repeatable audit fixture suite demonstrates comparable coverage and result quality. URL capacity and audit quality are separate measurements.
