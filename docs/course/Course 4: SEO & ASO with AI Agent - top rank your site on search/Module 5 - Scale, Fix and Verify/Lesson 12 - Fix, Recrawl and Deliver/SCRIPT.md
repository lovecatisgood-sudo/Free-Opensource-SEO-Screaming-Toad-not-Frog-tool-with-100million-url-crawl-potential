# Lesson 12 — Fix, Recrawl, and Deliver

**Target duration:** 5 minutes

## Narration

[On screen: Audit → prioritize → code → test → deploy → recrawl → compare.]

The audit creates value only when evidence becomes a safe, verified improvement.

Choose a small group of related, high-confidence findings. Template-wide problems are efficient: a title generator, canonical component, navigation link, sitemap builder, hreflang helper, or response configuration may explain hundreds of URLs.

Give a coding agent the evidence and boundaries:

[On screen: prompt card.]

“Review these Toad findings and the affected source code. Separate confirmed observations from hypotheses. Find the shared root cause. Propose the smallest safe fix, preserve accessibility, localization, security, and existing behavior, and add tests for the intended SEO behavior. Do not deploy until I review the diff.”

For JavaScript websites, ask the agent to inspect both raw HTML and the rendered document. Critical title, canonical, robots, hreflang, headings, content, and internal links should appear reliably for the intended crawler behavior.

Run tests and a production build. Review the code difference. Deploy only through an authorized process, then inspect representative live URLs before recrawling.

Create a new crawl with settings compatible with the baseline. Compare added, removed, and changed pages plus new and fixed issues. Open representative evidence instead of trusting only the summary count.

[On screen: one fixed issue and one regression caught by comparison.]

Watch for regressions. A canonical fix must not point every page to the home page. Removing `noindex` must not expose private utilities. A sitemap should not submit redirects or failures.

Deliver a concise report containing scope, configuration, terminal state, limitations, prioritized evidence, changes, tests, recrawl comparison, and remaining work. Separate what the crawler observed from what you inferred.

You now have the entire workflow: understand SEO and AI search, crawl responsibly, inspect evidence, use MCP to reduce technical barriers, fix with an AI coding agent, and verify the outcome.

[On screen: “Evidence. Judgment. Fix. Verification.” Then DJAI Academy and project GitHub calls to action.]
