# Lesson 8 — Essential SEO and ASO Checks

**Target duration:** 5 minutes

## Narration

[On screen: rapid technical audit checklist.]

Toad's audit rules cover the main technical signals needed for a normal website review.

Start with responses and redirects. A broken navigation target needs attention, while an intentionally retired URL may correctly return 404 or 410. Remove unnecessary redirect chains and point internal links toward the intended final URL.

Titles, descriptions, and headings clarify page purpose. Missing or duplicated values deserve review, but length thresholds are editorial guidance—not universal ranking laws. Look for template-wide causes.

Keep four systems distinct. `robots.txt` guides crawler access. A `noindex` directive asks compliant engines not to index a resource. A canonical tag suggests the preferred representative among duplicate or similar URLs. An XML sitemap helps discovery and submission. Sitemap inclusion does not guarantee indexing, and a canonical is a hint rather than an absolute command.

[On screen: align internal links, redirects, canonical, indexability, and sitemap.]

Duplicate checks reveal repeated content. Hreflang checks review language codes, targets, and reciprocity. Image checks identify missing alt attributes and failed image resources. Architecture checks use depth and link relationships to find deep or orphan-like pages.

Structured-data checks detect malformed JSON-LD and selected `@type` and `@context` problems. They do not validate the entire Schema.org vocabulary or guarantee rich-result eligibility. Structured data must truthfully match visible content.

For AI search, the same foundation matters: crawlable and indexable delivery, clear headings, coherent internal relationships, consistent entities, original information, and reliable evidence. Toad can test several technical preconditions; it cannot guarantee retrieval or citation by an AI system.

[On screen: “Observed evidence” versus “human judgment.”]

Prioritize by impact and confidence. A broken shared canonical may affect an entire section. A missing description on an intentionally noindex page may not matter. Fix source causes, not spreadsheet symptoms.
