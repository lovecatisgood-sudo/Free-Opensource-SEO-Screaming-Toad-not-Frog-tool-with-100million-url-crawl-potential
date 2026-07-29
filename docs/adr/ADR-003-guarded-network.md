# ADR-003: Guarded DNS resolution and connection pinning

Status: Accepted  
Date: 2026-07-30

## Context

A crawler is an outbound network capability. URL string validation alone does not prevent access to loopback, private services, cloud metadata, DNS-rebinding targets or redirects that cross the security boundary.

## Decision

Every request uses `internal/fetchpolicy`. URLs are normalized, schemes are limited to HTTP/HTTPS, userinfo is rejected and internationalized hostnames use a stable ASCII identity. All A/AAAA results are classified before dialing. If any answer is prohibited, the entire mixed answer set is rejected. The dialer connects only to an approved numeric address while the original hostname remains available to HTTP for the Host header and TLS SNI/certificate verification.

The guarded transport ignores ambient HTTP proxy environment variables. Every redirect will re-enter the complete policy. No insecure TLS retry is permitted.

## Consequences

- Fixture tests must cover IPv4, IPv6, mapped IPv4, loopback, private, link-local, CGNAT, multicast, documentation and mixed-answer cases.
- Renderer requests and sitemap discovery must call the same policy rather than implementing equivalents.
- Private target support, if ever added, is a trusted local installation policy and is never exposed through MCP.

