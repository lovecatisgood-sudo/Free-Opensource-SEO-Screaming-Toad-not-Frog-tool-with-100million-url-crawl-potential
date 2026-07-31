# Release and reproducibility

Release candidates are built from a clean, pinned commit in the sandbox with `SEO_AUDITOR_VERSION=VERSION scripts/build-release.sh`. The script rebuilds the embedded UI, runs Go tests and vet, cross-compiles raw-mode Go binaries for Linux x64/arm64, macOS x64/arm64, and Windows x64, generates a CycloneDX SBOM, and writes SHA-256 checksums.

The binaries use `-trimpath`, disable volatile VCS embedding, and receive version, commit and commit-derived build time through linker variables. Repeating the build with the same toolchains, lockfiles, source commit and `SOURCE_DATE_EPOCH` is intended to reproduce the same output. `SHA256SUMS` is relative to the release directory so it can be checked with `sha256sum -c SHA256SUMS` from that directory.

Cross-compilation proves that the Go source builds for each target; it is not a substitute for clean-machine runtime qualification. GitHub CI runs tests on Linux, macOS and Windows. Rendered mode additionally requires the exact Node/Playwright packages and Chromium runtime and is currently qualified through the Linux sandbox profile.

The local build does not sign artifacts. `.github/workflows/release.yml` runs the pinned qualification gates, creates cross-platform candidate archives, attaches GitHub OIDC artifact attestations, verifies the downloaded checksum set and opens a protected draft prerelease. Run `scripts/verify-release.sh RELEASE_DIR` on Unix or `scripts/verify-release.ps1 -ReleaseDirectory RELEASE_DIR` on Windows; with GitHub CLI installed, both also verify repository provenance attestations.

GitHub attestations prove workflow provenance but do not replace native operating-system trust. A stable public release still requires organization-owned Apple Developer ID/notarization and Windows Authenticode identities plus clean-machine qualification of the exact downloads. Until those identities and records exist, the automated workflow must remain a prerelease/candidate path and artifacts must not be described as natively signed. There is no automatic updater.
