# SEO Auditor final planning baseline

Product codename: **SEO Auditor**. Public naming may change without changing the product or architecture baseline.

This directory is the approved planning source of truth for the secure, local-first SEO crawler as of 2026-07-30.

1. [Final product plan and PRD](./PRD.md) — what the product must do and how success is measured.
2. [Final system architecture plan](./ARCHITECTURE.md) — how the product will be built and secured.
3. [Final implementation plan](./IMPLEMENTATION_PLAN.md) — sequence, milestones, traceability, tests, and release gates.
4. [SEonaut reference review](./references/SEONAUT_REVIEW.md) — reusable ideas and explicit non-reuse boundaries from the Go crawler.
5. [ADR-001: Go core](./adr/ADR-001-go-core.md) — accepted language and runtime decision.
6. [100M scale strategy](./SCALE_STRATEGY.md) — staged path to verified 100-million-URL crawl campaigns without reducing audit quality.
7. [Operations guide](./OPERATIONS.md) — installation, CLI/MCP use, updates, storage, backup and troubleshooting.
8. [Security model](./SECURITY_MODEL.md) — trust boundaries, controls and residual risks.
9. [Audit rule catalog](./RULE_CATALOG.md) — versioned rule coverage and limitations.
10. [Release and reproducibility](./RELEASE.md) — cross-platform artifacts, SBOM, checksums and signing limits.
11. [Risk register](./RISK_REGISTER.md) — owned security, integrity, quality, release and scale risks.
12. [Closed beta protocol](./BETA_PROTOCOL.md) — authorised-site validation and stable-release stop conditions.
13. [Project state checkpoint](./PROJECT_STATE.md) — latest implementation, verification, live-audit and resume state.

## Document precedence

If the documents disagree, use this order:

1. Security and privacy requirements in the PRD.
2. Product scope and acceptance criteria in the PRD.
3. Architectural decisions in the architecture document.
4. Scheduling and task breakdown in the implementation plan.

Changes that alter product scope, security boundaries, supported platforms, or stored data must update all affected documents in the same change.

## Baseline decisions

- Go core for crawler, security policy, persistence, rules, API, CLI and MCP.
- TypeScript for the React UI and isolated Node/Playwright renderer only.
- SQLite/WAL for the version-one local application.
- Loopback-only API and MCP over stdio.
- Version-one supported target: 100,000 URLs with bounded memory and durable recovery.
- Post-version-one north star: a published, qualifying 100M+ URL campaign without reducing audit quality.
- No “100M+ verified” public claim until the benchmark gate passes.
