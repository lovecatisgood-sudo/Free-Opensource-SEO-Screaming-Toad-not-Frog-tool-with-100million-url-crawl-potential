# Audit rule catalog

Rules are versioned and every finding stores its rule ID, version, subject, severity and JSON evidence. Severity is diagnostic priority, not a prediction of ranking impact.

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

Remediation and full limitation text are available through `issue-explain`, the page-detail UI, and the MCP `issue_explain` tool. Raw and rendered findings are stored as distinct subject types with an `extraction_mode` evidence field. Rendered evidence never silently replaces raw evidence.
