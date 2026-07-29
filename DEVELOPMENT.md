# Sandboxed development

SEO Auditor supports two development paths. Both keep generated state inside this workspace and outside source-controlled files.

## Container sandbox (recommended)

Run:

```bash
docker compose up -d dev
docker compose exec dev bash
make test
```

The development container:

- runs as the unprivileged `auditor` user;
- drops all Linux capabilities and enables `no-new-privileges`;
- does not mount the host Docker socket;
- publishes only port 7331 and only on host loopback;
- uses named caches for Go and pnpm dependencies;
- limits its ephemeral executable build/test tmpfs to 2 GiB (`nosuid`, `nodev`) and the Go runtime target to 4 GiB;
- stores application data, cache, exports and coverage under ignored workspace directories.

For tests that must prove they work without external network access:

```bash
make test-offline
```

That profile has `network_mode: none`; loopback fixture servers inside the container remain available.

## Host development

The host path uses the verified toolchain under `.tools/` and the same ignored state directories. No global package installation is required. The `.tools/` directory is disposable and excluded from container build contexts and version control.

## Trust boundary

`open-seo-crawler/` is third-party reference material. It is excluded from container build contexts and is not imported, executed, installed or linked by SEO Auditor. Any adopted behaviour must be independently implemented and covered by our own tests and notices.

Application crawling remains network-enabled only through the guarded fetch subsystem. Development isolation does not replace runtime SSRF, redirect, response-size, timeout and scope enforcement.
