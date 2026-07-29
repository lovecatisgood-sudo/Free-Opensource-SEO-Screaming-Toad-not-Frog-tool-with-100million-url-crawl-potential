# Closed beta protocol

The closed beta is a release gate, not an automated test and not yet recorded as complete. Participants must own or be authorised to audit each target and should include small static sites, common CMSs, multilingual sites, JavaScript applications, redirect migrations, large sitemaps, throttled hosts, and intentionally malformed fixtures.

For each target, record the immutable profile, coverage/terminal reason, representative raw and rendered evidence, false-positive and false-negative samples for every applicable Must rule, host load observations, cancellation behaviour, export consistency, and whether diagnostics exclude crawled content. Remove or anonymise customer URLs before sharing triage records.

Stop immediately on target instability, unexpected private-address resolution, unexplained scope expansion, data-integrity failure, or a high/critical security finding. Stable-release approval requires no unresolved high/critical issue, documented disposition of material rule errors, and explicit maintainer sign-off. Synthetic campaigns do not satisfy this beta.
