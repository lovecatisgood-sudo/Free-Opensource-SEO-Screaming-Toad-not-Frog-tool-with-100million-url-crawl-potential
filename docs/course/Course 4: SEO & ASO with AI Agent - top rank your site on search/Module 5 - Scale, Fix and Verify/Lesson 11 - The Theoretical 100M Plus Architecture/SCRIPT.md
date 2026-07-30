# Lesson 11 — The Theoretical 100M+ Architecture

**Target duration:** 4 minutes

## Narration

[On screen: claim boundary before the architecture animation.]

SEO Screaming Toad is architecturally designed for segmented campaigns beyond 100 million URLs. This is a theoretical scalability target, not a tested, supported, or guaranteed capacity.

A naive crawler might try to hold an enormous queue and every result in memory. Toad instead uses a durable frontier and bounded work leases. At very large scale, a campaign could process segments—conceptually 10,000 or 100,000 URLs at a time—while retaining one global campaign.

Completing a segment does not mean starting an unrelated audit. The design preserves one scope, configuration snapshot, authoritative URL identity, robots and politeness state, retry history, rule versions, checkpoints, and final completeness statement.

[On screen: coordinator → durable frontier → host ownership → workers → result segments → query store.]

A future distributed design separates coordination, host politeness, fetching, analysis, immutable result segments, and analytical queries. Workers use expiring leases and idempotent commits so retries do not silently create duplicate authoritative results.

Quality matters more than a large counter. One hundred million queue events would not equal one hundred million unique audited URLs. A credible result would require URL uniqueness, retained evidence, stable rule versions, completed frontier reconciliation, and explicit coverage limitations.

Current published evidence is smaller: a 100,000-URL gate and synthetic production-path campaigns at one million and five million URLs. These are not 100-million live-network crawls.

[On screen: “Architecture shows a path. Benchmarks prove capacity.”]

The segmented design is exciting because it can continue bounded processing without pretending everything fits in memory. Real capacity still depends on hardware, storage, rendering, network behavior, target politeness, and future distributed implementation.
