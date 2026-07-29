# SEO Auditor

SEO Auditor is a local-first, open-source technical SEO crawler focused on safe network behaviour, durable recovery and evidence-backed audits. The crawler core is Go; the browser interface and isolated optional renderer use TypeScript.

The v2.0 implementation is being release-qualified. Raw crawling, rendered-page evidence, local UI/API/CLI/MCP workflows, recovery, reports, lifecycle diagnostics, and the 100,000-URL single-crawl gate are implemented. A 100M+ campaign claim remains explicitly unverified until the separate distributed-scale evidence gate passes.

## Development

The recommended environment is the non-root container sandbox:

```bash
docker compose up -d dev
docker compose exec dev bash
make test
```

To prove the Go suite does not depend on external network access:

```bash
make test-offline
```

See [DEVELOPMENT.md](./DEVELOPMENT.md) for the sandbox boundaries and host-toolchain alternative.

## Safety status

Implemented foundation controls include:

- HTTP/HTTPS URL normalization with userinfo rejection and IDNA identity;
- prohibited IPv4/IPv6 classification and mixed DNS-answer rejection;
- approved-address connection pinning with normal TLS hostname verification;
- SQLite/WAL migrations, foreign keys and serialized writes;
- MCP over stdio using the official Go SDK;
- strict TypeScript renderer framing with an 8 MiB frame ceiling.

These controls are verified by automated tests, but no crawler can make arbitrary targets risk-free. Start with conservative limits and only audit systems you are authorized to access. See the [security model](./docs/SECURITY_MODEL.md) and [operations guide](./docs/OPERATIONS.md).

## Third-party reference boundary

The local `open-seo-crawler/` checkout is untrusted reference material. It is ignored by Git, excluded from Docker build contexts and never installed or executed by this project. See [the reuse checklist](./docs/REFERENCE_REUSE_CHECKLIST.md).

## License

SEO Auditor is available under the [MIT License](./LICENSE). Third-party dependency and source notices are tracked in [THIRD_PARTY_NOTICES.md](./THIRD_PARTY_NOTICES.md).
