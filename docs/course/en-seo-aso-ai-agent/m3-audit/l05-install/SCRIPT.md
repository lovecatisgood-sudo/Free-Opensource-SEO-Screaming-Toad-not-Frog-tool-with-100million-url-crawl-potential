# Lesson 5 — Install and Launch Safely

**Target duration:** 4 minutes

## Narration

[On screen: exact official GitHub owner and repository name.]

Install SEO Screaming Toad only from the official `lovecatisgood-sudo` GitHub repository. Before running an open-source project, verify the owner, read its README, security model, development guide, project state, license, and recent changes.

The easiest assisted setup is to open Codex or Claude Code in an empty development folder and use this prompt:

[On screen: prompt card.]

“Clone the official SEO Screaming Toad repository from `lovecatisgood-sudo`. Read the README, security model, development guide, and project state before executing anything. Explain the required Go and Node versions and inspect the build scripts. Ask before installing system-level dependencies. Build and start the local application, but do not crawl any website yet.”

The documented manual path is to clone the repository, enter its folder, run the bootstrap step, and start the Go application. Node and pnpm are required for frontend development or JavaScript-rendered mode. A sandboxed Docker development path is also available.

[On screen: `make bootstrap`, `go run ./cmd/seo-auditor`, `http://127.0.0.1:7331`.]

The default application listens on the local loopback interface at `127.0.0.1:7331`. It is intended for your own machine, not as a public internet service.

Start with a small authorized website and a conservative URL ceiling. Never tell an agent to bypass TLS verification, robots enforcement, scope rules, or network safety controls just to make a crawl run.

If setup fails, ask the agent to show the exact command, error output, and relevant documentation. Fix the cause; do not weaken the safety boundary.
