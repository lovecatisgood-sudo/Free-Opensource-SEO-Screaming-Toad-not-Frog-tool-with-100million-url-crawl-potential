# Changelog

All notable changes to SEO Screaming Toad are documented in this file.

The project follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/)
and uses [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased] - 2026-07-31

### Added

- Added 17 versioned SEO audit families with finding classification, evidence-source provenance, deterministic conformance fixtures, and evidence-backed explanations.
- Added Schema.org 30.0 vocabulary validation for types and properties, superseded-term detection, domain/range diagnostics, and bounded JSON-LD node validation.
- Added pinned Google Breadcrumb, Article, and merchant Product structured-data profiles.
- Added bounded CSS and XPath custom audits.
- Added raw-versus-rendered DOM comparison, managed rendered DOM and screenshot artifacts, browser console and resource diagnostics, and pinned axe-core accessibility checks.
- Added optional PageSpeed Insights, CrUX, Search Console, and GA4 integrations with separate lab, field, and external evidence.
- Added host-bound bearer, basic, and cookie authentication for raw crawling through opaque OS credential references and OAuth refresh support.
- Added bounded PDF-readiness, advanced image-delivery, responsive viewport, mobile-alternate, and AMP diagnostics.
- Added internal-link architecture analysis with segment, depth, inlink, outlink, score, and orphan evidence.
- Added recurring local audit schedules, artifact expiry, and orphan-file reconciliation.
- Expanded the MCP interface to 36 bounded tools covering crawls, evidence, architecture, schedules, custom audits, and integrations.
- Added a candidate-release workflow with deterministic cross-platform archives, CycloneDX SBOM generation, checksums, GitHub provenance attestations, verification scripts, and protected draft prereleases.
- Added professional product, implementation, commercial-quality, conformance, and course documentation.

### Changed

- Extended API, database, report, renderer, dashboard, and rule contracts to retain and expose evidence provenance, rendered diagnostics, integrations, schedules, custom audits, and architecture data.
- Updated the dashboard, public documentation, release guidance, rule catalog, and embedded production assets for the expanded professional-quality feature set.
- Shortened course paths and removed Windows-forbidden filename characters so the repository can be checked out consistently across supported platforms.
- Reworked renderer worker supervision so crashed Node processes are detected and reaped promptly on Windows without waiting for the outer render deadline.

### Security

- Added OS-backed secret storage, credential rotation, transport-error redaction, and protections against leaking credential values through profiles, crawl data, exports, diagnostics, or MCP responses.
- Bound raw-crawl credentials to configured host and subdomain boundaries so redirects cannot forward them outside the approved scope.
- Continued to reject authenticated rendered crawling until equivalent browser credential-boundary guarantees can be implemented and qualified.

### Verification

- Passed the full Go test suite, `go vet ./...`, and the full Go race-detector suite.
- Passed dashboard and renderer type checking, tests, and production builds.
- Passed real Chromium mediation, screenshot, axe, state-isolation, and hung-page tests.
- Achieved 21 true positives, zero false positives, and zero false negatives across all 17 represented conformance rule families.
- Cross-compiled Linux amd64/arm64, macOS amd64/arm64, and Windows amd64 artifacts and verified repeated archive hashes and package checksums.
- Passed release-script syntax and safety checks and `git diff --check`.

### Known limitations

- Apple notarization and Windows Authenticode still require DJAI-owned signing identities; clean-machine runtime qualification also requires macOS and Windows machines.
- The GitHub candidate workflow has not yet been executed or published as a formal release.
- Google rich-result validation currently covers three profiles rather than every Google profile.
- PDF and image analysis remain bounded diagnostics rather than full commercial-grade parsers.
- Search Console URL Inspection and sitemap enrichment, a locally isolated Lighthouse worker, and additional authorized closed-beta audits remain future work.
