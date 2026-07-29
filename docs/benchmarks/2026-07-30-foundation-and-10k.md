# Foundation and 10,000-URL benchmark

Date: 2026-07-30  
Environment: Linux amd64 development container, Go 1.26.5, SQLite/WAL using `modernc.org/sqlite` 1.55.0.

## Persistence spike

- Workload: insert 100,000 unique normalized URL-shaped rows in one serialized transaction.
- Result: passed.
- Observed elapsed time: approximately 305 ms.
- Reported live Go heap after completion: approximately 1 MiB.

## Durable crawl-engine spike

- Workload: deterministic 10,000-page, ten-way HTML link tree.
- Fetching: bounded in-process fixture; no external network.
- Persistence: every discovery, lease, fetch completion and counter transition persisted in SQLite.
- Concurrency: 32 global workers, 8 per host, no artificial fixture delay.
- Result: 10,000 unique discoveries and 10,000 fetches; zero duplicate fetches; terminal state `completed`.
- Observed elapsed time: approximately 4.42 seconds.

## Interpretation

These are regression baselines for storage and scheduler mechanics, not claims about live-site throughput. Real crawling includes DNS, TLS, server politeness, robots, response transfer, parsing and audit work. Public capacity statements remain governed by `SCALE_STRATEGY.md` and its qualifying evidence requirements.
