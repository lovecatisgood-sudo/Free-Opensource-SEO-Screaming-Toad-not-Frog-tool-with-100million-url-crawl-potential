# ADR-007: Renderer isolation and network mediation

Status: Accepted  
Date: 2026-07-30

## Context

JavaScript rendering executes untrusted site code in Chromium. It must not inherit database access, choose local files, bypass crawl scope, contact private networks, preserve browser state, or consume resources without bounds. Chromium's native Linux namespace sandbox is also unavailable inside the development container because the container intentionally drops every capability and enables `no-new-privileges`.

## Decision

- Go remains the security-policy and network authority. A renderer worker asks Go for every document and subresource over a versioned, length-prefixed stdio protocol. Go revalidates the URL with the guarded fetcher before returning bounded bytes.
- The worker has no database credentials or user-selected executable/file paths. Its Node binary, compiled worker path, browser cache, and sandbox mode are trusted process-start configuration only.
- Each page currently receives a fresh Node worker, Chromium browser, and context. This is deliberately less throughput-efficient than pooling, but gives deterministic crash recovery and prevents storage/cookie reuse. A future pool must retain fresh contexts and pass equivalent recycle/crash gates.
- Downloads, dialogs, permissions, service workers, WebSockets, fonts, and media are disabled or cancelled. Response cookies are stripped before fulfillment.
- Browser traffic is intercepted. Chromium is also configured with a dead loopback proxy and a failing host resolver so an un-intercepted request has no usable external path.
- Native installations use Chromium's own sandbox. In the hardened Docker development/runtime mode, Chromium runs without its namespace sandbox because Docker provides the outer process boundary: non-root user, all capabilities dropped, `no-new-privileges`, bounded tmpfs, and controlled mounts. Packaged native releases must not enable container mode.
- A render has independent deadline, request-count, per-resource, aggregate-byte, HTML-size, protocol-frame, and stderr limits. Failure is recorded with a stable taxonomy and raw extraction remains the fallback.
- Raw and rendered records are separate. Rendered issues use `subject_type=rendered_page` and evidence contains `extraction_mode=rendered`; page detail and exports expose both values and their differences.

## Consequences

Rendered crawling costs an additional guarded document request plus subresources and is therefore slower than raw crawling. External resources outside the configured crawl scope are rejected in this version. Development source mounts are writable for compilation, so untrusted live-site rendering should use the release runtime profile rather than treating the development container as a production multi-tenant boundary.

## Verification

The automated gates cover framed-protocol validation, worker crash reaping, JavaScript DOM mutation, client-link discovery, private-target mediation, blocked permissions and service workers, cancelled download behavior, non-persistent browser storage, renderer deadline enforcement, raw/rendered database provenance, and labeled exports.
