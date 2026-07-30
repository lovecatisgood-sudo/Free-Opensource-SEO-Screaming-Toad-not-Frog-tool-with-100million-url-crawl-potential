# SEO Auditor theoretical 100M+ architecture note

Status: Optional research note; not a release plan or capacity claim
Version: 2.0  
Date: 2026-07-30

## 1. Objective

Describe how one SEO crawl campaign could theoretically fetch, analyse, persist, recover and query more than 100 million unique URLs without weakening security, politeness, evidence or audit correctness.

This is not a v2.0 objective, release gate, supported limit or guarantee. It records architectural potential for optional future research. Current evidence remains limited to the published synthetic campaigns.

## 2. Definition of a crawl campaign

A campaign is one crawl with:

- one immutable scope and configuration snapshot;
- one global URL identity and authoritative deduplication set;
- one rule-pack version set;
- persistent frontier and retry history;
- continuous per-host robots and politeness state;
- global page, link, redirect, sitemap, canonical and hreflang relationships;
- recoverable checkpoints;
- a final terminal state and completeness statement.

A segment is an internal processing window, normally 10,000–100,000 leased URLs. Completing a segment checkpoints and compacts the campaign; it does not start a new crawl.

## 3. Quality contract at scale

The 100M goal does not permit a “lite” count-only crawl to be marketed as a full audit. A qualifying URL must retain, when applicable:

- requested and final URL identity;
- fetch outcome, timings, bytes and redirect evidence;
- indexability directives and canonical;
- title, description, headings and content hash;
- extracted internal links and discovery provenance;
- sitemap and robots evidence;
- hreflang, image and structured-data summaries configured for the benchmark;
- page-local rule results with stable rule versions.

Whole-campaign rules must produce exact results or be explicitly labelled as approximate/unsupported at that scale. Coverage limitations are part of the benchmark result.

## 4. Non-negotiable invariants

1. A committed URL result is idempotent.
2. No URL is omitted because an approximate cache reports a false positive.
3. Every frontier lease expires or reaches a terminal state.
4. Redirect, sitemap and renderer requests use the same target policy as seed requests.
5. Horizontal scaling never bypasses per-host rate and robots controls.
6. Segment boundaries do not change URL identity or crawl depth.
7. Cancellation and disk exhaustion stop new leasing safely.
8. Recovery cannot silently convert incomplete work into a completed campaign.
9. Rule and schema versions remain reproducible for the campaign lifetime.
10. Published capacity refers to unique qualifying URLs, not queue events or duplicate encounters.

## 5. Execution model

### 5.1 Local segmented execution

The Go supervisor leases a bounded segment from the disk-backed frontier. Workers fetch and extract using per-host scheduling. Results commit idempotently in smaller transactions. Newly discovered URLs enter the same authoritative frontier. At each segment boundary the system checkpoints counters and host state, compacts eligible temporary data, checks disk/time budgets, and continues automatically.

### 5.2 Distributed execution

At 100M scale, logical services separate while preserving the same domain contracts:

```text
Campaign coordinator
    ├── partitioned durable frontier
    ├── host/politeness ownership
    ├── fetch worker pools
    ├── extraction/rule workers
    ├── immutable result segments
    └── analytical/query store
```

Workers are replaceable and stateless beyond active leases. Commits include campaign, URL identity, attempt and schema versions so retries are idempotent. Coordinator failover uses durable state rather than reconstructing truth from worker memory.

## 6. Storage strategy

### Version 1

- SQLite WAL database.
- Normalised URL, fetch, page, link and issue tables.
- Managed export artifacts.

### Local campaigns

- Partition or rotate high-volume tables by campaign/segment if benchmarks require it.
- Use streaming compaction and indexes designed for keyset access.
- Keep optional raw bodies disabled by default.
- Forecast disk usage from observed bytes per URL/link before continuing.

### Distributed campaigns

- Operational store for frontier, leases, host state and idempotency.
- Immutable compressed columnar result segments in object storage.
- Analytical database for issue, page, link and aggregate queries.
- Explicit retention tiers for high-cardinality link and content data.

Storage technologies are selected through benchmarks rather than fixed prematurely. PostgreSQL and ClickHouse-class systems are candidates, not current commitments.

## 7. Analysis strategy

Rules are classified by scale behaviour:

| Class | Examples | Execution |
|---|---|---|
| Page-local | Missing title, noindex, missing H1 | Evaluate at page commit. |
| Edge-local | Broken link, internal nofollow | Evaluate when target outcome is known. |
| Incremental aggregate | Counts, status distributions | Merge exact partial aggregates. |
| Global graph | Orphans, final depth/inlinks, hreflang reciprocity | Partitioned graph passes or final analysis. |
| Similarity | Exact/near duplicates | Hash partitioning and bounded candidate generation. |

Every rule declares its execution class, data dependencies, exactness and supported scale. Rules that cannot retain their semantics at 100M are disabled with an explicit coverage notice rather than silently approximated.

## 8. Evidence and theoretical levels

| Level | Target | Current meaning |
|---|---:|---|
| Version-1 supported | 100,000 | Release regression and endurance suite. |
| Local campaign experimental | 1 million | Published single-machine benchmark and recovery test. |
| Local campaign experimental evidence | 5 million | One synthetic production-path run; not live or supported capacity. |
| Distributed architecture prototype | Unspecified | Coordinator and immutable-segment primitives, without a capacity claim. |
| Theoretical architecture direction | More than 100 million | Untested design potential only; not supported or guaranteed. |

The table is descriptive rather than a promotion ladder. v2.0 completion does not require additional large-scale campaigns.

The 1-million and single-run 5-million experimental synthetic production-path campaigns passed on 2026-07-30 with exact URL identity reconciliation, verified segments, zero outstanding work and SQLite integrity `ok`. The benchmark and its limitations are recorded in [the segmented campaign report](./benchmarks/2026-07-30-segmented-campaign.md).

## 9. Optional future validation protocol

If a future project ever seeks a supported capacity claim, a new approved PRD should require evidence including:

- exact unique qualifying URL count;
- product version, commit, schema and rule versions;
- benchmark dataset/generator and seed;
- enabled extraction fields and audit rules;
- hardware, operating system and storage topology;
- configured concurrency, delay and host distribution;
- wall-clock duration and sustained request rate;
- peak memory, CPU, network and storage use;
- total links and issue records;
- duplicate encounters, retries, skips and failures by reason;
- controlled worker/coordinator crash results;
- final frontier reconciliation and invariant checks;
- representative extraction/rule correctness samples;
- known limitations and features excluded from the run.

The verification tool must prove:

```text
accepted = committed + terminal_skipped + terminal_failed + outstanding
outstanding = 0 for a completed campaign
committed URL identities are unique
every committed segment checksum verifies
all required rule and schema versions are present
```

## 10. Public capability language

Approved language:

> Architecturally designed for segmented campaigns beyond 100 million URLs; this is a theoretical scalability target, not a tested, supported or guaranteed capacity.

Avoid “unlimited,” “number one,” “verified 100M,” “supports 100M,” “can crawl 100M,” and unqualified comparisons. The 1M and 5M evidence must always be labelled synthetic production-path evidence rather than live capacity.

## 11. Stop conditions

A campaign automatically pauses rather than risking corruption or host impact when:

- free disk crosses the configured safety reserve;
- database or object-store durability checks fail;
- coordinator state cannot prove lease ownership;
- a host exceeds adaptive error/rate-limit thresholds;
- configuration or rule-pack integrity changes unexpectedly;
- security policy cannot validate a target;
- resource projections exceed the campaign budget.

Operators receive the reason, recovery steps and completeness status. Automatic continuation resumes only after the invariant or budget condition is restored.
