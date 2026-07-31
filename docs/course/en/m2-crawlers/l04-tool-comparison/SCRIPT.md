# Lesson 4 — Meet SEO Screaming Toad and Compare the Tools

**Target duration:** 4 minutes

## Narration

[On screen: Toad logo, dashboard, GitHub repository, and “DJAI Toad.”]

SEO Screaming Toad — Not Frog, also called DJAI Toad, is a free, MIT-licensed technical SEO crawler. It was created by Siamese Cat Dev from DJAI Academy together with DJAI trainers and community members.

The crawler core is written in Go, the dashboard uses React, and audit evidence is kept locally in SQLite. It supports raw HTML auditing, optional JavaScript-rendered evidence, crawl history, comparisons, reports, a command-line interface, a local API, and an MCP server for AI agents.

For many normal technical website audits, Toad covers much of the core workflow: responses, redirects, broken internal targets, titles, descriptions, headings, canonicals, indexability, robots rules, sitemaps, duplicates, hreflang, images, architecture, basic structured-data checks, and raw-versus-rendered differences.

[On screen: balanced comparison.]

It does not claim complete Screaming Frog parity. Screaming Frog currently has broader mature capabilities such as custom extraction, scheduling, analytics and Search Console integrations, PageSpeed and link metrics, accessibility auditing, spelling and grammar checks, forms authentication, and commercial support.

Toad has different advantages. It is open source, local-first, inspectable, and free to use. Findings retain versioned rule IDs, structured evidence, remediation, and explicit limitations. Its first-class MCP server gives AI agents 36 bounded SEO tools without giving the crawler a generic shell, unrestricted SQL, or arbitrary web-fetch access.

The best comparison is a representative authorized crawl using similar settings. Compare discovered URLs, false positives, JavaScript evidence, reports, and the capabilities your workflow needs. Some experts may replace a core workflow with Toad; others may use both tools.

[On screen: “Choose by evidence and workflow—not slogans.”]
