# Segmented campaign benchmark log

Date: 2026-07-30  
Environment: Linux amd64 sandbox, Go 1.26.5, SQLite/WAL, synthetic in-process transport.

The public generator is `go run ./cmd/seo-auditor-scale run`. It exercises the production frontier, extraction, page rules, links, issue persistence, segment checkpoints, final graph rules and campaign verifier. It does not use the network fetcher and therefore is not evidence of live guarded-fetch capacity.

## 10,000-URL qualification trial

- requested/fetched/committed unique pages: 10,000 / 10,000 / 10,000;
- retained links: 9,999;
- retained issues: 21,011;
- completed checksummed segments: 1;
- outstanding work, invalid segments, missing required fields, failures: 0;
- SQLite integrity: `ok`;
- elapsed: 8.04 seconds before later batching changes.

## 1,000,000-URL campaign

Status: passed on the production persistence, extraction, audit, graph and verification path with the synthetic in-process transport. Started at `2026-07-29T19:57:47.055633315Z` and completed at `2026-07-29T20:27:23.368597417Z`. The approximately 29 minute 36 second wall-clock window includes controlled interruptions, defect investigation, storage relocation, recovery and finalisation; it is not a fetch-throughput benchmark.

The campaign deliberately survived writer contention and controlled process termination. At the first durable checkpoint it had 100,000 analysed pages and checksum `23e1e4dd08ee065df929676bf567921ee3643a66b3b1c503905d39345863cb27`. An interruption exposed a pre-fix two-transaction analysis/discovery gap: one retained internal-link target was absent from the frontier. The implementation now commits analysis and discoveries atomically, proves rollback behavior with a regression test and reconciles older retained-link evidence during restart. The same campaign recovered the missing identity and reached the exact target.

Final verifier result:

```json
{
  "crawl_id": "crawl_scale",
  "status": "completed",
  "discovered": 1000000,
  "committed_pages": 1000000,
  "unique_url_identities": 1000000,
  "outstanding": 0,
  "links": 999999,
  "issues": 3064710,
  "completed_segments": 10,
  "invalid_segments": 0,
  "missing_required_fields": 0,
  "database_integrity": "ok",
  "passed": true
}
```

The last resume processed 48,963 fetch calls and completed global finalisation in 418.13 seconds. Final SQLite storage was 2,182,422,528 bytes. All ten 100,000-URL segment metadata checksums verified; the first was `23e1e4dd08ee065df929676bf567921ee3643a66b3b1c503905d39345863cb27` and the tenth was `da714b806cf2225bda2a64d3e858dfae6da6f76f13f6b38d49372ae27361386a`.

This qualifies the 1-million level as experimental synthetic production-path evidence. It does not make 1 million live URLs a supported capacity, and it does not establish network throughput, multi-host politeness or guarded-fetch performance. A 5–10-million repeated supported-hardware benchmark is still required for the next local-campaign level.

This report cannot support a 100M+ public claim. That claim remains gated on a separate live guarded-fetch campaign with the complete evidence in `SCALE_STRATEGY.md`.
