# SEO Auditor

SEO Auditor is a local-first, open-source technical SEO crawler focused on safe network behaviour, durable recovery and evidence-backed audits. The crawler core is Go; the browser interface and isolated optional renderer use TypeScript.

The project is under active development and does not yet have a supported public release. The current supported development baseline is documented in [the planning index](./docs/README.md).

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

These are foundation components, not a claim that the crawler is release-ready. Do not use development builds against systems you are not authorized to audit.

## Third-party reference boundary

The local `open-seo-crawler/` checkout is untrusted reference material. It is ignored by Git, excluded from Docker build contexts and never installed or executed by this project. See [the reuse checklist](./docs/REFERENCE_REUSE_CHECKLIST.md).

## License

A product license will be selected before the first public release. Third-party dependency and source notices are tracked in [THIRD_PARTY_NOTICES.md](./THIRD_PARTY_NOTICES.md).
