# Course Screenshot and Visual Capture Plan

Status: proposed for review
Course target: 60 minutes
Planned core-course screenshot assets: **26** — 8 existing captures and 18 new captures. The optional real-world bonus adds 7 completed live-audit captures.

## Capture standards

- Capture new full-screen images at 1920×1080 where possible; retain a readable 16:9 safe area for video.
- Use a dedicated training project and authorized fixture/public target.
- Use real product output. Do not fabricate findings, counts, completed states, or MCP responses.
- Hide tokens, cookies, private URLs, client data, local usernames, and machine-specific home paths.
- Keep an untouched master and a presentation copy with callouts added by HyperFrames.
- Use browser zoom and table widths consistently so text remains readable on mobile course playback.
- Record product version, Git commit, crawl ID, target, scope, and capture date in a private capture log.
- Recheck the Screaming Frog price immediately before production; its pricing is time-sensitive.
- The installed comparison application is Screaming Frog SEO Spider 24.3 at `/usr/bin/screamingfrogseospider`. Record the version in each comparison caption and never expose licence details.
- For product comparisons, crawl the same authorized target in raw mode with equivalent hostname scope and exclusions. Use a maximum of 500 URLs so the exercise remains reproducible with Screaming Frog's free version.

## Lesson-by-lesson plan

### Lesson 1 — What SEO and AI Search Optimization Mean

**Screenshots:** none.

Use animated diagrams for discovery, crawling, understanding, indexing, conventional results, and AI-assisted answers. A product screenshot would distract from the foundational concepts.

### Lesson 2 — Why Websites Invest in SEO and ASO

**Screenshots:** none.

Use a compact animated value chain and a clearly labelled market-pricing chart derived from the cited survey data. Do not screenshot agency marketing pages or imply that survey ranges are DJAI prices.

### Lesson 3 — URL Crawling and Screaming Frog

**New screenshots:**

1. `lesson-03-screaming-frog-official-pricing.png` — official Screaming Frog pricing page showing the 500-URL free limit, £199 annual base price, and representative paid features. Teaching purpose: establish the commercial benchmark with a dated source.
2. `lesson-03-screaming-frog-local-app.png` — locally installed Screaming Frog SEO Spider 24.3 at its main crawl interface, with no licence key, account details, or private history visible. Teaching purpose: establish that Screaming Frog is also a local desktop application and introduce its technical workbench.

Use animation—not a screenshot—for the seed, fetch, extract, normalize, scope, and frontier loop.

### Lesson 4 — Meet SEO Screaming Toad and Compare the Tools

**Existing screenshots:**

1. `docs/images/dashboard/01-dashboard-home.png` — complete technical workbench and product identity.
2. `docs/images/dashboard/03-audit-findings.png` — evidence-backed audit findings.
3. `docs/images/dashboard/08-djai-services.png` — DJAI Academy, development, school-community, and service presence.

**New screenshot:**

4. `lesson-04-official-github-repository.png` — repository header with exact owner, full repository name, open-source license area, README hero, and MCP badge. Teaching purpose: show the official source and community identity.

**New local comparison screenshots:**

5. `lesson-04-screaming-frog-page-inventory.png` — Screaming Frog's Internal/HTML inventory after crawling the same authorized target used by Toad, with equivalent raw-mode scope and a maximum of 500 URLs.
6. `lesson-04-screaming-frog-issues.png` — Screaming Frog's issue overview for that same crawl. Teaching purpose: compare how each product organizes evidence and navigation, without claiming that issue names or counts are directly equivalent.

HyperFrames should pair capture 5 with Toad's page inventory and capture 6 with Toad's findings view. Captions must say that rule definitions, defaults, normalization, and crawl settings can cause different results even against the same target.

### Lesson 5 — Install and Launch Safely

**New screenshots:**

1. `lesson-05-agent-repository-review.png` — Codex or Claude Code in a clean workspace after reading the README, security model, development guide, and project state. Teaching purpose: model safe inspection before execution.
2. `lesson-05-build-and-local-launch.png` — successful documented build/start output with the loopback address visible and sensitive local paths cropped. Teaching purpose: confirm that installation reached a working local service.

**Reused screenshot:**

3. `docs/images/dashboard/01-dashboard-home.png` — first browser launch at `127.0.0.1:7331`.

### Lesson 6 — Configure Scope and Run the Crawl

**Existing screenshot:**

1. `docs/images/dashboard/02-crawl-configuration.png` — seed, URL ceiling, rendering mode, compression, subdomain policy, and exclusions.

**New screenshots:**

2. `lesson-06-scope-preview.png` — accepted and rejected candidate URLs with normalized values and reasons. Teaching purpose: prove scope before fetching.
3. `lesson-06-crawl-lifecycle.png` — one presentation composite made from genuine running, paused/resumed, and completed or `limit_reached` states. Teaching purpose: distinguish lifecycle and terminal meanings without spending three separate frames.

### Lesson 7 — Read the Dashboard and Evidence

**Existing screenshots:**

1. `docs/images/dashboard/03-audit-findings.png` — filtering and prioritization.
2. `docs/images/dashboard/04-page-inventory.png` — searchable response and metadata inventory.
3. `docs/images/dashboard/05-page-evidence.png` — headings, links, and page-level proof.
4. `docs/images/dashboard/06-rule-explanation.png` — rule ID, version, remediation, structured evidence, and limitation.

HyperFrames should pan and zoom into the relevant field rather than showing all four as static full-screen images.

### Lesson 8 — Essential SEO and ASO Checks

**New screenshots from an authorized deterministic fixture crawl:**

1. `lesson-08-response-metadata-findings.png` — broken target, redirect chain, duplicate title, missing description, and heading examples.
2. `lesson-08-indexability-signals.png` — canonical, robots/noindex, sitemap, and hreflang evidence arranged as one genuine dashboard composite. Teaching purpose: show that the systems are related but different.
3. `lesson-08-structured-and-rendered-evidence.png` — malformed JSON-LD evidence and one raw-versus-rendered difference. Teaching purpose: demonstrate both capability and limitation.

These should come from a controlled training fixture so every visible problem is intentional, reproducible, and safe to explain.

### Lesson 9 — What MCP Does and Why It Matters

**Screenshots:** none.

Use an animated architecture diagram for AI client → stdio MCP adapter → authenticated loopback API → guarded crawler and local evidence. Animate blocked generic shell, SQL, arbitrary fetch, and arbitrary-path actions. This relationship is clearer as a diagram than as terminal output.

### Lesson 10 — Connect and Run an Agent Audit

**New screenshots:**

1. `lesson-10-mcp-client-configuration.png` — current MCP client configuration with the absolute executable path anonymized and loopback host/port visible.
2. `lesson-10-health-and-scope.png` — actual `health_get` result with `api_connected: true`, followed by a genuine bounded scope-preview result.
3. `lesson-10-evidence-first-agent-result.png` — agent response containing rule ID, representative URL, observed evidence, limitation, likely cause, and verification step. Teaching purpose: demonstrate what a good AI-assisted audit answer looks like.

The video can animate intermediate tool calls rather than requiring a separate screenshot for every MCP operation.

### Lesson 11 — The Theoretical 100M+ Architecture

**Screenshots:** none.

Use a custom animation showing campaign coordinator, durable frontier, bounded URL segments, host-politeness ownership, workers, immutable results, and analytical store. Display the theoretical-capacity disclaimer throughout. A database or terminal screenshot would make an unverified capacity idea look like benchmark proof.

### Lesson 12 — Fix, Recrawl, and Deliver

**Existing screenshot:**

1. `docs/images/dashboard/07-crawl-history.png` — baseline and subsequent crawl history.

**New screenshots:**

2. `lesson-12-fix-diff-and-tests.png` — narrow code diff and passing relevant test from the training fixture. Teaching purpose: connect crawler evidence to a source-level correction.
3. `lesson-12-crawl-comparison.png` — genuine added, removed, changed, new-issue, and fixed-issue comparison after the recrawl.
4. `lesson-12-managed-report.png` — generated XLSX/CSV artifact metadata and a safe preview of the report. Teaching purpose: show the professional handoff.

**Optional reused outro screenshot:**

5. `docs/images/dashboard/08-djai-services.png` — use briefly with the DJAI Academy, community, GitHub-star, web-development, and software-development calls to action.

### Bonus Lesson — Real-World DJAI Academy Audit

**Completed live screenshots:**

1. `01-live-crawl-completed.png` — completed state: 287 discovered/fetched, 222 analysed, zero failures.
2. `02-live-audit-summary.png` — 8 errors, 66 warnings, 20 informational observations and 94 total findings.
3. `03-noindex-review.png` — eight AUD-05 observations filtered by their `noindex` evidence.
4. `04-indexability-rule-explanation.png` — versioned remediation, evidence, and rule limitation.
5. `05-near-duplicate-review.png` — 24 AUD-08 similarity findings for template-heavy tool pages.
6. `06-live-page-inventory.png` — real 200-status page inventory with raw titles and rendered mode clearly not requested.
7. `07-live-crawl-history.png` — durable completed crawl record.

These images live beneath the bonus lesson's `ASSETS/screenshots/` folder. They use genuine data from the authorized `www.djai.academy` raw crawl completed on 2026-07-30.

## New screenshot storage structure

```text
ASSETS/
  screenshots/
    lesson-03/
    lesson-04/
    lesson-05/
    lesson-06/
    lesson-08/
    lesson-10/
    lesson-12/
  capture-log.md
```

The existing repository dashboard screenshots remain in `docs/images/dashboard/` and should be referenced rather than duplicated.

## Capture sequence

1. Build one deterministic training fixture containing the issues needed by Lessons 8 and 12.
2. Run a raw baseline Toad crawl and capture Lessons 4, 6, 7, and 8.
3. Run Screaming Frog 24.3 against the same authorized target with equivalent raw-mode scope and a maximum of 500 URLs; record configuration differences and capture Lessons 3 and 4.
4. Connect MCP and capture Lesson 10 against the Toad baseline crawl.
5. Apply one narrow fixture fix, run tests, and capture the source diff.
6. Recrawl with identical Toad settings and capture Lesson 12 comparison/report evidence.
7. Capture time-sensitive external pages—GitHub and Screaming Frog pricing—last.
8. Verify every screenshot against its script, then create presentation crops and callouts.
