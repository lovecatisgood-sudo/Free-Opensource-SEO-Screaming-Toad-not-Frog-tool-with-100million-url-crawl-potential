# SEO conformance suite

Status: Phase 0 baseline implemented
Schema version: 1
Fixture release: `core-rules.v1`

## Purpose

The conformance suite tests SEO findings as application behavior. It runs deterministic local HTTP pages and bounded resources through response handling, extraction and the same rule evaluators used by crawls, then compares observed findings with explicit expected findings and expected absences.

Run it with:

```bash
make conformance GO=.tools/go/bin/go
```

For machine-readable output:

```bash
.tools/go/bin/go run ./cmd/seo-auditor-conformance -format json
```

The command exits non-zero when an expected finding is missing, an unexpected finding appears, an HTTP status differs or a current rule family lacks its baseline positive/clean-control pair. CI runs the JSON form on Linux, macOS and Windows.

## Current evidence

`core-rules.v1` currently contains:

- one clean document asserting the absence of all AUD-01 through AUD-17 findings;
- at least one positive finding for every current rule family;
- 21 expected findings across response, metadata, headings, canonical, indexability, sitemap, robots, duplicates, hreflang, images, architecture, transport, structured data, mobile/AMP and PDF diagnostics;
- explicit classification as deterministic, recommendation, review or information;
- explicit raw evidence-source attribution.

The current generated result is 21 true positives, zero false positives and zero false negatives: precision 1.0000 and recall 1.0000 for the represented cases.

This is a **baseline-covered** result, not complete rule validation and not a claim of overall commercial-crawler parity. Boundary, malformed, graph-level, rendered and adversarial fixtures still need to be added before stable-release accuracy gates are satisfied.

## Fixture contract

Each manifest contains:

- stable manifest and case IDs;
- expected HTTP status;
- bounded headers and HTML body;
- optional virtual document URL for HTTPS-only semantics;
- crawl context such as sitemap, robots, depth, inlinks and duplicate count;
- expected findings with rule, message, severity, classification and evidence subset;
- explicit expected absences.

The parser rejects unknown fields, unsupported schema versions, duplicate/invalid IDs, non-root-relative fixture paths, invalid status codes and oversized bodies. Reports use stable finding ordering.

## Classification contract

| Classification | Meaning |
|---|---|
| `deterministic` | The represented technical condition can be established from the captured evidence. |
| `recommendation` | The evidence is observed, but the desired action depends on editorial or business intent. |
| `review` | Available evidence is insufficient for an automatic conclusion. |
| `information` | Context that should not be presented as a failure. |

Severity and classification are independent. For example, a missing title is deterministically observable, while a short title is an editorial recommendation. A valid observation still does not predict ranking impact.

## Evidence-source contract

Stored issues now expose one of:

- `raw`
- `rendered`
- `graph`
- `sitemap`
- `external_api`
- `lab`
- `field`

Migration 009 backfills legacy issues and adds indexed, checked columns. New raw and rendered findings write their source and classification directly; graph finalisation normalises aggregate provenance. The fields are returned through database queries, HTTP API, MCP, CSV/NDJSON/XLSX paths and the dashboard evidence view.

## Next fixture increments

1. Boundary and malformed cases for every applicable page-level rule.
2. Full-crawl graph cases for broken links, canonical targets, reciprocal hreflang, sitemap coverage, duplicates and orphan/depth behavior.
3. Raw-versus-rendered cases using the isolated renderer.
4. Adversarial fixture limits for oversized bodies, schemas, links and decompression.
5. Versioned Schema.org vocabulary and Google search-feature profile fixtures.
