# Bonus Lesson — A Real-World SEO Audit of DJAI Academy

**Target duration:** 6 minutes, optional and outside the core 60-minute course

## Narration

[On screen: `01-live-crawl-completed.png`.]

Let us finish with a genuine real-world audit of DJAI Academy using SEO Screaming Toad.

We configured a raw HTML crawl of the final `www.djai.academy` host. Subdomains were excluded, robots rules remained enforced, and the URL ceiling was set to 100,000.

Remember: a ceiling is not a target or an achieved count. The website contained 287 discoverable same-host resources. Toad fetched all 287, analyzed 222 HTML pages, recorded zero failures, and completed in about 31 seconds.

[On screen: `02-live-audit-summary.png`. Highlight 8 errors, 66 warnings, 20 information, and 287 fetched.]

The dashboard reported 94 findings: 8 errors, 66 warnings, and 20 informational observations. It would be easy to announce that the website has eight serious SEO errors. That would also be poor auditing.

We must move from severity to evidence.

[On screen: `03-noindex-review.png`, followed by `04-indexability-rule-explanation.png`.]

The eight red findings were category-filter URLs returning 200 with `noindex`. We inspected them further. Each used `noindex, follow`, canonicalized to the correct Thai or English blog hub, and stayed out of the sitemap. Their repeated titles and descriptions also matched the hub.

This pattern looks deliberate. If these URLs are only interactive filters, the signals may be correct. Removing noindex merely to clear a red count could create indexable duplicates. If DJAI wants category pages to rank, the correct fix is larger: create stable clean URLs, add unique category content and metadata, use self-referencing canonicals, link them internally, and include them in the sitemap after verification.

So our high-priority result is surprising: this raw crawl found no confirmed high-priority technical failure. All accepted resources returned 2xx responses, zero fetches failed, the sitemap worked, and every analyzed page had a title, description, and canonical.

[On screen: `05-near-duplicate-review.png`.]

The most valuable medium-priority review is content similarity. Toad reported 24 near-duplicate signals across related QR, image, PDF, and media tools.

Similarity is expected in a product family. The warning does not prove that pages should be merged. We should manually compare each pair and ask: does each page answer a distinct intent? Does it have specific instructions, examples, limitations, and original value? If the only difference is a tool name, search and AI systems may struggle to identify why each URL deserves separate retrieval. If the tasks are genuinely different, we improve differentiation rather than consolidating blindly.

[On screen: `06-live-page-inventory.png`.]

Lower-priority observations included 12 titles above the configured length guideline, three below it, seven long descriptions, and 13 short descriptions. These thresholds are editorial diagnostics, not ranking rules. We should improve important snippets for clarity, not edit text mechanically to satisfy a number.

Seven informational findings showed pages where the title and H1 were identical. That is not inherently wrong. Changing one merely to make them different creates work without proven value.

[On screen: `07-live-crawl-history.png`.]

The correct conclusion is evidence-based and appropriately limited. DJAI Academy showed a strong raw technical baseline, one meaningful content-differentiation review area, an intentional category-filter strategy to confirm, and several lower-priority metadata improvements.

This audit did not test JavaScript-rendered output, performance, accessibility, Search Console, analytics, backlinks, or full rich-result eligibility. It cannot guarantee rankings or AI citations.

That is how Toad works in reality: crawl within an authorized scope, inspect the evidence behind every finding, reprioritize with human judgment, fix the source cause, and recrawl to verify.
