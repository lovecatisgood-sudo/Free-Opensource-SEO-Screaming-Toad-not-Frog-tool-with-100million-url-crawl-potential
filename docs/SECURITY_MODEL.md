# Security and threat model

SEO Auditor assumes target pages, DNS answers, redirects, HTML, JavaScript, browser downloads, and imported configuration are hostile. It also assumes a malicious web page may try to reach loopback, private, link-local, metadata, or Unix-local services.

The Go fetch policy normalizes URLs, rejects userinfo and prohibited address classes, rejects mixed safe/unsafe DNS answer sets, pins approved connection addresses, revalidates every redirect, and preserves TLS hostname verification. Request counts, response bytes, time, depth, retries, concurrency and total URLs are bounded. SQLite uses migrations, foreign keys, WAL, transactional leases and restart recovery.

The API binds to loopback. Browser mutations require an unguessable session cookie, exact Origin and CSRF token. Unknown server errors are not reflected to clients. MCP uses stdio. Managed artifact paths are application-owned and downloads require the local API session.

Rendered mode is optional and higher risk. The Playwright worker is disposable, denies permissions, downloads, dialogs, WebSockets and service workers, uses fresh storage, and cannot fetch resources directly: document and subresource requests are mediated by the Go guard. In the Docker profile Chromium's own namespace sandbox is unavailable under the deliberately dropped capabilities; the worker instead sits inside the non-root, capability-free container boundary with a dead proxy. Native deployments retain Chromium sandboxing. Do not run rendered mode outside one of these documented boundaries.

Residual risks include browser and parser vulnerabilities, DNS changes after validation, target-side denial of service caused by aggressive operator limits, disk exhaustion, sensitive local reports, and false-positive or false-negative SEO findings. Keep dependencies patched, use conservative limits, protect the data directory, and review diagnostics before sharing.

Report vulnerabilities privately as described in [SECURITY.md](../SECURITY.md). This document describes engineering controls; it is not a penetration-test certification.
