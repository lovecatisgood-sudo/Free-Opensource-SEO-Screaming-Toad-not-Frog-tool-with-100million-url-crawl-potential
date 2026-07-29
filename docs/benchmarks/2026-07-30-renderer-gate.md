# Renderer security and evidence gate — 2026-07-30

Environment: pinned Docker development image, Go 1.26.5, Node.js 22.23.1, Playwright 1.62.0, bundled Chromium revision 1234.

Verified outcomes:

- A deterministic HTML fixture loaded its script only through the Go supervisor and produced a client-mutated title, text, and link.
- A `127.0.0.1` image request reached the supervisor mediation boundary and was rejected; Chromium had a dead proxy/failing resolver as a second network-deny layer.
- Geolocation was denied, service-worker registration was rejected, downloads were cancelled, WebSockets were closed, and a second render did not observe local-storage state from the first.
- An infinite JavaScript loop was terminated within the supervisor bound, and an abruptly exiting worker was reaped promptly.
- Raw title/content remained in the `page` record. Rendered fields, costs, differences, links, and issues were stored with separate provenance and returned by page detail/query APIs.

Commands:

```text
pnpm --dir web/renderer typecheck
pnpm --dir web/renderer test
pnpm --dir web/renderer build
SEO_AUDITOR_RENDERER_INTEGRATION=1 go test ./internal/renderer -v
go test ./internal/database ./internal/crawler
```

Result: all renderer-specific gates passed in the sandbox. Performance claims are intentionally excluded; this gate proves behavior and containment, not crawl throughput.
