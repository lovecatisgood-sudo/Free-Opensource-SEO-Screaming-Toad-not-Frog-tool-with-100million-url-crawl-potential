# DJAI Academy Real-World Technical SEO Audit

Audit date: 2026-07-30
Target: `https://www.djai.academy/`
Authorization basis: course owner previously identified DJAI Academy as their website
Tool: SEO Screaming Toad `2.0.0-rc.4` release candidate
Crawl ID: `crawl_425921818cae33f4a03a417f7880cd3e`

Managed workbook: [`assets/reports/djai-audit-2026-07-30.xlsx`](assets/reports/djai-audit-2026-07-30.xlsx)
Workbook SHA-256: `0255dbd3f7dab4fe32add3777ce122cff2bce3c31a59affac43f4c7ef82ae44b`

## Scope and configuration

- Raw HTML mode; JavaScript rendering was not requested.
- Exact `www.djai.academy` host only; subdomains were excluded.
- Seed used the final canonical host after `djai.academy` returned a 308 redirect to `www.djai.academy`.
- Robots policy enforced; `robots.txt` allowed crawling and declared `https://www.djai.academy/sitemap.xml`.
- URL ceiling: 100,000. This was a safety maximum, not the achieved crawl size.
- Maximum depth: 50.
- Global concurrency: 16; per-host concurrency: 2; minimum host delay: 100 ms.
- Response compression disabled because this host previously returned incompatible responses when an `Accept-Encoding` header was explicitly sent.
- Same-host scope preview passed; `school.djai.academy` and unrelated hosts were excluded.

## Crawl outcome

| Measure | Observed result |
|---|---:|
| Terminal state | Completed |
| Discovered | 287 |
| Fetched | 287 |
| Analysed HTML pages | 222 |
| Fetch failures | 0 |
| 2xx responses | 287 |
| Sitemap records | 1, status `ok` |
| Errors reported by rules | 8 |
| Warnings reported by rules | 66 |
| Informational observations | 20 |
| Total findings | 94 |

The crawl ran from approximately 11:07:47 to 11:08:17 UTC—about 31 seconds. It completed because the same-host frontier emptied at 287 discovered URLs. It did **not** crawl 100,000 URLs.

## Practical priority: high to low

Crawler severity and practical priority are intentionally reported separately.

### High priority — no confirmed high-priority failure in this raw crawl

No broken responses, server failures, missing titles, missing descriptions, or missing canonicals were observed among the 222 analysed HTML pages. All 287 fetched resources returned a 2xx response and the declared sitemap was read successfully.

This does not prove the absence of every SEO problem. Rendered JavaScript, performance, accessibility, full Schema.org vocabulary and rich-result eligibility, Search Console coverage, analytics, backlinks, content quality, and real-user behavior were outside this crawl.

### Medium priority — review differentiation across 24 similar tool pages

AUD-08 reported 24 near-duplicate similarity observations, concentrated in QR generators, image tools, PDF tools, and media converters. Representative examples included:

- `/tools/qrgen/url-qr-code-generator/`
- `/tools/qrgen/wifi-qr-code-generator/`
- `/tools/resizeimg/image-to-100kb/`
- `/tools/media/compress-video/`
- `/tools/PDFTools/remove-pdf-metadata/en/`

**Possible implication:** if pages differ mainly by a tool label while sharing most explanatory content, search and AI retrieval systems may have difficulty identifying the unique purpose and value of each page. Similarity can also divide internal signals or create low-value landing pages.

**Required human review:** similarity is expected in a product family. Confirm that each page serves a distinct task, has unique instructions and examples, explains its specific input/output and limitations, and provides enough original value to justify a separate canonical indexable URL. Consolidate only when intent is genuinely the same.

The rule limitation states that hashes do not capture every semantic duplicate. A warning is a review queue, not proof that consolidation is required.

### Medium-to-low priority — evaluate eight category-filter URLs and their intended search role

AUD-05 labelled eight 200-status category-filter pages as indexability errors because they carried `noindex`. Inspection found a consistent pattern:

- each URL used `noindex, follow`;
- each canonicalized to its Thai or English blog hub;
- none appeared in the XML sitemap;
- the filter pages repeated the corresponding hub title and description.

Examples included `/blog/?category=Siamese+Cat+Dev` and `/blog/en/?category=Guides`.

**Interpretation:** this looks like deliberate duplicate-control rather than an accidental blocking error. If category-filter URLs exist only for user filtering, the current noindex, canonical, and sitemap alignment may be correct. If the business wants indexable category landing pages, then these pages need stable clean URLs, unique titles and descriptions, self-referencing canonicals, useful category copy, reciprocal internal links, and sitemap inclusion after verification.

**Possible implication if intent is wrong:** a valuable category landing page cannot compete in search while noindex. **Recommended action:** document the product decision; do not remove noindex simply to clear a red dashboard count.

### Low priority — metadata length review

AUD-02 reported:

| Observation | Count |
|---|---:|
| Titles above the configured 60-character threshold | 12 |
| Titles below the configured 30-character threshold | 3 |
| Meta descriptions above 160 characters | 7 |
| Meta descriptions below 70 characters | 13 informational |

These thresholds are editorial diagnostics, not ranking rules. Search engines may rewrite or truncate snippets. Review important pages for clarity and uniqueness rather than shortening text mechanically.

Representative longer descriptions appeared on the development service page and several English blog articles. Shorter descriptions were concentrated in spreadsheet/document tool pages.

### Low priority — repeated hub metadata on intentionally noindexed filters

Ten duplicate-title and ten duplicate-description warnings came from the Thai and English blog hubs plus their category-filter variations. Because the filters are noindex, canonicalized to the hubs, and excluded from the sitemap, these duplicates appear to be a consequence of the intended filter strategy.

No metadata change is required unless the product decision changes and the category pages become standalone indexable landing pages.

### Informational — seven title/H1 equality observations

Seven pages used the same text for the title and H1. Equality is not inherently an SEO defect. The title can include branding or search context while the H1 remains a clear visible heading, but changing text only to make two fields different would not automatically improve the page.

## Positive observations

- Zero fetch failures and no observed 3xx/4xx/5xx final responses within the accepted same-host set.
- All 222 analysed HTML pages had a title, meta description, and canonical URL.
- The sitemap was available and parsed successfully.
- The eight filter pages showed internally consistent noindex, canonical, and sitemap behavior.
- Raw evidence remained explicit; the report does not imply JavaScript-rendered verification.

## Recommended next audit pass

1. Manually sample the 24 similarity pairs and classify them as intentional product-family similarity, insufficient differentiation, or duplicate intent.
2. Record whether blog category pages are user-only filters or intended organic landing pages.
3. Improve titles/descriptions only on commercially or editorially important pages where clarity is genuinely weak.
4. Run a smaller rendered-mode sample for JavaScript-dependent templates.
5. Validate Article/Breadcrumb and other applicable structured data using appropriate vocabulary and search-feature validators.
6. Use Search Console for actual indexing and search performance; do not infer either from this crawl.

## Limitations

This was one raw-mode, same-host crawl at one point in time. It did not include JavaScript rendering, Search Console, analytics, field performance, PageSpeed, accessibility, backlink analysis, log-file analysis, semantic content-quality review, or full rich-result validation. Findings do not guarantee or predict rankings, traffic, conversions, indexing, or AI citations.
