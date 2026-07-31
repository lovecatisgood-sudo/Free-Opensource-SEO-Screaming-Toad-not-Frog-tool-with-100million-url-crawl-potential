# Lesson 6 — Configure Scope and Run the Crawl

**Target duration:** 4 minutes

## Narration

[On screen: Project → profile → scope preview → crawl.]

A project is the long-lived workspace for one audit boundary. A profile stores reusable crawl settings. A crawl is one run using a fixed snapshot of those settings.

Create a project, then enter the website's canonical public seed URL. Decide whether subdomains belong in scope. Exclude account, cart, checkout, internal-search, and faceted-navigation paths when they could create risk or crawler traps.

Set a maximum depth and URL ceiling. The ceiling is a safety budget, not a target. If a crawl stops with `limit_reached`, coverage is incomplete. Do not report it as a complete site audit.

[On screen: scope preview with accepted and rejected URLs plus reasons.]

Use scope preview before fetching. Confirm URL normalization, hostname behavior, exclusions, and redirect expectations. Start in raw mode because it is faster and preserves what the server originally returned. Use rendered mode when important content or metadata depends on JavaScript.

Confirm that you own the site or have permission to test it. Choose a crawl rate appropriate for the target's infrastructure. Then start the crawl.

Watch the lifecycle and terminal reason. `completed` means the application proved there was no outstanding accepted work. `limit_reached` means the budget stopped discovery. `cancelled` means the operator ended it. `failed` indicates a problem prevented completion.

Pause preserves durable state and stops new scheduling. Resume continues from the stored frontier. Crawl history and lifecycle events help distinguish an interrupted run from a completed audit.

[On screen: “Always report seed, scope, exclusions, limit, rendering mode, and terminal state.”]
