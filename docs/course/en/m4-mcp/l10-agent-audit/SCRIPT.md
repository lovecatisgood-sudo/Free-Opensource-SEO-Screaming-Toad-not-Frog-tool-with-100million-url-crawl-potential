# Lesson 10 — Connect and Run an Agent Audit

**Target duration:** 5 minutes

## Narration

[On screen: Codex or Claude Code prompt card.]

Ask your coding agent:

“Pull the official SEO Screaming Toad repository from `lovecatisgood-sudo`. Read `README.md`, `docs/MCP.md`, `docs/SECURITY_MODEL.md`, and `docs/PROJECT_STATE.md`. Build `seo-auditor` and `seo-auditor-mcp` in the local `bin` folder. Start the app on `127.0.0.1:7331`, then show the exact MCP configuration using the absolute executable path. Do not crawl until I approve the scope.”

The main `seo-auditor` process owns the dashboard, crawler, API, and evidence store. The `seo-auditor-mcp` process is the stdio adapter launched by the AI client. Start the main application first.

MCP configuration locations change between clients and versions. Ask the agent to check the current Codex or Claude Code documentation instead of guessing an old path. Once configured, call `health_get`. A ready connection reports `api_connected: true`.

[On screen: health → projects → profiles → preview → approval → start → status → evidence.]

Now use an evidence-first audit prompt:

“Using SEO Screaming Toad MCP, prepare an audit of this website, which I am authorized to crawl. Use raw mode, a maximum of 1,000 URLs, no subdomains, and exclude account, cart, checkout, and internal-search paths. Preview scope and explain exclusions. Do not start until I confirm.”

After confirmation, start once with an idempotency key and poll status until a terminal state appears. Read the summary, then paginate issues and pages. Do not dump unlimited output.

Ask the agent:

“Prioritize errors and template-wide warnings. For every recommendation, include the rule ID, affected count, representative URLs, observed evidence, limitation, likely root cause, and verification step. Separate facts from inference. Do not modify the website yet.”

The agent operates the machinery. The user retains control of authorization, interpretation, and changes.
