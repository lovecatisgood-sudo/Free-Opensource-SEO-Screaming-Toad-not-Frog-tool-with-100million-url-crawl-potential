# Security policy

SEO Auditor is under active development. Release candidates are not yet a supported stable public release.

Please do not disclose suspected vulnerabilities in a public issue. Use the repository host's private vulnerability-reporting feature when configured. Until then, contact the maintainers privately and include the affected version/commit, reproduction steps, impact and any suggested remediation. Do not include target credentials, private URLs, crawl databases, or raw reports unless an encrypted transfer has been agreed.

The crawler must only be used on sites the operator is authorized to audit. Security controls—including prohibited-network targeting, redirect validation, TLS verification, crawl budgets and robots policy—must not be bypassed to increase crawl coverage.

Security support currently covers the latest source revision and most recent release candidate. No fixed response SLA is offered before stable release. The threat model and residual risks are documented in `docs/SECURITY_MODEL.md`.
