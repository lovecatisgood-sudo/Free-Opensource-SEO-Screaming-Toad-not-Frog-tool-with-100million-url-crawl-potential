# Lesson 3 — URL Crawling and Screaming Frog

**Target duration:** 4 minutes

## Narration

[On screen: Seed URL → fetch → extract links → normalize → scope check → frontier.]

A URL crawler begins with a seed page. It requests that page, records the response, extracts links, normalizes the discovered URLs, rejects duplicates and out-of-scope destinations, and places accepted URLs into a queue called a frontier. The process repeats until the queue is empty or a safety limit is reached.

This produces a graph of the website. Pages are nodes and links are relationships. From this graph, an auditor can identify broken internal destinations, redirect chains, deep pages, duplicate content, isolated pages, and conflicting signals between sitemaps, canonicals, and internal links.

A responsible crawler also needs scope controls, URL and depth ceilings, crawler-trap protection, per-host rate limits, bounded retries, and robots-policy enforcement. Faster is not automatically better if the crawler overloads a website.

[On screen: Screaming Frog SEO Spider, with “Independent product; no affiliation.”]

Screaming Frog SEO Spider is one of the industry's most respected technical SEO crawlers. Professionals value its mature interface, extensive configuration, exports, JavaScript rendering, crawl comparisons, custom extraction, scheduling, integrations, and years of real-world use.

Its official free version is limited to 500 URLs. At the time this script was checked, the paid license was listed at 199 British pounds per user per year, with currency-adjusted pricing. That is roughly a two-hundred-pound—or two-hundred-plus-euro—annual software decision, not an exact permanent euro price.

Screaming Frog is respected because it does much more than download pages. It helps experts interrogate crawl evidence. That is the benchmark our open-source project learns from, while following its own architecture, code, and product identity.
