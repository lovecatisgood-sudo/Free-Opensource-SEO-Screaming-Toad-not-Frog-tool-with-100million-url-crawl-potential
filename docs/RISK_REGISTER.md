# Risk register

Last reviewed: 2026-07-30. Owners are project roles until named maintainers are assigned.

| Risk | Rating | Owner | Current mitigation | Release disposition / next review |
|---|---:|---|---|---|
| Guard bypass enables SSRF or private-service access | critical | security maintainer | central normalization/DNS/IP/redirect policy; pinned connections; renderer mediation; adversarial tests | no known bypass; review every outbound-path change |
| Browser exploit escapes renderer boundary | high | security maintainer | disposable worker, denied permissions/downloads/WebSockets/service workers, Go-mediated fetches, non-root capability-free container | residual browser/runtime risk; keep Playwright pinned and scanned |
| Crash loses discovered URLs after page commit | high | storage maintainer | fetch, extraction, issues, links, and accepted discoveries now share one transaction; legacy evidence/frontier reconciliation on restart | fixed and regression-tested; recheck after distributed result commits |
| SQLite write contention interrupts local campaigns | high | storage maintainer | single in-process writer, WAL, bounded batches, migration-free read-only diagnostics, durable recovery | 1M and 5M synthetic campaigns recovered and passed; repeated hardware profiles still required for support |
| Disk exhaustion corrupts or strands a campaign | high | storage maintainer | actual usage ceiling, projected final usage after one full segment, safe terminal reason, checksummed segments, backups | 5M campaign stayed within its absolute ceiling after estimator correction; validate reserve-based pause on repeated profiles |
| Distributed split-brain violates politeness or commits conflicting data | high | scale maintainer | coordinator and host-owner fencing epochs; expiring leases; immutable idempotent segment commits | prototype only; requires multi-node failover test before support |
| SEO findings are false positives/negatives | high | rules maintainer | versioned evidence, remediation limitations, exact/near duplicate separation, raw/rendered distinction | closed beta sampling still required before public stable release |
| Dependency or build compromise | high | release maintainer | pinned Go/pnpm/toolchain inputs, CI, Dependabot, audits, CycloneDX SBOM, checksums, no auto-update | signing identity not configured; unsigned candidates cannot be called signed |
| Reports or diagnostics leak sensitive crawl data | medium | privacy maintainer | managed paths, local auth, crawled-content exclusion in diagnostics, no raw-body storage | operators must review reports before sharing |
| Capacity claim overstates verified behaviour | high | release maintainer | qualifying-field contract, campaign verifier, explicit synthetic/live transport labels, public-language gate | single-run 5M synthetic production-path evidence passed; supported 5–10M and verified 100M+ claims remain prohibited until their gates pass |
| Cross-platform build is mistaken for runtime support | medium | release maintainer | CI matrix and explicit cross-build/runtime distinction | clean-machine Linux/macOS/Windows qualification remains a release gate |

No risk is accepted permanently. A release approval record must link evidence for every high/critical row and date any exception.
