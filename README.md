<div align="center">
  <img src="web/app/public/screaming-toad.png" alt="SEO Screaming Toad open-source SEO crawler logo" width="180">

# Free Open-Source SEO Screaming Toad — Not Frog

### A local-first Screaming Frog alternative with an MCP server and theoretical 100M+ URL campaign architecture

[![CI](https://github.com/lovecatisgood-sudo/seo-screaming-toad/actions/workflows/ci.yml/badge.svg)](https://github.com/lovecatisgood-sudo/seo-screaming-toad/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/license-MIT-8dc63f.svg)](LICENSE)
[![Core: Go](https://img.shields.io/badge/core-Go-00ADD8?logo=go&logoColor=white)](go.mod)
[![MCP server](https://img.shields.io/badge/MCP-server-8A2BE2)](docs/MCP.md)
[![Local first](https://img.shields.io/badge/data-local--first-f28c28)](docs/SECURITY_MODEL.md)

[Quick start](#quick-start) · [MCP for AI agents](#mcp-server-for-ai-agents) · [SEO checks](#seo-audit-coverage) · [Architecture](docs/ARCHITECTURE.md) · [Contributing](CONTRIBUTING.md)
</div>

SEO Screaming Toad, also called **DJAI Toad**, is a free, open-source technical SEO crawler and website audit tool. It crawls sites locally, preserves evidence in SQLite, audits raw and JavaScript-rendered pages, compares recrawls, exports reports, and exposes a guarded **Model Context Protocol (MCP) server** so AI agents can operate real SEO audits.

If you are searching for an **open-source Screaming Frog alternative**, a self-hosted SEO spider, a Go website crawler, or an MCP-powered SEO audit tool, this project is built for that workflow—without uploading your crawl database to a hosted service.

> [!IMPORTANT]
> SEO Screaming Toad is an independent project and is not affiliated with or endorsed by Screaming Frog Ltd. “Screaming Frog” is used only for descriptive comparison; this project does not claim feature parity. The current project is a release candidate, so review the [project state](docs/PROJECT_STATE.md) before production use.

## Why SEO Screaming Toad?

| Capability | What you get |
|---|---|
| Open source | MIT-licensed Go crawler, React dashboard, CLI, local API, and MCP server |
| AI-agent operation | 23 bounded MCP tools for projects, crawls, evidence, comparisons, and reports |
| Evidence-backed audits | Versioned findings retain rule ID, severity, subject, evidence, and limitations |
| Raw + rendered analysis | Raw HTML stays distinct from optional JavaScript-rendered evidence |
| Local-first data | SQLite/WAL storage and managed exports remain on your machine |
| Safe crawling | Robots policy, DNS/IP guards, redirect validation, TLS verification, budgets, and scope controls |
| Recoverable jobs | Pause, resume, cancel, event timeline, checkpoints, and interrupted-run recovery |
| Large-site architecture | Segmented campaigns and durable frontiers designed toward theoretical 100M+ URL operation |

## Quick start

Requirements: Go version from [`.go-version`](.go-version). Node and pnpm are needed only for rendered mode or frontend development.

```bash
git clone https://github.com/lovecatisgood-sudo/seo-screaming-toad.git
cd seo-screaming-toad
make bootstrap
go run ./cmd/seo-auditor
```

Open [http://127.0.0.1:7331](http://127.0.0.1:7331), create a project and crawl profile, preview its scope, and start an authorized public-site audit.

For a sandboxed development environment:

```bash
docker compose up -d dev
docker compose exec dev bash
make test
```

See [DEVELOPMENT.md](DEVELOPMENT.md) for container boundaries and the host-toolchain path. Only crawl websites you own or are authorized to test.

## MCP server for AI agents

SEO Screaming Toad includes `seo-auditor-mcp`, a stdio MCP server built with the official Go SDK. An MCP-compatible AI agent can create bounded profiles, preview scope, start or control crawls, inspect SEO issues and page evidence, compare runs, and export managed reports.

```text
AI agent / MCP client
        │ stdio
        ▼
seo-auditor-mcp
        │ authenticated loopback API
        ▼
SEO Screaming Toad → guarded crawler → SQLite evidence
```

Build the two executables:

```bash
mkdir -p bin
go build -o bin/seo-auditor ./cmd/seo-auditor
go build -o bin/seo-auditor-mcp ./cmd/seo-auditor-mcp
```

Start `bin/seo-auditor`, configure your MCP client to run the absolute path to `bin/seo-auditor-mcp`, and call `health_get`. A connected result reports `api_connected: true`.

The MCP interface deliberately exposes no generic shell, arbitrary filesystem access, SQL execution, or unrestricted HTTP fetch. Read the complete [MCP setup and tool guide](docs/MCP.md) or copy the [generic MCP client configuration](examples/mcp-client.json).

## SEO audit coverage

The versioned audit engine currently covers:

- HTTP failures, redirects, and broken internal targets;
- titles, meta descriptions, headings, and duplicate metadata;
- canonical tags, canonical chains, conflicts, and invalid targets;
- indexability, robots directives, robots.txt, and sitemap coverage;
- exact and near-duplicate content signals;
- hreflang validity, targets, and reciprocity;
- image alt attributes and failing image resources;
- crawl depth, orphan-like pages, internal links, and nofollow observations;
- mixed content and selected defensive headers;
- JSON-LD structured-data syntax and basic shape validation;
- raw-versus-rendered SEO differences for JavaScript sites.

Every check documents its limitations. Findings are technical observations—not promises about indexing, rankings, traffic, or rich-result eligibility. See the [audit rule catalog](docs/RULE_CATALOG.md).

## Technical crawler features

- Concurrent Go crawl engine with per-host politeness and bounded retries
- Robots.txt enforcement and sitemap discovery
- URL normalization, deduplication, query controls, and crawler-trap protection
- Prohibited-network and mixed-DNS-answer rejection for SSRF resistance
- Approved-address connection pinning with normal TLS hostname verification
- Response size, header, timeout, redirect, depth, and URL ceilings
- SQLite/WAL durable frontier with serialized writes and recovery
- Optional isolated TypeScript/Chromium rendering worker
- Dense technical React dashboard inspired by desktop SEO crawlers
- Local authenticated API, JSON CLI, MCP server, CSV, NDJSON, and XLSX reports
- Crawl-to-crawl comparison for added, removed, changed, new, and fixed results

## About the 100M+ URL architecture

The system has segmented-campaign and distributed-coordination prototypes intended to process bounded batches repeatedly instead of holding an entire enormous crawl in memory. Local synthetic persistence/analysis campaigns have completed at 1 million and 5 million URLs, and a 100,000-URL single-crawl gate has passed.

**100M+ is a theoretical architectural potential—not a verified live-network capacity, benchmark, support promise, or guarantee.** Actual scale depends on hardware, storage, site behavior, politeness settings, rendering, network conditions, and operational design. Read the [scale strategy](docs/SCALE_STRATEGY.md) and [benchmark evidence](docs/benchmarks/2026-07-30-segmented-campaign.md).

## Is this a Screaming Frog replacement?

It depends on the job. SEO Screaming Toad is attractive when you need open source, local data, inspectable rules, automation, or an MCP server for AI agents. Screaming Frog SEO Spider is a mature commercial product with years of integrations, workflows, and production refinement that this release candidate does not yet match.

Our goal is to build a credible community-owned alternative for normal technical website audits, while being transparent about gaps. See the [quality roadmap](docs/QUALITY_ROADMAP.md) and help prioritize the next improvement.

## Project structure

```text
cmd/                    Go application, CLI, MCP, SBOM, and scale commands
internal/crawler/       crawl engine, robots, sitemaps, links, and traps
internal/fetchpolicy/   target safety, transport, redirects, and scope
internal/rules/         versioned technical SEO audit rules
internal/database/      SQLite/WAL persistence, frontier, and recovery
internal/mcpserver/     bounded Model Context Protocol tools
web/app/                React technical audit dashboard
web/renderer/           isolated optional JavaScript renderer
docs/                   PRD, architecture, operations, security, and evidence
```

## Contributing

Bug reports, new SEO rules, UX improvements, platform testing, documentation, and thoughtful performance work are welcome. Start with [CONTRIBUTING.md](CONTRIBUTING.md), review the [security model](docs/SECURITY_MODEL.md), and choose an issue that matches your experience.

Please give the repository a ⭐ if an open-source, MCP-enabled SEO crawler is useful to you. Stars help more SEO practitioners and AI builders discover the project.

## Security and responsible use

Do not put vulnerabilities, private URLs, credentials, or raw client crawl data in public issues. Follow [SECURITY.md](SECURITY.md) for responsible disclosure. The crawler must only be used against systems you are authorized to audit.

## License and credits

Released under the [MIT License](LICENSE). Third-party dependency and source notices are recorded in [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md).

Created by **[Siamese Cat Dev](https://github.com/lovecatisgood-sudo)** with DJAI.
