# ADR-002: SQLite driver and writer strategy

Status: Accepted  
Date: 2026-07-30

## Context

The local application needs transactional recovery, bounded memory, streaming reads and a zero-service installation. The core must also build without a platform C toolchain.

## Decision

Use `modernc.org/sqlite` with foreign keys, WAL, a five-second busy timeout and normal synchronous mode. A single application-owned `database.Writer` serializes mutation transactions through a bounded queue. Query services use short-lived reads from the same connection pool. Migrations are embedded, ordered, transactional and recorded in `schema_migration`.

Generated artifacts remain files under the application-owned artifact directory; their metadata and checksums live in SQLite.

## Evidence

The foundation spike creates the full initial schema, verifies WAL and foreign keys, exercises concurrent submissions and rollback, and inserts 100,000 unique URL records in one bounded transaction. On the initial Linux development machine the insert took approximately 305 ms and reported approximately 1 MiB of live Go heap after completion. This is a schema/driver signal, not a crawl-throughput promise.

## Consequences

- The writer is the only mutation path used by application services.
- Long read transactions and unbounded result materialization are prohibited.
- The driver remains replaceable behind `internal/database`, but replacement requires a new ADR and migration tests.
- Campaign-scale storage beyond the supported local envelope must pass the scale-stage gates before becoming a product claim.

