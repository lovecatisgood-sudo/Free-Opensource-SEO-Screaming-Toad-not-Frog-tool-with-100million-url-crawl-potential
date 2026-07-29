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

Status: in progress. The campaign deliberately survived writer contention and controlled process termination. At the first durable checkpoint it had 100,000 analysed pages and checksum `23e1e4dd08ee065df929676bf567921ee3643a66b3b1c503905d39345863cb27`.

The interruption exposed a pre-fix two-transaction analysis/discovery gap: one retained internal-link target was absent from the frontier. The implementation now commits analysis and discoveries atomically and reconciles older retained-link evidence during restart. The same campaign recovered the missing identity and reached the exact 1,000,000 discovery count. Final metrics and verifier output will replace this status when complete.

This report cannot support a 100M+ public claim. That claim remains gated on a separate live guarded-fetch campaign with the complete evidence in `SCALE_STRATEGY.md`.
