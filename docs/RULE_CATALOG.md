# Audit rule catalog

Rules are versioned and every finding stores its rule ID, version, subject, severity and JSON evidence. Severity is diagnostic priority, not a prediction of ranking impact.

Every stored finding also includes a classification (`deterministic`, `recommendation`, `review` or `information`) and an evidence source (`raw`, `rendered`, `graph`, `sitemap`, `external_api`, `lab` or `field`). These fields prevent severity from being mistaken for certainty and prevent one evidence mode from silently replacing another. Baseline fixture evidence is documented in [CONFORMANCE.md](./CONFORMANCE.md).

| ID | Area | Default | Detects | Important limitation |
|---|---|---:|---|---|
| AUD-01 | Responses | error | failing pages, redirects and internal targets | failures may be temporary |
| AUD-02 | Metadata | warning | missing/long/short and duplicate titles/descriptions | length thresholds are editorial guidance |
| AUD-03 | Headings | warning | absent/multiple H1 and title equality | semantic templates differ |
| AUD-04 | Canonicals | warning | absent/conflicting/invalid targets, chains, failures and noindex targets | canonicals are hints |
| AUD-05 | Indexability | error | non-200 and noindex conditions | engines apply other policies |
| AUD-06 | Sitemaps | info | missing coverage, broken entries and noindex entries | sitemap inclusion is not required |
| AUD-07 | Robots | warning | robots-blocked coverage | blocked pages cannot be fully inspected |
| AUD-08 | Duplicates | warning | exact normalized content duplicates | semantic near-duplicates are not yet classified |
| AUD-09 | Hreflang | warning | invalid codes, missing targets and non-reciprocal links | regional intent needs human review |
| AUD-10 | Images | warning | absent alt attributes and failing image targets | visual intent is unknown |
| AUD-11 | Architecture | warning | deep/orphan-like pages and nofollow observations | utility pages may be intentionally isolated |
| AUD-12 | Transport | warning | mixed content and selected defensive-header observations | not a security audit or ranking factor |
| AUD-13 | Structured data and Schema.org vocabulary | error | malformed JSON-LD, invalid `@type`/`@context`, missing Schema.org context, unknown terms and superseded terms | vocabulary is pinned to Schema.org 30.0; domain/range validation remains incremental |
| AUD-14 | Google search-feature profiles | error | pinned required/recommended property checks for Breadcrumb, Article and merchant Product JSON-LD | diagnostics follow reviewed public documentation and never guarantee rich-result eligibility or display |
| AUD-15 | Mobile and alternates | warning | responsive viewport, AMP self-reference and mobile-alternate self-reference | does not emulate mobile indexing or validate every reciprocal alternate |
| AUD-16 | Image delivery | warning | missing dimensions, oversized image responses and selected legacy formats | does not judge visual quality or compression efficiency |
| AUD-17 | PDF readiness | warning | bounded PDF title, language, tagging and encryption markers | byte-pattern diagnostics are not a complete PDF parser or accessibility audit |

AUD-13 uses the bundled Schema.org 30.0 release published 2026-03-19. The generated registry records the frozen upstream URL and SHA-256. AUD-14 uses a separately versioned Google Search Central profile bundle reviewed 2026-07-30; keeping these sources separate prevents Schema.org validity from being confused with Google feature support. The conformance manifest covers all 17 public families with positive and clean controls.

Remediation and full limitation text are available through `issue-explain`, the page-detail UI, and the MCP `issue_explain` tool. Raw and rendered findings are stored as distinct subject types with an `extraction_mode` evidence field. Rendered evidence never silently replaces raw evidence.
