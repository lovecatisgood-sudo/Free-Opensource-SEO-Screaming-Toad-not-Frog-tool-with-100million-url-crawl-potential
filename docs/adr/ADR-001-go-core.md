# ADR-001: Use Go for the product core

Status: Accepted  
Date: 2026-07-30

## Context

SEO Auditor is a local-first, security-sensitive crawler expected to coordinate many bounded network requests, durable queue operations, parsing tasks, audit rules, exports, a local API, a CLI, and an MCP server. It initially targets normal website audits up to approximately 100,000 URLs, with room to grow.

The original architecture proposed TypeScript/Node for the whole product. Review of the requirements and the mature MIT-licensed SEonaut crawler demonstrated that Go maps naturally to crawler scheduling, cancellation, parsing, database services, and cross-platform delivery. TypeScript, Python, Go, and C# all have official Tier 1 MCP SDKs, so MCP does not require a Node core.

JavaScript rendering remains important, and Playwright's primary implementation is in the Node ecosystem.

## Decision

Use Go for:

- the crawl scheduler and persistent frontier;
- URL normalisation and target-security policy;
- DNS, HTTP, TLS, redirect and response-limit enforcement;
- robots.txt and sitemap handling;
- raw HTML extraction and audit rules;
- SQLite persistence, comparison and exports;
- the loopback API, CLI and stdio MCP server;
- worker supervision and release binaries.

Use TypeScript for:

- the React web interface;
- the optional isolated Playwright renderer sidecar;
- generated clients/types at the Go/JSON boundaries.

The renderer communicates with the Go supervisor through a versioned, framed stdio protocol. It has no database access or independent authority to broaden crawl scope. Go remains the source of truth for target validation and job state.

## Consequences

### Positive

- Goroutines, channels and contexts suit bounded per-host scheduling and cancellation.
- Native parallelism supports parsing, hashing, similarity and exports without an event-loop bottleneck.
- Go's race detector and built-in fuzzing directly support the threat model.
- Raw crawling, API, CLI and MCP can ship as compact cross-platform binaries.
- The core has a smaller runtime and dependency surface than an all-Node design.
- Internal contracts receive compile-time checking while public boundaries remain schema-validated.

### Negative

- The project uses two languages rather than one.
- UI and renderer contracts need generated schemas and compatibility tests.
- Go's HTML/data tooling is less convenient than Python's ecosystem.
- JavaScript rendering requires a supervised sidecar and separate packaging.
- SQLite driver and browser sandbox choices require platform spikes.

## Alternatives considered

### TypeScript/Node core

Fastest integration across UI, MCP and Playwright, but offers weaker memory predictability, more worker orchestration for CPU-heavy analysis, and less attractive native packaging for a long-running crawler.

### Python core

Fastest route to rich SEO analysis and excellent parsing/report libraries, but has higher packaging and memory costs and requires additional care for CPU parallelism and cross-platform distribution.

### Rust core

Offers the strongest low-level performance and memory safety, but materially increases development time and integration complexity without being necessary for the initial scale target.

## Validation requirements

Milestone 0 must validate:

1. guarded HTTP connections with DNS/IP pinning and correct TLS SNI;
2. SQLite write/read behaviour at 100,000 synthetic URLs;
3. Go MCP stdio initialization and a typed tool call;
4. Go-to-Node renderer framing, cancellation and crash handling;
5. builds on Linux, macOS and Windows.

If these spikes fail materially, this ADR must be revisited before expanding implementation.
