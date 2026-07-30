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

This qualifies the 1-million level as experimental synthetic production-path evidence. It does not make 1 million live URLs a supported capacity, and it does not establish network throughput, multi-host politeness or guarded-fetch performance.

## 5,000,000-URL campaign

Status: passed once on the production persistence, extraction, audit, graph and verification path with the synthetic in-process transport. It started at `2026-07-30T00:39:00.101673427Z`, committed its last 100,000-page segment at `2026-07-30T04:27:26.973594409Z`, and completed global audits at `2026-07-30T04:54:36.42251258Z`. The approximately 4 hour 15 minute 36 second wall-clock window includes controlled restarts and defect investigation, so it is not a fetch-throughput result.

The campaign exposed two scale-dependent defects. The projected-storage stop used an immature early sample and was changed to wait for one full segment while retaining the hard absolute disk limit. Global near-duplicate matching also generated overly broad candidate scans; it now uses five SimHash bands and ten exact two-band passes. A Hamming distance of at most three must preserve at least one band pair, which bounds candidate generation without weakening the configured threshold. The resumed campaign completed with the corrected implementation.

Final read-only `seo-auditor-scale verify` result:

```json
{
  "crawl_id": "crawl_scale",
  "status": "completed",
  "discovered": 5000000,
  "committed_pages": 5000000,
  "unique_url_identities": 5000000,
  "outstanding": 0,
  "links": 4999999,
  "issues": 15884167,
  "completed_segments": 50,
  "invalid_segments": 0,
  "missing_required_fields": 0,
  "database_integrity": "ok",
  "passed": true
}
```

Final SQLite storage after process close and checkpoint was 11,127,939,072 bytes. All fifty 100,000-URL segment metadata checksums verified; the first was `b2ed1da69bb576a0455f51c58a3afc554e1a8099c2711038eb387320699cfe1b` and the fiftieth was `dd0d42b535513d058a18bfbe9c005856dc0e398ee5653d747c1875f9c2f65629`.

This is stronger experimental evidence for a local multi-million campaign, but one synthetic run on one environment does not qualify 5–10 million URLs as supported. It does not exercise DNS, TLS, redirects, robots, host scheduling, network retries, or renderer isolation. Repeating larger campaigns is optional research rather than a v2.0 release gate.

This report cannot support a 100M+ capacity claim. The product describes only a theoretical architectural potential beyond 100 million URLs and explicitly labels it untested and unsupported.
