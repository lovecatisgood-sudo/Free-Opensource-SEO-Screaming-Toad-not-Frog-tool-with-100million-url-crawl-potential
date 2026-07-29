# Third-party notices

SEO Auditor is implemented from original source and distributed under the MIT License. No source was copied from either reference crawler reviewed during planning.

Runtime and build dependencies remain under their own licenses. The lockfiles are the authoritative version inventory. Direct dependencies are:

| Component | Purpose | License |
|---|---|---|
| Go MCP SDK | MCP stdio server | MIT |
| `golang.org/x/net` | network and IDNA support | BSD-3-Clause |
| `modernc.org/sqlite` | embedded SQLite driver | BSD-3-Clause |
| React / Vite / Vitest | local UI and tests | MIT |
| Playwright | isolated optional renderer | Apache-2.0 |
| TypeScript | renderer and UI compilation | Apache-2.0 |

Transitive JavaScript packages include MIT, Apache-2.0, BSD-3-Clause, ISC, and MPL-2.0 components. MPL-2.0 applies to the relevant dependency files; SEO Auditor does not modify or relicense those files. Run `pnpm licenses list --json` in the sandbox and `go list -m -json all` to reproduce the complete dependency inventory for a given lockfile revision.

The `open-seo-crawler/` directory is reference material and is not part of the SEO Auditor build, distribution, or runtime. No source from that directory may be copied into the product without an explicit provenance review and an entry in this file.
